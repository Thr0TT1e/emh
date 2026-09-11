package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/middleware"
	"codeberg.org/Thr0TT1e/emh/backend/pkg/env"
)

func main() {
	// 1. Конфигурация
	cfgPath := env.Get("CONFIG_PATH", "config/config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию.", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.App.LogLevel)
	logger.Info("Конфигурация загружена", "path", cfgPath, "app", cfg.App.Name)

	// 2. Инфраструктура (БД, S3)
	ctx := context.Background()
	dbPool, err := connectDatabase(ctx, cfg, logger)
	if err != nil {
		logger.Error("Не удалось подключиться к БД", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	mediaStorage, err := connectStorage(ctx, cfg, logger)
	if err != nil {
		logger.Error("Не удалось создать хранилище медиа", "error", err)
		os.Exit(1)
	}

	// 3. Wiring зависимостей
	deps, err := wireDependencies(cfg, logger, dbPool, mediaStorage)
	if err != nil {
		logger.Error("Не удалось инициализировать зависимости", "error", err)
		os.Exit(1)
	}

	// 4. Фоновые воркеры
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	startWorkers(workerCtx, cfg, logger, deps)

	// 5. Rate Limiting
	rateLimitMW := middleware.NewRateLimitMiddleware(cfg.RateLimit, cfg.Auth.TrustedProxies, logger)
	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()
	rateLimitMW.Start(rateLimitCtx)
	deps.ContactRateLimiter.Start(rateLimitCtx)

	// 6. HTTP-роутер
	mux := buildRouter(cfg, logger, deps, deps.AuthInterceptor, rateLimitMW)

	server := &http.Server{
		Addr:              cfg.App.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("Запуск сервера", "addr", cfg.App.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Сервер аварийно завершён", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Выключение сервера...")

	// 7.1. Останавливаем фоновые задачи через отмену контекстов.
	// Run() воркеров выходит из цикла через ctx.Done().
	workerCancel()    // воркеры (thumbnail, orphan, auth_audit, refresh_token)
	rateLimitCancel() // rate limit cleanup loop

	// 7.2. Drain очередей: ждём завершения in-flight отправок.
	// Stop() закрывает канал и вызывает wg.Wait().
	// Вызываем ПОСЛЕ workerCancel(), чтобы Run() уже вышел из цикла.
	if deps.submissionEmailWorker != nil {
		deps.submissionEmailWorker.Stop()
		logger.Info("submission email worker stopped")
	}
	if deps.alertWorker != nil {
		deps.alertWorker.Stop()
		logger.Info("alert worker stopped")
	}
	if deps.authAuditWorker != nil {
		deps.authAuditWorker.Stop()
		logger.Info("auth audit worker stopped")
	}

	// 7.3. Graceful shutdown HTTP-сервера.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Сервер был вынужден отключиться", "error", err)
	}
	logger.Info("Сервер завершил работу корректно.")
}
