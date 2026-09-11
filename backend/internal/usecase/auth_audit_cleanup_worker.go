package usecase

import (
	"context"
	"log/slog"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// AuthAuditCleanupWorker периодически удаляет старые записи аудита аутентификации.
// Паттерн аналогичен LLMLogCleanupWorker.
type AuthAuditCleanupWorker struct {
	repo            repository.AuthAuditRepository
	logger          *slog.Logger
	retentionPeriod time.Duration
	interval        time.Duration
}

// NewAuthAuditCleanupWorker создаёт воркер очистки аудита.
func NewAuthAuditCleanupWorker(
	repo repository.AuthAuditRepository,
	logger *slog.Logger,
	retentionPeriod time.Duration,
	interval time.Duration,
) *AuthAuditCleanupWorker {
	return &AuthAuditCleanupWorker{
		repo:            repo,
		logger:          logger,
		retentionPeriod: retentionPeriod,
		interval:        interval,
	}
}

// Run запускает цикл очистки. Блокируется до отмены контекста.
func (w *AuthAuditCleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("auth audit cleanup worker started",
		"interval", w.interval,
		"retention_period", w.retentionPeriod,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("auth audit cleanup worker stopped gracefully")
			return
		case <-ticker.C:
			w.cleanup(ctx)
		}
	}
}

// cleanup выполняет одну итерацию удаления старых записей.
func (w *AuthAuditCleanupWorker) cleanup(ctx context.Context) {
	start := time.Now()
	defer func() {
		metrics.AuthAuditCleanupDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	olderThan := time.Now().Add(-w.retentionPeriod)

	deleted, err := w.repo.DeleteOlderThan(ctx, olderThan)
	if err != nil {
		w.logger.Error("failed to cleanup old auth audit entries", "error", err)
		return
	}

	if deleted > 0 {
		metrics.AuthAuditDeletedTotal.Add(float64(deleted))
		w.logger.Info("successfully cleaned up old auth audit entries",
			"deleted_count", deleted,
			"older_than", olderThan.Format(time.RFC3339),
			"duration", time.Since(start),
		)
	}
}
