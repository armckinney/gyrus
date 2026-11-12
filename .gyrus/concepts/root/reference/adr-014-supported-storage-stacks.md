---
title: Support three complete storage stacks
id: adr-014-supported-storage-stacks
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
---

# Architecture Decision Record: Support three complete storage stacks

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

Cloud providers and arbitrary storage/index combinations broaden configuration and can expose branch-specific knowledge through a shared database index.

### 2. Objectives & Constraints

Provide clear local, Git, and PostgreSQL implementations. Validate compatibility before opening resources; prefer explicit breaking changes over compatibility layers.

---

## Decision

We will support only localfs/OKF, Git/OKF, and PostgreSQL/PostgreSQL/FTS. Remove cloud/blob implementations and aliases. The [configuration reference](tech-ref-002-config-schema.md) owns complete examples.

### Key Rationale

* **Consistency:** Each stack has a defined authoritative store and query model.
* **Simplicity:** Users choose among three supported implementations.

---

## Alternatives Considered

### Alternative 1: Retain all storage providers

* **Pros:** Existing cloud installations need fewer changes.
* **Cons:** Retains operational and validation obligations outside the MVP.

### Alternative 2: Allow arbitrary hybrid stacks

* **Pros:** More provider flexibility.
* **Cons:** Requires additional synchronization and branch-isolation contracts.

---

## Consequences

* **Positive / Gains:** Smaller dependency surface and predictable configuration behavior.
* **Negative / Trade-offs:** Users must relocate unsupported data and replace their configurations themselves.
* **Risks:** A configuration switch does not migrate existing data; operators must verify exports before changing stores.

---

## Compliance & Verification

* **Validation:** Generated profiles resolve to exactly three stacks; removed providers, aliases, and other combinations fail before resource creation.
