package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// PhotoUseCase бизнес-логика работы с фотографиями героев.
type PhotoUseCase interface {
	Add(ctx context.Context, p domain.AddPhotoParams) (string, error)
	BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error)
	Delete(ctx context.Context, heroID, photoID string) error
	DeleteBatch(ctx context.Context, heroID string, photoIDs []string) (int, error)
	ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error)
	Reorder(ctx context.Context, heroID string, photoIDs []string) error
	// SetMain назначает указанную фотографию главной для героя.
	SetMain(ctx context.Context, heroID, photoID string) error
	// Update обновляет метаданные фотографии.
	Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error)
}

// ThumbnailEnqueuer абстракция для постановки задач на генерацию превью.
// Позволяет тестировать PhotoUseCase без реального ThumbnailWorker.
type ThumbnailEnqueuer interface {
	Enqueue(task ThumbnailTask)
}

type photoUseCase struct {
	repo         repository.PhotoRepository
	heroRepo     repository.HeroRepository
	mediaStorage domain.MediaStorage
	worker       ThumbnailEnqueuer
	logger       *slog.Logger
}

// NewPhotoUseCase создаёт usecase с зависимостями для работы с БД и хранилищем.
func NewPhotoUseCase(
	repo repository.PhotoRepository,
	heroRepo repository.HeroRepository,
	mediaStorage domain.MediaStorage,
	worker ThumbnailEnqueuer, // ✅ интерфейс
	logger *slog.Logger,
) PhotoUseCase {
	return &photoUseCase{
		repo:         repo,
		heroRepo:     heroRepo,
		mediaStorage: mediaStorage,
		worker:       worker,
		logger:       logger,
	}
}

func (uc *photoUseCase) Add(ctx context.Context, p domain.AddPhotoParams) (string, error) {
	if p.URL == "" {
		return "", domain.ErrPhotoURLRequired
	}

	if _, err := uc.heroRepo.GetByID(ctx, p.HeroID); err != nil {
		return "", fmt.Errorf("hero %w", domain.ErrNotFound)
	}

	existing, err := uc.repo.ListByHero(ctx, p.HeroID)
	if err != nil {
		return "", fmt.Errorf("list existing photos: %w", err)
	}

	// Первое фото героя автоматически становится главным
	if len(existing) == 0 {
		p.IsMain = true
	}

	if p.SortOrder == 0 {
		p.SortOrder = len(existing)
	}

	id, err := uc.repo.Add(ctx, p)
	if err != nil {
		return "", err
	}

	// ставим задачу на генерацию thumbnail
	uc.worker.Enqueue(ThumbnailTask{
		PhotoID: id,
		HeroID:  p.HeroID,
		URL:     p.URL,
		FaceBox: p.FaceBox,
	})

	return id, nil
}

func (uc *photoUseCase) BatchAdd(ctx context.Context, heroID string, photos []domain.AddPhotoParams) ([]*domain.Photo, error) {
	if len(photos) == 0 {
		return nil, domain.ErrPhotosEmpty
	}

	if _, err := uc.heroRepo.GetByID(ctx, heroID); err != nil {
		return nil, fmt.Errorf("hero %w", domain.ErrNotFound)
	}

	// Вычисляем base для sort_order.
	// Это не критично для целостности данных, но даёт правильные порядковые номера.
	existing, err := uc.repo.ListByHero(ctx, heroID)
	if err != nil {
		return nil, fmt.Errorf("list existing photos: %w", err)
	}
	base := len(existing)

	for i := range photos {
		if photos[i].SortOrder == 0 {
			photos[i].SortOrder = base + i
		}
	}

	// Создаём маппинг URL -> FaceBox перед вызовом BatchAdd
	urlToFaceBox := make(map[string]*domain.FaceBox, len(photos))
	for _, p := range photos {
		urlToFaceBox[p.URL] = p.FaceBox
	}

	// Логика назначения главного фото теперь полностью в репозитории
	// (внутри транзакции, защита от race condition)
	added, err := uc.repo.BatchAdd(ctx, heroID, photos)
	if err != nil {
		return nil, err
	}

	// Используем маппинг для enqueue
	for _, p := range added {
		uc.worker.Enqueue(ThumbnailTask{
			PhotoID: p.ID,
			HeroID:  p.HeroID,
			URL:     p.URL,
			FaceBox: urlToFaceBox[p.URL],
		})
	}

	return added, nil
}

