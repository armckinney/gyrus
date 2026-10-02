package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		tenantResults, _ = h.engine.Search(ctx, "", gyrus.SearchFilter{OwnerGroup: tenantFilter})
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

	results, err := h.engine.Search(ctx, "", gyrus.SearchFilter{OwnerGroup: tenantFilter})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load graph nodes: %v", err), http.StatusInternalServerError)
		return
	}

	payload := GraphDataJSON{
		Nodes: make([]GraphNodeJSON, 0, len(results)),
		Edges: make([]GraphEdgeJSON, 0),
	}

	nodeMap := make(map[string]bool)
	for _, res := range results {
		doc := res.Document
		nodeMap[doc.ID] = true

		payload.Nodes = append(payload.Nodes, GraphNodeJSON{
			ID:       doc.ID,
			Title:    doc.Title,
			Category: string(doc.Category),
			Type:     string(doc.Type),
			Status:   doc.Status,
		})
	}

	// Fetch all edges for the nodes
	for docID := range nodeMap {
		edges, err := h.engine.Neighbors(ctx, docID, gyrus.EdgeFilter{Direction: "outgoing"})
		if err == nil {
			for _, edge := range edges {
				// If target node does not exist in our active document index, inject an external ghost node
				if !nodeMap[edge.ToDocumentID] {
					nodeMap[edge.ToDocumentID] = true
					payload.Nodes = append(payload.Nodes, GraphNodeJSON{
						ID:       edge.ToDocumentID,
						Title:    edge.ToDocumentID,
						Category: "external",
						Type:     "reference",
						Status:   "external",
					})
				}

				payload.Edges = append(payload.Edges, GraphEdgeJSON{
					Source:           edge.FromDocumentID,
					Target:           edge.ToDocumentID,
					RelationshipType: string(edge.RelationshipType),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode graph data: %v", err), http.StatusInternalServerError)
	}
}
