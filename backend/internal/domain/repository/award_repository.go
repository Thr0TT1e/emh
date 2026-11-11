package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// AwardRepository абстракция для работы с наградами.
type AwardRepository interface {
	Create(ctx context.Context, p domain.CreateAwardParams) (string, error)
	Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Award, error)
	List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error)
	HasHeroes(ctx context.Context, awardID string) (bool, error)

	// ExistsByKey — дешёвая проверка существования награды по составному ключу
	// сида (name, type, jurisdiction). Используется seeder'ом до загрузок в MinIO.
	ExistsByKey(ctx context.Context, name string, awardType domain.AwardType, jurisdiction domain.AwardJurisdiction) (bool, error)

	// InsertIfAbsent — INSERT ... ON CONFLICT (name, type, jurisdiction) DO NOTHING.
	// Возвращает true, если строка вставлена; false — если уже существует
	// (защита от гонки сида с админкой, ручные правки не перезаписываются).
	InsertIfAbsent(ctx context.Context, p domain.CreateAwardParams) (bool, error)
}

// HeroAwardRepository абстракция для связи герой-награда.
type HeroAwardRepository interface {
	Add(ctx context.Context, p domain.AddHeroAwardParams) error
	Remove(ctx context.Context, heroID, awardID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error)
}
