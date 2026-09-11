package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/mail"
	"time"

	"codeberg.org/Thr0TT1e/emh/backend/internal/delivery/interceptor"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
	"codeberg.org/Thr0TT1e/emh/backend/internal/metrics"
)

// SubmissionUseCase бизнес-логика пользовательских заявок.
type SubmissionUseCase interface {
	Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error)
	GetByID(ctx context.Context, id string) (*domain.Submission, error)
	List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error)
	Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error)
	ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error)
}

type submissionUseCase struct {
	repo        repository.SubmissionRepository
	heroRepo    repository.HeroRepository
	emailWorker *SubmissionEmailWorker
	logger      *slog.Logger
}

func NewSubmissionUseCase(
	repo repository.SubmissionRepository,
	heroRepo repository.HeroRepository,
	emailWorker *SubmissionEmailWorker,
	logger *slog.Logger,
) SubmissionUseCase {
	return &submissionUseCase{
		repo:        repo,
		heroRepo:    heroRepo,
		emailWorker: emailWorker,
		logger:      logger,
	}
}

func (uc *submissionUseCase) Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
	if p.SubmitterName == "" {
		return "", domain.ErrSubmitterNameRequired
	}
	if !isValidEmail(p.SubmitterEmail) {
		return "", domain.ErrInvalidSubmitterEmail
	}
	// payload_json должен быть валидным JSON до записи в JSONB
	if !json.Valid([]byte(p.PayloadJSON)) {
		return "", domain.ErrInvalidPayloadJSON
	}

	// Нормализация и хеширование
	normalized, err := normalizePayloadJSON(p.PayloadJSON)
	if err != nil {
		return "", domain.ErrInvalidPayloadJSON
	}
	p.ContentHash = hashPayload(normalized)

	// Проверка дубликата
	existing, err := uc.repo.FindByContentHash(ctx, p.ContentHash)
	if err == nil && existing != nil {
		return "", domain.ErrSubmissionDuplicate
	}

	if p.TargetHeroID != nil && *p.TargetHeroID != "" {
		if _, err := uc.heroRepo.GetByID(ctx, *p.TargetHeroID); err != nil {
			return "", fmt.Errorf("target hero not found: %w", err)
		}
	}

	id, err := uc.repo.Create(ctx, p)
	if err != nil {
		return "", err
	}

	// метрика созданной заявки
	metrics.SubmissionTotal.WithLabelValues("created").Inc()

	return id, nil
}

func (uc *submissionUseCase) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *submissionUseCase) List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	return uc.repo.List(ctx, f)
}

func (uc *submissionUseCase) Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
	existing, err := uc.repo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	// Защита от повторной обработки: рассматриваем только заявки на модерации
	if existing.Status != domain.StatusDraft {
		return nil, domain.ErrSubmissionAlreadyReviewed
	}

	result, err := uc.repo.Review(ctx, p)
	if err != nil {
		return nil, err
	}

	// Аудит модерации — best-effort
	reviewerName := "unknown"
	if claims, ok := interceptor.ClaimsFromContext(ctx); ok && claims != nil {
		reviewerName = claims.Username
	}
	if _, err := uc.repo.CreateReview(ctx, domain.CreateSubmissionReviewParams{
		SubmissionID: p.ID,
		ReviewerName: reviewerName,
		Decision:     p.Decision,
		Comment:      p.ModeratorComment,
	}); err != nil {
		uc.logger.Error("failed to create submission review audit",
			"submission_id", p.ID,
			"error", err,
		)
	}

	// Отправка email-уведомления заявителю (best-effort)
	if uc.emailWorker != nil && result.SubmitterEmail != "" {
		uc.emailWorker.Enqueue(domain.SubmissionEmailNotification{
			SubmitterName:    result.SubmitterName,
			Email:            result.SubmitterEmail,
			SubmissionID:     result.ID,
			Decision:         p.Decision,
			ModeratorComment: p.ModeratorComment,
			Timestamp:        time.Now(),
		})
	}

	status := "approved"
	if p.Decision == domain.ReviewDecisionReject {
		status = "rejected"
	}
	metrics.SubmissionTotal.WithLabelValues(status).Inc()

	return result, nil
}

// ListReviews возвращает историю модерации заявки.
func (uc *submissionUseCase) ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error) {
	return uc.repo.ListReviews(ctx, submissionID)
}

// isValidEmail проверяет корректность email через стандартный парсер.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// normalizePayloadJSON нормализует JSON для детерминированного хеширования.
// Парсит и сериализует обратно — json.Marshal сортирует ключи.
// Семантически идентичные JSON дают одинаковый результат.
func normalizePayloadJSON(raw string) (string, error) {
	var parsed interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", err
	}
	normalized, err := json.Marshal(parsed)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}

// hashPayload вычисляет SHA-256 хеш строки.
func hashPayload(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
