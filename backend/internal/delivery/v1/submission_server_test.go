package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"connectrpc.com/connect"
)

// --- Моки ---

type mockSubmissionUseCase struct {
	createFunc      func(ctx context.Context, p domain.CreateSubmissionParams) (string, error)
	getByIDFunc     func(ctx context.Context, id string) (*domain.Submission, error)
	listFunc        func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error)
	reviewFunc      func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error)
	listReviewsFunc func(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error)
}

func (m *mockSubmissionUseCase) Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return "submission-id-001", nil
}

func (m *mockSubmissionUseCase) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
}

func (m *mockSubmissionUseCase) List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return []*domain.Submission{}, "", 0, nil
}

func (m *mockSubmissionUseCase) Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
	if m.reviewFunc != nil {
		return m.reviewFunc(ctx, p)
	}
	return &domain.Submission{ID: p.ID, Status: domain.StatusPublished}, nil
}

// Добавить метод:
func (m *mockSubmissionUseCase) ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error) {
	if m.listReviewsFunc != nil {
		return m.listReviewsFunc(ctx, submissionID)
	}
	return []*domain.SubmissionReview{}, nil
}

// --- Хелперы ---

func setupSubmissionServer(
	t *testing.T,
	submissionUC *mockSubmissionUseCase,
) (emhv1connect.SubmissionServiceClient, *httptest.Server, func()) {
	t.Helper()

	server := NewSubmissionServer(submissionUC, discardLogger())

	path, handler := emhv1connect.NewSubmissionServiceHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := httptest.NewServer(mux)
	client := emhv1connect.NewSubmissionServiceClient(
		http.DefaultClient,
		httpServer.URL,
	)

	cleanup := func() {
		httpServer.Close()
	}

	return client, httpServer, cleanup
}

// --- Тесты CreateSubmission ---

func TestSubmissionServer_CreateSubmission_Success(t *testing.T) {
	submissionUC := &mockSubmissionUseCase{}
	client, _, cleanup := setupSubmissionServer(t, submissionUC)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.CreateSubmission(ctx, connect.NewRequest(&emhv1.CreateSubmissionRequest{
		SubmitterName:  "Иван Иванов",
		SubmitterEmail: "ivan@example.com",
		PayloadJson:    `{"bio":"updated"}`,
	}))
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}

	if resp.Msg.SubmissionId == "" {
		t.Error("SubmissionId should not be empty")
	}
}

func TestSubmissionServer_CreateSubmission_WithTargetHero(t *testing.T) {
	var capturedParams domain.CreateSubmissionParams
	submissionUC := &mockSubmissionUseCase{
		createFunc: func(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
			capturedParams = p
			return "submission-id-001", nil
		},
	}
	client, _, cleanup := setupSubmissionServer(t, submissionUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.CreateSubmission(ctx, connect.NewRequest(&emhv1.CreateSubmissionRequest{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroId:   "hero-123",
		PayloadJson:    `{"correction":"rank"}`,
	}))
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}

	if capturedParams.TargetHeroID == nil {
		t.Fatal("TargetHeroID should not be nil")
	}
	if *capturedParams.TargetHeroID != "hero-123" {
		t.Errorf("TargetHeroID = %q, want hero-123", *capturedParams.TargetHeroID)
	}
}

func TestSubmissionServer_CreateSubmission_NotFound(t *testing.T) {
	submissionUC := &mockSubmissionUseCase{
		createFunc: func(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
			return "", domain.ErrNotFound
		},
	}
	client, _, cleanup := setupSubmissionServer(t, submissionUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.CreateSubmission(ctx, connect.NewRequest(&emhv1.CreateSubmissionRequest{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroId:   "nonexistent",
		PayloadJson:    `{}`,
	}))
	if err == nil {
		t.Fatal("expected error for nonexistent hero")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeNotFound {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeNotFound)
	}
}

func TestSubmissionServer_CreateSubmission_InvalidArgument(t *testing.T) {
	submissionUC := &mockSubmissionUseCase{
		createFunc: func(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
			return "", domain.ErrInvalidSubmitterEmail
		},
	}
	client, _, cleanup := setupSubmissionServer(t, submissionUC)
	defer cleanup()

	ctx := context.Background()
	_, err := client.CreateSubmission(ctx, connect.NewRequest(&emhv1.CreateSubmissionRequest{
		SubmitterName:  "Иван",
		SubmitterEmail: "invalid-email",
		PayloadJson:    `{}`,
	}))
	if err == nil {
		t.Fatal("expected error for invalid email")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}
