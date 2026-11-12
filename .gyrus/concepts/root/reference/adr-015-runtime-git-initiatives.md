---
title: Select Git initiative branches at runtime
id: adr-015-runtime-git-initiatives
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
---

# Architecture Decision Record: Select Git initiative branches at runtime

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

An initiative spans several documents and sessions. Saving its branch in shared configuration can redirect unrelated agents.

### 2. Objectives & Constraints

Keep initiative names out of configuration, support parallel independent MCP sessions, and preserve explicit human PR review.

---

## Decision

We will configure `trunk_branch` and `commit_strategy: direct|branching`, supply initiative names at runtime, and keep the active selection in the Git store. CLI/MCP handlers access that store through the application's generic `ConceptStore`; neither `App` nor the provider factory owns runtime branch state.

### Key Rationale

* **Isolation:** Every adapter uses the same session-owned store.
* **Separation:** Provider construction uses configuration; provider-specific commands change runtime selection.

---

## Alternatives Considered

### Alternative 1: Create a branch for every document

* **Pros:** Each document has a small review unit.
* **Cons:** Fragments initiatives involving multiple concepts and schemas.

### Alternative 2: Persist one shared review branch

* **Pros:** Restarting requires no runtime selection.
* **Cons:** Couples parallel initiatives and redirects other sessions.

---

## Consequences

* **Positive / Gains:** Initiatives collect related knowledge changes and can be switched live within an MCP process.
* **Negative / Trade-offs:** CLI invocations repeat `--git-branch`; restarting MCP requires selecting the initiative again.
* **Risks:** Switches must wait for in-flight requests. Humans must integrate trunk changes externally; no automatic merges or rebases are performed.

---

## Compliance & Verification

* **Validation:** CLI/subprocess and MCP tests cover selection, restart behavior, failed-switch rollback, unchanged configuration, and independent branches across concepts, schemas, links, and queries.
