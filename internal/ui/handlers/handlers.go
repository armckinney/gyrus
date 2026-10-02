package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/templates"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// NavCount represents item counts for sidebar taxonomy navigation.
type NavCount struct {
	Name  string
	Count int
}

// DocListItem represents a document summary displayed in lists and search results.
type DocListItem struct {
	ID       string
	Title    string
	Category string
	Type     string
	Status   string
	Version  int
	Snippet  string
}

// Handlers encapsulates HTTP request handlers for the Gyrus Web UI.
type Handlers struct {
	app       *app.App
	engine    *lifecycle.Engine
	templates *templates.Manager
	renderer  *markdown.Renderer
}

// New creates a new Handlers instance.
func New(application *app.App, engine *lifecycle.Engine, tmpl *templates.Manager, renderer *markdown.Renderer) *Handlers {
	return &Handlers{
		app:       application,
		engine:    engine,
		templates: tmpl,
		renderer:  renderer,
	}
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

// Home handles `/` and `/docs`, rendering the documentation list with taxonomy sidebar.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryFilter := r.URL.Query().Get("category")
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	filter := gyrus.SearchFilter{
		Category: gyrus.Category(categoryFilter),
		Type:     gyrus.DocumentType(typeFilter),
		Status:   statusFilter,
	}

	// Fetch all documents for calculating taxonomy statistics
	allResults, err := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	if err != nil {
		allResults = nil
	}

	categoryMap := make(map[string]int)
	typeMap := make(map[string]int)
	for _, res := range allResults {
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

	// Fetch filtered documents
	filteredResults, err := h.engine.Search(ctx, "", filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load documents: %v", err), http.StatusInternalServerError)
		return
	}

	var docs []DocListItem
	for _, res := range filteredResults {
		docs = append(docs, DocListItem{
			ID:       res.Document.ID,
			Title:    res.Document.Title,
			Category: string(res.Document.Category),
			Type:     string(res.Document.Type),
			Status:   res.Document.Status,
			Version:  res.Document.Version,
			Snippet:  extractSnippet(res.Document.Content),
		})
	}

	filterTitle := "Open Knowledge Architecture"
	filterSubtitle := "Explore ADRs, specifications, and governance contracts managed by Gyrus."
	if categoryFilter != "" {
		filterTitle = fmt.Sprintf("Category: %s", categoryFilter)
		filterSubtitle = fmt.Sprintf("Documents categorized under %s.", categoryFilter)
	} else if typeFilter != "" {
		filterTitle = fmt.Sprintf("Type: %s", typeFilter)
		filterSubtitle = fmt.Sprintf("Documents of type %s.", typeFilter)
	}

	data := map[string]any{
		"Title":          "Documents",
		"StorageBackend": h.storageBackend(),
		"ActiveNav":      "all",
		"ActiveCategory": categoryFilter,
		"ActiveType":     typeFilter,
		"TotalDocs":      len(allResults),
		"CategoryCounts": categoryCounts,
		"TypeCounts":     typeCounts,
		"Documents":      docs,
		"FilterTitle":    filterTitle,
		"FilterSubtitle": filterSubtitle,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "index.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// Doc handles `GET /docs/{id}`, rendering rendered markdown with frontmatter badges and edge relationships.
func (h *Handlers) Doc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := strings.TrimPrefix(r.URL.Path, "/docs/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.Redirect(w, r, "/docs", http.StatusFound)
		return
	}

	doc, err := h.engine.Get(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Document '%s' not found: %v", id, err), http.StatusNotFound)
		return
	}

	// Fetch neighbor relationships
	outgoingEdges, _ := h.engine.Neighbors(ctx, id, gyrus.EdgeFilter{Direction: "outgoing"})
	incomingEdges, _ := h.engine.Neighbors(ctx, id, gyrus.EdgeFilter{Direction: "incoming"})

	renderedHTML, err := h.renderer.Render(doc.Content)
	if err != nil {
		renderedHTML = template.HTML(fmt.Sprintf("<pre>%s</pre>", template.HTMLEscapeString(doc.Content)))
	}

	// Fetch total doc counts for sidebar
	allResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	categoryCounts, typeCounts := h.buildTaxonomyCounts(allResults)

	data := map[string]any{
		"Title":          doc.Title,
		"StorageBackend": h.storageBackend(),
		"ActiveNav":      "all",
		"ActiveCategory": string(doc.Category),
		"ActiveType":     string(doc.Type),
		"TotalDocs":      len(allResults),
		"CategoryCounts": categoryCounts,
		"TypeCounts":     typeCounts,
		"Doc":            doc,
		"ContentHTML":    renderedHTML,
		"OutgoingEdges":  outgoingEdges,
		"IncomingEdges":  incomingEdges,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "doc.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// Search handles `GET /search`, executing search queries and rendering results.
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	allResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	categoryCounts, typeCounts := h.buildTaxonomyCounts(allResults)

	type SearchResultView struct {
		Document    gyrus.Document
		Snippet     string
		Score       float64
		MatchReason string
	}

	var resultsView []SearchResultView
	if query != "" {
		results, err := h.engine.Search(ctx, query, gyrus.SearchFilter{})
		if err == nil {
			for _, r := range results {
				resultsView = append(resultsView, SearchResultView{
					Document:    r.Document,
					Snippet:     extractSnippet(r.Document.Content),
					Score:       r.Score,
					MatchReason: r.MatchReason,
				})
			}
		}
	} else {
		// If query is blank, redirect back to docs or list docs
		for _, r := range allResults {
			resultsView = append(resultsView, SearchResultView{
				Document: r.Document,
				Snippet:  extractSnippet(r.Document.Content),
			})
		}
	}

	data := map[string]any{
		"Title":          fmt.Sprintf("Search: %s", query),
		"StorageBackend": h.storageBackend(),
		"SearchQuery":    query,
		"TotalDocs":      len(allResults),
		"CategoryCounts": categoryCounts,
		"TypeCounts":     typeCounts,
		"Results":        resultsView,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "search_results.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// Graph handles `GET /graph`, rendering the full Cytoscape interactive canvas.
func (h *Handlers) Graph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	focusID := strings.TrimSpace(r.URL.Query().Get("focus"))

	allResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	categoryCounts, typeCounts := h.buildTaxonomyCounts(allResults)

	data := map[string]any{
		"Title":          "Knowledge Graph",
		"StorageBackend": h.storageBackend(),
		"ActiveNav":      "graph",
		"TotalDocs":      len(allResults),
		"CategoryCounts": categoryCounts,
		"TypeCounts":     typeCounts,
		"FocusDocID":     focusID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "graph.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// GraphNodeJSON represents a node formatted for Cytoscape.js.
type GraphNodeJSON struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Status   string `json:"status"`
}

// GraphEdgeJSON represents a directed relationship edge for Cytoscape.js.
type GraphEdgeJSON struct {
	Source           string `json:"source"`
	Target           string `json:"target"`
	RelationshipType string `json:"relationship_type"`
}

// GraphDataJSON represents the complete topology response payload.
type GraphDataJSON struct {
	Nodes []GraphNodeJSON `json:"nodes"`
	Edges []GraphEdgeJSON `json:"edges"`
}

// APIGraph handles `GET /api/graph`, outputting the graph topology as JSON.
func (h *Handlers) APIGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	docs, err := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve documents: %v", err), http.StatusInternalServerError)
		return
	}

	var nodes []GraphNodeJSON
	var edges []GraphEdgeJSON
	edgeSet := make(map[string]bool)
	nodeMap := make(map[string]bool)

	for _, d := range docs {
		nodeMap[d.Document.ID] = true
		nodes = append(nodes, GraphNodeJSON{
			ID:       d.Document.ID,
			Title:    d.Document.Title,
			Category: string(d.Document.Category),
			Type:     string(d.Document.Type),
			Status:   d.Document.Status,
		})
	}

	for _, d := range docs {
		// Collect outgoing edges
		neighbors, err := h.engine.Neighbors(ctx, d.Document.ID, gyrus.EdgeFilter{Direction: "outgoing"})
		if err == nil {
			for _, edge := range neighbors {
				edgeKey := fmt.Sprintf("%s->%s:%s", edge.FromDocumentID, edge.ToDocumentID, edge.RelationshipType)
				if !edgeSet[edgeKey] {
					edgeSet[edgeKey] = true
					edges = append(edges, GraphEdgeJSON{
						Source:           edge.FromDocumentID,
						Target:           edge.ToDocumentID,
						RelationshipType: string(edge.RelationshipType),
					})
					if !nodeMap[edge.ToDocumentID] {
						nodeMap[edge.ToDocumentID] = true
						nodes = append(nodes, GraphNodeJSON{
							ID:       edge.ToDocumentID,
							Title:    edge.ToDocumentID,
							Category: "external",
							Type:     "reference",
							Status:   "external",
						})
					}
				}
			}
		}
	}


	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(GraphDataJSON{
		Nodes: nodes,
		Edges: edges,
	})
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
