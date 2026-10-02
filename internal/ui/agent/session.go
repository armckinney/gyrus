package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ActionItem represents a structured action or tool executed by the agent during a turn.
type ActionItem struct {
	ID        string `json:"id"`
	StepIndex int    `json:"step_index"`
	ToolName  string `json:"tool_name"`
	Summary   string `json:"summary"`
	Command   string `json:"command,omitempty"`
	Output    string `json:"output,omitempty"`
	Duration  string `json:"duration,omitempty"`
	State     string `json:"state"` // "ACTIVE", "DONE", "ERROR"
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	ID        string       `json:"id"`
	Role      string       `json:"role"` // "user" or "agent"
	Content   string       `json:"content"`
	Actions   []ActionItem `json:"actions,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// ChatSession represents an entire conversation session.
type ChatSession struct {
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	ConversationID string        `json:"conversation_id,omitempty"`
	Model          string        `json:"model,omitempty"`
	Effort         string        `json:"effort,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Messages       []ChatMessage `json:"messages"`
}

// SessionSummary provides lightweight metadata for listing sessions in the UI.
type SessionSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
	Model        string    `json:"model,omitempty"`
	Effort       string    `json:"effort,omitempty"`
}

// SessionStore manages persistence and thread-safe operations on chat sessions.
type SessionStore struct {
	mu          sync.RWMutex
	sessions    map[string]*ChatSession
	persistPath string
}

// NewSessionStore initializes a SessionStore, optionally loading saved sessions from persistPath.
func NewSessionStore(persistPath string) *SessionStore {
	store := &SessionStore{
		sessions:    make(map[string]*ChatSession),
		persistPath: persistPath,
	}
	if persistPath != "" {
		_ = store.load()
	}
	return store
}

func (s *SessionStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.persistPath)
	if err != nil {
		return err
	}
	var loaded map[string]*ChatSession
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}
	if loaded != nil {
		s.sessions = loaded
	}
	return nil
}

func (s *SessionStore) persistLocked() error {
	if s.persistPath == "" {
		return nil
	}
	dir := filepath.Dir(s.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.persistPath, data, 0644)
}

// ListSessions returns all sessions sorted with the most recently updated first.
func (s *SessionStore) ListSessions() []SessionSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]SessionSummary, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, SessionSummary{
			ID:           sess.ID,
			Title:        sess.Title,
			CreatedAt:    sess.CreatedAt,
			UpdatedAt:    sess.UpdatedAt,
			MessageCount: len(sess.Messages),
			Model:        sess.Model,
			Effort:       sess.Effort,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	return list
}

// GetSession returns a session by ID if found.
func (s *SessionStore) GetSession(id string) (*ChatSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	sessCopy := *sess
	sessCopy.Messages = make([]ChatMessage, len(sess.Messages))
	copy(sessCopy.Messages, sess.Messages)
	return &sessCopy, true
}

// CreateSession explicitly creates a new session.
func (s *SessionStore) CreateSession(id, title, model, effort string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == "" {
		id = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	if title == "" {
		title = "New Chat"
	}
	now := time.Now()
	sess := &ChatSession{
		ID:        id,
		Title:     title,
		Model:     model,
		Effort:    effort,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []ChatMessage{},
	}
	s.sessions[id] = sess
	_ = s.persistLocked()
	return sess
}

// GetOrCreateSession retrieves an existing session or creates a new one.
func (s *SessionStore) GetOrCreateSession(id, model, effort string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[id]; ok {
		if model != "" {
			sess.Model = model
		}
		if effort != "" {
			sess.Effort = effort
		}
		return sess
	}

	if id == "" {
		id = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	now := time.Now()
	sess := &ChatSession{
		ID:        id,
		Title:     "New Chat",
		Model:     model,
		Effort:    effort,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []ChatMessage{},
	}
	s.sessions[id] = sess
	_ = s.persistLocked()
	return sess
}

// AppendMessage appends a message to the session and updates the session's timestamp and title.
func (s *SessionStore) AppendMessage(sessionID string, msg ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		now := time.Now()
		sess = &ChatSession{
			ID:        sessionID,
			Title:     "Chat",
			CreatedAt: now,
			UpdatedAt: now,
			Messages:  []ChatMessage{},
		}
		s.sessions[sessionID] = sess
	}

	if sess.Title == "New Chat" || sess.Title == "Chat" || sess.Title == "" {
		if msg.Role == "user" && msg.Content != "" {
			lines := strings.Split(msg.Content, "\n")
			firstLine := strings.TrimSpace(lines[0])
			runes := []rune(firstLine)
			if len(runes) > 42 {
				sess.Title = string(runes[:40]) + "..."
			} else if len(runes) > 0 {
				sess.Title = string(runes)
			}
		}
	}

	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()
	_ = s.persistLocked()
}

// SetConversationID records the underlying agent CLI conversation ID for a session.
func (s *SessionStore) SetConversationID(sessionID string, convID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[sessionID]; ok {
		sess.ConversationID = convID
		_ = s.persistLocked()
	}
}

// GetConversationID returns the underlying agent CLI conversation ID if recorded.
func (s *SessionStore) GetConversationID(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sess, ok := s.sessions[sessionID]; ok {
		return sess.ConversationID
	}
	return ""
}

// DeleteSession removes a session from memory and disk.
func (s *SessionStore) DeleteSession(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[sessionID]; ok {
		delete(s.sessions, sessionID)
		_ = s.persistLocked()
		return true
	}
	return false
}

// DeleteAllSessions clears all sessions from memory and disk.
func (s *SessionStore) DeleteAllSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = make(map[string]*ChatSession)
	_ = s.persistLocked()
}
