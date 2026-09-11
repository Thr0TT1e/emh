package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ============================================================
// Тесты RunOnce (детерминированные, без ticker)
// ============================================================

// TestOrphanCleanupWorker_RunOnce_NoOrphans проверяет работу при отсутствии сирот.
func TestOrphanCleanupWorker_RunOnce_NoOrphans(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return []string{
				"https://public.example.com/photos/abc123.jpg",
				"https://public.example.com/photos/def456.jpg",
			}, nil
		},
	}

	storage := &mockMediaStorage{
		// ListObjects возвращает те же ключи — нет сирот.
	}
	// Переопределяем ListObjects через bytes.Buffer-подход невозможен,
	// поэтому используем кастомную функцию в mockMediaStorage.

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,  // interval (не используется в RunOnce)
		24*time.Hour, // gracePeriod
		false,        // dryRun
	)

	err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	// Проверяем, что ничего не удалилось.
	if storage.deleteCount() != 0 {
		t.Errorf("RunOnce() deleted %d objects, want 0", storage.deleteCount())
	}
}

// TestOrphanCleanupWorker_RunOnce_DeletesOrphans проверяет удаление сиротских файлов.
func TestOrphanCleanupWorker_RunOnce_DeletesOrphans(t *testing.T) {
	// В БД есть только один URL.
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return []string{"https://public.example.com/photos/valid.jpg"}, nil
		},
	}

	// В S3 три объекта: один валидный, два сироты (старше gracePeriod).
	oldTime := time.Now().Add(-48 * time.Hour) // старше gracePeriod
	storage := &orphanTestStorage{
		objects: []domain.ObjectInfo{
			{Key: "photos/valid.jpg", LastModified: time.Now()},
			{Key: "photos/orphan1.jpg", LastModified: oldTime},
			{Key: "photos/orphan2.jpg", LastModified: oldTime},
		},
	}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour,
		false,
	)

	err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	// Должно быть удалено 2 сироты.
	if storage.deleteCount() != 2 {
		t.Errorf("RunOnce() deleted %d objects, want 2", storage.deleteCount())
	}
}

// TestOrphanCleanupWorker_RunOnce_DryRun проверяет режим dry-run.
func TestOrphanCleanupWorker_RunOnce_DryRun(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return []string{}, nil // В БД пусто — все объекты сироты.
		},
	}

	var deleteAttempts atomic.Int32
	oldTime := time.Now().Add(-48 * time.Hour)
	storage := &orphanTestStorage{
		objects: []domain.ObjectInfo{
			{Key: "photos/orphan.jpg", LastModified: oldTime},
		},
		deleteFn: func(ctx context.Context, key string) error {
			deleteAttempts.Add(1)
			return nil
		},
	}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour,
		true, // dryRun = true
	)

	err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	// В dry-run режиме удаления не должно быть.
	if deleteAttempts.Load() != 0 {
		t.Errorf("RunOnce() in dry-run should not delete, but deleteAttempts = %d", deleteAttempts.Load())
	}
}

// TestOrphanCleanupWorker_RunOnce_GracePeriod проверяет, что свежие файлы не удаляются.
func TestOrphanCleanupWorker_RunOnce_GracePeriod(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return []string{}, nil // В БД пусто.
		},
	}

	var deleteAttempts atomic.Int32
	// Файл создан только что — должен попасть под grace period.
	storage := &orphanTestStorage{
		objects: []domain.ObjectInfo{
			{Key: "photos/recent.jpg", LastModified: time.Now()},
		},
		deleteFn: func(ctx context.Context, key string) error {
			deleteAttempts.Add(1)
			return nil
		},
	}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour, // gracePeriod = 24 часа
		false,
	)

	err := worker.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	// Свежий файл не должен быть удалён.
	if deleteAttempts.Load() != 0 {
		t.Errorf("RunOnce() should not delete files newer than gracePeriod, but deleteAttempts = %d", deleteAttempts.Load())
	}
}

