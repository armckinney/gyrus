package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/ui/agent"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Chat handles rendering the Agent Chat landing page view.
func (h *Handlers) Chat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}

	// Get total document count and taxonomy counts for sidebar navigation
	allSystemResults, err := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	if err != nil {
		allSystemResults = nil
	}
	availableTenants := h.extractAvailableTenants(allSystemResults)

	var tenantResults []gyrus.SearchResult
	if tenantFilter != "" {
		tenantResults, _ = h.engine.Search(ctx, "", gyrus.SearchFilter{OwnerGroup: tenantFilter})
	} else {
		tenantResults = allSystemResults
	}

	scopeMap := h.getDocScopeMap()
	scopeData := h.buildScopeData(tenantResults, scopeMap)
	categoryCounts, typeCounts := h.buildTaxonomyCounts(tenantResults)

	clientName := ""
	runnerName := ""
	runnerAvail := false
	binaryPath := ""
	if h.runner != nil {
		clientName = h.runner.ClientType()
		runnerName = h.runner.Name()
		runnerAvail = h.runner.IsAvailable()
		binaryPath = h.runner.BinaryPath()
	} else if h.app != nil && h.app.Config() != nil {
		clientName = h.app.Config().UIClient()
	}

	sessions := []agent.SessionSummary{}
	if h.sessionStore != nil {
		sessions = h.sessionStore.ListSessions()
	}

	models := h.getAvailableModels(clientName)

	data := ChatPageData{
		BasePageData: BasePageData{
			Title:            "Agent Chat",
			StorageBackend:   h.storageBackend(),
			ActiveNav:        "chat",
			UIClient:         clientName,
			SelectedTenant:   tenantFilter,
			AvailableTenants: availableTenants,
			TotalDocs:        len(tenantResults),
			Scopes:           scopeData,
			CategoryCounts:   categoryCounts,
			TypeCounts:       typeCounts,
		},
		RunnerName:      runnerName,
		RunnerAvailable: runnerAvail,
		RunnerBinary:    binaryPath,
		Sessions:        sessions,
		Models:          models,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "chat.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

func (h *Handlers) getAvailableModels(clientType string) []ModelOption {
	switch strings.ToLower(clientType) {
	case "antigravity", "agy":
		return []ModelOption{
			{ID: "default", Label: "Default (Auto)", Efforts: []string{"auto", "low", "medium", "high", "xhigh", "max"}, EffortList: "auto,low,medium,high,xhigh,max"},
			{ID: "gemini-3.8-flash", Label: "Gemini 3.8 Flash", Efforts: []string{"high", "medium", "low"}, EffortList: "high,medium,low"},
			{ID: "gemini-3.7-flash", Label: "Gemini 3.7 Flash", Efforts: []string{"high", "medium", "low"}, EffortList: "high,medium,low"},
			{ID: "gemini-3.1-pro", Label: "Gemini 3.1 Pro", Efforts: []string{"high", "low"}, EffortList: "high,low"},
			{ID: "claude-sonnet-4-6", Label: "Claude Sonnet 4.6", Efforts: nil, EffortList: "none"},
			{ID: "claude-opus-4-6-thinking", Label: "Claude Opus 4.6", Efforts: nil, EffortList: "none"},
			{ID: "gpt-oss-120b", Label: "GPT-OSS 120B", Efforts: []string{"medium"}, EffortList: "medium"},
		}
	case "claude", "claude-code":
		return []ModelOption{
			{ID: "default", Label: "Default (Auto)", Efforts: []string{"auto", "low", "medium", "high"}, EffortList: "auto,low,medium,high"},
			{ID: "claude-3-7-sonnet-latest", Label: "Claude 3.7 Sonnet", Efforts: []string{"auto", "low", "medium", "high"}, EffortList: "auto,low,medium,high"},
			{ID: "claude-3-5-sonnet-latest", Label: "Claude 3.5 Sonnet", Efforts: nil, EffortList: "none"},
			{ID: "claude-3-5-haiku-latest", Label: "Claude 3.5 Haiku", Efforts: nil, EffortList: "none"},
			{ID: "claude-3-opus-latest", Label: "Claude 3 Opus", Efforts: nil, EffortList: "none"},
		}
	default:
		return []ModelOption{
			{ID: "default", Label: "Default", Efforts: []string{"auto"}, EffortList: "auto"},
		}
	}
}

// ChatStreamPayload defines the JSON body for chat requests.
type ChatStreamPayload struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	Model     string `json:"model"`
	Effort    string `json:"effort"`
}

