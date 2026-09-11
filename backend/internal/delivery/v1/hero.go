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

type HeroServer struct {
	heroUC         usecase.HeroUseCase
	heroQueryUC    usecase.HeroQueryUseCase
	heroSourceUC   usecase.HeroSourceUseCase
	heroRelationUC usecase.HeroRelationUseCase
	logger         *slog.Logger
}

func NewHeroServer(
	uc usecase.HeroUseCase,
	queryUC usecase.HeroQueryUseCase,
	sourceUC usecase.HeroSourceUseCase,
	relationUC usecase.HeroRelationUseCase,
	logger *slog.Logger,
) *HeroServer {
	return &HeroServer{
		heroUC:         uc,
		heroQueryUC:    queryUC,
		heroSourceUC:   sourceUC,
		heroRelationUC: relationUC,
		logger:         logger,
	}
}

// GetHero возвращает полную карточку героя со всеми связями.
func (s *HeroServer) GetHero(
	ctx context.Context,
	req *connect.Request[emhv1.GetHeroRequest],
) (*connect.Response[emhv1.GetHeroResponse], error) {
	s.logger.InfoContext(ctx, "getting hero detail", "id", req.Msg.Id)

	detail, err := s.heroQueryUC.GetHeroDetail(ctx, req.Msg.Id)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "get hero detail failed")
	}

	return connect.NewResponse(&emhv1.GetHeroResponse{
		Hero: mapHeroDetailToProto(detail),
	}), nil
}

