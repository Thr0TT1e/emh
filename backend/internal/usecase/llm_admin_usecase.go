package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// LLMAdminUseCase бизнес-логика управления LLM-провайдерами.
type LLMAdminUseCase interface {
	ListProviders(ctx context.Context) ([]*domain.LLMProviderInfo, error)
	SetActiveProvider(ctx context.Context, id string) error
	TestProvider(ctx context.Context, id string, testPrompt string) (*domain.LLMTestResult, error)
	UpdateProvider(ctx context.Context, id string, priority int, notes string) error
}

type llmAdminUseCase struct {
	providerRepo    repository.LLMProviderRepository
	providerManager domain.LLMProviderManager
	logger          *slog.Logger
}

// NewLLMAdminUseCase создаёт usecase управления провайдерами.
func NewLLMAdminUseCase(
	providerRepo repository.LLMProviderRepository,
	providerManager domain.LLMProviderManager,
	logger *slog.Logger,
) LLMAdminUseCase {
	return &llmAdminUseCase{
		providerRepo:    providerRepo,
		providerManager: providerManager,
		logger:          logger,
	}
}

func (uc *llmAdminUseCase) ListProviders(ctx context.Context) ([]*domain.LLMProviderInfo, error) {
	records, err := uc.providerRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list providers: %w", err)
	}

	providers := make([]*domain.LLMProviderInfo, 0, len(records))
	for _, r := range records {
		// Получаем тип и модель из manager (если провайдер загружен)
		provider, _ := uc.providerManager.GetProviderByName(r.Name)

		info := &domain.LLMProviderInfo{
			ID:        r.ID,
			Name:      r.Name,
			IsActive:  r.IsActive,
			Priority:  r.Priority,
			Notes:     r.Notes,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}

		if provider != nil {
			info.Type = provider.Type()
			info.Model = provider.Model()
		}

		providers = append(providers, info)
	}

	return providers, nil
}

func (uc *llmAdminUseCase) SetActiveProvider(ctx context.Context, id string) error {
	record, err := uc.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}

	if err := uc.providerRepo.SetActive(ctx, record.ID); err != nil {
		return fmt.Errorf("set active provider: %w", err)
	}

	uc.logger.Info("Активный LLM-провайдер изменён",
		"provider_id", id,
		"provider_name", record.Name,
	)

	return nil
}

func (uc *llmAdminUseCase) TestProvider(ctx context.Context, id string, testPrompt string) (*domain.LLMTestResult, error) {
	record, err := uc.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	provider, err := uc.providerManager.GetProviderByName(record.Name)
	if err != nil {
		return &domain.LLMTestResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("provider not loaded: %v", err),
		}, nil
	}

	start := time.Now()
	rawResponse, err := provider.Generate(ctx, testPrompt)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return &domain.LLMTestResult{
			Success:          false,
			ProcessingTimeMs: elapsed,
			ErrorMessage:     err.Error(),
		}, nil
	}

	return &domain.LLMTestResult{
		Success:          true,
		RawResponse:      rawResponse,
		ProcessingTimeMs: elapsed,
	}, nil
}

func (uc *llmAdminUseCase) UpdateProvider(ctx context.Context, id string, priority int, notes string) error {
	record, err := uc.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}

	record.Priority = priority
	record.Notes = notes

	if err := uc.providerRepo.Update(ctx, record); err != nil {
		return fmt.Errorf("update provider: %w", err)
	}

	return nil
}
