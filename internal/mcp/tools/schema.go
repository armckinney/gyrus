package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (h *Handler) registerSchemaTools(s *server.MCPServer) {
	// 1. gyrus_schema_get
	schemaGetTool := mcp.NewTool("gyrus_schema_get",
		mcp.WithDescription("Get an OKF contract schema template by document type"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Document Type (adr|prd|guide|...)")),
	)
	s.AddTool(schemaGetTool, h.HandleSchemaGet)

	// 2. gyrus_schema_set
	schemaSetTool := mcp.NewTool("gyrus_schema_set",
		mcp.WithDescription("Save an OKF schema template into the persistence layer"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Document Type")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Markdown body containing YAML frontmatter schema template")),
	)
	s.AddTool(schemaSetTool, h.HandleSchemaSet)

	// 3. gyrus_schema_list
	schemaListTool := mcp.NewTool("gyrus_schema_list",
		mcp.WithDescription("List all available schemas across persistence layer and embedded templates"),
	)
	s.AddTool(schemaListTool, h.HandleSchemaList)

	// 4. gyrus_schema_delete
	schemaDeleteTool := mcp.NewTool("gyrus_schema_delete",
		mcp.WithDescription("Delete a custom schema template from the persistence layer"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Document Type")),
	)
	s.AddTool(schemaDeleteTool, h.HandleSchemaDelete)
}

// HandleSchemaGet handles the gyrus_schema_get tool request.
func (h *Handler) HandleSchemaGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	docType := getArgString(req, "type")
	if docType == "" {
		return mcp.NewToolResultError("missing required argument 'type'"), nil
	}

	content, err := h.engine.GetSchema(ctx, docType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get schema: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

// HandleSchemaSet handles the gyrus_schema_set tool request.
func (h *Handler) HandleSchemaSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	docType := getArgString(req, "type")
	if docType == "" {
		return mcp.NewToolResultError("missing required argument 'type'"), nil
	}
	content := getArgString(req, "content")
	if content == "" {
		return mcp.NewToolResultError("missing required argument 'content'"), nil
	}

	if err := h.engine.SaveSchema(ctx, docType, content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to save schema: %v", err)), nil
	}

	res := map[string]string{
		"status":   "saved",
		"type":     docType,
		"location": fmt.Sprintf(".gyrus/schemas/%s.md", docType),
	}
	data, _ := json.MarshalIndent(res, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleSchemaList handles the gyrus_schema_list tool request.
func (h *Handler) HandleSchemaList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schemas, err := h.engine.ListSchemas(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list schemas: %v", err)), nil
	}

	data, _ := json.MarshalIndent(schemas, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleSchemaDelete handles the gyrus_schema_delete tool request.
func (h *Handler) HandleSchemaDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	docType := getArgString(req, "type")
	if docType == "" {
		return mcp.NewToolResultError("missing required argument 'type'"), nil
	}

	if err := h.engine.DeleteSchema(ctx, docType); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete schema: %v", err)), nil
	}

	res := map[string]string{
		"status": "deleted",
		"type":   docType,
	}
	data, _ := json.MarshalIndent(res, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
