package usecase

import (
	"context"
	"fmt"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type AwardUseCase interface {
	Create(ctx context.Context, p domain.CreateAwardParams) (string, error)
	Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Award, error)
	List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error)
}

type awardUseCase struct {
	repo repository.AwardRepository
}

func NewAwardUseCase(repo repository.AwardRepository) AwardUseCase {
	return &awardUseCase{repo: repo}
}

func (uc *awardUseCase) Create(ctx context.Context, p domain.CreateAwardParams) (string, error) {
	if p.Name == "" {
		return "", domain.ErrAwardNameRequired
	}
	return uc.repo.Create(ctx, p)
}

func (uc *awardUseCase) Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error) {
	return uc.repo.Update(ctx, p)
}

func (uc *awardUseCase) Delete(ctx context.Context, id string) error {
	// Проверяем, привязана ли награда к героям
	hasHeroes, err := uc.repo.HasHeroes(ctx, id)
	if err != nil {
		return fmt.Errorf("check heroes: %w", err)
	}
	if hasHeroes {
		return domain.ErrEntityAssignedToHeroes
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *awardUseCase) GetByID(ctx context.Context, id string) (*domain.Award, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *awardUseCase) List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return uc.repo.List(ctx, f)
}
