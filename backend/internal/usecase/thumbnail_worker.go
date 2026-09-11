package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/deepteams/webp"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// Параметры batched backfill.
const (
	// backfillBatchSize — количество фото за один запрос к БД.
	backfillBatchSize = 100
	// backfillQueueThreshold — порог заполнения очереди (в %).
	// Когда очередь заполнена более чем на этот процент, backfill приостанавливается
	// до освобождения места. 50% даёт запас для пиковых Enqueue от API.
	backfillQueueThreshold = 0.5
	// backfillCheckInterval — интервал проверки заполнения очереди.
	backfillCheckInterval = 500 * time.Millisecond
)

// Защита от decompression bomb: злоумышленник может загрузить через
// публичный presigned URL маленький по размеру, но огромный по габаритам
// файл (например, 100000x100000 пикселей), при декодировании которого
// воркер исчерпает память.
const (
	maxDownloadBytes   = 50 * 1024 * 1024 // максимум 50 MB скачиваемого файла
	maxImageDimension  = 10000            // максимум пикселей по стороне
	maxImageMegapixels = 50               // максимум мегапикселей всего
)

// Параметры retry для Enqueue (exponential backoff).
const (
	enqueueMaxRetries = 3
	enqueueBaseDelay  = 100 * time.Millisecond
	enqueueMaxDelay   = 400 * time.Millisecond
)

// Параметры RescanLoop.
const (
	// Интервал периодического сканирования фото без thumbnail.
	// 5 минут — баланс между нагрузкой на БД и скоростью восстановления.
	rescanInterval = 5 * time.Minute
	// Максимум задач за один rescan цикл (защита от перегрузки очереди).
	rescanBatchSize = 100
)

// ThumbnailTask задача на генерацию thumbnail и конвертацию оригинала.
type ThumbnailTask struct {
	PhotoID string
	HeroID  string
	URL     string
	FaceBox *domain.FaceBox
}

// ThumbnailWorker фоновый обработчик задач генерации превью.
type ThumbnailWorker struct {
	photoRepo repository.PhotoRepository
	storage   domain.MediaStorage
	logger    *slog.Logger
	cfg       config.ThumbnailConfig
	queue     chan ThumbnailTask
}

// NewThumbnailWorker создаёт воркер с очередью задач.
func NewThumbnailWorker(
	photoRepo repository.PhotoRepository,
	storage domain.MediaStorage,
	logger *slog.Logger,
	cfg config.ThumbnailConfig,
) *ThumbnailWorker {
	return &ThumbnailWorker{
		photoRepo: photoRepo,
		storage:   storage,
		logger:    logger,
		cfg:       cfg,
		queue:     make(chan ThumbnailTask, cfg.QueueSize),
	}
}

// Enqueue добавляет задачу в очередь с retry при переполнении.
//
// Exponential backoff (100ms → 200ms → 400ms) позволяет переждать
// кратковременные пики нагрузки без блокировки API-запросов навечно.
//
// Если все retry исчерпаны — задача не ставится в очередь, но будет
// подхвачена периодическим RescanLoop (раз в 5 минут).
//
// Worst-case задержка для API: ~700ms (3 retry × ~230ms средняя).
func (w *ThumbnailWorker) Enqueue(task ThumbnailTask) {
	for attempt := 0; attempt <= enqueueMaxRetries; attempt++ {
		select {
		case w.queue <- task:
			w.logger.Debug("thumbnail task enqueued",
				"photo_id", task.PhotoID,
				"attempt", attempt,
			)
			return
		default:
			// Очередь переполнена — retry с backoff
			if attempt < enqueueMaxRetries {
				metrics.EnqueueRetriesTotal.Inc()
				delay := enqueueBaseDelay * time.Duration(1<<uint(attempt))
				if delay > enqueueMaxDelay {
					delay = enqueueMaxDelay
				}
				w.logger.Debug("thumbnail queue full, retrying",
					"photo_id", task.PhotoID,
					"attempt", attempt+1,
					"delay_ms", delay.Milliseconds(),
				)
				time.Sleep(delay)
			}
		}
	}

	// Все retry исчерпаны — задача будет восстановлена rescan
	metrics.EnqueueFailuresTotal.Inc()
	w.logger.Warn("thumbnail enqueue failed after retries, will be recovered by rescan",
		"photo_id", task.PhotoID,
		"max_retries", enqueueMaxRetries,
	)
}

