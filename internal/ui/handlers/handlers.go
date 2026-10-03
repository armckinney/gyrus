package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/ui/agent"
	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/templates"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Handlers encapsulates HTTP request handlers for the Gyrus Web UI.
type Handlers struct {
	app          *app.App
	engine       *lifecycle.Engine
	templates    *templates.Manager
	renderer     *markdown.Renderer
	runner       agent.Runner
	sessionStore *agent.SessionStore
}

// New creates a new Handlers instance.
func New(application *app.App, engine *lifecycle.Engine, tmpl *templates.Manager, renderer *markdown.Renderer) *Handlers {
	var runner agent.Runner
	var sessionStore *agent.SessionStore

	if application != nil {
		if application.Config() != nil {
			uiClient := application.Config().UIClient()
			uiCmd := application.Config().UICommand()
			if uiClient != "" || uiCmd != "" {
				runner = agent.NewRunner(uiClient, uiCmd, application.WorkspaceDir())
			}
		}
		persistPath := ""
		if application.WorkspaceDir() != "" {
			persistPath = fmt.Sprintf("%s/.gyrus/cache/ui_sessions.json", application.WorkspaceDir())
		}
		sessionStore = agent.NewSessionStore(persistPath)
	} else {
		sessionStore = agent.NewSessionStore("")
	}

	return &Handlers{
		app:          application,
		engine:       engine,
		templates:    tmpl,
		renderer:     renderer,
		runner:       runner,
		sessionStore: sessionStore,
	}
}

// SessionStore returns the active agent session store.
func (h *Handlers) SessionStore() *agent.SessionStore {
	return h.sessionStore
}

// SetRunner overrides the agent runner (useful for tests).
func (h *Handlers) SetRunner(r agent.Runner) {
	h.runner = r
}

func (h *Handlers) uiClientName() string {
	if h.runner != nil {
		return h.runner.ClientType()
	}
	if h.app != nil && h.app.Config() != nil {
		return h.app.Config().UIClient()
	}
	return ""
}

// isHTMX checks whether the request originates from an HTMX partial swap.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// storageBackend returns the active storage engine name.
func (h *Handlers) storageBackend() string {
	if h.app != nil && h.app.Config() != nil && h.app.Config().Storage.Provider != "" {
		return h.app.Config().Storage.Provider
	}
	return "localfs"
}

func (h *Handlers) defaultOwnerGroup() string {
	if h.app != nil && h.app.Config() != nil && h.app.Config().OwnerGroup() != "" {
		return h.app.Config().OwnerGroup()
	}
	return "root"
}

func (h *Handlers) extractAvailableTenants(results []gyrus.SearchResult) []string {
	tenantMap := make(map[string]bool)
	def := h.defaultOwnerGroup()
	if def != "" {
		tenantMap[def] = true
	}

	for _, res := range results {
		if res.Document.OwnerGroup != "" {
			tenantMap[res.Document.OwnerGroup] = true
		}
	}

	if h.app != nil && h.app.WorkspaceDir() != "" {
		docsDir := filepath.Join(h.app.WorkspaceDir(), ".gyrus", "docs")
		if entries, err := os.ReadDir(docsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					tenantMap[entry.Name()] = true
				}
			}
		}
	}

	var tenants []string
	for t := range tenantMap {
		tenants = append(tenants, t)
	}
	sort.Strings(tenants)
	return tenants
}

func (h *Handlers) getDocScopeMap() map[string]string {
	scopeMap := make(map[string]string)
	if h.app == nil || h.app.WorkspaceDir() == "" {
		return scopeMap
	}
	docsDir := filepath.Join(h.app.WorkspaceDir(), ".gyrus", "docs")
	_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(docsDir, path)
		if err != nil {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) >= 3 {
			id := strings.TrimSuffix(filepath.Base(path), ".md")
			if parts[1] == "reference" {
				scopeMap[id] = "reference"
			} else if parts[1] == "workspaces" && len(parts) >= 4 {
				scopeMap[id] = "workspaces/" + parts[2]
			}
		}
		return nil
	})
	return scopeMap
}

