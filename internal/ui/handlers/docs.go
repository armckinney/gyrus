package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Home handles `/`. If an agent client is configured, it renders the chat landing page.
// Otherwise, it delegates to Docs to display the document explorer.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if h.uiClientName() != "" {
		h.Chat(w, r)
		return
	}
	h.Docs(w, r)
}

// Docs handles `/docs`, rendering the documentation list with taxonomy sidebar.
func (h *Handlers) Docs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}
	scopeFilter := strings.TrimSpace(r.URL.Query().Get("scope"))
	categoryFilter := r.URL.Query().Get("category")
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	// Fetch all documents across system to discover available tenants
	allSystemResults, err := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	if err != nil {
		allSystemResults = nil
	}
	availableTenants := h.extractAvailableTenants(allSystemResults)

	// Fetch documents for active tenant (or all if empty)
	var tenantResults []gyrus.SearchResult
	if tenantFilter != "" {
		tenantResults, err = h.engine.Search(ctx, "", gyrus.SearchFilter{OwnerGroup: tenantFilter})
		if err != nil {
			tenantResults = nil
		}
	} else {
		tenantResults = allSystemResults
	}

	scopeMap := h.getDocScopeMap()
	scopeData := h.buildScopeData(tenantResults, scopeMap)

	// First pass: filter by scope only, to compute type counts within the active scope
	var scopeFilteredResults []gyrus.SearchResult
	for _, res := range tenantResults {
		if scopeFilter != "" {
			docScope := h.determineDocScope(&res.Document, scopeMap)
			if scopeFilter == "reference" {
				if docScope != "reference" {
					continue
				}
			} else if strings.HasPrefix(scopeFilter, "workspaces/") {
				if docScope != scopeFilter {
					continue
				}
			} else {
				if docScope != "workspaces/"+scopeFilter && docScope != scopeFilter {
					continue
				}
			}
		}
		scopeFilteredResults = append(scopeFilteredResults, res)
	}

	categoryCounts, typeCounts := h.buildTaxonomyCounts(scopeFilteredResults)

	// Second pass: apply category, type, and status filters on top of scope-filtered results
	var filteredResults []gyrus.SearchResult
	for _, res := range scopeFilteredResults {
		if categoryFilter != "" && string(res.Document.Category) != categoryFilter {
			continue
		}
		if typeFilter != "" && string(res.Document.Type) != typeFilter {
			continue
		}
		if statusFilter != "" && res.Document.Status != statusFilter {
			continue
		}
		filteredResults = append(filteredResults, res)
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

	filterTitle := "All Documents"
	filterSubtitle := "Explore ADRs, specifications, and governance contracts managed by Gyrus."
	if scopeFilter == "reference" {
		if typeFilter != "" {
			filterTitle = fmt.Sprintf("Scope: Reference · %s", typeFilter)
			filterSubtitle = fmt.Sprintf("Reference documents of type '%s'.", typeFilter)
		} else {
			filterTitle = "Scope: Reference"
			filterSubtitle = "Global reference documents, architecture standards, and specifications."
		}
	} else if scopeFilter != "" {
		wsName := strings.TrimPrefix(scopeFilter, "workspaces/")
		if typeFilter != "" {
			filterTitle = fmt.Sprintf("Workspace: %s · %s", wsName, typeFilter)
			filterSubtitle = fmt.Sprintf("Documents of type '%s' scoped to workspace '%s'.", typeFilter, wsName)
		} else {
			filterTitle = fmt.Sprintf("Workspace: %s", wsName)
			filterSubtitle = fmt.Sprintf("Documents and tickets scoped to workspace '%s'.", wsName)
		}
	} else if categoryFilter != "" {
		filterTitle = fmt.Sprintf("Category: %s", categoryFilter)
		filterSubtitle = fmt.Sprintf("Documents categorized under %s.", categoryFilter)
	} else if typeFilter != "" {
		filterTitle = fmt.Sprintf("Type: %s", typeFilter)
		filterSubtitle = fmt.Sprintf("Documents of type %s.", typeFilter)
	}

	data := DocsPageData{
		BasePageData: BasePageData{
			Title:            "Documents",
			StorageBackend:   h.storageBackend(),
			ActiveNav:        "docs",
			UIClient:         h.uiClientName(),
			ActiveScope:      scopeFilter,
			ActiveCategory:   categoryFilter,
			ActiveType:       typeFilter,
			SelectedTenant:   tenantFilter,
			AvailableTenants: availableTenants,
			TotalDocs:        len(tenantResults),
			Scopes:           scopeData,
			CategoryCounts:   categoryCounts,
			TypeCounts:       typeCounts,
		},
		Documents:      docs,
		FilterTitle:    filterTitle,
		FilterSubtitle: filterSubtitle,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "index.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// Doc handles `GET /docs/{id}`, rendering markdown with frontmatter badges and edge relationships.
func (h *Handlers) Doc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}

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

	// Fetch documents for stats and sidebar
	allSystemResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
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

	data := DocPageData{
		BasePageData: BasePageData{
			Title:            doc.Title,
			StorageBackend:   h.storageBackend(),
			ActiveNav:        "docs",
			UIClient:         h.uiClientName(),
			SelectedTenant:   tenantFilter,
			AvailableTenants: availableTenants,
			TotalDocs:        len(tenantResults),
			Scopes:           scopeData,
			CategoryCounts:   categoryCounts,
			TypeCounts:       typeCounts,
		},
		Doc:           doc,
		ContentHTML:   renderedHTML,
		OutgoingEdges: outgoingEdges,
		IncomingEdges: incomingEdges,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "doc.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// Search handles `GET /search?q={query}`, displaying ranked matches with snippets and badges.
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}
	scopeFilter := strings.TrimSpace(r.URL.Query().Get("scope"))

	allSystemResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
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

	var resultsView []SearchResultItem
	if query != "" {
		searchFilter := gyrus.SearchFilter{OwnerGroup: tenantFilter}
		results, err := h.engine.Search(ctx, query, searchFilter)
		if err == nil {
			for _, res := range results {
				if scopeFilter != "" {
					docScope := h.determineDocScope(&res.Document, scopeMap)
					if scopeFilter == "reference" && docScope != "reference" {
						continue
					}
					if strings.HasPrefix(scopeFilter, "workspaces/") && docScope != scopeFilter {
						continue
					}
				}
				resultsView = append(resultsView, SearchResultItem{
					Document:    res.Document,
					Snippet:     extractSnippet(res.Document.Content),
					Score:       res.Score,
					MatchReason: res.MatchReason,
				})
			}
		}
	} else {
		for _, res := range tenantResults {
			if scopeFilter != "" {
				docScope := h.determineDocScope(&res.Document, scopeMap)
				if scopeFilter == "reference" && docScope != "reference" {
					continue
				}
				if strings.HasPrefix(scopeFilter, "workspaces/") && docScope != scopeFilter {
					continue
				}
			}
			resultsView = append(resultsView, SearchResultItem{
				Document: res.Document,
				Snippet:  extractSnippet(res.Document.Content),
			})
		}
	}

	data := SearchPageData{
		BasePageData: BasePageData{
			Title:            fmt.Sprintf("Search: %s", query),
			StorageBackend:   h.storageBackend(),
			UIClient:         h.uiClientName(),
			SearchQuery:      query,
			ActiveScope:      scopeFilter,
			SelectedTenant:   tenantFilter,
			AvailableTenants: availableTenants,
			TotalDocs:        len(tenantResults),
			Scopes:           scopeData,
			CategoryCounts:   categoryCounts,
			TypeCounts:       typeCounts,
		},
		Results:      resultsView,
		TotalMatches: len(resultsView),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "search_results.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}