// ChatStream handles Server-Sent Events (SSE) streaming for agent turns.
func (h *Handlers) ChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var prompt string
	var sessionID string
	var model string
	var effort string

	if r.Header.Get("Content-Type") == "application/json" {
		var payload ChatStreamPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			prompt = strings.TrimSpace(payload.Message)
			sessionID = strings.TrimSpace(payload.SessionID)
			model = strings.TrimSpace(payload.Model)
			effort = strings.TrimSpace(payload.Effort)
		}
	} else {
		prompt = strings.TrimSpace(r.FormValue("message"))
		sessionID = strings.TrimSpace(r.FormValue("session_id"))
		model = strings.TrimSpace(r.FormValue("model"))
		effort = strings.TrimSpace(r.FormValue("effort"))
		if prompt == "" {
			prompt = strings.TrimSpace(r.URL.Query().Get("message"))
		}
		if sessionID == "" {
			sessionID = strings.TrimSpace(r.URL.Query().Get("session_id"))
		}
		if model == "" {
			model = strings.TrimSpace(r.URL.Query().Get("model"))
		}
		if effort == "" {
			effort = strings.TrimSpace(r.URL.Query().Get("effort"))
		}
	}

	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by client/server connection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	if h.runner == nil {
		writeSSE(w, flusher, "error", "No agent client configured in .gyrus.yaml (set 'ui.client')")
		return
	}

	if prompt == "" {
		writeSSE(w, flusher, "error", "Prompt message cannot be empty")
		return
	}

	if h.sessionStore != nil {
		_ = h.sessionStore.GetOrCreateSession(sessionID, model, effort)
		h.sessionStore.AppendMessage(sessionID, agent.ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "user",
			Content:   prompt,
			Timestamp: time.Now(),
		})
	}

	opts := agent.TurnOptions{
		SessionID: sessionID,
		Prompt:    prompt,
		Model:     model,
		Effort:    effort,
	}

	events := make(chan agent.StreamEvent, 100)
	ctx := r.Context()

	go func() {
		defer close(events)
		_ = h.runner.StreamTurn(ctx, opts, events)
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var accumulatedText strings.Builder
	actionsMap := make(map[int]*agent.ActionItem)
	var actionsOrder []int

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Turn finished: persist agent response
				if h.sessionStore != nil {
					collectedActions := make([]agent.ActionItem, 0, len(actionsOrder))
					for _, idx := range actionsOrder {
						if act, found := actionsMap[idx]; found && act != nil {
							collectedActions = append(collectedActions, *act)
						}
					}
					if accumulatedText.Len() > 0 || len(collectedActions) > 0 {
						h.sessionStore.AppendMessage(sessionID, agent.ChatMessage{
							ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
							Role:      "agent",
							Content:   accumulatedText.String(),
							Actions:   collectedActions,
							Timestamp: time.Now(),
						})
					}
				}
				return
			}

			if event.Type == "action" && event.Action != nil {
				idx := event.Action.StepIndex
				if _, exists := actionsMap[idx]; !exists {
					actionsOrder = append(actionsOrder, idx)
				}
				actionsMap[idx] = event.Action
			} else if event.Type == "token" {
				accumulatedText.WriteString(event.Data)
			}

			writeSSEEvent(w, flusher, event)

		case <-ticker.C:
			// Send standard SSE keepalive comment to prevent proxy/browser timeout
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

// APIChatSessions handles GET (list) and POST (create) sessions.
func (h *Handlers) APIChatSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.sessionStore == nil {
		http.Error(w, `{"error":"session store not configured"}`, http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sessions := h.sessionStore.ListSessions()
		_ = json.NewEncoder(w).Encode(sessions)
	case http.MethodPost:
		var req struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Model  string `json:"model"`
			Effort string `json:"effort"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		sess := h.sessionStore.CreateSession(req.ID, req.Title, req.Model, req.Effort)
		_ = json.NewEncoder(w).Encode(sess)
	case http.MethodDelete:
		h.sessionStore.DeleteAllSessions()
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "cleared": true})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// APIChatSessionDetail handles GET (retrieve) and DELETE (remove) for a specific session.
func (h *Handlers) APIChatSessionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.sessionStore == nil {
		http.Error(w, `{"error":"session store not configured"}`, http.StatusInternalServerError)
		return
	}

	// Extract session ID from /api/chat/sessions/{id}
	prefix := "/api/chat/sessions/"
	sessionID := strings.TrimPrefix(r.URL.Path, prefix)
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		http.Error(w, `{"error":"missing session id"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sess, ok := h.sessionStore.GetSession(sessionID)
		if !ok {
			http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(sess)
	case http.MethodDelete:
		deleted := h.sessionStore.DeleteSession(sessionID)
		if !deleted {
			http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "deleted": sessionID})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// APIChatModels returns available models for the active agent harness.
func (h *Handlers) APIChatModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientType := ""
	if h.runner != nil {
		clientType = h.runner.ClientType()
	}
	models := h.getAvailableModels(clientType)
	_ = json.NewEncoder(w).Encode(models)
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data string) {
	writeSSEEvent(w, flusher, agent.StreamEvent{
		Type: eventType,
		Data: data,
	})
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event agent.StreamEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
}
