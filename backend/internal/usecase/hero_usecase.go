package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// HeroUseCase интерфейс бизнес-логики героев.
type HeroUseCase interface {
	CreateHero(ctx context.Context, p domain.CreateHeroParams) (string, error)
	UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error)
	DeleteHero(ctx context.Context, id string, hardDelete bool) error
	GetByID(ctx context.Context, id string) (*domain.Hero, error)
	ListHeroes(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error)
}

type heroUseCase struct {
	repo repository.HeroRepository
}

// NewHeroUseCase создает экземпляр usecase.
func NewHeroUseCase(repo repository.HeroRepository) HeroUseCase {
	return &heroUseCase{repo: repo}
}

func (uc *heroUseCase) CreateHero(ctx context.Context, p domain.CreateHeroParams) (string, error) {
	if err := p.BirthDate.IsValid(); err != nil {
		return "", fmt.Errorf("validate birth_date: %w", err)
	}

	if err := p.DeathDate.IsValid(); err != nil {
		return "", fmt.Errorf("validate death_date: %w", err)
	}

	if err := p.ServiceStartDate.IsValid(); err != nil {
		return "", fmt.Errorf("validate service_start_date: %w", err)
	}

	if err := validateHeroDates(p.BirthDate, p.DeathDate, p.ServiceStartDate); err != nil {
		return "", err
	}

	return uc.repo.Create(ctx, p)
}

func (uc *heroUseCase) UpdateHero(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
	if p.ID == nil || strings.TrimSpace(*p.ID) == "" {
		return nil, domain.ErrHeroIDRequired
	}

	normalizedMask, err := normalizeHeroFieldMask(p.FieldMask)
	if err != nil {
		return nil, err
	}
	p.FieldMask = normalizedMask

	mask := make(map[string]bool, len(p.FieldMask))
	for _, f := range p.FieldMask {
		mask[f] = true
	}

	// Защита от очистки обязательных полей
	if mask["first_name"] && isNilOrBlank(p.FirstName) {
		return nil, fmt.Errorf("first_name: %w", domain.ErrRequiredFieldEmpty)
	}

	if mask["last_name"] && isNilOrBlank(p.LastName) {
		return nil, fmt.Errorf("last_name: %w", domain.ErrRequiredFieldEmpty)
	}

	if mask["status"] {
		if p.Status == nil {
			return nil, domain.ErrStatusRequired
		}
		if *p.Status == domain.StatusUnspecified {
			return nil, domain.ErrStatusUnspecified
		}
	}

	// Валидируем новые даты, если они переданы
	if mask["birth_date"] && p.BirthDate != nil {
		if err := p.BirthDate.IsValid(); err != nil {
			return nil, fmt.Errorf("validate birth_date: %w", err)
		}
	}

	if mask["death_date"] && p.DeathDate != nil {
		if err := p.DeathDate.IsValid(); err != nil {
			return nil, fmt.Errorf("validate death_date: %w", err)
		}
	}

	if mask["service_start_date"] && p.ServiceStartDate != nil {
		if err := p.ServiceStartDate.IsValid(); err != nil {
			return nil, fmt.Errorf("validate service_start_date: %w", err)
		}
	}

	existing, err := uc.repo.GetByID(ctx, *p.ID)
	if err != nil {
		return nil, err
	}

	// Вычисляем итоговые даты для бизнес-валидации
	birth := existing.BirthDate
	death := existing.DeathDate
	serviceStart := existing.ServiceStartDate

	if mask["birth_date"] {
		birth = flexibleValueOrUnknown(p.BirthDate)
	}

	if mask["death_date"] {
		death = flexibleValueOrUnknown(p.DeathDate)
	}

	if mask["service_start_date"] {
		serviceStart = flexibleValueOrUnknown(p.ServiceStartDate)
	}

	if err := validateHeroDates(birth, death, serviceStart); err != nil {
		return nil, err
	}

	return uc.repo.Update(ctx, p)
}

func (uc *heroUseCase) DeleteHero(ctx context.Context, id string, hardDelete bool) error {
	return uc.repo.Delete(ctx, id, hardDelete)
}

func (uc *heroUseCase) GetByID(ctx context.Context, id string) (*domain.Hero, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *heroUseCase) ListHeroes(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	if f.Limit > 100 {
		f.Limit = 100
	}

	return uc.repo.List(ctx, f)
}

// normalizeHeroFieldMask приводит имена полей к каноническому виду.
//
// Это позволяет поддержке старых и новых имён из proto:
// birth_date_info -> birth_date
// death_date_info -> death_date
// service_start_date_info -> service_start_date
//
// Если одновременно указаны старый и новый вариант одного поля,
// вернём ошибку, чтобы не было двусмысленности.
func normalizeHeroFieldMask(mask []string) ([]string, error) {
	aliases := map[string]string{
		"birth_date_info":         "birth_date",
		"death_date_info":         "death_date",
		"service_start_date_info": "service_start_date",
	}

	seen := make(map[string]bool, len(mask))
	normalized := make([]string, 0, len(mask))

	for _, raw := range mask {
		f := strings.TrimSpace(raw)
		if f == "" {
			continue
		}

		if canonical, ok := aliases[f]; ok {
			f = canonical
		}

		if seen[f] {
			return nil, fmt.Errorf("%s: %w", f, domain.ErrDuplicateFieldMask)
		}

		seen[f] = true
		normalized = append(normalized, f)
	}

	return normalized, nil
}

// isNilOrBlank проверяет указатель на строку на nil или пустое значение.
func isNilOrBlank(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == ""
}

// flexibleValueOrUnknown возвращает значение или неизвестную дату.
func flexibleValueOrUnknown(d *domain.FlexibleDate) domain.FlexibleDate {
	if d == nil {
		return domain.NewUnknownDate()
	}
	return *d
}

// validateHeroDates проверяет консистентность дат героя.
// Для гибких дат сравниваем только fully exact даты.
func validateHeroDates(birth, death, serviceStart domain.FlexibleDate) error {
	b := exactAnchor(birth)
	d := exactAnchor(death)
	s := exactAnchor(serviceStart)

	if b != nil && d != nil && d.Before(*b) {
		return domain.ErrDeathBeforeBirth
	}

	if s != nil && b != nil && s.Before(*b) {
		return domain.ErrServiceBeforeBirth
	}

	if s != nil && d != nil && s.After(*d) {
		return domain.ErrServiceAfterDeath
	}

	return nil
}

// exactAnchor возвращает anchor только для точных дат.
func exactAnchor(d domain.FlexibleDate) *time.Time {
	if d.Precision != domain.PrecisionExact {
		return nil
	}

	return d.Anchor
}

// ParseDateStr парсит строку даты из proto-запроса в *time.Time.
// Поддерживает YYYY-MM-DD и RFC3339. Пустая строка возвращает nil.
func ParseDateStr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	formats := []string{
		"2006-01-02",
		time.RFC3339,
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("%s: %w", s, domain.ErrInvalidDateFormat)
}
