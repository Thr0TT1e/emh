package usecase

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/config"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// ============================================================
// Хелперы
// ============================================================

// createPNGHeaderOnly создаёт минимальный валидный PNG только с IHDR-чанком
// (без пиксельных данных). image.DecodeConfig читает только заголовок,
// поэтому этого достаточно для проверки габаритов без потребления памяти.
//
// Использование: создаём файл с габаритами 100000×100000 (~80 байт на диске),
// который при полном декодировании занял бы 40 ГБ в памяти.
func createPNGHeaderOnly(width, height uint32) []byte {
	var buf bytes.Buffer

	// PNG signature (8 байт)
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	// IHDR chunk: ширина, высота, глубина цвета, тип цвета, сжатие, фильтр, интерлейс
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], width)
	binary.BigEndian.PutUint32(ihdr[4:8], height)
	ihdr[8] = 8  // bit depth
	ihdr[9] = 2  // color type: truecolor
	ihdr[10] = 0 // compression method
	ihdr[11] = 0 // filter method
	ihdr[12] = 0 // interlace method

	// Chunk length (4 байта)
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(ihdr)))
	buf.Write(lenBytes)

	// Chunk type + data
	buf.Write([]byte("IHDR"))
	buf.Write(ihdr)

	// CRC32 (обязателен для валидности — image/png проверяет его)
	crc := crc32.ChecksumIEEE(append([]byte("IHDR"), ihdr...))
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)
	buf.Write(crcBytes)

	return buf.Bytes()
}

// createTestPNG создаёт валидное маленькое PNG-изображение (100×100, красный цвет).
func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic("test helper: failed to encode PNG: " + err.Error())
	}
	return buf.Bytes()
}

// newTestThumbnailWorker создаёт воркер для тестов с отключённым логированием.
func newTestThumbnailWorker(
	storage domain.MediaStorage,
	photoRepo *mockPhotoRepositoryForWorker,
) *ThumbnailWorker {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.ThumbnailConfig{
		Width:           480,
		Height:          600,
		Quality:         80,
		OriginalQuality: 90,
		Workers:         2,
		QueueSize:       100,
	}
	return NewThumbnailWorker(photoRepo, storage, logger, cfg)
}

// sampleTask возвращает типовую задачу для воркера.
func sampleTask() ThumbnailTask {
	return ThumbnailTask{
		PhotoID: "photo-1",
		HeroID:  "hero-1",
		URL:     "https://example.com/photos/test.jpg",
	}
}

// ============================================================
// Тесты защиты от decompression bomb
// ============================================================

// TestProcessTask_RejectsOversizedDownload проверяет отклонение файлов > 50 MB.
// Атакующий может загрузить огромный файл через публичный presigned URL;
// воркер должен отклонить его до декодирования.
func TestProcessTask_RejectsOversizedDownload(t *testing.T) {
	oversizedData := make([]byte, maxDownloadBytes+1)

	storage := &mockMediaStorageForWorker{downloadData: oversizedData}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	worker.processTask(context.Background(), sampleTask())

	if photoRepo.updateAfterProcessingCalled {
		t.Error("UpdateAfterProcessing must not be called for oversized download")
	}
}

// TestProcessTask_RejectsLargeDimensions проверяет отклонение изображений
// с габаритами > 10000 пикселей по стороне. Это защита от маленьких по размеру,
// но огромных по габаритам файлов (decompression bomb).
func TestProcessTask_RejectsLargeDimensions(t *testing.T) {
	// 100000×100000 = 10 гигапикселей, ~80 байт в заголовке
	largeHeader := createPNGHeaderOnly(100000, 100000)

	storage := &mockMediaStorageForWorker{downloadData: largeHeader}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	worker.processTask(context.Background(), sampleTask())

	if photoRepo.updateAfterProcessingCalled {
		t.Error("UpdateAfterProcessing must not be called for images with dimensions > 10000")
	}
}

// TestProcessTask_RejectsLargeMegapixels проверяет отклонение изображений
// с общим числом пикселей > 50 мегапикселей, даже если каждая сторона < 10000.
// Пример: 8000×8000 = 64 МП (обе стороны < 10000, но общий объём слишком велик).
func TestProcessTask_RejectsLargeMegapixels(t *testing.T) {
	// 8000×8000 = 64 мегапикселя > лимита 50 МП
	largeHeader := createPNGHeaderOnly(8000, 8000)

	storage := &mockMediaStorageForWorker{downloadData: largeHeader}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	worker.processTask(context.Background(), sampleTask())

	if photoRepo.updateAfterProcessingCalled {
		t.Error("UpdateAfterProcessing must not be called for images > 50 megapixels")
	}
}

// TestProcessTask_ValidImage_PassesValidation проверяет, что валидное маленькое
// изображение проходит валидацию и успешно загружается (thumbnail + оригинал).
func TestProcessTask_ValidImage_PassesValidation(t *testing.T) {
	validPNG := createTestPNG(100, 100)

	storage := &mockMediaStorageForWorker{downloadData: validPNG}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("processTask panicked on valid image: %v", r)
		}
	}()

	worker.processTask(context.Background(), sampleTask())

	// Проверяем, что UpdateAfterProcessing был вызван (успешное завершение)
	if !photoRepo.updateAfterProcessingCalled {
		t.Error("expected UpdateAfterProcessing to be called for valid image")
	}

	// Проверяем, что Upload был вызван минимум 1 раз (thumbnail).
	// Оригинал в PNG → конвертируется в WebP → 2 вызова Upload.
	if len(storage.uploadCalls) < 1 {
		t.Errorf("expected at least 1 Upload call, got %d", len(storage.uploadCalls))
	}

	// Первая загрузка — thumbnail
	thumbUpload := storage.uploadCalls[0]
	if thumbUpload.Key != "thumbnails/photo-1.webp" {
		t.Errorf("thumbnail key = %q, want thumbnails/photo-1.webp", thumbUpload.Key)
	}
	if thumbUpload.ContentType != "image/webp" {
		t.Errorf("thumbnail content type = %q, want image/webp", thumbUpload.ContentType)
	}
	if thumbUpload.DataSize == 0 {
		t.Error("thumbnail data size must be > 0")
	}
}

// TestProcessTask_DownloadError проверяет корректное завершение при ошибке скачивания.
func TestProcessTask_DownloadError(t *testing.T) {
	storage := &mockMediaStorageForWorker{
		downloadErr: context.DeadlineExceeded,
	}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	worker.processTask(context.Background(), sampleTask())

	if photoRepo.updateAfterProcessingCalled {
		t.Error("UpdateAfterProcessing must not be called when download fails")
	}
}

// TestProcessTask_CancelledContext проверяет ранний выход при отменённом контексте.
func TestProcessTask_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	storage := &mockMediaStorageForWorker{downloadData: createTestPNG(100, 100)}
	photoRepo := &mockPhotoRepositoryForWorker{}
	worker := newTestThumbnailWorker(storage, photoRepo)

	worker.processTask(ctx, sampleTask())

	if photoRepo.updateAfterProcessingCalled {
		t.Error("UpdateAfterProcessing must not be called when context is cancelled")
	}
}
