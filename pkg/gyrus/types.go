package gyrus

import "time"

// Category represents a top-level category classification.
type Category string

const (
	CategoryArchitecture  Category = "architecture"
	CategoryBusinessLogic Category = "business-logic"
	CategoryProduct       Category = "product"
	CategoryOperations    Category = "operations"
	CategoryTechnical     Category = "technical"
)

// ConceptType represents an Open Knowledge Format (OKF) concept type.
type ConceptType string

const (
	TypeADR                 ConceptType = "adr"
	TypePRD                 ConceptType = "prd"
	TypeGuide               ConceptType = "guide"
	TypeImprovementProposal ConceptType = "improvement-proposal"
	TypeReleaseNote         ConceptType = "release-note"
	TypeSpecification       ConceptType = "specification"
	TypeStandards           ConceptType = "standards"
	TypeTechnicalReference  ConceptType = "technical-reference"
	TypeProduct             ConceptType = "product"
	TypeGlossary            ConceptType = "glossary"
	TypeFreeform            ConceptType = "freeform"
)

// Concept represents an Open Knowledge Format (OKF) contract concept.
type Concept struct {
	Name           string        `json:"name,omitempty" yaml:"-"`
	ID             string        `json:"id" yaml:"id"`
	Title          string        `json:"title" yaml:"title"`
	Category       Category      `json:"category" yaml:"category"`
	Type           ConceptType   `json:"type" yaml:"type"`
	Format         string        `json:"format" yaml:"format"`
	Tenant         string        `json:"tenant" yaml:"tenant"`
	Version        int           `json:"version" yaml:"version"`
	Status         string        `json:"status" yaml:"status"`
	Immutable      bool          `json:"immutable,omitempty" yaml:"immutable,omitempty"`
	LastModifiedBy string        `json:"last_modified_by" yaml:"last_modified_by"`
	LastUpdated    time.Time     `json:"last_updated" yaml:"last_updated"`
	Tags           []string      `json:"tags,omitempty" yaml:"tags,omitempty"`
	Links          []ConceptLink `json:"links,omitempty" yaml:"links,omitempty"`
	Scope          string        `json:"scope,omitempty" yaml:"-"`
	Project        string        `json:"project,omitempty" yaml:"-"`
	FilePath       string        `json:"filepath,omitempty" yaml:"-"`
	Content        string        `json:"content" yaml:"-"`
}

// ConceptRef holds lightweight reference metadata returned after a mutation.
type ConceptRef struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

// ConceptPatch represents a mutation request to update concept fields.
type ConceptPatch struct {
	Title   *string        `json:"title,omitempty"`
	Status  *string        `json:"status,omitempty"`
	Tags    *[]string      `json:"tags,omitempty"`
	Links   *[]ConceptLink `json:"links,omitempty"`
	Content *string        `json:"content,omitempty"`
	Reason  string         `json:"reason"`
}

// SearchFilter specifies filter constraints for lexical and metadata queries.
type SearchFilter struct {
	Category Category    `json:"category,omitempty"`
	Type     ConceptType `json:"type,omitempty"`
	Status   string      `json:"status,omitempty"`
	Tag      string      `json:"tag,omitempty"`
	Tenant   string      `json:"tenant,omitempty"`
	Scope    string      `json:"scope,omitempty"`   // "project", "reference", "all"
	Project  string      `json:"project,omitempty"` // target project name
}

// SearchQuery represents a search request.
type SearchQuery struct {
	Query      string       `json:"query"`
	Filter     SearchFilter `json:"filter"`
	MaxResults int          `json:"max_results"`
}

// SearchResult represents a matching concept result from search execution.
type SearchResult struct {
	Concept     Concept `json:"concept"`
	Score       float64 `json:"score"`
	MatchReason string  `json:"match_reason,omitempty"`
}

// RelationshipType specifies valid edge linkage types between concepts.
type RelationshipType string

// ConceptLink is a source-owned directed relationship in OKF frontmatter.
type ConceptLink struct {
	Concept string           `json:"concept" yaml:"concept"`
	Rel     RelationshipType `json:"rel" yaml:"rel"`
}

const (
	RelSupersedes RelationshipType = "supersedes"
	RelDependsOn  RelationshipType = "depends_on"
	RelImplements RelationshipType = "implements"
	RelMitigates  RelationshipType = "mitigates"
)

// ConceptEdge represents a directed relationship link between two concepts.
type ConceptEdge struct {
	FromConceptID    string           `json:"from_concept_id"`
	ToConceptID      string           `json:"to_concept_id"`
	RelationshipType RelationshipType `json:"relationship_type"`
}

// EdgeFilter represents parameters to query concept relationship edges.
type EdgeFilter struct {
	RelationshipType RelationshipType `json:"relationship_type,omitempty"`
	Direction        string           `json:"direction,omitempty"` // "outgoing" | "incoming" | "both"
}

// GraphQuery represents a multi-hop traversal request over relationship edges.
type GraphQuery struct {
	StartID  string `json:"start_id"`
	MaxDepth int    `json:"max_depth"`
}

// GraphPath represents a path traversed across concept edges.
type GraphPath struct {
	Nodes []string      `json:"nodes"`
	Edges []ConceptEdge `json:"edges"`
}

// SyncReport summarizes the output of an incremental file re-indexing run.
type SyncReport struct {
	ScannedFiles   int      `json:"scanned_files"`
	IndexedFiles   int      `json:"indexed_files"`
	UnchangedFiles int      `json:"unchanged_files"`
	RemovedFiles   int      `json:"removed_files"`
	Errors         []string `json:"errors,omitempty"`
}

// SchemaInfo represents metadata about an OKF schema template.
type SchemaInfo struct {
	Type     string `json:"type"`
	Source   string `json:"source"` // "persisted" or "embedded"
	Location string `json:"location,omitempty"`
}
