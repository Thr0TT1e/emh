package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/pkg/escape"
)

// heroFullColumns все колонки для детальной карточки.
const (
	heroFullColumns = `
		id,
		first_name,
		last_name,
		middle_name,
		short_bio,
		full_bio,
		rank,
		birth_date,
		birth_date_precision,
		birth_date_display,
		death_date,
		death_date_precision,
		death_date_display,
		status,
		nickname,
		unit,
		position,
		service_branch,
		cause_of_death,
		service_start_date,
		service_start_date_precision,
		service_start_date_display,
		memberships,
		created_at,
		updated_at
	`
)

type heroRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

// NewHeroRepository создает PostgreSQL-реализацию репозитория.
func NewHeroRepository(pool *pgxpool.Pool) repository.HeroRepository {
	return &heroRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *heroRepository) Create(ctx context.Context, p domain.CreateHeroParams) (string, error) {
	birthAnchor, birthPrecision, birthDisplay := flexibleDateParts(&p.BirthDate)
	deathAnchor, deathPrecision, deathDisplay := flexibleDateParts(&p.DeathDate)
	serviceAnchor, servicePrecision, serviceDisplay := flexibleDateParts(&p.ServiceStartDate)

	query := r.sb.Insert("heroes").
		Columns(
			"first_name",
			"last_name",
			"middle_name",
			"short_bio",
			"full_bio",
			"rank",
			"birth_date",
			"birth_date_precision",
			"birth_date_display",
			"death_date",
			"death_date_precision",
			"death_date_display",
			"status",
			"nickname",
			"unit",
			"position",
			"service_branch",
			"cause_of_death",
			"service_start_date",
			"service_start_date_precision",
			"service_start_date_display",
			"memberships",
		).
		Values(
			p.FirstName,
			p.LastName,
			p.MiddleName,
			p.ShortBio,
			p.FullBio,
			p.Rank,
			birthAnchor,
			birthPrecision,
			birthDisplay,
			deathAnchor,
			deathPrecision,
			deathDisplay,
			int(p.Status),
			p.Nickname,
			p.Unit,
			p.Position,
			p.ServiceBranch,
			p.CauseOfDeath,
			serviceAnchor,
			servicePrecision,
			serviceDisplay,
			squirrel.Expr("?::jsonb", marshalMemberships(p.Memberships)),
		).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create query: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}

	return id, nil
}

func (r *heroRepository) Update(ctx context.Context, p domain.UpdateHeroParams) (*domain.Hero, error) {
	if p.ID == nil || *p.ID == "" {
		return nil, fmt.Errorf("hero id is required")
	}

	update := r.sb.Update("heroes").Where(squirrel.Eq{"id": *p.ID})

	mask := make(map[string]bool, len(p.FieldMask))
	for _, f := range p.FieldMask {
		mask[f] = true
	}

	if mask["first_name"] {
		update = update.Set("first_name", stringPtrValue(p.FirstName))
	}

	if mask["last_name"] {
		update = update.Set("last_name", stringPtrValue(p.LastName))
	}

	if mask["middle_name"] {
		update = update.Set("middle_name", stringPtrValue(p.MiddleName))
	}

	if mask["short_bio"] {
		update = update.Set("short_bio", stringPtrValue(p.ShortBio))
	}

	if mask["full_bio"] {
		update = update.Set("full_bio", stringPtrValue(p.FullBio))
	}

	if mask["rank"] {
		update = update.Set("rank", stringPtrValue(p.Rank))
	}

	if mask["birth_date"] {
		anchor, precision, display := flexibleDateParts(p.BirthDate)
		update = update.
			Set("birth_date", anchor).
			Set("birth_date_precision", precision).
			Set("birth_date_display", display)
	}

	if mask["death_date"] {
		anchor, precision, display := flexibleDateParts(p.DeathDate)
		update = update.
			Set("death_date", anchor).
			Set("death_date_precision", precision).
			Set("death_date_display", display)
	}

	if mask["status"] {
		if p.Status == nil {
			return nil, fmt.Errorf("status is required when updating status")
		}

		update = update.Set("status", int(*p.Status))
	}

	if mask["nickname"] {
		update = update.Set("nickname", stringPtrValue(p.Nickname))
	}

	if mask["unit"] {
		update = update.Set("unit", stringPtrValue(p.Unit))
	}

	if mask["position"] {
		update = update.Set("position", stringPtrValue(p.Position))
	}

	if mask["service_branch"] {
		update = update.Set("service_branch", stringPtrValue(p.ServiceBranch))
	}

	if mask["cause_of_death"] {
		update = update.Set("cause_of_death", stringPtrValue(p.CauseOfDeath))
	}

	if mask["service_start_date"] {
		anchor, precision, display := flexibleDateParts(p.ServiceStartDate)
		update = update.
			Set("service_start_date", anchor).
			Set("service_start_date_precision", precision).
			Set("service_start_date_display", display)
	}

	if mask["memberships"] {
		update = update.Set(
			"memberships",
			squirrel.Expr("?::jsonb", marshalMemberships(p.Memberships)),
		)
	}

	update = update.
		Set("updated_at", time.Now()).
		Suffix("RETURNING " + heroFullColumns)

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	h, err := scanHeroFull(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("hero %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("exec update: %w", err)
	}

	return h, nil
}

func (r *heroRepository) Delete(ctx context.Context, id string, hardDelete bool) error {
	if hardDelete {
		query := r.sb.Delete("heroes").Where(squirrel.Eq{"id": id})
		sql, args, err := query.ToSql()
		if err != nil {
			return fmt.Errorf("build hard delete: %w", err)
		}

		ct, err := r.pool.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("exec hard delete: %w", err)
		}

		if ct.RowsAffected() == 0 {
			return fmt.Errorf("hero %w", domain.ErrNotFound)
		}
		return nil
	}

	// Soft delete: архивируем
	update := r.sb.Update("heroes").
		Set("status", int(domain.StatusArchived)).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id})

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build soft delete: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec soft delete: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("hero %w", domain.ErrNotFound)
	}

	return nil
}

