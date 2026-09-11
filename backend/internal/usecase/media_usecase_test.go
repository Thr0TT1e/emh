package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// --- Тесты GetUploadURL ---

// TestMediaUseCase_GetUploadURL_Success проверяет успешную генерацию presigned URL.
func TestMediaUseCase_GetUploadURL_Success(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	params := domain.UploadParams{
		Type:        domain.UploadTypeHeroPhoto,
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}

	result, err := uc.GetUploadURL(context.Background(), params)
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	if result.UploadURL == "" {
		t.Error("UploadURL should not be empty")
	}
	if result.PublicURL == "" {
		t.Error("PublicURL should not be empty")
	}
	if result.Key == "" {
		t.Error("Key should not be empty")
	}
	if result.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}

	// Проверяем, что ключ содержит правильный префикс
	if !strings.HasPrefix(result.Key, "heroes/") {
		t.Errorf("expected key to start with 'heroes/', got %s", result.Key)
	}
	// Проверяем, что ключ содержит расширение .jpg
	if !strings.HasSuffix(result.Key, ".jpg") {
		t.Errorf("expected key to end with '.jpg', got %s", result.Key)
	}
}

// TestMediaUseCase_GetUploadURL_AllMIMETypes проверяет все поддерживаемые MIME-типы.
func TestMediaUseCase_GetUploadURL_AllMIMETypes(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	tests := []struct {
		name        string
		uploadType  domain.UploadType
		contentType string
		expectedExt string
	}{
		{"HeroPhoto JPEG", domain.UploadTypeHeroPhoto, "image/jpeg", ".jpg"},
		{"HeroPhoto PNG", domain.UploadTypeHeroPhoto, "image/png", ".png"},
		{"HeroPhoto WebP", domain.UploadTypeHeroPhoto, "image/webp", ".webp"},
		{"AwardImage JPEG", domain.UploadTypeAwardImage, "image/jpeg", ".jpg"},
		{"AwardImage PNG", domain.UploadTypeAwardImage, "image/png", ".png"},
		{"AwardImage WebP", domain.UploadTypeAwardImage, "image/webp", ".webp"},
		{"Submission JPEG", domain.UploadTypeSubmissionAttachment, "image/jpeg", ".jpg"},
		{"Submission PNG", domain.UploadTypeSubmissionAttachment, "image/png", ".png"},
		{"Submission WebP", domain.UploadTypeSubmissionAttachment, "image/webp", ".webp"},
		{"Submission PDF", domain.UploadTypeSubmissionAttachment, "application/pdf", ".pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := domain.UploadParams{
				Type:        tt.uploadType,
				Filename:    "test" + tt.expectedExt,
				ContentType: tt.contentType,
			}

			result, err := uc.GetUploadURL(context.Background(), params)
			if err != nil {
				t.Fatalf("GetUploadURL failed: %v", err)
			}

			if !strings.HasSuffix(result.Key, tt.expectedExt) {
				t.Errorf("expected key to end with %s, got %s", tt.expectedExt, result.Key)
			}
		})
	}
}

// TestMediaUseCase_GetUploadURL_UnsupportedUploadType проверяет ошибку для неподдерживаемого типа.
func TestMediaUseCase_GetUploadURL_UnsupportedUploadType(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	params := domain.UploadParams{
		Type:        domain.UploadType(999),
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}

	_, err := uc.GetUploadURL(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for unsupported upload type")
	}
	if !strings.Contains(err.Error(), "unsupported upload type") {
		t.Errorf("expected 'unsupported upload type' error, got %v", err)
	}
}

// TestMediaUseCase_GetUploadURL_UnsupportedContentType проверяет ошибку для неподдерживаемого content-type.
func TestMediaUseCase_GetUploadURL_UnsupportedContentType(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	params := domain.UploadParams{
		Type:        domain.UploadTypeHeroPhoto,
		Filename:    "malware.exe",
		ContentType: "application/octet-stream",
	}

	_, err := uc.GetUploadURL(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for unsupported content type")
	}
	if !strings.Contains(err.Error(), "content type") {
		t.Errorf("expected 'content type not allowed' error, got %v", err)
	}
}

