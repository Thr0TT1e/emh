package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// SubmissionServer реализует публичный emhv1.SubmissionServiceHandler.
type SubmissionServer struct {
	submissionUC usecase.SubmissionUseCase
	logger       *slog.Logger
}

func NewSubmissionServer(uc usecase.SubmissionUseCase, logger *slog.Logger) *SubmissionServer {
	return &SubmissionServer{submissionUC: uc, logger: logger}
}

// CreateSubmission создаёт заявку от посетителя сайта.
func (s *SubmissionServer) CreateSubmission(
	ctx context.Context,
	req *connect.Request[emhv1.CreateSubmissionRequest],
) (*connect.Response[emhv1.CreateSubmissionResponse], error) {
	s.logger.InfoContext(ctx, "creating submission",
		"submitter", req.Msg.SubmitterName,
		"target_hero", req.Msg.TargetHeroId,
	)

	params := domain.CreateSubmissionParams{
		SubmitterName:  req.Msg.SubmitterName,
		SubmitterEmail: req.Msg.SubmitterEmail,
		PayloadJSON:    req.Msg.PayloadJson,
		AttachmentURLs: req.Msg.AttachmentUrls,
	}
	if req.Msg.TargetHeroId != "" {
		params.TargetHeroID = &req.Msg.TargetHeroId
	}

	id, err := s.submissionUC.Create(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create submission failed")
	}

	return connect.NewResponse(&emhv1.CreateSubmissionResponse{SubmissionId: id}), nil
}

// mapSubmissionToProto маппит доменную модель заявки в proto-сообщение.
func mapSubmissionToProto(sub *domain.Submission) *emhv1.Submission {
	proto := &emhv1.Submission{
		Id:             sub.ID,
		SubmitterName:  sub.SubmitterName,
		SubmitterEmail: sub.SubmitterEmail,
		PayloadJson:    sub.PayloadJSON,
		AttachmentUrls: sub.AttachmentURLs,
		Status:         emhv1.PublicationStatus(sub.Status),
		Audit: &emhv1.AuditInfo{
			CreatedAt: timestamppb.New(sub.CreatedAt),
			UpdatedAt: timestamppb.New(sub.UpdatedAt),
		},
		ContentHash: sub.ContentHash,
	}
	if sub.TargetHeroID != nil {
		proto.TargetHeroId = *sub.TargetHeroID
	}
	if sub.ModeratorComment != nil {
		proto.ModeratorComment = *sub.ModeratorComment
	}

	return proto
}
