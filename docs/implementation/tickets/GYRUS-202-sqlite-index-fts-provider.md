# [GYRUS-202] SQLite DDL, Indexing & FTS5 Search Engine

> **Status:** `COMPLETED`
> **Phase:** Phase 2 - Local Storage & SQLite Provider Engines
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-101], [GYRUS-201]



---

## 1. Description
Implement the SQLite indexer and full-text search provider in `internal/provider/sqlite`. Execute DDL migrations for `documents_index`, `documents_fts`, and secondary indexes, supporting lexical keyword search matching Specification 04.

---

## 2. Acceptance Criteria
- [ ] SQLite DDL execution for `documents_index` and `documents_fts` (FTS5).
- [ ] Index document metadata and content text on document creation/update.
- [ ] Implement `gyrus.SearchProvider` over FTS5 with query filtering by `category`, `type`, `status`, `tag`, and `owner_group`.

---

## 3. Implementation Tasks
1. Create `internal/provider/sqlite/migrations.go` with schema DDLs.
2. Create `internal/provider/sqlite/indexer.go` and `internal/provider/sqlite/search.go`.
3. Write test cases for metadata indexing, FTS token queries, and category filtering.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/provider/sqlite/... -v
```
