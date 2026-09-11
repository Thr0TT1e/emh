package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// --- Хелперы ---

// ptrString возвращает указатель на строку.
func ptrString(s string) *string { return &s }

// ptrStatus возвращает указатель на PublicationStatus.
func ptrStatus(s domain.PublicationStatus) *domain.PublicationStatus { return &s }

// ptrFlexibleDate возвращает указатель на FlexibleDate.
func ptrFlexibleDate(fd domain.FlexibleDate) *domain.FlexibleDate { return &fd }

// exactDate создаёт точную дату.
func exactDate(y, m, d int) domain.FlexibleDate {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return domain.NewExactDate(t)
}

// newValidCreateParams возвращает валидные параметры создания героя.
// DeathDate и ServiceStartDate явно заданы как "неизвестные",
// потому что zero-value FlexibleDate{} имеет PrecisionUnspecified и невалидна.
func newValidCreateParams() domain.CreateHeroParams {
	return domain.CreateHeroParams{
		FirstName:        "Иван",
		LastName:         "Петров",
		Status:           domain.StatusDraft,
		BirthDate:        exactDate(1990, 5, 15),
		DeathDate:        domain.NewUnknownDate(),
		ServiceStartDate: domain.NewUnknownDate(),
	}
}

// --- Тесты CreateHero ---

// TestHeroUseCase_CreateHero_Success проверяет успешное создание героя.
func TestHeroUseCase_CreateHero_Success(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	id, err := uc.CreateHero(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateHero failed: %v", err)
	}
	if id == "" {
		t.Error("CreateHero returned empty id")
	}
}

// TestHeroUseCase_CreateHero_InvalidBirthDate проверяет невалидную дату рождения.
func TestHeroUseCase_CreateHero_InvalidBirthDate(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	params.BirthDate = domain.FlexibleDate{Precision: domain.PrecisionExact} // без anchor

	_, err := uc.CreateHero(context.Background(), params)
	if err == nil {
		t.Fatal("CreateHero should fail with invalid birth_date")
	}
	if !errors.Is(err, domain.ErrExactDateRequiresAnchor) {
		t.Errorf("expected ErrExactDateRequiresAnchor, got %v", err)
	}
}

// TestHeroUseCase_CreateHero_DeathBeforeBirth проверяет бизнес-правило: death < birth.
func TestHeroUseCase_CreateHero_DeathBeforeBirth(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	params.BirthDate = exactDate(1990, 5, 15)
	params.DeathDate = exactDate(1985, 3, 10)         // раньше рождения
	params.ServiceStartDate = domain.NewUnknownDate() // оставляем неизвестной

	_, err := uc.CreateHero(context.Background(), params)
	if err == nil {
		t.Fatal("CreateHero should fail with death before birth")
	}
	if !errors.Is(err, domain.ErrDeathBeforeBirth) {
		t.Errorf("expected ErrDeathBeforeBirth, got %v", err)
	}
}

// TestHeroUseCase_CreateHero_ServiceBeforeBirth проверяет бизнес-правило: service < birth.
func TestHeroUseCase_CreateHero_ServiceBeforeBirth(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	params.BirthDate = exactDate(1990, 5, 15)
	params.DeathDate = domain.NewUnknownDate()      // оставляем неизвестной
	params.ServiceStartDate = exactDate(1985, 1, 1) // раньше рождения

	_, err := uc.CreateHero(context.Background(), params)
	if err == nil {
		t.Fatal("CreateHero should fail with service before birth")
	}
	if !errors.Is(err, domain.ErrServiceBeforeBirth) {
		t.Errorf("expected ErrServiceBeforeBirth, got %v", err)
	}
}

