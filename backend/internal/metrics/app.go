package metrics

import "github.com/prometheus/client_golang/prometheus"

var appInfo *prometheus.GaugeVec

func initAppMetrics() {
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "emh",
			Name:      "app_info",
			Help:      "Информация о приложении (version, commit, started_at). Всегда = 1.",
		},
		[]string{"version", "commit", "started_at"},
	)
	Registry.MustRegister(appInfo)
	// Устанавливаем 1 для текущей версии (значения берём из LDFLAGS при билде)
	appInfo.WithLabelValues("dev", "unknown", "").Set(1)
}
