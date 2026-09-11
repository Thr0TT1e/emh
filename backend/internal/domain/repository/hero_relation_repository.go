package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// HeroRelationRepository абстракция для связей между героями.
type HeroRelationRepository interface {
	Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error)
	Remove(ctx context.Context, id string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error)
}
