package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (h *Handler) registerDocumentTools(s *server.MCPServer) {
	// 1. gyrus_create
	createTool := mcp.NewTool("gyrus_create",
		mcp.WithDescription("Create a new OKF contract document"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID (^[a-z0-9-_]+$)")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Document Title")),
		mcp.WithString("category", mcp.Required(), mcp.Description("Category (architecture|business-logic|product|operations|technical)")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Type (adr|prd|guide|specification|...)")),
		mcp.WithString("owner_group", mcp.Required(), mcp.Description("Owner Group")),
		mcp.WithString("status", mcp.Description("Status (draft|proposed|active)")),
		mcp.WithString("content", mcp.Description("Body content")),
	)
	s.AddTool(createTool, h.HandleCreate)

	// 2. gyrus_get
	getTool := mcp.NewTool("gyrus_get",
		mcp.WithDescription("Get an OKF document by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID")),
	)
	s.AddTool(getTool, h.HandleGet)

	// 3. gyrus_update
	updateTool := mcp.NewTool("gyrus_update",
		mcp.WithDescription("Update an existing OKF document"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID")),
		mcp.WithString("title", mcp.Description("Updated Title")),
		mcp.WithString("status", mcp.Description("Updated Status")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		mcp.WithString("content", mcp.Description("Updated body content")),
		mcp.WithNumber("expected_version", mcp.Description("Optimistic concurrency expected version")),
	)
	s.AddTool(updateTool, h.HandleUpdate)

	// 4. gyrus_archive
	archiveTool := mcp.NewTool("gyrus_archive",
		mcp.WithDescription("Archive (delete) a document from storage and search index"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID to archive")),
	)
	s.AddTool(archiveTool, h.HandleArchive)
}

// HandleCreate handles the gyrus_create MCP tool request.
func (h *Handler) HandleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	ref, err := h.engine.Create(ctx, doc)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create failed: %v", err)), nil
	}

	data, _ := json.Marshal(ref)
	return mcp.NewToolResultText(string(data)), nil
}

// HandleGet handles the gyrus_get MCP tool request.
func (h *Handler) HandleGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	doc, err := h.engine.Get(ctx, id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("document not found: %v", err)), nil
	}

	data, _ := okf.SerializeMarkdown(&doc)
	return mcp.NewToolResultText(string(data)), nil
}

// HandleUpdate handles the gyrus_update MCP tool request.
func (h *Handler) HandleUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	title := getArgString(req, "title")
	status := getArgString(req, "status")
	tags := getArgString(req, "tags")
	content := getArgString(req, "content")
	expectedVersion := getArgInt(req, "expected_version")

	patch := gyrus.DocumentPatch{}
	if title != "" {
		patch.Title = &title
	}
	if status != "" {
		patch.Status = &status
	}
	if tags != "" {
		tList := strings.Split(tags, ",")
		patch.Tags = &tList
	}
	if content != "" {
		patch.Content = &content
	}

	ref, err := h.engine.Update(ctx, id, patch, expectedVersion)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("update failed: %v", err)), nil
	}

	data, _ := json.Marshal(ref)
	return mcp.NewToolResultText(string(data)), nil
}

// HandleArchive handles the gyrus_archive MCP tool request.
func (h *Handler) HandleArchive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	if err := h.engine.Archive(ctx, id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("archive failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Archived document '%s'", id)), nil
}
