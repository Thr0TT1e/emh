package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"connectrpc.com/connect"
)

// --- Хелперы ---

func setupConflictAdminServer(
	t *testing.T,
	conflictUC *mockConflictUseCase,
) (emhv1connect.ConflictAdminServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewConflictAdminServer(conflictUC, discardLogger())

	path, handler := emhv1connect.NewConflictAdminServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewConflictAdminServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты CreateConflict ---

func TestConflictAdminServer_CreateConflict_Success(t *testing.T) {
	conflictUC := &mockConflictUseCase{}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.CreateConflict(ctx, connect.NewRequest(&emhv1.CreateConflictRequest{
		Name:        "Новый конфликт",
		Description: "Описание",
		Type:        emhv1.ConflictType_CONFLICT_TYPE_GLOBAL,
		StartDate:   "2020-01-01",
		EndDate:     "2025-12-31",
	}))
	if err != nil {
		t.Fatalf("CreateConflict failed: %v", err)
	}

	if resp.Msg.Id == "" {
		t.Error("Id should not be empty")
	}
}

func TestConflictAdminServer_CreateConflict_InvalidStartDate(t *testing.T) {
	conflictUC := &mockConflictUseCase{}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.CreateConflict(ctx, connect.NewRequest(&emhv1.CreateConflictRequest{
		Name:      "Conflict",
		StartDate: "invalid-date",
	}))
	if err == nil {
		t.Fatal("expected error for invalid start_date")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

// --- Тесты UpdateConflict ---

func TestConflictAdminServer_UpdateConflict_Success(t *testing.T) {
	conflictUC := &mockConflictUseCase{}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.UpdateConflict(ctx, connect.NewRequest(&emhv1.UpdateConflictRequest{
		Id:          "conflict-123",
		Name:        "Updated Name",
		Description: "Updated description",
		FieldMask:   []string{"name", "description"},
	}))
	if err != nil {
		t.Fatalf("UpdateConflict failed: %v", err)
	}

	if resp.Msg.Conflict == nil {
		t.Fatal("Conflict is nil")
	}
	if resp.Msg.Conflict.Id != "conflict-123" {
		t.Errorf("Id = %q, want conflict-123", resp.Msg.Conflict.Id)
	}
}

// --- Тесты DeleteConflict ---

func TestConflictAdminServer_DeleteConflict_Success(t *testing.T) {
	conflictUC := &mockConflictUseCase{}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.DeleteConflict(ctx, connect.NewRequest(&emhv1.DeleteConflictRequest{
		Id: "conflict-123",
	}))
	if err != nil {
		t.Fatalf("DeleteConflict failed: %v", err)
	}

	if !resp.Msg.Success {
		t.Error("Success should be true")
	}
}

func TestConflictAdminServer_DeleteConflict_NotFound(t *testing.T) {
	conflictUC := &mockConflictUseCase{
		deleteFunc: func(ctx context.Context, id string) error {
			return domain.ErrNotFound
		},
	}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.DeleteConflict(ctx, connect.NewRequest(&emhv1.DeleteConflictRequest{
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

func TestConflictAdminServer_DeleteConflict_HasHeroes(t *testing.T) {
	conflictUC := &mockConflictUseCase{
		deleteFunc: func(ctx context.Context, id string) error {
			return domain.ErrEntityAssignedToHeroes
		},
	}
	client, _, cleanup := setupConflictAdminServer(t, conflictUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.DeleteConflict(ctx, connect.NewRequest(&emhv1.DeleteConflictRequest{
		Id: "conflict-with-heroes",
	}))
	if err == nil {
		t.Fatal("expected error for conflict with heroes")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeFailedPrecondition)
	}
}
