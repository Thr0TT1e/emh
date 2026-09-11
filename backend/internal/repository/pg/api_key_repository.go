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

const apiKeyColumns = "id, key_id, secret_hash, name, description, role, created_by, revoked_at, expires_at, last_used_at, last_used_ip, created_at, updated_at"

type apiKeyRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewAPIKeyRepository(pool *pgxpool.Pool) repository.APIKeyRepository {
	return &apiKeyRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// scanAPIKey маппит строку БД в доменную модель. scannable переиспользуется из пакета pg.
func scanAPIKey(s scannable) (*domain.APIKey, error) {
	k := &domain.APIKey{}
	if err := s.Scan(
		&k.ID, &k.KeyID, &k.SecretHash, &k.Name, &k.Description, &k.Role,
		&k.CreatedBy, &k.RevokedAt, &k.ExpiresAt, &k.LastUsedAt, &k.LastUsedIP,
		&k.CreatedAt, &k.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return k, nil
}

func (r *apiKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := r.sb.Insert("api_keys").
		Columns("key_id", "secret_hash", "name", "description", "role", "created_by", "expires_at").
		Values(key.KeyID, key.SecretHash, key.Name, key.Description, key.Role, key.CreatedBy, key.ExpiresAt).
		Suffix("RETURNING id, created_at, updated_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build create api key: %w", err)
	}

	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&key.ID, &key.CreatedAt, &key.UpdatedAt); err != nil {
		return fmt.Errorf("exec create api key: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error) {
	query := r.sb.Select(apiKeyColumns).
		From("api_keys").
		Where(squirrel.Eq{"key_id": keyID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get api key: %w", err)
	}

	k, err := scanAPIKey(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("api key not found")
		}
		return nil, fmt.Errorf("exec get api key: %w", err)
	}
	return k, nil
}

func (r *apiKeyRepository) List(ctx context.Context, includeRevoked bool) ([]*domain.APIKey, error) {
	sel := r.sb.Select(apiKeyColumns).
		From("api_keys").
		OrderBy("created_at DESC")

	if !includeRevoked {
		sel = sel.Where(squirrel.Eq{"revoked_at": nil})
	}

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list api keys: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id string) error {
	update := r.sb.Update("api_keys").
		Set("revoked_at", time.Now()).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id, "revoked_at": nil})

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build revoke api key: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec revoke api key: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("api key not found or already revoked")
	}
	return nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id string, ip string, at time.Time) error {
	update := r.sb.Update("api_keys").
		Set("last_used_at", at).
		Set("last_used_ip", ip).
		Where(squirrel.Eq{"id": id})

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build update last used: %w", err)
	}

	if _, err := r.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec update last used: %w", err)
	}
	return nil
}
