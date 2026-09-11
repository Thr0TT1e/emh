package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// LocationAdminServer реализует защищённый emhv1.LocationAdminServiceHandler.
type LocationAdminServer struct {
	locationUC usecase.LocationUseCase
	logger     *slog.Logger
}

func NewLocationAdminServer(uc usecase.LocationUseCase, logger *slog.Logger) *LocationAdminServer {
	return &LocationAdminServer{locationUC: uc, logger: logger}
}

// CreateLocation создаёт новую локацию.
func (s *LocationAdminServer) CreateLocation(
	ctx context.Context,
	req *connect.Request[emhv1.CreateLocationRequest],
) (*connect.Response[emhv1.CreateLocationResponse], error) {
	s.logger.InfoContext(ctx, "creating location", "name", req.Msg.Name)

	params := domain.CreateLocationParams{
		Name:           req.Msg.Name,
		HistoricalName: req.Msg.HistoricalName,
		Type:           domain.LocationType(req.Msg.Type),
	}
	if req.Msg.ParentId != "" {
		params.ParentID = &req.Msg.ParentId
	}
	// В proto3 скалярный double не отличает "не задано" от 0.
	// Интерпретируем 0 как отсутствие координаты (см. примечание ниже).
	if req.Msg.Latitude != 0 {
		params.Latitude = &req.Msg.Latitude
	}
	if req.Msg.Longitude != 0 {
		params.Longitude = &req.Msg.Longitude
	}

	id, err := s.locationUC.Create(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "create location failed")
	}

	return connect.NewResponse(&emhv1.CreateLocationResponse{Id: id}), nil
}

// UpdateLocation частично обновляет локацию по field_mask.
func (s *LocationAdminServer) UpdateLocation(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateLocationRequest],
) (*connect.Response[emhv1.UpdateLocationResponse], error) {
	s.logger.InfoContext(ctx, "updating location", "id", req.Msg.Id)

	params := domain.UpdateLocationParams{
		ID:             req.Msg.Id,
		Name:           req.Msg.Name,
		HistoricalName: req.Msg.HistoricalName,
		Type:           domain.LocationType(req.Msg.Type),
		Latitude:       &req.Msg.Latitude,
		Longitude:      &req.Msg.Longitude,
		FieldMask:      req.Msg.FieldMask,
	}
	if req.Msg.ParentId != "" {
		params.ParentID = &req.Msg.ParentId
	}

	location, err := s.locationUC.Update(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "update location failed")
	}

	return connect.NewResponse(&emhv1.UpdateLocationResponse{
		Location: mapLocationToProto(location),
	}), nil
}

// DeleteLocation удаляет локацию, если к ней не привязаны герои.
func (s *LocationAdminServer) DeleteLocation(
	ctx context.Context,
	req *connect.Request[emhv1.DeleteLocationRequest],
) (*connect.Response[emhv1.DeleteLocationResponse], error) {
	if err := s.locationUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "delete location failed")
	}

	return connect.NewResponse(&emhv1.DeleteLocationResponse{Success: true}), nil
}

// ReparentLocations массово переносит дочерние локации к новому родителю.
func (s *LocationAdminServer) ReparentLocations(
	ctx context.Context,
	req *connect.Request[emhv1.ReparentLocationsRequest],
) (*connect.Response[emhv1.ReparentLocationsResponse], error) {
	affected, err := s.locationUC.Reparent(ctx, req.Msg.OldParentId, req.Msg.NewParentId)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "reparent locations failed")
	}

	return connect.NewResponse(&emhv1.ReparentLocationsResponse{
		AffectedCount: int32(affected),
	}), nil
}
