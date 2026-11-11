package pg

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type awardRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

// awardColumns список колонок для чтения награды.
// COALESCE для nullable-колонок (image_url/description — NULL с миграции 00002,
// ribbon_image_url — NULL у строк до 00027): домен хранит пустые строки,
// pgx не сканирует NULL в *string (прецедент — фикс ribbon_image_url).
const awardColumns = `id, name, COALESCE(description, '') AS description, COALESCE(image_url, '') AS image_url, sort_order,
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
		// Конфликт idx_awards_seed_key (SQLSTATE 23505) — дубль составного ключа.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.ErrAwardDuplicate
		}
		return "", fmt.Errorf("exec create award: %w", err)
	}
	return id, nil
}

// ExistsByKey — проверка существования награды по составному ключу сида.
func (r *awardRepository) ExistsByKey(ctx context.Context, name string,
	awardType domain.AwardType, jurisdiction domain.AwardJurisdiction,
) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM awards WHERE name = $1 AND type = $2 AND jurisdiction = $3)`,
		name, int(awardType), int(jurisdiction),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check award exists by key: %w", err)
	}
	return exists, nil
}

// InsertIfAbsent — идемпотентная вставка: ON CONFLICT по idx_awards_seed_key
// (таргет совпадает с полным уникальным индексом без предиката, миграция 00028).
// id генерирует DEFAULT uuidv7() — как в Create.
func (r *awardRepository) InsertIfAbsent(ctx context.Context, p domain.CreateAwardParams) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO awards (name, description, image_url, sort_order,
		                    ribbon_image_url, type, jurisdiction, worn_without_bar, is_jubilee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (name, type, jurisdiction) DO NOTHING`,
		p.Name, p.Description, p.ImageURL, p.SortOrder,
		p.RibbonImageURL, int(p.Type), int(p.Jurisdiction), p.WornWithoutBar, p.IsJubilee,
	)
	if err != nil {
		return false, fmt.Errorf("insert award if absent: %w", err)
	}
	return tag.RowsAffected() > 0, nil
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
		// Смена name/type/jurisdiction могла попасть в idx_awards_seed_key.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrAwardDuplicate
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

// awardCursor — keyset-курсор списка наград: позиции последней строки страницы
// в порядке ORDER BY sort_order, name, id. Непрозрачный для клиента (base64url JSON).
type awardCursor struct {
	SortOrder int    `json:"s"`
	Name      string `json:"n"`
	ID        string `json:"i"`
}

// encodeAwardCursor сериализует курсор строки в непрозрачный токен.
func encodeAwardCursor(a *domain.Award) string {
	data, err := json.Marshal(awardCursor{SortOrder: a.SortOrder, Name: a.Name, ID: a.ID})
	if err != nil {
		// domain.Awardmarshal'ится всегда; ветка недостижима, fail-fast при порче модели.
		panic(fmt.Sprintf("encode award cursor: %v", err))
	}
	return base64.URLEncoding.EncodeToString(data)
}

// decodeAwardCursor разбирает токен; пустой вход — пустой курсор (первая страница).
func decodeAwardCursor(raw string) (awardCursor, error) {
	if raw == "" {
		return awardCursor{}, nil
	}
	data, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return awardCursor{}, fmt.Errorf("%w: %v", domain.ErrInvalidCursor, err)
	}
	var c awardCursor
	if err := json.Unmarshal(data, &c); err != nil {
		return awardCursor{}, fmt.Errorf("%w: %v", domain.ErrInvalidCursor, err)
	}
	if c.ID == "" {
		return awardCursor{}, fmt.Errorf("%w: empty id", domain.ErrInvalidCursor)
	}
	return c, nil
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
		c, err := decodeAwardCursor(f.Cursor)
		if err != nil {
			return nil, "", 0, err
		}
		// Keyset: лексикографическое сравнение кортежа ровно в порядке ORDER BY.
		// Старый WHERE id > cursor терял записи (сортировка не по id) — ADR-005.
		sel = sel.Where(squirrel.Expr("(sort_order, name, id) > (?, ?, ?)", c.SortOrder, c.Name, c.ID))
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
		nextCursor = encodeAwardCursor(awards[len(awards)-1])
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
