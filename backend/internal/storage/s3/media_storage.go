package s3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// Config параметры подключения к S3-совместимому хранилищу.
type Config struct {
	// Endpoint — внутренний адрес MinIO, доступный из docker-сети (minio_emh:9000).
	// Используется для реальных TCP-соединений и серверных операций.
	Endpoint string
	// PublicEndpoint — внешний адрес MinIO, доступный браузеру (localhost:4443).
	// Используется как host в presigned URL (подпись SigV4 привязана к host).
	PublicEndpoint string
	// PublicBaseURL — базовый URL для публичных ссылок на файлы.
	PublicBaseURL string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	// CACertPath — путь к CA-сертификату для самоподписанного TLS.
	CACertPath string
}

// MediaStorage реализует domain.MediaStorage на базе MinIO.
type MediaStorage struct {
	internal      *minio.Client // серверные операции (bucket, delete) — внутренний адрес
	public        *minio.Client // генерация presigned URL — внешний host, внутренний TCP
	bucket        string
	publicBaseURL string
}

// NewMediaStorage создаёт двух клиентов MinIO.
//
// Ключевой момент: presigned URL в SigV4 подписывается с учётом host, поэтому
// public-клиент сконфигурирован с ВНЕШНИМ endpoint (host для подписи), но его
// TCP-соединения через кастомный DialContext перенаправляются на ВНУТРЕННИЙ адрес,
// доступный из контейнера. Так host в URL остаётся доступным браузеру, а служебные
// запросы клиента (GetBucketLocation) доходят до MinIO.
func NewMediaStorage(cfg Config) (domain.MediaStorage, error) {
	// Internal-клиент: прямые соединения на внутренний адрес.
	internalTransport, err := buildTransport(cfg.CACertPath, "")
	if err != nil {
		return nil, fmt.Errorf("build internal transport: %w", err)
	}
	internal, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Transport: internalTransport,
	})
	if err != nil {
		return nil, fmt.Errorf("create internal minio client: %w", err)
	}

	// Public-клиент: endpoint = внешний адрес (host для подписи),
	// но TCP-соединения перенаправляются на внутренний адрес.
	publicTransport, err := buildTransport(cfg.CACertPath, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("build public transport: %w", err)
	}
	public, err := minio.New(cfg.PublicEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Transport: publicTransport,
	})
	if err != nil {
		return nil, fmt.Errorf("create public minio client: %w", err)
	}

	return &MediaStorage{
		internal:      internal,
		public:        public,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

// buildTransport создаёт HTTP-transport для MinIO-клиента.
//
// caCertPath — путь к CA-сертификату; если пуст, используется системное хранилище.
// redirectAddr — если не пуст, все TCP-соединения перенаправляются на этот адрес.
// Используется для public-клиента: host в HTTP-запросе и подписи остаётся внешним,
// но соединение устанавливается с внутренним адресом MinIO.
func buildTransport(caCertPath, redirectAddr string) (*http.Transport, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if redirectAddr != "" {
				addr = redirectAddr
			}
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:          256,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if caCertPath == "" {
		return transport, nil
	}

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert %q: %w", caCertPath, err)
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate from %q", caCertPath)
	}

	transport.TLSClientConfig = &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}
	return transport, nil
}

// EnsureBucket создаёт бакет при отсутствии и настраивает публичное чтение.
// Использует retry для устойчивости к кратковременным сбоям MinIO при старте.
func (s *MediaStorage) EnsureBucket(ctx context.Context) error {
	_, err := withRetry(ctx, defaultRetryConfig, "ensure_bucket", func() (struct{}, error) {
		exists, err := s.internal.BucketExists(ctx, s.bucket)
		if err != nil {
			return struct{}{}, fmt.Errorf("check bucket exists: %w", err)
		}

		if !exists {
			if err := s.internal.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
				return struct{}{}, fmt.Errorf("create bucket: %w", err)
			}
		}

		// Всегда настраиваем политику (идемпотентно).
		// Это решает проблему, когда бакет существует, но политика не настроена.
		if err := s.setPublicReadPolicy(ctx); err != nil {
			return struct{}{}, fmt.Errorf("set bucket policy: %w", err)
		}

		return struct{}{}, nil
	})

	return err
}

