package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type HeroAwardUseCase interface {
	Add(ctx context.Context, heroID, awardID string, awardDate *string, decreeNumber string) error
	Remove(ctx context.Context, heroID, awardID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error)
}

type heroAwardUseCase struct {
	repo      repository.HeroAwardRepository
	heroRepo  repository.HeroRepository
	awardRepo repository.AwardRepository
}

func NewHeroAwardUseCase(
	repo repository.HeroAwardRepository,
	heroRepo repository.HeroRepository,
	awardRepo repository.AwardRepository,
) HeroAwardUseCase {
	return &heroAwardUseCase{
		repo:      repo,
		heroRepo:  heroRepo,
		awardRepo: awardRepo,
	}
}

func (uc *heroAwardUseCase) Add(ctx context.Context, heroID, awardID string, awardDate *string, decreeNumber string) error {
	// Проверяем существование героя
	if _, err := uc.heroRepo.GetByID(ctx, heroID); err != nil {
		return fmt.Errorf("hero not found: %w", err)
	}
	// Проверяем существование награды
	if _, err := uc.awardRepo.GetByID(ctx, awardID); err != nil {
		return fmt.Errorf("award not found: %w", err)
	}
	return uc.repo.Add(ctx, heroID, awardID, awardDate, decreeNumber)
}

func (uc *heroAwardUseCase) Remove(ctx context.Context, heroID, awardID string) error {
	return uc.repo.Remove(ctx, heroID, awardID)
}

func (uc *heroAwardUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
