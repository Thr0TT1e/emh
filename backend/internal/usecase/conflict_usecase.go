package usecase

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// ConflictUseCase бизнес-логика справочника конфликтов.
type ConflictUseCase interface {
	Create(ctx context.Context, p domain.CreateConflictParams) (string, error)
	Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Conflict, error)
	List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error)
}

type conflictUseCase struct {
	repo repository.ConflictRepository
}

func NewConflictUseCase(repo repository.ConflictRepository) ConflictUseCase {
	return &conflictUseCase{repo: repo}
}

func (uc *conflictUseCase) Create(ctx context.Context, p domain.CreateConflictParams) (string, error) {
	if p.Name == "" {
		return "", domain.ErrConflictNameRequired
	}
	if err := validateConflictDates(p.StartDate, p.EndDate); err != nil {
		return "", err
	}
	// Проверяем существование родителя для понятной ошибки (иначе FK violation -> CodeInternal)
	if p.ParentConflictID != nil && *p.ParentConflictID != "" {
		if _, err := uc.repo.GetByID(ctx, *p.ParentConflictID); err != nil {
			return "", fmt.Errorf("parent conflict: %w", err)
		}
	}
	return uc.repo.Create(ctx, p)
}

func (uc *conflictUseCase) Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error) {
	existing, err := uc.repo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	// Вычисляем итоговые даты с учётом field_mask для бизнес-валидации
	start := existing.StartDate
	end := existing.EndDate
	for _, f := range p.FieldMask {
		switch f {
		case "start_date":
			start = p.StartDate
		case "end_date":
			end = p.EndDate
		}
	}
	if err := validateConflictDates(start, end); err != nil {
		return nil, err
	}

	// Валидация родителя при его обновлении
	if hasField(p.FieldMask, "parent_conflict_id") && p.ParentConflictID != nil && *p.ParentConflictID != "" {
		if *p.ParentConflictID == p.ID {
			return nil, domain.ErrOwnParent
		}
		if _, err := uc.repo.GetByID(ctx, *p.ParentConflictID); err != nil {
			return nil, fmt.Errorf("parent conflict: %w", err)
		}
		// Защита от транзитивных циклов (A → B → C → A).
		hasCycle, err := uc.repo.HasCyclicReference(ctx, p.ID, *p.ParentConflictID)
		if err != nil {
			return nil, fmt.Errorf("check cyclic reference: %w", err)
		}
		if hasCycle {
			return nil, domain.ErrCyclicReference
		}
	}

	return uc.repo.Update(ctx, p)
}

func (uc *conflictUseCase) Delete(ctx context.Context, id string) error {
	hasHeroes, err := uc.repo.HasHeroes(ctx, id)
	if err != nil {
		return fmt.Errorf("check heroes: %w", err)
	}
	if hasHeroes {
		return domain.ErrEntityAssignedToHeroes
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *conflictUseCase) GetByID(ctx context.Context, id string) (*domain.Conflict, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *conflictUseCase) List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return uc.repo.List(ctx, f)
}

// validateConflictDates проверяет, что дата окончания не раньше даты начала.
func validateConflictDates(start, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return domain.ErrConflictDateInvalid
	}
	return nil
}

// hasField проверяет наличие поля в маске обновления.
func hasField(mask []string, field string) bool {
	for _, f := range mask {
		if f == field {
			return true
		}
	}
	return false
}