// Run запускает N воркеров в фоне. Блокирует до отмены контекста.
func (w *ThumbnailWorker) Run(ctx context.Context) {
	for i := 0; i < w.cfg.Workers; i++ {
		go w.worker(ctx, i)
	}
	<-ctx.Done()
	w.logger.Info("thumbnail workers stopped")
}

// worker обрабатывает задачи из очереди.
func (w *ThumbnailWorker) worker(ctx context.Context, id int) {
	w.logger.Info("запущен обработчик миниатюр", "worker_id", id)
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("обработчик миниатюр остановился.", "worker_id", id)
			return
		case task := <-w.queue:
			w.processTask(ctx, task)
		}
	}
}

// processTask обрабатывает одну задачу: скачивает, конвертирует, загружает.
// Перед каждой длительной операцией проверяется отмена контекста
// для корректного завершения при graceful shutdown.
// processTask обрабатывает одну задачу: скачивает, конвертирует, загружает.
// Перед каждой длительной операцией проверяется отмена контекста
// для корректного завершения при graceful shutdown.
func (w *ThumbnailWorker) processTask(ctx context.Context, task ThumbnailTask) {
	totalStart := time.Now()
	logger := w.logger.With("photo_id", task.PhotoID, "hero_id", task.HeroID)

	// Финальная метрика — defer гарантирует учёт даже при early return
	var status string = "error"
	defer func() {
		metrics.ThumbnailProcessedTotal.WithLabelValues(status).Inc()
		metrics.ThumbnailProcessingDuration.WithLabelValues("total").Observe(time.Since(totalStart).Seconds())
		metrics.ThumbnailQueueSize.Dec()
	}()

	// Проверка отмены
	if err := ctx.Err(); err != nil {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()
		logger.Debug("task cancelled before processing", "error", err)

		return
	}

	// 1. Extract key
	key, err := w.storage.KeyFromPublicURL(task.URL)
	if err != nil {
		logger.Error("cannot extract key from URL", "error", err)

		return
	}

	// 2. Download
	if err := ctx.Err(); err != nil {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()

		return
	}

	downloadStart := time.Now()
	reader, err := w.storage.Download(ctx, key)
	if err != nil {
		metrics.ThumbnailSkippedTotal.WithLabelValues("download_error").Inc()
		logger.Error("failed to download original", "error", err)

		return
	}
	defer reader.Close()

	// Читаем в буфер с жёстким лимитом байт.
	// maxDownloadBytes+1 позволяет отличить "файл больше лимита" от "ровно лимит".
	var buf bytes.Buffer
	n, err := io.Copy(&buf, io.LimitReader(reader, maxDownloadBytes+1))
	if err != nil {
		logger.Error("failed to read image", "error", err)

		return
	}

	metrics.ThumbnailProcessingDuration.WithLabelValues("download").Observe(time.Since(downloadStart).Seconds())

	if n > maxDownloadBytes {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("oversized").Inc()
		logger.Error("image exceeds max download size, skipping", "bytes", n)

		return
	}

	// 3. Decode config
	if err := ctx.Err(); err != nil {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()

		return
	}

	decodeStart := time.Now()
	imgConfig, format, err := image.DecodeConfig(bytes.NewReader(buf.Bytes()))
	if err != nil {
		logger.Error("failed to decode image config", "error", err)

		return
	}
	if imgConfig.Width > maxImageDimension || imgConfig.Height > maxImageDimension {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("dimensions").Inc()
		logger.Error("image dimensions too large, skipping")

		return
	}
	if imgConfig.Width*imgConfig.Height > maxImageMegapixels*1_000_000 {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("megapixels").Inc()
		logger.Error("image megapixels too large, skipping")

		return
	}

	// 3b. Теперь безопасно декодируем полный пиксельный контент.
	img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		logger.Error("failed to decode image", "error", err)

		return
	}
	metrics.ThumbnailProcessingDuration.WithLabelValues("decode").Observe(time.Since(decodeStart).Seconds())

	// 4. Crop
	if err := ctx.Err(); err != nil {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()

		return
	}
	cropStart := time.Now()
	thumbnail := w.cropAndResize(img, task.FaceBox)
	metrics.ThumbnailProcessingDuration.WithLabelValues("crop").Observe(time.Since(cropStart).Seconds())

	// 5. Upload thumbnail
	uploadStart := time.Now()
	thumbnailKey := fmt.Sprintf("thumbnails/%s.webp", task.PhotoID)
	if err := w.uploadWebP(ctx, thumbnailKey, thumbnail, float32(w.cfg.Quality)); err != nil {
		logger.Error("failed to upload thumbnail", "error", err)

		return
	}
	thumbnailURL := w.storage.PublicURL(thumbnailKey)
	metrics.ThumbnailProcessingDuration.WithLabelValues("upload_thumbnail").Observe(time.Since(uploadStart).Seconds())

	// 6. Конвертируем оригинал в WebP (если ещё не WebP)
	originalURL := task.URL
	if strings.ToLower(format) != "webp" {
		// Проверка перед конвертацией оригинала (дорогостоящая операция)
		if err := ctx.Err(); err != nil {
			status = "skipped"
			metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()

			return
		}

		origStart := time.Now()
		originalKey := fmt.Sprintf("photos/%s.webp", task.PhotoID)
		if err := w.uploadWebP(ctx, originalKey, img, float32(w.cfg.OriginalQuality)); err != nil {
			logger.Error("failed to upload converted original", "error", err)
			// Загрузка не удалась — используем оригинальный URL.
			// Это не критично: данные корректны, просто без оптимизации WebP.
		} else {
			originalURL = w.storage.PublicURL(originalKey)

			// Строгое удаление старого оригинала.
			// Если не удалось удалить — задача прерывается и будет повторена.
			// При повторе: конвертация выполнится заново, БД обновится,
			// удаление будет повторено. Файл будет удалён гарантированно.
			if err := w.deleteOriginalWithRetry(ctx, key); err != nil {
				metrics.ThumbnailOriginalDeleteFailuresTotal.Inc()
				logger.Error("failed to delete original file, task will be retried",
					"error", err,
					"original_key", key,
					"new_key", originalKey,
				)
				// Прерываем задачу. БД НЕ обновлена, данные не потеряны.
				// При повторе задачи (через runWorkerWithRecovery или повторный вызов)
				// удаление будет выполнено повторно.
				return
			}

			logger.Debug("original file deleted after conversion", "key", key)
		}

		metrics.ThumbnailProcessingDuration.WithLabelValues("upload_original").Observe(time.Since(origStart).Seconds())
	}

	// 7. DB update
	if err := ctx.Err(); err != nil {
		status = "skipped"
		metrics.ThumbnailSkippedTotal.WithLabelValues("cancelled").Inc()
		return
	}
	dbStart := time.Now()
	if err := w.photoRepo.UpdateAfterProcessing(ctx, task.PhotoID, originalURL, thumbnailURL); err != nil {
		logger.Error("failed to update photo in DB", "error", err)
		return
	}
	metrics.ThumbnailProcessingDuration.WithLabelValues("db_update").Observe(time.Since(dbStart).Seconds())

	// Успех!
	status = "success"
	logger.Info("thumbnail generated successfully",
		"thumbnail_url", thumbnailURL,
		"original_url", originalURL,
		"duration_ms", time.Since(totalStart).Milliseconds(),
	)
}

