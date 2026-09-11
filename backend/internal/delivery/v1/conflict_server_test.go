package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"connectrpc.com/connect"
)

// --- Моки ---

type mockConflictUseCase struct {
	getByIDFunc func(ctx context.Context, id string) (*domain.Conflict, error)
	listFunc    func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error)
	createFunc  func(ctx context.Context, p domain.CreateConflictParams) (string, error)
	updateFunc  func(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error)
	deleteFunc  func(ctx context.Context, id string) error
}

func (m *mockConflictUseCase) GetByID(ctx context.Context, id string) (*domain.Conflict, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Conflict{ID: id, Name: "Test Conflict"}, nil
}

func (m *mockConflictUseCase) List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Conflict{}, "", 0, nil
}

func (m *mockConflictUseCase) Create(ctx context.Context, p domain.CreateConflictParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "conflict-id-001", nil
}

func (m *mockConflictUseCase) Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, p)
	}
	return &domain.Conflict{ID: p.ID, Name: p.Name}, nil
}

func (m *mockConflictUseCase) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// --- Хелперы ---

func setupConflictServer(
	t *testing.T,
	conflictUC *mockConflictUseCase,
) (emhv1connect.ConflictServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewConflictServer(conflictUC, discardLogger())

	path, handler := emhv1connect.NewConflictServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewConflictServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты GetConflict ---

func TestConflictServer_GetConflict_Success(t *testing.T) {
	startDate := time.Date(1941, 6, 22, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(1945, 5, 9, 0, 0, 0, 0, time.UTC)

	conflictUC := &mockConflictUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return &domain.Conflict{
				ID:          id,
				Name:        "Великая Отечественная война",
				Description: "1941-1945",
				Type:        domain.ConflictType(emhv1.ConflictType_CONFLICT_TYPE_GLOBAL),
				StartDate:   &startDate,
				EndDate:     &endDate,
			}, nil
		},
	}
	client, _, cleanup := setupConflictServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetConflict(ctx, connect.NewRequest(&emhv1.GetConflictRequest{
		Id: "conflict-123",
	}))
	if err != nil {
		t.Fatalf("GetConflict failed: %v", err)
	}

	if resp.Msg.Conflict == nil {
		t.Fatal("Conflict is nil")
	}
	if resp.Msg.Conflict.Id != "conflict-123" {
		t.Errorf("Id = %q, want conflict-123", resp.Msg.Conflict.Id)
	}
	if resp.Msg.Conflict.Name != "Великая Отечественная война" {
		t.Errorf("Name = %q", resp.Msg.Conflict.Name)
	}
	if resp.Msg.Conflict.Type != emhv1.ConflictType_CONFLICT_TYPE_GLOBAL {
		t.Errorf("Type = %v, want CONFLICT_TYPE_GLOBAL", resp.Msg.Conflict.Type)
	}
}

func TestConflictServer_GetConflict_NotFound(t *testing.T) {
	conflictUC := &mockConflictUseCase{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Conflict, error) {
			return nil, domain.ErrNotFound
		},
	}
	client, _, cleanup := setupConflictServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetConflict(ctx, connect.NewRequest(&emhv1.GetConflictRequest{
		Id: "nonexistent",
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent conflict")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

// --- Тесты ListConflicts ---

func TestConflictServer_ListConflicts_Success(t *testing.T) {
	conflictUC := &mockConflictUseCase{
		listFunc: func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
			return []*domain.Conflict{
				{ID: "c1", Name: "Conflict 1", Type: domain.ConflictType(emhv1.ConflictType_CONFLICT_TYPE_GLOBAL)},
				{ID: "c2", Name: "Conflict 2", Type: domain.ConflictType(emhv1.ConflictType_CONFLICT_TYPE_GLOBAL)},
			}, "", 2, nil
		},
	}
	client, _, cleanup := setupConflictServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListConflicts(ctx, connect.NewRequest(&emhv1.ListConflictsRequest{
		Type: emhv1.ConflictType_CONFLICT_TYPE_GLOBAL,
	}))
	if err != nil {
		t.Fatalf("ListConflicts failed: %v", err)
	}

	if len(resp.Msg.Conflicts) != 2 {
		t.Errorf("Conflicts count = %d, want 2", len(resp.Msg.Conflicts))
	}
	if resp.Msg.Conflicts[0].Id != "c1" {
		t.Errorf("Conflicts[0].Id = %q", resp.Msg.Conflicts[0].Id)
	}
}

func TestConflictServer_ListConflicts_WithParentID(t *testing.T) {
	var capturedFilter domain.ConflictFilter
	conflictUC := &mockConflictUseCase{
		listFunc: func(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
			capturedFilter = f
			return []*domain.Conflict{}, "", 0, nil
		},
	}
	client, _, cleanup := setupConflictServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.ListConflicts(ctx, connect.NewRequest(&emhv1.ListConflictsRequest{
		ParentConflictId: "parent-conflict-id",
	}))
	if err != nil {
		t.Fatalf("ListConflicts failed: %v", err)
	}
	if capturedFilter.ParentID != "parent-conflict-id" {
		t.Errorf("ParentID = %q, want parent-conflict-id", capturedFilter.ParentID)
	}
}
