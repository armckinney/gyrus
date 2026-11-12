---
title: Authoritative OKF scans and durable typed links
id: adr-010-okf-authoritative-providers
type: adr
format: markdown
category: architecture
version: 6
status: superseded
last_updated: 2026-11-12T05:07:17Z
last_modified_by: ""
tenant: root
links:
    - concept: root/reference/ip-006-okf-providers
      rel: implements
---

> Historical decision. Superseded by [Derive file-backed queries from authoritative OKF snapshots](adr-018-authoritative-okf-snapshots.md). The original decision content is retained within the restored template below.

# Authoritative OKF scans and durable typed links

## Context
Database-only graph mutations are not portable with stored concepts. Existing factories and profiles require databases even for file-backed collections.

### 2. Objectives & Constraints

Make stored file-backed concepts authoritative for metadata, links, and queries without requiring a secondary database.

---

## Decision
Provide independent OKF index, graph, and search adapters over a shared scanner-capable store. Persist source-owned typed links, retaining dependencies as depends_on edges. Queries rebuild ephemeral snapshots; no durable index or cache is introduced. Empty queries enumerate metadata, keyword queries default to ten results. Use explicit provider resolution and compatibility checks before I/O.

### Consistency
Lock local read-modify-write operations across processes, reject conflicting Git pushes, and use cloud generation/ETag conditions. Atomicity is per concept; cross-source batches report partial application. Missing targets remain valid, duplicate IDs and invalid concepts fail queries.

### Key Rationale

Direct scans reflect the stored collection and keep source-owned relationships portable.

---

## Alternatives Considered

The original record did not separately record alternatives deliberations. These comparisons summarize the baseline approaches described in its context.

### Alternative 1: Store graph mutations only in a database

* **Pros:** Reuses the existing database graph representation.
* **Cons:** Relationships are not portable with stored concepts.

### Alternative 2: Require a secondary database index for file-backed collections

* **Pros:** Reuses existing database index providers.
* **Cons:** Requires another store and synchronization before queries reflect file changes.

---

## Consequences
Collections remain portable and immediately reflect stored changes. Scans trade scalability for operational simplicity. Database providers remain explicit alternatives. SQLite edges must be exported before cutover; PostgreSQL migration is outside v1.

* **Positive / Gains:** Stored concepts and relationships remain portable and immediately visible to queries.
* **Negative / Trade-offs:** Direct scans trade scalability for simpler operation.
* **Risks:** Large collections incur scan costs. Historical cloud preconditions and migration notes in the preserved decision belong to the superseded design; current stack boundaries are defined by ADRs 014 and 018.

---

## Compliance & Verification

* **Validation:** The original record stated consistency requirements without a separate executed-verification report. Current snapshot, branch-isolation, and invalid-collection tests are documented in ADR 018.
