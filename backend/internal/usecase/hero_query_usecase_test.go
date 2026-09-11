package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ============================================================
// Хелперы
// ============================================================

// newHeroQueryUseCase создаёт HeroQueryUseCase с моками.
func newHeroQueryUseCase(
	heroRepo *mockHeroRepository,
	photoRepo *mockPhotoRepository,
	awardRepo *mockHeroAwardRepository,
	conflictRepo *mockHeroConflictRepository,
	locationRepo *mockHeroLocationRepository,
	sourceRepo *mockHeroSourceRepository,
	relationRepo *mockHeroRelationRepository,
) HeroQueryUseCase {
	return NewHeroQueryUseCase(
		heroRepo,
		photoRepo,
		awardRepo,
		conflictRepo,
		locationRepo,
		sourceRepo,
		relationRepo,
	)
}

// ListByHeroPaged — заглушка для удовлетворения интерфейса.
// Тесты пагинации будут добавлены отдельно.
func (m *mockPhotoRepository) ListByHeroPaged(ctx context.Context, heroID, cursor string, limit int) ([]*domain.Photo, string, int64, error) {
	// Возвращаем все фото без пагинации для существующих тестов
	photos, err := m.ListByHero(ctx, heroID)
	if err != nil {
		return nil, "", 0, err
	}
	return photos, "", int64(len(photos)), nil
}

// sampleHero возвращает тестового героя.
func sampleHero(id string) *domain.Hero {
	return &domain.Hero{
		ID:        id,
		FirstName: "Иван",
		LastName:  "Петров",
	}
}

// defaultRepos создаёт набор моков с успешными ответами.
func defaultRepos() (
	*mockHeroRepository,
	*mockPhotoRepository,
	*mockHeroAwardRepository,
	*mockHeroConflictRepository,
	*mockHeroLocationRepository,
	*mockHeroSourceRepository,
	*mockHeroRelationRepository,
) {
	return &mockHeroRepository{
			getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
				return sampleHero(id), nil
			},
		},
		&mockPhotoRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
				return []*domain.Photo{{ID: "photo-1"}}, nil
			},
		},
		&mockHeroAwardRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
				return []*domain.HeroAward{{AwardID: "award-1", AwardName: "Орден Мужества"}}, nil
			},
		},
		&mockHeroConflictRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
				return []*domain.HeroConflict{{ConflictID: "conflict-1"}}, nil
			},
		},
		&mockHeroLocationRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
				return []*domain.HeroLocation{{LocationID: "loc-1"}}, nil
			},
		},
		&mockHeroSourceRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
				return []*domain.HeroSource{{ID: "src-1"}}, nil
			},
		},
		&mockHeroRelationRepository{
			listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
				return []*domain.HeroRelation{{ID: "rel-1"}}, nil
			},
		}
}

// ============================================================
// Тесты GetHeroDetail
// ============================================================

// TestHeroQueryUseCase_GetHeroDetail_Success проверяет успешную сборку HeroDetail.
func TestHeroQueryUseCase_GetHeroDetail_Success(t *testing.T) {
	heroID := "hero-1"
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	detail, err := uc.GetHeroDetail(context.Background(), heroID)
	if err != nil {
		t.Fatalf("GetHeroDetail failed: %v", err)
	}

	// Проверяем корень агрегата
	if detail.Hero == nil {
		t.Fatal("detail.Hero is nil")
	}
	if detail.Hero.ID != heroID {
		t.Errorf("Hero.ID = %s, want %s", detail.Hero.ID, heroID)
	}

	// Проверяем все связи
	if len(detail.Photos) != 1 {
		t.Errorf("len(Photos) = %d, want 1", len(detail.Photos))
	}
	if len(detail.Awards) != 1 {
		t.Errorf("len(Awards) = %d, want 1", len(detail.Awards))
	}
	if len(detail.Conflicts) != 1 {
		t.Errorf("len(Conflicts) = %d, want 1", len(detail.Conflicts))
	}
	if len(detail.Locations) != 1 {
		t.Errorf("len(Locations) = %d, want 1", len(detail.Locations))
	}
	if len(detail.Sources) != 1 {
		t.Errorf("len(Sources) = %d, want 1", len(detail.Sources))
	}
	if len(detail.Relations) != 1 {
		t.Errorf("len(Relations) = %d, want 1", len(detail.Relations))
	}

	// Проверяем, что все 6 репозиториев связей были вызваны ровно 1 раз
	if photoRepo.listByHeroCalls.Load() != 1 {
		t.Errorf("photoRepo called %d times, want 1", photoRepo.listByHeroCalls.Load())
	}
	if awardRepo.callCount.Load() != 1 {
		t.Errorf("awardRepo called %d times, want 1", awardRepo.callCount.Load())
	}
	if conflictRepo.callCount.Load() != 1 {
		t.Errorf("conflictRepo called %d times, want 1", conflictRepo.callCount.Load())
	}
	if locationRepo.callCount.Load() != 1 {
		t.Errorf("locationRepo called %d times, want 1", locationRepo.callCount.Load())
	}
	if sourceRepo.callCount.Load() != 1 {
		t.Errorf("sourceRepo called %d times, want 1", sourceRepo.callCount.Load())
	}
	if relationRepo.callCount.Load() != 1 {
		t.Errorf("relationRepo called %d times, want 1", relationRepo.callCount.Load())
	}
}

