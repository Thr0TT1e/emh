package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroConflictUseCase бизнес-логика привязки конфликтов к героям.
type HeroConflictUseCase interface {
	Add(ctx context.Context, heroID, conflictID, specificLocation, rankAtConflict string) error
	Remove(ctx context.Context, heroID, conflictID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error)
}

type heroConflictUseCase struct {
	repo         repository.HeroConflictRepository
	heroRepo     repository.HeroRepository
	conflictRepo repository.ConflictRepository
}

func NewHeroConflictUseCase(
	repo repository.HeroConflictRepository,
	heroRepo repository.HeroRepository,
	conflictRepo repository.ConflictRepository,
) HeroConflictUseCase {
	return &heroConflictUseCase{
		repo:         repo,
		heroRepo:     heroRepo,
		conflictRepo: conflictRepo,
	}
}

func (uc *heroConflictUseCase) Add(ctx context.Context, heroID, conflictID, specificLocation, rankAtConflict string) error {
	// Проверяем существование героя и конфликта для понятных ошибок
	if _, err := uc.heroRepo.GetByID(ctx, heroID); err != nil {
		return fmt.Errorf("hero not found: %w", err)
	}
	if _, err := uc.conflictRepo.GetByID(ctx, conflictID); err != nil {
		return fmt.Errorf("conflict not found: %w", err)
	}
	return uc.repo.Add(ctx, heroID, conflictID, specificLocation, rankAtConflict)
}

func (uc *heroConflictUseCase) Remove(ctx context.Context, heroID, conflictID string) error {
	return uc.repo.Remove(ctx, heroID, conflictID)
}

func (uc *heroConflictUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
