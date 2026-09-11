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

const refreshTokenColumns = "id, username, token_hash, expires_at, created_at, revoked_at"

type refreshTokenRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) repository.RefreshTokenRepository {
	return &refreshTokenRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, p domain.CreateRefreshTokenParams) (string, error) {
	query := r.sb.Insert("refresh_tokens").
		Columns("username", "token_hash", "expires_at").
		Values(p.Username, p.TokenHash, p.ExpiresAt).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create refresh token: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create refresh token: %w", err)
	}
	return id, nil
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := r.sb.Select(refreshTokenColumns).
		From("refresh_tokens").
		Where(squirrel.Eq{
			"token_hash": hash,
			"revoked_at": nil,
		}).
		Where(squirrel.Gt{"expires_at": time.Now()})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get refresh token: %w", err)
	}

	token := &domain.RefreshToken{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&token.ID, &token.Username, &token.TokenHash,
		&token.ExpiresAt, &token.CreatedAt, &token.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("exec get refresh token: %w", err)
	}
	return token, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id string) error {
	update := r.sb.Update("refresh_tokens").
		Set("revoked_at", time.Now()).
		Where(squirrel.Eq{"id": id, "revoked_at": nil})

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build revoke refresh token: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec revoke refresh token: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found")
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := r.sb.Delete("refresh_tokens").
		Where(squirrel.Lt{"expires_at": time.Now()})

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build delete expired: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("exec delete expired: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (r *refreshTokenRepository) GetByHashIncludingRevoked(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := r.sb.Select(refreshTokenColumns).
		From("refresh_tokens").
		Where(squirrel.Eq{"token_hash": hash}) // БЕЗ фильтра revoked_at
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	token := &domain.RefreshToken{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&token.ID, &token.Username, &token.TokenHash,
		&token.ExpiresAt, &token.CreatedAt, &token.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("exec query: %w", err)
	}
	return token, nil
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, username string) (int64, error) {
	update := r.sb.Update("refresh_tokens").
		Set("revoked_at", time.Now()).
		Where(squirrel.Eq{"username": username, "revoked_at": nil})
	sql, args, err := update.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("exec revoke all: %w", err)
	}
	return ct.RowsAffected(), nil
}
