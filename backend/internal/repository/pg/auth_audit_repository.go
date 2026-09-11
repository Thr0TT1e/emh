package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type authAuditRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

// NewAuthAuditRepository создаёт PostgreSQL-реализацию репозитория аудита.
func NewAuthAuditRepository(pool *pgxpool.Pool) repository.AuthAuditRepository {
	return &authAuditRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *authAuditRepository) Add(ctx context.Context, entry domain.AuthAuditEntry) error {
	query := r.sb.Insert("auth_audit_log").
		Columns(
			"ip_address",
			"mechanism",
			"result",
			"identity",
			"procedure",
			"failure_reason",
			"user_agent",
		).
		Values(
			entry.IPAddress,
			string(entry.Mechanism),
			string(entry.Result),
			entry.Identity,
			entry.Procedure,
			entry.FailureReason,
			entry.UserAgent,
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build auth audit insert: %w", err)
	}

	if _, err := r.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec auth audit insert: %w", err)
	}

	return nil
}

// DeleteOlderThan удаляет записи аудита старше указанной даты.
// Использует индекс idx_auth_audit_log_created_at для эффективного DELETE.
func (r *authAuditRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := r.sb.Delete("auth_audit_log").
		Where(squirrel.Lt{"created_at": olderThan})

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build auth audit delete: %w", err)
	}

	result, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("exec auth audit delete: %w", err)
	}

	return result.RowsAffected(), nil
}
