package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (h *Handler) registerGraphTools(s *server.MCPServer) {
	// 7. gyrus_link
	linkTool := mcp.NewTool("gyrus_link",
		mcp.WithDescription("Create a directed relationship edge between two documents"),
		mcp.WithString("from_id", mcp.Required(), mcp.Description("Source document ID")),
		mcp.WithString("to_id", mcp.Required(), mcp.Description("Target document ID")),
		mcp.WithString("rel_type", mcp.Description("Relationship type (depends_on|supersedes|implements|mitigates)")),
	)
	s.AddTool(linkTool, h.HandleLink)

	// 8. gyrus_unlink
	unlinkTool := mcp.NewTool("gyrus_unlink",
		mcp.WithDescription("Remove a directed relationship edge between two documents"),
		mcp.WithString("from_id", mcp.Required(), mcp.Description("Source document ID")),
		mcp.WithString("to_id", mcp.Required(), mcp.Description("Target document ID")),
		mcp.WithString("rel_type", mcp.Description("Relationship type (depends_on|supersedes|implements|mitigates)")),
	)
	s.AddTool(unlinkTool, h.HandleUnlink)

	// 9. gyrus_neighbors
	neighborsTool := mcp.NewTool("gyrus_neighbors",
		mcp.WithDescription("Retrieve neighboring relationship edges for a document"),
		mcp.WithString("id", mcp.Required(), mcp.Description("Document ID")),
		mcp.WithString("direction", mcp.Description("Direction (outgoing|incoming|both)")),
		mcp.WithString("rel_type", mcp.Description("Filter by relationship type")),
	)
	s.AddTool(neighborsTool, h.HandleNeighbors)

	// 10. gyrus_traverse
	traverseTool := mcp.NewTool("gyrus_traverse",
		mcp.WithDescription("Execute graph path traversal from a starting document"),
		mcp.WithString("start_id", mcp.Required(), mcp.Description("Start Document ID")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum traversal depth (default: 2)")),
	)
	s.AddTool(traverseTool, h.HandleTraverse)
}

// HandleLink handles the gyrus_link MCP tool request.
func (h *Handler) HandleLink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID := getArgString(req, "from_id")
	toID := getArgString(req, "to_id")
	relStr := getArgString(req, "rel_type")

	relType := gyrus.RelationshipType(relStr)
	if relType == "" {
		relType = gyrus.RelDependsOn
	}

	if err := h.engine.Link(ctx, fromID, toID, relType); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("link failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Linked '%s' -[%s]-> '%s'", fromID, relType, toID)), nil
}

// HandleUnlink handles the gyrus_unlink MCP tool request.
func (h *Handler) HandleUnlink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID := getArgString(req, "from_id")
	toID := getArgString(req, "to_id")
	relStr := getArgString(req, "rel_type")

	relType := gyrus.RelationshipType(relStr)
	if relType == "" {
		relType = gyrus.RelDependsOn
	}

	if err := h.engine.Unlink(ctx, fromID, toID, relType); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("unlink failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Unlinked '%s' -[%s]-> '%s'", fromID, relType, toID)), nil
}

// HandleNeighbors handles the gyrus_neighbors MCP tool request.
func (h *Handler) HandleNeighbors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := getArgString(req, "id")
	direction := getArgString(req, "direction")
	relStr := getArgString(req, "rel_type")

	edges, err := h.engine.Neighbors(ctx, id, gyrus.EdgeFilter{
		Direction:        direction,
		RelationshipType: gyrus.RelationshipType(relStr),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("neighbors failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(edges, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleTraverse handles the gyrus_traverse MCP tool request.
func (h *Handler) HandleTraverse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startID := getArgString(req, "start_id")
	maxDepth := getArgInt(req, "max_depth")
	if maxDepth <= 0 {
		maxDepth = 2
	}

	paths, err := h.engine.Traverse(ctx, gyrus.GraphQuery{
		StartID:  startID,
		MaxDepth: maxDepth,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("traverse failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(paths, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
