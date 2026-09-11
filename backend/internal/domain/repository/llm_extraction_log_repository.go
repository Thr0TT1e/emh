package repository

import (
	"context"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// LLMExtractionLogRepository интерфейс для работы с логами LLM-экстракций.
type LLMExtractionLogRepository interface {
	Create(ctx context.Context, log *domain.LLMExtractionLog) (string, error)
	GetByID(ctx context.Context, id string) (*domain.LLMExtractionLog, error)
	List(ctx context.Context, limit, offset int) ([]*domain.LLMExtractionLog, error)
	// DeleteOlderThan удаляет записи старше указанного времени и возвращает количество удаленных строк.
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error)
}
