# [GYRUS-203] SQLite Edge Graph & Incremental File Sync

> **Status:** `NOT STARTED`
> **Phase:** Phase 2 - Local Storage & SQLite Provider Engines
> **Owner:** Unassigned
> **Dependencies:** [GYRUS-201], [GYRUS-202]

---

## 1. Description
Implement the `document_edges` graph store and the explicit incremental file synchronization engine (`Sync()`) in `internal/provider/sqlite`. Automatically extract relationships from OKF `dependencies` frontmatter fields and sync SQLite indexes with files modified on disk.

---

## 2. Acceptance Criteria
- [ ] Implement `gyrus.GraphStore` over SQLite `document_edges` table (`UpsertEdges`, `DeleteEdges`, `Neighbors`, `Traverse`).
- [ ] Automatically extract edges from OKF `dependencies: [...]` frontmatter lists on indexing.
- [ ] Implement `Sync(ctx, storageRoot)` calculating mtime/hash checksums to incrementally re-index updated files on disk.

---

## 3. Implementation Tasks
1. Create `internal/provider/sqlite/graph.go` with SQL queries for `document_edges`.
2. Create `internal/provider/sqlite/sync.go` with directory walk and hash comparison.
3. Write test cases for edge insertion, neighbor traversals, and file sync detection.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/provider/sqlite/... -v
```
