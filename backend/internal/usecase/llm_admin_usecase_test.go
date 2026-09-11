package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ─── Хелперы ────────────────────────────────────────────────────────────────

func newTestLLMAdminUC(
	repo *mockLLMProviderRepo,
	manager *mockLLMProviderManager,
) *llmAdminUseCase {
	return &llmAdminUseCase{
		providerRepo:    repo,
		providerManager: manager,
		logger:          slog.Default(),
	}
}

func testProviderRecord(id, name string, isActive bool, priority int) *domain.LLMProviderRecord {
	return &domain.LLMProviderRecord{
		ID:        id,
		Name:      name,
		IsActive:  isActive,
		Priority:  priority,
		Notes:     "test notes",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
}

// ─── Тесты ListProviders ────────────────────────────────────────────────────

// TestListProviders_Success проверяет успешное объединение БД и in-memory данных.
func TestListProviders_Success(t *testing.T) {
	repo := &mockLLMProviderRepo{
		listFn: func(ctx context.Context) ([]*domain.LLMProviderRecord, error) {
			return []*domain.LLMProviderRecord{
				testProviderRecord("id-1", "ollama_local", true, 1),
				testProviderRecord("id-2", "openrouter", false, 2),
			}, nil
		},
	}
	manager := &mockLLMProviderManager{
		getProviderByNameFn: func(name string) (domain.LLMProvider, error) {
			return &mockLLMProvider{
				nameFn:  func() string { return name },
				typeFn:  func() string { return "ollama" },
				modelFn: func() string { return "llama3.1:8b" },
			}, nil
		},
	}

	uc := newTestLLMAdminUC(repo, manager)
	providers, err := uc.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}
	if len(providers) != 2 {
		t.Errorf("ListProviders() returned %d providers, want 2", len(providers))
	}

	// Проверяем, что данные из БД и менеджера объединены.
	for _, p := range providers {
		if p.Type == "" {
			t.Errorf("provider %s has empty Type, want from manager", p.Name)
		}
		if p.Model == "" {
			t.Errorf("provider %s has empty Model, want from manager", p.Name)
		}
	}
}

// TestListProviders_RepoError проверяет обработку ошибки репозитория.
func TestListProviders_RepoError(t *testing.T) {
	repo := &mockLLMProviderRepo{
		listFn: func(ctx context.Context) ([]*domain.LLMProviderRecord, error) {
			return nil, errors.New("db connection failed")
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	_, err := uc.ListProviders(context.Background())
	if err == nil {
		t.Error("ListProviders() should return error on repo failure")
	}
}

// TestListProviders_ProviderNotLoaded проверяет, что провайдер без in-memory данных
// всё равно возвращается (с пустыми Type/Model).
func TestListProviders_ProviderNotLoaded(t *testing.T) {
	repo := &mockLLMProviderRepo{
		listFn: func(ctx context.Context) ([]*domain.LLMProviderRecord, error) {
			return []*domain.LLMProviderRecord{
				testProviderRecord("id-1", "unknown_provider", false, 1),
			}, nil
		},
	}
	manager := &mockLLMProviderManager{
		getProviderByNameFn: func(name string) (domain.LLMProvider, error) {
			return nil, domain.ErrNotFound // Провайдер не загружен в память.
		},
	}

	uc := newTestLLMAdminUC(repo, manager)
	providers, err := uc.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("ListProviders() returned %d providers, want 1", len(providers))
	}
	// Провайдер должен быть в списке, но без Type/Model.
	if providers[0].Type != "" {
		t.Errorf("provider Type = %q, want empty for unloaded provider", providers[0].Type)
	}
}

// ─── Тесты SetActiveProvider ────────────────────────────────────────────────

// TestSetActiveProvider_Success проверяет успешную установку активного провайдера.
func TestSetActiveProvider_Success(t *testing.T) {
	var setActiveCalled bool
	var setActiveID string
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "ollama_local", false, 1), nil
		},
		setActiveFn: func(ctx context.Context, id string) error {
			setActiveCalled = true
			setActiveID = id
			return nil
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	err := uc.SetActiveProvider(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("SetActiveProvider() error = %v", err)
	}
	if !setActiveCalled {
		t.Error("SetActiveProvider() should call repo.SetActive")
	}
	if setActiveID != "id-1" {
		t.Errorf("SetActiveProvider() called with id = %q, want %q", setActiveID, "id-1")
	}
}

// TestSetActiveProvider_NotFound проверяет обработку несуществующего провайдера.
func TestSetActiveProvider_NotFound(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return nil, domain.ErrNotFound
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	err := uc.SetActiveProvider(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("SetActiveProvider() should return error for nonexistent provider")
	}
}

// ─── Тесты UpdateProvider ───────────────────────────────────────────────────

