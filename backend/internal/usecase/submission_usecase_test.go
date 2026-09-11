package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"codeberg.org/Thr0TT1e/emh/backend/internal/auth"
	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// ============================================================
// Хелперы
// ============================================================

// newSubmissionUseCase создаёт SubmissionUseCase с моками.
func newSubmissionUseCase(
	repo repository.SubmissionRepository,
	heroRepo repository.HeroRepository,
) SubmissionUseCase {
	return NewSubmissionUseCase(repo, heroRepo, nil, discardTestLogger())
}

// ============================================================
// Тесты Create
// ============================================================

// TestSubmissionUseCase_Create_Success проверяет успешное создание заявки.
func TestSubmissionUseCase_Create_Success(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.CreateSubmissionParams{
		SubmitterName:  "Иван Иванов",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    `{"bio":"updated"}`,
	}

	id, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 Create call, got %d", len(repo.created))
	}
	if repo.created[0].SubmitterName != params.SubmitterName {
		t.Errorf("SubmitterName = %s, want %s", repo.created[0].SubmitterName, params.SubmitterName)
	}
}

// TestSubmissionUseCase_Create_WithTargetHero проверяет создание с привязкой к герою.
func TestSubmissionUseCase_Create_WithTargetHero(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			if id == "hero-1" {
				return &domain.Hero{ID: "hero-1"}, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	uc := newSubmissionUseCase(repo, heroRepo)

	heroID := "hero-1"
	params := domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroID:   &heroID,
		PayloadJSON:    `{"correction":"rank"}`,
	}

	id, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if id == "" {
		t.Error("Create returned empty id")
	}
}

// TestSubmissionUseCase_Create_EmptyTargetHero пропускает проверку при пустом TargetHeroID.
func TestSubmissionUseCase_Create_EmptyTargetHero(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			// Не должно вызываться для пустого TargetHeroID
			t.Fatal("heroRepo.GetByID should not be called for empty TargetHeroID")
			return nil, nil
		},
	}
	uc := newSubmissionUseCase(repo, heroRepo)

	emptyHeroID := ""
	params := domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroID:   &emptyHeroID,
		PayloadJSON:    `{}`,
	}

	_, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

// TestSubmissionUseCase_Create_NilTargetHero пропускает проверку при nil TargetHeroID.
func TestSubmissionUseCase_Create_NilTargetHero(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroID:   nil, // не привязано к герою
		PayloadJSON:    `{"new_hero":"data"}`,
	}

	_, err := uc.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

// TestSubmissionUseCase_Create_EmptyName проверяет ошибку для пустого имени.
func TestSubmissionUseCase_Create_EmptyName(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.CreateSubmissionParams{
		SubmitterName:  "",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    `{}`,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for empty submitter name")
	}
	if !strings.Contains(err.Error(), "submitter name is required") {
		t.Errorf("expected 'submitter name is required', got %v", err)
	}
}

// TestSubmissionUseCase_Create_InvalidEmail проверяет ошибку для невалидного email.
func TestSubmissionUseCase_Create_InvalidEmail(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	tests := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"no @", "ivanexample.com"},
		{"no domain", "ivan@"},
		{"spaces", "ivan @example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := domain.CreateSubmissionParams{
				SubmitterName:  "Иван",
				SubmitterEmail: tt.email,
				PayloadJSON:    `{}`,
			}

			_, err := uc.Create(context.Background(), params)
			if err == nil {
				t.Fatal("expected error for invalid email")
			}
			if !strings.Contains(err.Error(), "invalid submitter email") {
				t.Errorf("expected 'invalid submitter email', got %v", err)
			}
		})
	}
}

// TestSubmissionUseCase_Create_ValidEmails проверяет прохождение валидных email-адресов.
func TestSubmissionUseCase_Create_ValidEmails(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	validEmails := []string{
		"simple@example.com",
		"user+tag@example.com",
		"user.name@sub.domain.co.uk",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			params := domain.CreateSubmissionParams{
				SubmitterName:  "Иван",
				SubmitterEmail: email,
				PayloadJSON:    `{}`,
			}

			_, err := uc.Create(context.Background(), params)
			if err != nil {
				t.Errorf("expected success for %s, got %v", email, err)
			}
		})
	}
}

