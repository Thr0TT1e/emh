package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type heroSourceRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewHeroSourceRepository(pool *pgxpool.Pool) repository.HeroSourceRepository {
	return &heroSourceRepository{pool: pool, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *heroSourceRepository) Add(ctx context.Context, p domain.AddHeroSourceParams) (string, error) {
	sourceType := p.SourceType
	if sourceType == "" {
		sourceType = "website"
	}
	query := r.sb.Insert("hero_sources").
		Columns("hero_id", "url", "title", "source_type", "excerpt").
		Values(p.HeroID, p.URL, p.Title, sourceType, p.Excerpt).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build add hero source: %w", err)
	}
	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec add hero source: %w", err)
	}
	return id, nil
}

func (r *heroSourceRepository) Remove(ctx context.Context, heroID, sourceID string) error {
	query := r.sb.Delete("hero_sources").Where(squirrel.Eq{"id": sourceID, "hero_id": heroID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build remove hero source: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec remove hero source: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero source not found")
	}
	return nil
}

func (r *heroSourceRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroSource, error) {
	query := r.sb.Select("id, hero_id, url, title, source_type, excerpt, created_at").
		From("hero_sources").
		Where(squirrel.Eq{"hero_id": heroID}).
		OrderBy("created_at ASC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list hero sources: %w", err)
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list hero sources: %w", err)
	}
	defer rows.Close()

	var sources []*domain.HeroSource
	for rows.Next() {
		s := &domain.HeroSource{}
		if err := rows.Scan(&s.ID, &s.HeroID, &s.URL, &s.Title, &s.SourceType, &s.Excerpt, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan hero source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}
