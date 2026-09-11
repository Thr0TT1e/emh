package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// HeroRepository абстракция для работы с хранилищем героев.
type HeroRepository interface {
	Create(ctx context.Context, params domain.CreateHeroParams) (string, error)
	Update(ctx context.Context, params domain.UpdateHeroParams) (*domain.Hero, error)
	Delete(ctx context.Context, id string, hardDelete bool) error
	GetByID(ctx context.Context, id string) (*domain.Hero, error)
	List(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) // Возвращает список, next_cursor, total
}
