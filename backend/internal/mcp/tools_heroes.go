package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"codeberg.org/Thr0TT1e/emh/backend/internal/domain"
	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// registerSearchHeroesTool регистрирует инструмент поиска героев.
func registerSearchHeroesTool(s *server.MCPServer, heroUC usecase.HeroUseCase, logger *slog.Logger) {
	tool := mcp.NewTool("search_heroes",
		mcp.WithDescription("Поиск героев по ФИО, позывному, конфликту или локации. Возвращает список кратких карточек."),
		mcp.WithString("query",
			mcp.Description("Поисковая строка по ФИО или позывному героя"),
		),
		mcp.WithString("conflict_id",
			mcp.Description("UUID конфликта для фильтрации (опционально)"),
		),
		mcp.WithString("location_id",
			mcp.Description("UUID локации для фильтрации (опционально)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Максимум результатов (1-100, по умолчанию 20)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		query, _ := args["query"].(string)
		conflictID, _ := args["conflict_id"].(string)
		locationID, _ := args["location_id"].(string)

		limit := 20
		if v, ok := args["limit"].(float64); ok && v > 0 {
			limit = int(v)
			if limit > 100 {
				limit = 100
			}
		}

		heroes, nextCursor, total, err := heroUC.ListHeroes(ctx, domain.HeroFilter{
			SearchQuery: query,
			ConflictID:  conflictID,
			LocationID:  locationID,
			Limit:       limit,
		})
		if err != nil {
			logger.Error("search_heroes failed", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("поиск героев не удался: %v", err)), nil
		}

		// Формируем читаемый ответ для агента
		type heroSummary struct {
			ID         string   `json:"id"`
			LastName   string   `json:"last_name"`
			FirstName  string   `json:"first_name"`
			MiddleName string   `json:"middle_name,omitempty"`
			Nickname   string   `json:"nickname,omitempty"`
			Rank       string   `json:"rank,omitempty"`
			ShortBio   string   `json:"short_bio,omitempty"`
			AwardNames []string `json:"award_names,omitempty"`
			BirthDate  string   `json:"birth_date,omitempty"`
			DeathDate  string   `json:"death_date,omitempty"`
		}

		summaries := make([]heroSummary, 0, len(heroes))
		for _, h := range heroes {
			s := heroSummary{
				ID:         h.ID,
				LastName:   h.LastName,
				FirstName:  h.FirstName,
				MiddleName: h.MiddleName,
				Nickname:   h.Nickname,
				Rank:       h.Rank,
				ShortBio:   h.ShortBio,
				AwardNames: h.AwardNames,
			}
			if h.BirthDate.DisplayText != "" {
				s.BirthDate = h.BirthDate.DisplayText
			}
			if h.DeathDate.DisplayText != "" {
				s.DeathDate = h.DeathDate.DisplayText
			}
			summaries = append(summaries, s)
		}

		result := map[string]any{
			"total":       total,
			"count":       len(summaries),
			"next_cursor": nextCursor,
			"heroes":      summaries,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize response"), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

// registerGetHeroDetailTool регистрирует инструмент получения полной карточки героя.
func registerGetHeroDetailTool(s *server.MCPServer, heroQueryUC usecase.HeroQueryUseCase, logger *slog.Logger) {
	tool := mcp.NewTool("get_hero_detail",
		mcp.WithDescription("Возвращает полную карточку героя: биография, фото, награды, конфликты, локации, источники, связи с другими героями."),
		mcp.WithString("id",
			mcp.Description("UUID героя (обязательно)"),
			mcp.Required(),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		heroID, _ := args["id"].(string)
		if heroID == "" {
			return mcp.NewToolResultError("параметр 'id' обязателен"), nil
		}

		detail, err := heroQueryUC.GetHeroDetail(ctx, heroID)
		if err != nil {
			logger.Error("get_hero_detail failed", "hero_id", heroID, "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("герой не найден: %v", err)), nil
		}

		// Формируем полный ответ
		h := detail.Hero
		response := map[string]any{
			"id":             h.ID,
			"last_name":      h.LastName,
			"first_name":     h.FirstName,
			"middle_name":    h.MiddleName,
			"nickname":       h.Nickname,
			"rank":           h.Rank,
			"unit":           h.Unit,
			"position":       h.Position,
			"service_branch": h.ServiceBranch,
			"short_bio":      h.ShortBio,
			"full_bio":       h.FullBio,
			"cause_of_death": h.CauseOfDeath,
			"memberships":    h.Memberships,
			"status":         statusToString(h.Status),
		}

		// Гибкие даты
		response["birth_date"] = flexibleDateToMap(h.BirthDate)
		response["death_date"] = flexibleDateToMap(h.DeathDate)
		response["service_start_date"] = flexibleDateToMap(h.ServiceStartDate)

		// Награды
		awards := make([]map[string]any, 0, len(detail.Awards))
		for _, a := range detail.Awards {
			award := map[string]any{
				"name":          a.AwardName,
				"decree_number": a.DecreeNumber,
			}
			if a.AwardDate != nil {
				award["date"] = a.AwardDate.Format("2006-01-02")
			}
			awards = append(awards, award)
		}
		response["awards"] = awards

		// Конфликты
		conflicts := make([]map[string]any, 0, len(detail.Conflicts))
		for _, c := range detail.Conflicts {
			conflicts = append(conflicts, map[string]any{
				"name":              c.ConflictName,
				"specific_location": c.SpecificLocation,
				"rank_at_conflict":  c.RankAtConflict,
			})
		}
		response["conflicts"] = conflicts

		// Локации
		locations := make([]map[string]any, 0, len(detail.Locations))
		for _, l := range detail.Locations {
			loc := map[string]any{
				"type": heroLocationTypeToString(l.Type),
			}
			if l.Location != nil {
				loc["name"] = l.Location.Name
				loc["historical_name"] = l.Location.HistoricalName
			}
			locations = append(locations, loc)
		}
		response["locations"] = locations

		// Источники
		sources := make([]map[string]any, 0, len(detail.Sources))
		for _, src := range detail.Sources {
			sources = append(sources, map[string]any{
				"url":   src.URL,
				"title": src.Title,
				"type":  src.SourceType,
			})
		}
		response["sources"] = sources

		// Связи с другими героями
		relations := make([]map[string]any, 0, len(detail.Relations))
		for _, rel := range detail.Relations {
			relations = append(relations, map[string]any{
				"to_hero_id":    rel.ToHeroID,
				"hero_name":     rel.RelatedHeroName,
				"relation_type": rel.RelationType,
				"description":   rel.Description,
			})
		}
		response["relations"] = relations

		// Фото (только метаданные, без бинарных данных)
		photos := make([]map[string]any, 0, len(detail.Photos))
		for _, p := range detail.Photos {
			photos = append(photos, map[string]any{
				"url":         p.URL,
				"thumbnail":   p.ThumbnailURL,
				"description": p.Description,
				"is_main":     p.IsMain,
			})
		}
		response["photos"] = photos

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize response"), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

// --- Вспомогательные функции ---

func statusToString(s domain.PublicationStatus) string {
	switch s {
	case domain.StatusDraft:
		return "draft"
	case domain.StatusPublished:
		return "published"
	case domain.StatusArchived:
		return "archived"
	default:
		return "unspecified"
	}
}

func heroLocationTypeToString(t domain.HeroLocationType) string {
	switch t {
	case domain.HeroLocationTypeBirth:
		return "birth"
	case domain.HeroLocationTypeDeath:
		return "death"
	case domain.HeroLocationTypeBurial:
		return "burial"
	case domain.HeroLocationTypeResidence:
		return "residence"
	default:
		return "unspecified"
	}
}

func flexibleDateToMap(d domain.FlexibleDate) map[string]any {
	result := map[string]any{
		"precision": precisionToString(d.Precision),
	}
	if d.DisplayText != "" {
		result["display"] = d.DisplayText
	}
	if d.Anchor != nil {
		result["anchor"] = d.Anchor.Format("2006-01-02")
	}
	return result
}

func precisionToString(p domain.DatePrecision) string {
	switch p {
	case domain.PrecisionExact:
		return "exact"
	case domain.PrecisionMonth:
		return "month"
	case domain.PrecisionYear:
		return "year"
	case domain.PrecisionSeason:
		return "season"
	case domain.PrecisionDayMonth:
		return "day_month"
	case domain.PrecisionRange:
		return "range"
	case domain.PrecisionUnknown:
		return "unknown"
	default:
		return "unspecified"
	}
}
