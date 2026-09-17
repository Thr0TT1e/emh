package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/domain/repository"
)

// submissionColumns список колонок для единообразия в SELECT/RETURNING.
const submissionColumns = "id, submitter_name, submitter_email, target_hero_id, payload_json, attachment_urls, status, moderator_comment, content_hash, created_at, updated_at"

type submissionRepository struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewSubmissionRepository(pool *pgxpool.Pool) repository.SubmissionRepository {
	return &submissionRepository{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// scanSubmission маппит строку БД в доменную модель.
// Интерфейс scannable переиспользуется из conflict_repository.go.
func scanSubmission(s scannable) (*domain.Submission, error) {
	sub := &domain.Submission{}
	var status int
	var moderatorComment sql.NullString
	var contentHash sql.NullString // content_hash может быть NULL для старых записей

	if err := s.Scan(
		&sub.ID,
		&sub.SubmitterName,
		&sub.SubmitterEmail,
		&sub.TargetHeroID,
		&sub.PayloadJSON,
		&sub.AttachmentURLs,
		&status,
		&moderatorComment,
		&sub.ContentHash,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	); err != nil {
		return nil, err
	}

	sub.Status = domain.PublicationStatus(status)
	if moderatorComment.Valid {
		sub.ModeratorComment = &moderatorComment.String
	}
	if contentHash.Valid {
		sub.ContentHash = contentHash.String
	}

	return sub, nil
}

func (r *submissionRepository) Create(ctx context.Context, p domain.CreateSubmissionParams) (string, error) {
	// Защита от nil: pgx отправит NULL для nil-слайса, что нарушит NOT NULL constraint.
	// Пустой слайс []string{} корректно сериализуется в PostgreSQL '{}' (пустой массив).
	attachments := p.AttachmentURLs
	if attachments == nil {
		attachments = []string{}
	}

	// target_hero_id: пустая строка должна стать NULL в БД (nullable FK)
	var targetHeroID interface{}
	if p.TargetHeroID != nil && *p.TargetHeroID != "" {
		targetHeroID = *p.TargetHeroID
	}

	query := r.sb.Insert("submissions").
		Columns(
			"submitter_name",
			"submitter_email",
			"target_hero_id",
			"payload_json",
			"attachment_urls",
			"status",
			"content_hash",
		).
		Values(
			p.SubmitterName,
			p.SubmitterEmail,
			targetHeroID,
			p.PayloadJSON,
			attachments,
			int(domain.StatusDraft),
			p.ContentHash,
		).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create submission: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		// Обработка race condition: между FindByContentHash и INSERT
		// другая заявка с тем же хешем могла быть создана.
		// Unique constraint violation (SQLSTATE 23505) маппим в ErrSubmissionDuplicate.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.ErrSubmissionDuplicate
		}
		return "", fmt.Errorf("exec create submission: %w", err)
	}
	return id, nil
}

func (r *submissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	query := r.sb.Select(submissionColumns).
		From("submissions").
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get submission: %w", err)
	}

	sub, err := scanSubmission(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("submission not found")
		}
		return nil, fmt.Errorf("exec get submission: %w", err)
	}
	return sub, nil
}

