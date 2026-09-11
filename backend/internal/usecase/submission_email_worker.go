package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
	"codeberg.org/Thr0TT1e/emh/backend/internal/smtp"
)

// SubmissionEmailWorker асинхронно отправляет email-уведомления заявителям.
// Использует buffered channel, чтобы не блокировать обработку запросов модератора.
type SubmissionEmailWorker struct {
	smtpSender smtp.Sender
	logger     *slog.Logger
	queue      chan domain.SubmissionEmailNotification
	cfg        config.SubmissionEmailsConfig
	wg         sync.WaitGroup
	stopped    bool
	mu         sync.Mutex
}

// NewSubmissionEmailWorker создаёт воркер email-уведомлений.
func NewSubmissionEmailWorker(
	smtpSender smtp.Sender,
	logger *slog.Logger,
	cfg config.SubmissionEmailsConfig,
) *SubmissionEmailWorker {
	return &SubmissionEmailWorker{
		smtpSender: smtpSender,
		logger:     logger,
		queue:      make(chan domain.SubmissionEmailNotification, cfg.QueueSize),
		cfg:        cfg,
	}
}

// Enqueue добавляет уведомление в очередь.
// Не блокирует вызывающую горутину. При переполнении уведомление теряется.
func (w *SubmissionEmailWorker) Enqueue(notification domain.SubmissionEmailNotification) {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	select {
	case w.queue <- notification:
	default:
		// Очередь переполнена — логируем и пропускаем.
		w.logger.Warn("submission email queue full, notification dropped",
			"email", notification.Email,
			"submission_id", notification.SubmissionID,
		)
	}
}

// Run запускает воркер. Блокируется до отмены контекста.
//
// Контекст используется ТОЛЬКО для выхода из цикла (graceful shutdown).
// Отправка email идёт через context.Background(), чтобы отмена parent-контекста
// не обрывала in-flight и queued отправки.
func (w *SubmissionEmailWorker) Run(ctx context.Context) {
	w.logger.Info("submission email worker started")

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown: обрабатываем оставшиеся сообщения в очереди
			w.drainQueue()
			w.logger.Info("submission email worker stopped")
			return
		case notification, ok := <-w.queue:
			if !ok {
				// Канал закрыт — выходим (Stop() был вызван)
				return
			}
			w.wg.Add(1)
			go func(n domain.SubmissionEmailNotification) {
				defer w.wg.Done()
				// Используем Background: отправка не зависит от lifecycle воркера.
				// SMTP-слой имеет свои retry и таймауты.
				w.send(context.Background(), n)
			}(notification)
		}
	}
}

// drainQueue обрабатывает все оставшиеся в очереди сообщения после отмены контекста.
// Читает канал до опустошения (default case), запуская goroutine для каждого.
func (w *SubmissionEmailWorker) drainQueue() {
	for {
		select {
		case notification, ok := <-w.queue:
			if !ok {
				// Канал закрыт — ждём завершения in-flight и выходим
				w.wg.Wait()
				return
			}
			w.wg.Add(1)
			go func(n domain.SubmissionEmailNotification) {
				defer w.wg.Done()
				w.send(context.Background(), n)
			}(notification)
		default:
			// Канал пуст — ждём завершения всех запущенных горутин
			w.wg.Wait()
			return
		}
	}
}

// send отправляет одно email-уведомление.
func (w *SubmissionEmailWorker) send(ctx context.Context, n domain.SubmissionEmailNotification) {
	var subject, body, status string

	if n.Decision == domain.ReviewDecisionApprove {
		subject = w.cfg.ApproveSubject
		body = renderEmailTemplate(w.cfg.ApproveTemplate, n)
		status = "approved"
	} else {
		subject = w.cfg.RejectSubject
		body = renderEmailTemplate(w.cfg.RejectTemplate, n)
		status = "rejected"
	}

	msg := smtp.Message{
		Subject: fmt.Sprintf("[EMH] %s", subject),
		Body:    body,
	}

	if err := w.smtpSender.Send(ctx, msg); err != nil {
		metrics.SubmissionEmailSentTotal.WithLabelValues("failed").Inc()
		w.logger.Error("failed to send submission email",
			"email", n.Email,
			"submission_id", n.SubmissionID,
			"decision", n.Decision,
			"error", err,
		)
		return
	}

	metrics.SubmissionEmailSentTotal.WithLabelValues(status).Inc()
	w.logger.Info("submission email sent",
		"email", n.Email,
		"submission_id", n.SubmissionID,
		"decision", n.Decision,
	)
}

// Stop gracefully останавливает воркер.
// Вызывается ПОСЛЕ отмены контекста (которая триггерит drainQueue в Run).
// Закрывает канал и ждёт завершения всех in-flight отправок.
func (w *SubmissionEmailWorker) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	w.mu.Unlock()

	// Закрываем канал — Run() (если ещё работает) выйдет через case ok
	close(w.queue)
	// Ждём завершения всех in-flight отправок
	w.wg.Wait()
}

// renderEmailTemplate подставляет данные в шаблон.
func renderEmailTemplate(template string, n domain.SubmissionEmailNotification) string {
	result := template
	result = strings.ReplaceAll(result, "{{.SubmitterName}}", n.SubmitterName)
	result = strings.ReplaceAll(result, "{{.SubmissionID}}", n.SubmissionID)
	result = strings.ReplaceAll(result, "{{.Timestamp}}", n.Timestamp.Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "{{.ModeratorComment}}", n.ModeratorComment)
	return result
}
