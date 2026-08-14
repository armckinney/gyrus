package lifecycle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// TransitionError indicates an illegal state transition attempt.
type TransitionError struct {
	DocType       gyrus.DocumentType
	CurrentStatus string
	NewStatus     string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid lifecycle transition for type '%s': cannot transition from '%s' to '%s'", e.DocType, e.CurrentStatus, e.NewStatus)
}

// adrTransitions defines valid state transitions for ADR documents.
var adrTransitions = map[string]map[string]bool{
	"proposed": {
		"accepted": true,
		"rejected": true,
	},
	"accepted": {
		"superseded": true,
		"deprecated": true,
	},
}

// ipTransitions defines valid state transitions for Improvement Proposals.
var ipTransitions = map[string]map[string]bool{
	"draft": {
		"reviewing": true,
		"abandoned": true,
	},
	"reviewing": {
		"approved":  true,
		"rejected":  true,
		"abandoned": true,
	},
	"approved": {
		"implemented": true,
		"abandoned":   true,
	},
}

// generalTransitions defines default state transitions for standard document types.
var generalTransitions = map[string]map[string]bool{
	"draft": {
		"active":   true,
		"archived": true,
	},
	"active": {
		"deprecated": true,
		"archived":   true,
	},
	"deprecated": {
		"archived": true,
	},
}

// ValidateTransition verifies whether moving from currentStatus to newStatus is permitted for docType.
func ValidateTransition(docType gyrus.DocumentType, currentStatus, newStatus string) error {
	if currentStatus == newStatus {
		return nil
	}

	var transitions map[string]map[string]bool

	switch docType {
	case gyrus.TypeADR:
		transitions = adrTransitions
	case gyrus.TypeImprovementProposal:
		transitions = ipTransitions
	default:
		transitions = generalTransitions
	}

	allowedNewStatuses, exists := transitions[currentStatus]
	if !exists || !allowedNewStatuses[newStatus] {
		return &TransitionError{
			DocType:       docType,
			CurrentStatus: currentStatus,
			NewStatus:     newStatus,
		}
	}

	return nil
}

// ImmutabilityError indicates an illegal attempt to mutate content on a locked/immutable document.
type ImmutabilityError struct {
	DocType gyrus.DocumentType
	Status  string
}

func (e *ImmutabilityError) Error() string {
	return fmt.Sprintf("cannot modify content of immutable document type '%s' in status '%s'", e.DocType, e.Status)
}

// IsImmutableType returns true if the document type is an immutable historical decision log.
func IsImmutableType(docType gyrus.DocumentType) bool {
	switch docType {
	case gyrus.TypeADR, gyrus.TypeImprovementProposal, gyrus.TypeReleaseNote:
		return true
	default:
		return false
	}
}

// IsLockedStatus returns true if the document status represents a finalized/locked state.
func IsLockedStatus(docType gyrus.DocumentType, status string, isExplicitlyImmutable bool) bool {
	if isExplicitlyImmutable {
		return status != "draft" && status != "proposed"
	}
	switch status {
	case "accepted", "rejected", "superseded", "deprecated", "approved", "abandoned", "implemented", "published", "archived":
		return true
	default:
		return false
	}
}

// ValidateMutation verifies whether content modifications are permitted for the given docType, status, and immutability flag.
func ValidateMutation(docType gyrus.DocumentType, currentStatus string, isExplicitlyImmutable bool, contentChanged bool) error {
	if contentChanged && (isExplicitlyImmutable || IsImmutableType(docType)) && IsLockedStatus(docType, currentStatus, isExplicitlyImmutable) {
		return &ImmutabilityError{
			DocType: docType,
			Status:  currentStatus,
		}
	}
	return nil
}

// Engine orchestrates document lifecycle, indexing, search, and knowledge graph edge operations.
type Engine struct {
	store       gyrus.DocumentStore
	search      gyrus.SearchProvider
	indexer     gyrus.IndexStore
	graph       gyrus.GraphStore
	storageRoot string
}

// NewEngine constructs a new lifecycle Engine.
func NewEngine(store gyrus.DocumentStore, search gyrus.SearchProvider, indexer gyrus.IndexStore, graph gyrus.GraphStore, storageRoot string) *Engine {
	return &Engine{
		store:       store,
		search:      search,
		indexer:     indexer,
		graph:       graph,
		storageRoot: storageRoot,
	}
}

