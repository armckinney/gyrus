package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Evaluation
// [Purpose]: Verifies that Agent Skills (gyrus-cli and gyrus-mcp) provide clear, unambiguous instructions for LLM agent routing.
// [Execution Surface]: Packaging Skills Manifests (SKILL.md)
// [Assertions]: Skills contain exact command/tool names corresponding to user developer intents.
// -----------------------------------------------------------------------------
func TestAgentPromptRoutingSimulation(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	skillBytes, err := os.ReadFile(filepath.Join(repoRoot, "packaging", "plugins", "gyrus", "skills", "gyrus", "SKILL.md"))
	if err != nil {
		t.Fatalf("Failed reading gyrus SKILL.md: %v", err)
	}

	prompt := string(skillBytes)

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
			expectedCLISubcommand: "gyrus get",
			expectedMCPTool:       "gyrus_get",
		},
		{
			intent:                "Suggest context for relational SQL databases",
			expectedCLISubcommand: "gyrus suggest-context",
			expectedMCPTool:       "gyrus_suggest_context",
		},
		{
			intent:                "Create a new PRD for feature X",
			expectedCLISubcommand: "gyrus create",
			expectedMCPTool:       "gyrus_create",
		},
		{
			intent:                "Link two documents with relationship edge",
			expectedCLISubcommand: "gyrus link",
			expectedMCPTool:       "gyrus_link",
		},
		{
			intent:                "Reindex workspace markdown files",
			expectedCLISubcommand: "gyrus sync",
			expectedMCPTool:       "gyrus_sync",
		},
	}

	for _, sc := range scenarios {
		t.Run("CLI_Routing/"+sc.intent, func(t *testing.T) {
			if !strings.Contains(prompt, sc.expectedCLISubcommand) {
				t.Errorf("gyrus skill does not instruct agent on %s (expected %s)", sc.intent, sc.expectedCLISubcommand)
			}
		})

		t.Run("MCP_Routing/"+sc.intent, func(t *testing.T) {
			if !strings.Contains(prompt, sc.expectedMCPTool) {
				t.Errorf("gyrus skill does not instruct agent on %s (expected %s)", sc.intent, sc.expectedMCPTool)
			}
		})
	}
}
