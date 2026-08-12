---
id: spec-004-core-engine
title: Core SDK Provider Architecture & SQLite Engine Specs
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Specification 04: Core Engine SDK & Validation

The core logic of Gyrus is engineered as the **Gyrus Core SDK**—a standalone Go library (`pkg/gyrus`) independent of CLI or MCP presentation layers. The SDK handles contract schema enforcement, lifecycle transition checks, full-text search indexing, and relationship edge traversals by composing pluggable provider implementations.

---

## 1. Gyrus Core SDK Abstraction Interfaces

The Gyrus Core SDK decouples storage, indexing, knowledge graph, and search operations using clean Go interfaces:

```go
// DocumentStore manages durable CRUD persistence for document payloads.
type DocumentStore interface {
    Create(ctx context.Context, doc Document) (DocumentRef, error)
    Get(ctx context.Context, id string) (Document, error)
    Update(ctx context.Context, id string, patch DocumentPatch) (DocumentRef, error)
    Delete(ctx context.Context, id string) error // Soft-delete or archive
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
    DeleteEdges(ctx context.Context, fromID string, toID string, relType string) error
    Neighbors(ctx context.Context, id string, filter EdgeFilter) ([]DocumentEdge, error)
    Traverse(ctx context.Context, query GraphQuery) ([]GraphPath, error)
}

// SearchProvider handles lexical or semantic text search queries.
type SearchProvider interface {
    Search(ctx context.Context, query string, filter SearchFilter) ([]SearchResult, error)
}
```


---

## 2. Complete SQLite Index & Edge DDL Schemas

In the Local Profile (OKF storage mode), Gyrus uses a local SQLite database (`index.db` inside the Gyrus storage directory) for fast indexing, search, and relationship mapping:

```sql
-- 1. Metadata Index Table
CREATE TABLE IF NOT EXISTS documents_index (
    id                VARCHAR(255) PRIMARY KEY,
    title             VARCHAR(500) NOT NULL,
    category          VARCHAR(100) NOT NULL,
    type              VARCHAR(100) NOT NULL,
    format            VARCHAR(50)  NOT NULL DEFAULT 'markdown',
    owner_group       VARCHAR(255) NOT NULL,
    version           INTEGER      NOT NULL DEFAULT 1,
    status            VARCHAR(100) NOT NULL,
    last_modified_by  VARCHAR(255) NOT NULL,
    last_updated      TIMESTAMP    NOT NULL,
    tags_json         TEXT,        -- JSON array of tags e.g. ["sqlite","storage"]
    dependencies_json TEXT,        -- JSON array of document IDs e.g. ["prd-001"]
    file_path         TEXT         NOT NULL,
    content_hash      VARCHAR(64)  NOT NULL
);

-- 2. Full-Text Search (FTS5) Virtual Table
CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    id UNINDEXED,
    title,
    tags,
    content,
    tokenize = 'porter ascii'
);

-- 3. SQL Edge-Table Relationship Graph
CREATE TABLE IF NOT EXISTS document_edges (
    from_document_id   VARCHAR(255) NOT NULL,
    to_document_id     VARCHAR(255) NOT NULL,
    relationship_type  VARCHAR(100) NOT NULL, -- 'supersedes', 'depends_on', 'implements', 'mitigates'
    created_by         VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (from_document_id, to_document_id, relationship_type)
);

-- Secondary Indexes for Fast Traversal & Filtering
CREATE INDEX IF NOT EXISTS idx_docs_category ON documents_index(category);
CREATE INDEX IF NOT EXISTS idx_docs_type ON documents_index(type);
CREATE INDEX IF NOT EXISTS idx_docs_status ON documents_index(status);
CREATE INDEX IF NOT EXISTS idx_edges_to ON document_edges(to_document_id);
```

### Relationship Extraction (OKF vs Enterprise DB):
* **Open Knowledge Format (Local Profile):** Gyrus automatically extracts edges from document frontmatter attributes (`dependencies: [...]`, `supersedes: [...]`) during indexing and populates `document_edges`. Explicit CLI/MCP `link` commands update the frontmatter and the `document_edges` table.
* **Enterprise DB (Team/Enterprise Profile):** Relationships are stored directly in PostgreSQL/Cosmos DB edge tables, independent of whether the document body contains frontmatter array fields.

---

## 3. Comprehensive Lifecycle State-Transition Machine

Gyrus enforces strict lifecycle validations for ALL supported document types. Any state update violating these transition matrices is **rejected with exit code 2 (`TRANSITION_ERROR`)**.

```
ADR Lifecycle:
  [proposed] ──► [accepted] ──► [superseded]
      │              │
      ▼              ▼
  [rejected]     [deprecated]

Improvement Proposal Lifecycle:
  [draft] ──► [reviewing] ──► [approved] ──► [implemented]
                  │               │
                  ▼               ▼
              [rejected]      [abandoned]

General Document Lifecycle (PRD, Spec, Guide, Standards, Technical-Ref, Product, Glossary, Release-Note, Freeform):
  [draft] ──► [active] ──► [deprecated] ──► [archived]
```

### Transition Validation Matrix

| Document Type | Current Status | Legal Next Statuses |
| :--- | :--- | :--- |
| **`adr`** | `proposed` | `accepted`, `rejected` |
| | `accepted` | `superseded`, `deprecated` |
| | `superseded` | None (Immutable) |
| | `deprecated` | None (Immutable) |
| | `rejected` | None (Immutable) |
| **`improvement-proposal`** | `draft` | `reviewing`, `abandoned` |
| | `reviewing` | `approved`, `rejected` |
| | `approved` | `implemented`, `abandoned` |
| | `implemented` | None (Immutable) |
| | `rejected` | `draft` (If re-opened) |
| | `abandoned` | `draft` (If re-opened) |
| **General Types**<br>(`prd`, `specification`, `guide`, `standards`, `technical-reference`, `product`, `glossary`, `release-note`, `freeform`) | `draft` | `active`, `archived` |
| | `active` | `deprecated`, `archived` |
| | `deprecated` | `archived` |
| | `archived` | `draft` (If unarchived) |

---

## 4. Optimistic Concurrency Control

When an update is submitted:
1. Gyrus fetches the existing record and compares its stored `version` against the client's `expected_version`.
2. If `expected_version != version`, Gyrus aborts the update and returns exit code 4 (`CONCURRENCY_CONFLICT`).
3. If versions match, Gyrus increments `version = version + 1`, updates `last_updated` and `last_modified_by`, persists changes, and updates the search index.
