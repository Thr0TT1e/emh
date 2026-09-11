package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

const limit = 1024

// ─── Хелперы ────────────────────────────────────────────────────────────────

// validExtractionJSON валидный JSON-ответ от LLM.
// Поля соответствуют json-тегам доменной структуры domain.ExtractedHero.
func validExtractionJSON() string {
	return `{
		"hero": {
			"last_name": "Иванов",
			"first_name": "Иван",
			"middle_name": "Иванович",
			"rank": "Сержант",
			"unit": "3-й батальон",
			"position": "Командир отделения",
			"birth_date": "1920-05-15",
			"death_date": "1944-08-20",
			"short_bio": "Краткая биография",
			"full_bio": "Полная биография героя"
		},
		"conflicts": [
			{
				"name": "Великая Отечественная война",
				"specific_location": "Сталинград"
			}
		],
		"awards": [],
		"locations": [],
		"warnings": []
	}`
}

// ─── Тесты ExtractFromText ──────────────────────────────────────────────────

// TestExtractFromText_Success проверяет успешный пайплайн извлечения.
func TestExtractFromText_Success(t *testing.T) {
	provider := &mockLLMProvider{
		nameFn:  func() string { return "ollama_local" },
		modelFn: func() string { return "llama3.1:8b" },
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return validExtractionJSON(), nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	result, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() error = %v", err)
	}
	if result == nil {
		t.Fatal("ExtractFromText() should return result")
	}
	if result.Hero == nil {
		t.Fatal("ExtractFromText() result.Hero should not be nil")
	}
	if result.Hero.LastName != "Иванов" {
		t.Errorf("hero.LastName = %q, want %q", result.Hero.LastName, "Иванов")
	}
	if result.Hero.FirstName != "Иван" {
		t.Errorf("hero.FirstName = %q, want %q", result.Hero.FirstName, "Иван")
	}
	if len(result.Conflicts) != 1 {
		t.Errorf("conflicts count = %d, want 1", len(result.Conflicts))
	}
}

// TestExtractFromText_InputTooLarge_Boundary проверяет лимит символов (рун)
// с точными граничными значениями.
func TestExtractFromText_InputTooLarge_Boundary(t *testing.T) {

	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return validExtractionJSON(), nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}
	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "below limit (ASCII)",
			text:    strings.Repeat("x", limit-1),
			wantErr: false,
		},
		{
			name:    "exactly at limit (ASCII)",
			text:    strings.Repeat("x", limit),
			wantErr: false,
		},
		{
			name:    "one char over (ASCII)",
			text:    strings.Repeat("x", limit+1),
			wantErr: true,
		},
		{
			name:    "cyrillic below limit",
			text:    strings.Repeat("а", limit-1), // "а" = 1 руна, 2 байта
			wantErr: false,
		},
		{
			name:    "cyrillic exactly at limit",
			text:    strings.Repeat("а", limit),
			wantErr: false,
		},
		{
			name:    "cyrillic one char over",
			text:    strings.Repeat("а", limit+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.ExtractFromText(context.Background(), tt.text, "")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, domain.ErrLLMInputTooLarge) {
					t.Errorf("want ErrLLMInputTooLarge, got: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestExtractFromText_ProviderNil проверяет обработку отключённого LLM.
func TestExtractFromText_ProviderNil(t *testing.T) {
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(nil, logRepo, heroRepo, limit, slog.Default())

	_, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err == nil {
		t.Error("ExtractFromText() should return error when provider is nil")
	}
	if !errors.Is(err, domain.ErrLLMDisabled) {
		t.Errorf("ExtractFromText() error = %v, want ErrLLMDisabled", err)
	}
}

// TestExtractFromText_GenerateError проверяет обработку ошибки генерации.
func TestExtractFromText_GenerateError(t *testing.T) {
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return "", errors.New("connection timeout")
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	_, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err == nil {
		t.Error("ExtractFromText() should return error on generate failure")
	}
}

// TestExtractFromText_InvalidJSON проверяет обработку невалидного JSON от LLM.
func TestExtractFromText_InvalidJSON(t *testing.T) {
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return `{invalid json}`, nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	_, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err == nil {
		t.Error("ExtractFromText() should return error for invalid JSON")
	}
}

// TestExtractFromText_AuditLogCreated проверяет, что аудит-лог создаётся.
func TestExtractFromText_AuditLogCreated(t *testing.T) {
	var auditLogCreated bool
	var auditLogProvider string
	provider := &mockLLMProvider{
		nameFn:  func() string { return "ollama_local" },
		modelFn: func() string { return "llama3.1:8b" },
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return validExtractionJSON(), nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{
		createFn: func(ctx context.Context, log *domain.LLMExtractionLog) (string, error) {
			auditLogCreated = true
			auditLogProvider = log.Provider
			return "log-id", nil
		},
	}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	_, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() error = %v", err)
	}
	if !auditLogCreated {
		t.Error("ExtractFromText() should create audit log")
	}
	if auditLogProvider != "ollama_local" {
		t.Errorf("audit log provider = %q, want %q", auditLogProvider, "ollama_local")
	}
}

// TestExtractFromText_ContextCanceled проверяет обработку отмены контекста.
func TestExtractFromText_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем сразу.

	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return "", ctx.Err()
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	_, err := uc.ExtractFromText(ctx, "Текст о герое", "")
	if err == nil {
		t.Error("ExtractFromText() should return error on canceled context")
	}
}

// ─── Тесты маппинга ─────────────────────────────────────────────────────────

// TestExtractFromText_MappingFields проверяет корректность маппинга полей.
func TestExtractFromText_MappingFields(t *testing.T) {
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return validExtractionJSON(), nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	result, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() error = %v", err)
	}

	// Проверяем маппинг основных полей героя.
	if result.Hero.Rank != "Сержант" {
		t.Errorf("hero.Rank = %q, want %q", result.Hero.Rank, "Сержант")
	}
	if result.Hero.Unit != "3-й батальон" {
		t.Errorf("hero.Unit = %q, want %q", result.Hero.Unit, "3-й батальон")
	}
	if result.Hero.Position != "Командир отделения" {
		t.Errorf("hero.Position = %q, want %q", result.Hero.Position, "Командир отделения")
	}
	if result.Hero.ShortBio != "Краткая биография" {
		t.Errorf("hero.ShortBio = %q, want %q", result.Hero.ShortBio, "Краткая биография")
	}

	// Проверяем маппинг конфликтов.
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts count = %d, want 1", len(result.Conflicts))
	}
	if result.Conflicts[0].Name != "Великая Отечественная война" {
		t.Errorf("conflict.Name = %q, want %q", result.Conflicts[0].Name, "Великая Отечественная война")
	}
}

