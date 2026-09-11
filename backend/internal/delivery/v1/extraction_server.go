package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	emhv1 "codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1"
	"codeberg.org/Thr0TT1e/emh/backend/internal/gen/emh/v1/emhv1connect"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// ExtractionServer реализует emhv1connect.ExtractionServiceHandler.
type ExtractionServer struct {
	emhv1connect.UnimplementedExtractionServiceHandler
	uc     usecase.ExtractionUseCase
	logger *slog.Logger
}

// NewExtractionServer создаёт сервер извлечения данных.
func NewExtractionServer(uc usecase.ExtractionUseCase, logger *slog.Logger) *ExtractionServer {
	return &ExtractionServer{
		uc:     uc,
		logger: logger,
	}
}

// ExtractHeroData извлекает данные героя из текста или URL.
func (s *ExtractionServer) ExtractHeroData(
	ctx context.Context,
	req *connect.Request[emhv1.ExtractHeroDataRequest],
) (*connect.Response[emhv1.ExtractHeroDataResponse], error) {
	msg := req.Msg

	// Вызываем usecase
	result, err := s.uc.ExtractFromText(ctx, msg.RawText, msg.SourceUrl)
	if err != nil {
		return nil, mapDomainError(ctx, s.logger, err, "extract_hero_data")
	}

	// Маппинг в proto
	resp := &emhv1.ExtractHeroDataResponse{}

	if result.Hero != nil {
		resp.Hero = &emhv1.ExtractedHero{
			LastName:         result.Hero.LastName,
			FirstName:        result.Hero.FirstName,
			MiddleName:       result.Hero.MiddleName,
			Nickname:         result.Hero.Nickname,
			Rank:             result.Hero.Rank,
			Unit:             result.Hero.Unit,
			Position:         result.Hero.Position,
			ServiceBranch:    result.Hero.ServiceBranch,
			BirthDate:        result.Hero.BirthDate,
			DeathDate:        result.Hero.DeathDate,
			CauseOfDeath:     result.Hero.CauseOfDeath,
			ServiceStartDate: result.Hero.ServiceStartDate,
			ShortBio:         result.Hero.ShortBio,
			FullBio:          result.Hero.FullBio,
			Memberships:      result.Hero.Memberships,
		}
	}

	for _, c := range result.Conflicts {
		resp.Conflicts = append(resp.Conflicts, &emhv1.ExtractedConflict{
			Name:                  c.Name,
			SpecificLocation:      c.SpecificLocation,
			RankAtConflict:        c.RankAtConflict,
			SuggestedConflictType: c.SuggestedConflictType,
		})
	}

	for _, a := range result.Awards {
		resp.Awards = append(resp.Awards, &emhv1.ExtractedAward{
			Name:         a.Name,
			AwardDate:    a.AwardDate,
			DecreeNumber: a.DecreeNumber,
		})
	}

	for _, l := range result.Locations {
		resp.Locations = append(resp.Locations, &emhv1.ExtractedLocation{
			Name:             l.Name,
			HistoricalName:   l.HistoricalName,
			LocationType:     l.LocationType,
			HeroLocationType: l.HeroLocationType,
		})
	}

	resp.SourceUrls = result.SourceURLs
	resp.Warnings = result.Warnings
	resp.DuplicateHeroIds = result.DuplicateHeroIDs

	return connect.NewResponse(resp), nil
}