func (h *Handlers) determineDocScope(doc *gyrus.Document, scopeMap map[string]string) string {
	if doc.Scope == "workspace" && doc.Workspace != "" {
		return "workspaces/" + doc.Workspace
	}
	if doc.Scope == "reference" {
		return "reference"
	}
	if scopeMap != nil {
		if s, ok := scopeMap[doc.ID]; ok && s != "" {
			return s
		}
	}
	if doc.Category == gyrus.CategoryBusinessLogic || doc.Category == gyrus.CategoryProduct {
		return "workspaces/gyrus"
	}
	return "reference"
}

func (h *Handlers) buildScopeData(results []gyrus.SearchResult, scopeMap map[string]string) ScopeNavData {
	var refCount int
	wsMap := make(map[string]int)

	for _, res := range results {
		scope := h.determineDocScope(&res.Document, scopeMap)
		if scope == "reference" {
			refCount++
		} else if strings.HasPrefix(scope, "workspaces/") {
			wsName := strings.TrimPrefix(scope, "workspaces/")
			wsMap[wsName]++
		}
	}

	if h.app != nil && h.app.WorkspaceDir() != "" {
		docsDir := filepath.Join(h.app.WorkspaceDir(), ".gyrus", "docs")
		if owners, err := os.ReadDir(docsDir); err == nil {
			for _, owner := range owners {
				if !owner.IsDir() {
					continue
				}
				wsDir := filepath.Join(docsDir, owner.Name(), "workspaces")
				if wss, err := os.ReadDir(wsDir); err == nil {
					for _, ws := range wss {
						if ws.IsDir() && !strings.HasPrefix(ws.Name(), ".") {
							if _, exists := wsMap[ws.Name()]; !exists {
								wsMap[ws.Name()] = 0
							}
						}
					}
				}
			}
		}
	}

	if len(wsMap) == 0 {
		wsMap["main"] = 0
	}

	var wsList []WorkspaceScopeCount
	for name, count := range wsMap {
		wsList = append(wsList, WorkspaceScopeCount{Name: name, Count: count})
	}
	sort.Slice(wsList, func(i, j int) bool {
		return wsList[i].Name < wsList[j].Name
	})

	return ScopeNavData{
		ReferenceCount: refCount,
		Workspaces:     wsList,
	}
}

func (h *Handlers) buildTaxonomyCounts(results []gyrus.SearchResult) ([]NavCount, []NavCount) {
	categoryMap := make(map[string]int)
	typeMap := make(map[string]int)
	for _, res := range results {
		if res.Document.Category != "" {
			categoryMap[string(res.Document.Category)]++
		}
		if res.Document.Type != "" {
			typeMap[string(res.Document.Type)]++
		}
	}

	var categoryCounts []NavCount
	for k, v := range categoryMap {
		categoryCounts = append(categoryCounts, NavCount{Name: k, Count: v})
	}
	sort.Slice(categoryCounts, func(i, j int) bool {
		return categoryCounts[i].Name < categoryCounts[j].Name
	})

	var typeCounts []NavCount
	for k, v := range typeMap {
		typeCounts = append(typeCounts, NavCount{Name: k, Count: v})
	}
	sort.Slice(typeCounts, func(i, j int) bool {
		return typeCounts[i].Name < typeCounts[j].Name
	})

	return categoryCounts, typeCounts
}

func extractSnippet(content string) string {
	lines := strings.Split(content, "\n")
	var textLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") {
			continue
		}
		textLines = append(textLines, trimmed)
		if len(strings.Join(textLines, " ")) > 160 {
			break
		}
	}
	res := strings.Join(textLines, " ")
	if len(res) > 160 {
		return res[:157] + "..."
	}
	return res
}
