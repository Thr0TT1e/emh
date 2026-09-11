package llm

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// providerManager реализует domain.LLMProviderManager.
type providerManager struct {
	providers   map[string]domain.LLMProvider // Кеш провайдеров по имени
	defaultName string                        // Имя провайдера по умолчанию из конфига
	repo        repository.LLMProviderRepository
	logger      *slog.Logger
	mu          sync.RWMutex
}

func NewProviderManager(cfg config.LLMConfig, repo repository.LLMProviderRepository, logger *slog.Logger) (domain.LLMProviderManager, error) {
	pm := &providerManager{
		providers:   make(map[string]domain.LLMProvider),
		defaultName: cfg.DefaultProvider,
		repo:        repo,
		logger:      logger,
	}

	// Инициализируем провайдеры из конфига
	for _, providerCfg := range cfg.Providers {
		provider, err := pm.createProvider(providerCfg)
		if err != nil {
			logger.Warn("failed to create provider", "name", providerCfg.Name, "error", err)
			continue
		}
		pm.providers[providerCfg.Name] = provider
	}

	// Синхронизируем с БД (создаём записи для новых провайдеров)
	if err := pm.syncWithDatabase(context.Background()); err != nil {
		logger.Error("failed to sync providers with database", "error", err)
	}

	return pm, nil
}

func (pm *providerManager) createProvider(cfg config.LLMProviderConfig) (domain.LLMProvider, error) {
	// Получаем API key из env (если требуется)
	apiKey := ""
	if cfg.APIKeyEnv != "" {
		apiKey = os.Getenv(cfg.APIKeyEnv)
		if apiKey == "" {
			return nil, fmt.Errorf("API key env %s is not set", cfg.APIKeyEnv)
		}
	}

	switch cfg.Type {
	case "ollama":
		return NewOllamaProvider(cfg.Name, cfg.Endpoint, cfg.Model, cfg.Timeout.Duration), nil
	case "openai_compatible":
		return NewOpenAICompatibleProvider(
			cfg.Name,
			cfg.Endpoint,
			cfg.Model,
			apiKey,
			cfg.Timeout.Duration,
			cfg.MaxTokens,
			cfg.Temperature,
		), nil
	case "anthropic":
		return NewAnthropicProvider(
			cfg.Name,
			cfg.Endpoint,
			cfg.Model,
			apiKey,
			cfg.Timeout.Duration,
			cfg.MaxTokens,
			cfg.Temperature,
		), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}

func (pm *providerManager) syncWithDatabase(ctx context.Context) error {
	// Получаем все провайдеры из БД
	dbProviders, err := pm.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list providers from database: %w", err)
	}

	dbProviderMap := make(map[string]*domain.LLMProviderRecord)
	for _, p := range dbProviders {
		dbProviderMap[p.Name] = p
	}

	// Создаём записи в БД для новых провайдеров из конфига
	for name := range pm.providers {
		if _, exists := dbProviderMap[name]; !exists {
			record := &domain.LLMProviderRecord{
				Name:     name,
				IsActive: name == pm.defaultName,
				Priority: 0,
			}
			if _, err := pm.repo.Create(ctx, record); err != nil {
				pm.logger.Error("failed to create provider record", "name", name, "error", err)
			}
		}
	}

	return nil
}

func (pm *providerManager) GetActiveProvider(ctx context.Context) (domain.LLMProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Ищем активный провайдер в БД
	activeRecord, err := pm.repo.GetActive(ctx)
	if err != nil {
		pm.logger.Warn("failed to get active provider from database, using default", "error", err)
	} else if activeRecord != nil {
		if provider, exists := pm.providers[activeRecord.Name]; exists {
			return provider, nil
		}
		pm.logger.Warn("active provider from database not found in config", "name", activeRecord.Name)
	}

	// Fallback на default из конфига
	if provider, exists := pm.providers[pm.defaultName]; exists {
		return provider, nil
	}

	return nil, domain.ErrLLMUnavailable
}

func (pm *providerManager) GetProviderByName(name string) (domain.LLMProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

func (pm *providerManager) ListProviders() []domain.LLMProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]domain.LLMProvider, 0, len(pm.providers))
	for _, p := range pm.providers {
		result = append(result, p)
	}
	return result
}
