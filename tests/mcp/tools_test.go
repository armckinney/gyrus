package mcp_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies end-to-end execution of all 11 Gyrus memory tools over Stdio JSON-RPC 2.0.
// [Execution Surface]: 'gyrus mcp serve' Subprocess over Stdio
// [Assertions]: All 11 tools execute successfully, mutate state, and respond with valid MCP content envelopes.
// -----------------------------------------------------------------------------
func TestMCP_All11ToolsEndToEnd(t *testing.T) {
	workspaceDir := t.TempDir()

	// Init workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = workspaceDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	client := startMCPClient(t, workspaceDir)

	// Initialize Handshake
	client.Call(t, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test-runner"},
	})
	client.Notify(t, "notifications/initialized", map[string]any{})

	// Helper to extract text content from tool response
	parseToolContent := func(t *testing.T, resp *JSONRPCResponse) string {
		if resp.Error != nil {
			t.Fatalf("Tool call returned JSON-RPC error: %+v", resp.Error)
		}
		var toolRes struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		}
		if err := json.Unmarshal(resp.Result, &toolRes); err != nil {
			t.Fatalf("Failed unmarshaling tool call result: %v", err)
		}
		if toolRes.IsError {
			t.Fatalf("Tool returned isError=true with content: %+v", toolRes.Content)
		}
		if len(toolRes.Content) == 0 {
			t.Fatalf("Tool returned empty content slice")
		}
		return toolRes.Content[0].Text
	}

	// 1. gyrus_create (Doc A - PRD)
	createRespA := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_create",
		"arguments": map[string]any{
			"id":          "prd-mcp-e2e-001",
			"title":       "MCP Interface Spec",
			"category":    "product",
			"type":        "prd",
			"owner_group": "platform-eng",
			"status":      "active",
			"content":     "Full specification for AI agent memory interface.",
		},
	})
	textA := parseToolContent(t, createRespA)
	if !strings.Contains(textA, "prd-mcp-e2e-001") {
		t.Errorf("Expected doc ID in create response, got: %s", textA)
	}

	// 2. gyrus_create (Doc B - ADR)
	createRespB := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_create",
		"arguments": map[string]any{
			"id":          "adr-mcp-e2e-001",
			"title":       "Stdio JSON-RPC Transport",
			"category":    "architecture",
			"type":        "adr",
			"owner_group": "platform-eng",
			"status":      "proposed",
			"content":     "Architecture for bidirectional standard streams RPC communication.",
		},
	})
	parseToolContent(t, createRespB)

	// 3. gyrus_get
	getResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_get",
		"arguments": map[string]any{
			"id": "prd-mcp-e2e-001",
		},
	})
	getText := parseToolContent(t, getResp)
	if !strings.Contains(getText, "title: MCP Interface Spec") {
		t.Errorf("Expected markdown frontmatter in get response, got:\n%s", getText)
	}

	// 4. gyrus_update
	updateResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_update",
		"arguments": map[string]any{
			"id":               "prd-mcp-e2e-001",
			"title":            "MCP Interface Spec v2",
			"expected_version": 1,
		},
	})
	parseToolContent(t, updateResp)

	// 5. gyrus_search
	searchResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_search",
		"arguments": map[string]any{
			"query":    "bidirectional",
			"category": "architecture",
		},
	})
	searchText := parseToolContent(t, searchResp)
	if !strings.Contains(searchText, "adr-mcp-e2e-001") {
		t.Errorf("Expected adr-mcp-e2e-001 in search results, got:\n%s", searchText)
	}

	// 6. gyrus_suggest_context
	suggestResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_suggest_context",
		"arguments": map[string]any{
			"prompt": "How does the Stdio JSON-RPC transport work?",
		},
	})
	suggestText := parseToolContent(t, suggestResp)
	if len(suggestText) == 0 {
		t.Errorf("Expected non-empty context suggestion")
	}

	// 7. gyrus_link
	linkResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_link",
		"arguments": map[string]any{
			"from_id":  "prd-mcp-e2e-001",
			"to_id":    "adr-mcp-e2e-001",
			"rel_type": "depends_on",
		},
	})
	linkText := parseToolContent(t, linkResp)
	if !strings.Contains(linkText, "Linked") {
		t.Errorf("Expected link confirmation, got: %s", linkText)
	}

	// 8. gyrus_neighbors
	neighborsResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_neighbors",
		"arguments": map[string]any{
			"id":        "prd-mcp-e2e-001",
			"direction": "outgoing",
		},
	})
	neighborsText := parseToolContent(t, neighborsResp)
	if !strings.Contains(neighborsText, "adr-mcp-e2e-001") {
		t.Errorf("Expected neighbor edge in response, got: %s", neighborsText)
	}

	// 9. gyrus_traverse
	traverseResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_traverse",
		"arguments": map[string]any{
			"start_id":  "prd-mcp-e2e-001",
			"max_depth": 2,
		},
	})
	traverseText := parseToolContent(t, traverseResp)
	if !strings.Contains(traverseText, "adr-mcp-e2e-001") {
		t.Errorf("Expected traversal path to adr-mcp-e2e-001, got: %s", traverseText)
	}

	// 10. gyrus_unlink
	unlinkResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_unlink",
		"arguments": map[string]any{
			"from_id":  "prd-mcp-e2e-001",
			"to_id":    "adr-mcp-e2e-001",
			"rel_type": "depends_on",
		},
	})
	unlinkText := parseToolContent(t, unlinkResp)
	if !strings.Contains(unlinkText, "Unlinked") {
		t.Errorf("Expected unlink confirmation, got: %s", unlinkText)
	}

	// 11. gyrus_sync
	syncResp := client.Call(t, "tools/call", map[string]any{
		"name":      "gyrus_sync",
		"arguments": map[string]any{},
	})
	parseToolContent(t, syncResp)

	// 12. gyrus_archive
	archiveResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_archive",
		"arguments": map[string]any{
			"id": "adr-mcp-e2e-001",
		},
	})
	archiveText := parseToolContent(t, archiveResp)
	if !strings.Contains(archiveText, "Archived") {
		t.Errorf("Expected archive confirmation, got: %s", archiveText)
	}

	// 13. gyrus_schema_get
	schemaGetResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_schema_get",
		"arguments": map[string]any{
			"type": "adr",
		},
	})
	schemaGetText := parseToolContent(t, schemaGetResp)
	if !strings.Contains(schemaGetText, "Architecture Decision Record") {
		t.Errorf("Expected ADR schema template, got: %s", schemaGetText)
	}

	// 14. gyrus_schema_set
	customRunbook := `---
id: <unique-id>
title: <Runbook>
category: operations
type: runbook
owner_group: ops
version: 1
status: active
---

# Runbook Template
`
	schemaSetResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_schema_set",
		"arguments": map[string]any{
			"type":    "runbook",
			"content": customRunbook,
		},
	})
	schemaSetText := parseToolContent(t, schemaSetResp)
	if !strings.Contains(schemaSetText, "saved") {
		t.Errorf("Expected saved status, got: %s", schemaSetText)
	}

	// 15. gyrus_schema_list
	schemaListResp := client.Call(t, "tools/call", map[string]any{
		"name":      "gyrus_schema_list",
		"arguments": map[string]any{},
	})
	schemaListText := parseToolContent(t, schemaListResp)
	if !strings.Contains(schemaListText, "runbook") || !strings.Contains(schemaListText, "persisted") {
		t.Errorf("Expected runbook persisted in schema list, got: %s", schemaListText)
	}

	// 16. Read resource memory://schema/runbook
	readResResp := client.Call(t, "resources/read", map[string]any{
		"uri": "memory://schema/runbook",
	})
	if readResResp.Error != nil {
		t.Fatalf("resources/read memory://schema/runbook returned error: %+v", readResResp.Error)
	}

	// 17. gyrus_schema_delete
	schemaDelResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_schema_delete",
		"arguments": map[string]any{
			"type": "runbook",
		},
	})
	schemaDelText := parseToolContent(t, schemaDelResp)
	if !strings.Contains(schemaDelText, "deleted") {
		t.Errorf("Expected deleted status, got: %s", schemaDelText)
	}

	// 18. gyrus_schema_get for non-existent schema returns isError: true
	schemaGetDeletedResp := client.Call(t, "tools/call", map[string]any{
		"name": "gyrus_schema_get",
		"arguments": map[string]any{
			"type": "runbook",
		},
	})
	var toolErrRes struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(schemaGetDeletedResp.Result, &toolErrRes); err != nil {
		t.Fatalf("Failed unmarshaling tool call result: %v", err)
	}
	if !toolErrRes.IsError {
		t.Errorf("Expected isError=true for non-existent schema, got false")
	}
	if len(toolErrRes.Content) == 0 || !strings.Contains(toolErrRes.Content[0].Text, "not found") {
		t.Errorf("Expected 'not found' error text, got: %+v", toolErrRes.Content)
	}
}
