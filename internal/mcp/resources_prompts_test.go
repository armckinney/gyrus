package mcp_test

import (
	"os"
	"testing"

	"github.com/armckinney/gyrus/internal/mcp"
)

func TestMCPResourcesAndPromptsRegistration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-mcp-res-*")
	if err != nil {
		t.Fatalf("Failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	server, err := mcp.NewServer(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize MCP Server with resources/prompts: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server instance")
	}
}
