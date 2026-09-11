package usecase

import (
	"context"
	"log/slog"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// RefreshTokenCleanupWorker периодически удаляет истёкшие refresh-токены.
type RefreshTokenCleanupWorker struct {
	repo     repository.RefreshTokenRepository
	logger   *slog.Logger
	interval time.Duration
}

func NewRefreshTokenCleanupWorker(
	repo repository.RefreshTokenRepository,
	logger *slog.Logger,
	interval time.Duration,
) *RefreshTokenCleanupWorker {
	return &RefreshTokenCleanupWorker{
		repo:     repo,
		logger:   logger,
		interval: interval,
	}
}

func (w *RefreshTokenCleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.logger.Info("refresh token cleanup worker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("refresh token cleanup worker stopped")
			return
		case <-ticker.C:
			deleted, err := w.repo.DeleteExpired(ctx)
			if err != nil {
				w.logger.Error("failed to delete expired refresh tokens", "error", err)
				continue
			}
			if deleted > 0 {
				metrics.RefreshTokensExpiredTotal.Add(float64(deleted))
				w.logger.Info("expired refresh tokens deleted", "count", deleted)
			}
		}
	}
}
