# OKF Document Schemas Reference

This reference documents all 11 Open Knowledge Format (OKF) document types, their YAML frontmatter specifications, allowed lifecycle states, and mutability rules.

## 📋 Allowed Document Types & Categories

### Categories (`--category`)
- `architecture`: Software architecture, design patterns, decisions, infrastructure topology.
- `business-logic`: Requirements, specifications, domain logic, product features.
- `product`: Value proposition, positioning, roadmaps, release schedules.
- `operations`: Runbooks, deployment guides, monitoring, incident response.
- `technical`: Developer guides, codebase standards, API documentation, technical reference.

### Document Types (`--type`)
1. `adr`: Architecture Design Record (Immutable once accepted).
2. `prd`: Product Requirement Document (Living context).
3. `specification`: Technical Specification (Living context).
4. `guide`: Developer or Operational Guide (Living context).
5. `standards`: Engineering Standards & Coding Guidelines (Living context).
6. `technical-reference`: System Architecture & Component Reference (Living context).
7. `product`: Product Vision & Feature Spec (Living context).
8. `release-note`: Release Notes (Immutable log).
9. `improvement-proposal`: Improvement Proposal (Immutable once accepted).
10. `glossary`: Domain Glossary & Terminology (Living context).
11. `freeform`: General Unstructured Markdown (Living context).

---

## 📜 YAML Frontmatter Schema

Every OKF document starts with YAML frontmatter:

```yaml
---
id: adr-001-storage-engine
title: SQLite FTS5 for Local Search
category: architecture
type: adr
format: markdown
owner_group: platform
version: 1
status: accepted
last_modified_by: armckinney
last_updated: 2026-07-23T20:00:00Z
immutable: true
tags:
  - storage
  - sqlite
dependencies:
  - prd-001-master-spec
---

# Document Title
Markdown content goes here...
```

---

## ⚙️ Lifecycle State Transitions

| Document Type | Allowed Statuses | Valid Transitions |
| :--- | :--- | :--- |
| `adr` | `draft`, `proposed`, `accepted`, `rejected`, `deprecated`, `superseded` | `draft` ➔ `proposed` ➔ `accepted`/`rejected` ➔ `deprecated`/`superseded` |
| `improvement-proposal` | `draft`, `proposed`, `accepted`, `rejected`, `withdrawn` | `draft` ➔ `proposed` ➔ `accepted`/`rejected`/`withdrawn` |
| `release-note` | `draft`, `published` | `draft` ➔ `published` |
| Living Docs (`prd`, `spec`, etc.) | `draft`, `proposed`, `active`, `deprecated`, `archived` | `draft` ➔ `proposed` ➔ `active` ➔ `deprecated`/`archived` |