// setPublicReadPolicy разрешает анонимное чтение объектов (фото видны всем посетителям сайта).
func (s *MediaStorage) setPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/*"]
		}]
	}`, s.bucket)
	return s.internal.SetBucketPolicy(ctx, s.bucket, policy)
}

// PresignedPutURL генерирует подписанный URL для прямой загрузки файла.
//
// Content-Type включается в подпись (signed header): MinIO отклонит
// PUT с другим Content-Type, поэтому анонимный владелец URL не может
// загрузить контент произвольного типа.
func (s *MediaStorage) PresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	extraHeaders := http.Header{}
	if contentType != "" {
		extraHeaders.Set("Content-Type", contentType)
	}
	u, err := s.public.PresignHeader(ctx, http.MethodPut, s.bucket, key, expiry, nil, extraHeaders)
	if err != nil {
		return "", fmt.Errorf("presign put object: %w", err)
	}

	// ❗ КРИТИЧНО: AWS SigV4 не включает схему (http/https) в Canonical Request.
	// Безопасно меняем схему на HTTPS, чтобы браузер не блокировал Mixed Content.
	// Caddy терминирует TLS снаружи, а внутрь Docker-сети проксирует обычный HTTP.
	u.Scheme = "https"

	return u.String(), nil
}

// PublicURL формирует публичную ссылку на объект.
func (s *MediaStorage) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, s.bucket, key)
}

// Delete удаляет объект из хранилища по ключу.
func (s *MediaStorage) Delete(ctx context.Context, key string) error {
	_, err := withRetry(ctx, defaultRetryConfig, "delete", func() (struct{}, error) {
		err := s.internal.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})

		return struct{}{}, err
	})

	return err
}

// KeyFromPublicURL извлекает ключ объекта из публичного URL (для удаления по URL из БД).
func (s *MediaStorage) KeyFromPublicURL(rawURL string) (string, error) {
	prefix := fmt.Sprintf("%s/%s/", s.publicBaseURL, s.bucket)
	if !strings.HasPrefix(rawURL, prefix) {
		return "", fmt.Errorf("url does not belong to configured bucket")
	}
	return strings.TrimPrefix(rawURL, prefix), nil
}

// Download скачивает объект из хранилища.
// Использует retry для устойчивости к кратковременным сбоям MinIO.
// Внутри вызывается Stat() для проверки доступности объекта,
// так как GetObject в minio-go выполняет запрос лениво.
func (s *MediaStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := withRetry(ctx, defaultRetryConfig, "download", func() (*minio.Object, error) {
		o, err := s.internal.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}

		// GetObject ленивый — делаем Stat для проверки доступности.
		// Если MinIO недоступен, ошибка будет здесь, и retry сработает.
		if _, err := o.Stat(); err != nil {
			_ = o.Close()
			return nil, err
		}

		return o, nil
	})

	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	// Оборачиваем reader для подсчёта прочитанных байт
	return &countingReader{
		ReadCloser: obj,
		operation:  "download",
	}, nil
}

// Upload загружает объект в бакет с указанным Content-Type.
// Использует retry с exponential backoff для устойчивости к сбоям MinIO.
//
// Этот метод закрывает зависимость usecase от конкретной реализации S3:
// раньше ThumbnailWorker делал type assertion `w.storage.(*s3.MediaStorage)`
// для доступа к Internal().PutObject(), что нарушало DIP и делало
// невозможным тестирование с моками.
func (s *MediaStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	// Оборачиваем reader для подсчёта загруженных байт
	cr := &countingReader{ReadCloser: io.NopCloser(data), operation: "upload"}
	// Для retry нужен Seeker — сохраняем оригинал если возможно
	seeker, isSeeker := data.(io.ReadSeeker)

	_, err := withRetry(ctx, defaultRetryConfig, "upload", func() (struct{}, error) {
		if isSeeker {
			// Сбрасываем позицию на начало для повторных попыток
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return struct{}{}, fmt.Errorf("seek to start: %w", err)
			}

			cr.reset() // сбрасываем счётчик перед retry
		}
		_, err := s.internal.PutObject(
			ctx, s.bucket, key, cr, -1,
			minio.PutObjectOptions{ContentType: contentType},
		)

		return struct{}{}, err
	})

	return err
}

// ListObjects возвращает список всех объектов в бакете.
// Используется orphan cleanup воркером для поиска сиротских файлов.
// Использует retry для устойчивости к кратковременным сбоям MinIO.
func (s *MediaStorage) ListObjects(ctx context.Context) ([]domain.ObjectInfo, error) {
	result, err := withRetry(ctx, defaultRetryConfig, "list_objects", func() ([]domain.ObjectInfo, error) {
		var objects []domain.ObjectInfo

		opts := minio.ListObjectsOptions{
			Recursive: true,
		}

		for obj := range s.internal.ListObjects(ctx, s.bucket, opts) {
			if obj.Err != nil {
				return nil, fmt.Errorf("list objects iteration: %w", obj.Err)
			}
			objects = append(objects, domain.ObjectInfo{
				Key:          obj.Key,
				Size:         obj.Size,
				LastModified: obj.LastModified,
			})
		}

		return objects, nil
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	return result, nil
}

// ListObjectsPaged возвращает страницу объектов начиная с marker (exclusive).
// Использует minio.ListObjectsOptions.StartAfter для эффективной пагинации.
//
// Алгоритм:
//  1. StartAfter=marker — MinIO пропускает все ключи ≤ marker
//  2. Итерируем до limit объектов или конца списка
//  3. Возвращаем lastKey как next_marker для следующей страницы
//
// Пустой next_marker означает, что это последняя страница.
// Не использует withRetry: при сбое MinIO OrphanCleanupWorker повторит весь цикл.
func (s *MediaStorage) ListObjectsPaged(
	ctx context.Context,
	marker string,
	limit int,
) ([]domain.ObjectInfo, string, error) {
	if limit <= 0 {
		limit = 500
	}
	if limit > 1000 {
		limit = 1000 // S3 максимум
	}

	start := time.Now()
	opts := minio.ListObjectsOptions{
		Recursive:  true,
		StartAfter: marker,
	}

	objects := make([]domain.ObjectInfo, 0, limit)
	var lastKey string

	for obj := range s.internal.ListObjects(ctx, s.bucket, opts) {
		// Проверка отмены контекста для корректного graceful shutdown
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}

		if obj.Err != nil {
			metrics.S3OperationsTotal.WithLabelValues("list_objects_paged", "error").Inc()
			metrics.S3DurationSeconds.WithLabelValues("list_objects_paged").Observe(time.Since(start).Seconds())
			return nil, "", fmt.Errorf("list objects iteration: %w", obj.Err)
		}

		objects = append(objects, domain.ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})
		lastKey = obj.Key

		if len(objects) >= limit {
			break
		}
	}

	// Если получили ровно limit объектов — вероятно есть ещё
	// Если меньше — это последняя страница, lastKey = ""
	var nextMarker string
	if len(objects) == limit {
		nextMarker = lastKey
	}

	metrics.S3OperationsTotal.WithLabelValues("list_objects_paged", "success").Inc()
	metrics.S3DurationSeconds.WithLabelValues("list_objects_paged").Observe(time.Since(start).Seconds())

	return objects, nextMarker, nil
}

// Internal возвращает внутренний MinIO-клиент (для воркера).
func (s *MediaStorage) Internal() *minio.Client {
	return s.internal
}

// Bucket возвращает имя бакета.
func (s *MediaStorage) Bucket() string {
	return s.bucket
}

// retryConfig параметры повторных попыток для операций с MinIO.
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// defaultRetryConfig настройки по умолчанию:
// 3 повторные попытки, начальная задержка 100ms, максимум 2s.
var defaultRetryConfig = retryConfig{
	maxRetries: 3,
	baseDelay:  100 * time.Millisecond,
	maxDelay:   2 * time.Second,
}

// withRetry выполняет операцию с exponential backoff.
// Повторяет только при retryable-ошибках (сетевые сбои, 5xx, 429).
func withRetry[T any](ctx context.Context, cfg retryConfig, operation string, fn func() (T, error)) (T, error) {
	start := time.Now()
	var result T
	var lastErr error
	var retries int

	for attempt := 0; attempt <= cfg.maxRetries; attempt++ {
		if attempt > 0 {
			retries++
			// Exponential backoff: 100ms, 200ms, 400ms
			delay := cfg.baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > cfg.maxDelay {
				delay = cfg.maxDelay
			}

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, lastErr = fn()
		if lastErr == nil {
			// Успех
			metrics.S3OperationsTotal.WithLabelValues(operation, "success").Inc()
			metrics.S3DurationSeconds.WithLabelValues(operation).Observe(time.Since(start).Seconds())
			if retries > 0 {
				metrics.S3RetriesTotal.WithLabelValues(operation).Add(float64(retries))
			}

			return result, nil
		}

		if !isRetryableError(lastErr) {
			break
		}
	}

	// Исчерпаны попытки или не-retryable ошибка
	status := "error"
	if retries >= cfg.maxRetries {
		status = "retry_exhausted"
	}

	metrics.S3OperationsTotal.WithLabelValues(operation, status).Inc()
	metrics.S3DurationSeconds.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	if retries > 0 {
		metrics.S3RetriesTotal.WithLabelValues(operation).Add(float64(retries))
	}

	return result, fmt.Errorf("operation %s failed after %d retries: %w", operation, cfg.maxRetries, lastErr)
}

// isRetryableError определяет, стоит ли повторять операцию при данной ошибке.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Контекст отменён — не повторяем
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Проверяем MinIO ErrorResponse
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		// 4xx (кроме 429) не повторяем: 404, 403 и т.д.
		if minioErr.StatusCode >= 400 && minioErr.StatusCode < 500 && minioErr.StatusCode != 429 {
			return false
		}
		// 429 (rate limit) и 5xx повторяем
		return true
	}

	// Сетевые ошибки повторяем
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// По умолчанию повторяем: сетевые сбои могут быть обёрнуты
	return true
}

// countingReader оборачивает io.Reader для подсчёта переданных байт.
// При Close() инкрементирует метрику emh_s3_bytes_transferred_total.
type countingReader struct {
	io.ReadCloser
	operation string
	bytes     int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.bytes += int64(n)
	return n, err
}

func (r *countingReader) Close() error {
	if r.bytes > 0 {
		direction := "out" // для upload — байты уходят в S3
		if r.operation == "download" {
			direction = "in" // для download — байты приходят из S3
		}
		metrics.S3BytesTransferred.WithLabelValues(r.operation, direction).Add(float64(r.bytes))
	}
	if r.ReadCloser != nil {
		return r.ReadCloser.Close()
	}
	return nil
}

// reset сбрасывает счётчик байт (для retry при Seek).
func (r *countingReader) reset() {
	r.bytes = 0
}
