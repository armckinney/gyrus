package gyrus

import "time"

// Category represents a top-level category classification.
type Category string

const (
	CategoryArchitecture Category = "architecture"
	CategoryBusinessLogic Category = "business-logic"
	CategoryProduct       Category = "product"
	CategoryOperations    Category = "operations"
	CategoryTechnical     Category = "technical"
)

// DocumentType represents an Open Knowledge Format (OKF) document type.
type DocumentType string

const (
	TypeADR                 DocumentType = "adr"
	TypePRD                 DocumentType = "prd"
	TypeGuide               DocumentType = "guide"
	TypeImprovementProposal DocumentType = "improvement-proposal"
	TypeReleaseNote         DocumentType = "release-note"
	TypeSpecification       DocumentType = "specification"
	TypeStandards           DocumentType = "standards"
	TypeTechnicalReference DocumentType = "technical-reference"
	TypeProduct             DocumentType = "product"
	TypeGlossary            DocumentType = "glossary"
	TypeFreeform            DocumentType = "freeform"
)

// Document represents an Open Knowledge Format (OKF) contract document.
type Document struct {
	ID             string       `json:"id" yaml:"id"`
	Title          string       `json:"title" yaml:"title"`
	Category       Category     `json:"category" yaml:"category"`
	Type           DocumentType `json:"type" yaml:"type"`
	Format         string       `json:"format" yaml:"format"`
	OwnerGroup     string       `json:"owner_group" yaml:"owner_group"`
	Version        int          `json:"version" yaml:"version"`
	Status         string       `json:"status" yaml:"status"`
	LastModifiedBy string       `json:"last_modified_by" yaml:"last_modified_by"`
	LastUpdated    time.Time    `json:"last_updated" yaml:"last_updated"`
	Tags           []string     `json:"tags,omitempty" yaml:"tags,omitempty"`
	Dependencies   []string     `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Content        string       `json:"content" yaml:"-"`
}

// DocumentRef holds lightweight reference metadata returned after a mutation.
type DocumentRef struct {
	ID          string    `json:"id"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

// DocumentPatch represents a mutation request to update document fields.
type DocumentPatch struct {
	Title        *string   `json:"title,omitempty"`
	Status       *string   `json:"status,omitempty"`
	Tags         *[]string `json:"tags,omitempty"`
	Dependencies *[]string `json:"dependencies,omitempty"`
	Content      *string   `json:"content,omitempty"`
	Reason       string    `json:"reason"`
}

// SearchFilter specifies filter constraints for lexical and metadata queries.
type SearchFilter struct {
	Category   Category     `json:"category,omitempty"`
	Type       DocumentType `json:"type,omitempty"`
	Status     string       `json:"status,omitempty"`
	Tag        string       `json:"tag,omitempty"`
	OwnerGroup string       `json:"owner_group,omitempty"`
}

// SearchQuery represents a search request.
type SearchQuery struct {
	Query      string       `json:"query"`
	Filter     SearchFilter `json:"filter"`
	MaxResults int          `json:"max_results"`
}

// SearchResult represents a matching document result from search execution.
type SearchResult struct {
	Document    Document `json:"document"`
	Score       float64  `json:"score"`
	MatchReason string   `json:"match_reason,omitempty"`
}

// RelationshipType specifies valid edge linkage types between documents.
type RelationshipType string

const (
	RelSupersedes RelationshipType = "supersedes"
	RelDependsOn  RelationshipType = "depends_on"
	RelImplements RelationshipType = "implements"
	RelMitigates  RelationshipType = "mitigates"
)

// DocumentEdge represents a directed relationship link between two documents.
type DocumentEdge struct {
	FromDocumentID   string           `json:"from_document_id"`
	ToDocumentID     string           `json:"to_document_id"`
	RelationshipType RelationshipType `json:"relationship_type"`
	CreatedBy        string           `json:"created_by"`
	CreatedAt        time.Time        `json:"created_at"`
}

// EdgeFilter represents parameters to query document relationship edges.
type EdgeFilter struct {
	RelationshipType RelationshipType `json:"relationship_type,omitempty"`
	Direction        string           `json:"direction,omitempty"` // "outgoing" | "incoming" | "both"
}

// GraphQuery represents a multi-hop traversal request over relationship edges.
type GraphQuery struct {
	StartID  string `json:"start_id"`
	MaxDepth int    `json:"max_depth"`
}

// GraphPath represents a path traversed across document edges.
type GraphPath struct {
	Nodes []string       `json:"nodes"`
	Edges []DocumentEdge `json:"edges"`
}

// SyncReport summarizes the output of an incremental file re-indexing run.
type SyncReport struct {
	ScannedFiles   int      `json:"scanned_files"`
	IndexedFiles   int      `json:"indexed_files"`
	UnchangedFiles int      `json:"unchanged_files"`
	RemovedFiles   int      `json:"removed_files"`
	Errors         []string `json:"errors,omitempty"`
}
