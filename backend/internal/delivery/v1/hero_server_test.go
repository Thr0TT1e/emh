package v1

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
)

// --- Моки ---

// mockHeroUseCase мок HeroUseCase для тестов handler.
type mockHeroUseCase struct {
	createFunc  func(ctx context.Context, p domain.CreateHeroParams) (string, error)
	updateFunc  func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
	deleteFunc  func(ctx context.Context, id string, hardDelete bool) error
	getByIDFunc func(ctx context.Context, id string) (*domain.Hero, error)
	listFunc    func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
}

func (m *mockHeroUseCase) CreateHero(ctx context.Context, p domain.CreateHeroParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "hero-id-001", nil
}

func (m *mockHeroUseCase) UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Hero{ID: "hero-id-001"}, nil
}

func (m *mockHeroUseCase) DeleteHero(ctx context.Context, id string, hardDelete bool) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, hardDelete)
	}
	return nil
}

func (m *mockHeroUseCase) GetByID(ctx context.Context, id string) (*domain.Hero, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Hero{ID: id}, nil
}

func (m *mockHeroUseCase) ListHeroes(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Hero{}, "", 0, nil
}

// mockHeroQueryUseCase мок HeroQueryUseCase.
type mockHeroQueryUseCase struct {
	getHeroDetailFunc  func(ctx context.Context, id string) (*domain.HeroDetail, error)
	listHeroPhotosFunc func(ctx context.Context, heroID string) ([]*domain.Photo, error)
}

func (m *mockHeroQueryUseCase) GetHeroDetail(ctx context.Context, id string) (*domain.HeroDetail, error) {
	if m.getHeroDetailFunc != nil {
		return m.getHeroDetailFunc(ctx, id)
	}
	return &domain.HeroDetail{
		Hero: &domain.Hero{ID: id, FirstName: "Иван", LastName: "Петров"},
	}, nil
}

func (m *mockHeroQueryUseCase) ListHeroPhotos(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	if m.listHeroPhotosFunc != nil {
		return m.listHeroPhotosFunc(ctx, heroID)
	}
	return []*domain.Photo{}, nil
}

// ListHeroPhotosPaged — заглушка для удовлетворения интерфейса.
// Публичный HeroServer не вызывает этот метод (использует ListHeroPhotos).
func (m *mockHeroQueryUseCase) ListHeroPhotosPaged(ctx context.Context, heroID, cursor string, limit int) ([]*domain.Photo, string, int64, error) {
	return nil, "", 0, nil
}

// mockHeroSourceUseCase мок HeroSourceUseCase.
type mockHeroSourceUseCase struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroSource, error)
}

func (m *mockHeroSourceUseCase) Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error) {
	return "source-id-001", nil
}

func (m *mockHeroSourceUseCase) Remove(ctx context.Context, heroID, sourceID string) error {
	return nil
}

func (m *mockHeroSourceUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroSource{}, nil
}

// mockHeroRelationUseCase мок HeroRelationUseCase.
type mockHeroRelationUseCase struct {
	listByHeroFunc func(ctx context.Context, heroID string) ([]*domain.HeroRelation, error)
}

func (m *mockHeroRelationUseCase) Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
	return "relation-id-001", nil
}

func (m *mockHeroRelationUseCase) Remove(ctx context.Context, id string) error {
	return nil
}

func (m *mockHeroRelationUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
	if m.listByHeroFunc != nil {
		return m.listByHeroFunc(ctx, heroID)
	}
	return []*domain.HeroRelation{}, nil
}

// --- Хелперы ---

// discardLogger возвращает логгер, который ничего не пишет.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newUUID генерирует новый UUID для тестов (локальная версия testutil.NewUUID).
// Без build-тега, доступна в unit-тестах.
func newUUID() string {
	return uuid.New().String()
}

