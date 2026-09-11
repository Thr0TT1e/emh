package v1

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// ConflictAdminServer реализует защищённый emhv1.ConflictAdminServiceHandler.
type ConflictAdminServer struct {
	conflictUC usecase.ConflictUseCase
	logger     *slog.Logger
}

func NewConflictAdminServer(uc usecase.ConflictUseCase, logger *slog.Logger) *ConflictAdminServer {
	return &ConflictAdminServer{conflictUC: uc, logger: logger}
}

// CreateConflict создаёт новый конфликт.
func (s *ConflictAdminServer) CreateConflict(
	ctx context.Context,
	req *connect.Request[emhv1.CreateConflictRequest],
) (*connect.Response[emhv1.CreateConflictResponse], error) {
	s.logger.InfoContext(ctx, "creating conflict", "name", req.Msg.Name)

	start, err := usecase.ParseDateStr(req.Msg.StartDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start_date: %w", err))
	}
	end, err := usecase.ParseDateStr(req.Msg.EndDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end_date: %w", err))
	}

	// Парсинг гибких дат
	var startDateInfo, endDateInfo *domain.FlexibleDate
	if req.Msg.StartDateInfo != nil {
		d, err := mapProtoFlexibleDate(req.Msg.StartDateInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start_date_info: %w", err))
		}
		startDateInfo = d
	}
	if req.Msg.EndDateInfo != nil {
		d, err := mapProtoFlexibleDate(req.Msg.EndDateInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end_date_info: %w", err))
		}
		endDateInfo = d
	}

	params := domain.CreateConflictParams{
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		Type:          domain.ConflictType(req.Msg.Type),
		StartDate:     start,
		EndDate:       end,
		StartDateInfo: startDateInfo,
		EndDateInfo:   endDateInfo,
	}
	if req.Msg.ParentConflictId != "" {
		params.ParentConflictID = &req.Msg.ParentConflictId
	}

	id, err := s.conflictUC.Create(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create conflict failed")
	}

	return connect.NewResponse(&emhv1.CreateConflictResponse{Id: id}), nil
}

// UpdateConflict частично обновляет конфликт по field_mask.
func (s *ConflictAdminServer) UpdateConflict(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateConflictRequest],
) (*connect.Response[emhv1.UpdateConflictResponse], error) {
	s.logger.InfoContext(ctx, "updating conflict", "id", req.Msg.Id)

	start, err := usecase.ParseDateStr(req.Msg.StartDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start_date: %w", err))
	}
	end, err := usecase.ParseDateStr(req.Msg.EndDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end_date: %w", err))
	}

	// Парсинг гибких дат
	var startDateInfo, endDateInfo *domain.FlexibleDate
	if req.Msg.StartDateInfo != nil {
		d, err := mapProtoFlexibleDate(req.Msg.StartDateInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start_date_info: %w", err))
		}
		startDateInfo = d
	}
	if req.Msg.EndDateInfo != nil {
		d, err := mapProtoFlexibleDate(req.Msg.EndDateInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end_date_info: %w", err))
		}
		endDateInfo = d
	}

	params := domain.UpdateConflictParams{
		ID:            req.Msg.Id,
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		Type:          domain.ConflictType(req.Msg.Type),
		StartDate:     start,
		EndDate:       end,
		FieldMask:     req.Msg.FieldMask,
		StartDateInfo: startDateInfo,
		EndDateInfo:   endDateInfo,
	}
	if req.Msg.ParentConflictId != "" {
		params.ParentConflictID = &req.Msg.ParentConflictId
	}

	conflict, err := s.conflictUC.Update(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "update conflict failed")
	}

	return connect.NewResponse(&emhv1.UpdateConflictResponse{
		Conflict: mapConflictToProto(conflict),
	}), nil
}

// DeleteConflict удаляет конфликт, если к нему не привязаны герои.
func (s *ConflictAdminServer) DeleteConflict(
	ctx context.Context,
	req *connect.Request[emhv1.DeleteConflictRequest],
) (*connect.Response[emhv1.DeleteConflictResponse], error) {
	s.logger.InfoContext(ctx, "deleting conflict", "id", req.Msg.Id)

	if err := s.conflictUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "delete conflict failed")
	}

	return connect.NewResponse(&emhv1.DeleteConflictResponse{Success: true}), nil
}
