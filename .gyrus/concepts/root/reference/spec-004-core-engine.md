---
title: Core SDK Provider Architecture
id: spec-004-core-engine
type: specification
format: markdown
category: technical
version: 8
status: active
last_updated: 2026-10-17T23:57:58Z
last_modified_by: ""
tenant: root
---

Gyrus composes ConceptStore, IndexStore, GraphStore, and SearchProvider through the application container. Storage is constructed once, then injected into the selected providers. Effective provider validation runs before opening storage/database/network resources.

## Provider contracts

ConceptStore provides Create, Get, version-checked Update, Delete, and Archive. Optional ConceptScanner.Scan(context.Context, func(Concept) error) (SyncReport, error) enumerates authoritative validated concepts. Localfs, Git, and blob implement this capability; PostgreSQL storage does not in this release.

IndexStore exposes Index, Remove, and Search. GraphStore exposes UpsertEdges, DeleteEdges, Neighbors, and Traverse. SearchProvider accepts SearchQuery including Query, Filter, and MaxResults. Engine.Search retains the default-limit convenience method; Engine.SearchQuery accepts explicit limits. PostgreSQL storage/index/search delegate to one shared full-text query implementation. SQLite index/graph/search adapters and CLI/MCP sync operations have been removed; internal/persistence/db retains SQLite connection support for future cache work.

## Direct-scan OKF providers

OKF index/search share filtering and ranking over ephemeral snapshots. Index and Remove validate requests without mutating storage. Scanner errors include record locations; queries fail explicitly rather than returning partial results. Keyword requests default to ten results; positive MaxResults values override the default. Empty-query browsing is unlimited unless capped explicitly. CLI and MCP pass limits to the provider before retrieval; PostgreSQL applies the same scope/project filters. Vector search, embedding clients/configuration, and its profile have been removed; shared SQLite persistence remains for future cache work.

Graph reads derive source-owned {concept, rel} links keyed by qualified concept names. Neighbors supports incoming/outgoing/both, defaulting to both. Traverse follows outgoing edges breadth-first and returns one GraphPath of reached nodes/edges, with deterministic ordering and cycle detection. Nonpositive depth defaults to one. Missing target nodes remain visible and are not expanded.

## Mutations and consistency

Localfs/fileblob serialize cooperating processes with a stable root lock and replace complete payloads atomically. Git refreshes before scans/mutations and pushes without force; rejected pushes are conflicts and cannot leak unconfirmed data. S3/Azure/GCS updates and deletes require the observed ETag/generation; creates require nonexistence. A token comes from the same object read as the content. Unsupported conditional mutation capabilities fail explicitly.

Positive expected versions are checked before mutation. Conflicts map to CLI code 4. Multi-source graph batches preflight sources, then apply one patch per source in qualified-name order; later failure reports already-applied sources. There is no cross-concept transaction, automatic retry, cloud-wide snapshot, or automatic cross-concept transaction. Local IDs may repeat across tenants/projects; duplicate qualified names and inconsistent locations fail queries.

PostgreSQL checks expected versions under its row lock and returns gyrus.ConflictError with the qualified name, update operation, and expected/actual versions. Rejection rolls back without concept/history writes; MCP returns a tool error and CLI returns code 4.

The UI graph reads outgoing edges only for a fixed list of originally visible concepts. Immediate filtered/missing targets are display-only ghosts, never traversal sources. Nodes are unique by qualified name; edges are unique by source/target/relation. Both collections are sorted deterministically. Edge-query errors return HTTP 500 before any partial graph is encoded.

Context suggestions expand depends_on edges from the selected graph provider for project concepts, including PostgreSQL database-only edges. Other relation types do not automatically expand context. OKF graph mutations refresh an explicitly selected PostgreSQL metadata index; a post-persistence indexing failure is reported as requiring repair without rolling back source data.

## Lifecycle

ADR transitions: proposed to accepted/rejected; accepted to superseded/deprecated. Improvement proposals: draft to reviewing/abandoned; reviewing to approved/rejected/abandoned; approved to implemented/abandoned. Other types: draft to active/archived; active to deprecated/archived; deprecated to archived. Same-state updates are accepted. Finalized immutable bodies are protected; metadata updates remain permitted.

## Qualified concept identity

Creation accepts a local ID and tenant/scope/project. All existing-concept operations require tenant/reference/id or tenant/projects/project/id. Concept envelopes and mutation references expose the derived name alongside the local id. Storage paths, database keys/history, graph endpoints, UI routes, search identity ranking, and context deduplication use the qualified name. Short IDs are rejected. Persist only links with concept and rel; relationship creation metadata and the dependencies field are removed.

## Shared package boundaries

Provider contracts remain in pkg/gyrus. internal/persistence/fileio supplies local locking, safe paths, and atomic file replacement; internal/persistence/db manages shared SQLite/PostgreSQL connections. These are concrete infrastructure packages, not providers or a second persistence interface. internal/provider/shared/okf supplies scanner validation, snapshots, filtering, and ranking used by the independent OKF adapters. Legacy conversion is an agent-led procedure in the packaged gyrus-migrate-sqlite-to-okf skill, using installed Gyrus and available file/database tools. It has no Go helper package or runtime imports and does not require a Go installation or source checkout. The packaged gyrus-test-lifecycle skill likewise verifies public CLI, MCP, and HTTP behavior through agent instructions without a bundled language runtime.
