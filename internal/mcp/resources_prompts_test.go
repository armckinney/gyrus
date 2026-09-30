package mcp_test

import (
	"os"
	"testing"

	"github.com/armckinney/gyrus/internal/mcp"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that MCP Server exposes resources and prompts.
// [Execution Surface]: In-Memory MCP Server Instance
// [Assertions]: Resources and prompts registered without panic.
// -----------------------------------------------------------------------------
func TestMCPResourcesAndPromptsRegistration(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	server, err := mcp.NewServer()
	if err != nil {
		t.Fatalf("Failed to initialize MCP Server with resources/prompts: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server instance")
	}
}
