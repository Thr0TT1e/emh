package usecase

import (
	"context"
	"log/slog"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type LLMLogCleanupWorker struct {
	logRepo         repository.LLMExtractionLogRepository
	logger          *slog.Logger
	retentionPeriod time.Duration
	interval        time.Duration
}

func NewLLMLogCleanupWorker(
	logRepo repository.LLMExtractionLogRepository,
	logger *slog.Logger,
	retentionPeriod time.Duration,
	interval time.Duration,
) *LLMLogCleanupWorker {
	return &LLMLogCleanupWorker{
		logRepo:         logRepo,
		logger:          logger,
		retentionPeriod: retentionPeriod,
		interval:        interval,
	}
}

func (w *LLMLogCleanupWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("LLM log cleanup worker started",
		"interval", w.interval,
		"retention_period", w.retentionPeriod,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("LLM log cleanup worker stopped gracefully")
			return
		case <-ticker.C:
			w.cleanup(ctx)
		}
	}
}

func (w *LLMLogCleanupWorker) cleanup(ctx context.Context) {
	olderThan := time.Now().Add(-w.retentionPeriod)

	deleted, err := w.logRepo.DeleteOlderThan(ctx, olderThan)
	if err != nil {
		w.logger.Error("failed to cleanup old LLM extraction logs", "error", err)
		return
	}

	if deleted > 0 {
		w.logger.Info("successfully cleaned up old LLM extraction logs",
			"deleted_count", deleted,
			"older_than", olderThan.Format(time.RFC3339),
		)
	}
}
