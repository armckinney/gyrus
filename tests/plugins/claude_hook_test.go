package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies the Claude Code plugin's prompt hook.
// [Execution Surface]: Filesystem Canonical Plugin Bundle (packaging/plugins/gyrus/)
// [Assertions]: hooks/hooks.json registers one UserPromptSubmit command running 'gyrus suggest-context --hook' with
// a token budget and a timeout.
// -----------------------------------------------------------------------------
func TestClaudePluginPromptHook(t *testing.T) {
	var hooks struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "packaging", "plugins", "gyrus", "hooks", "hooks.json"))
	if err != nil {
		t.Fatalf("missing hooks/hooks.json: %v", err)
	}
	if err := json.Unmarshal(data, &hooks); err != nil {
		t.Fatalf("invalid hooks.json: %v", err)
	}
	entries := hooks.Hooks["UserPromptSubmit"]
	if len(entries) != 1 || len(entries[0].Hooks) != 1 {
		t.Fatalf("expected one UserPromptSubmit hook, got %+v", hooks.Hooks)
	}
	h := entries[0].Hooks[0]
	if h.Type != "command" || !strings.HasPrefix(h.Command, "gyrus suggest-context --hook") || !strings.Contains(h.Command, "--max-tokens") || h.Timeout <= 0 {
		t.Errorf("unexpected prompt hook: %+v", h)
	}
}
