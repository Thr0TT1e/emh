package domain

import "time"

// ReviewDecision решение модератора по заявке. Значения совпадают с proto enum.
type ReviewDecision int

const (
	ReviewDecisionUnspecified ReviewDecision = iota
	ReviewDecisionApprove
	ReviewDecisionReject
)

// Submission представляет пользовательскую заявку на добавление/исправление данных.
type Submission struct {
	ID               string
	SubmitterName    string
	SubmitterEmail   string
	TargetHeroID     *string
	PayloadJSON      string
	AttachmentURLs   []string
	Status           PublicationStatus
	ModeratorComment *string
	ContentHash      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CreateSubmissionParams параметры создания заявки.
type CreateSubmissionParams struct {
	SubmitterName  string
	SubmitterEmail string
	TargetHeroID   *string
	PayloadJSON    string
	AttachmentURLs []string
	ContentHash    string
}

// ReviewSubmissionParams параметры модерации заявки.
type ReviewSubmissionParams struct {
	ID               string
	Decision         ReviewDecision
	ModeratorComment string
}

// SubmissionFilter параметры фильтрации списка заявок.
type SubmissionFilter struct {
	Status *PublicationStatus
	Cursor string
	Limit  int
}

// SubmissionReview запись аудита модерации.
type SubmissionReview struct {
	ID           string
	SubmissionID string
	ReviewerName string
	Decision     ReviewDecision
	Comment      string
	CreatedAt    time.Time
}

// CreateSubmissionReviewParams параметры записи аудита.
type CreateSubmissionReviewParams struct {
	SubmissionID string
	ReviewerName string
	Decision     ReviewDecision
	Comment      string
}
