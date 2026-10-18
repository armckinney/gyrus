package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

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

// GraphDataJSON represents the payload returned by `/api/graph`.
type GraphDataJSON struct {
	Nodes []GraphNodeJSON `json:"nodes"`
	Edges []GraphEdgeJSON `json:"edges"`
}

// Graph handles `GET /graph`, rendering the full Cytoscape interactive canvas.
func (h *Handlers) Graph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	focusID := strings.TrimSpace(r.URL.Query().Get("focus"))
	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}

	allSystemResults, _ := h.engine.Search(ctx, "", gyrus.SearchFilter{})
	availableTenants := h.extractAvailableTenants(allSystemResults)

	var tenantResults []gyrus.SearchResult
	if tenantFilter != "" {
		tenantResults, _ = h.engine.Search(ctx, "", gyrus.SearchFilter{Tenant: tenantFilter})
	} else {
		tenantResults = allSystemResults
	}

	scopeMap := h.getDocScopeMap()
	scopeData := h.buildScopeData(tenantResults, scopeMap)
	categoryCounts, typeCounts := h.buildTaxonomyCounts(tenantResults)

	data := GraphPageData{
		BasePageData: BasePageData{
			Title:            "Knowledge Graph",
			StorageBackend:   h.storageBackend(),
			ActiveNav:        "graph",
			UIClient:         h.uiClientName(),
			SelectedTenant:   tenantFilter,
			AvailableTenants: availableTenants,
			TotalDocs:        len(tenantResults),
			Scopes:           scopeData,
			CategoryCounts:   categoryCounts,
			TypeCounts:       typeCounts,
		},
		FocusDocID: focusID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Render(w, "graph.html", data, isHTMX(r)); err != nil {
		http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
	}
}

// APIGraph handles `GET /api/graph`, providing node and edge elements for Cytoscape.js.
func (h *Handlers) APIGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantFilter := strings.TrimSpace(r.URL.Query().Get("tenant"))
	if tenantFilter == "" {
		tenantFilter = strings.TrimSpace(r.URL.Query().Get("owner"))
	}

	results, err := h.engine.Search(ctx, "", gyrus.SearchFilter{Tenant: tenantFilter})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load graph nodes: %v", err), http.StatusInternalServerError)
		return
	}

	payload := GraphDataJSON{
		Nodes: make([]GraphNodeJSON, 0, len(results)),
		Edges: make([]GraphEdgeJSON, 0),
	}

	nodeMap := make(map[string]bool)
	sources := make([]string, 0, len(results))
	for _, res := range results {
		doc := res.Concept
		if nodeMap[doc.Name] {
			continue
		}
		nodeMap[doc.Name] = true
		sources = append(sources, doc.Name)

		payload.Nodes = append(payload.Nodes, GraphNodeJSON{
			ID:       doc.Name,
			Title:    doc.Title,
			Category: string(doc.Category),
			Type:     string(doc.Type),
			Status:   doc.Status,
		})
	}

	sort.Strings(sources)
	type edgeKey struct {
		source, target string
		rel            gyrus.RelationshipType
	}
	seenEdges := make(map[edgeKey]bool)
	for _, docID := range sources {
		edges, err := h.engine.Neighbors(ctx, docID, gyrus.EdgeFilter{Direction: "outgoing"})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load graph edges: %v", err), http.StatusInternalServerError)
			return
		}
		for _, edge := range edges {
			key := edgeKey{edge.FromConceptID, edge.ToConceptID, edge.RelationshipType}
			if seenEdges[key] {
				continue
			}
			seenEdges[key] = true
			if !nodeMap[edge.ToConceptID] {
				nodeMap[edge.ToConceptID] = true
				payload.Nodes = append(payload.Nodes, GraphNodeJSON{
					ID: edge.ToConceptID, Title: edge.ToConceptID,
					Category: "external", Type: "reference", Status: "external",
				})
			}
			payload.Edges = append(payload.Edges, GraphEdgeJSON{
				Source: edge.FromConceptID, Target: edge.ToConceptID,
				RelationshipType: string(edge.RelationshipType),
			})
		}
	}
	sort.Slice(payload.Nodes, func(i, j int) bool { return payload.Nodes[i].ID < payload.Nodes[j].ID })
	sort.Slice(payload.Edges, func(i, j int) bool {
		a, b := payload.Edges[i], payload.Edges[j]
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.Target != b.Target {
			return a.Target < b.Target
		}
		return a.RelationshipType < b.RelationshipType
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode graph data: %v", err), http.StatusInternalServerError)
	}
}
