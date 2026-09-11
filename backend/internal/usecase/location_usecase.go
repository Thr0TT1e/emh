package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// LocationUseCase бизнес-логика справочника мест.
type LocationUseCase interface {
	Create(ctx context.Context, p domain.CreateLocationParams) (string, error)
	Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Location, error)
	List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error)
	Reparent(ctx context.Context, oldParentID, newParentID string) (int, error)
}

type locationUseCase struct {
	repo repository.LocationRepository
}

func NewLocationUseCase(repo repository.LocationRepository) LocationUseCase {
	return &locationUseCase{repo: repo}
}

func (uc *locationUseCase) Create(ctx context.Context, p domain.CreateLocationParams) (string, error) {
	if p.Name == "" {
		return "", domain.ErrLocationNameRequired
	}
	if err := validateCoordinates(p.Latitude, p.Longitude); err != nil {
		return "", err
	}
	// Проверяем существование родителя для понятной ошибки
	if p.ParentID != nil && *p.ParentID != "" {
		if _, err := uc.repo.GetByID(ctx, *p.ParentID); err != nil {
			return "", fmt.Errorf("parent location: %w", err)
		}
	}
	return uc.repo.Create(ctx, p)
}

func (uc *locationUseCase) Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error) {
	existing, err := uc.repo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	// Вычисляем итоговые координаты с учётом field_mask
	lat := existing.Latitude
	lon := existing.Longitude
	for _, f := range p.FieldMask {
		switch f {
		case "latitude":
			lat = p.Latitude
		case "longitude":
			lon = p.Longitude
		}
	}
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	// Валидация родителя при его обновлении
	if hasField(p.FieldMask, "parent_id") && p.ParentID != nil && *p.ParentID != "" {
		if *p.ParentID == p.ID {
			return nil, domain.ErrOwnParent
		}
		if _, err := uc.repo.GetByID(ctx, *p.ParentID); err != nil {
			return nil, fmt.Errorf("parent location: %w", err)
		}
		// Защита от транзитивных циклов (A → B → C → A).
		// Проверяем, не является ли обновляемая локация потомком нового родителя.
		hasCycle, err := uc.repo.HasCyclicReference(ctx, p.ID, *p.ParentID)
		if err != nil {
			return nil, fmt.Errorf("check cyclic reference: %w", err)
		}
		if hasCycle {
			return nil, domain.ErrCyclicReference
		}
	}

	return uc.repo.Update(ctx, p)
}

func (uc *locationUseCase) Delete(ctx context.Context, id string) error {
	hasHeroes, err := uc.repo.HasHeroes(ctx, id)
	if err != nil {
		return fmt.Errorf("check heroes: %w", err)
	}
	if hasHeroes {
		return domain.ErrEntityAssignedToHeroes
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *locationUseCase) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *locationUseCase) List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return uc.repo.List(ctx, f)
}

func (uc *locationUseCase) Reparent(ctx context.Context, oldParentID, newParentID string) (int, error) {
	// Новый родитель должен существовать, если задан (пустая строка = сделать корневыми)
	if newParentID != "" {
		if _, err := uc.repo.GetByID(ctx, newParentID); err != nil {
			return 0, fmt.Errorf("new parent location not found: %w", err)
		}
	}
	return uc.repo.Reparent(ctx, oldParentID, newParentID)
}

// validateCoordinates проверяет допустимость географических координат.
func validateCoordinates(lat, lon *float64) error {
	if lat != nil && (*lat < -90 || *lat > 90) {
		return domain.ErrInvalidLatitude
	}
	if lon != nil && (*lon < -180 || *lon > 180) {
		return domain.ErrInvalidLongitude
	}
	return nil
}
