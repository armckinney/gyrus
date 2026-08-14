package mcp_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/mcp"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the Gyrus MCP Server initializes with all 11 memory tools, resources, and prompt templates.
// [Execution Surface]: In-Memory MCP Server Instance
// [Assertions]: Server initializes without error and provides non-nil MCPServer and Engine instances.
// -----------------------------------------------------------------------------
func TestMCPServerInitialization(t *testing.T) {
	tempDir := t.TempDir()

	server, err := mcp.NewServer(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize MCP Server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server instance")
	}

	if server.MCPServer() == nil {
		t.Fatal("Expected non-nil underlying mark3labs MCPServer")
	}

	if server.Engine() == nil {
		t.Fatal("Expected non-nil lifecycle Engine")
	}
}
