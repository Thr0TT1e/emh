package repository

import (
	"context"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ContactRepository интерфейс доступа к сообщениям обратной связи.
type ContactRepository interface {
	// Create сохраняет новое сообщение и возвращает его UUID.
	Create(ctx context.Context, p domain.CreateContactMessageParams) (string, error)

	// UpdateEmailStatus обновляет статус отправки email-уведомления.
	// При status = EmailStatusSent устанавливает email_sent_at = now().
	// При status = EmailStatusFailed записывает emailErr в email_error.
	// Возвращает domain.ErrNotFound если запись не существует.
	UpdateEmailStatus(ctx context.Context, id string, status domain.EmailStatus, emailErr string) error

	// FindRecentDuplicate ищет сообщение с тем же email и message_hash,
	// созданное не ранее since. Возвращает ID найденного сообщения
	// или пустую строку, если дубликат не найден.
	FindRecentDuplicate(ctx context.Context, email, messageHash string, since time.Time) (string, error)
}