// ─── Тесты валидации входных данных ─────────────────────────────────────────

// TestExtractFromText_JSONWithExtraFields проверяет устойчивость к лишним полям в JSON.
func TestExtractFromText_JSONWithExtraFields(t *testing.T) {
	jsonWithExtra := `{
		"hero": {
			"last_name": "Иванов",
			"first_name": "Иван",
			"unknown_field": "should be ignored"
		},
		"conflicts": [],
		"awards": [],
		"locations": [],
		"extra_top_level": "ignored",
		"warnings": []
	}`
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return jsonWithExtra, nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	result, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() should handle extra fields gracefully, got error: %v", err)
	}
	if result.Hero.LastName != "Иванов" {
		t.Errorf("hero.LastName = %q, want %q", result.Hero.LastName, "Иванов")
	}
}

// ─── Проверка структуры домена ──────────────────────────────────────────────

// TestExtractionResult_JSONRoundTrip проверяет корректность сериализации.
func TestExtractionResult_JSONRoundTrip(t *testing.T) {
	original := &domain.ExtractionResult{
		Hero: &domain.ExtractedHero{
			LastName:  "Иванов",
			FirstName: "Иван",
			Rank:      "Сержант",
			Unit:      "3-й батальон",
		},
		Conflicts: []domain.ExtractedConflict{
			{Name: "Великая Отечественная война"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded domain.ExtractionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Hero == nil {
		t.Fatal("decoded.Hero should not be nil")
	}
	if decoded.Hero.LastName != original.Hero.LastName {
		t.Errorf("roundtrip hero.LastName = %q, want %q", decoded.Hero.LastName, original.Hero.LastName)
	}
	if len(decoded.Conflicts) != len(original.Conflicts) {
		t.Errorf("roundtrip conflicts count = %d, want %d", len(decoded.Conflicts), len(original.Conflicts))
	}
}

// ─── Тесты поиска дубликатов ────────────────────────────────────────────────

// TestExtractFromText_FindDuplicates проверяет, что при совпадении ФИО
// в БД находятся кандидаты на дубликаты.
func TestExtractFromText_FindDuplicates(t *testing.T) {
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return validExtractionJSON(), nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{
		listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
			// Имитируем найденного дубликата
			return []*domain.Hero{
				{ID: "existing-hero-1", LastName: "Иванов", FirstName: "Иван"},
			}, "", 1, nil
		},
	}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	result, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() error = %v", err)
	}
	if len(result.DuplicateHeroIDs) != 1 {
		t.Errorf("DuplicateHeroIDs count = %d, want 1", len(result.DuplicateHeroIDs))
	}
	if len(result.DuplicateHeroIDs) > 0 && result.DuplicateHeroIDs[0] != "existing-hero-1" {
		t.Errorf("DuplicateHeroIDs[0] = %q, want %q", result.DuplicateHeroIDs[0], "existing-hero-1")
	}
}

// TestExtractFromText_NoDuplicatesWhenHeroEmpty проверяет, что поиск дубликатов
// не вызывается, если LLM не вернул героя.
func TestExtractFromText_NoDuplicatesWhenHeroEmpty(t *testing.T) {
	var listCalled bool
	provider := &mockLLMProvider{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return `{"hero": null, "conflicts": [], "awards": [], "locations": [], "warnings": []}`, nil
		},
	}
	logRepo := &mockLLMExtractionLogRepo{}
	heroRepo := &mockHeroRepository{
		listFunc: func(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
			listCalled = true
			return nil, "", 0, nil
		},
	}

	uc := NewExtractionUseCase(provider, logRepo, heroRepo, limit, slog.Default())

	result, err := uc.ExtractFromText(context.Background(), "Текст о герое", "")
	if err != nil {
		t.Fatalf("ExtractFromText() error = %v", err)
	}
	if listCalled {
		t.Error("heroRepo.List should not be called when hero is nil")
	}
	if len(result.DuplicateHeroIDs) != 0 {
		t.Errorf("DuplicateHeroIDs should be empty, got %v", result.DuplicateHeroIDs)
	}
}
