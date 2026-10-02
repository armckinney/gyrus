package handlers

import (
	"html/template"

	"github.com/armckinney/gyrus/internal/ui/agent"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// NavCount represents item counts for sidebar taxonomy navigation.
type NavCount struct {
	Name  string
	Count int
}

// WorkspaceScopeCount represents item counts for a specific workspace under Scopes.
type WorkspaceScopeCount struct {
	Name  string
	Count int
}

// ScopeNavData represents navigation hierarchy counts for Reference and Workspaces scopes.
type ScopeNavData struct {
	ReferenceCount int
	Workspaces     []WorkspaceScopeCount
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

// SearchResultItem represents an individual search result card.
type SearchResultItem struct {
	Document    gyrus.Document
	Snippet     string
	Score       float64
	MatchReason string
}

// BasePageData encapsulates common fields required across all page templates by layout.html.
type BasePageData struct {
	Title            string
	StorageBackend   string
	ActiveNav        string
	UIClient         string
	SearchQuery      string
	ActiveScope      string
	ActiveCategory   string
	ActiveType       string
	SelectedTenant   string
	AvailableTenants []string
	TotalDocs        int
	Scopes           ScopeNavData
	CategoryCounts   []NavCount
	TypeCounts       []NavCount
}

// DocsPageData represents view data for index.html (document list & filtered view).
type DocsPageData struct {
	BasePageData
	Documents      []DocListItem
	FilterTitle    string
	FilterSubtitle string
}

// DocPageData represents view data for doc.html (single document reader & lineage).
type DocPageData struct {
	BasePageData
	Doc           gyrus.Document
	ContentHTML   template.HTML
	OutgoingEdges []gyrus.DocumentEdge
	IncomingEdges []gyrus.DocumentEdge
}

// SearchPageData represents view data for search_results.html.
type SearchPageData struct {
	BasePageData
	Results      []SearchResultItem
	TotalMatches int
}

// GraphPageData represents view data for graph.html.
type GraphPageData struct {
	BasePageData
	FocusDocID string
}

// ModelOption represents an available model in the chat UI.
type ModelOption struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Efforts    []string `json:"efforts,omitempty"`
	EffortList string   `json:"effort_list,omitempty"`
}

// ChatPageData represents view data for chat.html.
type ChatPageData struct {
	BasePageData
	RunnerName      string
	RunnerAvailable bool
	RunnerBinary    string
	Sessions        []agent.SessionSummary
	Models          []ModelOption
}
