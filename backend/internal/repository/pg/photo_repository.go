package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

type photoRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewPhotoRepository(pool *pgxpool.Pool) repository.PhotoRepository {
	return &photoRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *photoRepository) Add(ctx context.Context, p domain.AddPhotoParams) (string, error) {
	query := r.sb.Insert("photos").
		Columns("hero_id", "url", "thumbnail_url", "description", "sort_order", "is_main", "face_box").
		Values(p.HeroID, p.URL, p.ThumbnailURL, p.Description, p.SortOrder, p.IsMain,
			squirrel.Expr("?::jsonb", marshalFaceBox(p.FaceBox))).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build add photo: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		// Нарушение FK означает, что герой не существует
		return "", fmt.Errorf("exec add photo: %w", err)
	}
	return id, nil
}

// Delete удаляет фотографию и возвращает URL файла для удаления из S3.
// Проверяет принадлежность фото герою через heroID (защита от IDOR).
func (r *photoRepository) Delete(ctx context.Context, heroID, photoID string) (string, error) {
	query := r.sb.Delete("photos").
		Where(squirrel.Eq{"id": photoID, "hero_id": heroID}).
		Suffix("RETURNING url")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build delete photo: %w", err)
	}

	var url string
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("photo %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("exec delete photo: %w", err)
	}

	return url, nil
}

func (r *photoRepository) ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	query := r.sb.Select(
		"id", "hero_id", "url", "thumbnail_url", "description",
		"sort_order", "is_main", "face_box", "created_at", "updated_at",
	).
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID}).
		OrderBy("sort_order ASC, id ASC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list photos: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list photos: %w", err)
	}
	defer rows.Close()

	var photos []*domain.Photo
	for rows.Next() {
		p := &domain.Photo{}
		var faceBoxJSON []byte
		if err := rows.Scan(
			&p.ID, &p.HeroID, &p.URL, &p.ThumbnailURL, &p.Description,
			&p.SortOrder, &p.IsMain, &faceBoxJSON, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		if len(faceBoxJSON) > 0 {
			_ = json.Unmarshal(faceBoxJSON, &p.FaceBox)
		}
		photos = append(photos, p)
	}
	return photos, rows.Err()
}

// ListByHeroPaged возвращает фото героя с курсорной пагинацией.
// Курсор кодирует пару (sort_order, id) в формате "{sort_order}:{id}".
// Использует покрывающий индекс idx_photos_hero_sort.
func (r *photoRepository) ListByHeroPaged(
	ctx context.Context,
	heroID string,
	cursor string,
	limit int,
) ([]*domain.Photo, string, int64, error) {
	sel := r.sb.Select(
		"id", "hero_id", "url", "thumbnail_url", "description",
		"sort_order", "is_main", "face_box", "created_at", "updated_at",
	).
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID})

	// Декодируем курсор: "{sort_order}:{id}"
	if cursor != "" {
		cursorSortOrder, cursorID, err := decodePhotoCursor(cursor)
		if err != nil {
			return nil, "", 0, fmt.Errorf("invalid cursor: %w", err)
		}
		// (sort_order, id) > (cursor_sort_order, cursor_id)
		sel = sel.Where(squirrel.Or{
			squirrel.Expr("sort_order > ?", cursorSortOrder),
			squirrel.And{
				squirrel.Eq{"sort_order": cursorSortOrder},
				squirrel.Expr("id > ?", cursorID),
			},
		})
	}

	sel = sel.OrderBy("sort_order ASC, id ASC").Limit(uint64(limit + 1))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list photos paged: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("exec list photos paged: %w", err)
	}
	defer rows.Close()

	photos := make([]*domain.Photo, 0, limit)
	for rows.Next() {
		p := &domain.Photo{}
		var faceBoxJSON []byte
		if err := rows.Scan(
			&p.ID, &p.HeroID, &p.URL, &p.ThumbnailURL, &p.Description,
			&p.SortOrder, &p.IsMain, &faceBoxJSON, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, "", 0, fmt.Errorf("scan photo: %w", err)
		}
		if len(faceBoxJSON) > 0 {
			_ = json.Unmarshal(faceBoxJSON, &p.FaceBox)
		}
		photos = append(photos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration: %w", err)
	}

	// Определяем наличие следующей страницы
	var nextCursor string
	if len(photos) > limit {
		photos = photos[:limit]
		last := photos[len(photos)-1]
		nextCursor = encodePhotoCursor(last.SortOrder, last.ID)
	}

	// Total считаем только для первой страницы
	var total int64
	if cursor == "" {
		countQuery := r.sb.Select("COUNT(*)").
			From("photos").
			Where(squirrel.Eq{"hero_id": heroID})
		countSQL, countArgs, err := countQuery.ToSql()
		if err != nil {
			return nil, "", 0, fmt.Errorf("build count photos: %w", err)
		}
		if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
			return nil, "", 0, fmt.Errorf("exec count photos: %w", err)
		}
	}

	return photos, nextCursor, total, nil
}

