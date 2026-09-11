package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ErrorsTotal считает ошибки по коду и слою.
	// Labels: code (доменный код ошибки), layer (domain/delivery/infra)
	ErrorsTotal *prometheus.CounterVec
)

func initErrorsMetrics() {
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "errors",
			Name:      "total",
			Help:      "Количество ошибок по доменному коду и слою приложения.",
		},
		[]string{"code", "layer"},
	)

	Registry.MustRegister(ErrorsTotal)
}
