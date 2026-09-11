package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// contactMessageRepository PostgreSQL-реализация репозитория сообщений обратной связи.
type contactMessageRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

// NewContactMessageRepository создаёт PostgreSQL-реализацию ContactRepository.
func NewContactMessageRepository(pool *pgxpool.Pool) repository.ContactRepository {
	return &contactMessageRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create сохраняет новое сообщение обратной связи и возвращает его UUID.
// email_status по умолчанию 'pending' (устанавливается БД).
func (r *contactMessageRepository) Create(ctx context.Context, p domain.CreateContactMessageParams) (string, error) {
	query := r.sb.Insert("contact_messages").
		Columns(
			"name",
			"email",
			"subject",
			"message",
			"message_hash",
			"page_url",
			"ip_hash",
			"user_agent",
			"consent",
			"is_honeypot",
		).
		Values(
			p.Name,
			p.Email,
			p.Subject,
			p.Message,
			p.MessageHash,
			p.PageURL,
			nullableString(p.IPHash),
			nullableString(p.UserAgent),
			p.Consent,
			p.IsHoneypot,
		).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create contact message query: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create contact message: %w", err)
	}

	return id, nil
}

// UpdateEmailStatus обновляет статус отправки email-уведомления.
// При status = EmailStatusSent устанавливает email_sent_at = now().
// При status = EmailStatusFailed записывает emailErr в email_error.
func (r *contactMessageRepository) UpdateEmailStatus(
	ctx context.Context,
	id string,
	status domain.EmailStatus,
	emailErr string,
) error {
	update := r.sb.Update("contact_messages").
		Set("email_status", string(status)).
		Where(squirrel.Eq{"id": id})

	if status == domain.EmailStatusSent {
		update = update.Set("email_sent_at", time.Now())
	}
	if emailErr != "" {
		update = update.Set("email_error", emailErr)
		// При skipped email_error остаётся NULL.
	}

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build update email status: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec update email status: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("contact message %w", domain.ErrNotFound)
	}

	return nil
}

// FindRecentDuplicate ищет сообщение с тем же email и message_hash,
// созданное не ранее since. Возвращает ID найденного сообщения
// или пустую строку, если дубликат не найден.
func (r *contactMessageRepository) FindRecentDuplicate(
	ctx context.Context,
	email, messageHash string,
	since time.Time,
) (string, error) {
	query := r.sb.Select("id").
		From("contact_messages").
		Where(squirrel.Eq{"email": email}).
		Where(squirrel.Eq{"message_hash": messageHash}).
		Where(squirrel.GtOrEq{"created_at": since}).
		OrderBy("created_at DESC").
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build find duplicate query: %w", err)
	}

	var id string
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Дубликат не найден — не ошибка.
		}
		return "", fmt.Errorf("exec find duplicate: %w", err)
	}

	return id, nil
}