// cropAndResize обрезает и масштабирует изображение до 480×600 (4:5).
func (w *ThumbnailWorker) cropAndResize(img image.Image, faceBox *domain.FaceBox) image.Image {
	var cropped image.Image
	if faceBox != nil {
		// Crop по координатам face_box
		cropped = imaging.Crop(img, image.Rect(
			faceBox.X,
			faceBox.Y,
			faceBox.X+faceBox.Width,
			faceBox.Y+faceBox.Height,
		))
	} else {
		// Fallback: центральный crop 4:5
		bounds := img.Bounds()
		imgW, imgH := bounds.Dx(), bounds.Dy()
		targetRatio := float64(w.cfg.Width) / float64(w.cfg.Height)
		currentRatio := float64(imgW) / float64(imgH)
		if currentRatio > targetRatio {
			// Изображение шире — обрезаем по бокам
			newW := int(float64(imgH) * targetRatio)
			x := (imgW - newW) / 2
			cropped = imaging.Crop(img, image.Rect(x, 0, x+newW, imgH))
		} else {
			// Изображение выше — обрезаем сверху/снизу
			newH := int(float64(imgW) / targetRatio)
			y := (imgH - newH) / 2
			cropped = imaging.Crop(img, image.Rect(0, y, imgW, y+newH))
		}
	}

	// Resize до 480×600 (fill)
	return imaging.Fill(cropped, w.cfg.Width, w.cfg.Height, imaging.Center, imaging.Lanczos)
}

