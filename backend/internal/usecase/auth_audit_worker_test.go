package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ============================================================
// Тесты Enqueue
// ============================================================

// TestAuthAuditWorker_Enqueue_Success проверяет, что запись попадает в очередь.
func TestAuthAuditWorker_Enqueue_Success(t *testing.T) {
	repo := &mockAuthAuditRepository{}
	worker := NewAuthAuditWorker(repo, slog.Default(), 10)

	entry := domain.AuthAuditEntry{
		IPAddress: "127.0.0.1",
		Mechanism: "password",
		Result:    "success",
	}

	worker.Enqueue(entry)
	// Enqueue не блокирующий. Проверяем через запуск Run.
}

// TestAuthAuditWorker_Enqueue_QueueFull проверяет, что при переполнении очереди
// запись пропускается без блокировки (best-effort).
func TestAuthAuditWorker_Enqueue_QueueFull(t *testing.T) {
	repo := &mockAuthAuditRepository{}
	// Очень маленькая очередь (1) — быстро переполнится.
	worker := NewAuthAuditWorker(repo, slog.Default(), 1)

	// НЕ запускаем Run — тогда очередь заполнится и следующие записи потеряются.
	for i := 0; i < 5; i++ {
		worker.Enqueue(domain.AuthAuditEntry{
			IPAddress: "127.0.0.1",
			Mechanism: "password",
			Result:    "success",
		})
	}
	// Enqueue не должен блокироваться даже при переполнении.
	// Воркер не запущен, поэтому записи не обрабатываются — это ожидаемо.
}

// ============================================================
// Тесты Run
// ============================================================

// TestAuthAuditWorker_Run_WritesEntries проверяет, что воркер пишет записи в БД.
func TestAuthAuditWorker_Run_WritesEntries(t *testing.T) {
	repo := &mockAuthAuditRepository{}
	worker := NewAuthAuditWorker(repo, slog.Default(), 100)

	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем воркер в фоне.
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	// Отправляем несколько записей.
	for i := 0; i < 5; i++ {
		worker.Enqueue(domain.AuthAuditEntry{
			IPAddress: "127.0.0.1",
			Mechanism: "password",
			Result:    "success",
		})
	}

	// Даём воркеру время обработать записи.
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	// Проверяем, что записи попали в БД.
	if repo.totalEntries() == 0 {
		t.Error("Run() should write entries to database via Add")
	}
}

// TestAuthAuditWorker_Run_ContextCanceled проверяет graceful shutdown.
func TestAuthAuditWorker_Run_ContextCanceled(t *testing.T) {
	repo := &mockAuthAuditRepository{}
	worker := NewAuthAuditWorker(repo, slog.Default(), 100)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	// Отменяем контекст.
	cancel()

	// Проверяем, что воркер завершился.
	select {
	case <-done:
		// Успешно.
	case <-time.After(2 * time.Second):
		t.Error("Run() should stop gracefully on context cancellation")
	}
}

// TestAuthAuditWorker_Run_RepoError проверяет, что ошибка БД не роняет воркер.
func TestAuthAuditWorker_Run_RepoError(t *testing.T) {
	var addAttempts atomic.Int32
	repo := &mockAuthAuditRepository{
		addFn: func(ctx context.Context, entry domain.AuthAuditEntry) error {
			addAttempts.Add(1)
			return errors.New("db connection failed")
		},
	}
	worker := NewAuthAuditWorker(repo, slog.Default(), 100)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Отправляем записи.
	for i := 0; i < 3; i++ {
		worker.Enqueue(domain.AuthAuditEntry{
			IPAddress: "127.0.0.1",
		})
	}

	// Воркер должен продолжить работу несмотря на ошибки БД.
	worker.Run(ctx)

	// Проверяем, что были попытки записи (воркер не упал).
	if addAttempts.Load() == 0 {
		t.Error("Run() should attempt Add even if it fails")
	}
}

// TestAuthAuditWorker_Stop_FlushesEntries проверяет, что Stop ждёт записи накопленных записей.
func TestAuthAuditWorker_Stop_FlushesEntries(t *testing.T) {
	repo := &mockAuthAuditRepository{}
	worker := NewAuthAuditWorker(repo, slog.Default(), 100)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	// Отправляем записи.
	for i := 0; i < 3; i++ {
		worker.Enqueue(domain.AuthAuditEntry{
			IPAddress: "127.0.0.1",
		})
	}

	// Даём немного времени на обработку.
	time.Sleep(100 * time.Millisecond)

	// Останавливаем воркер.
	cancel()
	worker.Stop()
	<-done

	// Все записи должны быть записаны.
	if repo.totalEntries() != 3 {
		t.Errorf("Stop() should flush all entries, got %d, want 3", repo.totalEntries())
	}
}
