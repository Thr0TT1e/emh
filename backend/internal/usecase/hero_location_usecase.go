package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroLocationUseCase бизнес-логика привязки локаций к героям.
type HeroLocationUseCase interface {
	Add(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error
	Remove(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error)
}

type heroLocationUseCase struct {
	repo         repository.HeroLocationRepository
	heroRepo     repository.HeroRepository
	locationRepo repository.LocationRepository
}

func NewHeroLocationUseCase(
	repo repository.HeroLocationRepository,
	heroRepo repository.HeroRepository,
	locationRepo repository.LocationRepository,
) HeroLocationUseCase {
	return &heroLocationUseCase{
		repo:         repo,
		heroRepo:     heroRepo,
		locationRepo: locationRepo,
	}
}

func (uc *heroLocationUseCase) Add(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	if locType == domain.HeroLocationTypeUnspecified {
		return fmt.Errorf("location type is required")
	}
	// Проверяем существование героя и локации для понятных ошибок
	if _, err := uc.heroRepo.GetByID(ctx, heroID); err != nil {
		return fmt.Errorf("hero not found: %w", err)
	}
	if _, err := uc.locationRepo.GetByID(ctx, locationID); err != nil {
		return fmt.Errorf("location not found: %w", err)
	}
	return uc.repo.Add(ctx, heroID, locationID, locType)
}

func (uc *heroLocationUseCase) Remove(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	return uc.repo.Remove(ctx, heroID, locationID, locType)
}

func (uc *heroLocationUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
