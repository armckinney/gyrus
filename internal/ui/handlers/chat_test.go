package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/ui/agent"
	"github.com/armckinney/gyrus/internal/ui/handlers"
	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/templates"
)

type mockRunner struct {
	clientType string
	available  bool
	tokens     []string
	actions    []agent.ActionItem
}

func (m *mockRunner) Name() string { return "Mock Agent" }
func (m *mockRunner) ClientType() string {
	if m.clientType != "" {
		return m.clientType
	}
	return "antigravity"
}
func (m *mockRunner) IsAvailable() bool  { return m.available }
func (m *mockRunner) BinaryPath() string { return "/bin/mock" }
func (m *mockRunner) StreamTurn(ctx context.Context, opts agent.TurnOptions, out chan<- agent.StreamEvent) error {
	for _, act := range m.actions {
		out <- agent.StreamEvent{
			Type:   "action",
			Data:   act.Summary,
			Action: &act,
		}
	}
	for _, tok := range m.tokens {
		out <- agent.StreamEvent{Type: "token", Data: tok}
	}
	out <- agent.StreamEvent{Type: "done", Data: "[DONE]"}
	return nil
}

func setupTestHandlers(t *testing.T) (*handlers.Handlers, *app.App, *lifecycle.Engine) {
	tempWorkspace := t.TempDir()
	application, err := app.NewWithWorkspace(tempWorkspace)
	if err != nil {
		t.Fatalf("failed initializing app: %v", err)
	}

	engine, err := application.Engine()
	if err != nil {
		t.Fatalf("failed initializing engine: %v", err)
	}

	tmplMgr, err := templates.NewManager()
	if err != nil {
		t.Fatalf("failed initializing templates: %v", err)
	}

	renderer := markdown.NewRenderer()
	h := handlers.New(application, engine, tmplMgr, renderer)
	return h, application, engine
}

func TestChat_HandlerWithRunner(t *testing.T) {
	h, _, _ := setupTestHandlers(t)
	mock := &mockRunner{available: true, tokens: []string{"test"}}
	h.SetRunner(mock)

	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Agent Chat") {
		t.Errorf("expected response to contain 'Agent Chat', got: %s", body)
	}
	if !strings.Contains(body, "Mock Agent") {
		t.Errorf("expected response to contain 'Mock Agent', got: %s", body)
	}
}

func TestChat_HomeDelegation(t *testing.T) {
	h, _, _ := setupTestHandlers(t)

	// When no runner is set, Home delegates to Docs
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.Home(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "All Documents") {
		t.Errorf("expected 'All Documents' on default home page without runner")
	}

	// When runner is set, Home delegates to Chat
	mock := &mockRunner{available: true, tokens: []string{}}
	h.SetRunner(mock)

	reqChat := httptest.NewRequest(http.MethodGet, "/", nil)
	recChat := httptest.NewRecorder()
	h.Home(recChat, reqChat)
	if recChat.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recChat.Code)
	}
	if !strings.Contains(recChat.Body.String(), "Agent Chat") {
		t.Errorf("expected 'Agent Chat' on home page when runner is configured")
	}
}

