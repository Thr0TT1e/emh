package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ConflictRepository абстракция для работы с конфликтами.
type ConflictRepository interface {
	Create(ctx context.Context, p domain.CreateConflictParams) (string, error)
	Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Conflict, error)
	List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error)
	HasHeroes(ctx context.Context, conflictID string) (bool, error)
	// HasCyclicReference проверяет, приведёт ли установка parent_conflict_id = candidateParentID
	// для конфликта selfID к циклу в иерархии.
	HasCyclicReference(ctx context.Context, selfID, candidateParentID string) (bool, error)
}

// HeroConflictRepository абстракция для связи герой-конфликт.
type HeroConflictRepository interface {
	Add(ctx context.Context, heroID, conflictID, specificLocation, rankAtConflict string) error
	Remove(ctx context.Context, heroID, conflictID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error)
}
