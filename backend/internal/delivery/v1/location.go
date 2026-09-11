package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// LocationServer реализует публичный emhv1.LocationServiceHandler.
type LocationServer struct {
	locationUC usecase.LocationUseCase
	logger     *slog.Logger
}

func NewLocationServer(uc usecase.LocationUseCase, logger *slog.Logger) *LocationServer {
	return &LocationServer{locationUC: uc, logger: logger}
}

// GetLocation возвращает локацию по ID.
func (s *LocationServer) GetLocation(
	ctx context.Context,
	req *connect.Request[emhv1.GetLocationRequest],
) (*connect.Response[emhv1.GetLocationResponse], error) {
	location, err := s.locationUC.GetByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get location failed")
	}

	return connect.NewResponse(&emhv1.GetLocationResponse{
		Location: mapLocationToProto(location),
	}), nil
}

// ListLocations возвращает отфильтрованный список локаций.
func (s *LocationServer) ListLocations(
	ctx context.Context,
	req *connect.Request[emhv1.ListLocationsRequest],
) (*connect.Response[emhv1.ListLocationsResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := domain.LocationFilter{
		ParentID:    req.Msg.ParentId,
		SearchQuery: req.Msg.SearchQuery,
		Type:        domain.LocationType(req.Msg.Type),
		Cursor:      req.Msg.Pagination.GetCursor(),
		Limit:       limit,
	}

	locations, nextCursor, total, err := s.locationUC.List(ctx, filter)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list locations failed")
	}

	protoLocations := make([]*emhv1.Location, 0, len(locations))
	for _, l := range locations {
		protoLocations = append(protoLocations, mapLocationToProto(l))
	}

	return connect.NewResponse(&emhv1.ListLocationsResponse{
		Locations: protoLocations,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// mapLocationToProto маппит доменную модель локации в proto-сообщение.
func mapLocationToProto(l *domain.Location) *emhv1.Location {
	proto := &emhv1.Location{
		Id:             l.ID,
		Name:           l.Name,
		HistoricalName: l.HistoricalName,
		Type:           emhv1.LocationType(l.Type),
	}
	if l.ParentID != nil {
		proto.ParentId = *l.ParentID
	}
	if l.Latitude != nil {
		proto.Latitude = *l.Latitude
	}
	if l.Longitude != nil {
		proto.Longitude = *l.Longitude
	}
	return proto
}
