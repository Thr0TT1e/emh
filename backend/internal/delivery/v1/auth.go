package v1

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// AuthServer реализует emhv1.AuthServiceHandler.
type AuthServer struct {
	authUC usecase.AuthUseCase
	logger *slog.Logger
}

func NewAuthServer(uc usecase.AuthUseCase, logger *slog.Logger) *AuthServer {
	return &AuthServer{authUC: uc, logger: logger}
}

// Login выдаёт пару access/refresh токенов при успешной проверке учётных данных.
func (s *AuthServer) Login(
	ctx context.Context,
	req *connect.Request[emhv1.LoginRequest],
) (*connect.Response[emhv1.LoginResponse], error) {
	s.logger.InfoContext(ctx, "admin login attempt", "username", req.Msg.Username)

	result, err := s.authUC.Login(ctx, req.Msg.Username, req.Msg.Password)
	if err != nil {
		s.logger.WarnContext(ctx, "login failed", "username", req.Msg.Username)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	s.logger.InfoContext(ctx, "admin login success", "username", req.Msg.Username)
	return connect.NewResponse(&emhv1.LoginResponse{
		AccessToken:  result.AccessToken,
		TokenType:    "Bearer",
		ExpiresAt:    result.ExpiresAt.Format(time.RFC3339),
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}), nil
}

// Refresh обменивает действующий refresh-токен на новую пару.
func (s *AuthServer) Refresh(
	ctx context.Context,
	req *connect.Request[emhv1.RefreshTokenRequest],
) (*connect.Response[emhv1.RefreshTokenResponse], error) {
	s.logger.InfoContext(ctx, "token refresh attempt")

	result, err := s.authUC.Refresh(ctx, req.Msg.RefreshToken)
	if err != nil {
		s.logger.WarnContext(ctx, "token refresh failed", "error", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or expired refresh token"))
	}

	s.logger.InfoContext(ctx, "token refresh success")
	return connect.NewResponse(&emhv1.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		TokenType:    "Bearer",
		ExpiresAt:    result.ExpiresAt.Format(time.RFC3339),
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}), nil
}

// Logout отзывает refresh-токен.
func (s *AuthServer) Logout(
	ctx context.Context,
	req *connect.Request[emhv1.LogoutRequest],
) (*connect.Response[emhv1.LogoutResponse], error) {
	s.logger.InfoContext(ctx, "logout attempt")

	if err := s.authUC.Logout(ctx, req.Msg.RefreshToken); err != nil {
		s.logger.ErrorContext(ctx, "logout failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	s.logger.InfoContext(ctx, "logout success")
	return connect.NewResponse(&emhv1.LogoutResponse{Success: true}), nil
}