// TestSubmissionUseCase_Create_InvalidJSON проверяет ошибку для невалидного JSON.
func TestSubmissionUseCase_Create_InvalidJSON(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	invalidJSONs := []struct {
		name    string
		payload string
	}{
		{"empty", ""},
		{"incomplete", `{`},
		{"broken", `{"key":}`},
		{"plain text", `not a json`},
	}

	for _, tt := range invalidJSONs {
		t.Run(tt.name, func(t *testing.T) {
			params := domain.CreateSubmissionParams{
				SubmitterName:  "Иван",
				SubmitterEmail: "ivan@example.com",
				PayloadJSON:    tt.payload,
			}

			_, err := uc.Create(context.Background(), params)
			if err == nil {
				t.Fatal("expected error for invalid JSON")
			}
			if !strings.Contains(err.Error(), "not valid JSON") {
				t.Errorf("expected 'not valid JSON', got %v", err)
			}
		})
	}
}

// TestSubmissionUseCase_Create_ValidJSON проверяет прохождение валидного JSON.
func TestSubmissionUseCase_Create_ValidJSON(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	validJSONs := []string{
		`{}`,
		`{"key":"value"}`,
		`{"nested":{"data":123}}`,
		`[]`,
		`["array","of","strings"]`,
	}

	for _, payload := range validJSONs {
		t.Run(payload, func(t *testing.T) {
			params := domain.CreateSubmissionParams{
				SubmitterName:  "Иван",
				SubmitterEmail: "ivan@example.com",
				PayloadJSON:    payload,
			}

			_, err := uc.Create(context.Background(), params)
			if err != nil {
				t.Errorf("expected success for JSON %q, got %v", payload, err)
			}
		})
	}
}

// TestSubmissionUseCase_Create_TargetHeroNotFound проверяет ошибку при несуществующем герое.
func TestSubmissionUseCase_Create_TargetHeroNotFound(t *testing.T) {
	repo := &mockSubmissionRepository{}
	heroRepo := &mockHeroRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Hero, error) {
			return nil, domain.ErrNotFound
		},
	}
	uc := newSubmissionUseCase(repo, heroRepo)

	heroID := "nonexistent-hero"
	params := domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		TargetHeroID:   &heroID,
		PayloadJSON:    `{}`,
	}

	_, err := uc.Create(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for nonexistent target hero")
	}
	if !strings.Contains(err.Error(), "target hero not found") {
		t.Errorf("expected 'target hero not found', got %v", err)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected wrapped ErrNotFound, got %v", err)
	}
}

// ============================================================
// Тесты GetByID
// ============================================================

// TestSubmissionUseCase_GetByID_Success проверяет успешный возврат заявки.
func TestSubmissionUseCase_GetByID_Success(t *testing.T) {
	expected := &domain.Submission{
		ID:             "sub-1",
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    `{"bio":"test"}`,
		Status:         domain.StatusDraft,
	}
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return expected, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	result, err := uc.GetByID(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("ID = %s, want %s", result.ID, expected.ID)
	}
	if result.SubmitterName != expected.SubmitterName {
		t.Errorf("SubmitterName = %s, want %s", result.SubmitterName, expected.SubmitterName)
	}
}

// TestSubmissionUseCase_GetByID_NotFound проверяет ошибку при отсутствии заявки.
func TestSubmissionUseCase_GetByID_NotFound(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return nil, domain.ErrNotFound
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, err := uc.GetByID(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ============================================================
// Тесты List
// ============================================================

// TestSubmissionUseCase_List_DefaultLimit проверяет установку default limit.
func TestSubmissionUseCase_List_DefaultLimit(t *testing.T) {
	var capturedFilter domain.SubmissionFilter
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			capturedFilter = f
			return []*domain.Submission{}, "", 0, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, _, _, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if capturedFilter.Limit != 20 {
		t.Errorf("Limit = %d, want 20 (default)", capturedFilter.Limit)
	}
}

// TestSubmissionUseCase_List_NegativeLimit проверяет установку default limit для отрицательного значения.
func TestSubmissionUseCase_List_NegativeLimit(t *testing.T) {
	var capturedFilter domain.SubmissionFilter
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			capturedFilter = f
			return []*domain.Submission{}, "", 0, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, _, _, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: -5})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if capturedFilter.Limit != 20 {
		t.Errorf("Limit = %d, want 20 (default for negative)", capturedFilter.Limit)
	}
}

