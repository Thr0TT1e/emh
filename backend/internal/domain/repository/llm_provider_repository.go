package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// LLMProviderRepository интерфейс для управления провайдерами в БД.
type LLMProviderRepository interface {
	Create(ctx context.Context, record *domain.LLMProviderRecord) (string, error)
	GetByID(ctx context.Context, id string) (*domain.LLMProviderRecord, error)
	GetByName(ctx context.Context, name string) (*domain.LLMProviderRecord, error)
	GetActive(ctx context.Context) (*domain.LLMProviderRecord, error)
	List(ctx context.Context) ([]*domain.LLMProviderRecord, error)
	Update(ctx context.Context, record *domain.LLMProviderRecord) error
	SetActive(ctx context.Context, id string) error
}
