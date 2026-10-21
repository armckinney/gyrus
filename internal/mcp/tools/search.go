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
		mcp.WithDescription("Search concepts using the configured provider; OKF returns the top 10 keyword matches or all empty-query metadata matches"),
		mcp.WithString("query", mcp.Description("Search query text")),
		mcp.WithString("category", mcp.Description("Filter by category")),
		mcp.WithString("type", mcp.Description("Filter by type")),
		mcp.WithString("status", mcp.Description("Filter by status")),
		mcp.WithString("tag", mcp.Description("Filter by tag")),
		mcp.WithString("tenant", mcp.Description("Filter by tenant")),
		mcp.WithString("scope", mcp.Description("Filter by scope (project, reference, all)")),
		mcp.WithString("project", mcp.Description("Filter by project name")),
		mcp.WithNumber("max_results", mcp.Description("Maximum results; defaults to 10 for keywords, unlimited for empty-query browsing")),
	)
	s.AddTool(searchTool, h.HandleSearch)

	// 6. gyrus_suggest_context
	suggestTool := mcp.NewTool("gyrus_suggest_context",
		mcp.WithDescription("Suggest relevant context concepts for an agent prompt"),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Agent prompt context")),
		mcp.WithString("scope", mcp.Description("Context scope (project, reference, all)")),
		mcp.WithString("project", mcp.Description("Explicit project name")),
		mcp.WithNumber("max_concepts", mcp.Description("Maximum concepts to return")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum estimated tokens; defaults to 4000. The layer holds the best-matching sections that fit")),
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
	tenantStr := getArgString(req, "tenant")
	scopeStr := getArgString(req, "scope")
	projectStr := getArgString(req, "project")

	filter := gyrus.SearchFilter{
		Category: gyrus.Category(catStr),
		Type:     gyrus.ConceptType(typeStr),
		Status:   statusStr,
		Tag:      tagStr,
		Tenant:   tenantStr,
		Scope:    scopeStr,
		Project:  projectStr,
	}

	results, err := h.engine.SearchQuery(ctx, gyrus.SearchQuery{Query: queryStr, Filter: filter, MaxResults: getArgInt(req, "max_results")})
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
	projectStr := getArgString(req, "project")
	maxConcepts := getArgInt(req, "max_concepts")
	if maxConcepts <= 0 {
		maxConcepts = 5
	}
	maxTokens := getArgInt(req, "max_tokens")
	if maxTokens <= 0 {
		maxTokens = 4000
	}

	filter := gyrus.SearchFilter{
		Scope:   scopeStr,
		Project: projectStr,
	}

	contextLayer, err := h.engine.SuggestContextWithBudget(ctx, prompt, filter, maxConcepts, maxTokens)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("suggest failed: %v", err)), nil
	}

	return mcp.NewToolResultText(contextLayer), nil
}
