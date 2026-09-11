package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/netutil"
)

// rateLimiter хранит лимитер и время последнего обращения.
type rateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware chi-middleware для ограничения запросов по IP и группе.
type RateLimitMiddleware struct {
	cfg          config.RateLimitConfig
	proxyChecker *netutil.ProxyChecker
	logger       *slog.Logger
	mu           sync.Mutex
	limiters     map[string]*rateLimiter
}

// NewRateLimitMiddleware создаёт middleware с конфигурацией.
// trustedProxies — список доверенных прокси (Caddy): только для них
// учитывается X-Forwarded-For, иначе IP подделывается и лимиты обходятся.
func NewRateLimitMiddleware(cfg config.RateLimitConfig, trustedProxies []string, logger *slog.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cfg:          cfg,
		proxyChecker: netutil.NewProxyChecker(trustedProxies),
		logger:       logger,
		limiters:     make(map[string]*rateLimiter),
	}
}

// Start запускает фоновую горутину очистки записей.
// Вызывается в main.go, завершается при отмене контекста.
func (m *RateLimitMiddleware) Start(ctx context.Context) {
	if !m.cfg.Enabled {
		return
	}
	go m.cleanupLoop(ctx)
	m.logger.Info("Запущено промежуточное ПО для ограничения скорости.",
		"auth_limit", m.cfg.AuthLimit,
		"submission_limit", m.cfg.SubmissionLimit,
		"media_limit", m.cfg.MediaLimit,
		"read_limit", m.cfg.ReadLimit,
		"admin_limit", m.cfg.AdminLimit,
	)
}

// Handler возвращает chi-compatible middleware.
func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate-limiting отключён — пропускаем
		if !m.cfg.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Preflight OPTIONS не лимитируем
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Health check не лимитируем
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		ip := netutil.ClientIP(m.proxyChecker, r.Header, r.RemoteAddr)
		group := classifyService(r.URL.Path)
		key := ip + ":" + group

		limit, interval := m.getLimit(group)
		limiter := m.getLimiter(key, limit, interval)

		if !limiter.Allow() {
			m.logger.Warn("rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path,
				"group", group,
			)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":"resource_exhausted","message":"rate limit exceeded, try again later"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getLimit возвращает лимит и интервал для группы сервисов из конфига.
func (m *RateLimitMiddleware) getLimit(group string) (int, time.Duration) {
	switch group {
	case "auth":
		return m.cfg.AuthLimit, m.cfg.AuthInterval.Duration
	case "submission":
		return m.cfg.SubmissionLimit, m.cfg.SubmissionInterval.Duration
	case "media":
		return m.cfg.MediaLimit, m.cfg.MediaInterval.Duration
	case "admin":
		return m.cfg.AdminLimit, m.cfg.AdminInterval.Duration
	default:
		return m.cfg.ReadLimit, m.cfg.ReadInterval.Duration
	}
}

// getLimiter возвращает существующий или создаёт новый лимитер.
func (m *RateLimitMiddleware) getLimiter(key string, limit int, interval time.Duration) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rl, exists := m.limiters[key]; exists {
		rl.lastSeen = time.Now()
		return rl.limiter
	}

	// rate.Every(interval/limit) = равномерное распределение запросов
	r := rate.NewLimiter(rate.Every(interval/time.Duration(limit)), limit)
	m.limiters[key] = &rateLimiter{limiter: r, lastSeen: time.Now()}
	return r
}

// cleanupLoop периодически удаляет записи без активности.
// Завершается при отмене контекста (graceful shutdown).
func (m *RateLimitMiddleware) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.CleanupInterval.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("rate limit cleanup loop stopped")
			return
		case <-ticker.C:
			m.mu.Lock()
			expired := 0
			for key, rl := range m.limiters {
				if time.Since(rl.lastSeen) > m.cfg.EntryTTL.Duration {
					delete(m.limiters, key)
					expired++
				}
			}
			m.mu.Unlock()
			if expired > 0 {
				m.logger.Debug("rate limit cleanup", "expired_entries", expired)
			}
		}
	}
}

// classifyService определяет группу по URL-пути Connect RPC.
func classifyService(path string) string {
	switch {
	case strings.Contains(path, "AuthService"):
		return "auth"
	case strings.Contains(path, "SubmissionService"):
		return "submission"
	case strings.Contains(path, "MediaService"):
		return "media"
	case strings.Contains(path, "AdminService"):
		return "admin"
	default:
		return "read"
	}
}
