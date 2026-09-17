package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// MediaUseCase бизнес-логика работы с медиа-контентом.
type MediaUseCase interface {
	GetUploadURL(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error)
	// Batch-генерация presigned URL
	BatchGetUploadURLs(ctx context.Context, fileType domain.UploadType, files []domain.FileToUpload) ([]domain.UploadUrlInfo, error)
}

type mediaUseCase struct {
	storage domain.MediaStorage
	expiry  time.Duration
}

// NewMediaUseCase создаёт usecase с настраиваемым временем жизни presigned URL.
func NewMediaUseCase(storage domain.MediaStorage, presignExpiry time.Duration) MediaUseCase {
	if presignExpiry <= 0 {
		presignExpiry = 15 * time.Minute
	}
	return &mediaUseCase{
		storage: storage,
		expiry:  presignExpiry,
	}
}

// allowedMIME — whitelist допустимых MIME-типов с маппингом на расширения.
// Жёсткая валидация типов критична для безопасности (защита от загрузки исполняемых файлов).
var allowedMIME = map[domain.UploadType]map[string]string{
	domain.UploadTypeHeroPhoto: {
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	},
	domain.UploadTypeAwardImage: {
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	},
	domain.UploadTypeAwardRibbon: {
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	},
	domain.UploadTypeSubmissionAttachment: {
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/webp":      ".webp",
		"application/pdf": ".pdf",
	},
}

// typePrefix — префиксы путей в бакете для группировки объектов.
var typePrefix = map[domain.UploadType]string{
	domain.UploadTypeHeroPhoto:            "heroes",
	domain.UploadTypeAwardImage:           "awards",
	domain.UploadTypeAwardRibbon:          "awards/ribbons",
	domain.UploadTypeSubmissionAttachment: "submissions",
}

// generateKey формирует ключ объекта в S3 для заданного типа файла и content-type.
// Вынесенная общая логика для GetUploadURL и BatchGetUploadURLs.
func (uc *mediaUseCase) generateKey(fileType domain.UploadType, contentType string) (string, error) {
	exts, ok := allowedMIME[fileType]
	if !ok {
		return "", fmt.Errorf("unsupported upload type: %d", domain.ErrUnsupportedUploadType)
	}
	ext, ok := exts[contentType]
	if !ok {
		return "", fmt.Errorf("content type %q not allowed for this upload type", domain.ErrContentTypeNotAllowed)
	}
	// Ключ вида: heroes/2026/07/30/<uuid>.jpg
	// Группировка по датам упрощает обслуживание и избегает коллизий.
	now := time.Now().UTC()
	return fmt.Sprintf("%s/%d/%02d/%02d/%s%s",
		typePrefix[fileType],
		now.Year(), int(now.Month()), now.Day(),
		uuid.NewString(), ext,
	), nil
}

func (uc *mediaUseCase) GetUploadURL(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error) {
	exts, ok := allowedMIME[p.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported upload type: %d", domain.ErrUnsupportedUploadType)
	}

	ext, ok := exts[p.ContentType]
	if !ok {
		return nil, fmt.Errorf("content type %q not allowed for this upload type", domain.ErrContentTypeNotAllowed)
	}

	// Ключ вида: heroes/2026/07/30/<uuid>.jpg
	// Группировка по датам упрощает обслуживание и избегает коллизий.
	now := time.Now().UTC()
	key := fmt.Sprintf("%s/%d/%02d/%02d/%s%s",
		typePrefix[p.Type],
		now.Year(), int(now.Month()), now.Day(),
		uuid.NewString(), ext,
	)

	uploadURL, err := uc.storage.PresignedPutURL(ctx, key, p.ContentType, uc.expiry)
	if err != nil {
		return nil, fmt.Errorf("generate presigned url: %w", err)
	}

	return &domain.UploadResult{
		UploadURL: uploadURL,
		PublicURL: uc.storage.PublicURL(key),
		Key:       key,
		ExpiresAt: now.Add(uc.expiry),
	}, nil
}

// BatchGetUploadURLs генерирует несколько presigned URL за один запрос.
func (uc *mediaUseCase) BatchGetUploadURLs(ctx context.Context, fileType domain.UploadType, files []domain.FileToUpload) ([]domain.UploadUrlInfo, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("files must not be empty")
	}
	if len(files) > 50 {
		return nil, fmt.Errorf("too many files (max 50)")
	}

	results := make([]domain.UploadUrlInfo, 0, len(files))
	for _, file := range files {
		key, err := uc.generateKey(fileType, file.ContentType)
		if err != nil {
			return nil, fmt.Errorf("invalid file %s: %w", file.Filename, err)
		}

		uploadURL, err := uc.storage.PresignedPutURL(ctx, key, file.ContentType, uc.expiry)
		if err != nil {
			return nil, fmt.Errorf("presign url for %s: %w", file.Filename, err)
		}

		results = append(results, domain.UploadUrlInfo{
			Filename:  file.Filename,
			UploadURL: uploadURL,
			PublicURL: uc.storage.PublicURL(key),
			ExpiresAt: time.Now().Add(uc.expiry).UTC().Format(time.RFC3339),
		})
	}
	return results, nil
}