// TestHeroUseCase_CreateHero_ServiceAfterDeath проверяет бизнес-правило: service > death.
func TestHeroUseCase_CreateHero_ServiceAfterDeath(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	params.BirthDate = exactDate(1990, 5, 15)
	params.DeathDate = exactDate(2020, 8, 20)
	params.ServiceStartDate = exactDate(2021, 1, 1) // позже смерти

	_, err := uc.CreateHero(context.Background(), params)
	if err == nil {
		t.Fatal("CreateHero should fail with service after death")
	}
	if !errors.Is(err, domain.ErrServiceAfterDeath) {
		t.Errorf("expected ErrServiceAfterDeath, got %v", err)
	}
}

// TestHeroUseCase_CreateHero_NonExactDatesIgnored проверяет, что не-точные даты не сравниваются.
func TestHeroUseCase_CreateHero_NonExactDatesIgnored(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := newValidCreateParams()
	params.BirthDate = domain.FlexibleDate{Precision: domain.PrecisionYear, DisplayText: "1990"}
	params.DeathDate = domain.FlexibleDate{Precision: domain.PrecisionMonth, DisplayText: "Август 2020"}

	// Не-точные даты не должны триггерить бизнес-правила.
	id, err := uc.CreateHero(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateHero should succeed with non-exact dates: %v", err)
	}
	if id == "" {
		t.Error("CreateHero returned empty id")
	}
}

// --- Тесты UpdateHero ---

