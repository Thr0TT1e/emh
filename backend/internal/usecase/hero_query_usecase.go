package usecase

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroQueryUseCase — read-модель для сборки полного агрегата героя (CQRS).
type HeroQueryUseCase interface {
	GetHeroDetail(ctx context.Context, id string) (*domain.HeroDetail, error)
	ListHeroPhotos(ctx context.Context, heroID string) ([]*domain.Photo, error)
	// ListHeroPhotosPaged возвращает фото героя с курсорной пагинацией (публичный API).
	ListHeroPhotosPaged(ctx context.Context, heroID, cursor string, limit int) ([]*domain.Photo, string, int64, error)
}

type heroQueryUseCase struct {
	heroRepo     repository.HeroRepository
	photoRepo    repository.PhotoRepository
	awardRepo    repository.HeroAwardRepository
	conflictRepo repository.HeroConflictRepository
	locationRepo repository.HeroLocationRepository
	sourceRepo   repository.HeroSourceRepository
	relationRepo repository.HeroRelationRepository
}

func NewHeroQueryUseCase(
	heroRepo repository.HeroRepository,
	photoRepo repository.PhotoRepository,
	awardRepo repository.HeroAwardRepository,
	conflictRepo repository.HeroConflictRepository,
	locationRepo repository.HeroLocationRepository,
	sourceRepo repository.HeroSourceRepository,
	relationRepo repository.HeroRelationRepository,
) HeroQueryUseCase {
	return &heroQueryUseCase{
		heroRepo:     heroRepo,
		photoRepo:    photoRepo,
		awardRepo:    awardRepo,
		conflictRepo: conflictRepo,
		locationRepo: locationRepo,
		sourceRepo:   sourceRepo,
		relationRepo: relationRepo,
	}
}

func (uc *heroQueryUseCase) GetHeroDetail(ctx context.Context, id string) (*domain.HeroDetail, error) {
	// Сначала загружаем корень агрегата. Если героя нет — дальше нет смысла.
	hero, err := uc.heroRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &domain.HeroDetail{Hero: hero}

	// Все связи независимы — загружаем параллельно.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		photos, err := uc.photoRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load photos: %w", err)
		}
		detail.Photos = photos
		return nil
	})

	g.Go(func() error {
		awards, err := uc.awardRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load awards: %w", err)
		}
		detail.Awards = awards
		return nil
	})

	g.Go(func() error {
		conflicts, err := uc.conflictRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load conflicts: %w", err)
		}
		detail.Conflicts = conflicts
		return nil
	})

	g.Go(func() error {
		locations, err := uc.locationRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load locations: %w", err)
		}
		detail.Locations = locations
		return nil
	})

	g.Go(func() error {
		sources, err := uc.sourceRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load sources: %w", err)
		}
		detail.Sources = sources
		return nil
	})

	g.Go(func() error {
		relations, err := uc.relationRepo.ListByHero(gctx, id)
		if err != nil {
			return fmt.Errorf("load relations: %w", err)
		}
		detail.Relations = relations
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return detail, nil
}

func (uc *heroQueryUseCase) ListHeroPhotos(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	return uc.photoRepo.ListByHero(ctx, heroID)
}

func (uc *heroQueryUseCase) ListHeroPhotosPaged(
	ctx context.Context,
	heroID, cursor string,
	limit int,
) ([]*domain.Photo, string, int64, error) {
	return uc.photoRepo.ListByHeroPaged(ctx, heroID, cursor, limit)
}
