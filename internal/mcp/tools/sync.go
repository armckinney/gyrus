package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (h *Handler) registerSyncTools(s *server.MCPServer) {
	// 11. gyrus_sync
	syncTool := mcp.NewTool("gyrus_sync",
		mcp.WithDescription("Scan concept storage and validate or rebuild the selected index"),
	)
	s.AddTool(syncTool, h.HandleSync)
}

// HandleSync handles the gyrus_sync MCP tool request.
func (h *Handler) HandleSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	report, err := h.engine.Sync(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("sync failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	result := mcp.NewToolResultText(string(data))
	result.IsError = len(report.Errors) > 0
	return result, nil
}
