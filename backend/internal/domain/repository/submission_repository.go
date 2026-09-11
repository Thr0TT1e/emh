package repository

import (
	"context"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
)

// SubmissionRepository абстракция для работы с пользовательскими заявками.
type SubmissionRepository interface {
	Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error)
	GetByID(ctx context.Context, id string) (*domain.Submission, error)
	List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error)
	Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error)
	FindByContentHash(ctx context.Context, hash string) (*domain.Submission, error)
	CreateReview(ctx context.Context, p domain.CreateSubmissionReviewParams) (string, error)
	ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error)
}
