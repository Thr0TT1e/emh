package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"connectrpc.com/connect"
)

// --- Моки ---

type mockMediaUseCase struct {
	getUploadURLFunc       func(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error)
	batchGetUploadURLsFunc func(ctx context.Context, fileType domain.UploadType, files []domain.FileToUpload) ([]domain.UploadUrlInfo, error)
}

func (m *mockMediaUseCase) GetUploadURL(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error) {
	if m.getUploadURLFunc != nil {
		return m.getUploadURLFunc(ctx, p)
	}
	return &domain.UploadResult{
		UploadURL: "https://s3.example.com/presigned",
		PublicURL: "https://s3.example.com/public/file.jpg",
		Key:       "heroes/2026/08/25/uuid.jpg",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

func (m *mockMediaUseCase) BatchGetUploadURLs(ctx context.Context, fileType domain.UploadType, files []domain.FileToUpload) ([]domain.UploadUrlInfo, error) {
	if m.batchGetUploadURLsFunc != nil {
		return m.batchGetUploadURLsFunc(ctx, fileType, files)
	}
	results := make([]domain.UploadUrlInfo, len(files))
	for i, f := range files {
		results[i] = domain.UploadUrlInfo{
			Filename:  f.Filename,
			UploadURL: "https://s3.example.com/presigned/" + f.Filename,
			PublicURL: "https://s3.example.com/public/" + f.Filename,
			ExpiresAt: time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		}
	}
	return results, nil
}

// --- Хелперы ---

func setupMediaServer(
	t *testing.T,
	mediaUC *mockMediaUseCase,
) (emhv1connect.MediaServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewMediaServer(mediaUC, discardLogger())

	path, handler := emhv1connect.NewMediaServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewMediaServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты GetUploadUrl ---

func TestMediaServer_GetUploadUrl_Success(t *testing.T) {
	mediaUC := &mockMediaUseCase{}
	client, _, cleanup := setupMediaServer(t, mediaUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetUploadUrl(ctx, connect.NewRequest(&emhv1.GetUploadUrlRequest{
		Type:        emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Filename:    "photo.jpg",
		ContentType: "image/jpeg",
	}))
	if err != nil {
		t.Fatalf("GetUploadUrl failed: %v", err)
	}

	if resp.Msg.UploadUrl == "" {
		t.Error("UploadUrl should not be empty")
	}
	if resp.Msg.PublicUrl == "" {
		t.Error("PublicUrl should not be empty")
	}
	if resp.Msg.ExpiresAt == "" {
		t.Error("ExpiresAt should not be empty")
	}
}

func TestMediaServer_GetUploadUrl_InvalidContentType(t *testing.T) {
	mediaUC := &mockMediaUseCase{
		getUploadURLFunc: func(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error) {
			return nil, domain.ErrContentTypeNotAllowed
		},
	}
	client, _, cleanup := setupMediaServer(t, mediaUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetUploadUrl(ctx, connect.NewRequest(&emhv1.GetUploadUrlRequest{
		Type:        emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Filename:    "malware.exe",
		ContentType: "application/octet-stream",
	}))
	if err == nil {
		t.Fatal("expected error for invalid content type")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

func TestMediaServer_GetUploadUrl_UnsupportedType(t *testing.T) {
	mediaUC := &mockMediaUseCase{
		getUploadURLFunc: func(ctx context.Context, p domain.UploadParams) (*domain.UploadResult, error) {
			return nil, domain.ErrUnsupportedUploadType
		},
	}
	client, _, cleanup := setupMediaServer(t, mediaUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.GetUploadUrl(ctx, connect.NewRequest(&emhv1.GetUploadUrlRequest{
		Type:        emhv1.UploadType_UPLOAD_TYPE_UNSPECIFIED,
		Filename:    "file.txt",
		ContentType: "text/plain",
	}))
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

// --- Тесты BatchGetUploadUrls ---

func TestMediaServer_BatchGetUploadUrls_Success(t *testing.T) {
	mediaUC := &mockMediaUseCase{}
	client, _, cleanup := setupMediaServer(t, mediaUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.BatchGetUploadUrls(ctx, connect.NewRequest(&emhv1.BatchGetUploadUrlsRequest{
		Type: emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Files: []*emhv1.FileToUpload{
			{Filename: "photo1.jpg", ContentType: "image/jpeg"},
			{Filename: "photo2.png", ContentType: "image/png"},
		},
	}))
	if err != nil {
		t.Fatalf("BatchGetUploadUrls failed: %v", err)
	}

	if len(resp.Msg.Urls) != 2 {
		t.Errorf("Urls count = %d, want 2", len(resp.Msg.Urls))
	}
	if resp.Msg.Count != 2 {
		t.Errorf("Count = %d, want 2", resp.Msg.Count)
	}
	if resp.Msg.Urls[0].Filename != "photo1.jpg" {
		t.Errorf("Urls[0].Filename = %q, want photo1.jpg", resp.Msg.Urls[0].Filename)
	}
}

func TestMediaServer_BatchGetUploadUrls_InvalidFile(t *testing.T) {
	mediaUC := &mockMediaUseCase{
		batchGetUploadURLsFunc: func(ctx context.Context, fileType domain.UploadType, files []domain.FileToUpload) ([]domain.UploadUrlInfo, error) {
			return nil, domain.ErrContentTypeNotAllowed
		},
	}
	client, _, cleanup := setupMediaServer(t, mediaUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.BatchGetUploadUrls(ctx, connect.NewRequest(&emhv1.BatchGetUploadUrlsRequest{
		Type: emhv1.UploadType_UPLOAD_TYPE_HERO_PHOTO,
		Files: []*emhv1.FileToUpload{
			{Filename: "malware.exe", ContentType: "application/octet-stream"},
		},
	}))
	if err == nil {
		t.Fatal("expected error for invalid file")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}