// setupHeroServer создаёт тестовый HTTP-сервер с HeroServer и возвращает Connect-клиент.
func setupHeroServer(
	t *testing.T,
	heroUC *mockHeroUseCase,
	queryUC *mockHeroQueryUseCase,
	sourceUC *mockHeroSourceUseCase,
	relationUC *mockHeroRelationUseCase,
) (emhv1connect.HeroServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewHeroServer(heroUC, queryUC, sourceUC, relationUC, discardLogger())

	validateInterceptor := validate.NewInterceptor()
	path, handler := emhv1connect.NewHeroServiceHandler(
		server,
		connect.WithInterceptors(validateInterceptor),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)

	client := emhv1connect.NewHeroServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты GetHero ---

// TestHeroServer_GetHero_Success проверяет успешное получение героя.
func TestHeroServer_GetHero_Success(t *testing.T) {
	birthAnchor := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	queryUC := &mockHeroQueryUseCase{
		getHeroDetailFunc: func(ctx context.Context, id string) (*domain.HeroDetail, error) {
			return &domain.HeroDetail{
				Hero: &domain.Hero{
					ID:         id,
					FirstName:  "Иван",
					LastName:   "Петров",
					MiddleName: "Сергеевич",
					Rank:       "Старший лейтенант",
					ShortBio:   "Краткая биография",
					BirthDate:  domain.FlexibleDate{Precision: domain.PrecisionExact, Anchor: &birthAnchor},
					DeathDate:  domain.NewUnknownDate(),
					Status:     domain.StatusPublished,
				},
				Photos: []*domain.Photo{
					{ID: "p1", URL: "https://s3.example.com/1.jpg", IsMain: true},
				},
				Awards: []*domain.HeroAward{
					{AwardID: "a1", AwardName: "Герой России"},
				},
			}, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, queryUC, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetHero(ctx, connect.NewRequest(&emhv1.GetHeroRequest{Id: "hero-123"}))
	if err != nil {
		t.Fatalf("GetHero failed: %v", err)
	}

	if resp.Msg.Hero == nil {
		t.Fatal("response hero is nil")
	}
	if resp.Msg.Hero.Summary == nil {
		t.Fatal("summary is nil")
	}
	if resp.Msg.Hero.Summary.Id != "hero-123" {
		t.Errorf("id = %q, want %q", resp.Msg.Hero.Summary.Id, "hero-123")
	}
	if resp.Msg.Hero.Summary.FirstName != "Иван" {
		t.Errorf("first_name = %q", resp.Msg.Hero.Summary.FirstName)
	}
	if len(resp.Msg.Hero.Photos) != 1 {
		t.Errorf("photos count = %d, want 1", len(resp.Msg.Hero.Photos))
	}
	if len(resp.Msg.Hero.Awards) != 1 {
		t.Errorf("awards count = %d, want 1", len(resp.Msg.Hero.Awards))
	}
}

// TestHeroServer_GetHero_NotFound проверяет маппинг ErrNotFound на Connect CodeNotFound.
func TestHeroServer_GetHero_NotFound(t *testing.T) {
	queryUC := &mockHeroQueryUseCase{
		getHeroDetailFunc: func(ctx context.Context, id string) (*domain.HeroDetail, error) {
			return nil, domain.ErrNotFound
		},
	}

	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, queryUC, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetHero(ctx, connect.NewRequest(&emhv1.GetHeroRequest{Id: "nonexistent"}))
	if err == nil {
		t.Fatal("expected error for nonexistent hero")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

// --- Тесты ListHeroes ---

// TestHeroServer_ListHeroes_Success проверяет успешный список с пагинацией.
func TestHeroServer_ListHeroes_Success(t *testing.T) {
	heroUC := &mockHeroUseCase{
		listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
			heroes := []*domain.Hero{
				{ID: "hero-1", FirstName: "Иван", LastName: "Петров", Status: domain.StatusPublished},
				{ID: "hero-2", FirstName: "Пётр", LastName: "Иванов", Status: domain.StatusPublished},
			}
			return heroes, "next-cursor-123", 10, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, heroUC, &mockHeroQueryUseCase{}, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListHeroes(ctx, connect.NewRequest(&emhv1.ListHeroesRequest{
		Pagination: &emhv1.PaginationRequest{PageSize: 20},
	}))
	if err != nil {
		t.Fatalf("ListHeroes failed: %v", err)
	}

	if len(resp.Msg.Heroes) != 2 {
		t.Errorf("heroes count = %d, want 2", len(resp.Msg.Heroes))
	}
	if resp.Msg.Pagination.NextCursor != "next-cursor-123" {
		t.Errorf("next_cursor = %q", resp.Msg.Pagination.NextCursor)
	}
	if resp.Msg.Pagination.TotalCount != 10 {
		t.Errorf("total_count = %d, want 10", resp.Msg.Pagination.TotalCount)
	}
}

// TestHeroServer_ListHeroes_WithFilters проверяет передачу фильтров в usecase.
func TestHeroServer_ListHeroes_WithFilters(t *testing.T) {
	var capturedFilter domain.HeroFilter
	heroUC := &mockHeroUseCase{
		listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
			capturedFilter = f
			return []*domain.Hero{}, "", 0, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, heroUC, &mockHeroQueryUseCase{}, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	dateFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.ListHeroes(ctx, connect.NewRequest(&emhv1.ListHeroesRequest{
		SearchQuery: "Иван",
		ConflictId:  "conflict-1",
		LocationId:  "location-1",
		DateFrom:    timestamppb.New(dateFrom),
		Pagination:  &emhv1.PaginationRequest{PageSize: 50, Cursor: "cursor-abc"},
	}))
	if err != nil {
		t.Fatalf("ListHeroes failed: %v", err)
	}

	if capturedFilter.SearchQuery != "Иван" {
		t.Errorf("search_query = %q", capturedFilter.SearchQuery)
	}
	if capturedFilter.ConflictID != "conflict-1" {
		t.Errorf("conflict_id = %q", capturedFilter.ConflictID)
	}
	if capturedFilter.LocationID != "location-1" {
		t.Errorf("location_id = %q", capturedFilter.LocationID)
	}
	if capturedFilter.DateFrom == nil || *capturedFilter.DateFrom != "2020-01-01" {
		t.Errorf("date_from = %v", capturedFilter.DateFrom)
	}
	if capturedFilter.Limit != 50 {
		t.Errorf("limit = %d, want 50", capturedFilter.Limit)
	}
	if capturedFilter.Cursor != "cursor-abc" {
		t.Errorf("cursor = %q", capturedFilter.Cursor)
	}
}

// TestHeroServer_ListHeroes_PaginationNormalization проверяет нормализацию page_size.
func TestHeroServer_ListHeroes_PaginationNormalization(t *testing.T) {
	var capturedLimit int
	heroUC := &mockHeroUseCase{
		listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
			capturedLimit = f.Limit
			return []*domain.Hero{}, "", 0, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, heroUC, &mockHeroQueryUseCase{}, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	tests := []struct {
		name     string
		pageSize int32
		expected int
	}{
		{"zero becomes 20", 0, 20},
		{"negative becomes 20", -10, 20},
		{"valid preserved", 50, 50},
		{"max capped at 100", 150, 100},
		{"exactly 100", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := client.ListHeroes(ctx, connect.NewRequest(&emhv1.ListHeroesRequest{
				Pagination: &emhv1.PaginationRequest{PageSize: tt.pageSize},
			}))
			if err != nil {
				t.Fatalf("ListHeroes failed: %v", err)
			}
			if capturedLimit != tt.expected {
				t.Errorf("limit = %d, want %d", capturedLimit, tt.expected)
			}
		})
	}
}

// --- Тесты ListHeroPhotos ---

// TestHeroServer_ListHeroPhotos_Success проверяет получение списка фотографий.
func TestHeroServer_ListHeroPhotos_Success(t *testing.T) {
	heroID := newUUID()
	queryUC := &mockHeroQueryUseCase{
		listHeroPhotosFunc: func(ctx context.Context, hid string) ([]*domain.Photo, error) {
			if hid != heroID {
				t.Errorf("heroID = %q, want %q", hid, heroID)
			}
			return []*domain.Photo{
				{ID: "p1", URL: "https://s3.example.com/1.jpg", IsMain: true, SortOrder: 0},
				{ID: "p2", URL: "https://s3.example.com/2.jpg", IsMain: false, SortOrder: 1},
			}, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, queryUC, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListHeroPhotos(ctx, connect.NewRequest(&emhv1.ListHeroPhotosRequest{HeroId: heroID}))
	if err != nil {
		t.Fatalf("ListHeroPhotos failed: %v", err)
	}

	if len(resp.Msg.Photos) != 2 {
		t.Errorf("photos count = %d, want 2", len(resp.Msg.Photos))
	}
	if !resp.Msg.Photos[0].IsMain {
		t.Error("first photo should be main")
	}
}

// --- Тесты ListHeroSources ---

// TestHeroServer_ListHeroSources_Success проверяет получение источников.
func TestHeroServer_ListHeroSources_Success(t *testing.T) {
	heroID := newUUID()
	sourceUC := &mockHeroSourceUseCase{
		listByHeroFunc: func(ctx context.Context, hid string) ([]*domain.HeroSource, error) {
			if hid != heroID {
				t.Errorf("heroID = %q, want %q", hid, heroID)
			}
			return []*domain.HeroSource{
				{ID: "s1", URL: "https://example.com/article", Title: "Статья"},
			}, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, &mockHeroQueryUseCase{}, sourceUC, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListHeroSources(ctx, connect.NewRequest(&emhv1.ListHeroSourcesRequest{HeroId: heroID}))
	if err != nil {
		t.Fatalf("ListHeroSources failed: %v", err)
	}

	if len(resp.Msg.Sources) != 1 {
		t.Errorf("sources count = %d, want 1", len(resp.Msg.Sources))
	}
	if resp.Msg.Sources[0].Title != "Статья" {
		t.Errorf("title = %q", resp.Msg.Sources[0].Title)
	}
}

// TestHeroServer_ListHeroRelations_Success проверяет получение связей.
func TestHeroServer_ListHeroRelations_Success(t *testing.T) {
	heroID := newUUID()
	relationUC := &mockHeroRelationUseCase{
		listByHeroFunc: func(ctx context.Context, hid string) ([]*domain.HeroRelation, error) {
			if hid != heroID {
				t.Errorf("heroID = %q, want %q", hid, heroID)
			}
			return []*domain.HeroRelation{
				{ID: "r1", FromHeroID: heroID, ToHeroID: "hero-2", RelationType: "comrade"},
			}, nil
		},
	}

	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, &mockHeroQueryUseCase{}, &mockHeroSourceUseCase{}, relationUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListHeroRelations(ctx, connect.NewRequest(&emhv1.ListHeroRelationsRequest{HeroId: heroID}))
	if err != nil {
		t.Fatalf("ListHeroRelations failed: %v", err)
	}

	if len(resp.Msg.Relations) != 1 {
		t.Errorf("relations count = %d, want 1", len(resp.Msg.Relations))
	}
	if resp.Msg.Relations[0].RelationType != "comrade" {
		t.Errorf("relation_type = %q", resp.Msg.Relations[0].RelationType)
	}
}

// --- Тесты валидации proto ---

// TestHeroServer_ListHeroPhotos_InvalidUUID проверяет валидацию hero_id.
func TestHeroServer_ListHeroPhotos_InvalidUUID(t *testing.T) {
	client, _, cleanup := setupHeroServer(t, &mockHeroUseCase{}, &mockHeroQueryUseCase{}, &mockHeroSourceUseCase{}, &mockHeroRelationUseCase{})
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListHeroPhotos(ctx, connect.NewRequest(&emhv1.ListHeroPhotosRequest{HeroId: "invalid"}))
	if err == nil {
		t.Fatal("expected validation error")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}