// TestSubmissionUseCase_List_LimitExceeded проверяет cap на 100.
func TestSubmissionUseCase_List_LimitExceeded(t *testing.T) {
	var capturedFilter domain.SubmissionFilter
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			capturedFilter = f
			return []*domain.Submission{}, "", 0, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, _, _, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: 500})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if capturedFilter.Limit != 100 {
		t.Errorf("Limit = %d, want 100 (max cap)", capturedFilter.Limit)
	}
}

// TestSubmissionUseCase_List_NormalLimit проверяет что нормальный limit не меняется.
func TestSubmissionUseCase_List_NormalLimit(t *testing.T) {
	var capturedFilter domain.SubmissionFilter
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			capturedFilter = f
			return []*domain.Submission{}, "", 0, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, _, _, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: 50})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if capturedFilter.Limit != 50 {
		t.Errorf("Limit = %d, want 50 (unchanged)", capturedFilter.Limit)
	}
}

// TestSubmissionUseCase_List_Success проверяет успешный возврат списка.
func TestSubmissionUseCase_List_Success(t *testing.T) {
	expected := []*domain.Submission{
		{ID: "sub-1", Status: domain.StatusDraft},
		{ID: "sub-2", Status: domain.StatusDraft},
	}
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			return expected, "next-cursor", 42, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	results, cursor, total, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: 20})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(results))
	}
	if cursor != "next-cursor" {
		t.Errorf("cursor = %s, want next-cursor", cursor)
	}
	if total != 42 {
		t.Errorf("total = %d, want 42", total)
	}
}

// TestSubmissionUseCase_List_Error проверяет ошибку репозитория.
func TestSubmissionUseCase_List_Error(t *testing.T) {
	repo := &mockSubmissionRepository{
		listFunc: func(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
			return nil, "", 0, errors.New("db error")
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	_, _, _, err := uc.List(context.Background(), domain.SubmissionFilter{Limit: 20})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

// ============================================================
// Тесты Review
// ============================================================

// TestSubmissionUseCase_Review_ApproveSuccess проверяет успешное одобрение.
func TestSubmissionUseCase_Review_ApproveSuccess(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
		},
		reviewFunc: func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
			return &domain.Submission{ID: p.ID, Status: domain.StatusPublished}, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.ReviewSubmissionParams{
		ID:               "sub-1",
		Decision:         domain.ReviewDecisionApprove,
		ModeratorComment: "Принято",
	}

	result, err := uc.Review(context.Background(), params)
	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}
	if result.Status != domain.StatusPublished {
		t.Errorf("Status = %v, want Published", result.Status)
	}
	if len(repo.reviewed) != 1 {
		t.Fatalf("expected 1 Review call, got %d", len(repo.reviewed))
	}
}

// TestSubmissionUseCase_Review_RejectSuccess проверяет успешное отклонение.
func TestSubmissionUseCase_Review_RejectSuccess(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.ReviewSubmissionParams{
		ID:               "sub-1",
		Decision:         domain.ReviewDecisionReject,
		ModeratorComment: "Данных недостаточно",
	}

	_, err := uc.Review(context.Background(), params)
	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}
}

// TestSubmissionUseCase_Review_AlreadyReviewed проверяет ошибку для уже обработанной заявки.
func TestSubmissionUseCase_Review_AlreadyReviewed(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{ID: id, Status: domain.StatusPublished}, nil
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.ReviewSubmissionParams{
		ID:       "sub-1",
		Decision: domain.ReviewDecisionApprove,
	}

	_, err := uc.Review(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for already reviewed submission")
	}
	if !strings.Contains(err.Error(), "submission already reviewed") {
		t.Errorf("expected 'submission already reviewed', got %v", err)
	}
}

// TestSubmissionUseCase_Review_NotFound проверяет ошибку при отсутствии заявки.
func TestSubmissionUseCase_Review_NotFound(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return nil, domain.ErrNotFound
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.ReviewSubmissionParams{
		ID:       "nonexistent",
		Decision: domain.ReviewDecisionApprove,
	}

	_, err := uc.Review(context.Background(), params)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestSubmissionUseCase_Review_RepoError проверяет ошибку от репозитория.
func TestSubmissionUseCase_Review_RepoError(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
		},
		reviewFunc: func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
			return nil, errors.New("db error")
		},
	}
	heroRepo := &mockHeroRepository{}
	uc := newSubmissionUseCase(repo, heroRepo)

	params := domain.ReviewSubmissionParams{
		ID:       "sub-1",
		Decision: domain.ReviewDecisionApprove,
	}

	_, err := uc.Review(context.Background(), params)
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

