// Реализация HeroAdminService

package v1

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
	"connectrpc.com/connect"
)

// HeroAdminServer реализует emhv1.HeroAdminServiceHandler.
type HeroAdminServer struct {
	heroUC         usecase.HeroUseCase
	heroQueryUC    usecase.HeroQueryUseCase
	heroAwardUC    usecase.HeroAwardUseCase
	heroConflictUC usecase.HeroConflictUseCase
	heroLocationUC usecase.HeroLocationUseCase
	heroSourceUC   usecase.HeroSourceUseCase
	heroRelationUC usecase.HeroRelationUseCase
	photoUC        usecase.PhotoUseCase
	logger         *slog.Logger
}

// NewHeroAdminServer создает новый экземпляр сервера.
func NewHeroAdminServer(
	uc usecase.HeroUseCase,
	queryUC usecase.HeroQueryUseCase,
	awardUC usecase.HeroAwardUseCase,
	conflictUC usecase.HeroConflictUseCase,
	locationUC usecase.HeroLocationUseCase,
	sourceUC usecase.HeroSourceUseCase,
	relationUC usecase.HeroRelationUseCase,
	photoUC usecase.PhotoUseCase,
	logger *slog.Logger) *HeroAdminServer {
	return &HeroAdminServer{
		heroUC:         uc,
		heroQueryUC:    queryUC,
		heroAwardUC:    awardUC,
		heroConflictUC: conflictUC,
		heroLocationUC: locationUC,
		heroSourceUC:   sourceUC,
		heroRelationUC: relationUC,
		photoUC:        photoUC,
		logger:         logger,
	}
}

// CreateHero обрабатывает создание нового героя.
func (s *HeroAdminServer) CreateHero(
	ctx context.Context,
	req *connect.Request[emhv1.CreateHeroRequest],
) (*connect.Response[emhv1.CreateHeroResponse], error) {
	s.logger.InfoContext(ctx, "creating hero",
		"name", req.Msg.LastName+" "+req.Msg.FirstName,
	)

	birthDate, err := resolveCreateFlexibleDate(req.Msg.BirthDateInfo, req.Msg.BirthDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid birth_date: %w", err),
		)
	}

	deathDate, err := resolveCreateFlexibleDate(req.Msg.DeathDateInfo, req.Msg.DeathDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid death_date: %w", err),
		)
	}

	serviceStartDate, err := resolveCreateFlexibleDate(req.Msg.ServiceStartDateInfo, req.Msg.ServiceStartDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid service_start_date: %w", err),
		)
	}

	params := domain.CreateHeroParams{
		FirstName:        req.Msg.FirstName,
		LastName:         req.Msg.LastName,
		MiddleName:       req.Msg.MiddleName,
		ShortBio:         req.Msg.ShortBio,
		FullBio:          req.Msg.FullBio,
		Rank:             req.Msg.Rank,
		BirthDate:        birthDate,
		DeathDate:        deathDate,
		Status:           domain.PublicationStatus(req.Msg.Status),
		Nickname:         req.Msg.Nickname,
		Unit:             req.Msg.Unit,
		Position:         req.Msg.Position,
		ServiceBranch:    req.Msg.ServiceBranch,
		CauseOfDeath:     req.Msg.CauseOfDeath,
		ServiceStartDate: serviceStartDate,
		Memberships:      req.Msg.Memberships,
	}

	id, err := s.heroUC.CreateHero(ctx, params)
	if err != nil {
		return nil, s.mapUseCaseError(ctx, err, "create hero failed")
	}

	return connect.NewResponse(&emhv1.CreateHeroResponse{Id: id}), nil
}

