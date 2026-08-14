package tools

import (
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Handler provides MCP tool execution handlers backed by the lifecycle engine.
type Handler struct {
	engine *lifecycle.Engine
}

// NewHandler creates a new tool Handler instance.
func NewHandler(engine *lifecycle.Engine) *Handler {
	return &Handler{engine: engine}
}

// Register registers all 11 Gyrus memory MCP tools with the server.
func Register(s *server.MCPServer, engine *lifecycle.Engine) {
	h := NewHandler(engine)
	h.registerDocumentTools(s)
	h.registerSearchTools(s)
	h.registerGraphTools(s)
	h.registerSyncTools(s)
}

func getArgString(req mcp.CallToolRequest, key string) string {
	if req.Params.Arguments == nil {
		return ""
	}
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return ""
	}
	val, ok := args[key]
	if !ok || val == nil {
		return ""
	}
	str, _ := val.(string)
	return str
}

func getArgInt(req mcp.CallToolRequest, key string) int {
	if req.Params.Arguments == nil {
		return 0
	}
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return 0
	}
	val, ok := args[key]
	if !ok || val == nil {
		return 0
	}
	if f, ok := val.(float64); ok {
		return int(f)
	}
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}
