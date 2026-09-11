package repository

import (
	"context"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// AuthAuditRepository интерфейс для записи аудита аутентификации.
type AuthAuditRepository interface {
	// Add записывает одну запись аудита.
	Add(ctx context.Context, entry domain.AuthAuditEntry) error

	// DeleteOlderThan удаляет записи аудита старше указанной даты.
	// Возвращает количество удалённых записей.
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error)
}
