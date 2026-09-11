package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// LocationRepository абстракция для работы с локациями.
type LocationRepository interface {
	Create(ctx context.Context, p domain.CreateLocationParams) (string, error)
	Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Location, error)
	List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error)
	Reparent(ctx context.Context, oldParentID, newParentID string) (int, error)
	HasHeroes(ctx context.Context, locationID string) (bool, error)
	// HasCyclicReference проверяет, приведёт ли установка parentID = candidateParentID
	// для локации selfID к циклу в иерархии.
	// Возвращает true, если candidateParentID является потомком selfID
	// (т.е. selfID уже находится в цепочке предков candidateParentID).
	HasCyclicReference(ctx context.Context, selfID, candidateParentID string) (bool, error)
}

// HeroLocationRepository абстракция для связи герой-локация.
type HeroLocationRepository interface {
	Add(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error
	Remove(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error)
}
