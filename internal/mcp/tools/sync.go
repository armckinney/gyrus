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
		mcp.WithDescription("Scan workspace storage and rebuild/sync the search index"),
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
	return mcp.NewToolResultText(string(data)), nil
}
