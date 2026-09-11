package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

type AwardAdminServer struct {
	awardUC usecase.AwardUseCase
	logger  *slog.Logger
}

func NewAwardAdminServer(uc usecase.AwardUseCase, logger *slog.Logger) *AwardAdminServer {
	return &AwardAdminServer{awardUC: uc, logger: logger}
}

func (s *AwardAdminServer) CreateAward(
	ctx context.Context,
	req *connect.Request[emhv1.CreateAwardRequest],
) (*connect.Response[emhv1.CreateAwardResponse], error) {
	s.logger.InfoContext(ctx, "creating award", "name", req.Msg.Name)

	params := domain.CreateAwardParams{
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
		ImageURL:    req.Msg.ImageUrl,
		SortOrder:   int(req.Msg.SortOrder),
	}

	id, err := s.awardUC.Create(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create award failed")
	}

	return connect.NewResponse(&emhv1.CreateAwardResponse{Id: id}), nil
}

func (s *AwardAdminServer) UpdateAward(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateAwardRequest],
) (*connect.Response[emhv1.UpdateAwardResponse], error) {
	s.logger.InfoContext(ctx, "updating award", "id", req.Msg.Id)

	params := domain.UpdateAwardParams{
		ID:          req.Msg.Id,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
		ImageURL:    req.Msg.ImageUrl,
		SortOrder:   int(req.Msg.SortOrder),
		FieldMask:   req.Msg.FieldMask,
	}

	award, err := s.awardUC.Update(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "update award failed")
	}

	return connect.NewResponse(&emhv1.UpdateAwardResponse{
		Award: mapAwardToProto(award),
	}), nil
}

func (s *AwardAdminServer) DeleteAward(
	ctx context.Context,
	req *connect.Request[emhv1.DeleteAwardRequest],
) (*connect.Response[emhv1.DeleteAwardResponse], error) {
	s.logger.InfoContext(ctx, "deleting award", "id", req.Msg.Id)

	if err := s.awardUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "delete award failed")
	}

	return connect.NewResponse(&emhv1.DeleteAwardResponse{Success: true}), nil
}
