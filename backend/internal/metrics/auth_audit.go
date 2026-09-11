package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// AuthAuditDeletedTotal счётчик удалённых записей аудита аутентификации.
	AuthAuditDeletedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "auth_audit",
		Name:      "deleted_total",
		Help:      "Total number of auth audit log entries deleted by cleanup worker",
	})

	// AuthAuditCleanupDurationSeconds гистограмма длительности операций очистки.
	AuthAuditCleanupDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "emh",
		Subsystem: "auth_audit",
		Name:      "cleanup_duration_seconds",
		Help:      "Duration of auth audit log cleanup operations in seconds",
		Buckets:   prometheus.DefBuckets,
	})
)

func init() {
	prometheus.MustRegister(AuthAuditDeletedTotal)
	prometheus.MustRegister(AuthAuditCleanupDurationSeconds)
}
