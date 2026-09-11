package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"connectrpc.com/connect"
)

// --- Моки ---

type mockLocationUseCase struct {
	getByIDFunc   func(ctx context.Context, id string) (*domain.Location, error)
	listFunc      func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error)
	createFunc    func(ctx context.Context, p domain.CreateLocationParams) (string, error)
	updateFunc    func(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error)
	deleteFunc    func(ctx context.Context, id string) error
	reparentFunc  func(ctx context.Context, oldParentID, newParentID string) (int, error)
	hasHeroesFunc func(ctx context.Context, locationID string) (bool, error)
}

func (m *mockLocationUseCase) Create(ctx context.Context, p domain.CreateLocationParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "location-id-001", nil
}

func (m *mockLocationUseCase) Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Location{ID: p.ID}, nil
}

func (m *mockLocationUseCase) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockLocationUseCase) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Location{ID: id, Name: "Москва"}, nil
}

func (m *mockLocationUseCase) List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Location{}, "", 0, nil
}

func (m *mockLocationUseCase) Reparent(ctx context.Context, oldParentID, newParentID string) (int, error) {
	if m.reparentFunc != nil {
		return m.reparentFunc(ctx, oldParentID, newParentID)
	}
	return 0, nil
}

func (m *mockLocationUseCase) HasHeroes(ctx context.Context, locationID string) (bool, error) {
	if m.hasHeroesFunc != nil {
		return m.hasHeroesFunc(ctx, locationID)
	}
	return false, nil
}

// --- Хелперы ---

func setupLocationServer(
	t *testing.T,
	locationUC *mockLocationUseCase,
) (emhv1connect.LocationServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewLocationServer(locationUC, discardLogger())

	path, handler := emhv1connect.NewLocationServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewLocationServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты GetLocation ---

// TestLocationServer_GetLocation_Success проверяет успешное получение локации.
func TestLocationServer_GetLocation_Success(t *testing.T) {
	lat := 55.7558
	lon := 37.6173
	parentID := "region-1"
	locationUC := &mockLocationUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return &domain.Location{
				ID:             id,
				Name:           "Москва",
				HistoricalName: "Москва",
				Type:           domain.LocationType(emhv1.LocationType_LOCATION_TYPE_CITY),
				ParentID:       &parentID,
				Latitude:       &lat,
				Longitude:      &lon,
			}, nil
		},
	}
	client, _, cleanup := setupLocationServer(t, locationUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetLocation(ctx, connect.NewRequest(&emhv1.GetLocationRequest{
		Id: "location-123",
	}))
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}

	if resp.Msg.Location == nil {
		t.Fatal("Location is nil")
	}
	if resp.Msg.Location.Id != "location-123" {
		t.Errorf("Id = %q, want location-123", resp.Msg.Location.Id)
	}
	if resp.Msg.Location.Name != "Москва" {
		t.Errorf("Name = %q", resp.Msg.Location.Name)
	}
	if resp.Msg.Location.Type != emhv1.LocationType_LOCATION_TYPE_CITY {
		t.Errorf("Type = %v, want CITY", resp.Msg.Location.Type)
	}
}

// TestLocationServer_GetLocation_NotFound проверяет обработку ErrNotFound.
func TestLocationServer_GetLocation_NotFound(t *testing.T) {
	locationUC := &mockLocationUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	client, _, cleanup := setupLocationServer(t, locationUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetLocation(ctx, connect.NewRequest(&emhv1.GetLocationRequest{
		Id: "nonexistent",
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent location")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

// --- Тесты ListLocations ---

// TestLocationServer_ListLocations_Success проверяет успешное получение списка.
func TestLocationServer_ListLocations_Success(t *testing.T) {
	locationUC := &mockLocationUseCase{
		listFunc: func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
			return []*domain.Location{
				{ID: "loc-1", Name: "Москва", Type: domain.LocationType(emhv1.LocationType_LOCATION_TYPE_CITY)},
				{ID: "loc-2", Name: "Санкт-Петербург", Type: domain.LocationType(emhv1.LocationType_LOCATION_TYPE_CITY)},
			}, "", 2, nil
		},
	}
	client, _, cleanup := setupLocationServer(t, locationUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListLocations(ctx, connect.NewRequest(&emhv1.ListLocationsRequest{
		Type: emhv1.LocationType_LOCATION_TYPE_CITY,
	}))
	if err != nil {
		t.Fatalf("ListLocations failed: %v", err)
	}

	if len(resp.Msg.Locations) != 2 {
		t.Errorf("Locations count = %d, want 2", len(resp.Msg.Locations))
	}
	if resp.Msg.Locations[0].Id != "loc-1" {
		t.Errorf("Locations[0].Id = %q", resp.Msg.Locations[0].Id)
	}
}

// TestLocationServer_ListLocations_WithFilters проверяет передачу всех фильтров.
func TestLocationServer_ListLocations_WithFilters(t *testing.T) {
	var capturedFilter domain.LocationFilter
	locationUC := &mockLocationUseCase{
		listFunc: func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
			capturedFilter = f
			return []*domain.Location{}, "", 0, nil
		},
	}
	client, _, cleanup := setupLocationServer(t, locationUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListLocations(ctx, connect.NewRequest(&emhv1.ListLocationsRequest{
		ParentId:    "country-1",
		SearchQuery: "Моск",
		Type:        emhv1.LocationType_LOCATION_TYPE_CITY,
	}))
	if err != nil {
		t.Fatalf("ListLocations failed: %v", err)
	}
	if capturedFilter.ParentID != "country-1" {
		t.Errorf("ParentID = %q, want country-1", capturedFilter.ParentID)
	}
	if capturedFilter.SearchQuery != "Моск" {
		t.Errorf("SearchQuery = %q, want Моск", capturedFilter.SearchQuery)
	}
	if capturedFilter.Type != domain.LocationType(emhv1.LocationType_LOCATION_TYPE_CITY) {
		t.Errorf("Type = %v, want CITY", capturedFilter.Type)
	}
}

// TestLocationServer_ListLocations_RepoError проверяет обработку ошибки репозитория.
func TestLocationServer_ListLocations_RepoError(t *testing.T) {
	locationUC := &mockLocationUseCase{
		listFunc: func(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
			return nil, "", 0, errors.New("db error")
		},
	}
	client, _, cleanup := setupLocationServer(t, locationUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListLocations(ctx, connect.NewRequest(&emhv1.ListLocationsRequest{}))
	if err == nil {
		t.Fatal("expected error from repo")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