// UpdateHero обрабатывает частичное обновление героя.
func (s *HeroAdminServer) UpdateHero(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateHeroRequest],
) (*connect.Response[emhv1.UpdateHeroResponse], error) {
	s.logger.InfoContext(ctx, "updating hero", "id", req.Msg.Id)

	birthDate, err := resolveUpdateFlexibleDate(req.Msg.BirthDateInfo, req.Msg.BirthDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid birth_date: %w", err),
		)
	}

	deathDate, err := resolveUpdateFlexibleDate(req.Msg.DeathDateInfo, req.Msg.DeathDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid death_date: %w", err),
		)
	}

	serviceStartDate, err := resolveUpdateFlexibleDate(req.Msg.ServiceStartDateInfo, req.Msg.ServiceStartDate)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("invalid service_start_date: %w", err),
		)
	}

	var statusPtr *domain.PublicationStatus
	if req.Msg.Status != 0 {
		st := domain.PublicationStatus(req.Msg.Status)
		statusPtr = &st
	}

	id := req.Msg.Id

	params := domain.UpdateHeroParams{
		ID:               &id,
		FirstName:        stringPtr(req.Msg.FirstName),
		LastName:         stringPtr(req.Msg.LastName),
		MiddleName:       stringPtr(req.Msg.MiddleName),
		ShortBio:         stringPtr(req.Msg.ShortBio),
		FullBio:          stringPtr(req.Msg.FullBio),
		Rank:             stringPtr(req.Msg.Rank),
		BirthDate:        birthDate,
		DeathDate:        deathDate,
		Status:           statusPtr,
		FieldMask:        req.Msg.FieldMask,
		Nickname:         stringPtr(req.Msg.Nickname),
		Unit:             stringPtr(req.Msg.Unit),
		Position:         stringPtr(req.Msg.Position),
		ServiceBranch:    stringPtr(req.Msg.ServiceBranch),
		CauseOfDeath:     stringPtr(req.Msg.CauseOfDeath),
		ServiceStartDate: serviceStartDate,
		Memberships:      req.Msg.Memberships,
	}

	updated, err := s.heroUC.UpdateHero(ctx, params)
	if err != nil {
		return nil, s.mapUseCaseError(ctx, err, "update hero failed")
	}

	// Загружаем полный агрегат для возврата в ответе.
	detail, err := s.heroQueryUC.GetHeroDetail(ctx, req.Msg.Id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load hero detail after update", "error", err)

		// Fallback: возвращаем хотя бы базовую карточку из обновлённого героя.
		return connect.NewResponse(&emhv1.UpdateHeroResponse{
			Hero: &emhv1.HeroDetail{
				Summary: mapHeroToSummary(updated),
				FullBio: updated.FullBio,
				Status:  emhv1.PublicationStatus(updated.Status),
			},
		}), nil
	}

	return connect.NewResponse(&emhv1.UpdateHeroResponse{
		Hero: mapHeroDetailToProto(detail),
	}), nil
}

// DeleteHero обрабатывает удаление или архивацию.
func (s *HeroAdminServer) DeleteHero(
	ctx context.Context,
	req *connect.Request[emhv1.DeleteHeroRequest],
) (*connect.Response[emhv1.DeleteHeroResponse], error) {
	mode := "soft"
	if req.Msg.HardDelete {
		mode = "hard"
	}

	s.logger.InfoContext(ctx, "deleting hero", "id", req.Msg.Id, "mode", mode)

	if err := s.heroUC.DeleteHero(ctx, req.Msg.Id, req.Msg.HardDelete); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "delete hero failed")
	}

	return connect.NewResponse(&emhv1.DeleteHeroResponse{Success: true}), nil
}

// ListHeroes возвращает список героев для админ-панели с фильтрацией по статусу.
func (s *HeroAdminServer) ListHeroes(
	ctx context.Context,
	req *connect.Request[emhv1.ListAdminHeroesRequest],
) (*connect.Response[emhv1.ListAdminHeroesResponse], error) {
	limit := int(req.Msg.Pagination.GetPageSize())
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := domain.HeroFilter{
		SearchQuery:     req.Msg.SearchQuery,
		Cursor:          req.Msg.Pagination.GetCursor(),
		Limit:           limit,
		IncludeArchived: req.Msg.IncludeArchived,
	}

	// Если статус передан явно (не 0 / Unspecified), используем его
	if req.Msg.Status != emhv1.PublicationStatus_PUBLICATION_STATUS_UNSPECIFIED {
		status := domain.PublicationStatus(req.Msg.Status)
		filter.Status = &status
	}

	heroes, nextCursor, total, err := s.heroUC.ListHeroes(ctx, filter)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "list admin heroes failed")
	}

	protoHeroes := make([]*emhv1.HeroSummary, 0, len(heroes))
	for _, h := range heroes {
		protoHeroes = append(protoHeroes, mapHeroToSummary(h))
	}

	return connect.NewResponse(&emhv1.ListAdminHeroesResponse{
		Heroes: protoHeroes,
		Pagination: &emhv1.PaginationResponse{
			NextCursor: nextCursor,
			TotalCount: total,
		},
	}), nil
}

