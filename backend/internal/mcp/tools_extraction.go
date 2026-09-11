package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

func registerExtractHeroDataTool(s *server.MCPServer, extractionUC usecase.ExtractionUseCase, logger *slog.Logger) {
	tool := mcp.NewTool("extract_hero_data",
		mcp.WithDescription(
			"Извлекает структурированные данные героя из сырого текста через LLM. "+
				"Возвращает ФИО, звание, конфликты, награды, локации. "+
				"Результат НЕ сохраняется в БД — используйте create_submission для этого."),
		mcp.WithString("text",
			mcp.Description("Сырой текст для анализа (статья, пост, письмо)"),
		),
		mcp.WithString("source_url",
			mcp.Description("URL источника (альтернатива тексту)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Проверяем доступность LLM
		if extractionUC == nil {
			return mcp.NewToolResultError(
				"LLM-пайплайн отключён. Включите секцию 'llm' в config.yaml"), nil
		}

		args := req.GetArguments()
		text, _ := args["text"].(string)
		sourceURL, _ := args["source_url"].(string)

		if text == "" && sourceURL == "" {
			return mcp.NewToolResultError(
				"укажите 'text' или 'source_url'"), nil
		}

		result, err := extractionUC.ExtractFromText(ctx, text, sourceURL)
		if err != nil {
			logger.Error("extract_hero_data failed", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("ошибка извлечения: %v", err)), nil
		}

		// Формируем ответ
		response := map[string]any{
			"warnings":           result.Warnings,
			"duplicate_hero_ids": result.DuplicateHeroIDs,
			"source_urls":        result.SourceURLs,
		}

		if result.Hero != nil {
			response["hero"] = result.Hero
		}
		if len(result.Conflicts) > 0 {
			response["conflicts"] = result.Conflicts
		}
		if len(result.Awards) > 0 {
			response["awards"] = result.Awards
		}
		if len(result.Locations) > 0 {
			response["locations"] = result.Locations
		}

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialize response"), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