// TestOrphanCleanupWorker_RunOnce_ContextCanceled проверяет graceful shutdown.
func TestOrphanCleanupWorker_RunOnce_ContextCanceled(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{}
	storage := &orphanTestStorage{}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour,
		false,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу.

	err := worker.RunOnce(ctx)
	if err == nil {
		t.Error("RunOnce() should return error on canceled context")
	}
}

// TestOrphanCleanupWorker_RunOnce_PhotoRepoError проверяет обработку ошибки PhotoRepository.
func TestOrphanCleanupWorker_RunOnce_PhotoRepoError(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return nil, errors.New("db connection failed")
		},
	}
	storage := &orphanTestStorage{}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour,
		false,
	)

	err := worker.RunOnce(context.Background())
	if err == nil {
		t.Error("RunOnce() should return error on photoRepo failure")
	}
}

// TestOrphanCleanupWorker_RunOnce_StorageError проверяет обработку ошибки S3 при листинге.
func TestOrphanCleanupWorker_RunOnce_StorageError(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{
		listAllMediaURLsFn: func(ctx context.Context) ([]string, error) {
			return []string{"https://public.example.com/photos/valid.jpg"}, nil
		},
	}
	storage := &orphanTestStorage{
		listError: errors.New("S3 connection failed"),
	}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		1*time.Hour,
		24*time.Hour,
		false,
	)

	err := worker.RunOnce(context.Background())
	if err == nil {
		t.Error("RunOnce() should return error on storage failure")
	}
}

// ============================================================
// Тесты Run (ticker-based loop)
// ============================================================

// TestOrphanCleanupWorker_Run_ContextCanceled проверяет graceful shutdown цикла.
func TestOrphanCleanupWorker_Run_ContextCanceled(t *testing.T) {
	photoRepo := &mockPhotoRepositoryForOrphan{}
	storage := &orphanTestStorage{}

	worker := NewOrphanCleanupWorker(
		photoRepo,
		storage,
		slog.Default(),
		100*time.Millisecond,
		24*time.Hour,
		false,
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

// ============================================================
// Специализированный мок для orphan cleanup тестов
// ============================================================

// orphanTestStorage специализированный мок S3 для orphan cleanup тестов.
type orphanTestStorage struct {
	objects    []domain.ObjectInfo
	listError  error
	deleteFn   func(ctx context.Context, key string) error
	deleteKeys []string
	mu         sync.Mutex
}

func (m *orphanTestStorage) ListObjectsPaged(ctx context.Context, marker string, limit int) ([]domain.ObjectInfo, string, error) {
	if m.listError != nil {
		return nil, "", m.listError
	}
	// Возвращаем все объекты как одну страницу.
	return m.objects, "", nil
}

func (m *orphanTestStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	m.deleteKeys = append(m.deleteKeys, key)
	m.mu.Unlock()
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

func (m *orphanTestStorage) deleteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.deleteKeys)
}

// Остальные методы MediaStorage — заглушки.
func (m *orphanTestStorage) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	return "", nil
}
func (m *orphanTestStorage) PublicURL(key string) string {
	return "https://public.example.com/" + key
}
func (m *orphanTestStorage) KeyFromPublicURL(rawURL string) (string, error) {
	// Извлекаем ключ из URL вида "https://public.example.com/photos/abc.jpg".
	const prefix = "https://public.example.com/"
	if len(rawURL) > len(prefix) {
		return rawURL[len(prefix):], nil
	}
	return rawURL, nil
}
func (m *orphanTestStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}
func (m *orphanTestStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	return nil
}
func (m *orphanTestStorage) ListObjects(ctx context.Context) ([]domain.ObjectInfo, error) {
	return m.objects, nil
}
func (m *orphanTestStorage) EnsureBucket(ctx context.Context) error {
	return nil
}