// AddHeroAward привязывает награду.
func (s *HeroAdminServer) AddHeroAward(
	ctx context.Context,
	req *connect.Request[emhv1.AddHeroAwardRequest],
) (*connect.Response[emhv1.AddHeroAwardRequest], error) {
	s.logger.InfoContext(ctx, "adding award to hero",
		"hero_id", req.Msg.HeroId,
		"award_id", req.Msg.AwardId,
	)

	var awardDate *string
	if req.Msg.AwardDate != "" {
		awardDate = &req.Msg.AwardDate
	}

	if err := s.heroAwardUC.Add(ctx, req.Msg.HeroId, req.Msg.AwardId, awardDate, req.Msg.DecreeNumber); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero award failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// RemoveHeroAward удаляет связь с наградой.
func (s *HeroAdminServer) RemoveHeroAward(
	ctx context.Context,
	req *connect.Request[emhv1.RemoveHeroAwardRequest],
) (*connect.Response[emhv1.RemoveHeroAwardRequest], error) {
	s.logger.InfoContext(ctx, "removing award from hero",
		"hero_id", req.Msg.HeroId,
		"award_id", req.Msg.AwardId,
	)

	if err := s.heroAwardUC.Remove(ctx, req.Msg.HeroId, req.Msg.AwardId); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "remove hero award failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// AddHeroPhoto добавляет фото.
func (s *HeroAdminServer) AddHeroPhoto(
	ctx context.Context,
	req *connect.Request[emhv1.AddHeroPhotoRequest],
) (*connect.Response[emhv1.AddHeroPhotoRequest], error) {
	s.logger.InfoContext(ctx, "adding photo to hero",
		"hero_id", req.Msg.HeroId,
		"is_main", req.Msg.IsMain,
	)

	params := domain.AddPhotoParams{
		HeroID:      req.Msg.HeroId,
		URL:         req.Msg.Url,
		Description: req.Msg.Description,
		SortOrder:   int(req.Msg.SortOrder),
		IsMain:      req.Msg.IsMain,
		FaceBox:     mapFaceBoxToDomain(req.Msg.FaceBox),
	}

	if _, err := s.photoUC.Add(ctx, params); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero photo failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// ReorderHeroPhotos меняет порядок фото.
func (s *HeroAdminServer) ReorderHeroPhotos(
	ctx context.Context,
	req *connect.Request[emhv1.ReorderHeroPhotosRequest],
) (*connect.Response[emhv1.ReorderHeroPhotosRequest], error) {
	s.logger.InfoContext(ctx, "reordering hero photos",
		"hero_id", req.Msg.HeroId,
		"count", len(req.Msg.PhotoIds),
	)

	if err := s.photoUC.Reorder(ctx, req.Msg.HeroId, req.Msg.PhotoIds); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "reorder hero photos failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// BatchAddHeroPhotos добавляет несколько фотографий за один запрос.
func (s *HeroAdminServer) BatchAddHeroPhotos(
	ctx context.Context,
	req *connect.Request[emhv1.BatchAddHeroPhotosRequest],
) (*connect.Response[emhv1.BatchAddHeroPhotosResponse], error) {
	s.logger.InfoContext(ctx, "batch adding photos",
		"hero_id", req.Msg.HeroId,
		"count", len(req.Msg.Photos),
	)

	params := make([]domain.AddPhotoParams, 0, len(req.Msg.Photos))
	for _, p := range req.Msg.Photos {
		params = append(params, domain.AddPhotoParams{
			HeroID:      req.Msg.HeroId,
			URL:         p.Url,
			Description: p.Description,
			IsMain:      p.IsMain,
			SortOrder:   int(p.SortOrder),
			FaceBox:     mapFaceBoxToDomain(p.FaceBox),
		})
	}

	added, err := s.photoUC.BatchAdd(ctx, req.Msg.HeroId, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "batch add photos failed")
	}

	protoAdded := make([]*emhv1.AddedPhoto, 0, len(added))
	for _, p := range added {
		protoAdded = append(protoAdded, &emhv1.AddedPhoto{
			Id:     p.ID,
			Url:    p.URL,
			IsMain: p.IsMain,
		})
	}

	return connect.NewResponse(&emhv1.BatchAddHeroPhotosResponse{
		Added:      protoAdded,
		AddedCount: int32(len(protoAdded)),
	}), nil
}

// UpdateHeroPhoto обновляет метаданные фотографии (описание, face_box).
func (s *HeroAdminServer) UpdateHeroPhoto(
	ctx context.Context,
	req *connect.Request[emhv1.UpdateHeroPhotoRequest],
) (*connect.Response[emhv1.UpdateHeroPhotoResponse], error) {
	s.logger.InfoContext(ctx, "updating hero photo",
		"hero_id", req.Msg.HeroId,
		"photo_id", req.Msg.PhotoId,
	)

	params := domain.UpdatePhotoParams{
		PhotoID:     req.Msg.PhotoId,
		HeroID:      req.Msg.HeroId,
		Description: req.Msg.Description,
		FaceBox:     mapFaceBoxToDomain(req.Msg.FaceBox),
		FieldMask:   req.Msg.FieldMask,
	}

	photo, err := s.photoUC.Update(ctx, params)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "update hero photo failed")
	}

	return connect.NewResponse(&emhv1.UpdateHeroPhotoResponse{
		Photo: mapPhotoToProto(photo),
	}), nil
}

// DeleteHeroPhotos удаляет отмеченные фотографии пакетно.
func (s *HeroAdminServer) DeleteHeroPhotos(
	ctx context.Context,
	req *connect.Request[emhv1.DeleteHeroPhotosRequest],
) (*connect.Response[emhv1.DeleteHeroPhotosResponse], error) {
	s.logger.InfoContext(ctx, "batch deleting photos",
		"hero_id", req.Msg.HeroId,
		"count", len(req.Msg.PhotoIds),
	)

	deleted, err := s.photoUC.DeleteBatch(ctx, req.Msg.HeroId, req.Msg.PhotoIds)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "batch delete photos failed")
	}

	return connect.NewResponse(&emhv1.DeleteHeroPhotosResponse{
		DeletedCount: int32(deleted),
		Success:      true,
	}), nil
}

