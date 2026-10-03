package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (h *Handler) registerSearchTools(s *server.MCPServer) {
	// 5. gyrus_search
	searchTool := mcp.NewTool("gyrus_search",
		mcp.WithDescription("Search OKF documents using FTS5 lexical keyword matching and metadata filters"),
		mcp.WithString("query", mcp.Description("Search query text")),
		mcp.WithString("category", mcp.Description("Filter by category")),
		mcp.WithString("type", mcp.Description("Filter by type")),
		mcp.WithString("status", mcp.Description("Filter by status")),
		mcp.WithString("tag", mcp.Description("Filter by tag")),
		mcp.WithString("scope", mcp.Description("Filter by scope (workspace, reference, all)")),
		mcp.WithString("workspace", mcp.Description("Filter by workspace name")),
	)
	s.AddTool(searchTool, h.HandleSearch)

	// 6. gyrus_suggest_context
	suggestTool := mcp.NewTool("gyrus_suggest_context",
		mcp.WithDescription("Suggest relevant context documents for an agent prompt"),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Agent prompt context")),
		mcp.WithString("scope", mcp.Description("Context scope (workspace, reference, all)")),
		mcp.WithString("workspace", mcp.Description("Explicit workspace name")),
		mcp.WithNumber("max_docs", mcp.Description("Maximum documents to return")),
	)
	s.AddTool(suggestTool, h.HandleSuggest)
}

// HandleSearch handles the gyrus_search MCP tool request.
func (h *Handler) HandleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryStr := getArgString(req, "query")
	catStr := getArgString(req, "category")
	typeStr := getArgString(req, "type")
	statusStr := getArgString(req, "status")
	tagStr := getArgString(req, "tag")
	scopeStr := getArgString(req, "scope")
	workspaceStr := getArgString(req, "workspace")

	filter := gyrus.SearchFilter{
		Category:  gyrus.Category(catStr),
		Type:      gyrus.DocumentType(typeStr),
		Status:    statusStr,
		Tag:       tagStr,
		Scope:     scopeStr,
		Workspace: workspaceStr,
	}

	results, err := h.engine.Search(ctx, queryStr, filter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleSuggest handles the gyrus_suggest_context MCP tool request.
func (h *Handler) HandleSuggest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := getArgString(req, "prompt")
	scopeStr := getArgString(req, "scope")
	workspaceStr := getArgString(req, "workspace")
	maxDocs := getArgInt(req, "max_docs")
	if maxDocs <= 0 {
		maxDocs = 5
	}

	filter := gyrus.SearchFilter{
		Scope:     scopeStr,
		Workspace: workspaceStr,
	}

	contextLayer, err := h.engine.SuggestContextWithFilter(ctx, prompt, filter, maxDocs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("suggest failed: %v", err)), nil
	}

	return mcp.NewToolResultText(contextLayer), nil
}
