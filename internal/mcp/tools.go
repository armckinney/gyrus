package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerTools() {
	// memory.create
	createTool := mcp.NewTool("memory_create",
		mcp.WithDescription("Create a new OKF contract document"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID (^[a-z0-9-_]+$)")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Document Title")),
		mcp.WithString("category", mcp.Required(), mcp.Description("Category (architecture|business-logic|product|operations|technical)")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Type (adr|prd|guide|specification|...)")),
		mcp.WithString("owner_group", mcp.Required(), mcp.Description("Owner Group")),
		mcp.WithString("status", mcp.Description("Status (draft|proposed|active)")),
		mcp.WithString("content", mcp.Description("Body content")),
	)
	s.mcpServer.AddTool(createTool, s.handleCreate)

	// memory.get
	getTool := mcp.NewTool("memory_get",
		mcp.WithDescription("Get an OKF document by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID")),
	)
	s.mcpServer.AddTool(getTool, s.handleGet)

	// memory.search
	searchTool := mcp.NewTool("memory_search",
		mcp.WithDescription("Search OKF documents using FTS5 lexical keyword matching"),
		mcp.WithString("query", mcp.Description("Search query text")),
		mcp.WithString("category", mcp.Description("Filter by category")),
		mcp.WithString("type", mcp.Description("Filter by type")),
	)
	s.mcpServer.AddTool(searchTool, s.handleSearch)

	// memory.suggest_context
	suggestTool := mcp.NewTool("memory_suggest_context",
		mcp.WithDescription("Suggest relevant context documents for an agent prompt"),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Agent prompt context")),
	)
	s.mcpServer.AddTool(suggestTool, s.handleSuggest)

	// memory.link
	linkTool := mcp.NewTool("memory_link",
		mcp.WithDescription("Create a directed relationship edge between two documents"),
		mcp.WithString("from_id", mcp.Required(), mcp.Description("Source document ID")),
		mcp.WithString("to_id", mcp.Required(), mcp.Description("Target document ID")),
		mcp.WithString("rel_type", mcp.Description("Relationship type (depends_on|supersedes|implements)")),
	)
	s.mcpServer.AddTool(linkTool, s.handleLink)

	// memory.archive
	archiveTool := mcp.NewTool("memory_archive",
		mcp.WithDescription("Archive (delete) a document from storage and search index"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID to archive")),
	)
	s.mcpServer.AddTool(archiveTool, s.handleArchive)
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

func (s *Server) handleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	title := getArgString(req, "title")
	category := getArgString(req, "category")
	docType := getArgString(req, "type")
	ownerGroup := getArgString(req, "owner_group")
	status := getArgString(req, "status")
	content := getArgString(req, "content")

	if status == "" {
		status = "draft"
		if docType == string(gyrus.TypeADR) {
			status = "proposed"
		}
	}

	doc := gyrus.Document{
		ID:         id,
		Title:      title,
		Category:   gyrus.Category(category),
		Type:       gyrus.DocumentType(docType),
		OwnerGroup: ownerGroup,
		Version:    1,
		Status:     status,
		Content:    content,
	}

	ref, err := s.store.Create(ctx, doc)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create failed: %v", err)), nil
	}

	_ = s.indexer.Index(ctx, doc)

	data, _ := json.Marshal(ref)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	doc, err := s.store.Get(ctx, id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("document not found: %v", err)), nil
	}

	data, _ := okf.SerializeMarkdown(&doc)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryStr := getArgString(req, "query")
	catStr := getArgString(req, "category")
	typeStr := getArgString(req, "type")

	q := gyrus.SearchQuery{
		Query: queryStr,
		Filter: gyrus.SearchFilter{
			Category: gyrus.Category(catStr),
			Type:     gyrus.DocumentType(typeStr),
		},
		MaxResults: 10,
	}

	results, err := s.indexer.Search(ctx, q)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleSuggest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := getArgString(req, "prompt")
	results, err := s.indexer.Search(ctx, gyrus.SearchQuery{Query: prompt, MaxResults: 5})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("suggest failed: %v", err)), nil
	}

	var sb strings.Builder
	for _, res := range results {
		if doc, err := s.store.Get(ctx, res.Document.ID); err == nil {
			sb.WriteString(fmt.Sprintf("--- %s (%s) ---\n%s\n\n", doc.ID, doc.Title, doc.Content))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleLink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID := getArgString(req, "from_id")
	toID := getArgString(req, "to_id")
	relStr := getArgString(req, "rel_type")

	relType := gyrus.RelationshipType(relStr)
	if relType == "" {
		relType = gyrus.RelDependsOn
	}

	edge := gyrus.DocumentEdge{
		FromDocumentID:   fromID,
		ToDocumentID:     toID,
		RelationshipType: relType,
		CreatedBy:        "mcp",
		CreatedAt:        time.Now().Truncate(time.Second),
	}

	if err := s.indexer.UpsertEdges(ctx, []gyrus.DocumentEdge{edge}); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("link failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Linked '%s' -> '%s'", fromID, toID)), nil
}

func (s *Server) handleArchive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	if err := s.store.Archive(ctx, id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("archive failed: %v", err)), nil
	}
	_ = s.indexer.Remove(ctx, id)

	return mcp.NewToolResultText(fmt.Sprintf("Archived document '%s'", id)), nil
}
