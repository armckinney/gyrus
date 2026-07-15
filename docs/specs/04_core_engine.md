# Specification 04: Core Engine & Validation

The Gyrus Core Engine is independent of input adapters (CLI/MCP) and handles contract schema enforcement, lifecycle transition checks, searchindexing, and edge-table relationship traversals.

---

## 1. Core Go Abstraction Interfaces

The engine decouples storage, search, and relationship mapping using standard interfaces:

```go
type DocumentStore interface {
    Create(ctx context.Context, doc Document) (DocumentRef, error)
    Get(ctx context.Context, id string) (Document, error)
    Update(ctx context.Context, id string, patch DocumentPatch) (DocumentRef, error)
    Delete(ctx context.Context, id string) error // Soft-delete or archive only
}

type IndexStore interface {
    Index(ctx context.Context, doc Document) error
    Remove(ctx context.Context, id string) error
    Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)
}

type GraphStore interface {
    UpsertEdges(ctx context.Context, edges []DocumentEdge) error
    Neighbors(ctx context.Context, id string, filter EdgeFilter) ([]DocumentEdge, error)
    Traverse(ctx context.Context, query GraphQuery) ([]GraphPath, error)
}
```

---

## 2. SQL Edge-Table Relationship Graph

To avoid the operational overhead of a dedicated Graph Database (like Neo4j) during the MVP, Gyrus models relationships using a simple `document_edges` index table stored in SQLite (Local Profile) or Postgres (Team/Enterprise Profiles):

```sql
CREATE TABLE document_edges (
    from_document_id   VARCHAR(255) NOT NULL,
    to_document_id     VARCHAR(255) NOT NULL,
    relationship_type  VARCHAR(100) NOT NULL, -- e.g., 'supersedes', 'depends_on', 'implements'
    created_by         VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (from_document_id, to_document_id, relationship_type)
);
```

### Standard Relationship Types:
* `supersedes` / `superseded_by` (e.g. `adr` supersedes `adr`)
* `depends_on` (e.g. `improvement-proposal` depends on `specification`)
* `implements` (e.g. `improvement-proposal` implements `prd`)
* `mitigates` (e.g. `guide` mitigates `known-issue`)

---

## 3. Lifecycle State-Transition Machine

Gyrus enforces strict lifecycle validations. An agent or developer cannot transition a document to an illegal state.

```
ADR Lifecycle:
  [proposed] ──► [accepted] ──► [superseded]
      │              │
      ▼              ▼
  [rejected]     [deprecated]

Proposal Lifecycle:
  [draft] ──► [reviewing] ──► [approved] ──► [implemented]
                  │               │
                  ▼               ▼
              [rejected]      [abandoned]
```

### Transition Validation Matrix

| Current State | Allowed Next States |
| :--- | :--- |
| **ADR - `proposed`** | `accepted`, `rejected` |
| **ADR - `accepted`** | `superseded`, `deprecated` |
| **ADR - `superseded`** | None (Immutable) |
| **ADR - `deprecated`** | None (Immutable) |
| **ADR - `rejected`** | None (Immutable) |
| **Proposal - `draft`** | `reviewing`, `abandoned` |
| **Proposal - `reviewing`** | `approved`, `rejected` |
| **Proposal - `approved`** | `implemented`, `abandoned` |
| **Proposal - `implemented`**| None (Immutable) |
| **Proposal - `rejected`** | `draft` (If re-opened) |

### Validation Rule Enforcement:
Any write/patch operation that violates these transition rules is **rejected at the Core Engine boundary** with a `VALIDATION_FAILED` error status, preventing agents from corrupting the index.
