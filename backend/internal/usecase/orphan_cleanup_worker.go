package usecase

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// OrphanCleanupWorker фоновая задача удаления сиротских файлов из S3.
// Периодически сравнивает объекты в хранилище с записями в БД
// и удаляет файлы, не привязанные ни к одной записи.
type OrphanCleanupWorker struct {
	photoRepo    repository.PhotoRepository
	mediaStorage domain.MediaStorage
	logger       *slog.Logger
	interval     time.Duration // периодичность запуска
	gracePeriod  time.Duration // не удалять файлы моложе этого возраста
	dryRun       bool          // если true — только логирует, не удаляет
}

// NewOrphanCleanupWorker создаёт воркер очистки сиротских файлов.
func NewOrphanCleanupWorker(
	photoRepo repository.PhotoRepository,
	mediaStorage domain.MediaStorage,
	logger *slog.Logger,
	interval time.Duration,
	gracePeriod time.Duration,
	dryRun bool,
) *OrphanCleanupWorker {
	return &OrphanCleanupWorker{
		photoRepo:    photoRepo,
		mediaStorage: mediaStorage,
		logger:       logger,
		interval:     interval,
		gracePeriod:  gracePeriod,
		dryRun:       dryRun,
	}
}

// Run запускает цикл очистки. Блокируется до отмены контекста.
func (w *OrphanCleanupWorker) Run(ctx context.Context) {
	w.logger.Info("Запущена очистка осиротевших файлов.",
		"interval", w.interval,
		"grace_period", w.gracePeriod,
		"dry_run", w.dryRun,
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("обработчик артифактов в S3 остановился.")
			return
		case <-ticker.C:
			if err := w.cleanup(ctx); err != nil {
				w.logger.Error("Не удалось завершить очистку от осиротевших файлов.", "error", err)
			}
		}
	}
}

// RunOnce выполняет один проход очистки.
// Используется для ручного запуска или тестов.
func (w *OrphanCleanupWorker) RunOnce(ctx context.Context) error {
	return w.cleanup(ctx)
}

// cleanup выполняет один проход: загрузка knownKeys из БД,
// затем streaming-сканирование S3 по страницам с немедленным удалением сирот.
//
// Память: пропорциональна числу URL в БД + размеру одной страницы S3.
// Не зависит от общего числа объектов в S3.
func (w *OrphanCleanupWorker) cleanup(ctx context.Context) error {
	start := time.Now()
	w.logger.Info("orphan cleanup started")

	// 1. Собираем все URL из БД в in-memory set.
	// Это единственный крупный расход памяти — пропорционален числу записей в photos.
	knownURLs, err := w.photoRepo.ListAllMediaURLs(ctx)
	if err != nil {
		w.logger.Error("orphan cleanup: failed to list media urls from db", "error", err)
		return err
	}

	knownKeys := make(map[string]struct{}, len(knownURLs))
	for _, u := range knownURLs {
		key, err := w.mediaStorage.KeyFromPublicURL(u)
		if err != nil {
			continue // URL не принадлежит нашему бакету
		}
		knownKeys[key] = struct{}{}
	}
	w.logger.Info("orphan cleanup: collected known keys", "count", len(knownKeys))

	// 2. Streaming-сканирование S3 по страницам.
	// Для каждой страницы:
	//   - Находим сирот (не в knownKeys, старше gracePeriod)
	//   - Немедленно удаляем их (если не dry_run)
	//   - Accumulate stats
	const pageSize = 500
	cutoff := time.Now().Add(-w.gracePeriod)

	var (
		totalScanned int
		totalOrphans int
		totalDeleted atomic.Int64
		totalErrors  atomic.Int64
		pagesScanned int
		marker       string
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Warn("orphan cleanup cancelled",
				"scanned", totalScanned,
				"deleted", totalDeleted,
			)
			return ctx.Err()
		default:
		}

		page, nextMarker, err := w.mediaStorage.ListObjectsPaged(ctx, marker, pageSize)
		if err != nil {
			w.logger.Error("orphan cleanup: failed to list objects page",
				"marker", marker,
				"error", err,
			)
			return err
		}

		pagesScanned++
		totalScanned += len(page)

		// Находим сирот на этой странице
		var orphans []domain.ObjectInfo
		for _, obj := range page {
			if _, known := knownKeys[obj.Key]; known {
				continue
			}
			if obj.LastModified.After(cutoff) {
				continue // grace period
			}
			orphans = append(orphans, obj)
		}
		totalOrphans += len(orphans)

		// Dry run: только логируем
		if w.dryRun {
			for _, o := range orphans {
				w.logger.Info("orphan cleanup [DRY RUN]: would delete",
					"key", o.Key, "size", o.Size,
				)
			}
		} else if len(orphans) > 0 {
			// Удаляем параллельно с ограничением конкурентности
			g, gctx := errgroup.WithContext(ctx)
			g.SetLimit(5)
			for _, obj := range orphans {
				obj := obj
				g.Go(func() error {
					if err := w.mediaStorage.Delete(gctx, obj.Key); err != nil {
						totalErrors.Add(1) // атомарный инкремент
						w.logger.Warn("orphan cleanup: failed to delete",
							"key", obj.Key, "error", err,
						)
						return nil
					}
					totalDeleted.Add(1) // атомарный инкремент
					w.logger.Debug("orphan cleanup: deleted",
						"key", obj.Key, "size", obj.Size,
					)
					return nil
				})
			}
			_ = g.Wait()
		}

		// Если next_marker пустой — последняя страница
		if nextMarker == "" {
			break
		}
		marker = nextMarker
	}

	// Метрики
	metrics.OrphanCleanupScannedTotal.Add(float64(totalScanned))
	metrics.OrphanCleanupPagesScannedTotal.Add(float64(pagesScanned))
	metrics.OrphanCleanupDurationSeconds.Observe(time.Since(start).Seconds())

	deletedCount := totalDeleted.Load()
	errorsCount := totalErrors.Load()

	if w.dryRun {
		metrics.OrphanCleanupDeletedTotal.WithLabelValues("dry_run").Add(float64(totalOrphans))
	} else {
		metrics.OrphanCleanupDeletedTotal.WithLabelValues("deleted").Add(float64(deletedCount))
		if errorsCount > 0 {
			metrics.OrphanCleanupDeletedTotal.WithLabelValues("error").Add(float64(errorsCount))
		}
	}

	w.logger.Info("orphan cleanup finished",
		"pages_scanned", pagesScanned,
		"objects_scanned", totalScanned,
		"known_keys", len(knownKeys),
		"orphans_found", totalOrphans,
		"deleted", deletedCount,
		"errors", errorsCount,
		"duration", time.Since(start),
	)

	return nil
}
