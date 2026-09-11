package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// LLMRequestsTotal считает LLM-запросы по провайдеру, модели и статусу.
	// Labels: provider (ollama/openai/anthropic), model, status (success/error/timeout)
	LLMRequestsTotal *prometheus.CounterVec

	// LLMDurationSeconds измеряет время ответа LLM.
	// Labels: provider, model
	LLMDurationSeconds *prometheus.HistogramVec

	// LLMParseErrorsTotal считает ошибки парсинга JSON от LLM.
	// Labels: error_type (invalid_json/missing_fields/unknown_format)
	LLMParseErrorsTotal *prometheus.CounterVec
)

func initLLMMetrics() {
	LLMRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "llm",
			Name:      "requests_total",
			Help:      "Общее количество LLM-запросов по провайдеру, модели и статусу.",
		},
		[]string{"provider", "model", "status"},
	)

	LLMDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "emh",
			Subsystem: "llm",
			Name:      "duration_seconds",
			Help:      "Время ответа LLM в секундах.",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 20, 30, 60, 120},
		},
		[]string{"provider", "model"},
	)

	LLMParseErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "llm",
			Name:      "parse_errors_total",
			Help:      "Количество ошибок парсинга JSON от LLM по типу.",
		},
		[]string{"error_type"},
	)

	Registry.MustRegister(
		LLMRequestsTotal,
		LLMDurationSeconds,
		LLMParseErrorsTotal,
	)
}
