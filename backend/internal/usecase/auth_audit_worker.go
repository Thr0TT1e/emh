package usecase

import (
	"context"
	"log/slog"
	"sync"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// AuthAuditWorker асинхронно записывает записи аудита аутентификации в БД.
// Использует buffered channel, чтобы не блокировать обработку запросов.
type AuthAuditWorker struct {
	repo    repository.AuthAuditRepository
	logger  *slog.Logger
	queue   chan domain.AuthAuditEntry
	wg      sync.WaitGroup
	stopped bool
	mu      sync.Mutex
}

// NewAuthAuditWorker создаёт воркер аудита.
// queueSize — размер buffered channel (например, 1000).
func NewAuthAuditWorker(
	repo repository.AuthAuditRepository,
	logger *slog.Logger,
	queueSize int,
) *AuthAuditWorker {
	return &AuthAuditWorker{
		repo:   repo,
		logger: logger,
		queue:  make(chan domain.AuthAuditEntry, queueSize),
	}
}

// Enqueue добавляет запись аудита в очередь.
// Не блокирует вызывающую горутину. Если очередь переполнена, запись теряется.
func (w *AuthAuditWorker) Enqueue(entry domain.AuthAuditEntry) {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	select {
	case w.queue <- entry:
	default:
		// Очередь переполнена — логируем и пропускаем.
		// Аудит не должен блокировать обработку запросов.
		w.logger.Warn("auth audit queue full, entry dropped",
			"ip", entry.IPAddress,
			"mechanism", string(entry.Mechanism),
			"result", string(entry.Result),
		)
	}
}

// Run запускает воркер. Блокируется до отмены контекста.
func (w *AuthAuditWorker) Run(ctx context.Context) {
	w.logger.Info("auth audit worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("auth audit worker stopped")
			return
		case entry := <-w.queue:
			w.wg.Add(1)
			go func(e domain.AuthAuditEntry) {
				defer w.wg.Done()
				if err := w.repo.Add(ctx, e); err != nil {
					w.logger.Error("auth audit write failed",
						"error", err,
						"ip", e.IPAddress,
						"mechanism", string(e.Mechanism),
					)
				}
			}(entry)
		}
	}
}

// Stop gracefully останавливает воркер и ждёт записи накопленных записей.
func (w *AuthAuditWorker) Stop() {
	w.mu.Lock()
	w.stopped = true
	w.mu.Unlock()

	close(w.queue)
	w.wg.Wait()
}
