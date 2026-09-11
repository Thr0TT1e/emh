package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// HeroSourceRepository абстракция для работы с источниками данных героя.
type HeroSourceRepository interface {
	Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error)
	Remove(ctx context.Context, heroID, sourceID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error)
}
