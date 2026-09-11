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

// ConflictServer реализует публичный emhv1.ConflictServiceHandler.
type ConflictServer struct {
	conflictUC usecase.ConflictUseCase
	logger     *slog.Logger
}

func NewConflictServer(uc usecase.ConflictUseCase, logger *slog.Logger) *ConflictServer {
	return &ConflictServer{conflictUC: uc, logger: logger}
}

// GetConflict возвращает конфликт по ID.
func (s *ConflictServer) GetConflict(
	ctx context.Context,
	req *connect.Request[emhv1.GetConflictRequest],
) (*connect.Response[emhv1.GetConflictResponse], error) {
	conflict, err := s.conflictUC.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get conflict failed")
	}
	return connect.NewResponse(&emhv1.GetConflictResponse{
		Conflict: mapConflictToProto(conflict),
	}), nil
}

// ListConflicts возвращает отфильтрованный список конфликтов.
func (s *ConflictServer) ListConflicts(
	ctx context.Context,
	req *connect.Request[emhv1.ListConflictsRequest],
) (*connect.Response[emhv1.ListConflictsResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := domain.ConflictFilter{
		Type:     domain.ConflictType(req.Msg.Type),
		ParentID: req.Msg.ParentConflictId,
		Cursor:   req.Msg.Pagination.GetCursor(),
		Limit:    limit,
	}

	conflicts, nextCursor, total, err := s.conflictUC.List(ctx, filter)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list conflicts failed")
	}

	protoConflicts := make([]*emhv1.Conflict, 0, len(conflicts))
	for _, c := range conflicts {
		protoConflicts = append(protoConflicts, mapConflictToProto(c))
	}

	return connect.NewResponse(&emhv1.ListConflictsResponse{
		Conflicts: protoConflicts,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// mapConflictToProto маппит доменную модель в proto-сообщение.
func mapConflictToProto(c *domain.Conflict) *emhv1.Conflict {
	proto := &emhv1.Conflict{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Type:        emhv1.ConflictType(c.Type),
	}
	if c.StartDate != nil {
		proto.StartDate = timestamppb.New(*c.StartDate)
	}
	if c.EndDate != nil {
		proto.EndDate = timestamppb.New(*c.EndDate)
	}
	if c.ParentConflictID != nil {
		proto.ParentConflictId = *c.ParentConflictID
	}

	// Маппинг гибких дат
	if c.StartDateInfo != nil {
		proto.StartDateInfo = mapFlexibleDateToProto(*c.StartDateInfo)
	}
	if c.EndDateInfo != nil {
		proto.EndDateInfo = mapFlexibleDateToProto(*c.EndDateInfo)
	}

	return proto
}