// TestHeroQueryUseCase_GetHeroDetail_HeroNotFound проверяет ранний выход при отсутствии героя.
func TestHeroQueryUseCase_GetHeroDetail_HeroNotFound(t *testing.T) {
	heroID := "nonexistent"
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return nil, domain.ErrNotFound
		},
	}
	photoRepo := &mockPhotoRepository{}
	awardRepo := &mockHeroAwardRepository{}
	conflictRepo := &mockHeroConflictRepository{}
	locationRepo := &mockHeroLocationRepository{}
	sourceRepo := &mockHeroSourceRepository{}
	relationRepo := &mockHeroRelationRepository{}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), heroID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Проверяем, что репозитории связей НЕ вызывались
	if heroRepo.getByIDCalls.Load() != 1 {
		t.Errorf("heroRepo called %d times, want 1", heroRepo.getByIDCalls.Load())
	}
	if photoRepo.listByHeroCalls.Load() != 0 {
		t.Errorf("photoRepo called %d times, want 0 (early exit)", photoRepo.listByHeroCalls.Load())
	}
	if awardRepo.callCount.Load() != 0 {
		t.Errorf("awardRepo called %d times, want 0", awardRepo.callCount.Load())
	}
	if conflictRepo.callCount.Load() != 0 {
		t.Errorf("conflictRepo called %d times, want 0", conflictRepo.callCount.Load())
	}
	if locationRepo.callCount.Load() != 0 {
		t.Errorf("locationRepo called %d times, want 0", locationRepo.callCount.Load())
	}
	if sourceRepo.callCount.Load() != 0 {
		t.Errorf("sourceRepo called %d times, want 0", sourceRepo.callCount.Load())
	}
	if relationRepo.callCount.Load() != 0 {
		t.Errorf("relationRepo called %d times, want 0", relationRepo.callCount.Load())
	}
}

// TestHeroQueryUseCase_GetHeroDetail_EmptyRelations проверяет героя без связей.
func TestHeroQueryUseCase_GetHeroDetail_EmptyRelations(t *testing.T) {
	heroID := "hero-lonely"
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return sampleHero(id), nil
		},
	}
	photoRepo := &mockPhotoRepository{}
	awardRepo := &mockHeroAwardRepository{}
	conflictRepo := &mockHeroConflictRepository{}
	locationRepo := &mockHeroLocationRepository{}
	sourceRepo := &mockHeroSourceRepository{}
	relationRepo := &mockHeroRelationRepository{}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	detail, err := uc.GetHeroDetail(context.Background(), heroID)
	if err != nil {
		t.Fatalf("GetHeroDetail failed: %v", err)
	}

	if detail.Hero == nil {
		t.Fatal("detail.Hero is nil")
	}
	if len(detail.Photos) != 0 {
		t.Errorf("len(Photos) = %d, want 0", len(detail.Photos))
	}
	if len(detail.Awards) != 0 {
		t.Errorf("len(Awards) = %d, want 0", len(detail.Awards))
	}
	if len(detail.Conflicts) != 0 {
		t.Errorf("len(Conflicts) = %d, want 0", len(detail.Conflicts))
	}
	if len(detail.Locations) != 0 {
		t.Errorf("len(Locations) = %d, want 0", len(detail.Locations))
	}
	if len(detail.Sources) != 0 {
		t.Errorf("len(Sources) = %d, want 0", len(detail.Sources))
	}
	if len(detail.Relations) != 0 {
		t.Errorf("len(Relations) = %d, want 0", len(detail.Relations))
	}
}

