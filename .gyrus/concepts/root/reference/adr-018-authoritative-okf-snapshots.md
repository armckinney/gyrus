---
title: Derive file-backed queries from authoritative OKF snapshots
id: adr-018-authoritative-okf-snapshots
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
links:
    - concept: root/reference/adr-010-okf-authoritative-providers
      rel: supersedes
---

# Architecture Decision Record: Derive file-backed queries from authoritative OKF snapshots

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

File-backed queries and graph reads must reflect stored knowledge, including the currently selected Git initiative.

### 2. Objectives & Constraints

Use the same authoritative store across adapters and avoid a second durable index or synchronization step.

---

## Decision

We will derive ephemeral query snapshots through direct-scan OKF index/graph/search adapters over the same local or Git store. Qualified typed-link identity remains governed by [ADR 011](adr-011-qualified-concept-links.md).

### Key Rationale

* **Freshness:** Queries reflect confirmed storage state.
* **Branch isolation:** Git selection applies naturally to every adapter using that store.

---

## Alternatives Considered

### Alternative 1: Mirror file-backed data into a database

* **Pros:** Supports database-native query optimization.
* **Cons:** Adds synchronization and branch-isolation requirements.

### Alternative 2: Require explicit index synchronization

* **Pros:** Separates indexing cost from queries.
* **Cons:** Queries can use stale state and require another maintenance workflow.

---

## Consequences

* **Positive / Gains:** Portable collections and immediate visibility of confirmed stored changes.
* **Negative / Trade-offs:** Query scans cost more as collections grow; PostgreSQL remains a separate complete stack.
* **Risks:** Invalid or duplicate concepts must fail queries rather than produce misleading partial results.

---

## Compliance & Verification

* **Validation:** Tests compare persisted concepts, source-owned links, search results, independent initiative snapshots, and invalid/duplicate collection failures.
