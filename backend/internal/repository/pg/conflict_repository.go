package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// conflictColumns список колонок для единообразия в SELECT/RETURNING.
const conflictColumns = "id, name, description, type, start_date, end_date, parent_conflict_id, start_date_info, end_date_info, created_at, updated_at"

type conflictRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewConflictRepository(pool *pgxpool.Pool) repository.ConflictRepository {
	return &conflictRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// scannable абстракция над pgx.Row и pgx.Rows для переиспользования сканера.
type scannable interface {
	Scan(dest ...any) error
}

// Хелпер для сериализации в JSONB
func flexDateToJSON(d *domain.FlexibleDate) []byte {
	if d == nil {
		return nil
	}
	b, _ := json.Marshal(d)
	return b
}

// scanConflict маппит строку БД в доменную модель.
func scanConflict(s scannable) (*domain.Conflict, error) {
	c := &domain.Conflict{}
	var conflictType int
	var startDateInfoJSON, endDateInfoJSON []byte
	if err := s.Scan(
		&c.ID, &c.Name, &c.Description, &conflictType,
		&c.StartDate, &c.EndDate, &c.ParentConflictID,
		&startDateInfoJSON, &endDateInfoJSON,
		&c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, err
	}
	c.Type = domain.ConflictType(conflictType)

	// Десериализация гибких дат
	if len(startDateInfoJSON) > 0 {
		var fd domain.FlexibleDate
		if err := json.Unmarshal(startDateInfoJSON, &fd); err == nil {
			c.StartDateInfo = &fd
		}
	}
	if len(endDateInfoJSON) > 0 {
		var fd domain.FlexibleDate
		if err := json.Unmarshal(endDateInfoJSON, &fd); err == nil {
			c.EndDateInfo = &fd
		}
	}

	return c, nil
}

func (r *conflictRepository) Create(ctx context.Context, p domain.CreateConflictParams) (string, error) {
	query := r.sb.Insert("conflicts").
		Columns("name", "description", "type", "start_date", "end_date", "parent_conflict_id", "start_date_info", "end_date_info").
		Values(p.Name, p.Description, int(p.Type), p.StartDate, p.EndDate, p.ParentConflictID,
			flexDateToJSON(p.StartDateInfo), flexDateToJSON(p.EndDateInfo)).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create conflict: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create conflict: %w", err)
	}
	return id, nil
}

func (r *conflictRepository) Update(ctx context.Context, p domain.UpdateConflictParams) (*domain.Conflict, error) {
	update := r.sb.Update("conflicts").Where(squirrel.Eq{"id": p.ID})

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
	if mask["type"] {
		update = update.Set("type", int(p.Type))
	}
	if mask["start_date"] {
		update = update.Set("start_date", p.StartDate)
	}
	if mask["end_date"] {
		update = update.Set("end_date", p.EndDate)
	}
	if mask["parent_conflict_id"] {
		update = update.Set("parent_conflict_id", p.ParentConflictID)
	}
	if mask["start_date_info"] {
		update = update.Set("start_date_info", flexDateToJSON(p.StartDateInfo))
	}
	if mask["end_date_info"] {
		update = update.Set("end_date_info", flexDateToJSON(p.EndDateInfo))
	}

	update = update.Set("updated_at", time.Now())
	update = update.Suffix("RETURNING " + conflictColumns)

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update conflict: %w", err)
	}

	c, err := scanConflict(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("conflict not found")
		}
		return nil, fmt.Errorf("exec update conflict: %w", err)
	}
	return c, nil
}

func (r *conflictRepository) Delete(ctx context.Context, id string) error {
	query := r.sb.Delete("conflicts").Where(squirrel.Eq{"id": id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete conflict: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec delete conflict: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("conflict not found")
	}
	return nil
}

func (r *conflictRepository) GetByID(ctx context.Context, id string) (*domain.Conflict, error) {
	query := r.sb.Select(conflictColumns).
		From("conflicts").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get conflict: %w", err)
	}

	c, err := scanConflict(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("conflict not found")
		}
		return nil, fmt.Errorf("exec get conflict: %w", err)
	}
	return c, nil
}

func (r *conflictRepository) List(ctx context.Context, f domain.ConflictFilter) ([]*domain.Conflict, string, int64, error) {
	// 1. COUNT(*) для total_count
	countQuery := r.sb.Select("COUNT(*)").From("conflicts")
	if f.Type != domain.ConflictTypeUnspecified {
		countQuery = countQuery.Where(squirrel.Eq{"type": int(f.Type)})
	}
	if f.ParentID != "" {
		countQuery = countQuery.Where(squirrel.Eq{"parent_conflict_id": f.ParentID})
	}
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build count conflicts: %w", err)
	}
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("exec count conflicts: %w", err)
	}

	// 2. SELECT с пагинацией
	sel := r.sb.Select(conflictColumns).
		From("conflicts").
		OrderBy("start_date ASC NULLS LAST, name ASC, id ASC")

	if f.Type != domain.ConflictTypeUnspecified {
		sel = sel.Where(squirrel.Eq{"type": int(f.Type)})
	}
	if f.ParentID != "" {
		sel = sel.Where(squirrel.Eq{"parent_conflict_id": f.ParentID})
	}
	if f.Cursor != "" {
		sel = sel.Where(squirrel.Gt{"id": f.Cursor})
	}

	limit := f.Limit + 1 // +1 для определения has_next_page
	sel = sel.Limit(uint64(limit))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list conflicts: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("exec list conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []*domain.Conflict
	for rows.Next() {
		c, err := scanConflict(rows)
		if err != nil {
			return nil, "", 0, fmt.Errorf("scan conflict: %w", err)
		}
		conflicts = append(conflicts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration error: %w", err)
	}

	// 3. Определяем next_cursor
	var nextCursor string
	if len(conflicts) > f.Limit {
		conflicts = conflicts[:f.Limit] // обрезаем до limit
		nextCursor = conflicts[len(conflicts)-1].ID
	}

	return conflicts, nextCursor, total, nil
}

func (r *conflictRepository) HasHeroes(ctx context.Context, conflictID string) (bool, error) {
	query := r.sb.Select("COUNT(*)").
		From("hero_conflicts").
		Where(squirrel.Eq{"conflict_id": conflictID})

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
// См. документацию в location_repository.go.
func (r *conflictRepository) HasCyclicReference(ctx context.Context, selfID, candidateParentID string) (bool, error) {
	const query = `
        WITH RECURSIVE ancestors AS (
            SELECT id, parent_conflict_id, 1 AS depth
            FROM conflicts
            WHERE id = $1
            UNION ALL
            SELECT c.id, c.parent_conflict_id, a.depth + 1
            FROM conflicts c
            INNER JOIN ancestors a ON c.id = a.parent_conflict_id
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
