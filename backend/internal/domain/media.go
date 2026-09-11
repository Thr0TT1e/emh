package domain

import (
	"context"
	"io"
	"time"
)

// UploadType назначение загружаемого файла. Значения совпадают с proto enum.
type UploadType int

const (
	UploadTypeUnspecified UploadType = iota
	UploadTypeHeroPhoto
	UploadTypeAwardImage
	UploadTypeSubmissionAttachment
)

// UploadParams параметры запроса на получение presigned URL.
type UploadParams struct {
	Type        UploadType
	Filename    string
	ContentType string
}

// UploadResult результат генерации presigned URL.
type UploadResult struct {
	UploadURL string
	PublicURL string
	Key       string
	ExpiresAt time.Time
}

// FileToUpload описание файла для batch-загрузки.
type FileToUpload struct {
	Filename    string
	ContentType string
}

// UploadUrlInfo результат генерации presigned URL для одного файла.
type UploadUrlInfo struct {
	Filename  string
	UploadURL string
	PublicURL string
	ExpiresAt string
}

// ObjectInfo информация об объекте в S3-хранилище.
type ObjectInfo struct {
	// Key - путь объекта в бакете.
	Key string
	// Size - размер в байтах.
	Size int64
	// LastModified - время последнего изменения.
	LastModified time.Time
}

// MediaStorage абстракция объектного хранилища.
type MediaStorage interface {
	// PresignedPutURL генерирует подписанный PUT URL. Content-Type включается
	// в подпись: хранилище отклонит загрузку с другим Content-Type.
	PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	PublicURL(key string) string
	Delete(ctx context.Context, key string) error
	KeyFromPublicURL(rawURL string) (string, error)
	EnsureBucket(ctx context.Context) error
	// Download скачивает объект из хранилища.
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// ListObjects возвращает список всех объектов в бакете.
	// DEPRECATED: используйте ListObjectsPaged для больших бакетов.
	ListObjects(ctx context.Context) ([]ObjectInfo, error)

	// ListObjectsPaged возвращает страницу объектов начиная с marker.
	// marker — ключ объекта, с которого начинается страница (exclusive).
	// Пустой marker — начало списка.
	// Возвращает: объекты страницы, next_marker (пустой если последняя страница), ошибку.
	// Используется OrphanCleanupWorker для streaming-обработки без загрузки всего S3 в память.
	ListObjectsPaged(ctx context.Context, marker string, limit int) ([]ObjectInfo, string, error)

	// Upload загружает данные в хранилище с указанным Content-Type.
	// Используется ThumbnailWorker для загрузки превью и сконвертированных оригиналов.
	// Реализация должна включать retry при временных сбоях.
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error
}