// TestHeroUseCase_UpdateHero_EmptyID проверяет, что пустой ID возвращает ошибку.
func TestHeroUseCase_UpdateHero_EmptyID(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{ID: nil}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrHeroIDRequired) {
		t.Errorf("expected ErrHeroIDRequired, got %v", err)
	}

	params.ID = ptrString("")
	_, err = uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrHeroIDRequired) {
		t.Errorf("expected ErrHeroIDRequired for blank id, got %v", err)
	}

	params.ID = ptrString("   ")
	_, err = uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrHeroIDRequired) {
		t.Errorf("expected ErrHeroIDRequired for whitespace id, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_DuplicateFieldMask проверяет дубликаты в field_mask.
func TestHeroUseCase_UpdateHero_DuplicateFieldMask(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		FieldMask: []string{"first_name", "first_name"},
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrDuplicateFieldMask) {
		t.Errorf("expected ErrDuplicateFieldMask, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_FieldMaskNormalization проверяет нормализацию алиасов.
func TestHeroUseCase_UpdateHero_FieldMaskNormalization(t *testing.T) {
	repo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return &domain.Hero{ID: id, BirthDate: exactDate(1990, 1, 1)}, nil
		},
		updateFunc: func(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
			// Проверяем, что birth_date_info нормализован в birth_date.
			if len(p.FieldMask) != 1 || p.FieldMask[0] != "birth_date" {
				t.Errorf("expected field_mask=[birth_date], got %v", p.FieldMask)
			}
			return &domain.Hero{ID: *p.ID}, nil
		},
	}
	uc := NewHeroUseCase(repo)

	newDate := exactDate(1995, 5, 15)
	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		BirthDate: ptrFlexibleDate(newDate),
		FieldMask: []string{"birth_date_info"}, // алиас
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdateHero failed: %v", err)
	}
}

// TestHeroUseCase_UpdateHero_ConflictingAliases проверяет конфликт старого и нового имён.
func TestHeroUseCase_UpdateHero_ConflictingAliases(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		FieldMask: []string{"birth_date", "birth_date_info"}, // конфликт
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrDuplicateFieldMask) {
		t.Errorf("expected ErrDuplicateFieldMask for conflicting aliases, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_ClearRequiredField проверяет защиту от очистки first_name.
func TestHeroUseCase_UpdateHero_ClearRequiredField(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		FirstName: ptrString(""),
		FieldMask: []string{"first_name"},
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrRequiredFieldEmpty) {
		t.Errorf("expected ErrRequiredFieldEmpty for first_name, got %v", err)
	}

	params.FirstName = nil
	_, err = uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrRequiredFieldEmpty) {
		t.Errorf("expected ErrRequiredFieldEmpty for nil first_name, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_StatusUnspecified проверяет запрет UNSPECIFIED статуса.
func TestHeroUseCase_UpdateHero_StatusUnspecified(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		Status:    ptrStatus(domain.StatusUnspecified),
		FieldMask: []string{"status"},
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrStatusUnspecified) {
		t.Errorf("expected ErrStatusUnspecified, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_StatusRequired проверяет, что status обязателен при обновлении.
func TestHeroUseCase_UpdateHero_StatusRequired(t *testing.T) {
	repo := &mockHeroRepository{}
	uc := NewHeroUseCase(repo)

	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		Status:    nil,
		FieldMask: []string{"status"},
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrStatusRequired) {
		t.Errorf("expected ErrStatusRequired, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_DeathBeforeBirth проверяет бизнес-правило при partial update.
func TestHeroUseCase_UpdateHero_DeathBeforeBirth(t *testing.T) {
	repo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return &domain.Hero{
				ID:        id,
				BirthDate: exactDate(1990, 5, 15),
				DeathDate: domain.NewUnknownDate(),
			}, nil
		},
	}
	uc := NewHeroUseCase(repo)

	newDeath := exactDate(1985, 3, 10) // раньше существующего birth
	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		DeathDate: ptrFlexibleDate(newDeath),
		FieldMask: []string{"death_date"},
	}
	_, err := uc.UpdateHero(context.Background(), params)
	if !errors.Is(err, domain.ErrDeathBeforeBirth) {
		t.Errorf("expected ErrDeathBeforeBirth, got %v", err)
	}
}

// TestHeroUseCase_UpdateHero_Success проверяет успешное обновление.
func TestHeroUseCase_UpdateHero_Success(t *testing.T) {
	repo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return &domain.Hero{ID: id, BirthDate: exactDate(1990, 5, 15)}, nil
		},
	}
	uc := NewHeroUseCase(repo)

	newFirstName := "Пётр"
	params := domain.UpdateHeroParams{
		ID:        ptrString("hero-id"),
		FirstName: ptrString(newFirstName),
		FieldMask: []string{"first_name"},
	}
	hero, err := uc.UpdateHero(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdateHero failed: %v", err)
	}
	if hero == nil {
		t.Fatal("UpdateHero returned nil hero")
	}
}

// --- Тесты DeleteHero ---

// TestHeroUseCase_DeleteHero_Success проверяет делегирование в репозиторий.
func TestHeroUseCase_DeleteHero_Success(t *testing.T) {
	var calledHardDelete bool
	repo := &mockHeroRepository{
		deleteFunc: func(ctx context.Context, id string, hardDelete bool) error {
			if id != "hero-id" {
				t.Errorf("expected id=hero-id, got %s", id)
			}
			calledHardDelete = hardDelete
			return nil
		},
	}
	uc := NewHeroUseCase(repo)

	err := uc.DeleteHero(context.Background(), "hero-id", true)
	if err != nil {
		t.Fatalf("DeleteHero failed: %v", err)
	}
	if !calledHardDelete {
		t.Error("DeleteHero did not pass hardDelete=true")
	}
}

// --- Тесты ListHeroes ---

// TestHeroUseCase_ListHeroes_NormalizesLimit проверяет нормализацию limit.
func TestHeroUseCase_ListHeroes_NormalizesLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero becomes 20", 0, 20},
		{"negative becomes 20", -5, 20},
		{"valid limit preserved", 50, 50},
		{"max 100", 150, 100},
		{"exactly 100", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLimit int
			repo := &mockHeroRepository{
				listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
					capturedLimit = f.Limit
					return []*domain.Hero{}, "", 0, nil
				},
			}
			uc := NewHeroUseCase(repo)

			filter := domain.HeroFilter{Limit: tt.input}
			_, _, _, err := uc.ListHeroes(context.Background(), filter)
			if err != nil {
				t.Fatalf("ListHeroes failed: %v", err)
			}
			if capturedLimit != tt.expected {
				t.Errorf("expected limit=%d, got %d", tt.expected, capturedLimit)
			}
		})
	}
}

