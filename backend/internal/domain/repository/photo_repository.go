package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// PhotoRepository абстракция для работы с фотографиями.
type PhotoRepository interface {
	Add(ctx context.Context, p domain.AddPhotoParams) (string, error)
	BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error)
	// Delete удаляет фото и возвращает URL файла для последующего удаления из S3.
	Delete(ctx context.Context, heroID, photoID string) (string, error)
	// DeleteByIDs удаляет фото героя пакетно и возвращает URL удалённых файлов.
	DeleteByIDs(ctx context.Context, heroID string, photoIDs []string) ([]string, error)
	ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error)
	// ListByHeroPaged возвращает фото героя с курсорной пагинацией.
	// Курсор кодирует пару (sort_order, id). Пустой курсор — первая страница.
	// Возвращает список, next_cursor (пустой если последняя страница), total.
	ListByHeroPaged(ctx context.Context, heroID string, cursor string, limit int) ([]*domain.Photo, string, int64, error)
	Reorder(ctx context.Context, heroID string, photoIDs []string) error
	GetMainPhotoURL(ctx context.Context, heroID string) (string, error)
	// SetMain назначает фотографию главной для героя.
	SetMain(ctx context.Context, heroID, photoID string) error
	// Update обновляет метаданные фотографии (description, face_box).
	Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error)
	// UpdateAfterProcessing обновляет URL и thumbnail_url после обработки воркером.
	UpdateAfterProcessing(ctx context.Context, photoID, url, thumbnailURL string) error
	// ListWithoutThumbnails возвращает фото без thumbnail для startup backfill.
	ListWithoutThumbnails(ctx context.Context) ([]*domain.Photo, error)
	// ListAllMediaURLs возвращает все URL медиафайлов для orphan cleanup.
	ListAllMediaURLs(ctx context.Context) ([]string, error)
	// ListWithoutThumbnailsPaged возвращает фото без thumbnail с курсорной пагинацией.
	// Используется StartupBackfill для batched-обработки без перегрузки очереди.
	// Курсор — id последнего обработанного фото. Пустой курсор — начало списка.
	// Возвращает: фото, next_cursor (пустой если последняя страница).
	ListWithoutThumbnailsPaged(ctx context.Context, cursor string, limit int) ([]*domain.Photo, string, error)
}
