package gyrus

import "context"

// ConceptStore manages durable CRUD persistence for concept payloads.
type ConceptStore interface {
	Create(ctx context.Context, doc Concept) (ConceptRef, error)
	Get(ctx context.Context, id string) (Concept, error)
	Update(ctx context.Context, id string, patch ConceptPatch, expectedVersion int) (ConceptRef, error)
	Delete(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
}

// IndexStore manages metadata indexing and search retrieval.
type IndexStore interface {
	Index(ctx context.Context, doc Concept) error
	Remove(ctx context.Context, id string) error
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)
	Sync(ctx context.Context, storageRoot string) (SyncReport, error)
}

// GraphStore manages concept linkages and lineage traversals.
type GraphStore interface {
	UpsertEdges(ctx context.Context, edges []ConceptEdge) error
	DeleteEdges(ctx context.Context, fromID string, toID string, relType RelationshipType) error
	Neighbors(ctx context.Context, id string, filter EdgeFilter) ([]ConceptEdge, error)
	Traverse(ctx context.Context, query GraphQuery) ([]GraphPath, error)
}

// SearchProvider handles lexical or semantic text search queries.
type SearchProvider interface {
	Search(ctx context.Context, query string, filter SearchFilter) ([]SearchResult, error)
}

// StorageProvider is the main repository abstraction wrapper over storage and indexing providers.
type StorageProvider interface {
	GetConcept(ctx context.Context, id string, userGroups []string) (*Concept, error)
	SaveConcept(ctx context.Context, doc *Concept) error
	SearchConcepts(ctx context.Context, query string, userGroups []string) ([]SearchResult, error)
}

// SchemaStore manages durable CRUD persistence for OKF contract schema templates in the storage provider.
type SchemaStore interface {
	GetSchema(ctx context.Context, conceptType string) (string, error)
	SaveSchema(ctx context.Context, conceptType string, content string) error
	ListSchemas(ctx context.Context) ([]string, error)
	DeleteSchema(ctx context.Context, conceptType string) error
}
