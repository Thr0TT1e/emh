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

type mockAwardUseCase struct {
	getByIDFunc func(ctx context.Context, id string) (*domain.Award, error)
	listFunc    func(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error)
	createFunc  func(ctx context.Context, p domain.CreateAwardParams) (string, error)
	updateFunc  func(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error)
	deleteFunc  func(ctx context.Context, id string) error
}

func (m *mockAwardUseCase) Create(ctx context.Context, p domain.CreateAwardParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "award-id-001", nil
}

func (m *mockAwardUseCase) Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Award{ID: p.ID}, nil
}

func (m *mockAwardUseCase) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockAwardUseCase) GetByID(ctx context.Context, id string) (*domain.Award, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Award{ID: id, Name: "Герой России"}, nil
}

func (m *mockAwardUseCase) List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Award{}, "", 0, nil
}

// --- Хелперы ---

func setupAwardServer(
	t *testing.T,
	awardUC *mockAwardUseCase,
) (emhv1connect.AwardServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewAwardServer(awardUC, discardLogger())

	path, handler := emhv1connect.NewAwardServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewAwardServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты GetAward ---

// TestAwardServer_GetAward_Success проверяет успешное получение награды.
func TestAwardServer_GetAward_Success(t *testing.T) {
	awardUC := &mockAwardUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Award, error) {
			return &domain.Award{
				ID:          id,
				Name:        "Герой Российской Федерации",
				Description: "Высшее звание РФ",
			}, nil
		},
	}
	client, _, cleanup := setupAwardServer(t, awardUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetAward(ctx, connect.NewRequest(&emhv1.GetAwardRequest{
		Id: "award-123",
	}))
	if err != nil {
		t.Fatalf("GetAward failed: %v", err)
	}

	if resp.Msg.Award == nil {
		t.Fatal("Award is nil")
	}
	if resp.Msg.Award.Id != "award-123" {
		t.Errorf("Id = %q, want award-123", resp.Msg.Award.Id)
	}
	if resp.Msg.Award.Name != "Герой Российской Федерации" {
		t.Errorf("Name = %q", resp.Msg.Award.Name)
	}
}

// TestAwardServer_GetAward_NotFound проверяет обработку ErrNotFound.
func TestAwardServer_GetAward_NotFound(t *testing.T) {
	awardUC := &mockAwardUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Award, error) {
			return nil, domain.ErrNotFound
		},
	}
	client, _, cleanup := setupAwardServer(t, awardUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetAward(ctx, connect.NewRequest(&emhv1.GetAwardRequest{
		Id: "nonexistent",
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent award")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

// --- Тесты ListAwards ---

// TestAwardServer_ListAwards_Success проверяет успешный возврат списка.
func TestAwardServer_ListAwards_Success(t *testing.T) {
	awardUC := &mockAwardUseCase{
		listFunc: func(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
			return []*domain.Award{
				{ID: "a1", Name: "Орден Мужества"},
				{ID: "a2", Name: "Медаль «За отвагу»"},
			}, "", 2, nil
		},
	}
	client, _, cleanup := setupAwardServer(t, awardUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListAwards(ctx, connect.NewRequest(&emhv1.ListAwardsRequest{}))
	if err != nil {
		t.Fatalf("ListAwards failed: %v", err)
	}

	if len(resp.Msg.Awards) != 2 {
		t.Errorf("Awards count = %d, want 2", len(resp.Msg.Awards))
	}
	if resp.Msg.Awards[0].Name != "Орден Мужества" {
		t.Errorf("Awards[0].Name = %q", resp.Msg.Awards[0].Name)
	}
}

// TestAwardServer_ListAwards_WithSearch проверяет передачу search_query.
func TestAwardServer_ListAwards_WithSearch(t *testing.T) {
	var capturedFilter domain.AwardFilter
	awardUC := &mockAwardUseCase{
		listFunc: func(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
			capturedFilter = f
			return []*domain.Award{
				{ID: "a1", Name: "Орден Мужества"},
			}, "", 1, nil
		},
	}
	client, _, cleanup := setupAwardServer(t, awardUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListAwards(ctx, connect.NewRequest(&emhv1.ListAwardsRequest{
		SearchQuery: "Орден",
	}))
	if err != nil {
		t.Fatalf("ListAwards failed: %v", err)
	}
	if capturedFilter.SearchQuery != "Орден" {
		t.Errorf("SearchQuery = %q, want Орден", capturedFilter.SearchQuery)
	}
}

// TestAwardServer_ListAwards_RepoError проверяет обработку ошибки репозитория.
func TestAwardServer_ListAwards_RepoError(t *testing.T) {
	awardUC := &mockAwardUseCase{
		listFunc: func(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
			return nil, "", 0, errors.New("db error")
		},
	}
	client, _, cleanup := setupAwardServer(t, awardUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListAwards(ctx, connect.NewRequest(&emhv1.ListAwardsRequest{}))
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