// ListHeroes — без изменений (использует heroUC).
func (s *HeroServer) ListHeroes(
	ctx context.Context,
	req *connect.Request[emhv1.ListHeroesRequest],
) (*connect.Response[emhv1.ListHeroesResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var dateFrom, dateTo *string
	if req.Msg.DateFrom != nil {
		v := req.Msg.DateFrom.AsTime().Format("2006-01-02")
		dateFrom = &v
	}
	if req.Msg.DateTo != nil {
		v := req.Msg.DateTo.AsTime().Format("2006-01-02")
		dateTo = &v
	}

	filter := domain.HeroFilter{
		SearchQuery: req.Msg.SearchQuery,
		ConflictID:  req.Msg.ConflictId,
		LocationID:  req.Msg.LocationId,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Cursor:      req.Msg.Pagination.GetCursor(),
		Limit:       limit,
	}

	// 🔒 Явно запрашиваем только опубликованных героев для публичного API
	publishedStatus := domain.StatusPublished
	filter.Status = &publishedStatus

	heroes, nextCursor, total, err := s.heroUC.ListHeroes(ctx, filter)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list heroes failed")
	}

	protoHeroes := make([]*emhv1.HeroSummary, 0, len(heroes))
	for _, h := range heroes {
		protoHeroes = append(protoHeroes, mapHeroToSummary(h))
	}

	return connect.NewResponse(&emhv1.ListHeroesResponse{
		Heroes: protoHeroes,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// --- Мапперы domain -> proto ---

func mapHeroDetailToProto(d *domain.HeroDetail) *emhv1.HeroDetail {
	h := d.Hero

	summary := mapHeroToSummary(h)
	summary.MainPhotoUrl = d.MainPhotoURL()
	summary.AwardNames = d.AwardNames()

	detail := &emhv1.HeroDetail{
		Summary: summary,
		FullBio: h.FullBio,
		Status:  emhv1.PublicationStatus(h.Status),
		Audit: &emhv1.AuditInfo{
			CreatedAt: timestamppb.New(h.CreatedAt),
			UpdatedAt: timestamppb.New(h.UpdatedAt),
		},
	}

	detail.Photos = make([]*emhv1.Photo, 0, len(d.Photos))
	for _, p := range d.Photos {
		detail.Photos = append(detail.Photos, mapPhotoToProto(p))
	}

	detail.Awards = make([]*emhv1.HeroAward, 0, len(d.Awards))
	for _, a := range d.Awards {
		detail.Awards = append(detail.Awards, mapHeroAwardToProto(a))
	}

	detail.Conflicts = make([]*emhv1.HeroConflict, 0, len(d.Conflicts))
	for _, c := range d.Conflicts {
		detail.Conflicts = append(detail.Conflicts, mapHeroConflictToProto(c))
	}

	detail.Locations = make([]*emhv1.HeroLocation, 0, len(d.Locations))
	for _, l := range d.Locations {
		detail.Locations = append(detail.Locations, mapHeroLocationToProto(l))
	}

	detail.Sources = make([]*emhv1.HeroSource, 0, len(d.Sources))
	for _, src := range d.Sources {
		detail.Sources = append(detail.Sources, mapHeroSourceToProto(src))
	}

	detail.Relations = make([]*emhv1.HeroRelation, 0, len(d.Relations))
	for _, rel := range d.Relations {
		detail.Relations = append(detail.Relations, mapHeroRelationToProto(rel))
	}

	detail.Position = h.Position
	detail.CauseOfDeath = h.CauseOfDeath
	detail.Memberships = h.Memberships
	if h.ServiceStartDate.Anchor != nil {
		detail.ServiceStartDate = timestamppb.New(*h.ServiceStartDate.Anchor)
	}

	detail.ServiceStartDateInfo = mapFlexibleDateToProto(h.ServiceStartDate)

	return detail
}

func mapHeroToSummary(h *domain.Hero) *emhv1.HeroSummary {
	summary := &emhv1.HeroSummary{
		Id:            h.ID,
		FirstName:     h.FirstName,
		LastName:      h.LastName,
		MiddleName:    h.MiddleName,
		Rank:          h.Rank,
		ShortBio:      h.ShortBio,
		AwardNames:    h.AwardNames,
		Nickname:      h.Nickname,
		Unit:          h.Unit,
		ServiceBranch: h.ServiceBranch,
		Status:        emhv1.PublicationStatus(h.Status),
	}

	if h.MainPhotoURL != nil {
		summary.MainPhotoUrl = *h.MainPhotoURL
	}

	if h.MainThumbnailURL != nil {
		summary.MainThumbnailUrl = *h.MainThumbnailURL
	}

	// Старые поля для обратной совместимости.
	if h.BirthDate.Anchor != nil {
		summary.BirthDate = timestamppb.New(*h.BirthDate.Anchor)
	}

	if h.DeathDate.Anchor != nil {
		summary.DeathDate = timestamppb.New(*h.DeathDate.Anchor)
	}

	// Новые гибкие даты.
	summary.BirthDateInfo = mapFlexibleDateToProto(h.BirthDate)
	summary.DeathDateInfo = mapFlexibleDateToProto(h.DeathDate)

	return summary
}

func mapPhotoToProto(p *domain.Photo) *emhv1.Photo {
	return &emhv1.Photo{
		Id:           p.ID,
		Url:          p.URL,
		ThumbnailUrl: p.ThumbnailURL,
		Description:  p.Description,
		SortOrder:    int32(p.SortOrder),
		IsMain:       p.IsMain,
		FaceBox:      mapFaceBoxToProto(p.FaceBox),
	}
}

// mapHeroSourceToProto маппит источник данных героя в proto.
func mapHeroSourceToProto(s *domain.HeroSource) *emhv1.HeroSource {
	return &emhv1.HeroSource{
		Id:         s.ID,
		Url:        s.URL,
		Title:      s.Title,
		SourceType: s.SourceType,
		Excerpt:    s.Excerpt,
	}
}

// mapHeroRelationToProto маппит связь героя с другим героем в proto.
func mapHeroRelationToProto(r *domain.HeroRelation) *emhv1.HeroRelation {
	return &emhv1.HeroRelation{
		Id:              r.ID,
		FromHeroId:      r.FromHeroID,
		ToHeroId:        r.ToHeroID,
		RelationType:    r.RelationType,
		Description:     r.Description,
		RelatedHeroName: r.RelatedHeroName,
	}
}

func mapHeroAwardToProto(a *domain.HeroAward) *emhv1.HeroAward {
	proto := &emhv1.HeroAward{
		AwardId:      a.AwardID,
		AwardName:    a.AwardName,
		DecreeNumber: a.DecreeNumber,
	}
	if a.AwardDate != nil {
		proto.AwardDate = timestamppb.New(*a.AwardDate)
	}
	return proto
}

func mapHeroConflictToProto(c *domain.HeroConflict) *emhv1.HeroConflict {
	return &emhv1.HeroConflict{
		ConflictId:       c.ConflictID,
		ConflictName:     c.ConflictName,
		SpecificLocation: c.SpecificLocation,
		RankAtConflict:   c.RankAtConflict,
	}
}

func mapHeroLocationToProto(hl *domain.HeroLocation) *emhv1.HeroLocation {
	return &emhv1.HeroLocation{
		LocationId: hl.LocationID,
		Type:       emhv1.HeroLocationType(hl.Type),
		Location:   mapLocationToProto(hl.Location), // переиспользование
	}
}

// ListHeroPhotos возвращает фотографии героя.
//
// Поддерживает два режима:
//   - Без пагинации (pagination == nil или пустой): возвращает ВСЕ фото (legacy-режим).
//     Используется для внутренней сборки HeroDetail через GetHero.
//   - С пагинацией: курсорная пагинация через ListHeroPhotosPaged.
//     Используется клиентами для галерей с большим числом фото.
func (s *HeroServer) ListHeroPhotos(
	ctx context.Context,
	req *connect.Request[emhv1.ListHeroPhotosRequest],
) (*connect.Response[emhv1.ListHeroPhotosResponse], error) {
	pag := req.Msg.Pagination
	hasCursor := pag != nil && pag.GetCursor() != ""
	hasPageSize := pag != nil && pag.GetPageSize() > 0

	// Legacy-режим: без пагинации возвращаем все фото (для совместимости).
	// Используется внутри GetHero через HeroQueryUseCase.
	if !hasCursor && !hasPageSize {
		photos, err := s.heroQueryUC.ListHeroPhotos(ctx, req.Msg.HeroId)
		if err != nil {
			return nil, mapDomainError(ctx, s.logger, err, "list hero photos failed")
		}
		protoPhotos := make([]*emhv1.Photo, 0, len(photos))
		for _, p := range photos {
			protoPhotos = append(protoPhotos, mapPhotoToProto(p))
		}
		return connect.NewResponse(&emhv1.ListHeroPhotosResponse{
			Photos: protoPhotos,
			Pagination: &emhv1.PaginationResponse{
				NextCursor: "",
				TotalCount: int64(len(photos)),
			},
		}), nil
	}

	// Пагинированный режим
	limit := int(pag.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	photos, nextCursor, total, err := s.heroQueryUC.ListHeroPhotosPaged(
		ctx, req.Msg.HeroId, pag.GetCursor(), limit,
	)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list hero photos paged failed")
	}

	protoPhotos := make([]*emhv1.Photo, 0, len(photos))
	for _, p := range photos {
		protoPhotos = append(protoPhotos, mapPhotoToProto(p))
	}

	return connect.NewResponse(&emhv1.ListHeroPhotosResponse{
		Photos: protoPhotos,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// ListHeroSources возвращает все источники данных героя.
func (s *HeroServer) ListHeroSources(
	ctx context.Context,
	req *connect.Request[emhv1.ListHeroSourcesRequest],
) (*connect.Response[emhv1.ListHeroSourcesResponse], error) {
	sources, err := s.heroSourceUC.ListByHero(ctx, req.Msg.HeroId)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list hero sources failed")
	}

	protoSources := make([]*emhv1.HeroSource, 0, len(sources))
	for _, src := range sources {
		protoSources = append(protoSources, mapHeroSourceToProto(src))
	}

	return connect.NewResponse(&emhv1.ListHeroSourcesResponse{
		Sources: protoSources,
	}), nil
}

// ListHeroRelations возвращает все связи героя с другими героями.
func (s *HeroServer) ListHeroRelations(
	ctx context.Context,
	req *connect.Request[emhv1.ListHeroRelationsRequest],
) (*connect.Response[emhv1.ListHeroRelationsResponse], error) {
	relations, err := s.heroRelationUC.ListByHero(ctx, req.Msg.HeroId)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list hero relations failed")
	}

	protoRelations := make([]*emhv1.HeroRelation, 0, len(relations))
	for _, rel := range relations {
		protoRelations = append(protoRelations, mapHeroRelationToProto(rel))
	}

	return connect.NewResponse(&emhv1.ListHeroRelationsResponse{
		Relations: protoRelations,
	}), nil
}
