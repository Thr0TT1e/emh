package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
	"connectrpc.com/connect"
)

// --- Моки ---

type mockAuthUseCase struct {
	loginFunc   func(ctx context.Context, username, password string) (*usecase.AuthResult, error)
	refreshFunc func(ctx context.Context, refreshToken string) (*usecase.AuthResult, error)
	logoutFunc  func(ctx context.Context, refreshToken string) error
}

func (m *mockAuthUseCase) Login(ctx context.Context, username, password string) (*usecase.AuthResult, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, username, password)
	}
	return &usecase.AuthResult{
		AccessToken:  "access-token-mock",
		RefreshToken: "refresh-token-mock",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		ExpiresIn:    900,
	}, nil
}

func (m *mockAuthUseCase) Refresh(ctx context.Context, refreshToken string) (*usecase.AuthResult, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, refreshToken)
	}
	return &usecase.AuthResult{
		AccessToken:  "new-access-token-mock",
		RefreshToken: "new-refresh-token-mock",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		ExpiresIn:    900,
	}, nil
}

func (m *mockAuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, refreshToken)
	}
	return nil
}

// --- Хелперы ---

func setupAuthServer(
	t *testing.T,
	authUC *mockAuthUseCase,
) (emhv1connect.AuthServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewAuthServer(authUC, discardLogger())

	path, handler := emhv1connect.NewAuthServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewAuthServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты Login ---

func TestAuthServer_Login_Success(t *testing.T) {
	authUC := &mockAuthUseCase{}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "password",
	}))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.Msg.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if resp.Msg.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if resp.Msg.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", resp.Msg.TokenType)
	}
	if resp.Msg.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", resp.Msg.ExpiresIn)
	}
	if resp.Msg.ExpiresAt == "" {
		t.Error("ExpiresAt should not be empty")
	}
}

func TestAuthServer_Login_InvalidCredentials(t *testing.T) {
	authUC := &mockAuthUseCase{
		loginFunc: func(ctx context.Context, username, password string) (*usecase.AuthResult, error) {
			return nil, errors.New("invalid credentials")
		},
	}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.Login(ctx, connect.NewRequest(&emhv1.LoginRequest{
		Username: "admin",
		Password: "wrong",
	}))
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}
}

// --- Тесты Refresh ---

func TestAuthServer_Refresh_Success(t *testing.T) {
	authUC := &mockAuthUseCase{}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.Refresh(ctx, connect.NewRequest(&emhv1.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}))
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if resp.Msg.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if resp.Msg.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if resp.Msg.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", resp.Msg.TokenType)
	}
}

func TestAuthServer_Refresh_InvalidToken(t *testing.T) {
	authUC := &mockAuthUseCase{
		refreshFunc: func(ctx context.Context, refreshToken string) (*usecase.AuthResult, error) {
			return nil, errors.New("invalid or expired refresh token")
		},
	}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.Refresh(ctx, connect.NewRequest(&emhv1.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}))
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeUnauthenticated)
	}
}

// --- Тесты Logout ---

func TestAuthServer_Logout_Success(t *testing.T) {
	authUC := &mockAuthUseCase{}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.Logout(ctx, connect.NewRequest(&emhv1.LogoutRequest{
		RefreshToken: "valid-refresh-token",
	}))
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	if !resp.Msg.Success {
		t.Error("Success should be true")
	}
}

func TestAuthServer_Logout_UseCaseError(t *testing.T) {
	authUC := &mockAuthUseCase{
		logoutFunc: func(ctx context.Context, refreshToken string) error {
			return errors.New("database error")
		},
	}
	client, _, cleanup := setupAuthServer(t, authUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.Logout(ctx, connect.NewRequest(&emhv1.LogoutRequest{
		RefreshToken: "valid-refresh-token",
	}))
	if err == nil {
		t.Fatal("expected error from usecase")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}
