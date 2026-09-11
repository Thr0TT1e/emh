package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroSourceUseCase бизнес-логика источников данных героя.
type HeroSourceUseCase interface {
	Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error)
	Remove(ctx context.Context, heroID, sourceID string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error)
}

type heroSourceUseCase struct {
	repo     repository.HeroSourceRepository
	heroRepo repository.HeroRepository
}

func NewHeroSourceUseCase(repo repository.HeroSourceRepository, heroRepo repository.HeroRepository) HeroSourceUseCase {
	return &heroSourceUseCase{repo: repo, heroRepo: heroRepo}
}

func (uc *heroSourceUseCase) Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error) {
	if p.URL == "" {
		return "", fmt.Errorf("source url is required")
	}
	if _, err := uc.heroRepo.GetByID(ctx, p.HeroID); err != nil {
		return "", fmt.Errorf("hero not found: %w", err)
	}
	return uc.repo.Add(ctx, p)
}

func (uc *heroSourceUseCase) Remove(ctx context.Context, heroID, sourceID string) error {
	return uc.repo.Remove(ctx, heroID, sourceID)
}

func (uc *heroSourceUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