// Create validates and persists a new OKF contract document.
func (e *Engine) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	if err := okf.Validate(&doc); err != nil {
		return gyrus.DocumentRef{}, fmt.Errorf("document validation failed: %w", err)
	}

	ref, err := e.store.Create(ctx, doc)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if e.indexer != nil {
		_ = e.indexer.Index(ctx, doc)
	}

	return ref, nil
}

// Get retrieves a document by ID.
func (e *Engine) Get(ctx context.Context, id string) (gyrus.Document, error) {
	return e.store.Get(ctx, id)
}

// Update mutates an existing document.
func (e *Engine) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	existing, err := e.store.Get(ctx, id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	contentChanged := patch.Content != nil && *patch.Content != existing.Content
	if err := ValidateMutation(existing.Type, existing.Status, existing.Immutable, contentChanged); err != nil {
		return gyrus.DocumentRef{}, err
	}

	if patch.Status != nil {
		if err := ValidateTransition(existing.Type, existing.Status, *patch.Status); err != nil {
			return gyrus.DocumentRef{}, err
		}
	}

	ref, err := e.store.Update(ctx, id, patch, expectedVersion)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if e.indexer != nil {
		if updated, err := e.store.Get(ctx, id); err == nil {
			_ = e.indexer.Index(ctx, updated)
		}
	}

	return ref, nil
}

// Archive marks a document as archived.
func (e *Engine) Archive(ctx context.Context, id string) error {
	if err := e.store.Archive(ctx, id); err != nil {
		return err
	}
	if e.indexer != nil {
		_ = e.indexer.Remove(ctx, id)
	}
	return nil
}

// Search executes search over the configured search provider.
func (e *Engine) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	if e.search != nil {
		return e.search.Search(ctx, query, filter)
	}
	if e.indexer != nil {
		return e.indexer.Search(ctx, gyrus.SearchQuery{
			Query:  query,
			Filter: filter,
		})
	}
	return nil, fmt.Errorf("no search or index provider configured")
}

// SuggestContext analyzes a prompt and returns relevant context layer documents.
func (e *Engine) SuggestContext(ctx context.Context, prompt string, category string, maxDocs int) (string, error) {
	if maxDocs <= 0 {
		maxDocs = 5
	}

	filter := gyrus.SearchFilter{}
	if category != "" {
		filter.Category = gyrus.Category(category)
	}

	results, err := e.Search(ctx, prompt, filter)
	if err != nil {
		return "", err
	}

	if len(results) > maxDocs {
		results = results[:maxDocs]
	}

	var sb strings.Builder
	sb.WriteString("=== SUGGESTED CONTEXT LAYER ===\n\n")
	for _, res := range results {
		sb.WriteString(fmt.Sprintf("--- [%s] %s (%s/%s) ---\n", res.Document.ID, res.Document.Title, res.Document.Category, res.Document.Type))
		sb.WriteString(res.Document.Content)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}

// Link connects two documents with a directed relationship edge.
func (e *Engine) Link(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if e.graph == nil {
		return fmt.Errorf("no graph store configured")
	}

	if relType == "" {
		relType = gyrus.RelDependsOn
	}

	edge := gyrus.DocumentEdge{
		FromDocumentID:   fromID,
		ToDocumentID:     toID,
		RelationshipType: relType,
		CreatedBy:        "engine",
		CreatedAt:        time.Now().Truncate(time.Second),
	}

	return e.graph.UpsertEdges(ctx, []gyrus.DocumentEdge{edge})
}

// Unlink removes a directed relationship edge between two documents.
func (e *Engine) Unlink(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if e.graph == nil {
		return fmt.Errorf("no graph store configured")
	}

	if relType == "" {
		relType = gyrus.RelDependsOn
	}

	return e.graph.DeleteEdges(ctx, fromID, toID, relType)
}

// Neighbors retrieves neighboring relationship edges for a document.
func (e *Engine) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.DocumentEdge, error) {
	if e.graph == nil {
		return nil, fmt.Errorf("no graph store configured")
	}
	return e.graph.Neighbors(ctx, id, filter)
}

// Traverse executes a graph path search from a start document.
func (e *Engine) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if e.graph == nil {
		return nil, fmt.Errorf("no graph store configured")
	}
	return e.graph.Traverse(ctx, query)
}

// Sync scans storageRoot for Markdown files and re-indexes them.
func (e *Engine) Sync(ctx context.Context) (gyrus.SyncReport, error) {
	if e.indexer == nil {
		return gyrus.SyncReport{}, fmt.Errorf("no index store configured")
	}
	return e.indexer.Sync(ctx, e.storageRoot)
}