// TestMediaUseCase_GetUploadURL_StorageError проверяет обработку ошибки от storage.
func TestMediaUseCase_GetUploadURL_StorageError(t *testing.T) {
	storage := &mockMediaStorage{
		presignError: errors.New("storage error"),
	}
	uc := newMediaUseCase(storage)

	params := domain.UploadParams{
		Type:        domain.UploadTypeHeroPhoto,
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}

	_, err := uc.GetUploadURL(context.Background(), params)
	if err == nil {
		t.Fatal("expected error from storage")
	}
	if !strings.Contains(err.Error(), "generate presigned url") {
		t.Errorf("expected 'generate presigned url' error, got %v", err)
	}
}

// TestMediaUseCase_GetUploadURL_Expiry проверяет, что expires_at правильно вычисляется.
func TestMediaUseCase_GetUploadURL_Expiry(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := NewMediaUseCase(storage, 30*time.Minute)

	params := domain.UploadParams{
		Type:        domain.UploadTypeHeroPhoto,
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}

	before := time.Now()
	result, err := uc.GetUploadURL(context.Background(), params)
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}
	after := time.Now()

	// ExpiresAt должно быть примерно через 30 минут
	expectedExpiry := before.Add(30 * time.Minute)
	if result.ExpiresAt.Before(expectedExpiry) || result.ExpiresAt.After(after.Add(30*time.Minute)) {
		t.Errorf("ExpiresAt = %v, expected around %v", result.ExpiresAt, expectedExpiry)
	}
}

// --- Тесты BatchGetUploadURLs ---

// TestMediaUseCase_BatchGetUploadURLs_EmptyList проверяет ошибку для пустого списка.
func TestMediaUseCase_BatchGetUploadURLs_EmptyList(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	_, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, []domain.FileToUpload{})
	if err == nil {
		t.Fatal("expected error for empty files list")
	}
	if !strings.Contains(err.Error(), "files must not be empty") {
		t.Errorf("expected 'files must not be empty' error, got %v", err)
	}
}

// TestMediaUseCase_BatchGetUploadURLs_TooManyFiles проверяет ошибку для слишком большого списка.
func TestMediaUseCase_BatchGetUploadURLs_TooManyFiles(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	files := make([]domain.FileToUpload, 51)
	for i := range files {
		files[i] = domain.FileToUpload{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
		}
	}

	_, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, files)
	if err == nil {
		t.Fatal("expected error for too many files")
	}
	if !strings.Contains(err.Error(), "too many files") {
		t.Errorf("expected 'too many files' error, got %v", err)
	}
}

// TestMediaUseCase_BatchGetUploadURLs_MaxFiles проверяет границу в 50 файлов.
func TestMediaUseCase_BatchGetUploadURLs_MaxFiles(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	files := make([]domain.FileToUpload, 50)
	for i := range files {
		files[i] = domain.FileToUpload{
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
		}
	}

	results, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, files)
	if err != nil {
		t.Fatalf("BatchGetUploadURLs failed: %v", err)
	}

	if len(results) != 50 {
		t.Errorf("expected 50 results, got %d", len(results))
	}
}