// Reorder атомарно обновляет sort_order согласно переданному порядку photo_ids.
// Использует транзакцию, чтобы избежать частичного обновления при ошибке.
func (r *photoRepository) Reorder(ctx context.Context, heroID string, photoIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Проверяем, что все переданные фото принадлежат данному герою.
	// Защита от подмены ID чужих фото.
	checkQuery := r.sb.Select("COUNT(*)").
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID, "id": photoIDs})

	checkSQL, checkArgs, err := checkQuery.ToSql()
	if err != nil {
		return fmt.Errorf("build check query: %w", err)
	}

	var count int
	if err := tx.QueryRow(ctx, checkSQL, checkArgs...).Scan(&count); err != nil {
		return fmt.Errorf("exec check query: %w", err)
	}
	if count != len(photoIDs) {
		return fmt.Errorf("%w", domain.ErrPhotosNotBelongToHero)
	}

	// Обновляем sort_order для каждого фото
	for i, photoID := range photoIDs {
		update := r.sb.Update("photos").
			Set("sort_order", i).
			Where(squirrel.Eq{"id": photoID, "hero_id": heroID})

		sql, args, err := update.ToSql()
		if err != nil {
			return fmt.Errorf("build reorder update: %w", err)
		}
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			return fmt.Errorf("exec reorder update: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *photoRepository) GetMainPhotoURL(ctx context.Context, heroID string) (string, error) {
	query := r.sb.Select("url").
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID, "is_main": true}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build get main photo: %w", err)
	}

	var url string
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Нет главного фото — не ошибка
		}
		return "", fmt.Errorf("exec get main photo: %w", err)
	}
	return url, nil
}

// BatchAdd атомарно добавляет несколько фотографий в транзакции.
// Проверка наличия главного фото выполняется внутри транзакции,
// что исключает race condition при параллельных запросах.
func (r *photoRepository) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Проверяем, есть ли уже главное фото у героя (внутри транзакции).
	// Это защита от race condition: между проверкой и вставкой
	// другой запрос не может добавить главное фото.
	checkMainQuery := r.sb.Select("COUNT(*)").
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID, "is_main": true})
	checkMainSQL, checkMainArgs, err := checkMainQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build check main query: %w", err)
	}
	var mainCount int
	if err := tx.QueryRow(ctx, checkMainSQL, checkMainArgs...).Scan(&mainCount); err != nil {
		return nil, fmt.Errorf("exec check main query: %w", err)
	}
	hasMainInDB := mainCount > 0

	// Определяем индекс фото, которое станет главным.
	// Если в БД уже есть главное — ни одно фото из пакета не становится главным.
	mainIndex := -1
	if !hasMainInDB {
		for i, p := range photos {
			if p.IsMain {
				mainIndex = i
				break
			}
		}
		if mainIndex == -1 && len(photos) > 0 {
			mainIndex = 0
		}
	}

	result := make([]*domain.Photo, 0, len(photos))
	for i, p := range photos {
		isMain := i == mainIndex

		query := r.sb.Insert("photos").
			Columns("hero_id", "url", "thumbnail_url", "description", "sort_order", "is_main", "face_box").
			Values(heroID, p.URL, p.ThumbnailURL, p.Description, p.SortOrder, isMain,
				squirrel.Expr("?::jsonb", marshalFaceBox(p.FaceBox))).
			Suffix("RETURNING id, url, sort_order, is_main")
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, fmt.Errorf("build batch add photo: %w", err)
		}
		photo := &domain.Photo{HeroID: heroID, Description: p.Description}
		if err := tx.QueryRow(ctx, sql, args...).Scan(
			&photo.ID, &photo.URL, &photo.SortOrder, &photo.IsMain,
		); err != nil {
			return nil, fmt.Errorf("exec batch add photo: %w", err)
		}
		result = append(result, photo)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil
}

