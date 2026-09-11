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

type awardRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewAwardRepository(pool *pgxpool.Pool) repository.AwardRepository {
	return &awardRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *awardRepository) Create(ctx context.Context, p domain.CreateAwardParams) (string, error) {
	query := r.sb.Insert("awards").
		Columns("name", "description", "image_url", "sort_order").
		Values(p.Name, p.Description, p.ImageURL, p.SortOrder).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create award: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create award: %w", err)
	}
	return id, nil
}

func (r *awardRepository) Update(ctx context.Context, p domain.UpdateAwardParams) (*domain.Award, error) {
	update := r.sb.Update("awards").Where(squirrel.Eq{"id": p.ID})

	mask := make(map[string]bool, len(p.FieldMask))
	for _, f := range p.FieldMask {
		mask[f] = true
	}

	if mask["name"] {
		update = update.Set("name", p.Name)
	}
	if mask["description"] {
		update = update.Set("description", p.Description)
	}
	if mask["image_url"] {
		update = update.Set("image_url", p.ImageURL)
	}
	if mask["sort_order"] {
		update = update.Set("sort_order", p.SortOrder)
	}

	update = update.Set("updated_at", time.Now())
	update = update.Suffix("RETURNING id, name, description, image_url, sort_order, created_at, updated_at")

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update award: %w", err)
	}

	a := &domain.Award{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&a.ID, &a.Name, &a.Description, &a.ImageURL, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("award not found")
		}
		return nil, fmt.Errorf("exec update award: %w", err)
	}
	return a, nil
}

func (r *awardRepository) Delete(ctx context.Context, id string) error {
	query := r.sb.Delete("awards").Where(squirrel.Eq{"id": id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete award: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec delete award: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("award not found")
	}
	return nil
}

func (r *awardRepository) GetByID(ctx context.Context, id string) (*domain.Award, error) {
	query := r.sb.Select("id, name, description, image_url, sort_order, created_at, updated_at").
		From("awards").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get award: %w", err)
	}

	a := &domain.Award{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&a.ID, &a.Name, &a.Description, &a.ImageURL, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("award not found")
		}
		return nil, fmt.Errorf("exec get award: %w", err)
	}
	return a, nil
}

func (r *awardRepository) List(ctx context.Context, f domain.AwardFilter) ([]*domain.Award, string, int64, error) {
	// 1. COUNT(*)
	countQuery := r.sb.Select("COUNT(*)").From("awards")
	if f.SearchQuery != "" {
		countQuery = countQuery.Where(squirrel.ILike{"name": "%" + f.SearchQuery + "%"})
	}
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build count awards: %w", err)
	}
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("exec count awards: %w", err)
	}

	// 2. SELECT с пагинацией
	sel := r.sb.Select("id, name, description, image_url, sort_order, created_at, updated_at").
		From("awards").
		OrderBy("sort_order ASC, name ASC, id ASC")

	if f.SearchQuery != "" {
		sel = sel.Where(squirrel.ILike{"name": "%" + f.SearchQuery + "%"})
	}
	if f.Cursor != "" {
		sel = sel.Where(squirrel.Gt{"id": f.Cursor})
	}

	limit := f.Limit + 1
	sel = sel.Limit(uint64(limit))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list awards: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("exec list awards: %w", err)
	}
	defer rows.Close()

	var awards []*domain.Award
	for rows.Next() {
		a := &domain.Award{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.ImageURL, &a.SortOrder, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, "", 0, fmt.Errorf("scan award: %w", err)
		}
		awards = append(awards, a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration error: %w", err)
	}

	var nextCursor string
	if len(awards) > f.Limit {
		awards = awards[:f.Limit]
		nextCursor = awards[len(awards)-1].ID
	}

	return awards, nextCursor, total, nil
}

func (r *awardRepository) HasHeroes(ctx context.Context, awardID string) (bool, error) {
	query := r.sb.Select("COUNT(*)").
		From("hero_awards").
		Where(squirrel.Eq{"award_id": awardID})

	sql, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("build has heroes: %w", err)
	}

	var count int
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("exec has heroes: %w", err)
	}
	return count > 0, nil
}
