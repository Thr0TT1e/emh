package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// SubmissionTotal считает заявки по статусу.
	// Labels: status (created/approved/rejected)
	SubmissionTotal *prometheus.CounterVec

	// SMTPSendTotal считает попытки отправки email.
	// Labels: status (success/failed/skipped)
	SMTPSendTotal *prometheus.CounterVec

	// ContactMessagesTotal считает сообщения обратной связи.
	// Labels: status (created/rate_limited/duplicate)
	ContactMessagesTotal *prometheus.CounterVec

	// ContactRateLimitedTotal считает срабатывания rate-limit для контактов.
	// Labels: reason (ip/email)
	ContactRateLimitedTotal *prometheus.CounterVec

	// SMTPRetriesTotal считает повторные попытки отправки email.
	SMTPRetriesTotal prometheus.Counter

	// SMTPDurationSeconds измеряет длительность отправки (включая ретраи).
	SMTPDurationSeconds prometheus.Histogram

	// ContactEmailStatusUpdateFailuresTotal считает неудачные обновления email_status.
	// Рост метрики = проблемы с доступностью БД или crash сервера.
	ContactEmailStatusUpdateFailuresTotal prometheus.Counter
	LLMExtractionTextTooLongTotal         prometheus.Counter
)

func initBusinessMetrics() {
	SubmissionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "submission",
			Name:      "total",
			Help:      "Количество заявок по статусу.",
		},
		[]string{"status"},
	)

	SMTPSendTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "smtp",
			Name:      "send_total",
			Help:      "Количество попыток отправки email по статусу.",
		},
		[]string{"status"},
	)

	ContactMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "contact",
			Name:      "messages_total",
			Help:      "Количество сообщений обратной связи по статусу.",
		},
		[]string{"status"},
	)

	ContactRateLimitedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "contact",
			Name:      "rate_limited_total",
			Help:      "Количество срабатываний rate-limit для контактов по причине.",
		},
		[]string{"reason"},
	)

	SMTPRetriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "smtp",
		Name:      "retries_total",
		Help:      "Количество повторных попыток отправки email при временных ошибках.",
	})

	SMTPDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "emh",
		Subsystem: "smtp",
		Name:      "duration_seconds",
		Help:      "Длительность отправки email (включая все ретраи).",
		Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
	})

	ContactEmailStatusUpdateFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "emh",
		Subsystem: "contact",
		Name:      "email_status_update_failures_total",
		Help:      "Количество неудачных обновлений email_status после отправки SMTP.",
	})

	LLMExtractionTextTooLongTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "emh_llm_extraction_text_too_long_total",
			Help: "Количество запросов с превышением лимита длины текста",
		})

	Registry.MustRegister(
		SubmissionTotal,
		SMTPSendTotal,
		ContactMessagesTotal,
		ContactRateLimitedTotal,
		SMTPRetriesTotal,
		SMTPDurationSeconds,
		ContactEmailStatusUpdateFailuresTotal,
		LLMExtractionTextTooLongTotal,
	)
}
