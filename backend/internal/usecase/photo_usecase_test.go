package usecase

import (
	"context"
	"errors"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// --- Хелперы ---

// newPhotoUseCase создаёт PhotoUseCase с моками.
// newPhotoUseCase создаёт PhotoUseCase с моками.
func newPhotoUseCase(
	photoRepo *mockPhotoRepository,
	heroRepo *mockHeroRepository,
	storage *mockMediaStorage,
	worker *mockThumbnailWorker,
) PhotoUseCase {
	return NewPhotoUseCase(photoRepo, heroRepo, storage, worker, discardLogger())
}

// --- Тесты Add ---

// TestPhotoUseCase_Add_EmptyURL проверяет, что пустой URL возвращает ошибку.
func TestPhotoUseCase_Add_EmptyURL(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	params := domain.AddPhotoParams{URL: ""}
	_, err := uc.Add(context.Background(), params)
	if !errors.Is(err, domain.ErrPhotoURLRequired) {
		t.Errorf("expected ErrPhotoURLRequired, got %v", err)
	}
}

// TestPhotoUseCase_Add_HeroNotFound проверяет, что несуществующий герой возвращает ошибку.
func TestPhotoUseCase_Add_HeroNotFound(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return nil, domain.ErrNotFound
		},
	}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	params := domain.AddPhotoParams{HeroID: "nonexistent", URL: "https://example.com/photo.jpg"}
	_, err := uc.Add(context.Background(), params)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestPhotoUseCase_Add_FirstPhotoBecomesMain проверяет автоназначение главного фото.
func TestPhotoUseCase_Add_FirstPhotoBecomesMain(t *testing.T) {
	photoRepo := &mockPhotoRepository{
		listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
			return []*domain.Photo{}, nil // пустой список
		},
	}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	params := domain.AddPhotoParams{
		HeroID: "hero-1",
		URL:    "https://example.com/photo.jpg",
		IsMain: false, // явно false, но должно стать true
	}
	id, err := uc.Add(context.Background(), params)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if id == "" {
		t.Error("Add returned empty id")
	}

	// Проверяем, что Add был вызван с IsMain=true.
	if len(photoRepo.addCalls) != 1 {
		t.Fatalf("expected 1 add call, got %d", len(photoRepo.addCalls))
	}
	if !photoRepo.addCalls[0].IsMain {
		t.Error("first photo should be marked as main")
	}
}

// TestPhotoUseCase_Add_AutoSortOrder проверяет автоинкремент sort_order.
func TestPhotoUseCase_Add_AutoSortOrder(t *testing.T) {
	photoRepo := &mockPhotoRepository{
		listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
			return []*domain.Photo{{ID: "existing-1"}, {ID: "existing-2"}}, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	params := domain.AddPhotoParams{
		HeroID:    "hero-1",
		URL:       "https://example.com/photo.jpg",
		SortOrder: 0, // должно стать 2
	}
	_, err := uc.Add(context.Background(), params)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if photoRepo.addCalls[0].SortOrder != 2 {
		t.Errorf("expected sort_order=2, got %d", photoRepo.addCalls[0].SortOrder)
	}
}

// --- Тесты BatchAdd ---

// TestPhotoUseCase_BatchAdd_EmptyList проверяет, что пустой список возвращает ошибку.
func TestPhotoUseCase_BatchAdd_EmptyList(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	_, err := uc.BatchAdd(context.Background(), "hero-1", []domain.AddPhotoParams{})
	if !errors.Is(err, domain.ErrPhotosEmpty) {
		t.Errorf("expected ErrPhotosEmpty, got %v", err)
	}
}

// TestPhotoUseCase_BatchAdd_Success проверяет успешное пакетное добавление.
func TestPhotoUseCase_BatchAdd_Success(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	photos := []domain.AddPhotoParams{
		{URL: "https://example.com/photo1.jpg"},
		{URL: "https://example.com/photo2.jpg"},
	}
	added, err := uc.BatchAdd(context.Background(), "hero-1", photos)
	if err != nil {
		t.Fatalf("BatchAdd failed: %v", err)
	}
	if len(added) != 2 {
		t.Errorf("expected 2 added photos, got %d", len(added))
	}
}

// --- Тесты DeleteBatch ---

// TestPhotoUseCase_DeleteBatch_EmptyList проверяет, что пустой список возвращает ошибку.
func TestPhotoUseCase_DeleteBatch_EmptyList(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	count, err := uc.DeleteBatch(context.Background(), "hero-1", []string{})
	if !errors.Is(err, domain.ErrPhotoIDsEmpty) {
		t.Errorf("expected ErrPhotoIDsEmpty, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected count=0, got %d", count)
	}
}

// TestPhotoUseCase_DeleteBatch_Success проверяет успешное пакетное удаление.
func TestPhotoUseCase_DeleteBatch_Success(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	count, err := uc.DeleteBatch(context.Background(), "hero-1", []string{"photo-1", "photo-2"})
	if err != nil {
		t.Fatalf("DeleteBatch failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
}

// --- Тесты Reorder ---

// TestPhotoUseCase_Reorder_EmptyList проверяет, что пустой список возвращает ошибку.
func TestPhotoUseCase_Reorder_EmptyList(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	err := uc.Reorder(context.Background(), "hero-1", []string{})
	if !errors.Is(err, domain.ErrPhotoIDsEmpty) {
		t.Errorf("expected ErrPhotoIDsEmpty, got %v", err)
	}
}

// --- Тесты SetMain ---

// TestPhotoUseCase_SetMain_HeroNotFound проверяет, что несуществующий герой возвращает ошибку.
func TestPhotoUseCase_SetMain_HeroNotFound(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return nil, domain.ErrNotFound
		},
	}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	err := uc.SetMain(context.Background(), "nonexistent", "photo-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestPhotoUseCase_SetMain_Success проверяет успешное назначение главного фото.
func TestPhotoUseCase_SetMain_Success(t *testing.T) {
	photoRepo := &mockPhotoRepository{}
	heroRepo := &mockHeroRepository{}
	storage := &mockMediaStorage{}
	worker := &mockThumbnailWorker{}
	uc := newPhotoUseCase(photoRepo, heroRepo, storage, worker)

	err := uc.SetMain(context.Background(), "hero-1", "photo-1")
	if err != nil {
		t.Fatalf("SetMain failed: %v", err)
	}
}
