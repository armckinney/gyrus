---
title: Reject invalid discovered configuration without fallback
id: adr-017-strict-configuration
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
links:
    - concept: root/reference/adr-005-global-configuration-and-precedence-hierarchy
      rel: supersedes
---

# Architecture Decision Record: Reject invalid discovered configuration without fallback

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

Ignoring unknown fields or falling back after malformed YAML can silently choose another store or target the wrong branch.

### 2. Objectives & Constraints

Preserve file-driven discovery and full project-over-global replacement while failing before resource creation for invalid settings.

---

## Decision

We will decode exactly one strict YAML document and reject unknown fields, invalid provider stacks, and invalid Git settings. Runtime branch selection never writes configuration. See the [configuration reference](tech-ref-002-config-schema.md) for discovery and command details.

### Key Rationale

* **Predictability:** An invalid selected file cannot redirect writes through fallback.
* **Clarity:** Removed fields produce explicit errors instead of being silently ignored.

---

## Alternatives Considered

### Alternative 1: Keep permissive YAML decoding

* **Pros:** Obsolete or misspelled fields do not prevent startup.
* **Cons:** Settings can be ignored without the user knowing.

### Alternative 2: Fall back after invalid configuration

* **Pros:** Startup may succeed using another file.
* **Cons:** Successful startup can use an unintended store or branch.

---

## Consequences

* **Positive / Gains:** Configuration errors are explicit and occur before provider resources open.
* **Negative / Trade-offs:** Existing users must replace obsolete fields and provider selections.
* **Risks:** Malformed configuration blocks operations until corrected; no automatic migration or legacy aliases are provided.

---

## Compliance & Verification

* **Validation:** Tests cover malformed project/global files, removed keys, multiple YAML documents, supported stacks, and failure before resource creation.
