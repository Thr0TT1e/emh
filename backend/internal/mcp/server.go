// Package mcp реализует MCP-сервер для интеграции с внешними LLM-агентами.
package mcp

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/server"

	"codeberg.org/Thr0TT1e/emh/backend/internal/usecase"
)

// ServerConfig параметры инициализации MCP-сервера.
type ServerConfig struct {
	HeroQueryUC  usecase.HeroQueryUseCase
	HeroUC       usecase.HeroUseCase
	SubmissionUC usecase.SubmissionUseCase
	ConflictUC   usecase.ConflictUseCase
	ExtractionUC usecase.ExtractionUseCase
	Logger       *slog.Logger
}

// NewMCPServer создаёт и настраивает MCP-сервер с регистрацией tools.
func NewMCPServer(cfg ServerConfig) *server.MCPServer {
	s := server.NewMCPServer(
		"emh-mcp-server",
		"0.2.0",
		server.WithToolCapabilities(false),
		server.WithLogging(),
	)

	// Read-only
	registerPingTool(s, cfg.Logger)
	registerSearchHeroesTool(s, cfg.HeroUC, cfg.Logger)
	registerGetHeroDetailTool(s, cfg.HeroQueryUC, cfg.Logger)
	registerListConflictsTool(s, cfg.ConflictUC, cfg.Logger)

	// Write (human-in-the-loop через Submission)
	registerCreateSubmissionTool(s, cfg.SubmissionUC, cfg.Logger)

	// LLM-пайплайн (опционально)
	registerExtractHeroDataTool(s, cfg.ExtractionUC, cfg.Logger)

	return s
}
