package usecase

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ContactRateLimitConfig параметры rate limiting для ContactService.
// Лимиты применяются отдельно к IP и к email, в двух окнах: час и сутки.
type ContactRateLimitConfig struct {
	// HourLimit — максимум сообщений в час (на один IP или email).
	HourLimit int
	// DayLimit — максимум сообщений в сутки (на один IP или email).
	DayLimit int
	// EntryTTL — время жизни записи лимитера без активности (защита от утечки памяти).
	EntryTTL time.Duration
	// CleanupInterval — интервал фоновой горутины очистки записей.
	CleanupInterval time.Duration
}

// contactRateEntry счётчики fixed window для одного ключа (IP или email).
type contactRateEntry struct {
	// hourCount — количество запросов в текущем часовом окне.
	hourCount int
	// hourReset — время сброса часового окна.
	hourReset time.Time
	// dayCount — количество запросов в текущем суточном окне.
	dayCount int
	// dayReset — время сброса суточного окна.
	dayReset time.Time
	// lastSeen — время последнего обращения (для TTL-очистки).
	lastSeen time.Time
}

// ContactRateLimiter in-memory rate limiter для ContactService.
// Проверяет лимиты по двум ключам (IP и email) и двум окнам (час и сутки).
// Использует fixed window counter — проще и предсказуемее token bucket
// для низкочастотных лимитов (5/час, 20/сутки).
type ContactRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*contactRateEntry
	cfg     ContactRateLimitConfig
	logger  *slog.Logger
}

// NewContactRateLimiter создаёт rate limiter для ContactService.
func NewContactRateLimiter(cfg ContactRateLimitConfig, logger *slog.Logger) *ContactRateLimiter {
	return &ContactRateLimiter{
		entries: make(map[string]*contactRateEntry),
		cfg:     cfg,
		logger:  logger,
	}
}

// Start запускает фоновую горутину очистки записей.
// Вызывается в main.go, завершается при отмене контекста.
func (rl *ContactRateLimiter) Start(ctx context.Context) {
	go rl.cleanupLoop(ctx)
	rl.logger.Info("contact rate limiter запущен",
		"hour_limit", rl.cfg.HourLimit,
		"day_limit", rl.cfg.DayLimit,
	)
}

// Allow проверяет, разрешён ли запрос для данного ключа.
// Ключ формируется вызывающим кодом: "ip:<адрес>" или "email:<адрес>".
// Если лимит превышен, возвращает domain.ErrContactRateLimited.
// При разрешении инкрементирует счётчики обоих окон.
func (rl *ContactRateLimiter) Allow(key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	entry, exists := rl.entries[key]
	if !exists {
		entry = &contactRateEntry{
			hourReset: now.Add(time.Hour),
			dayReset:  now.Add(24 * time.Hour),
		}
		rl.entries[key] = entry
	}

	// Сброс часового окна при истечении.
	if now.After(entry.hourReset) {
		entry.hourCount = 0
		entry.hourReset = now.Add(time.Hour)
	}

	// Сброс суточного окна при истечении.
	if now.After(entry.dayReset) {
		entry.dayCount = 0
		entry.dayReset = now.Add(24 * time.Hour)
	}

	// Проверка лимитов до инкремента.
	if entry.hourCount >= rl.cfg.HourLimit {
		return domain.ErrContactRateLimited
	}
	if entry.dayCount >= rl.cfg.DayLimit {
		return domain.ErrContactRateLimited
	}

	// Инкремент обоих окон.
	entry.hourCount++
	entry.dayCount++
	entry.lastSeen = now

	return nil
}

// cleanupLoop периодически удаляет записи без активности.
// Завершается при отмене контекста (graceful shutdown).
func (rl *ContactRateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rl.logger.Info("contact rate limiter cleanup loop остановлен")
			return
		case <-ticker.C:
			rl.mu.Lock()
			expired := 0
			for key, entry := range rl.entries {
				if time.Since(entry.lastSeen) > rl.cfg.EntryTTL {
					delete(rl.entries, key)
					expired++
				}
			}
			rl.mu.Unlock()

			if expired > 0 {
				rl.logger.Debug("contact rate limiter cleanup", "expired_entries", expired)
			}
		}
	}
}
