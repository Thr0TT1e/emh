package domain

import "context"

// LLMProvider абстракция над любым LLM-провайдером.
type LLMProvider interface {
	// Name возвращает уникальное имя провайдера (из конфига).
	Name() string
	// Type возвращает тип провайдера (ollama, openai_compatible, anthropic).
	Type() string
	// Model возвращает название используемой модели.
	Model() string
	// Generate отправляет промпт и возвращает сырой текст (ожидается JSON).
	Generate(ctx context.Context, prompt string) (string, error)
	// HealthCheck проверяет доступность провайдера (для Admin API).
	HealthCheck(ctx context.Context) error
}

// LLMProviderManager управляет провайдерами: загрузка из конфига, получение активного из БД.
type LLMProviderManager interface {
	// GetActiveProvider возвращает текущий активный провайдер (из БД или default из конфига).
	GetActiveProvider(ctx context.Context) (LLMProvider, error)
	// GetProviderByName возвращает провайдер по имени (для админки).
	GetProviderByName(name string) (LLMProvider, error)
	// ListProviders возвращает список всех настроенных провайдеров.
	ListProviders() []LLMProvider
}
