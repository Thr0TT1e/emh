package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroRelationUseCase бизнес-логика связей между героями.
type HeroRelationUseCase interface {
	Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error)
	Remove(ctx context.Context, id string) error
	ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error)
}

type heroRelationUseCase struct {
	repo     repository.HeroRelationRepository
	heroRepo repository.HeroRepository
}

func NewHeroRelationUseCase(repo repository.HeroRelationRepository, heroRepo repository.HeroRepository) HeroRelationUseCase {
	return &heroRelationUseCase{repo: repo, heroRepo: heroRepo}
}

func (uc *heroRelationUseCase) Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
	if p.FromHeroID == p.ToHeroID {
		return "", fmt.Errorf("hero cannot be related to itself")
	}
	// Проверяем существование обоих героев
	if _, err := uc.heroRepo.GetByID(ctx, p.FromHeroID); err != nil {
		return "", fmt.Errorf("source hero not found: %w", err)
	}
	if _, err := uc.heroRepo.GetByID(ctx, p.ToHeroID); err != nil {
		return "", fmt.Errorf("target hero not found: %w", err)
	}
	return uc.repo.Add(ctx, p)
}

func (uc *heroRelationUseCase) Remove(ctx context.Context, id string) error {
	return uc.repo.Remove(ctx, id)
}

func (uc *heroRelationUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
	return uc.repo.ListByHero(ctx, heroID)
}