// ─── Анти-дубликат ─────────────────────────────────────────────────────────

// TestSubmissionUseCase_Create_DuplicateDetected проверяет, что при попытке
// создать заявку с контентом, идентичным существующей активной заявке,
// возвращается ErrSubmissionDuplicate.
func TestSubmissionUseCase_Create_DuplicateDetected(t *testing.T) {
	repo := &mockSubmissionRepository{
		findByContentHashFn: func(ctx context.Context, hash string) (*domain.Submission, error) {
			// Симулируем найденный дубликат со статусом DRAFT (активная)
			return &domain.Submission{
				ID:          "existing-duplicate-id",
				Status:      domain.StatusDraft,
				ContentHash: hash,
			}, nil
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	_, err := uc.Create(context.Background(), domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    `{"first_name": "Пётр", "last_name": "Иванов"}`,
	})

	if err == nil {
		t.Fatal("expected ErrSubmissionDuplicate, got nil")
	}
	if !errors.Is(err, domain.ErrSubmissionDuplicate) {
		t.Errorf("want ErrSubmissionDuplicate, got: %v", err)
	}

	// Create в репозиторий НЕ должен был вызваться
	repo.mu.Lock()
	createCount := len(repo.createCalls)
	repo.mu.Unlock()
	if createCount != 0 {
		t.Errorf("repo.Create should not be called on duplicate, got %d calls", createCount)
	}
}

// TestSubmissionUseCase_Create_RejectedCanBeResubmitted проверяет, что
// отклонённая ранее заявка не считается дубликатом — пользователь может
// отправить её снова после исправлений.
func TestSubmissionUseCase_Create_RejectedCanBeResubmitted(t *testing.T) {
	repo := &mockSubmissionRepository{
		findByContentHashFn: func(ctx context.Context, hash string) (*domain.Submission, error) {
			// FindByContentHash в реальной реализации фильтрует status != REJECTED,
			// поэтому для отклонённой заявки возвращает ErrNotFound.
			return nil, domain.ErrNotFound
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	id, err := uc.Create(context.Background(), domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    `{"first_name": "Пётр", "last_name": "Иванов"}`,
	})

	if err != nil {
		t.Fatalf("rejected submission should be re-submittable, got error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty submission ID")
	}
}

// TestSubmissionUseCase_Create_NormalizesJSON проверяет, что семантически
// идентичные JSON (разный порядок ключей, форматирование) дают одинаковый
// content_hash и детектируются как дубликат.
func TestSubmissionUseCase_Create_NormalizesJSON(t *testing.T) {
	var capturedHash string
	repo := &mockSubmissionRepository{
		findByContentHashFn: func(ctx context.Context, hash string) (*domain.Submission, error) {
			capturedHash = hash
			return nil, domain.ErrNotFound // не дубликат
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	// Два JSON с одинаковым содержанием, но разным форматированием
	jsonA := `{"first_name": "Пётр", "last_name": "Иванов", "rank": "рядовой"}`
	jsonB := `{
		"rank": "рядовой",
		"last_name": "Иванов",
		"first_name": "Пётр"
	}`

	// Отправляем первый
	_, err := uc.Create(context.Background(), domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    jsonA,
	})
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	hashA := capturedHash

	// Отправляем второй
	_, err = uc.Create(context.Background(), domain.CreateSubmissionParams{
		SubmitterName:  "Иван",
		SubmitterEmail: "ivan@example.com",
		PayloadJSON:    jsonB,
	})
	if err != nil {
		t.Fatalf("second Create failed: %v", err)
	}
	hashB := capturedHash

	// Хеши должны совпадать — нормализация устранила различия в форматировании
	if hashA != hashB {
		t.Errorf("normalized JSON should produce identical hashes:\n  hashA=%s\n  hashB=%s", hashA, hashB)
	}

	// Базовая проверка: SHA-256 hex = 64 символа
	if len(hashA) != 64 {
		t.Errorf("expected SHA-256 hex (64 chars), got %d chars", len(hashA))
	}
}

// ─── Аудит модерации ───────────────────────────────────────────────────────

// TestSubmissionUseCase_Review_CreatesAuditRecord проверяет, что после
// успешной модерации создаётся запись аудита с правильным reviewerName,
// decision и comment.
func TestSubmissionUseCase_Review_CreatesAuditRecord(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{
				ID:     "sub-001",
				Status: domain.StatusDraft,
			}, nil
		},
		reviewFunc: func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
			comment := p.ModeratorComment
			return &domain.Submission{
				ID:               p.ID,
				Status:           domain.StatusPublished,
				ModeratorComment: &comment,
			}, nil
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	// Имитируем JWT-контекст модератора
	ctx := interceptor.ContextWithClaims(context.Background(), &auth.Claims{
		Username: "moderator-ivanov",
		Role:     "admin",
	})

	_, err := uc.Review(ctx, domain.ReviewSubmissionParams{
		ID:               "sub-001",
		Decision:         domain.ReviewDecisionApprove,
		ModeratorComment: "Данные проверены, добавляем",
	})
	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	// Проверяем, что CreateReview был вызван ровно 1 раз
	repo.mu.Lock()
	calls := repo.createReviewCalls
	repo.mu.Unlock()

	if len(calls) != 1 {
		t.Fatalf("expected 1 CreateReview call, got %d", len(calls))
	}

	got := calls[0]
	if got.SubmissionID != "sub-001" {
		t.Errorf("SubmissionID = %q, want 'sub-001'", got.SubmissionID)
	}
	if got.ReviewerName != "moderator-ivanov" {
		t.Errorf("ReviewerName = %q, want 'moderator-ivanov'", got.ReviewerName)
	}
	if got.Decision != domain.ReviewDecisionApprove {
		t.Errorf("Decision = %v, want Approve", got.Decision)
	}
	if got.Comment != "Данные проверены, добавляем" {
		t.Errorf("Comment = %q, want expected text", got.Comment)
	}
}

// TestSubmissionUseCase_Review_UnknownReviewerWhenNoClaims проверяет,
// что при отсутствии JWT-claims в контексте reviewerName = "unknown".
// Это защита от потери аудита при нештатных вызовах.
func TestSubmissionUseCase_Review_UnknownReviewerWhenNoClaims(t *testing.T) {
	repo := &mockSubmissionRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Submission, error) {
			return &domain.Submission{ID: id, Status: domain.StatusDraft}, nil
		},
		reviewFunc: func(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
			return &domain.Submission{ID: p.ID, Status: domain.StatusPublished}, nil
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	// Пустой контекст без claims
	_, err := uc.Review(context.Background(), domain.ReviewSubmissionParams{
		ID:       "sub-002",
		Decision: domain.ReviewDecisionReject,
	})
	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	repo.mu.Lock()
	calls := repo.createReviewCalls
	repo.mu.Unlock()

	if len(calls) != 1 {
		t.Fatalf("expected 1 CreateReview call, got %d", len(calls))
	}
	if calls[0].ReviewerName != "unknown" {
		t.Errorf("ReviewerName = %q, want 'unknown' (no claims in ctx)", calls[0].ReviewerName)
	}
}

// ─── ListReviews ───────────────────────────────────────────────────────────

// TestSubmissionUseCase_ListReviews делегирует в репозиторий.
func TestSubmissionUseCase_ListReviews(t *testing.T) {
	expectedReviews := []*domain.SubmissionReview{
		{
			ID:           "r-001",
			SubmissionID: "sub-001",
			ReviewerName: "mod-1",
			Decision:     domain.ReviewDecisionApprove,
			Comment:      "ok",
		},
	}
	repo := &mockSubmissionRepository{
		listReviewsFn: func(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error) {
			if submissionID != "sub-001" {
				t.Errorf("submissionID = %q, want 'sub-001'", submissionID)
			}
			return expectedReviews, nil
		},
	}
	uc := newSubmissionUseCase(repo, &mockHeroRepository{})

	reviews, err := uc.ListReviews(context.Background(), "sub-001")
	if err != nil {
		t.Fatalf("ListReviews failed: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("got %d reviews, want 1", len(reviews))
	}
	if reviews[0].ReviewerName != "mod-1" {
		t.Errorf("ReviewerName = %q, want 'mod-1'", reviews[0].ReviewerName)
	}
}
