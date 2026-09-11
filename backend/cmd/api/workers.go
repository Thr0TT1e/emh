package main

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// startWorkers запускает все фоновые воркеры с автоперезапуском.
func startWorkers(ctx context.Context, cfg *config.Config, logger *slog.Logger, deps *Dependencies) {
	// Thumbnail worker
	runWorkerWithRecovery(ctx, logger, "thumbnail", deps.thumbnailWorker.Run, 5*time.Second)
	runWorkerWithRecovery(ctx, logger, "thumbnail_rescan", deps.thumbnailWorker.RescanLoop, 60*time.Second)

	// Orphan cleanup
	if cfg.OrphanCleanup.Enabled {
		runWorkerWithRecovery(ctx, logger, "orphan_cleanup", deps.orphanWorker.Run, 60*time.Second)
	}

	// Auth audit
	if cfg.AuthAudit.Enabled {
		runWorkerWithRecovery(ctx, logger, "auth_audit", deps.authAuditWorker.Run, 5*time.Second)
	}

	// Alert worker
	if cfg.AuthAlerts.Enabled {
		runWorkerWithRecovery(ctx, logger, "alert", deps.alertWorker.Run, 5*time.Second)
	}

	// Auth audit
	if cfg.AuthAudit.Enabled {
		runWorkerWithRecovery(ctx, logger, "auth_audit", deps.authAuditWorker.Run, 5*time.Second)
		// Auth audit cleanup
		runWorkerWithRecovery(ctx, logger, "auth_audit_cleanup", deps.authAuditCleanupWorker.Run, 60*time.Second)
	}

	// Submission email worker
	if deps.submissionEmailWorker != nil {
		runWorkerWithRecovery(ctx, logger, "submission_email", deps.submissionEmailWorker.Run, 5*time.Second)
	}

	// LLM log cleanup
	runWorkerWithRecovery(ctx, logger, "llm_log_cleanup", deps.llmLogCleanupWorker.Run, 60*time.Second)

	// Startup backfill для thumbnail (асинхронно)
	go deps.thumbnailWorker.StartupBackfill(ctx)

	logger.Info("Фоновые воркеры запущены")
}

// runWorkerWithRecovery запускает воркер с автоперезапуском при panic.
func runWorkerWithRecovery(
	ctx context.Context,
	logger *slog.Logger,
	name string,
	run func(context.Context),
	restartDelay time.Duration,
) {
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						// метрика panic
						metrics.WorkerPanicsTotal.WithLabelValues(name).Inc()
						logger.Error("worker panicked, will restart",
							"worker", name,
							"panic", r,
							"stack", string(debug.Stack()),
						)
					}
				}()
				run(ctx)
			}()

			select {
			case <-ctx.Done():
				logger.Info("worker stopped", "worker", name)
				return
			case <-time.After(restartDelay):
				// метрика перезапуска
				metrics.WorkerRestartsTotal.WithLabelValues(name).Inc()
				logger.Info("restarting worker", "worker", name)
			}
		}
	}()
}
