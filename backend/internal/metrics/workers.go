package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ─── Thumbnail Worker ──────────────────────────────────────

	// ThumbnailQueueSize текущий размер очереди задач (Gauge).
	// Растёт при Enqueue, уменьшается при processTask.
	ThumbnailQueueSize prometheus.Gauge

	// ThumbnailProcessedTotal считает обработанные задачи по статусу.
	// Labels: status (success/error/skipped)
	ThumbnailProcessedTotal *prometheus.CounterVec

	// ThumbnailProcessingDuration измеряет время обработки по этапам.
	// Labels: step (download/decode/crop/upload_original/upload_thumbnail/db_update/total)
	ThumbnailProcessingDuration *prometheus.HistogramVec

	// ThumbnailSkippedTotal считает пропущенные задачи по причине.
	// Labels: reason (oversized/dimensions/megapixels/queue_full/cancelled/download_error)
	ThumbnailSkippedTotal *prometheus.CounterVec

	// ─── Orphan Cleanup Worker ────────────────────────────────

	// OrphanCleanupScannedTotal считает просканированные объекты в S3.
	OrphanCleanupScannedTotal prometheus.Counter

	// OrphanCleanupDeletedTotal считает удалённые/пропущенные сироты.
	// Labels: status (deleted/dry_run/error)
	OrphanCleanupDeletedTotal *prometheus.CounterVec

	// OrphanCleanupDurationSeconds измеряет длительность полного цикла очистки.
	OrphanCleanupDurationSeconds prometheus.Histogram

	// ─── Refresh Token Cleanup Worker ─────────────────────────

	// RefreshTokensExpiredTotal считает удалённые истёкшие refresh-токены.
	RefreshTokensExpiredTotal prometheus.Counter

	// ─── Общие worker-метрики ─────────────────────────────────

	// WorkerRestartsTotal считает перезапуски воркеров после panic.
	// Labels: worker (thumbnail/orphan_cleanup/auth_audit/refresh_token_cleanup)
	WorkerRestartsTotal *prometheus.CounterVec

	// WorkerPanicsTotal считает panic'и воркеров.
	// Labels: worker
	WorkerPanicsTotal *prometheus.CounterVec

	// EnqueueRetriesTotal считает retry-попытки постановки задач в очередь.
	// Срабатывает при кратковременном переполнении очереди.
	EnqueueRetriesTotal prometheus.Counter

	// EnqueueFailuresTotal считает финальные неудачи постановки в очередь.
	// Задача будет подхвачена RescanLoop.
	EnqueueFailuresTotal prometheus.Counter

	// RescanTriggeredTotal считает задачи, подхваченные rescan loop.
	// Рост метрики = индикатор проблем с очередью или воркером.
	RescanTriggeredTotal prometheus.Counter

	// OrphanCleanupPagesScannedTotal считает количество обработанных страниц S3.
	// Полезно для мониторинга: если растёт при неизменном числе объектов —
	// возможно S3 возвращает мало объектов на страницу (проблема с лимитом).
	OrphanCleanupPagesScannedTotal prometheus.Counter

	// StartupBackfillProcessedTotal считает фото, обработанные при cold start.
	// Рост при каждом старте — индикатор незавершённых задач с прошлой сессии.
	StartupBackfillProcessedTotal prometheus.Counter

	// StartupBackfillDurationSeconds измеряет длительность полного цикла backfill.
	StartupBackfillDurationSeconds prometheus.Histogram

	// ThumbnailOriginalDeleteFailuresTotal считает неудачные удаления оригинального файла
	// после конвертации в WebP. Рост метрики = проблемы с доступностью S3.
	ThumbnailOriginalDeleteFailuresTotal prometheus.Counter
)

func initWorkerMetrics() {
	// Thumbnail
	ThumbnailQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "queue_size",
		Help:      "Текущий размер очереди задач генерации превью.",
	})

	ThumbnailProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "thumbnail",
			Name:      "processed_total",
			Help:      "Количество обработанных задач по статусу.",
		},
		[]string{"status"},
	)

	ThumbnailProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "thumbnail",
			Name:      "processing_duration_seconds",
			Help:      "Длительность этапов обработки thumbnail.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"step"},
	)

	ThumbnailSkippedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "thumbnail",
			Name:      "skipped_total",
			Help:      "Количество пропущенных задач по причине.",
		},
		[]string{"reason"},
	)

	// Orphan Cleanup
	OrphanCleanupScannedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "orphan_cleanup",
		Name:      "scanned_total",
		Help:      "Общее количество просканированных объектов в S3.",
	})

	OrphanCleanupDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "orphan_cleanup",
			Name:      "deleted_total",
			Help:      "Количество обработанных сирот по статусу.",
		},
		[]string{"status"},
	)

	OrphanCleanupDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "emh",
		Subsystem: "orphan_cleanup",
		Name:      "duration_seconds",
		Help:      "Длительность полного цикла очистки сирот.",
		Buckets:   []float64{1, 5, 10, 30, 60, 300, 600},
	})

	// Refresh Token Cleanup
	RefreshTokensExpiredTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "refresh_token",
		Name:      "expired_total",
		Help:      "Количество удалённых истёкших refresh-токенов.",
	})

	// Общие worker-метрики
	WorkerRestartsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "worker",
			Name:      "restarts_total",
			Help:      "Количество перезапусков воркеров после panic.",
		},
		[]string{"worker"},
	)

	WorkerPanicsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "worker",
			Name:      "panics_total",
			Help:      "Количество panic'ов воркеров.",
		},
		[]string{"worker"},
	)

	EnqueueRetriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "enqueue_retries_total",
		Help:      "Количество retry-попыток постановки задач в очередь при переполнении.",
	})

	EnqueueFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "enqueue_failures_total",
		Help:      "Количество финальных неудач постановки задач (подхватываются rescan).",
	})

	RescanTriggeredTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "rescan_triggered_total",
		Help:      "Количество задач, подхваченных периодическим rescan loop.",
	})

	OrphanCleanupPagesScannedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "orphan_cleanup",
		Name:      "pages_scanned_total",
		Help:      "Количество обработанных страниц S3 при orphan cleanup.",
	})

	StartupBackfillProcessedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "startup_backfill_processed_total",
		Help:      "Количество фото, поставленных в очередь при cold start.",
	})

	StartupBackfillDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "startup_backfill_duration_seconds",
		Help:      "Длительность полного цикла StartupBackfill.",
		Buckets:   []float64{1, 5, 10, 30, 60, 300, 600, 1800},
	})

	ThumbnailOriginalDeleteFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "thumbnail",
		Name:      "original_delete_failures_total",
		Help:      "Количество неудачных удалений оригинального файла после конвертации в WebP.",
	})

	Registry.MustRegister(
		ThumbnailQueueSize,
		ThumbnailProcessedTotal,
		ThumbnailProcessingDuration,
		ThumbnailSkippedTotal,
		OrphanCleanupScannedTotal,
		OrphanCleanupDeletedTotal,
		OrphanCleanupDurationSeconds,
		RefreshTokensExpiredTotal,
		WorkerRestartsTotal,
		WorkerPanicsTotal,
		EnqueueRetriesTotal,
		EnqueueFailuresTotal,
		RescanTriggeredTotal,
		OrphanCleanupPagesScannedTotal,
		StartupBackfillProcessedTotal,
		StartupBackfillDurationSeconds,
		ThumbnailOriginalDeleteFailuresTotal,
	)
}
