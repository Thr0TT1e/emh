package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// S3OperationsTotal считает S3-операции по типу и статусу.
	// Labels: operation (presign/upload/download/delete/list/ensure_bucket), status (success/error/retry_exhausted)
	S3OperationsTotal *prometheus.CounterVec

	// S3DurationSeconds измеряет длительность S3-операций.
	// Labels: operation
	S3DurationSeconds *prometheus.HistogramVec

	// S3RetriesTotal считает повторные попытки (только при retry, не при успехе с первой).
	// Labels: operation
	S3RetriesTotal *prometheus.CounterVec

	// S3BytesTransferred считает переданные байты по операции и направлению.
	// Labels: operation (upload/download), direction (in/out)
	S3BytesTransferred *prometheus.CounterVec
)

func initS3Metrics() {
	S3OperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "s3",
			Name:      "operations_total",
			Help:      "Количество S3-операций по типу и статусу.",
		},
		[]string{"operation", "status"},
	)

	S3DurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "s3",
			Name:      "duration_seconds",
			Help:      "Длительность S3-операций (включая retry).",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"operation"},
	)

	S3RetriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "s3",
			Name:      "retries_total",
			Help:      "Количество повторных попыток S3-операций.",
		},
		[]string{"operation"},
	)

	S3BytesTransferred = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "s3",
			Name:      "bytes_transferred_total",
			Help:      "Объём переданных байт по операции и направлению.",
		},
		[]string{"operation", "direction"},
	)

	Registry.MustRegister(
		S3OperationsTotal,
		S3DurationSeconds,
		S3RetriesTotal,
		S3BytesTransferred,
	)
}
