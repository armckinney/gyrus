package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAgentPromptRoutingSimulation tests that both skills provide clear, unambiguous
// instructions so that an AI Agent (like Google Antigravity) will reliably select the correct tool.
func TestAgentPromptRoutingSimulation(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	cliSkillBytes, err := os.ReadFile(filepath.Join(repoRoot, "skills", "gyrus-cli", "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed reading gyrus-cli SKILL.md: %v", err)
	}

	mcpSkillBytes, err := os.ReadFile(filepath.Join(repoRoot, "skills", "gyrus-mcp", "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed reading gyrus-mcp SKILL.md: %v", err)
	}

	cliPrompt := string(cliSkillBytes)
	mcpPrompt := string(mcpSkillBytes)

	// User Intent Scenarios to test prompt instruction clarity
	scenarios := []struct {
		intent                string
		expectedCLISubcommand string
		expectedMCPTool       string
	}{
		{
			intent:                "Find architecture decision records for storage",
			expectedCLISubcommand: "gyrus search --query",
			expectedMCPTool:       "gyrus_search",
		},
		{
			intent:                "Get document details for a document ID",
			expectedCLISubcommand: "gyrus get <document-id>",
			expectedMCPTool:       "gyrus_get_document",
		},
		{
			intent:                "Suggest context for relational SQL databases",
			expectedCLISubcommand: "gyrus suggest-context --prompt",
			expectedMCPTool:       "gyrus_suggest_context",
		},
		{
			intent:                "Create a new PRD for feature X",
			expectedCLISubcommand: "gyrus create",
			expectedMCPTool:       "gyrus_create_document",
		},
	}

	for _, sc := range scenarios {
		t.Run("CLI_Routing/"+sc.intent, func(t *testing.T) {
			if !strings.Contains(cliPrompt, sc.expectedCLISubcommand) {
				t.Errorf("gyrus-cli skill does not explicitly instruct agent on %s (expected %s)", sc.intent, sc.expectedCLISubcommand)
			}
		})

		t.Run("MCP_Routing/"+sc.intent, func(t *testing.T) {
			if !strings.Contains(mcpPrompt, sc.expectedMCPTool) {
				t.Errorf("gyrus-mcp skill does not explicitly instruct agent on %s (expected %s)", sc.intent, sc.expectedMCPTool)
			}
		})
	}
}
