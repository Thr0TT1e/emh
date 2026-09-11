package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// Тесты Run
// ============================================================

// TestLLMLogCleanupWorker_Run_Success проверяет успешную очистку старых логов.
func TestLLMLogCleanupWorker_Run_Success(t *testing.T) {
	var deleteCalls atomic.Int32
	var capturedOlderThan time.Time

	repo := &mockLLMExtractionLogRepo{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			deleteCalls.Add(1)
			capturedOlderThan = olderThan
			return 10, nil
		},
	}

	worker := NewLLMLogCleanupWorker(
		repo,
		slog.Default(),
		90*24*time.Hour,      // retention: 90 дней
		100*time.Millisecond, // interval: 100ms для быстрого теста
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	worker.Run(ctx)

	if deleteCalls.Load() == 0 {
		t.Error("Run() should call DeleteOlderThan at least once")
	}

	// Проверяем корректность вычисления olderThan.
	expectedOlderThan := time.Now().Add(-90 * 24 * time.Hour)
	diff := capturedOlderThan.Sub(expectedOlderThan)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Run() olderThan = %v, want ~%v (diff: %v)", capturedOlderThan, expectedOlderThan, diff)
	}
}

// TestLLMLogCleanupWorker_Run_ContextCanceled проверяет graceful shutdown.
func TestLLMLogCleanupWorker_Run_ContextCanceled(t *testing.T) {
	repo := &mockLLMExtractionLogRepo{}
	worker := NewLLMLogCleanupWorker(
		repo,
		slog.Default(),
		90*24*time.Hour,
		100*time.Millisecond,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу.

	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Успешно.
	case <-time.After(500 * time.Millisecond):
		t.Error("Run() should stop gracefully on context cancellation")
	}
}

// TestLLMLogCleanupWorker_Run_RepoError проверяет, что ошибка БД не роняет воркер.
func TestLLMLogCleanupWorker_Run_RepoError(t *testing.T) {
	var callCount atomic.Int32
	repo := &mockLLMExtractionLogRepo{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			callCount.Add(1)
			return 0, errors.New("db connection failed")
		},
	}

	worker := NewLLMLogCleanupWorker(
		repo,
		slog.Default(),
		90*24*time.Hour,
		100*time.Millisecond,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	worker.Run(ctx)

	if callCount.Load() == 0 {
		t.Error("Run() should call DeleteOlderThan even if it fails")
	}
}

// TestLLMLogCleanupWorker_Run_ZeroDeleted проверяет работу при отсутствии старых логов.
func TestLLMLogCleanupWorker_Run_ZeroDeleted(t *testing.T) {
	repo := &mockLLMExtractionLogRepo{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			return 0, nil // Нет старых записей.
		},
	}

	worker := NewLLMLogCleanupWorker(
		repo,
		slog.Default(),
		90*24*time.Hour,
		100*time.Millisecond,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	worker.Run(ctx)
	// Не должно падать.
}

// TestLLMLogCleanupWorker_Run_LargeBatch проверяет обработку большого количества удалений.
func TestLLMLogCleanupWorker_Run_LargeBatch(t *testing.T) {
	repo := &mockLLMExtractionLogRepo{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			return 10000, nil // Большая партия.
		},
	}

	worker := NewLLMLogCleanupWorker(
		repo,
		slog.Default(),
		90*24*time.Hour,
		100*time.Millisecond,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	worker.Run(ctx)
	// Не должно падать при большом количестве удалений.
}