// Delete удаляет одно фото из БД и его файл из S3 (best-effort).
// Проверяет принадлежность фото герою через heroID.
func (uc *photoUseCase) Delete(ctx context.Context, heroID, photoID string) error {
	url, err := uc.repo.Delete(ctx, heroID, photoID)
	if err != nil {
		return err
	}

	uc.deleteFilesFromStorage(ctx, []string{url})
	return nil
}

// DeleteBatch удаляет отмеченные фото из БД и их файлы из S3.
// Возвращает количество удалённых из БД записей.
func (uc *photoUseCase) DeleteBatch(ctx context.Context, heroID string, photoIDs []string) (int, error) {
	if len(photoIDs) == 0 {
		return 0, domain.ErrPhotoIDsEmpty
	}

	// 1. Удаляем записи из БД (источник правды) и получаем URL файлов.
	urls, err := uc.repo.DeleteByIDs(ctx, heroID, photoIDs)
	if err != nil {
		return 0, err
	}

	// 2. Best-effort удаление файлов из S3.
	// Ошибки хранилища логируются, но не влияют на результат операции.
	uc.deleteFilesFromStorage(ctx, urls)

	return len(urls), nil
}

func (uc *photoUseCase) ListByHero(ctx context.Context, heroID string) ([]*domain.Photo, error) {
	return uc.repo.ListByHero(ctx, heroID)
}

func (uc *photoUseCase) Reorder(ctx context.Context, heroID string, photoIDs []string) error {
	if len(photoIDs) == 0 {
		return domain.ErrPhotoIDsEmpty
	}

	return uc.repo.Reorder(ctx, heroID, photoIDs)
}

// deleteFilesFromStorage параллельно удаляет объекты из S3 по публичным URL.
// Ошибки отдельных удалений логируются и не возвращаются: сбой хранилища
// не должен откатывать уже выполненное удаление записей из БД.
// Неудалённые файлы становятся сиротами и убираются задачей orphan cleanup.
func (uc *photoUseCase) deleteFilesFromStorage(ctx context.Context, urls []string) {
	if len(urls) == 0 {
		return
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // ограничиваем конкурентность, чтобы не перегружать S3

	for _, url := range urls {
		g.Go(func() error {
			key, err := uc.mediaStorage.KeyFromPublicURL(url)
			if err != nil {
				// URL не принадлежит нашему бакету (например, внешняя ссылка) — пропускаем.
				// Это защита от случайного удаления чужих объектов.
				uc.logger.WarnContext(gctx, "skip s3 delete: cannot extract key",
					"url", url, "error", err)
				return nil
			}

			if err := uc.mediaStorage.Delete(gctx, key); err != nil {
				uc.logger.WarnContext(gctx, "failed to delete s3 object",
					"key", key, "error", err)
				return nil
			}

			uc.logger.InfoContext(gctx, "s3 object deleted", "key", key)
			return nil
		})
	}

	// Все ошибки уже залогированы внутри горутин, возвращаемых ошибок нет.
	_ = g.Wait()
}

// SetMain назначает указанную фотографию главной для героя.
func (uc *photoUseCase) SetMain(ctx context.Context, heroID, photoID string) error {
	if _, err := uc.heroRepo.GetByID(ctx, heroID); err != nil {
		return fmt.Errorf("hero %w", domain.ErrNotFound)
	}

	return uc.repo.SetMain(ctx, heroID, photoID)
}

// Update обновляет метаданные фотографии (description, face_box).
// При изменении face_box ставит задачу на перегенерацию thumbnail.
func (uc *photoUseCase) Update(ctx context.Context, p domain.UpdatePhotoParams) (*domain.Photo, error) {
	if _, err := uc.heroRepo.GetByID(ctx, p.HeroID); err != nil {
		return nil, fmt.Errorf("hero %w", domain.ErrNotFound)
	}

	photo, err := uc.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	// Если изменился face_box, ставим задачу на перегенерацию thumbnail
	for _, f := range p.FieldMask {
		if f == "face_box" && uc.worker != nil {
			uc.worker.Enqueue(ThumbnailTask{
				PhotoID: photo.ID,
				HeroID:  photo.HeroID,
				URL:     photo.URL,
				FaceBox: photo.FaceBox,
			})
			break
		}
	}

	return photo, nil
}
