package gyrus

import "context"

// DocumentStore manages durable CRUD persistence for document payloads.
type DocumentStore interface {
	Create(ctx context.Context, doc Document) (DocumentRef, error)
	Get(ctx context.Context, id string) (Document, error)
	Update(ctx context.Context, id string, patch DocumentPatch, expectedVersion int) (DocumentRef, error)
	Delete(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
}

// IndexStore manages metadata indexing and search retrieval.
type IndexStore interface {
	Index(ctx context.Context, doc Document) error
	Remove(ctx context.Context, id string) error
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)
	Sync(ctx context.Context, storageRoot string) (SyncReport, error)
}

// GraphStore manages document linkages and lineage traversals.
type GraphStore interface {
	UpsertEdges(ctx context.Context, edges []DocumentEdge) error
	DeleteEdges(ctx context.Context, fromID string, toID string, relType RelationshipType) error
	Neighbors(ctx context.Context, id string, filter EdgeFilter) ([]DocumentEdge, error)
	Traverse(ctx context.Context, query GraphQuery) ([]GraphPath, error)
}

// SearchProvider handles lexical or semantic text search queries.
type SearchProvider interface {
	Search(ctx context.Context, query string, filter SearchFilter) ([]SearchResult, error)
}

// StorageProvider is the main repository abstraction wrapper over storage and indexing providers.
type StorageProvider interface {
	GetDocument(ctx context.Context, id string, userGroups []string) (*Document, error)
	SaveDocument(ctx context.Context, doc *Document) error
	SearchDocuments(ctx context.Context, query string, userGroups []string) ([]SearchResult, error)
}