func (r *heroRepository) GetByID(ctx context.Context, id string) (*domain.Hero, error) {
	query := r.sb.Select(heroFullColumns).From("heroes").Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get query: %w", err)
	}

	h, err := scanHeroFull(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("hero %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("exec get: %w", err)
	}

	return h, nil
}

// List реализует cursor-based пагинацию с денормализацией главного фото и наград.
// Итерация 2: заменены 7 коррелированных подзапросов на 2 LEFT JOIN LATERAL.
// Было: 20 героев × 7 подзапросов = 140 запросов.
// Стало: 1 запрос на страницу.
func (r *heroRepository) List(ctx context.Context, f domain.HeroFilter) ([]*domain.Hero, string, int64, error) {
	sel := r.sb.Select(
		"h.id",
		"h.first_name",
		"h.last_name",
		"h.middle_name",
		"h.short_bio",
		"h.rank",
		"h.birth_date",
		"h.birth_date_precision",
		"h.birth_date_display",
		"h.death_date",
		"h.death_date_precision",
		"h.death_date_display",
		"h.status",
		"h.nickname",
		"h.unit",
		"h.service_branch",
		// Главное фото: одно фото на героя с приоритетом is_main,
		// fallback на первое по sort_order. LATERAL гарантирует
		// выполнение подзапроса один раз на героя, а не на каждый столбец.
		"mp.url AS main_photo_url",
		// Превью: берём thumbnail если есть, иначе fallback на url.
		"COALESCE(NULLIF(mp.thumbnail_url, ''), mp.url) AS main_thumbnail_url",
		// Награды: агрегация имён в массив.
		"aw.names AS award_names",
	).From("heroes h")

	// LATERAL JOIN для главного фото.
	// ORDER BY is_main DESC ставит главное фото первым;
	// если его нет — берётся первое по sort_order, id.
	// COALESCE(NULLIF(thumbnail_url, ''), url) — fallback если
	// thumbnail ещё не сгенерирован воркером.
	sel = sel.
		LeftJoin(`LATERAL (
			SELECT p.url, p.thumbnail_url
			FROM photos p
			WHERE p.hero_id = h.id
			ORDER BY p.is_main DESC, p.sort_order ASC, p.id ASC
			LIMIT 1
		) mp ON true`)

	// LATERAL JOIN для наград.
	// Если наград нет — aw.names = NULL, ARRAY_AGG вернёт пустой массив
	// через COALESCE в scan.
	sel = sel.
		LeftJoin(`LATERAL (
			SELECT ARRAY_AGG(a.name ORDER BY a.sort_order) AS names
			FROM hero_awards ha
			JOIN awards a ON ha.award_id = a.id
			WHERE ha.hero_id = h.id
		) aw ON true`)

	sel = applyHeroFilters(sel, f)

	if f.Cursor != "" {
		sel = sel.Where(squirrel.Lt{"h.id": f.Cursor})
	}

	sel = sel.OrderBy("h.id DESC").Limit(uint64(f.Limit + 1))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("list heroes: %w", err)
	}
	defer rows.Close()

	heroes := make([]*domain.Hero, 0, f.Limit)

	for rows.Next() {
		h := &domain.Hero{}

		var (
			birthAnchor    *time.Time
			deathAnchor    *time.Time
			birthDisplay   *string
			deathDisplay   *string
			birthPrecision int
			deathPrecision int
			status         int
		)

		if err := rows.Scan(
			&h.ID,
			&h.FirstName,
			&h.LastName,
			&h.MiddleName,
			&h.ShortBio,
			&h.Rank,
			&birthAnchor,
			&birthPrecision,
			&birthDisplay,
			&deathAnchor,
			&deathPrecision,
			&deathDisplay,
			&status,
			&h.Nickname,
			&h.Unit,
			&h.ServiceBranch,
			&h.MainPhotoURL,
			&h.MainThumbnailURL,
			&h.AwardNames,
		); err != nil {
			return nil, "", 0, fmt.Errorf("scan hero row: %w", err)
		}

		h.Status = domain.PublicationStatus(status)
		h.BirthDate = buildFlexibleDate(birthAnchor, birthPrecision, birthDisplay)
		h.DeathDate = buildFlexibleDate(deathAnchor, deathPrecision, deathDisplay)
		heroes = append(heroes, h)
	}

	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration error: %w", err)
	}

	var nextCursor string
	if len(heroes) > f.Limit {
		heroes = heroes[:f.Limit]
		nextCursor = heroes[len(heroes)-1].ID
	}

	// Total считаем только для первой страницы (без курсора).
	var total int64
	if f.Cursor == "" {
		total, err = r.countHeroes(ctx, f)
		if err != nil {
			return nil, "", 0, err
		}
	}

	return heroes, nextCursor, total, nil
}

