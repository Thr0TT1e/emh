package v1

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// SubmissionAdminServer реализует защищённый emhv1.SubmissionAdminServiceHandler.
type SubmissionAdminServer struct {
	submissionUC usecase.SubmissionUseCase
	logger       *slog.Logger
}

func NewSubmissionAdminServer(uc usecase.SubmissionUseCase, logger *slog.Logger) *SubmissionAdminServer {
	return &SubmissionAdminServer{submissionUC: uc, logger: logger}
}

// ListSubmissions возвращает список заявок с фильтрацией и пагинацией.
func (s *SubmissionAdminServer) ListSubmissions(
	ctx context.Context,
	req *connect.Request[emhv1.ListSubmissionsRequest],
) (*connect.Response[emhv1.ListSubmissionsResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := domain.SubmissionFilter{
		Cursor: req.Msg.Pagination.GetCursor(),
		Limit:  limit,
	}
	if req.Msg.Status != emhv1.PublicationStatus_PUBLICATION_STATUS_UNSPECIFIED {
		st := domain.PublicationStatus(req.Msg.Status)
		filter.Status = &st
	}

	submissions, nextCursor, total, err := s.submissionUC.List(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "list submissions failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	protoSubs := make([]*emhv1.Submission, 0, len(submissions))
	for _, sub := range submissions {
		protoSubs = append(protoSubs, mapSubmissionToProto(sub))
	}

	return connect.NewResponse(&emhv1.ListSubmissionsResponse{
		Submissions: protoSubs,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// GetSubmission возвращает заявку по ID.
func (s *SubmissionAdminServer) GetSubmission(
	ctx context.Context,
	req *connect.Request[emhv1.GetSubmissionRequest],
) (*connect.Response[emhv1.GetSubmissionResponse], error) {
	sub, err := s.submissionUC.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get submission failed")
	}

	return connect.NewResponse(&emhv1.GetSubmissionResponse{
		Submission: mapSubmissionToProto(sub),
	}), nil
}

// ReviewSubmission одобряет или отклоняет заявку.
func (s *SubmissionAdminServer) ReviewSubmission(
	ctx context.Context,
	req *connect.Request[emhv1.ReviewSubmissionRequest],
) (*connect.Response[emhv1.ReviewSubmissionResponse], error) {
	s.logger.InfoContext(ctx, "reviewing submission",
		"id", req.Msg.Id,
		"decision", req.Msg.Decision.String(),
	)

	params := domain.ReviewSubmissionParams{
		ID:               req.Msg.Id,
		Decision:         domain.ReviewDecision(req.Msg.Decision),
		ModeratorComment: req.Msg.ModeratorComment,
	}

	sub, err := s.submissionUC.Review(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "review submission failed")
	}

	return connect.NewResponse(&emhv1.ReviewSubmissionResponse{
		Submission: mapSubmissionToProto(sub),
	}), nil
}

// ListSubmissionReviews возвращает историю модерации заявки.
func (s *SubmissionAdminServer) ListSubmissionReviews(
	ctx context.Context,
	req *connect.Request[emhv1.ListSubmissionReviewsRequest],
) (*connect.Response[emhv1.ListSubmissionReviewsResponse], error) {
	reviews, err := s.submissionUC.ListReviews(ctx, req.Msg.SubmissionId)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list submission reviews failed")
	}

	protoReviews := make([]*emhv1.SubmissionReview, 0, len(reviews))
	for _, r := range reviews {
		protoReviews = append(protoReviews, &emhv1.SubmissionReview{
			Id:           r.ID,
			SubmissionId: r.SubmissionID,
			ReviewerName: r.ReviewerName,
			Decision:     emhv1.SubmissionReviewDecision(r.Decision),
			Comment:      r.Comment,
			CreatedAt:    timestamppb.New(r.CreatedAt),
		})
	}

	return connect.NewResponse(&emhv1.ListSubmissionReviewsResponse{
		Reviews: protoReviews,
	}), nil
}