// AddHeroConflict привязывает конфликт к герою.
func (s *HeroAdminServer) AddHeroConflict(
	ctx context.Context,
	req *connect.Request[emhv1.AddHeroConflictRequest],
) (*connect.Response[emhv1.AddHeroConflictRequest], error) {
	s.logger.InfoContext(ctx, "adding conflict to hero",
		"hero_id", req.Msg.HeroId,
		"conflict_id", req.Msg.ConflictId,
	)

	if err := s.heroConflictUC.Add(ctx, req.Msg.HeroId, req.Msg.ConflictId, req.Msg.SpecificLocation, req.Msg.RankAtConflict); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero conflict failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// RemoveHeroConflict удаляет связь героя с конфликтом.
func (s *HeroAdminServer) RemoveHeroConflict(
	ctx context.Context,
	req *connect.Request[emhv1.RemoveHeroConflictRequest],
) (*connect.Response[emhv1.RemoveHeroConflictRequest], error) {
	s.logger.InfoContext(ctx, "removing conflict from hero",
		"hero_id", req.Msg.HeroId,
		"conflict_id", req.Msg.ConflictId,
	)

	if err := s.heroConflictUC.Remove(ctx, req.Msg.HeroId, req.Msg.ConflictId); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "remove hero conflict failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// AddHeroLocation привязывает локацию к герою с указанным типом связи.
func (s *HeroAdminServer) AddHeroLocation(
	ctx context.Context,
	req *connect.Request[emhv1.AddHeroLocationRequest],
) (*connect.Response[emhv1.AddHeroLocationRequest], error) {
	s.logger.InfoContext(ctx, "adding location to hero",
		"hero_id", req.Msg.HeroId,
		"location_id", req.Msg.LocationId,
		"type", req.Msg.Type.String(),
	)

	if err := s.heroLocationUC.Add(ctx, req.Msg.HeroId, req.Msg.LocationId, domain.HeroLocationType(req.Msg.Type)); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero location failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// RemoveHeroLocation удаляет связь героя с локацией указанного типа.
func (s *HeroAdminServer) RemoveHeroLocation(
	ctx context.Context,
	req *connect.Request[emhv1.RemoveHeroLocationRequest],
) (*connect.Response[emhv1.RemoveHeroLocationRequest], error) {
	s.logger.InfoContext(ctx, "removing location from hero",
		"hero_id", req.Msg.HeroId,
		"location_id", req.Msg.LocationId,
		"type", req.Msg.Type.String(),
	)

	if err := s.heroLocationUC.Remove(ctx, req.Msg.HeroId, req.Msg.LocationId, domain.HeroLocationType(req.Msg.Type)); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "remove hero location failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// AddHeroSource добавляет источник данных к герою.
func (s *HeroAdminServer) AddHeroSource(ctx context.Context, req *connect.Request[emhv1.AddHeroSourceRequest]) (*connect.Response[emhv1.AddHeroSourceRequest], error) {
	params := domain.AddHeroSourceParams{
		HeroID:     req.Msg.HeroId,
		URL:        req.Msg.Url,
		Title:      req.Msg.Title,
		SourceType: req.Msg.SourceType,
		Excerpt:    req.Msg.Excerpt,
	}

	if _, err := s.heroSourceUC.Add(ctx, params); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero source failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// RemoveHeroSource удаляет источник.
func (s *HeroAdminServer) RemoveHeroSource(ctx context.Context, req *connect.Request[emhv1.RemoveHeroSourceRequest]) (*connect.Response[emhv1.RemoveHeroSourceRequest], error) {
	if err := s.heroSourceUC.Remove(ctx, req.Msg.HeroId, req.Msg.SourceId); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "remove hero source failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// AddHeroRelation создаёт связь между героями.
func (s *HeroAdminServer) AddHeroRelation(ctx context.Context, req *connect.Request[emhv1.AddHeroRelationRequest]) (*connect.Response[emhv1.AddHeroRelationRequest], error) {
	params := domain.AddHeroRelationParams{
		FromHeroID:   req.Msg.FromHeroId,
		ToHeroID:     req.Msg.ToHeroId,
		RelationType: req.Msg.RelationType,
		Description:  req.Msg.Description,
	}

	if _, err := s.heroRelationUC.Add(ctx, params); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "add hero relation failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// RemoveHeroRelation удаляет связь.
func (s *HeroAdminServer) RemoveHeroRelation(ctx context.Context, req *connect.Request[emhv1.RemoveHeroRelationRequest]) (*connect.Response[emhv1.RemoveHeroRelationRequest], error) {
	if err := s.heroRelationUC.Remove(ctx, req.Msg.Id); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "remove hero relation failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// SetMainHeroPhoto назначает указанную фотографию главной для героя.
func (s *HeroAdminServer) SetMainHeroPhoto(
	ctx context.Context,
	req *connect.Request[emhv1.SetMainHeroPhotoRequest],
) (*connect.Response[emhv1.SetMainHeroPhotoRequest], error) {
	s.logger.InfoContext(ctx, "setting main photo",
		"hero_id", req.Msg.HeroId,
		"photo_id", req.Msg.PhotoId,
	)

	if err := s.photoUC.SetMain(ctx, req.Msg.HeroId, req.Msg.PhotoId); err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "set main photo failed")
	}

	return connect.NewResponse(req.Msg), nil
}

// mapFaceBoxToProto конвертирует доменный FaceBox в proto.
func mapFaceBoxToProto(fb *domain.FaceBox) *emhv1.FaceBox {
	if fb == nil {
		return nil
	}
	return &emhv1.FaceBox{
		X:      int32(fb.X),
		Y:      int32(fb.Y),
		Width:  int32(fb.Width),
		Height: int32(fb.Height),
	}
}

// mapFaceBoxToDomain конвертирует proto FaceBox в доменный.
func mapFaceBoxToDomain(fb *emhv1.FaceBox) *domain.FaceBox {
	if fb == nil {
		return nil
	}
	return &domain.FaceBox{
		X:      int(fb.X),
		Y:      int(fb.Y),
		Width:  int(fb.Width),
		Height: int(fb.Height),
	}
}

// stringPtr возвращает указатель на строку.
func stringPtr(v string) *string {
	return &v
}

// parseFlexibleDate парсит старую строковую дату YYYY-MM-DD в FlexibleDate.
// Пустая строка трактуется как неизвестная дата.
func parseFlexibleDate(s string) (domain.FlexibleDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return domain.NewUnknownDate(), nil
	}

	t, err := usecase.ParseDateStr(s)
	if err != nil {
		return domain.FlexibleDate{}, err
	}

	if t == nil {
		return domain.NewUnknownDate(), nil
	}

	return domain.NewExactDate(*t), nil
}

// parseFlexibleDatePtr парсит старую строковую дату в указатель на FlexibleDate.
// Пустая строка возвращает nil, что для Update означает очистку даты.
func parseFlexibleDatePtr(s string) (*domain.FlexibleDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	fd, err := parseFlexibleDate(s)
	if err != nil {
		return nil, err
	}

	return &fd, nil
}

// mapUseCaseError маппит ошибки usecase в Connect-ошибки.
// Использует errors.Is() для типизированных ошибок вместо strings.Contains().
func (s *HeroAdminServer) mapUseCaseError(ctx context.Context, err error, action string) error {
	return mapDomainError(ctx, s.logger, err, action)
}