// countHeroes возвращает общее количество героев с учётом фильтров.
func (r *heroRepository) countHeroes(ctx context.Context, f domain.HeroFilter) (int64, error) {
	// ВАЖНО: было COUNT(), исправлено на COUNT(*)
	countSel := applyHeroFilters(r.sb.Select("COUNT(*)").From("heroes h"), f)
	sql, args, err := countSel.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}
	var total int64
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("exec count: %w", err)
	}
	return total, nil
}

// applyHeroFilters применяет фильтры поиска к билдеру.
func applyHeroFilters(sel squirrel.SelectBuilder, f domain.HeroFilter) squirrel.SelectBuilder {
	// 1. Безопасная фильтрация по статусу
	if f.Status != nil {
		// Явный запрос конкретного статуса (Published, Draft или Archived)
		sel = sel.Where(squirrel.Eq{"h.status": int(*f.Status)})
	} else if !f.IncludeArchived {
		// По умолчанию (для публичного API или админки без флага "включить архивные")
		// исключаем архивные записи, показывая только Published и Draft
		sel = sel.Where(squirrel.NotEq{"h.status": int(domain.StatusArchived)})
	}

	// 2. Остальные фильтры
	if f.SearchQuery != "" {
		escaped := escape.LIKE(f.SearchQuery)
		sel = sel.Where(squirrel.Or{
			squirrel.Expr("h.first_name ILIKE ? ESCAPE '\\'", "%"+escaped+"%"),
			squirrel.Expr("h.last_name ILIKE ? ESCAPE '\\'", "%"+escaped+"%"),
			squirrel.Expr("h.middle_name ILIKE ? ESCAPE '\\'", "%"+escaped+"%"),
			squirrel.Expr("h.nickname ILIKE ? ESCAPE '\\'", "%"+escaped+"%"),
		})
	}

	if len(f.SearchWords) > 0 {
		// websearch_to_tsquery безопасно парсит входную строку,
		// экранируя спецсимволы. Слова соединяются через пробел →
		// PostgreSQL интерпретирует их как AND.
		query := strings.Join(f.SearchWords, " ")
		sel = sel.Where(
			"h.search_vector @@ websearch_to_tsquery('russian', ?)",
			query,
		)
	}

	if f.ConflictID != "" {
		sel = sel.Join("hero_conflicts hc ON h.id = hc.hero_id").
			Where(squirrel.Eq{"hc.conflict_id": f.ConflictID})
	}
	if f.LocationID != "" {
		sel = sel.Join("hero_locations hl ON h.id = hl.hero_id").
			Where(squirrel.Eq{"hl.location_id": f.LocationID})
	}
	if f.DateFrom != nil {
		sel = sel.Where(squirrel.GtOrEq{"h.death_date": f.DateFrom})
	}
	if f.DateTo != nil {
		sel = sel.Where(squirrel.LtOrEq{"h.death_date": f.DateTo})
	}
	return sel
}

