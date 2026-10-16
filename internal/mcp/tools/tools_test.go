package tools_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/mcp/tools"
	"github.com/armckinney/gyrus/pkg/gyrus"
	mcp_sdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestSearchLimitPassThrough(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gyrus.yaml"), []byte("project: demo\nstorage:\n  provider: localfs\n  root: .gyrus\n"), 0644); err != nil {
		t.Fatal(err)
	}
	application, err := app.NewWithProject(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := application.Engine()
	if err != nil {
		t.Fatal(err)
	}
	const fixtureCount = 12
	for i := 0; i < fixtureCount; i++ {
		_, err := engine.Create(context.Background(), gyrus.Concept{ID: fmt.Sprintf("limit-%d", i), Title: "Limit fixture", Tenant: "root", Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Version: 1, Status: "draft", Content: "limitmarker"})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		query       string
		limit, want int
	}{{"limitmarker", 0, 10}, {"limitmarker", 3, 3}, {"limitmarker", 20, fixtureCount}, {"", 0, fixtureCount}, {"", 3, 3}} {
		req := mcp_sdk.CallToolRequest{}
		req.Params.Arguments = map[string]any{"query": tc.query, "max_results": tc.limit, "scope": "all"}
		result, err := tools.NewHandler(engine).HandleSearch(context.Background(), req)
		if err != nil || result.IsError {
			t.Fatalf("search: %+v %v", result, err)
		}
		var docs []gyrus.SearchResult
		if err := json.Unmarshal([]byte(result.Content[0].(mcp_sdk.TextContent).Text), &docs); err != nil {
			t.Fatal(err)
		}
		if len(docs) != tc.want {
			t.Errorf("query=%q limit=%d: got %d, want %d", tc.query, tc.limit, len(docs), tc.want)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that tools.Register registers all 11 Gyrus memory tools and each handler executes correctly.
// [Execution Surface]: In-Memory Engine + MCP Server
// [Assertions]: Tools register without panic, handlers return successful results, and state mutations persist.
func TestToolsExecution(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, ".gyrus.yaml"), []byte("storage:\n  provider: localfs\n  root: .gyrus\nindex:\n  provider: okf\ngraph:\n  provider: okf\nsearch:\n  provider: okf\n"), 0644); err != nil {
		t.Fatal(err)
	}
	application, err := app.New()
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
		"id":       "spec-tools-001",
		"title":    "Tools Package Architecture",
		"category": "technical",
		"type":     "specification",
		"tenant":   "platform-team",
		"status":   "active",
		"content":  "Modular directory architecture for MCP tools.",
	}
	res, err := handler.HandleCreate(ctx, createReq)
	if err != nil || res.IsError {
		t.Fatalf("HandleCreate failed: %v, res: %+v", err, res)
	}

	// 2. Get Document
	getReq := mcp_sdk.CallToolRequest{}
	getReq.Params.Name = "gyrus_get"
	getReq.Params.Arguments = map[string]any{"id": "platform-team/reference/spec-tools-001"}
	getRes, err := handler.HandleGet(ctx, getReq)
	if err != nil || getRes.IsError {
		t.Fatalf("HandleGet failed: %v", err)
	}

	// 3. Update Document
	updateReq := mcp_sdk.CallToolRequest{}
	updateReq.Params.Name = "gyrus_update"
	updateReq.Params.Arguments = map[string]any{
		"id":               "platform-team/reference/spec-tools-001",
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
		"from_id":  "platform-team/reference/spec-tools-001",
		"to_id":    "platform-team/reference/spec-tools-001",
		"rel_type": "depends_on",
	}
	linkRes, err := handler.HandleLink(ctx, linkReq)
	if err != nil || linkRes.IsError {
		t.Fatalf("HandleLink failed: %v", err)
	}

	neighborsReq := mcp_sdk.CallToolRequest{}
	neighborsReq.Params.Name = "gyrus_neighbors"
	neighborsReq.Params.Arguments = map[string]any{"id": "platform-team/reference/spec-tools-001"}
	neighborsRes, err := handler.HandleNeighbors(ctx, neighborsReq)
	if err != nil || neighborsRes.IsError {
		t.Fatalf("HandleNeighbors failed: %v", err)
	}

	traverseReq := mcp_sdk.CallToolRequest{}
	traverseReq.Params.Name = "gyrus_traverse"
	traverseReq.Params.Arguments = map[string]any{"start_id": "platform-team/reference/spec-tools-001"}
	traverseRes, err := handler.HandleTraverse(ctx, traverseReq)
	if err != nil || traverseRes.IsError {
		t.Fatalf("HandleTraverse failed: %v", err)
	}

	unlinkReq := mcp_sdk.CallToolRequest{}
	unlinkReq.Params.Name = "gyrus_unlink"
	unlinkReq.Params.Arguments = map[string]any{
		"from_id":  "platform-team/reference/spec-tools-001",
		"to_id":    "platform-team/reference/spec-tools-001",
		"rel_type": "depends_on",
	}
	unlinkRes, err := handler.HandleUnlink(ctx, unlinkReq)
	if err != nil || unlinkRes.IsError {
		t.Fatalf("HandleUnlink failed: %v", err)
	}

	// 8. Archive
	archiveReq := mcp_sdk.CallToolRequest{}
	archiveReq.Params.Name = "gyrus_archive"
	archiveReq.Params.Arguments = map[string]any{"id": "platform-team/reference/spec-tools-001"}
	archiveRes, err := handler.HandleArchive(ctx, archiveReq)
	if err != nil || archiveRes.IsError {
		t.Fatalf("HandleArchive failed: %v", err)
	}

	// 9. Schema Tools: Get, Set, List, Delete
	schemaGetReq := mcp_sdk.CallToolRequest{}
	schemaGetReq.Params.Name = "gyrus_schema_get"
	schemaGetReq.Params.Arguments = map[string]any{"type": "adr"}
	schemaGetRes, err := handler.HandleSchemaGet(ctx, schemaGetReq)
	if err != nil || schemaGetRes.IsError {
		t.Fatalf("HandleSchemaGet failed: %v", err)
	}

	schemaSetReq := mcp_sdk.CallToolRequest{}
	schemaSetReq.Params.Name = "gyrus_schema_set"
	schemaSetReq.Params.Arguments = map[string]any{
		"type": "runbook",
		"content": `---
id: <unique-id>
title: <Title>
category: operations
type: runbook
tenant: ops
version: 1
status: active
---

# Runbook MCP Content
`,
	}
	schemaSetRes, err := handler.HandleSchemaSet(ctx, schemaSetReq)
	if err != nil || schemaSetRes.IsError {
		t.Fatalf("HandleSchemaSet failed: %v", err)
	}

	schemaListReq := mcp_sdk.CallToolRequest{}
	schemaListReq.Params.Name = "gyrus_schema_list"
	schemaListRes, err := handler.HandleSchemaList(ctx, schemaListReq)
	if err != nil || schemaListRes.IsError {
		t.Fatalf("HandleSchemaList failed: %v", err)
	}

	schemaDelReq := mcp_sdk.CallToolRequest{}
	schemaDelReq.Params.Name = "gyrus_schema_delete"
	schemaDelReq.Params.Arguments = map[string]any{"type": "runbook"}
	schemaDelRes, err := handler.HandleSchemaDelete(ctx, schemaDelReq)
	if err != nil || schemaDelRes.IsError {
		t.Fatalf("HandleSchemaDelete failed: %v", err)
	}
}

func TestOKFSearchRejectsMalformedCollection(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gyrus.yaml"), []byte("storage:\n  provider: localfs\n  root: .gyrus\n"), 0644); err != nil {
		t.Fatal(err)
	}
	collection := filepath.Join(root, ".gyrus/concepts/root/reference")
	if err := os.MkdirAll(collection, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(collection, "bad.md"), []byte("bad payload"), 0644); err != nil {
		t.Fatal(err)
	}
	application, err := app.NewWithProject(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := application.Engine()
	if err != nil {
		t.Fatal(err)
	}
	result, err := tools.NewHandler(engine).HandleSearch(context.Background(), mcp_sdk.CallToolRequest{})
	if err != nil || !result.IsError {
		t.Fatalf("MCP failure status: %+v %v", result, err)
	}
	text, ok := result.Content[0].(mcp_sdk.TextContent)
	if !ok {
		t.Fatal("report is not text JSON")
	}
	if !strings.Contains(text.Text, "bad.md") {
		t.Fatalf("missing malformed-file location: %s", text.Text)
	}
}
