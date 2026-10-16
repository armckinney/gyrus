package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// TransitionError indicates an illegal state transition attempt.
type TransitionError struct {
	DocType       gyrus.ConceptType
	CurrentStatus string
	NewStatus     string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid lifecycle transition for type '%s': cannot transition from '%s' to '%s'", e.DocType, e.CurrentStatus, e.NewStatus)
}

// adrTransitions defines valid state transitions for ADR concepts.
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

// generalTransitions defines default state transitions for standard concept types.
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
func ValidateTransition(docType gyrus.ConceptType, currentStatus, newStatus string) error {
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

// ImmutabilityError indicates an illegal attempt to mutate content on a locked/immutable concept.
type ImmutabilityError struct {
	DocType gyrus.ConceptType
	Status  string
}

func (e *ImmutabilityError) Error() string {
	return fmt.Sprintf("cannot modify content of immutable concept type '%s' in status '%s'", e.DocType, e.Status)
}

// IsImmutableType returns true if the concept type is an immutable historical decision log.
func IsImmutableType(docType gyrus.ConceptType) bool {
	switch docType {
	case gyrus.TypeADR, gyrus.TypeImprovementProposal, gyrus.TypeReleaseNote:
		return true
	default:
		return false
	}
}

// IsLockedStatus returns true if the concept status represents a finalized/locked state.
func IsLockedStatus(docType gyrus.ConceptType, status string, isExplicitlyImmutable bool) bool {
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
func ValidateMutation(docType gyrus.ConceptType, currentStatus string, isExplicitlyImmutable bool, contentChanged bool) error {
	if contentChanged && (isExplicitlyImmutable || IsImmutableType(docType)) && IsLockedStatus(docType, currentStatus, isExplicitlyImmutable) {
		return &ImmutabilityError{
			DocType: docType,
			Status:  currentStatus,
		}
	}
	return nil
}

// Engine orchestrates concept lifecycle, indexing, search, and knowledge graph edge operations.
type Engine struct {
	store          gyrus.ConceptStore
	schemaStore    gyrus.SchemaStore
	search         gyrus.SearchProvider
	indexer        gyrus.IndexStore
	graph          gyrus.GraphStore
	defaultProject string
}

// NewEngine constructs a new lifecycle Engine.
func NewEngine(store gyrus.ConceptStore, search gyrus.SearchProvider, indexer gyrus.IndexStore, graph gyrus.GraphStore, storageRoot string) *Engine {
	return NewEngineWithProject(store, search, indexer, graph, storageRoot, "")
}

// NewEngineWithProject constructs a new lifecycle Engine with a default project name.
func NewEngineWithProject(store gyrus.ConceptStore, search gyrus.SearchProvider, indexer gyrus.IndexStore, graph gyrus.GraphStore, storageRoot string, defaultProject string) *Engine {
	var schemaStore gyrus.SchemaStore
	if ss, ok := store.(gyrus.SchemaStore); ok {
		schemaStore = ss
	}

	return &Engine{
		store:          store,
		schemaStore:    schemaStore,
		search:         search,
		indexer:        indexer,
		graph:          graph,
		defaultProject: defaultProject,
	}
}

// DefaultProject returns the engine's default project name.
func (e *Engine) DefaultProject() string {
	return e.defaultProject
}

// Create validates and persists a new OKF contract concept.
func (e *Engine) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	if doc.Project == "" && doc.Scope != "reference" && e.defaultProject != "" {
		doc.Project = e.defaultProject
		doc.Scope = "project"
	}

	if err := okf.SetName(&doc, e.defaultProject); err != nil {
		return gyrus.ConceptRef{}, err
	}
	if err := okf.Validate(&doc); err != nil {
		return gyrus.ConceptRef{}, fmt.Errorf("concept validation failed: %w", err)
	}

	ref, err := e.store.Create(ctx, doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	if e.indexer != nil {
		_ = e.indexer.Index(ctx, doc)
	}

	return ref, nil
}

// Get retrieves a concept by ID.
func (e *Engine) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	if err := okf.ValidateName(id); err != nil {
		return gyrus.Concept{}, err
	}

	return e.store.Get(ctx, id)
}

// Update mutates an existing concept.
func (e *Engine) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	if err := okf.ValidateName(id); err != nil {
		return gyrus.ConceptRef{}, err
	}

	existing, err := e.store.Get(ctx, id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	contentChanged := patch.Content != nil && *patch.Content != existing.Content
	if err := ValidateMutation(existing.Type, existing.Status, existing.Immutable, contentChanged); err != nil {
		return gyrus.ConceptRef{}, err
	}

	if patch.Status != nil {
		if err := ValidateTransition(existing.Type, existing.Status, *patch.Status); err != nil {
			return gyrus.ConceptRef{}, err
		}
	}

	ref, err := e.store.Update(ctx, id, patch, expectedVersion)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	if e.indexer != nil {
		if updated, err := e.store.Get(ctx, id); err == nil {
			_ = e.indexer.Index(ctx, updated)
		}
	}

	return ref, nil
}

// Archive marks a concept as archived.
func (e *Engine) Archive(ctx context.Context, id string) error {
	if err := okf.ValidateName(id); err != nil {
		return err
	}

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
	return e.SearchQuery(ctx, gyrus.SearchQuery{Query: query, Filter: filter})
}

func (e *Engine) SearchQuery(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	if query.Filter.Project == "" && query.Filter.Scope != "all" && query.Filter.Scope != "reference" && e.defaultProject != "" {
		query.Filter.Project = e.defaultProject
	}

	if e.search != nil {
		return e.search.Search(ctx, query)
	}
	if e.indexer != nil {
		return e.indexer.Search(ctx, query)
	}
	return nil, fmt.Errorf("no search or index provider configured")
}

// SuggestContext analyzes a prompt and returns relevant context layer concepts.
func (e *Engine) SuggestContext(ctx context.Context, prompt string, category string, maxConcepts int) (string, error) {
	filter := gyrus.SearchFilter{}
	if category != "" {
		filter.Category = gyrus.Category(category)
	}
	return e.SuggestContextWithFilter(ctx, prompt, filter, maxConcepts)
}

// SuggestContextWithFilter analyzes a prompt with explicit search filter constraints, prioritizes project concepts,
// expands linked reference dependencies, and returns a formatted context layer.
func (e *Engine) SuggestContextWithFilter(ctx context.Context, prompt string, filter gyrus.SearchFilter, maxConcepts int) (string, error) {
	if maxConcepts <= 0 {
		maxConcepts = 5
	}

	if filter.Project == "" && filter.Scope != "all" && filter.Scope != "reference" && e.defaultProject != "" {
		filter.Project = e.defaultProject
	}

	results, err := e.SearchQuery(ctx, gyrus.SearchQuery{Query: prompt, Filter: filter, MaxResults: maxConcepts})
	if err != nil {
		return "", err
	}

	seenIDs := make(map[string]bool)
	var projectDocs []gyrus.Concept
	var refDocs []gyrus.Concept

	for _, res := range results {
		seenIDs[res.Concept.Name] = true
		if res.Concept.Scope == "project" {
			projectDocs = append(projectDocs, res.Concept)
		} else {
			refDocs = append(refDocs, res.Concept)
		}
	}

	// Dependency link expansion: For project concepts, expand linked reference dependencies
	var linkedRefDocs []gyrus.Concept
	for _, projectDoc := range projectDocs {
		edges := gyrus.ConceptEdges(projectDoc)
		if e.graph != nil {
			edges, err = e.Neighbors(ctx, projectDoc.Name, gyrus.EdgeFilter{Direction: "outgoing", RelationshipType: gyrus.RelDependsOn})
			if err != nil {
				return "", err
			}
		}
		for _, edge := range edges {
			if edge.RelationshipType != gyrus.RelDependsOn {
				continue
			}
			depID := edge.ToConceptID
			if !seenIDs[depID] {
				seenIDs[depID] = true
				if linkedDoc, err := e.Get(ctx, depID); err == nil {
					linkedRefDocs = append(linkedRefDocs, linkedDoc)
				}
			}
		}
	}

	// Assemble combined prioritized concepts up to maxConcepts
	// Priority: 1) Project concepts, 2) Linked reference concepts, 3) General reference concepts
	var finalDocs []gyrus.Concept
	for _, d := range projectDocs {
		if len(finalDocs) < maxConcepts {
			finalDocs = append(finalDocs, d)
		}
	}
	for _, d := range linkedRefDocs {
		if len(finalDocs) < maxConcepts {
			finalDocs = append(finalDocs, d)
		}
	}
	for _, d := range refDocs {
		if len(finalDocs) < maxConcepts {
			finalDocs = append(finalDocs, d)
		}
	}

	var sb strings.Builder
	sb.WriteString("=== SUGGESTED CONTEXT LAYER ===\n\n")

	// Group output by section: Project context first, then Shared Reference context
	var finalProject []gyrus.Concept
	var finalRef []gyrus.Concept
	for _, d := range finalDocs {
		if d.Scope == "project" {
			finalProject = append(finalProject, d)
		} else {
			finalRef = append(finalRef, d)
		}
	}

	if len(finalProject) > 0 {
		projectName := filter.Project
		if projectName == "" && len(finalProject) > 0 && finalProject[0].Project != "" {
			projectName = finalProject[0].Project
		}
		if projectName != "" {
			sb.WriteString(fmt.Sprintf("[PROJECT CONTEXT: %s]\n\n", projectName))
		} else {
			sb.WriteString("[PROJECT CONTEXT]\n\n")
		}
		for _, doc := range finalProject {
			sb.WriteString(fmt.Sprintf("--- [%s] %s (%s/%s) ---\n", doc.Name, doc.Title, doc.Category, doc.Type))
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	if len(finalRef) > 0 {
		tenant := "root"
		if len(finalRef) > 0 && finalRef[0].Tenant != "" {
			tenant = finalRef[0].Tenant
		}
		sb.WriteString(fmt.Sprintf("[SHARED REFERENCE LAYER: %s]\n\n", tenant))
		for _, doc := range finalRef {
			sb.WriteString(fmt.Sprintf("--- [%s] %s (%s/%s) ---\n", doc.Name, doc.Title, doc.Category, doc.Type))
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	return sb.String(), nil
}

// Link connects two concepts with a directed relationship edge.
func (e *Engine) Link(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if e.graph == nil {
		return fmt.Errorf("no graph store configured")
	}

	if relType == "" {
		relType = gyrus.RelDependsOn
	}
	if err := okf.ValidateName(fromID); err != nil {
		return err
	}
	if err := okf.ValidateName(toID); err != nil {
		return err
	}
	if !gyrus.ValidRelationship(relType) {
		return &okf.ValidationError{Field: "rel_type", Message: "invalid relationship type"}
	}

	edge := gyrus.ConceptEdge{
		FromConceptID:    fromID,
		ToConceptID:      toID,
		RelationshipType: relType,
	}

	if err := e.graph.UpsertEdges(ctx, []gyrus.ConceptEdge{edge}); err != nil {
		return err
	}
	return e.refreshLinkedConcept(ctx, fromID)
}

// Unlink removes a directed relationship edge between two concepts.
func (e *Engine) Unlink(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if e.graph == nil {
		return fmt.Errorf("no graph store configured")
	}

	if relType == "" {
		relType = gyrus.RelDependsOn
	}
	if err := okf.ValidateName(fromID); err != nil {
		return err
	}
	if err := okf.ValidateName(toID); err != nil {
		return err
	}
	if !gyrus.ValidRelationship(relType) {
		return &okf.ValidationError{Field: "rel_type", Message: "invalid relationship type"}
	}

	if err := e.graph.DeleteEdges(ctx, fromID, toID, relType); err != nil {
		return err
	}
	return e.refreshLinkedConcept(ctx, fromID)
}

// Neighbors retrieves neighboring relationship edges for a concept.
func (e *Engine) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
	if err := okf.ValidateName(id); err != nil {
		return nil, err
	}

	if e.graph == nil {
		return nil, fmt.Errorf("no graph store configured")
	}
	return e.graph.Neighbors(ctx, id, filter)
}

// Traverse executes a graph path search from a start concept.
func (e *Engine) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if err := okf.ValidateName(query.StartID); err != nil {
		return nil, err
	}

	if e.graph == nil {
		return nil, fmt.Errorf("no graph store configured")
	}
	return e.graph.Traverse(ctx, query)
}

// GetSchema retrieves the schema template for a concept type.
// Precedence: Active Persistence Layer -> Pre-compiled Binary Embedded Templates.
// Returns an error if the schema template is not found in either source.
func (e *Engine) GetSchema(ctx context.Context, conceptType string) (string, error) {
	if e.schemaStore != nil {
		if content, err := e.schemaStore.GetSchema(ctx, conceptType); err == nil && strings.TrimSpace(content) != "" {
			return content, nil
		}
	}
	return okf.GetTemplate(conceptType, "")
}

// SaveSchema persists an OKF schema template into the active persistence layer.
func (e *Engine) SaveSchema(ctx context.Context, conceptType string, content string) error {
	if e.schemaStore == nil {
		return fmt.Errorf("active storage provider does not support schema persistence")
	}
	return e.schemaStore.SaveSchema(ctx, conceptType, content)
}

// ListSchemas enumerates all available schemas across persistence layer and embedded templates.
func (e *Engine) ListSchemas(ctx context.Context) ([]gyrus.SchemaInfo, error) {
	seen := make(map[string]gyrus.SchemaInfo)

	// 1. Embedded binary templates
	for _, t := range okf.ListEmbeddedTemplates() {
		seen[t] = gyrus.SchemaInfo{
			Type:   t,
			Source: "embedded",
		}
	}

	// 2. Persisted schemas from active storage provider
	if e.schemaStore != nil {
		persisted, err := e.schemaStore.ListSchemas(ctx)
		if err == nil {
			for _, p := range persisted {
				seen[p] = gyrus.SchemaInfo{
					Type:     p,
					Source:   "persisted",
					Location: fmt.Sprintf(".gyrus/schemas/%s.md", p),
				}
			}
		}
	}

	var result []gyrus.SchemaInfo
	for _, info := range seen {
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})

	return result, nil
}

// DeleteSchema removes a custom schema template from the persistence layer.
func (e *Engine) DeleteSchema(ctx context.Context, conceptType string) error {
	if e.schemaStore == nil {
		return fmt.Errorf("active storage provider does not support schema persistence")
	}
	return e.schemaStore.DeleteSchema(ctx, conceptType)
}

func (e *Engine) refreshLinkedConcept(ctx context.Context, id string) error {
	owned, ok := e.graph.(interface{ StoresLinksInConcepts() bool })
	if !ok || !owned.StoresLinksInConcepts() || e.indexer == nil {
		return nil
	}
	doc, err := e.store.Get(ctx, id)
	if err == nil {
		err = e.indexer.Index(ctx, doc)
	}
	if err != nil {
		return fmt.Errorf("source concept %q was persisted but its secondary index needs repair: %w", id, err)
	}
	return nil
}