func (r *submissionRepository) List(ctx context.Context, f domain.SubmissionFilter) ([]*domain.Submission, string, int64, error) {
	// uuidv7 монотонен, поэтому id DESC даёт стабильный хронологический порядок
	sel := r.sb.Select(submissionColumns).
		From("submissions").
		OrderBy("id DESC")

	if f.Status != nil {
		sel = sel.Where(squirrel.Eq{"status": int(*f.Status)})
	}
	if f.Cursor != "" {
		sel = sel.Where(squirrel.Lt{"id": f.Cursor})
	}

	sel = sel.Limit(uint64(f.Limit + 1))

	sql, args, err := sel.ToSql()
	if err != nil {
		return nil, "", 0, fmt.Errorf("build list submissions: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("exec list submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*domain.Submission
	for rows.Next() {
		sub, err := scanSubmission(rows)
		if err != nil {
			return nil, "", 0, fmt.Errorf("scan submission: %w", err)
		}
		submissions = append(submissions, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, "", 0, fmt.Errorf("rows iteration error: %w", err)
	}

	var nextCursor string
	if len(submissions) > f.Limit {
		submissions = submissions[:f.Limit]
		nextCursor = submissions[len(submissions)-1].ID
	}

	var total int64
	if f.Cursor == "" {
		countSel := r.sb.Select("COUNT(*)").From("submissions")
		if f.Status != nil {
			countSel = countSel.Where(squirrel.Eq{"status": int(*f.Status)})
		}
		countSQL, countArgs, _ := countSel.ToSql()
		_ = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	}

	return submissions, nextCursor, total, nil
}

func (r *submissionRepository) Review(ctx context.Context, p domain.ReviewSubmissionParams) (*domain.Submission, error) {
	var newStatus domain.PublicationStatus
	switch p.Decision {
	case domain.ReviewDecisionApprove:
		newStatus = domain.StatusPublished
	case domain.ReviewDecisionReject:
		newStatus = domain.StatusArchived
	default:
		return nil, fmt.Errorf("unknown review decision")
	}

	update := r.sb.Update("submissions").
		Set("status", int(newStatus)).
		Set("moderator_comment", p.ModeratorComment).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": p.ID, "status": int(domain.StatusDraft)}).
		Suffix("RETURNING " + submissionColumns)

	sql, args, err := update.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build review submission: %w", err)
	}

	sub, err := scanSubmission(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("submission not found")
		}
		return nil, fmt.Errorf("exec review submission: %w", err)
	}
	return sub, nil
}

func (r *submissionRepository) FindByContentHash(ctx context.Context, hash string) (*domain.Submission, error) {
	// status хранится как smallint: 0=UNSPECIFIED, 1=DRAFT, 2=PUBLISHED, 3=ARCHIVED
	// Ищем активные заявки: DRAFT (1) и PUBLISHED (2) считаются дубликатами.
	// ARCHIVED (3, отклонённые) — можно переотправлять после исправлений.
	// Значение должно совпадать с WHERE в unique index idx_submissions_content_hash_active.
	const archivedStatus = 3

	query := r.sb.Select(submissionColumns).
		From("submissions").
		Where(squirrel.Eq{"content_hash": hash}).
		Where(squirrel.NotEq{"status": archivedStatus}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find by content hash: %w", err)
	}

	sub, err := scanSubmission(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("exec find by content hash: %w", err)
	}
	return sub, nil
}

func (r *submissionRepository) CreateReview(ctx context.Context, p domain.CreateSubmissionReviewParams) (string, error) {
	query := r.sb.Insert("submission_reviews").
		Columns("submission_id", "reviewer_name", "decision", "comment").
		Values(p.SubmissionID, p.ReviewerName, strconv.Itoa(int(p.Decision)), p.Comment).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("build create review: %w", err)
	}

	var id string
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		return "", fmt.Errorf("exec create review: %w", err)
	}
	return id, nil
}

func (r *submissionRepository) ListReviews(ctx context.Context, submissionID string) ([]*domain.SubmissionReview, error) {
	query := r.sb.Select(
		"id", "submission_id", "reviewer_name", "decision", "comment", "created_at",
	).From("submission_reviews").
		Where(squirrel.Eq{"submission_id": submissionID}).
		OrderBy("created_at DESC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list reviews: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*domain.SubmissionReview
	for rows.Next() {
		r := &domain.SubmissionReview{}
		var decision string
		if err := rows.Scan(&r.ID, &r.SubmissionID, &r.ReviewerName,
			&decision, &r.Comment, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		decisionInt, err := strconv.Atoi(decision)
		if err != nil {
			return nil, fmt.Errorf("parse review decision %q: %w", decision, err)
		}
		r.Decision = domain.ReviewDecision(decisionInt)
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}
