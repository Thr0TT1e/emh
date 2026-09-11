package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type heroConflictRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewHeroConflictRepository(pool *pgxpool.Pool) repository.HeroConflictRepository {
	return &heroConflictRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *heroConflictRepository) Add(ctx context.Context, heroID, conflictID, specificLocation, rankAtConflict string) error {
	query := r.sb.Insert("hero_conflicts").
		Columns("hero_id", "conflict_id", "specific_location", "rank_at_conflict").
		Values(heroID, conflictID, specificLocation, rankAtConflict).
		Suffix("ON CONFLICT (hero_id, conflict_id) DO NOTHING")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build add hero conflict: %w", err)
	}
	if _, err := r.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec add hero conflict: %w", err)
	}
	return nil
}

func (r *heroConflictRepository) Remove(ctx context.Context, heroID, conflictID string) error {
	query := r.sb.Delete("hero_conflicts").
		Where(squirrel.Eq{"hero_id": heroID, "conflict_id": conflictID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build remove hero conflict: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec remove hero conflict: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero conflict not found")
	}
	return nil
}

func (r *heroConflictRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroConflict, error) {
	query := r.sb.Select(
		"hc.hero_id", "hc.conflict_id", "c.name", "hc.specific_location", "hc.rank_at_conflict",
	).
		From("hero_conflicts hc").
		Join("conflicts c ON hc.conflict_id = c.id").
		Where(squirrel.Eq{"hc.hero_id": heroID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list hero conflicts: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list hero conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []*domain.HeroConflict
	for rows.Next() {
		hc := &domain.HeroConflict{}
		if err := rows.Scan(&hc.HeroID, &hc.ConflictID, &hc.ConflictName, &hc.SpecificLocation, &hc.RankAtConflict); err != nil {
			return nil, fmt.Errorf("scan hero conflict: %w", err)
		}
		conflicts = append(conflicts, hc)
	}
	return conflicts, rows.Err()
}
