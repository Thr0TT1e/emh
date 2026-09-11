// Package metrics — централизованный Prometheus-реестр и HTTP-handler для /metrics.
// Все модули проекта регистрируют свои метрики здесь через MustRegister().
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry — глобальный Prometheus-реестр приложения.
// Содержит стандартные Go-метрики (goroutines, GC, memory) + кастомные EMH-метрики.
var Registry *prometheus.Registry

// Handler возвращает http.Handler для эндпоинта /metrics.
// Используется в server.go: mux.Handle("/metrics", metrics.Handler()).
var Handler http.Handler

func init() {
	Registry = prometheus.NewRegistry()

	// Стандартные Go-метрики: goroutines, GC, memory, threads.
	Registry.MustRegister(collectors.NewGoCollector())
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Инициализируем кастомные метрики (описаны в отдельных файлах пакета).
	initRPCMetrics()
	initLLMMetrics()
	initAuthMetrics()
	initS3Metrics()
	initWorkerMetrics()
	initBusinessMetrics()
	initErrorsMetrics()
	initAppMetrics()

	Handler = promhttp.HandlerFor(Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: false, // Включаем только если нужен OTLP
	})
}