// TestHeroQueryUseCase_GetHeroDetail_PhotoError проверяет ошибку загрузки фото.
func TestHeroQueryUseCase_GetHeroDetail_PhotoError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	photoRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
		return nil, errors.New("db connection failed")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from photo repo")
	}
	if !strings.Contains(err.Error(), "load photos") {
		t.Errorf("expected 'load photos' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_AwardError проверяет ошибку загрузки наград.
func TestHeroQueryUseCase_GetHeroDetail_AwardError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	awardRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
		return nil, errors.New("db timeout")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from award repo")
	}
	if !strings.Contains(err.Error(), "load awards") {
		t.Errorf("expected 'load awards' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_ConflictError проверяет ошибку загрузки конфликтов.
func TestHeroQueryUseCase_GetHeroDetail_ConflictError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	conflictRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
		return nil, errors.New("conflict db error")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from conflict repo")
	}
	if !strings.Contains(err.Error(), "load conflicts") {
		t.Errorf("expected 'load conflicts' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_LocationError проверяет ошибку загрузки локаций.
func TestHeroQueryUseCase_GetHeroDetail_LocationError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	locationRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
		return nil, errors.New("location db error")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from location repo")
	}
	if !strings.Contains(err.Error(), "load locations") {
		t.Errorf("expected 'load locations' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_SourceError проверяет ошибку загрузки источников.
func TestHeroQueryUseCase_GetHeroDetail_SourceError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	sourceRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
		return nil, errors.New("source db error")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from source repo")
	}
	if !strings.Contains(err.Error(), "load sources") {
		t.Errorf("expected 'load sources' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_RelationError проверяет ошибку загрузки связей.
func TestHeroQueryUseCase_GetHeroDetail_RelationError(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	relationRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
		return nil, errors.New("relation db error")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from relation repo")
	}
	if !strings.Contains(err.Error(), "load relations") {
		t.Errorf("expected 'load relations' in error, got %v", err)
	}
}

// TestHeroQueryUseCase_GetHeroDetail_ContextCancelled проверяет отмену контекста.
func TestHeroQueryUseCase_GetHeroDetail_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // сразу отменяем

	// Context-aware моки: проверяют ctx.Err() перед возвратом данных
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return sampleHero(id), nil
		},
	}
	photoRepo := &mockPhotoRepository{
		listByHeroFunc: func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return []*domain.Photo{{ID: "photo-1"}}, nil
		},
	}
	awardRepo := &mockHeroAwardRepository{}
	conflictRepo := &mockHeroConflictRepository{}
	locationRepo := &mockHeroLocationRepository{}
	sourceRepo := &mockHeroSourceRepository{}
	relationRepo := &mockHeroRelationRepository{}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.GetHeroDetail(ctx, "hero-1")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ============================================================
// Тесты ListHeroPhotos
// ============================================================

// TestHeroQueryUseCase_ListHeroPhotos_Success проверяет успешный возврат фото.
func TestHeroQueryUseCase_ListHeroPhotos_Success(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	expectedPhotos := []*domain.Photo{
		{ID: "photo-1", URL: "https://example.com/1.jpg", IsMain: true},
		{ID: "photo-2", URL: "https://example.com/2.jpg"},
		{ID: "photo-3", URL: "https://example.com/3.jpg"},
	}
	photoRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
		return expectedPhotos, nil
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	photos, err := uc.ListHeroPhotos(context.Background(), "hero-1")
	if err != nil {
		t.Fatalf("ListHeroPhotos failed: %v", err)
	}
	if len(photos) != 3 {
		t.Errorf("len(photos) = %d, want 3", len(photos))
	}
	for i, p := range photos {
		if p.ID != expectedPhotos[i].ID {
			t.Errorf("photos[%d].ID = %s, want %s", i, p.ID, expectedPhotos[i].ID)
		}
	}
}

// TestHeroQueryUseCase_ListHeroPhotos_Empty проверяет пустой список фото.
func TestHeroQueryUseCase_ListHeroPhotos_Empty(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	photoRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
		return []*domain.Photo{}, nil
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	photos, err := uc.ListHeroPhotos(context.Background(), "hero-1")
	if err != nil {
		t.Fatalf("ListHeroPhotos failed: %v", err)
	}
	if len(photos) != 0 {
		t.Errorf("len(photos) = %d, want 0", len(photos))
	}
}

// TestHeroQueryUseCase_ListHeroPhotos_Error проверяет ошибку репозитория.
func TestHeroQueryUseCase_ListHeroPhotos_Error(t *testing.T) {
	heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo := defaultRepos()
	photoRepo.listByHeroFunc = func(ctx context.Context, heroID string) ([]*domain.Photo, error) {
		return nil, errors.New("db error")
	}

	uc := newHeroQueryUseCase(heroRepo, photoRepo, awardRepo, conflictRepo, locationRepo, sourceRepo, relationRepo)

	_, err := uc.ListHeroPhotos(context.Background(), "hero-1")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
