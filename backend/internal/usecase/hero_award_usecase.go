package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type HeroAwardUseCase interface {
	Add(ctx context.Context, p domain.AddHeroAwardParams) error
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

func (uc *heroAwardUseCase) Add(ctx context.Context, p domain.AddHeroAwardParams) error {
	// Проверяем существование героя
	if _, err := uc.heroRepo.GetByID(ctx, p.HeroID); err != nil {
		return fmt.Errorf("hero not found: %w", err)
	}
	// Проверяем существование награды
	if _, err := uc.awardRepo.GetByID(ctx, p.AwardID); err != nil {
		return fmt.Errorf("award not found: %w", err)
	}
	if p.AwardDateInfo != nil {
		if err := p.AwardDateInfo.IsValid(); err != nil {
			return fmt.Errorf("validate award_date_info: %w", err)
		}
	}
	return uc.repo.Add(ctx, p)
}

func (uc *heroAwardUseCase) Remove(ctx context.Context, heroID, awardID string) error {
	return uc.repo.Remove(ctx, heroID, awardID)
}

func (uc *heroAwardUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