// TestUpdateProvider_Success проверяет успешное обновление приоритета и заметок.
func TestUpdateProvider_Success(t *testing.T) {
	var updatedRecord *domain.LLMProviderRecord
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "ollama_local", true, 1), nil
		},
		updateFn: func(ctx context.Context, record *domain.LLMProviderRecord) error {
			updatedRecord = record
			return nil
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	err := uc.UpdateProvider(context.Background(), "id-1", 5, "updated notes")
	if err != nil {
		t.Fatalf("UpdateProvider() error = %v", err)
	}
	if updatedRecord == nil {
		t.Fatal("UpdateProvider() should call repo.Update")
	}
	if updatedRecord.Priority != 5 {
		t.Errorf("UpdateProvider() priority = %d, want 5", updatedRecord.Priority)
	}
	if updatedRecord.Notes != "updated notes" {
		t.Errorf("UpdateProvider() notes = %q, want %q", updatedRecord.Notes, "updated notes")
	}
}

// TestUpdateProvider_NotFound проверяет обработку несуществующего провайдера.
func TestUpdateProvider_NotFound(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return nil, domain.ErrNotFound
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	err := uc.UpdateProvider(context.Background(), "nonexistent-id", 5, "notes")
	if err == nil {
		t.Error("UpdateProvider() should return error for nonexistent provider")
	}
}

// TestUpdateProvider_RepoUpdateError проверяет обработку ошибки обновления в БД.
func TestUpdateProvider_RepoUpdateError(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "ollama_local", true, 1), nil
		},
		updateFn: func(ctx context.Context, record *domain.LLMProviderRecord) error {
			return errors.New("db update failed")
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	err := uc.UpdateProvider(context.Background(), "id-1", 5, "notes")
	if err == nil {
		t.Error("UpdateProvider() should return error on repo.Update failure")
	}
}

// ─── Тесты TestProvider ─────────────────────────────────────────────────────

// TestTestProvider_Success проверяет успешное тестирование провайдера.
func TestTestProvider_Success(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "ollama_local", true, 1), nil
		},
	}
	manager := &mockLLMProviderManager{
		getProviderByNameFn: func(name string) (domain.LLMProvider, error) {
			return &mockLLMProvider{
				generateFn: func(ctx context.Context, prompt string) (string, error) {
					return `{"hero": {"name": "Test Hero"}}`, nil
				},
			}, nil
		},
	}

	uc := newTestLLMAdminUC(repo, manager)
	result, err := uc.TestProvider(context.Background(), "id-1", "test prompt")
	if err != nil {
		t.Fatalf("TestProvider() error = %v", err)
	}
	if result == nil {
		t.Fatal("TestProvider() should return result")
	}
	if !result.Success {
		t.Errorf("TestProvider() Success = %v, want true (ErrorMessage: %s)", result.Success, result.ErrorMessage)
	}
}

// TestTestProvider_ProviderNotLoaded проверяет, что незагруженный провайдер
// возвращает Success=false без ошибки.
func TestTestProvider_ProviderNotLoaded(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "unknown_provider", false, 1), nil
		},
	}
	manager := &mockLLMProviderManager{
		getProviderByNameFn: func(name string) (domain.LLMProvider, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := newTestLLMAdminUC(repo, manager)
	result, err := uc.TestProvider(context.Background(), "id-1", "test prompt")
	if err != nil {
		t.Fatalf("TestProvider() should not return error for unloaded provider, got %v", err)
	}
	if result.Success {
		t.Error("TestProvider() Success = true, want false for unloaded provider")
	}
	if result.ErrorMessage == "" {
		t.Error("TestProvider() ErrorMessage should not be empty for unloaded provider")
	}
}

// TestTestProvider_GenerateError проверяет обработку ошибки генерации.
func TestTestProvider_GenerateError(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return testProviderRecord(id, "ollama_local", true, 1), nil
		},
	}
	manager := &mockLLMProviderManager{
		getProviderByNameFn: func(name string) (domain.LLMProvider, error) {
			return &mockLLMProvider{
				generateFn: func(ctx context.Context, prompt string) (string, error) {
					return "", errors.New("connection timeout")
				},
			}, nil
		},
	}

	uc := newTestLLMAdminUC(repo, manager)
	result, err := uc.TestProvider(context.Background(), "id-1", "test prompt")
	if err != nil {
		t.Fatalf("TestProvider() should not return error for generate failure, got %v", err)
	}
	if result.Success {
		t.Error("TestProvider() Success = true, want false for generate error")
	}
}

// TestTestProvider_RecordNotFound проверяет обработку несуществующего провайдера.
func TestTestProvider_RecordNotFound(t *testing.T) {
	repo := &mockLLMProviderRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
			return nil, domain.ErrNotFound
		},
	}
	manager := &mockLLMProviderManager{}

	uc := newTestLLMAdminUC(repo, manager)
	_, err := uc.TestProvider(context.Background(), "nonexistent-id", "test prompt")
	if err == nil {
		t.Error("TestProvider() should return error for nonexistent provider")
	}
}
