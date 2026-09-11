package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// SubmissionEmailSentTotal счётчик отправленных email-уведомлений заявителям.
	// Labels: status (approved, rejected, failed).
	SubmissionEmailSentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "emh",
			Subsystem: "submission_email",
			Name:      "sent_total",
			Help:      "Total number of submission notification emails sent to submitters",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(SubmissionEmailSentTotal)
}
