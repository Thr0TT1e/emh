// EMH MCP Server — отдельный бинарник для интеграции с LLM-агентами.
//
// Запуск:
//
//	./emh-mcp-server                    # stdio-транспорт (для Claude, Cursor)
//	./emh-mcp-server --transport=http   # HTTP+SSE (для удалённых агентов)
//
// Конфигурация переиспользует config.yaml из cmd/api.
package main

import (
	"context"
	"crypto/subtle"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mark3labs/mcp-go/server"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/llm"
	"codeberg.org/Thr0TT1e/emh/backend/internal/mcp"
	"codeberg.org/Thr0TT1e/emh/backend/internal/repository/pg"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
	"codeberg.org/Thr0TT1e/emh/backend/pkg/env"
)

func main() {
	// Флаги командной строки.
	// дефолтный адрес — только loopback. Публикация наружу должна быть
	// осознанным решением оператора, а не значением по умолчанию.
	transport := flag.String("transport", "stdio", "Транспорт: stdio или http")
	addr := flag.String("addr", "127.0.0.1:3481", "Адрес HTTP-сервера (только для --transport=http)")
	flag.Parse()

	// 1. Загрузка конфигурации (переиспользуем из cmd/api)
	cfgPath := env.Get("CONFIG_PATH", "config/config.yaml")
	cfg, err := config.LoadMCP(cfgPath)
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.App.LogLevel)
	logger.Info("Конфигурация загружена", "path", cfgPath, "app", cfg.App.Name)

	// 2. Подключение к PostgreSQL
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		logger.Error("Не удалось разобрать конфигурацию БД", "error", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = 5 // MCP-серверу не нужен большой пул
	poolConfig.MinConns = 1

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("Не удалось создать пул БД", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("Не удалось проверить связь с БД", "error", err)
		os.Exit(1)
	}
	logger.Info("БД подключена", "max_conns", poolConfig.MaxConns)

	// 3. Wiring зависимостей
	heroRepo := pg.NewHeroRepository(dbPool)
	heroUC := usecase.NewHeroUseCase(heroRepo)

	photoRepo := pg.NewPhotoRepository(dbPool)
	heroAwardRepo := pg.NewHeroAwardRepository(dbPool)
	heroConflictRepo := pg.NewHeroConflictRepository(dbPool)
	heroLocationRepo := pg.NewHeroLocationRepository(dbPool)
	heroSourceRepo := pg.NewHeroSourceRepository(dbPool)
	heroRelationRepo := pg.NewHeroRelationRepository(dbPool)

	heroQueryUC := usecase.NewHeroQueryUseCase(
		heroRepo, photoRepo, heroAwardRepo, heroConflictRepo,
		heroLocationRepo, heroSourceRepo, heroRelationRepo,
	)

	submissionRepo := pg.NewSubmissionRepository(dbPool)
	submissionUC := usecase.NewSubmissionUseCase(submissionRepo, heroRepo, nil, logger)

	conflictRepo := pg.NewConflictRepository(dbPool)
	conflictUC := usecase.NewConflictUseCase(conflictRepo)

	// LLM-провайдер (опционально)
	var extractionUC usecase.ExtractionUseCase
	if cfg.LLM.Enabled {
		provider, err := llm.NewLLMProvider(cfg.LLM)
		if err != nil {
			logger.Error("Не удалось создать LLM-провайдер", "error", err)
			os.Exit(1)
		}
		extractionLogRepo := pg.NewLLMExtractionLogRepository(dbPool)
		extractionUC = usecase.NewExtractionUseCase(
			provider,
			extractionLogRepo,
			heroRepo,
			cfg.LLM.MaxInputChars,
			logger,
		)
		logger.Info("LLM-провайдер инициализирован",
			"provider", provider.Name(),
			"model", provider.Model(),
		)
	} else {
		logger.Info("LLM отключён, extract_hero_data будет недоступен")
	}

	// 4. Создание MCP-сервера
	mcpServer := mcp.NewMCPServer(mcp.ServerConfig{
		HeroQueryUC:  heroQueryUC,
		HeroUC:       heroUC,
		SubmissionUC: submissionUC,
		ConflictUC:   conflictUC,
		ExtractionUC: extractionUC,
		Logger:       logger,
	})

	// 5. Запуск транспорта
	switch *transport {
	case "stdio":
		logger.Info("Запуск MCP-сервера в stdio-режиме")
		if err := server.ServeStdio(mcpServer); err != nil {
			logger.Error("MCP stdio server failed", "error", err)
			os.Exit(1)
		}

	case "http":
		// H3 (security.md): HTTP-режим отдаёт инструменты (создание заявок,
		// LLM-экстракция — расход токенов внешних провайдеров) без авторизации.
		// Требуем непустой AUTH_API_KEYS и проверяем Bearer-токен на каждый запрос.
		if len(cfg.Auth.APIKeys) == 0 {
			logger.Error("HTTP-режим MCP требует хотя бы один ключ в AUTH_API_KEYS; запуск отклонён")
			os.Exit(1)
		}

		sseServer := server.NewSSEServer(mcpServer)
		httpServer := &http.Server{
			Addr:    *addr,
			Handler: requireAPIKey(cfg.Auth.APIKeys, logger)(sseServer),
			// Slowloris-защита. WriteTimeout намеренно не задан — SSE держит
			// соединение открытым для потоковой отправки событий.
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		}

		logger.Info("Запуск MCP-сервера в HTTP-режиме", "addr", *addr)
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Error("MCP HTTP server failed", "error", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Неизвестный транспорт: %s\n", *transport)
		flag.Usage()
		os.Exit(1)
	}
}

// requireAPIKey возвращает middleware, требующий заголовок
// "Authorization: Bearer <ключ>" с одним из ключей AUTH_API_KEYS.
//
// Сравнение — константным по времени для каждого known-ключа (subtle.ConstantTimeCompare),
// что приемлемо для случайных ключей достаточной длины (см. security.md, low priority hardening).
func requireAPIKey(keys []string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" || !matchesAnyKey(token, keys) {
				logger.Warn("MCP HTTP: unauthorized request", "remote", r.RemoteAddr, "path", r.URL.Path)
				w.Header().Set("WWW-Authenticate", `Bearer realm="emh-mcp"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func matchesAnyKey(token string, keys []string) bool {
	for _, k := range keys {
		if k == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(k)) == 1 {
			return true
		}
	}
	return false
}

// newLogger создаёт структурированный логгер с уровнем из конфигурации.
func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}
