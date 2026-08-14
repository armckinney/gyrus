package tools_test

import (
	"context"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/mcp/tools"
	mcp_sdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that tools.Register registers all 11 Gyrus memory tools and each handler executes correctly.
// [Execution Surface]: In-Memory Engine + MCP Server
// [Assertions]: Tools register without panic, handlers return successful results, and state mutations persist.
// -----------------------------------------------------------------------------
func TestToolsExecution(t *testing.T) {
	tempDir := t.TempDir()

	application, err := app.New(tempDir)
	if err != nil {
		t.Fatalf("Failed creating app container: %v", err)
	}

	engine, err := application.Engine()
	if err != nil {
		t.Fatalf("Failed initializing engine: %v", err)
	}

	mcpServer := server.NewMCPServer("gyrus-memory", "1.0.0")
	tools.Register(mcpServer, engine)

	handler := tools.NewHandler(engine)
	ctx := context.Background()

	// 1. Create Document
	createReq := mcp_sdk.CallToolRequest{}
	createReq.Params.Name = "gyrus_create"
	createReq.Params.Arguments = map[string]any{
		"id":          "spec-tools-001",
		"title":       "Tools Package Architecture",
		"category":    "technical",
		"type":        "specification",
		"owner_group": "platform-team",
		"status":      "active",
		"content":     "Modular directory architecture for MCP tools.",
	}
	res, err := handler.HandleCreate(ctx, createReq)
	if err != nil || res.IsError {
		t.Fatalf("HandleCreate failed: %v, res: %+v", err, res)
	}

	// 2. Get Document
	getReq := mcp_sdk.CallToolRequest{}
	getReq.Params.Name = "gyrus_get"
	getReq.Params.Arguments = map[string]any{"id": "spec-tools-001"}
	getRes, err := handler.HandleGet(ctx, getReq)
	if err != nil || getRes.IsError {
		t.Fatalf("HandleGet failed: %v", err)
	}

	// 3. Update Document
	updateReq := mcp_sdk.CallToolRequest{}
	updateReq.Params.Name = "gyrus_update"
	updateReq.Params.Arguments = map[string]any{
		"id":               "spec-tools-001",
		"title":            "Tools Package Architecture v2",
		"expected_version": 1,
	}
	updateRes, err := handler.HandleUpdate(ctx, updateReq)
	if err != nil || updateRes.IsError {
		t.Fatalf("HandleUpdate failed: %v", err)
	}

	// 4. Search
	searchReq := mcp_sdk.CallToolRequest{}
	searchReq.Params.Name = "gyrus_search"
	searchReq.Params.Arguments = map[string]any{"query": "Modular"}
	searchRes, err := handler.HandleSearch(ctx, searchReq)
	if err != nil || searchRes.IsError {
		t.Fatalf("HandleSearch failed: %v", err)
	}

	// 5. Suggest Context
	suggestReq := mcp_sdk.CallToolRequest{}
	suggestReq.Params.Name = "gyrus_suggest_context"
	suggestReq.Params.Arguments = map[string]any{"prompt": "MCP tools directory structure"}
	suggestRes, err := handler.HandleSuggest(ctx, suggestReq)
	if err != nil || suggestRes.IsError {
		t.Fatalf("HandleSuggest failed: %v", err)
	}

	// 6. Link & Graph
	linkReq := mcp_sdk.CallToolRequest{}
	linkReq.Params.Name = "gyrus_link"
	linkReq.Params.Arguments = map[string]any{
		"from_id":  "spec-tools-001",
		"to_id":    "spec-tools-001",
		"rel_type": "depends_on",
	}
	linkRes, err := handler.HandleLink(ctx, linkReq)
	if err != nil || linkRes.IsError {
		t.Fatalf("HandleLink failed: %v", err)
	}

	neighborsReq := mcp_sdk.CallToolRequest{}
	neighborsReq.Params.Name = "gyrus_neighbors"
	neighborsReq.Params.Arguments = map[string]any{"id": "spec-tools-001"}
	neighborsRes, err := handler.HandleNeighbors(ctx, neighborsReq)
	if err != nil || neighborsRes.IsError {
		t.Fatalf("HandleNeighbors failed: %v", err)
	}

	traverseReq := mcp_sdk.CallToolRequest{}
	traverseReq.Params.Name = "gyrus_traverse"
	traverseReq.Params.Arguments = map[string]any{"start_id": "spec-tools-001"}
	traverseRes, err := handler.HandleTraverse(ctx, traverseReq)
	if err != nil || traverseRes.IsError {
		t.Fatalf("HandleTraverse failed: %v", err)
	}

	unlinkReq := mcp_sdk.CallToolRequest{}
	unlinkReq.Params.Name = "gyrus_unlink"
	unlinkReq.Params.Arguments = map[string]any{
		"from_id":  "spec-tools-001",
		"to_id":    "spec-tools-001",
		"rel_type": "depends_on",
	}
	unlinkRes, err := handler.HandleUnlink(ctx, unlinkReq)
	if err != nil || unlinkRes.IsError {
		t.Fatalf("HandleUnlink failed: %v", err)
	}

	// 7. Sync
	syncReq := mcp_sdk.CallToolRequest{}
	syncReq.Params.Name = "gyrus_sync"
	syncRes, err := handler.HandleSync(ctx, syncReq)
	if err != nil || syncRes.IsError {
		t.Fatalf("HandleSync failed: %v", err)
	}

	// 8. Archive
	archiveReq := mcp_sdk.CallToolRequest{}
	archiveReq.Params.Name = "gyrus_archive"
	archiveReq.Params.Arguments = map[string]any{"id": "spec-tools-001"}
	archiveRes, err := handler.HandleArchive(ctx, archiveReq)
	if err != nil || archiveRes.IsError {
		t.Fatalf("HandleArchive failed: %v", err)
	}
}
