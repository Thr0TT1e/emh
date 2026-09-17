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

// awardColumns список колонок для чтения награды.
// COALESCE для ribbon_image_url: у строк, созданных до миграции 00027, значение NULL.
const awardColumns = `id, name, description, image_url, sort_order,
	COALESCE(ribbon_image_url, '') AS ribbon_image_url, type, jurisdiction, worn_without_bar, is_jubilee,
	created_at, updated_at`

// scanAward маппит строку БД в доменную модель (порядок колонок — awardColumns).
func scanAward(s scannable) (*domain.Award, error) {
	a := &domain.Award{}
	var awardType, jurisdiction int

	if err := s.Scan(
		&a.ID, &a.Name, &a.Description, &a.ImageURL, &a.SortOrder,
		&a.RibbonImageURL, &awardType, &jurisdiction, &a.WornWithoutBar, &a.IsJubilee,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, err
	}

	a.Type = domain.AwardType(awardType)
	a.Jurisdiction = domain.AwardJurisdiction(jurisdiction)
	return a, nil
}

func NewAwardRepository(pool *pgxpool.Pool) repository.AwardRepository {
	return &awardRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *awardRepository) Create(ctx context.Context, p domain.CreateAwardParams) (string, error) {
	query := r.sb.Insert("awards").
		Columns("name", "description", "image_url", "sort_order",
			"ribbon_image_url", "type", "jurisdiction", "worn_without_bar", "is_jubilee").
		Values(p.Name, p.Description, p.ImageURL, p.SortOrder,
			p.RibbonImageURL, int(p.Type), int(p.Jurisdiction), p.WornWithoutBar, p.IsJubilee).
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
	if mask["ribbon_image_url"] {
		update = update.Set("ribbon_image_url", p.RibbonImageURL)
	}
	if mask["type"] {
		update = update.Set("type", int(p.Type))
	}
	if mask["jurisdiction"] {
		update = update.Set("jurisdiction", int(p.Jurisdiction))
	}
	if mask["worn_without_bar"] {
		update = update.Set("worn_without_bar", p.WornWithoutBar)
	}
	if mask["is_jubilee"] {
		update = update.Set("is_jubilee", p.IsJubilee)
	}

	update = update.Set("updated_at", time.Now())
	update = update.Suffix("RETURNING " + awardColumns)

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update award: %w", err)
	}

	a, err := scanAward(r.pool.QueryRow(ctx, sql, args...))
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
	query := r.sb.Select(awardColumns).
		From("awards").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get award: %w", err)
	}

	a, err := scanAward(r.pool.QueryRow(ctx, sql, args...))
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
	sel := r.sb.Select(awardColumns).
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
		a, err := scanAward(rows)
		if err != nil {
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
