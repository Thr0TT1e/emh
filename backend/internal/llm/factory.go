package llm

import (
	"fmt"
	"os"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// NewLLMProvider создаёт LLMProvider на основе конфигурации.
// Находит провайдер по DefaultProvider (или берёт первый из списка).
func NewLLMProvider(cfg config.LLMConfig) (domain.LLMProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("LLM is disabled")
	}

	// Ищем провайдер по имени DefaultProvider
	var providerCfg *config.LLMProviderConfig
	for i := range cfg.Providers {
		if cfg.Providers[i].Name == cfg.DefaultProvider {
			providerCfg = &cfg.Providers[i]
			break
		}
	}
	// Fallback: первый провайдер в списке
	if providerCfg == nil && len(cfg.Providers) > 0 {
		providerCfg = &cfg.Providers[0]
	}
	if providerCfg == nil {
		return nil, fmt.Errorf("no LLM providers configured")
	}

	return newProviderFromConfig(*providerCfg)
}

// newProviderFromConfig создаёт конкретный провайдер по типу.
func newProviderFromConfig(cfg config.LLMProviderConfig) (domain.LLMProvider, error) {
	// Читаем API key из env (если требуется)
	apiKey := ""
	if cfg.APIKeyEnv != "" {
		apiKey = os.Getenv(cfg.APIKeyEnv)
		if apiKey == "" {
			return nil, fmt.Errorf("API key env %q is not set for provider %q", cfg.APIKeyEnv, cfg.Name)
		}
	}

	switch cfg.Type {
	case "ollama":
		return NewOllamaProvider(
			cfg.Name,
			cfg.Endpoint,
			cfg.Model,
			cfg.Timeout.Duration,
		), nil

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
		return nil, fmt.Errorf("unknown LLM provider type: %q", cfg.Type)
	}
}
