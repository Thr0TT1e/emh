package pg

import (
	"context"
	"encoding/json"
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

func (r *heroAwardRepository) Add(ctx context.Context, p domain.AddHeroAwardParams) error {
	// award_date и award_date_info — независимые колонки:
	// contract отдаёт приоритет info, но legacy-колонка пишется как есть (может быть NULL).
	query := r.sb.Insert("hero_awards").
		Columns("hero_id", "award_id", "award_date", "decree_number", "award_date_info").
		Values(p.HeroID, p.AwardID, p.AwardDate, p.DecreeNumber, flexDateToJSON(p.AwardDateInfo)).
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
	// Денормализация полей справочника для орденской планки (ADR-005):
	// COALESCE для ribbon_image_url — у строк, созданных до миграции 00027, значение NULL.
	query := r.sb.Select(
		"ha.hero_id", "ha.award_id", "a.name", "ha.award_date", "ha.decree_number",
		"ha.award_date_info",
		"COALESCE(a.ribbon_image_url, '') AS ribbon_image_url",
		"COALESCE(a.image_url, '') AS image_url",
		"a.type", "a.jurisdiction", "a.worn_without_bar", "a.is_jubilee",
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
		var awardType, jurisdiction int
		var awardDateInfoJSON []byte
		if err := rows.Scan(
			&ha.HeroID, &ha.AwardID, &ha.AwardName, &ha.AwardDate, &ha.DecreeNumber,
			&awardDateInfoJSON,
			&ha.RibbonImageURL, &ha.ImageURL, &awardType, &jurisdiction,
			&ha.WornWithoutBar, &ha.IsJubilee,
		); err != nil {
			return nil, fmt.Errorf("scan hero award: %w", err)
		}
		// Десериализация гибкой даты (паттерн как в scanConflict): битый JSON не роняет выдачу.
		if len(awardDateInfoJSON) > 0 {
			var fd domain.FlexibleDate
			if err := json.Unmarshal(awardDateInfoJSON, &fd); err == nil {
				ha.AwardDateInfo = &fd
			}
		}
		ha.Type = domain.AwardType(awardType)
		ha.Jurisdiction = domain.AwardJurisdiction(jurisdiction)
		awards = append(awards, ha)
	}
	return awards, rows.Err()
}