// uploadWebP кодирует изображение в WebP (lossy) и загружает в S3 через интерфейсный метод.
// quality — качество сжатия от 0 до 100 (чем выше, тем лучше качество и больше размер).
// Method=4 обеспечивает хороший баланс между скоростью и размером файла.
func (w *ThumbnailWorker) uploadWebP(ctx context.Context, key string, img image.Image, quality float32) error {
	var buf bytes.Buffer

	opts := &webp.EncoderOptions{
		Lossless: false,
		Quality:  quality,
		Method:   4,
	}

	if err := webp.Encode(&buf, img, opts); err != nil {
		return fmt.Errorf("encode webp: %w", err)
	}
	// Используем интерфейсный метод — без type assertion к *s3.MediaStorage.
	// bytes.NewReader возвращает *bytes.Reader, реализующий io.ReadSeeker,
	// поэтому retry внутри Upload работает корректно.
	if err := w.storage.Upload(ctx, key, bytes.NewReader(buf.Bytes()), "image/webp"); err != nil {
		return fmt.Errorf("upload object: %w", err)
	}

	return nil
}

// deleteOriginalWithRetry удаляет оригинальный файл после успешной конвертации в WebP.
//
// Использует 3 дополнительных попытки сверх внутренних retry в MediaStorage.Delete.
// Итого: до 3 × 4 = 12 попыток (внешние × внутренние с экспоненциальной задержкой).
//
// Если все попытки исчерпаны — возвращает ошибку для прерывания задачи.
// При повторе задачи файл будет удалён (идемпотентно).
//
// Возвращает:
//   - nil — файл удалён или уже не существует
//   - error — все попытки исчерпаны, файл НЕ удалён
func (w *ThumbnailWorker) deleteOriginalWithRetry(ctx context.Context, key string) error {
	const (
		maxAttempts = 3
		baseDelay   = 200 * time.Millisecond
	)

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := w.storage.Delete(ctx, key)
		if err == nil {
			return nil // успех
		}

		// 404 = файл уже удалён (идемпотентно) — считаем успехом.
		// Это может произойти при повторе задачи после частичного сбоя.
		if isNotFoundError(err) {
			w.logger.Debug("original file already deleted", "key", key)
			return nil
		}

		lastErr = err

		// Задержка перед повтором (кроме последней попытки)
		if attempt < maxAttempts-1 {
			delay := baseDelay * time.Duration(1<<uint(attempt))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return fmt.Errorf("delete original file %q after %d attempts: %w", key, maxAttempts, lastErr)
}

// StartupBackfill ставит фото без thumbnail в очередь с контролем нагрузки.
//
// Batched algorithm:
//  1. Загружает фото из БД страницами по backfillBatchSize (LIMIT + cursor).
//  2. Ставит задачи в очередь с retry (через Enqueue).
//  3. Приостанавливается (backpressure) если очередь заполнена > 50%.
//  4. Повторяет пока есть необработанные фото или не отменён контекст.
//
// Защита от edge cases:
//   - 10k+ фото без thumbnail после массового импорта → не перегружает очередь
//   - Graceful shutdown во время backfill → чистый выход через ctx.Done()
//   - MinIO недоступен при старте → задачи в очереди, retry при обработке
//
// Worst-case память: backfillBatchSize * ~500 bytes = 50 KB на страницу.
// Не зависит от общего числа фото.
func (w *ThumbnailWorker) StartupBackfill(ctx context.Context) error {
	start := time.Now()
	w.logger.Info("startup backfill started",
		"batch_size", backfillBatchSize,
		"queue_threshold", backfillQueueThreshold,
	)

	// Метрика длительности
	defer func() {
		metrics.StartupBackfillDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	var (
		cursor            string
		totalProcessed    int
		totalPages        int
		totalBackpressure int // сколько раз приостанавливались
	)

	for {
		// Проверка отмены контекста перед каждым запросом к БД
		if err := ctx.Err(); err != nil {
			w.logger.Info("startup backfill cancelled",
				"processed", totalProcessed,
				"pages", totalPages,
			)
			return err
		}

		// 1. Загружаем страницу фото
		photos, nextCursor, err := w.photoRepo.ListWithoutThumbnailsPaged(ctx, cursor, backfillBatchSize)
		if err != nil {
			w.logger.Error("startup backfill: failed to list photos",
				"cursor", cursor,
				"error", err,
			)
			return err
		}

		if len(photos) == 0 {
			break // Все фото обработаны
		}

		totalPages++

		// 2. Ставим задачи в очередь с backpressure
		for _, p := range photos {
			// Проверка отмены перед каждой задачей
			if err := ctx.Err(); err != nil {
				w.logger.Info("startup backfill cancelled mid-batch",
					"processed", totalProcessed,
					"pages", totalPages,
				)
				return err
			}

			// Backpressure: ждём пока очередь не освободится до порога
			if err := w.waitForQueueCapacity(ctx); err != nil {
				return err
			}

			w.Enqueue(ThumbnailTask{
				PhotoID: p.ID,
				HeroID:  p.HeroID,
				URL:     p.URL,
				FaceBox: p.FaceBox,
			})
			totalProcessed++
		}

		// Если получили меньше batch_size — это последняя страница
		if len(photos) < backfillBatchSize {
			break
		}

		cursor = nextCursor
	}

	metrics.StartupBackfillProcessedTotal.Add(float64(totalProcessed))

	w.logger.Info("startup backfill finished",
		"processed", totalProcessed,
		"pages", totalPages,
		"backpressure_waits", totalBackpressure,
		"duration", time.Since(start),
	)
	return nil
}

// RescanLoop периодически сканирует фото без thumbnail и ставит их в очередь.
//
// Защита от edge cases:
//   - Переполнение очереди при пиковой нагрузке (Enqueue failed)
//   - Сбой воркера (panic при обработке задачи)
//   - Недоступность MinIO при обработке (thumbnail не сгенерирован)
//   - Перезапуск сервера во время обработки
//
// Отличается от StartupBackfill:
//   - Работает постоянно, не только при холодном старте
//   - Ограничен batchSize для контроля нагрузки на БД/очередь
//   - Инкрементальная метрика rescan_triggered_total
func (w *ThumbnailWorker) RescanLoop(ctx context.Context) {
	ticker := time.NewTicker(rescanInterval)
	defer ticker.Stop()

	w.logger.Info("thumbnail rescan loop started",
		"interval", rescanInterval,
		"batch_size", rescanBatchSize,
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("thumbnail rescan loop stopped")
			return
		case <-ticker.C:
			w.rescanOnce(ctx)
		}
	}
}

// rescanOnce выполняет один цикл сканирования.
func (w *ThumbnailWorker) rescanOnce(ctx context.Context) {
	// Используем ListWithoutThumbnails с LIMIT через контекст.
	// Для простоты получаем все и обрезаем до batchSize —
	// обычно таких фото немного (единицы), не сотни.
	photos, err := w.photoRepo.ListWithoutThumbnails(ctx)
	if err != nil {
		w.logger.Error("rescan: failed to list photos without thumbnails", "error", err)
		return
	}

	if len(photos) == 0 {
		return
	}

	// Ограничиваем batch для контроля нагрузки
	batch := photos
	if len(batch) > rescanBatchSize {
		batch = batch[:rescanBatchSize]
	}

	recovered := 0
	for _, p := range batch {
		w.Enqueue(ThumbnailTask{
			PhotoID: p.ID,
			HeroID:  p.HeroID,
			URL:     p.URL,
			FaceBox: p.FaceBox,
		})
		recovered++
	}

	if recovered > 0 {
		metrics.RescanTriggeredTotal.Add(float64(recovered))
		w.logger.Info("rescan: recovered photos without thumbnails",
			"recovered", recovered,
			"total_without_thumbnail", len(photos),
		)
	}
}

// waitForQueueCapacity блокирует выполнение пока очередь не освободится
// ниже порога backfillQueueThreshold или не будет отменён контекст.
//
// Используется в StartupBackfill для backpressure — предотвращает
// переполнение очереди при массовом импорте фото.
func (w *ThumbnailWorker) waitForQueueCapacity(ctx context.Context) error {
	threshold := int(float64(cap(w.queue)) * backfillQueueThreshold)
	ticker := time.NewTicker(backfillCheckInterval)
	defer ticker.Stop()

	for {
		if len(w.queue) < threshold {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Продолжаем проверку
		}
	}
}

// isNotFoundError проверяет, является ли ошибкой "объект не найден" (404).
// В MinIO/S3 удаление несуществующего объекта обычно возвращает 204 (успех),
// но некоторые конфигурации возвращают 404. Обрабатываем оба случая.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		return minioErr.StatusCode == http.StatusNotFound
	}
	// Fallback: проверяем текст ошибки для обёрнутых ошибок
	errStr := err.Error()
	return strings.Contains(errStr, "404") || strings.Contains(errStr, "NoSuchKey")
}
