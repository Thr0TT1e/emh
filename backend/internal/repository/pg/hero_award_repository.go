package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type heroAwardRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewHeroAwardRepository(pool *pgxpool.Pool) repository.HeroAwardRepository {
	return &heroAwardRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *heroAwardRepository) Add(ctx context.Context, heroID, awardID string, awardDate *string, decreeNumber string) error {
	query := r.sb.Insert("hero_awards").
		Columns("hero_id", "award_id", "award_date", "decree_number").
		Values(heroID, awardID, awardDate, decreeNumber).
		Suffix("ON CONFLICT (hero_id, award_id) DO NOTHING")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build add hero award: %w", err)
	}

	if _, err := r.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec add hero award: %w", err)
	}
	return nil
}

func (r *heroAwardRepository) Remove(ctx context.Context, heroID, awardID string) error {
	query := r.sb.Delete("hero_awards").
		Where(squirrel.Eq{"hero_id": heroID, "award_id": awardID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build remove hero award: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec remove hero award: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero award not found")
	}
	return nil
}

func (r *heroAwardRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroAward, error) {
	query := r.sb.Select(
		"ha.hero_id", "ha.award_id", "a.name", "ha.award_date", "ha.decree_number",
	).
		From("hero_awards ha").
		Join("awards a ON ha.award_id = a.id").
		Where(squirrel.Eq{"ha.hero_id": heroID}).
		OrderBy("a.sort_order ASC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list hero awards: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list hero awards: %w", err)
	}
	defer rows.Close()

	var awards []*domain.HeroAward
	for rows.Next() {
		ha := &domain.HeroAward{}
		if err := rows.Scan(&ha.HeroID, &ha.AwardID, &ha.AwardName, &ha.AwardDate, &ha.DecreeNumber); err != nil {
			return nil, fmt.Errorf("scan hero award: %w", err)
		}
		awards = append(awards, ha)
	}
	return awards, rows.Err()
}