// DeleteByIDs удаляет фотографии, принадлежащие герою, и возвращает их URL
// для последующего удаления объектов из S3.
// Атомарная операция: если хотя бы одно фото не принадлежит герою,
// транзакция откатывается и возвращается ошибка.
func (r *photoRepository) DeleteByIDs(ctx context.Context, heroID string, photoIDs []string) ([]string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Проверяем, что все переданные фото принадлежат данному герою.
	// Защита от подмены ID чужих фото.
	checkQuery := r.sb.Select("COUNT(*)").
		From("photos").
		Where(squirrel.Eq{"hero_id": heroID, "id": photoIDs})
	checkSQL, checkArgs, err := checkQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build check query: %w", err)
	}
	var count int
	if err := tx.QueryRow(ctx, checkSQL, checkArgs...).Scan(&count); err != nil {
		return nil, fmt.Errorf("exec check query: %w", err)
	}
	if count != len(photoIDs) {
		return nil, fmt.Errorf("%w", domain.ErrPhotosNotBelongToHero)
	}

	// Удаляем фото
	deleteQuery := r.sb.Delete("photos").
		Where(squirrel.Eq{"hero_id": heroID, "id": photoIDs}).
		Suffix("RETURNING url")
	deleteSQL, deleteArgs, err := deleteQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build delete query: %w", err)
	}

	rows, err := tx.Query(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return nil, fmt.Errorf("exec delete photos: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, fmt.Errorf("scan deleted photo url: %w", err)
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return urls, nil
}

// SetMain назначает фотографию главной для героя.
// Триггер ensure_single_main_photo автоматически сбросит флаг у остальных фото.
func (r *photoRepository) SetMain(ctx context.Context, heroID, photoID string) error {
	update := r.sb.Update("photos").
		Set("is_main", true).
		Where(squirrel.Eq{"id": photoID, "hero_id": heroID})

	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build set main photo: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec set main photo: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("photo %w", domain.ErrNotFound)
	}
	return nil
}

// Update обновляет метаданные фотографии (description, face_box).
func (r *photoRepository) Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
	update := r.sb.Update("photos").
		Where(squirrel.Eq{"id": p.PhotoID, "hero_id": p.HeroID})

	mask := make(map[string]bool, len(p.FieldMask))
	for _, f := range p.FieldMask {
		mask[f] = true
	}

	if mask["description"] {
		update = update.Set("description", p.Description)
	}
	if mask["face_box"] {
		update = update.Set("face_box", squirrel.Expr("?::jsonb", marshalFaceBox(p.FaceBox)))
	}

	update = update.Set("updated_at", time.Now()).
		Suffix("RETURNING id, hero_id, url, thumbnail_url, description, sort_order, is_main, face_box, created_at, updated_at")

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update photo: %w", err)
	}

	photo := &domain.Photo{}
	var faceBoxJSON []byte
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(
		&photo.ID, &photo.HeroID, &photo.URL, &photo.ThumbnailURL, &photo.Description,
		&photo.SortOrder, &photo.IsMain, &faceBoxJSON, &photo.CreatedAt, &photo.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("photo %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("exec update photo: %w", err)
	}

	if len(faceBoxJSON) > 0 {
		_ = json.Unmarshal(faceBoxJSON, &photo.FaceBox)
	}

	return photo, nil
}

// UpdateAfterProcessing обновляет URL и thumbnail_url после обработки воркером.
func (r *photoRepository) UpdateAfterProcessing(ctx context.Context, photoID, url, thumbnailURL string) error {
	update := r.sb.Update("photos").
		Set("url", url).
		Set("thumbnail_url", thumbnailURL).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": photoID})
	sql, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("build update after processing: %w", err)
	}
	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec update after processing: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("photo %w", domain.ErrNotFound)
	}
	return nil
}

