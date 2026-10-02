package agent_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/ui/agent"
)

func TestSessionStore_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	persistPath := filepath.Join(tempDir, "ui_sessions.json")

	store := agent.NewSessionStore(persistPath)

	// Create a session
	sess := store.CreateSession("test-1", "First Session", "gemini-3.8-flash-high", "high")
	if sess.ID != "test-1" || sess.Title != "First Session" {
		t.Fatalf("unexpected session: %+v", sess)
	}

	// Append a user message
	store.AppendMessage("test-1", agent.ChatMessage{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Can you explain the storage engine?",
		Timestamp: time.Now(),
	})

	// Append an agent message with actions
	store.AppendMessage("test-1", agent.ChatMessage{
		ID:      "msg-2",
		Role:    "agent",
		Content: "The storage engine supports localfs, sqlite, and postgres.",
		Actions: []agent.ActionItem{
			{
				ID:        "step-1",
				StepIndex: 1,
				ToolName:  "view_file",
				Summary:   "internal/provider/storage/localfs/localfs.go",
				Duration:  "0.04s",
				State:     "DONE",
			},
		},
		Timestamp: time.Now(),
	})

	// List sessions
	list := store.ListSessions()
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}
	if list[0].MessageCount != 2 {
		t.Errorf("expected 2 messages, got %d", list[0].MessageCount)
	}

	// Reload from disk into a new store instance
	reloaded := agent.NewSessionStore(persistPath)
	reloadedSess, ok := reloaded.GetSession("test-1")
	if !ok {
		t.Fatalf("failed to retrieve reloaded session")
	}
	if len(reloadedSess.Messages) != 2 {
		t.Fatalf("expected 2 reloaded messages, got %d", len(reloadedSess.Messages))
	}
	if len(reloadedSess.Messages[1].Actions) != 1 {
		t.Errorf("expected 1 action in agent message, got %d", len(reloadedSess.Messages[1].Actions))
	}
	if reloadedSess.Messages[1].Actions[0].ToolName != "view_file" {
		t.Errorf("unexpected action tool name: %s", reloadedSess.Messages[1].Actions[0].ToolName)
	}

	// Delete session
	deleted := reloaded.DeleteSession("test-1")
	if !deleted {
		t.Errorf("expected true on delete")
	}
	if len(reloaded.ListSessions()) != 0 {
		t.Errorf("expected 0 sessions after deletion")
	}

	// Create multiple and test DeleteAllSessions
	reloaded.CreateSession("test-2", "Second", "", "")
	reloaded.CreateSession("test-3", "Third", "", "")
	if len(reloaded.ListSessions()) != 2 {
		t.Fatalf("expected 2 sessions before DeleteAllSessions, got %d", len(reloaded.ListSessions()))
	}
	reloaded.DeleteAllSessions()
	if len(reloaded.ListSessions()) != 0 {
		t.Errorf("expected 0 sessions after DeleteAllSessions, got %d", len(reloaded.ListSessions()))
	}
}

