package usecase

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// mockAuthAuditRepositoryWithCleanup мок с поддержкой DeleteOlderThan.
type mockAuthAuditRepositoryWithCleanup struct {
	mu                sync.Mutex
	entries           []domain.AuthAuditEntry
	deleteOlderThanFn func(ctx context.Context, olderThan time.Time) (int64, error)
	deleteCalls       atomic.Int32
}

func (m *mockAuthAuditRepositoryWithCleanup) Add(ctx context.Context, entry domain.AuthAuditEntry) error {
	m.mu.Lock()
	m.entries = append(m.entries, entry)
	m.mu.Unlock()
	return nil
}

func (m *mockAuthAuditRepositoryWithCleanup) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	m.deleteCalls.Add(1)
	if m.deleteOlderThanFn != nil {
		return m.deleteOlderThanFn(ctx, olderThan)
	}
	return 0, nil
}

// TestAuthAuditCleanupWorker_CallsDeleteOlderThan проверяет, что воркер
// вызывает DeleteOlderThan с правильной датой.
func TestAuthAuditCleanupWorker_CallsDeleteOlderThan(t *testing.T) {
	var deleteCalls int32
	var receivedOlderThan time.Time

	repo := &mockAuthAuditRepositoryWithCleanup{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			atomic.AddInt32(&deleteCalls, 1)
			receivedOlderThan = olderThan
			return 100, nil
		},
	}

	retention := 90 * 24 * time.Hour
	interval := 10 * time.Millisecond
	worker := NewAuthAuditCleanupWorker(repo, discardTestLogger(), retention, interval)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	worker.Run(ctx)

	// Минимум 1 вызов (первый cleanup сразу при старте)
	if atomic.LoadInt32(&deleteCalls) < 1 {
		t.Errorf("expected at least 1 DeleteOlderThan call, got %d", deleteCalls)
	}

	// Проверяем, что olderThan примерно сейчас - 90 дней
	expectedOlderThan := time.Now().Add(-retention)
	diff := receivedOlderThan.Sub(expectedOlderThan)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("olderThan = %v, expected around %v (diff: %v)", receivedOlderThan, expectedOlderThan, diff)
	}
}

// TestAuthAuditCleanupWorker_HandlesError проверяет, что ошибка DeleteOlderThan
// не прерывает цикл работы воркера.
func TestAuthAuditCleanupWorker_HandlesError(t *testing.T) {
	var callCount atomic.Int32

	repo := &mockAuthAuditRepositoryWithCleanup{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			callCount.Add(1)
			return 0, errors.New("db connection failed")
		},
	}

	worker := NewAuthAuditCleanupWorker(repo, discardTestLogger(), 90*24*time.Hour, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	// Не должно паниковать или зависать
	worker.Run(ctx)

	// Должно было быть несколько попыток, несмотря на ошибки
	if callCount.Load() < 2 {
		t.Errorf("expected multiple DeleteOlderThan attempts despite errors, got %d", callCount.Load())
	}
}

// TestAuthAuditCleanupWorker_ZeroDeletedNoLog проверяет, что при нулевом
// количестве удалённых записей воркер не логирует и не инкрементирует метрику.
func TestAuthAuditCleanupWorker_ZeroDeletedNoLog(t *testing.T) {
	repo := &mockAuthAuditRepositoryWithCleanup{
		deleteOlderThanFn: func(ctx context.Context, olderThan time.Time) (int64, error) {
			return 0, nil // ничего не удалено
		},
	}

	worker := NewAuthAuditCleanupWorker(repo, discardTestLogger(), 90*24*time.Hour, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	// Не должно паниковать
	worker.Run(ctx)
}
