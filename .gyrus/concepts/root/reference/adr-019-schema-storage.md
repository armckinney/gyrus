---
title: Keep schema overrides with their knowledge collection
id: adr-019-schema-storage
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
links:
    - concept: root/reference/adr-003-persistence-layer-schema-storage-and-remote-linkage
      rel: supersedes
---

# Architecture Decision Record: Keep schema overrides with their knowledge collection

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

Custom templates must follow the same storage and review lifecycle as the concepts that use them.

### 2. Objectives & Constraints

Keep overrides shared through the selected store and distinguish an absent override from a storage or session-selection failure.

---

## Decision

We will persist overrides in the selected store. Git schema changes use the same branch and context-aware commit path as concepts. Fall back to embedded templates only when an override is absent; propagate storage and selection failures.

### Key Rationale

* **Coherence:** Schemas and concepts move through the same review workflow.
* **Visible failures:** Embedded defaults cannot conceal inaccessible storage or an unselected initiative.

---

## Alternatives Considered

### Alternative 1: Keep schema overrides only in local files

* **Pros:** Local customization requires no remote write.
* **Cons:** Templates diverge across clients and bypass shared Git review.

### Alternative 2: Fall back on every storage failure

* **Pros:** Template retrieval succeeds during more outages.
* **Cons:** Authentication and selection mistakes are hidden by unrelated defaults.

---

## Consequences

* **Positive / Gains:** Shared, branch-specific templates and consistent mutation behavior.
* **Negative / Trade-offs:** Schema retrieval can fail when the store is unavailable or an initiative is unselected.
* **Risks:** Callers must propagate failures and only interpret missing overrides as permission to use embedded templates.

---

## Compliance & Verification

* **Validation:** Tests cover branch-specific templates, unchanged saves, delete/fallback behavior, cancellation, rejected pushes, and missing-selection errors.
