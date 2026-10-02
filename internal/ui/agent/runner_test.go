package agent_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/ui/agent"
)

func TestNewRunner_Resolution(t *testing.T) {
	tests := []struct {
		clientType string
		customCmd  string
		expected   string
	}{
		{"antigravity", "", "antigravity"},
		{"agy", "", "antigravity"},
		{"claude", "", "claude"},
		{"claude-code", "", "claude"},
		{"codex", "", "codex"},
		{"copilot", "", "copilot"},
		{"custom", "echo hello", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.clientType, func(t *testing.T) {
			runner := agent.NewRunner(tt.clientType, tt.customCmd, ".")
			if runner.ClientType() != tt.expected {
				t.Errorf("expected client type %s, got %s", tt.expected, runner.ClientType())
			}
		})
	}
}

func TestAntigravityRunner_Availability(t *testing.T) {
	runner := agent.NewAntigravityRunner(".")
	if runner.Name() != "Antigravity" {
		t.Errorf("expected runner name 'Antigravity', got '%s'", runner.Name())
	}
	if runner.ClientType() != "antigravity" {
		t.Errorf("expected client type 'antigravity', got '%s'", runner.ClientType())
	}
	if !runner.IsAvailable() {
		t.Logf("agy CLI not found in test environment (expected in minimal containers)")
	} else {
		if runner.BinaryPath() == "" {
			t.Errorf("runner is marked available but BinaryPath is empty")
		}
	}
}

func TestClaudeRunner_Properties(t *testing.T) {
	runner := agent.NewClaudeRunner(".")
	if runner.Name() != "Claude Code" {
		t.Errorf("expected runner name 'Claude Code', got '%s'", runner.Name())
	}
	if runner.ClientType() != "claude" {
		t.Errorf("expected client type 'claude', got '%s'", runner.ClientType())
	}
	if !runner.IsAvailable() {
		t.Logf("claude CLI not found in test environment")
	} else {
		if runner.BinaryPath() == "" {
			t.Errorf("runner is marked available but BinaryPath is empty")
		}
	}
}

func TestGenericRunner_Execution(t *testing.T) {
	runner := agent.NewGenericRunner("Echo", "echo", "echo", ".")
	if !runner.IsAvailable() {
		t.Skip("echo command not found in PATH")
	}

	out := make(chan agent.StreamEvent, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := runner.StreamTurn(ctx, agent.TurnOptions{
		SessionID: "session-1",
		Prompt:    "test-prompt",
	}, out)
	if err != nil {
		t.Fatalf("StreamTurn failed: %v", err)
	}

	var collected strings.Builder
	hasDone := false
	for event := range out {
		if event.Type == "token" {
			collected.WriteString(event.Data)
		} else if event.Type == "done" {
			hasDone = true
			break
		}
	}

	if !hasDone {
		t.Errorf("expected 'done' event from StreamTurn")
	}
	if !strings.Contains(collected.String(), "test-prompt") {
		t.Errorf("expected output to contain 'test-prompt', got: %s", collected.String())
	}
}
