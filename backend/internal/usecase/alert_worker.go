package usecase

import (
	"context"
	"log/slog"
	"sync"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// AlertSender интерфейс отправки алертов безопасности.
type AlertSender interface {
	SendAlert(ctx context.Context, alert domain.Alert) error
}

// AlertWorker асинхронно отправляет алерты безопасности.
// Использует buffered channel, чтобы не блокировать обработку запросов.
// Паттерн аналогичен AuthAuditWorker.
type AlertWorker struct {
	sender  AlertSender
	logger  *slog.Logger
	queue   chan domain.Alert
	wg      sync.WaitGroup
	stopped bool
	mu      sync.Mutex
}

// NewAlertWorker создаёт воркер алертов.
// queueSize — размер buffered channel.
func NewAlertWorker(
	sender AlertSender,
	logger *slog.Logger,
	queueSize int,
) *AlertWorker {
	return &AlertWorker{
		sender: sender,
		logger: logger,
		queue:  make(chan domain.Alert, queueSize),
	}
}

// Enqueue добавляет алерт в очередь.
// Не блокирует вызывающую горутину. При переполнении алерт теряется.
func (w *AlertWorker) Enqueue(alert domain.Alert) {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	select {
	case w.queue <- alert:
	default:
		// Очередь переполнена — логируем и пропускаем.
		// Алерты не должны блокировать обработку запросов.
		w.logger.Warn("alert queue full, alert dropped",
			"alert_type", string(alert.Type),
		)
	}
}

// Run запускает воркер. Блокируется до отмены контекста.
func (w *AlertWorker) Run(ctx context.Context) {
	w.logger.Info("alert worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("alert worker stopped")
			return
		case alert := <-w.queue:
			w.wg.Add(1)
			go func(a domain.Alert) {
				defer w.wg.Done()
				if w.sender == nil {
					return
				}
				if err := w.sender.SendAlert(ctx, a); err != nil {
					w.logger.Error("alert send failed",
						"error", err,
						"alert_type", string(a.Type),
					)
				}
			}(alert)
		}
	}
}

// Stop gracefully останавливает воркер и ждёт отправки накопленных алертов.
func (w *AlertWorker) Stop() {
	w.mu.Lock()
	w.stopped = true
	w.mu.Unlock()

	close(w.queue)
	w.wg.Wait()
}
