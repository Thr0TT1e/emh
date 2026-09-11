package pg

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type heroLocationRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewHeroLocationRepository(pool *pgxpool.Pool) repository.HeroLocationRepository {
	return &heroLocationRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *heroLocationRepository) Add(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	query := r.sb.Insert("hero_locations").
		Columns("hero_id", "location_id", "type").
		Values(heroID, locationID, int(locType)).
		Suffix("ON CONFLICT (hero_id, location_id, type) DO NOTHING")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build add hero location: %w", err)
	}
	if _, err := r.pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec add hero location: %w", err)
	}
	return nil
}

func (r *heroLocationRepository) Remove(ctx context.Context, heroID, locationID string, locType domain.HeroLocationType) error {
	query := r.sb.Delete("hero_locations").
		Where(squirrel.Eq{"hero_id": heroID, "location_id": locationID, "type": int(locType)})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build remove hero location: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec remove hero location: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero location not found")
	}
	return nil
}

func (r *heroLocationRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.HeroLocation, error) {
	query := r.sb.Select(
		"hl.hero_id", "hl.type",
		"l.id", "l.name", "l.historical_name", "l.type", "l.parent_id", "l.latitude", "l.longitude",
	).
		From("hero_locations hl").
		Join("locations l ON hl.location_id = l.id").
		Where(squirrel.Eq{"hl.hero_id": heroID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list hero locations: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list hero locations: %w", err)
	}
	defer rows.Close()

	var locations []*domain.HeroLocation
	for rows.Next() {
		hl := &domain.HeroLocation{Location: &domain.Location{}}
		var locType int
		if err := rows.Scan(
			&hl.HeroID, &locType,
			&hl.Location.ID, &hl.Location.Name, &hl.Location.HistoricalName,
			&hl.Location.Type, &hl.Location.ParentID, &hl.Location.Latitude, &hl.Location.Longitude,
		); err != nil {
			return nil, fmt.Errorf("scan hero location: %w", err)
		}
		hl.LocationID = hl.Location.ID
		hl.Type = domain.HeroLocationType(locType)
		locations = append(locations, hl)
	}
	return locations, rows.Err()
}
