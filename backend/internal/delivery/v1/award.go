package v1

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

type AwardServer struct {
	awardUC usecase.AwardUseCase
	logger  *slog.Logger
}

func NewAwardServer(uc usecase.AwardUseCase, logger *slog.Logger) *AwardServer {
	return &AwardServer{awardUC: uc, logger: logger}
}

func (s *AwardServer) GetAward(
	ctx context.Context,
	req *connect.Request[emhv1.GetAwardRequest],
) (*connect.Response[emhv1.GetAwardResponse], error) {
	award, err := s.awardUC.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get award failed")
	}

	return connect.NewResponse(&emhv1.GetAwardResponse{
		Award: mapAwardToProto(award),
	}), nil
}

func (s *AwardServer) ListAwards(
	ctx context.Context,
	req *connect.Request[emhv1.ListAwardsRequest],
) (*connect.Response[emhv1.ListAwardsResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := domain.AwardFilter{
		SearchQuery: req.Msg.SearchQuery,
		Cursor:      req.Msg.Pagination.GetCursor(),
		Limit:       limit,
	}

	awards, nextCursor, total, err := s.awardUC.List(ctx, filter)
	if err != nil {
		s.logger.ErrorContext(ctx, "list awards failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}

	protoAwards := make([]*emhv1.Award, 0, len(awards))
	for _, a := range awards {
		protoAwards = append(protoAwards, mapAwardToProto(a))
	}

	return connect.NewResponse(&emhv1.ListAwardsResponse{
		Awards: protoAwards,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

func mapAwardToProto(a *domain.Award) *emhv1.Award {
	return &emhv1.Award{
		Id:             a.ID,
		Name:           a.Name,
		Description:    a.Description,
		ImageUrl:       a.ImageURL,
		SortOrder:      int32(a.SortOrder),
		RibbonImageUrl: a.RibbonImageURL,
		Type:           emhv1.AwardType(a.Type),
		WornWithoutBar: a.WornWithoutBar,
		IsJubilee:      a.IsJubilee,
		Jurisdiction:   emhv1.AwardJurisdiction(a.Jurisdiction),
	}
}