// ListWithoutThumbnails возвращает фото без thumbnail для startup backfill.
// Использует COALESCE для эффективной проверки пустых и NULL значений.
func (r *photoRepository) ListWithoutThumbnails(ctx context.Context) ([]*domain.Photo, error) {
	query := r.sb.Select("id", "hero_id", "url", "face_box").
		From("photos").
		Where(squirrel.Expr("COALESCE(thumbnail_url, '') = ''"))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list without thumbnails: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list without thumbnails: %w", err)
	}
	defer rows.Close()

	var photos []*domain.Photo
	for rows.Next() {
		p := &domain.Photo{}
		var faceBoxJSON []byte
		if err := rows.Scan(&p.ID, &p.HeroID, &p.URL, &faceBoxJSON); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		if len(faceBoxJSON) > 0 {
			_ = json.Unmarshal(faceBoxJSON, &p.FaceBox)
		}
		photos = append(photos, p)
	}

	return photos, rows.Err()
}

// ListWithoutThumbnailsPaged возвращает фото без thumbnail с курсорной пагинацией.
// Использует покрывающий частичный индекс idx_photos_no_thumbnail для быстрого поиска.
//
// Курсор — id последнего обработанного фото (UUIDv7, монотонный).
// Сортировка по id ASC для стабильной пагинации.
func (r *photoRepository) ListWithoutThumbnailsPaged(
	ctx context.Context,
	cursor string,
	limit int,
) ([]*domain.Photo, string, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	query := r.sb.Select("id", "hero_id", "url", "face_box").
		From("photos").
		Where(squirrel.Expr("COALESCE(thumbnail_url, '') = ''"))

	// Курсорная пагинация: id > cursor (UUIDv7 монотонный)
	if cursor != "" {
		query = query.Where(squirrel.Gt{"id": cursor})
	}

	query = query.OrderBy("id ASC").Limit(uint64(limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, "", fmt.Errorf("build list without thumbnails paged: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", fmt.Errorf("exec list without thumbnails paged: %w", err)
	}
	defer rows.Close()

	var photos []*domain.Photo
	for rows.Next() {
		p := &domain.Photo{}
		var faceBoxJSON []byte
		if err := rows.Scan(&p.ID, &p.HeroID, &p.URL, &faceBoxJSON); err != nil {
			return nil, "", fmt.Errorf("scan photo: %w", err)
		}
		if len(faceBoxJSON) > 0 {
			_ = json.Unmarshal(faceBoxJSON, &p.FaceBox)
		}
		photos = append(photos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("rows iteration: %w", err)
	}

	// Определяем next_cursor
	var nextCursor string
	if len(photos) == limit {
		nextCursor = photos[len(photos)-1].ID
	}

	return photos, nextCursor, nil
}

// marshalFaceBox сериализует FaceBox для записи в JSONB.
func marshalFaceBox(fb *domain.FaceBox) []byte {
	if fb == nil {
		return nil
	}
	b, _ := json.Marshal(fb)
	return b
}

// ListAllMediaURLs возвращает все URL медиафайлов из таблицы photos.
// Включает как оригиналы (url), так и превью (thumbnail_url).
// Используется orphan cleanup воркером для определения известных файлов.
func (r *photoRepository) ListAllMediaURLs(ctx context.Context) ([]string, error) {
	query := r.sb.Select("url", "thumbnail_url").From("photos")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list media urls: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list media urls: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url, thumbnailURL *string
		if err := rows.Scan(&url, &thumbnailURL); err != nil {
			return nil, fmt.Errorf("scan media url: %w", err)
		}
		if url != nil && *url != "" {
			urls = append(urls, *url)
		}
		if thumbnailURL != nil && *thumbnailURL != "" {
			urls = append(urls, *thumbnailURL)
		}
	}

	return urls, rows.Err()
}

// encodePhotoCursor кодирует пару (sort_order, id) в строку курсора.
func encodePhotoCursor(sortOrder int, id string) string {
	return fmt.Sprintf("%d:%s", sortOrder, id)
}

// decodePhotoCursor декодирует строку курсора в (sort_order, id).
func decodePhotoCursor(cursor string) (int, string, error) {
	idx := strings.Index(cursor, ":")
	if idx < 0 {
		return 0, "", fmt.Errorf("malformed cursor format")
	}
	sortOrder, err := strconv.Atoi(cursor[:idx])
	if err != nil {
		return 0, "", fmt.Errorf("malformed sort_order in cursor: %w", err)
	}
	id := cursor[idx+1:]
	if id == "" {
		return 0, "", fmt.Errorf("empty id in cursor")
	}
	return sortOrder, id, nil
}
