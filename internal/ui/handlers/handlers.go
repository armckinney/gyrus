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

// Handlers encapsulates dependencies for rendering Web UI views and managing chat sessions.
type Handlers struct {
	app          *app.App
	engine       *lifecycle.Engine
	templates    *templates.Manager
	renderer     *markdown.Renderer
	runner       agent.Runner
	sessionStore *agent.SessionStore
}

// New constructs a Handlers instance with the given dependencies.
func New(application *app.App, engine *lifecycle.Engine, tmpl *templates.Manager, renderer *markdown.Renderer) *Handlers {
	var runner agent.Runner
	var sessionStore *agent.SessionStore

	if application != nil {
		if application.Config() != nil {
			uiClient := application.Config().UIClient()
			uiCmd := application.Config().UICommand()
			if uiClient != "" || uiCmd != "" {
				runner = agent.NewRunner(uiClient, uiCmd, application.ProjectDir())
			}
		}
		persistPath := ""
		if application.ProjectDir() != "" {
			persistPath = fmt.Sprintf("%s/.gyrus/cache/ui_sessions.json", application.ProjectDir())
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

// isHTMX checks whether the request originates from an HTMX partial swap.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
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

// storageBackend returns the active storage engine name.
func (h *Handlers) storageBackend() string {
	if h.app != nil && h.app.Config() != nil && h.app.Config().Storage.Provider != "" {
		return h.app.Config().Storage.Provider
	}
	return "localfs"
}

func (h *Handlers) defaultTenant() string {
	if h.app != nil && h.app.Config() != nil && h.app.Config().Tenant != "" {
		return h.app.Config().Tenant
	}
	return "root"
}

func (h *Handlers) extractAvailableTenants(results []gyrus.SearchResult) []string {
	tenantMap := make(map[string]bool)
	def := h.defaultTenant()
	if def != "" {
		tenantMap[def] = true
	}

	for _, res := range results {
		if res.Concept.Tenant != "" {
			tenantMap[res.Concept.Tenant] = true
		}
	}

	if h.app != nil && h.app.ProjectDir() != "" {
		conceptsDir := filepath.Join(h.app.ProjectDir(), ".gyrus", "concepts")
		if entries, err := os.ReadDir(conceptsDir); err == nil {
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
	if h.app == nil || h.app.ProjectDir() == "" {
		return scopeMap
	}
	conceptsDir := filepath.Join(h.app.ProjectDir(), ".gyrus", "concepts")

	_ = filepath.Walk(conceptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(conceptsDir, path)
		if err != nil {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) >= 3 {
			id := strings.TrimSuffix(filepath.Base(path), ".md")
			if parts[1] == "reference" {
				scopeMap[id] = "reference"
			} else if parts[1] == "projects" && len(parts) >= 4 {
				scopeMap[id] = "projects/" + parts[2]
			}
		}
		return nil
	})
	return scopeMap
}

func (h *Handlers) determineDocScope(doc *gyrus.Concept, scopeMap map[string]string) string {
	if doc.Scope == "project" && doc.Project != "" {
		return "projects/" + doc.Project
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
		return "projects/gyrus"
	}
	return "reference"
}

func (h *Handlers) buildScopeData(results []gyrus.SearchResult, scopeMap map[string]string) ScopeNavData {
	var refCount int
	projMap := make(map[string]int)

	for _, res := range results {
		scope := h.determineDocScope(&res.Concept, scopeMap)
		if scope == "reference" {
			refCount++
		} else if strings.HasPrefix(scope, "projects/") {
			pName := strings.TrimPrefix(scope, "projects/")
			projMap[pName]++
		}
	}

	if h.app != nil && h.app.ProjectDir() != "" {
		conceptsDir := filepath.Join(h.app.ProjectDir(), ".gyrus", "concepts")

		if tenants, err := os.ReadDir(conceptsDir); err == nil {
			for _, tenant := range tenants {
				if !tenant.IsDir() {
					continue
				}
				pDir := filepath.Join(conceptsDir, tenant.Name(), "projects")
				if projs, err := os.ReadDir(pDir); err == nil {
					for _, p := range projs {
						if p.IsDir() && !strings.HasPrefix(p.Name(), ".") {
							if _, exists := projMap[p.Name()]; !exists {
								projMap[p.Name()] = 0
							}
						}
					}
				}
			}
		}
	}

	if len(projMap) == 0 {
		projMap["gyrus"] = 0
	}

	var projList []ProjectScopeCount
	for name, count := range projMap {
		projList = append(projList, ProjectScopeCount{Name: name, Count: count})
	}
	sort.Slice(projList, func(i, j int) bool {
		return projList[i].Name < projList[j].Name
	})

	return ScopeNavData{
		ReferenceCount: refCount,
		Projects:       projList,
	}
}

func (h *Handlers) buildTaxonomyCounts(results []gyrus.SearchResult) ([]NavCount, []NavCount) {
	categoryMap := make(map[string]int)
	typeMap := make(map[string]int)
	for _, res := range results {
		if res.Concept.Category != "" {
			categoryMap[string(res.Concept.Category)]++
		}
		if res.Concept.Type != "" {
			typeMap[string(res.Concept.Type)]++
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