// TestMediaUseCase_BatchGetUploadURLs_Success проверяет успешную генерацию нескольких presigned URL.
func TestMediaUseCase_BatchGetUploadURLs_Success(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	files := []domain.FileToUpload{
		{Filename: "photo1.jpg", ContentType: "image/jpeg"},
		{Filename: "photo2.png", ContentType: "image/png"},
		{Filename: "photo3.webp", ContentType: "image/webp"},
	}

	results, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, files)
	if err != nil {
		t.Fatalf("BatchGetUploadURLs failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Проверяем, что все поля заполнены
	for i, result := range results {
		if result.Filename != files[i].Filename {
			t.Errorf("result[%d].Filename = %s, want %s", i, result.Filename, files[i].Filename)
		}
		if result.UploadURL == "" {
			t.Errorf("result[%d].UploadURL should not be empty", i)
		}
		if result.PublicURL == "" {
			t.Errorf("result[%d].PublicURL should not be empty", i)
		}
		if result.ExpiresAt == "" {
			t.Errorf("result[%d].ExpiresAt should not be empty", i)
		}
	}
}

// TestMediaUseCase_BatchGetUploadURLs_DifferentTypes проверяет генерацию для разных типов.
func TestMediaUseCase_BatchGetUploadURLs_DifferentTypes(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	tests := []struct {
		name        string
		uploadType  domain.UploadType
		expectedPfx string
	}{
		{"HeroPhoto", domain.UploadTypeHeroPhoto, "heroes/"},
		{"AwardImage", domain.UploadTypeAwardImage, "awards/"},
		{"Submission", domain.UploadTypeSubmissionAttachment, "submissions/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := []domain.FileToUpload{
				{Filename: "test.jpg", ContentType: "image/jpeg"},
			}

			results, err := uc.BatchGetUploadURLs(context.Background(), tt.uploadType, files)
			if err != nil {
				t.Fatalf("BatchGetUploadURLs failed: %v", err)
			}

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}

			// Проверяем, что upload URL содержит правильный префикс
			if !strings.Contains(results[0].UploadURL, tt.expectedPfx) {
				t.Errorf("expected upload URL to contain %s, got %s", tt.expectedPfx, results[0].UploadURL)
			}
		})
	}
}

// TestMediaUseCase_BatchGetUploadURLs_InvalidFile проверяет ошибку для невалидного файла в пакете.
func TestMediaUseCase_BatchGetUploadURLs_InvalidFile(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := newMediaUseCase(storage)

	files := []domain.FileToUpload{
		{Filename: "photo1.jpg", ContentType: "image/jpeg"},
		{Filename: "malware.exe", ContentType: "application/octet-stream"}, // невалидный
	}

	_, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, files)
	if err == nil {
		t.Fatal("expected error for invalid file")
	}
	if !strings.Contains(err.Error(), "invalid file") {
		t.Errorf("expected 'invalid file' error, got %v", err)
	}
}

// TestMediaUseCase_BatchGetUploadURLs_StorageError проверяет обработку ошибки от storage.
func TestMediaUseCase_BatchGetUploadURLs_StorageError(t *testing.T) {
	storage := &mockMediaStorage{
		presignError: errors.New("storage error"),
	}
	uc := newMediaUseCase(storage)

	files := []domain.FileToUpload{
		{Filename: "photo1.jpg", ContentType: "image/jpeg"},
	}

	_, err := uc.BatchGetUploadURLs(context.Background(), domain.UploadTypeHeroPhoto, files)
	if err == nil {
		t.Fatal("expected error from storage")
	}
	if !strings.Contains(err.Error(), "presign url") {
		t.Errorf("expected 'presign url' error, got %v", err)
	}
}

// TestMediaUseCase_NewMediaUseCase_DefaultExpiry проверяет дефолтное время жизни.
func TestMediaUseCase_NewMediaUseCase_DefaultExpiry(t *testing.T) {
	storage := &mockMediaStorage{}
	uc := NewMediaUseCase(storage, 0) // должно стать 15 минут

	params := domain.UploadParams{
		Type:        domain.UploadTypeHeroPhoto,
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}

	before := time.Now()
	result, err := uc.GetUploadURL(context.Background(), params)
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}
	after := time.Now()

	// ExpiresAt должно быть примерно через 15 минут
	expectedExpiry := before.Add(15 * time.Minute)
	if result.ExpiresAt.Before(expectedExpiry) || result.ExpiresAt.After(after.Add(15*time.Minute)) {
		t.Errorf("ExpiresAt = %v, expected around %v (default 15 minutes)", result.ExpiresAt, expectedExpiry)
	}
}
