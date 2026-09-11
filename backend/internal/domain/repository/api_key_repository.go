package repository

import (
	"context"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// APIKeyRepository абстракция для работы с управляемыми API-ключами.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error)
	List(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error)
	Revoke(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string, ip string, at time.Time) error
}
