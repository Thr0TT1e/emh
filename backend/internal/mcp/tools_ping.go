package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPingTool(s *server.MCPServer, logger *slog.Logger) {
	tool := mcp.NewTool("ping",
		mcp.WithDescription("Проверка доступности EMH MCP-сервера"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("ping tool called")
		return mcp.NewToolResultText("pong"), nil
	})
}
