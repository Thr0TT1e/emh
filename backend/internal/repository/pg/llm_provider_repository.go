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

type llmProviderRepository struct {
	pool *pgxpool.Pool
	ps   squirrel.StatementBuilderType
}

// NewLLMProviderRepository создаёт PostgreSQL-реализацию LLMProviderRepository.
func NewLLMProviderRepository(pool *pgxpool.Pool) repository.LLMProviderRepository {
	return &llmProviderRepository{
		pool: pool,
		ps:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *llmProviderRepository) Create(ctx context.Context, record *domain.LLMProviderRecord) (string, error) {
	query := `
		INSERT INTO llm_providers (name, is_active, priority, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		record.Name,
		record.IsActive,
		record.Priority,
		record.Notes,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)

	if err != nil {
		return "", fmt.Errorf("insert llm_provider: %w", err)
	}
	return record.ID, nil
}

func (r *llmProviderRepository) scanRecord(row pgx.Row) (*domain.LLMProviderRecord, error) {
	var rec domain.LLMProviderRecord
	err := row.Scan(
		&rec.ID,
		&rec.Name,
		&rec.IsActive,
		&rec.Priority,
		&rec.Notes,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *llmProviderRepository) GetByID(ctx context.Context, id string) (*domain.LLMProviderRecord, error) {
	query := `SELECT id, name, is_active, priority, notes, created_at, updated_at
	          FROM llm_providers WHERE id = $1`

	rec, err := r.scanRecord(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select llm_provider by id: %w", err)
	}
	return rec, nil
}

func (r *llmProviderRepository) GetByName(ctx context.Context, name string) (*domain.LLMProviderRecord, error) {
	query := `SELECT id, name, is_active, priority, notes, created_at, updated_at
	          FROM llm_providers WHERE name = $1`

	rec, err := r.scanRecord(r.pool.QueryRow(ctx, query, name))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select llm_provider by name: %w", err)
	}
	return rec, nil
}

func (r *llmProviderRepository) GetActive(ctx context.Context) (*domain.LLMProviderRecord, error) {
	query := `SELECT id, name, is_active, priority, notes, created_at, updated_at
	          FROM llm_providers WHERE is_active = true LIMIT 1`

	rec, err := r.scanRecord(r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Нет активного провайдера, это нормальная ситуация (fallback на default)
		}
		return nil, fmt.Errorf("select active llm_provider: %w", err)
	}
	return rec, nil
}

func (r *llmProviderRepository) List(ctx context.Context) ([]*domain.LLMProviderRecord, error) {
	query := `SELECT id, name, is_active, priority, notes, created_at, updated_at
	          FROM llm_providers ORDER BY priority DESC, name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select llm_providers: %w", err)
	}
	defer rows.Close()

	var records []*domain.LLMProviderRecord
	for rows.Next() {
		rec, err := r.scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan llm_provider: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *llmProviderRepository) Update(ctx context.Context, record *domain.LLMProviderRecord) error {
	qb := r.ps.Update("llm_providers").
		Where(squirrel.Eq{"id": record.ID})

	if record.Priority != 0 {
		qb = qb.Set("priority", record.Priority)
	}
	if record.Notes != "" {
		qb = qb.Set("notes", record.Notes)
	}
	qb = qb.Set("updated_at", time.Now())

	sql, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("build update llm_provider query: %w", err)
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update llm_provider: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *llmProviderRepository) SetActive(ctx context.Context, id string) error {
	// Триггер trg_single_active_provider в БД сам сбросит is_active у остальных записей.
	query := `UPDATE llm_providers SET is_active = true, updated_at = NOW() WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("set active llm_provider: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
