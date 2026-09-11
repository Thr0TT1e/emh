package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type heroRelationRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewHeroRelationRepository(pool *pgxpool.Pool) repository.HeroRelationRepository {
	return &heroRelationRepository{pool: pool, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *heroRelationRepository) Add(ctx context.Context, p domain.AddHeroRelationParams) (string, error) {
	query := r.sb.Insert("hero_relations").
		Columns("from_hero_id", "to_hero_id", "relation_type", "description").
		Values(p.FromHeroID, p.ToHeroID, p.RelationType, p.Description).
		Suffix("ON CONFLICT (from_hero_id, to_hero_id, relation_type) DO NOTHING RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build add hero relation: %w", err)
	}
	var id string
	// При конфликте (дубликат) RETURNING не вернёт строк — это не ошибка
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec add hero relation: %w", err)
	}
	return id, nil
}

func (r *heroRelationRepository) Remove(ctx context.Context, id string) error {
	query := r.sb.Delete("hero_relations").Where(squirrel.Eq{"id": id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build remove hero relation: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec remove hero relation: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero relation not found")
	}
	return nil
}

// ListByHero возвращает связи героя в обе стороны с денормализацией имени.
func (r *heroRelationRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroRelation, error) {
	// UNION: связи где герой является from_hero_id И где to_hero_id
	query := r.sb.Select(
		"hr.id", "hr.from_hero_id", "hr.to_hero_id", "hr.relation_type", "hr.description",
		"CONCAT_WS(' ', h.last_name, h.first_name, h.middle_name) AS related_hero_name",
	).
		From("hero_relations hr").
		Join("heroes h ON h.id = CASE WHEN hr.from_hero_id = ? THEN hr.to_hero_id ELSE hr.from_hero_id END", heroID).
		Where(squirrel.Or{
			squirrel.Eq{"hr.from_hero_id": heroID},
			squirrel.Eq{"hr.to_hero_id": heroID},
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list hero relations: %w", err)
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list hero relations: %w", err)
	}
	defer rows.Close()

	var relations []*domain.HeroRelation
	for rows.Next() {
		rel := &domain.HeroRelation{}
		if err := rows.Scan(&rel.ID, &rel.FromHeroID, &rel.ToHeroID, &rel.RelationType, &rel.Description, &rel.RelatedHeroName); err != nil {
			return nil, fmt.Errorf("scan hero relation: %w", err)
		}
		relations = append(relations, rel)
	}
	return relations, rows.Err()
}
