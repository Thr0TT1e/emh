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

// locationColumns список колонок для единообразия в SELECT/RETURNING.
const locationColumns = "id, name, historical_name, type, parent_id, latitude, longitude, created_at, updated_at"

type locationRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewLocationRepository(pool *pgxpool.Pool) repository.LocationRepository {
	return &locationRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// scanLocation маппит строку БД в доменную модель.
// Интерфейс scannable переиспользуется из conflict_repository.go (общий для пакета pg).
func scanLocation(s scannable) (*domain.Location, error) {
	l := &domain.Location{}
	var locType int
	if err := s.Scan(
		&l.ID, &l.Name, &l.HistoricalName, &locType,
		&l.ParentID, &l.Latitude, &l.Longitude,
		&l.CreatedAt, &l.UpdatedAt,
	); err != nil {
		return nil, err
	}
	l.Type = domain.LocationType(locType)
	return l, nil
}

func (r *locationRepository) Create(ctx context.Context, p domain.CreateLocationParams) (string, error) {
	query := r.sb.Insert("locations").
		Columns("name", "historical_name", "type", "parent_id", "latitude", "longitude").
		Values(p.Name, p.HistoricalName, int(p.Type), p.ParentID, p.Latitude, p.Longitude).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create location: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create location: %w", err)
	}
	return id, nil
}

func (r *locationRepository) Update(ctx context.Context, p domain.UpdateLocationParams) (*domain.Location, error) {
	update := r.sb.Update("locations").Where(squirrel.Eq{"id": p.ID})

	mask := make(map[string]bool, len(p.FieldMask))
	for _, f := range p.FieldMask {
		mask[f] = true
	}

	if mask["name"] {
		update = update.Set("name", p.Name)
	}
	if mask["historical_name"] {
		update = update.Set("historical_name", p.HistoricalName)
	}
	if mask["type"] {
		update = update.Set("type", int(p.Type))
	}
	if mask["parent_id"] {
		update = update.Set("parent_id", p.ParentID)
	}
	if mask["latitude"] {
		update = update.Set("latitude", p.Latitude)
	}
	if mask["longitude"] {
		update = update.Set("longitude", p.Longitude)
	}

	update = update.Set("updated_at", time.Now())
	update = update.Suffix("RETURNING " + locationColumns)

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update location: %w", err)
	}

	l, err := scanLocation(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, fmt.Errorf("exec update location: %w", err)
	}
	return l, nil
}

func (r *locationRepository) Delete(ctx context.Context, id string) error {
	query := r.sb.Delete("locations").Where(squirrel.Eq{"id": id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete location: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec delete location: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("location not found")
	}
	return nil
}

func (r *locationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	query := r.sb.Select(locationColumns).
		From("locations").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get location: %w", err)
	}

	l, err := scanLocation(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, fmt.Errorf("exec get location: %w", err)
	}
	return l, nil
}

func (r *locationRepository) List(ctx context.Context, f domain.LocationFilter) ([]*domain.Location, string, int64, error) {
	// 1. COUNT(*)
	countQuery := r.sb.Select("COUNT(*)").From("locations")
	if f.ParentID != "" {
		countQuery = countQuery.Where(squirrel.Eq{"parent_id": f.ParentID})
	}
	if f.SearchQuery != "" {
		countQuery = countQuery.Where(squirrel.Or{
			squirrel.ILike{"name": "%" + f.SearchQuery + "%"},
			squirrel.ILike{"historical_name": "%" + f.SearchQuery + "%"},
		})
	}
	if f.Type != domain.LocationTypeUnspecified {
		countQuery = countQuery.Where(squirrel.Eq{"type": int(f.Type)})
	}
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build count locations: %w", err)
	}
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("exec count locations: %w", err)
	}

	// 2. SELECT с пагинацией
	sel := r.sb.Select(locationColumns).
		From("locations").
		OrderBy("name ASC, id ASC")

	if f.ParentID != "" {
		sel = sel.Where(squirrel.Eq{"parent_id": f.ParentID})
	}
	if f.SearchQuery != "" {
		sel = sel.Where(squirrel.Or{
			squirrel.ILike{"name": "%" + f.SearchQuery + "%"},
			squirrel.ILike{"historical_name": "%" + f.SearchQuery + "%"},
		})
	}
	if f.Type != domain.LocationTypeUnspecified {
		sel = sel.Where(squirrel.Eq{"type": int(f.Type)})
	}
	if f.Cursor != "" {
		sel = sel.Where(squirrel.Gt{"id": f.Cursor})
	}

	limit := f.Limit + 1
	sel = sel.Limit(uint64(limit))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list locations: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("exec list locations: %w", err)
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, "", 0, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, l)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration error: %w", err)
	}

	var nextCursor string
	if len(locations) > f.Limit {
		locations = locations[:f.Limit]
		nextCursor = locations[len(locations)-1].ID
	}

	return locations, nextCursor, total, nil
}

// Reparent массово переносит дочерние локации от одного родителя к другому.
// Если newParentID пустая, дети становятся корневыми (parent_id = NULL).
func (r *locationRepository) Reparent(ctx context.Context, oldParentID, newParentID string) (int, error) {
	// Пустая строка означает NULL (корневая локация), а не литерал ""
	var newParent interface{}
	if newParentID == "" {
		newParent = nil
	} else {
		newParent = newParentID
	}

	update := r.sb.Update("locations").
		Set("parent_id", newParent).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"parent_id": oldParentID})

	sql, args, err := update.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build reparent locations: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("exec reparent locations: %w", err)
	}
	return int(ct.RowsAffected()), nil
}

func (r *locationRepository) HasHeroes(ctx context.Context, locationID string) (bool, error) {
	query := r.sb.Select("COUNT(*)").
		From("hero_locations").
		Where(squirrel.Eq{"location_id": locationID})

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

// HasCyclicReference проверяет наличие транзитивного цикла через WITH RECURSIVE.
//
// Идея: строим цепочку предков candidateParentID (идём вверх по parent_id).
// Если среди предков встречается selfID — значит candidateParentID уже является
// потомком selfID, и установка parent_id = candidateParentID для selfID создаст цикл.
//
// Пример: есть A → B → C.
//   - HasCyclicReference(C, A): предки A = {} → C не найден → false (OK, можно делать A.parent=C? НЕТ!)
//   - HasCyclicReference(A, C): предки C = {B, A} → A найден → true (цикл, отклоняем)
//
// Глубина ограничена 100 уровнями как защита от бесконечной рекурсии
// при уже сломанных данных.
func (r *locationRepository) HasCyclicReference(ctx context.Context, selfID, candidateParentID string) (bool, error) {
	const query = `
        WITH RECURSIVE ancestors AS (
            SELECT id, parent_id, 1 AS depth
            FROM locations
            WHERE id = $1
            UNION ALL
            SELECT l.id, l.parent_id, a.depth + 1
            FROM locations l
            INNER JOIN ancestors a ON l.id = a.parent_id
            WHERE a.depth < 100
        )
        SELECT EXISTS(SELECT 1 FROM ancestors WHERE id = $2)
    `
	var hasCycle bool
	if err := r.pool.QueryRow(ctx, query, candidateParentID, selfID).Scan(&hasCycle); err != nil {
		return false, fmt.Errorf("check cyclic reference: %w", err)
	}
	return hasCycle, nil
}
