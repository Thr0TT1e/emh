package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// RefreshTokenRepository абстракция для работы с refresh-токенами.
type RefreshTokenRepository interface {
	// Create сохраняет новый refresh-токен (хеш) в БД.
	Create(ctx context.Context, p domain.CreateRefreshTokenParams) (string, error)
	// GetByHash возвращает активный (не отозванный, не истёкший) токен по хешу.
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	// GetByHashIncludingRevoked возвращает токен даже если отозван (для детекта кражи)
	GetByHashIncludingRevoked(ctx context.Context, hash string) (*domain.RefreshToken, error)
	// Revoke помечает токен как отозванный.
	Revoke(ctx context.Context, id string) error
	// RevokeAllForUser отозвать ВСЕ активные токены пользователя (при детекте кражи)
	RevokeAllForUser(ctx context.Context, username string) (int64, error)
	// DeleteExpired удаляет истёкшие токены (периодическая очистка).
	DeleteExpired(ctx context.Context) (int64, error)
}