func TestChatStream_SuccessAndPersistence(t *testing.T) {
	h, _, _ := setupTestHandlers(t)
	mock := &mockRunner{
		available: true,
		actions: []agent.ActionItem{
			{
				StepIndex: 1,
				ToolName:  "view_file",
				Summary:   "README.md",
				Duration:  "0.02s",
				State:     "DONE",
			},
		},
		tokens: []string{"Hello", " world!"},
	}
	h.SetRunner(mock)

	sessionID := "test-sess-stream-unique"
	h.SessionStore().DeleteSession(sessionID)

	payload := map[string]string{
		"message":    "test message",
		"session_id": sessionID,
		"model":      "gemini-3.8-flash",
		"effort":     "high",
	}
	jsonBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChatStream(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("expected Content-Type text/event-stream, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "view_file") {
		t.Errorf("expected stream to contain action event, got: %s", body)
	}
	if !strings.Contains(body, "Hello") || !strings.Contains(body, "world!") {
		t.Errorf("expected stream to contain tokens, got: %s", body)
	}
	if !strings.Contains(body, "[DONE]") {
		t.Errorf("expected stream to end with [DONE], got: %s", body)
	}

	// Verify session persistence in SessionStore
	sess, ok := h.SessionStore().GetSession(sessionID)
	if !ok {
		t.Fatalf("session %s was not persisted", sessionID)
	}
	if len(sess.Messages) != 2 {
		t.Fatalf("expected 2 messages (user + agent), got %d", len(sess.Messages))
	}
	if sess.Messages[0].Role != "user" || sess.Messages[0].Content != "test message" {
		t.Errorf("unexpected user message: %+v", sess.Messages[0])
	}
	if sess.Messages[1].Role != "agent" || sess.Messages[1].Content != "Hello world!" {
		t.Errorf("unexpected agent message: %+v", sess.Messages[1])
	}
	if len(sess.Messages[1].Actions) != 1 || sess.Messages[1].Actions[0].ToolName != "view_file" {
		t.Errorf("unexpected agent actions: %+v", sess.Messages[1].Actions)
	}
}

func TestChatSessions_APIs(t *testing.T) {
	h, _, _ := setupTestHandlers(t)

	// Create session via POST /api/chat/sessions
	createBody := bytes.NewReader([]byte(`{"id":"api-sess-1","title":"API Test Session","model":"gemini-3.8-flash-high"}`))
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/chat/sessions", createBody)
	recCreate := httptest.NewRecorder()
	h.APIChatSessions(recCreate, reqCreate)
	if recCreate.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recCreate.Code)
	}

	// List sessions via GET /api/chat/sessions
	reqList := httptest.NewRequest(http.MethodGet, "/api/chat/sessions", nil)
	recList := httptest.NewRecorder()
	h.APIChatSessions(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recList.Code)
	}
	if !strings.Contains(recList.Body.String(), "API Test Session") {
		t.Errorf("expected session list to contain 'API Test Session', got: %s", recList.Body.String())
	}

	// Get session detail via GET /api/chat/sessions/api-sess-1
	reqGet := httptest.NewRequest(http.MethodGet, "/api/chat/sessions/api-sess-1", nil)
	recGet := httptest.NewRecorder()
	h.APIChatSessionDetail(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recGet.Code)
	}
	if !strings.Contains(recGet.Body.String(), "api-sess-1") {
		t.Errorf("expected detail to contain api-sess-1, got: %s", recGet.Body.String())
	}

	// Delete session via DELETE /api/chat/sessions/api-sess-1
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/chat/sessions/api-sess-1", nil)
	recDel := httptest.NewRecorder()
	h.APIChatSessionDetail(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recDel.Code)
	}

	// Verify it's deleted
	reqGetAfter := httptest.NewRequest(http.MethodGet, "/api/chat/sessions/api-sess-1", nil)
	recGetAfter := httptest.NewRecorder()
	h.APIChatSessionDetail(recGetAfter, reqGetAfter)
	if recGetAfter.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", recGetAfter.Code)
	}

	// Create another session and test clearing all sessions via DELETE /api/chat/sessions
	h.SessionStore().CreateSession("api-sess-2", "Clear All Test", "", "")
	reqDelAll := httptest.NewRequest(http.MethodDelete, "/api/chat/sessions", nil)
	recDelAll := httptest.NewRecorder()
	h.APIChatSessions(recDelAll, reqDelAll)
	if recDelAll.Code != http.StatusOK {
		t.Fatalf("expected 200 on DELETE /api/chat/sessions, got %d", recDelAll.Code)
	}
	if len(h.SessionStore().ListSessions()) != 0 {
		t.Errorf("expected 0 sessions after clearing all")
	}
}

func TestChatModels_API(t *testing.T) {
	h, _, _ := setupTestHandlers(t)
	mock := &mockRunner{available: true}
	h.SetRunner(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/chat/models", nil)
	rec := httptest.NewRecorder()
	h.APIChatModels(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "gemini-3.8-flash") {
		t.Errorf("expected models response to contain gemini-3.8-flash, got: %s", body)
	}
	if strings.Contains(body, "gemini-3.8-flash-high") {
		t.Errorf("did not expect duplicate gemini-3.8-flash-high model in models list, got: %s", body)
	}
}
