package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// AuthAttemptsTotal считает попытки аутентификации по механизму и результату.
	// Labels: mechanism (jwt/api_key/static_key/none), result (success/failure)
	AuthAttemptsTotal *prometheus.CounterVec

	// AuthActiveSessions отслеживает количество активных refresh-токенов по username.
	// Используется для мониторинга одновременных сессий.
	AuthActiveSessions *prometheus.GaugeVec

	// AuthTokensIssuedTotal считает выданные токены по типу.
	// Labels: type (access/refresh)
	AuthTokensIssuedTotal *prometheus.CounterVec

	// AuthBruteForceAlertsTotal считает срабатывания алертов brute-force.
	AuthBruteForceAlertsTotal prometheus.Counter
)

func initAuthMetrics() {
	AuthAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "auth",
			Name:      "attempts_total",
			Help:      "Общее количество попыток аутентификации по механизму и результату.",
		},
		[]string{"mechanism", "result"},
	)

	AuthActiveSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "emh",
			Subsystem: "auth",
			Name:      "active_sessions",
			Help:      "Количество активных refresh-токенов по username.",
		},
		[]string{"username"},
	)

	AuthTokensIssuedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "auth",
			Name:      "tokens_issued_total",
			Help:      "Количество выданных токенов по типу (access/refresh).",
		},
		[]string{"type"},
	)

	AuthBruteForceAlertsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "auth",
			Name:      "brute_force_alerts_total",
			Help:      "Количество срабатываний алертов brute-force по IP.",
		},
	)

	Registry.MustRegister(
		AuthAttemptsTotal,
		AuthActiveSessions,
		AuthTokensIssuedTotal,
		AuthBruteForceAlertsTotal,
	)
}