// --- Тесты вспомогательных функций ---

// TestNormalizeHeroFieldMask_Success проверяет нормализацию без ошибок.
func TestNormalizeHeroFieldMask_Success(t *testing.T) {
	mask := []string{"first_name", "birth_date_info", "status"}
	normalized, err := normalizeHeroFieldMask(mask)
	if err != nil {
		t.Fatalf("normalizeHeroFieldMask failed: %v", err)
	}

	expected := []string{"first_name", "birth_date", "status"}
	if len(normalized) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, normalized)
	}
	for i := range normalized {
		if normalized[i] != expected[i] {
			t.Errorf("normalized[%d] = %q, want %q", i, normalized[i], expected[i])
		}
	}
}

// TestNormalizeHeroFieldMask_EmptyAndWhitespace проверяет фильтрацию пустых значений.
func TestNormalizeHeroFieldMask_EmptyAndWhitespace(t *testing.T) {
	mask := []string{"first_name", "", "  ", "last_name"}
	normalized, err := normalizeHeroFieldMask(mask)
	if err != nil {
		t.Fatalf("normalizeHeroFieldMask failed: %v", err)
	}

	if len(normalized) != 2 {
		t.Errorf("expected 2 fields, got %d: %v", len(normalized), normalized)
	}
}

// TestParseDateStr_ValidFormats проверяет парсинг валидных форматов.
func TestParseDateStr_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		year  int
		month time.Month
		day   int
	}{
		{"YYYY-MM-DD", "2024-03-15", 2024, time.March, 15},
		{"RFC3339", "2024-03-15T10:30:00Z", 2024, time.March, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseDateStr(tt.input)
			if err != nil {
				t.Fatalf("ParseDateStr(%q) failed: %v", tt.input, err)
			}
			if parsed == nil {
				t.Fatal("ParseDateStr returned nil")
			}
			if parsed.Year() != tt.year || parsed.Month() != tt.month || parsed.Day() != tt.day {
				t.Errorf("parsed date = %v, want %d-%02d-%02d", parsed, tt.year, tt.month, tt.day)
			}
		})
	}
}

// TestParseDateStr_EmptyString проверяет, что пустая строка возвращает nil.
func TestParseDateStr_EmptyString(t *testing.T) {
	parsed, err := ParseDateStr("")
	if err != nil {
		t.Fatalf("ParseDateStr(\"\") failed: %v", err)
	}
	if parsed != nil {
		t.Errorf("ParseDateStr(\"\") = %v, want nil", parsed)
	}
}

// TestParseDateStr_InvalidFormat проверяет невалидный формат.
func TestParseDateStr_InvalidFormat(t *testing.T) {
	_, err := ParseDateStr("15.03.2024")
	if err == nil {
		t.Fatal("ParseDateStr should fail for invalid format")
	}
	if !errors.Is(err, domain.ErrInvalidDateFormat) {
		t.Errorf("expected ErrInvalidDateFormat, got %v", err)
	}
}

// TestValidateHeroDates_AllValid проверяет валидные комбинации.
func TestValidateHeroDates_AllValid(t *testing.T) {
	birth := exactDate(1990, 5, 15)
	death := exactDate(2020, 8, 20)
	service := exactDate(2010, 9, 1)

	err := validateHeroDates(birth, death, service)
	if err != nil {
		t.Errorf("validateHeroDates returned error for valid dates: %v", err)
	}
}

// TestValidateHeroDates_SameDay проверяет граничный случай: все даты в один день.
func TestValidateHeroDates_SameDay(t *testing.T) {
	sameDay := exactDate(2020, 8, 20)

	err := validateHeroDates(sameDay, sameDay, sameDay)
	if err != nil {
		t.Errorf("validateHeroDates should allow same day: %v", err)
	}
}