// splitSearchWords разбивает строку на отдельные слова для поиска.
// Пустые слова игнорируются.
func splitSearchWords(query string) []string {
	fields := strings.Fields(query)
	words := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			words = append(words, f)
		}
	}
	return words
}

// scanHeroFull маппит полную строку героя.
func scanHeroFull(s scannable) (*domain.Hero, error) {
	h := &domain.Hero{}

	var (
		membershipsJSON []byte

		birthAnchor   *time.Time
		deathAnchor   *time.Time
		serviceAnchor *time.Time

		birthPrecision   int
		deathPrecision   int
		servicePrecision int

		birthDisplay   *string
		deathDisplay   *string
		serviceDisplay *string

		status int
	)

	if err := s.Scan(
		&h.ID,
		&h.FirstName,
		&h.LastName,
		&h.MiddleName,
		&h.ShortBio,
		&h.FullBio,
		&h.Rank,
		&birthAnchor,
		&birthPrecision,
		&birthDisplay,
		&deathAnchor,
		&deathPrecision,
		&deathDisplay,
		&status,
		&h.Nickname,
		&h.Unit,
		&h.Position,
		&h.ServiceBranch,
		&h.CauseOfDeath,
		&serviceAnchor,
		&servicePrecision,
		&serviceDisplay,
		&membershipsJSON,
		&h.CreatedAt,
		&h.UpdatedAt,
	); err != nil {
		return nil, err
	}

	h.Status = domain.PublicationStatus(status)

	h.BirthDate = buildFlexibleDate(birthAnchor, birthPrecision, birthDisplay)
	h.DeathDate = buildFlexibleDate(deathAnchor, deathPrecision, deathDisplay)
	h.ServiceStartDate = buildFlexibleDate(serviceAnchor, servicePrecision, serviceDisplay)

	if len(membershipsJSON) > 0 {
		_ = json.Unmarshal(membershipsJSON, &h.Memberships)
	}

	return h, nil
}

// marshalMemberships сериализует memberships для записи в JSONB.
func marshalMemberships(m []string) []byte {
	if m == nil {
		m = []string{}
	}
	b, _ := json.Marshal(m)
	return b
}

// stringPtrValue безопасно разыменовывает указатель на строку.
func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

// nullableString возвращает nil для пустой строки,
// чтобы не хранить пустые display-тексты как пустые строки.
func nullableString(s string) any {
	if s == "" {
		return nil
	}

	return s
}

// flexibleDateParts разбирает FlexibleDate на три значения для SQL.
// Возвращает untyped nil для anchor когда он отсутствует — это критично
// для корректной работы pgx с SQL NULL (typed nil внутри interface{}
// может быть интерпретирован драйвером как не-NULL значение).
func flexibleDateParts(d *domain.FlexibleDate) (any, int, any) {
	if d == nil {
		return nil, int(domain.PrecisionUnknown), nil
	}
	// Defense-in-depth: защита от PrecisionUnspecified (0), которая нарушает
	// CHECK constraint в БД (precision BETWEEN 1 AND 7).
	if d.Precision == domain.PrecisionUnspecified {
		return nil, int(domain.PrecisionUnknown), nil
	}
	// Явно конвертируем typed nil в untyped nil для pgx.
	// Без этого (*time.Time)(nil) внутри interface{} может вызвать проблемы
	// при передаче в SQL-запрос как параметр.
	var anchor any
	if d.Anchor != nil {
		anchor = d.Anchor
	}
	return anchor, int(d.Precision), nullableString(d.DisplayText)
}

// buildFlexibleDate собирает FlexibleDate из значений БД.
func buildFlexibleDate(anchor *time.Time, precision int, display *string) domain.FlexibleDate {
	p := domain.DatePrecision(precision)

	// Защита от значений вне известного диапазона.
	if precision < int(domain.PrecisionExact) || precision > int(domain.PrecisionUnknown) {
		p = domain.PrecisionUnknown
	}

	// Если в БД вдруг оказался 0, считаем это неизвестной датой.
	if p == domain.PrecisionUnspecified {
		p = domain.PrecisionUnknown
	}

	// Защита от некорректных комбинаций.
	if p == domain.PrecisionExact && anchor == nil {
		p = domain.PrecisionUnknown
	}

	if (p == domain.PrecisionUnknown || p == domain.PrecisionDayMonth) && anchor != nil {
		anchor = nil
	}

	d := domain.FlexibleDate{
		Precision: p,
	}

	if anchor != nil {
		d.Anchor = anchor
	}

	if display != nil {
		d.DisplayText = *display
	}

	return d
}
