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

func registerCreateSubmissionTool(s *server.MCPServer, submissionUC usecase.SubmissionUseCase, logger *slog.Logger) {
	tool := mcp.NewTool("create_submission",
		mcp.WithDescription(
			"Создаёт заявку на добавление или исправление данных героя. "+
				"Заявка создаётся со статусом DRAFT и будет рассмотрена модератором. "+
				"LLM не пишет в БД напрямую — только через заявки (human-in-the-loop)."),
		mcp.WithString("submitter_name",
			mcp.Description("Имя отправителя заявки"),
			mcp.Required(),
		),
		mcp.WithString("submitter_email",
			mcp.Description("Email отправителя для обратной связи"),
			mcp.Required(),
		),
		mcp.WithString("payload_json",
			mcp.Description(
				"JSON-строка с данными героя. Формат: "+
					`{"hero": {"first_name": "...", "last_name": "...", ...}, "conflicts": [...], "awards": [...]}`),
			mcp.Required(),
		),
		mcp.WithString("target_hero_id",
			mcp.Description("UUID существующего героя, если заявка на исправление данных."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		submitterName, _ := args["submitter_name"].(string)
		submitterEmail, _ := args["submitter_email"].(string)
		payloadJSON, _ := args["payload_json"].(string)
		targetHeroID, _ := args["target_hero_id"].(string)

		// Базовая валидация до вызова usecase
		if submitterName == "" {
			return mcp.NewToolResultError("параметр 'submitter_name' обязателен"), nil
		}
		if submitterEmail == "" {
			return mcp.NewToolResultError("параметр 'submitter_email' обязателен"), nil
		}
		if payloadJSON == "" {
			return mcp.NewToolResultError("параметр 'payload_json' обязателен"), nil
		}

		// Проверяем валидность JSON до отправки в usecase
		if !json.Valid([]byte(payloadJSON)) {
			return mcp.NewToolResultError("payload_json должен быть валидным JSON"), nil
		}

		params := domain.CreateSubmissionParams{
			SubmitterName:  submitterName,
			SubmitterEmail: submitterEmail,
			PayloadJSON:    payloadJSON,
		}
		if targetHeroID != "" {
			params.TargetHeroID = &targetHeroID
		}

		id, err := submissionUC.Create(ctx, params)
		if err != nil {
			logger.Error("create_submission failed", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("ошибка создания заявки: %v", err)), nil
		}

		// Возвращаем понятный ответ для агента
		response := map[string]any{
			"submission_id": id,
			"status":        "draft",
			"message": "Заявка создана и будет рассмотрена модератором. " +
				"Данные не добавлены в базу напрямую — только через модерацию.",
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
