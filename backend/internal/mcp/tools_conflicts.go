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

func registerListConflictsTool(s *server.MCPServer, conflictUC usecase.ConflictUseCase, logger *slog.Logger) {
	tool := mcp.NewTool("list_conflicts",
		mcp.WithDescription("Справочник военных конфликтов. Возвращает список конфликтов с иерархией."),
		mcp.WithString("type",
			mcp.Description("Фильтр по типу: global, local, peacekeeping, counter_terrorism, special_operation. Пусто = все."),
		),
		mcp.WithString("parent_id",
			mcp.Description("UUID родительского конфликта для получения подчинённых операций."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		typeStr, _ := args["type"].(string)
		parentID, _ := args["parent_id"].(string)

		conflictType := parseConflictType(typeStr)

		conflicts, _, _, err := conflictUC.List(ctx, domain.ConflictFilter{
			Type:     conflictType,
			ParentID: parentID,
			Limit:    100, // разумный лимит для MCP-агентов
		})
		if err != nil {
			logger.Error("list_conflicts failed", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("ошибка загрузки конфликтов: %v", err)), nil
		}

		type conflictItem struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			Description      string `json:"description,omitempty"`
			Type             string `json:"type"`
			StartDate        string `json:"start_date,omitempty"`
			EndDate          string `json:"end_date,omitempty"`
			ParentConflictID string `json:"parent_conflict_id,omitempty"`
		}

		items := make([]conflictItem, 0, len(conflicts))
		for _, c := range conflicts {
			item := conflictItem{
				ID:          c.ID,
				Name:        c.Name,
				Description: c.Description,
				Type:        conflictTypeToString(c.Type),
			}
			if c.StartDate != nil {
				item.StartDate = c.StartDate.Format("2006-01-02")
			}
			if c.EndDate != nil {
				item.EndDate = c.EndDate.Format("2006-01-02")
			}
			if c.ParentConflictID != nil {
				item.ParentConflictID = *c.ParentConflictID
			}
			items = append(items, item)
		}

		result := map[string]any{
			"count":     len(items),
			"conflicts": items,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize response"), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

func parseConflictType(s string) domain.ConflictType {
	switch s {
	case "global":
		return domain.ConflictTypeGlobal
	case "local":
		return domain.ConflictTypeLocal
	case "peacekeeping":
		return domain.ConflictTypePeacekeeping
	case "counter_terrorism":
		return domain.ConflictTypeCounterTerrorism
	case "special_operation":
		return domain.ConflictTypeSpecialOperation
	default:
		return domain.ConflictTypeUnspecified
	}
}

func conflictTypeToString(t domain.ConflictType) string {
	switch t {
	case domain.ConflictTypeGlobal:
		return "global"
	case domain.ConflictTypeLocal:
		return "local"
	case domain.ConflictTypePeacekeeping:
		return "peacekeeping"
	case domain.ConflictTypeCounterTerrorism:
		return "counter_terrorism"
	case domain.ConflictTypeSpecialOperation:
		return "special_operation"
	default:
		return "unspecified"
	}
}
