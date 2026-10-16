---
id: adr-009-okf-concepts-projects-tenant-nomenclature
title: OKF Concepts, Projects, and Tenant Nomenclature Refactoring
category: architecture
type: adr
format: markdown
tenant: root
version: 1
status: accepted
immutable: true
last_modified_by: antigravity
last_updated: 2026-10-16T03:30:00Z
tags: ["nomenclature", "okf", "concepts", "projects", "tenant", "cleanup"]
dependencies: ["adr-002-simplified-directory-topology", "adr-007-repository-context-and-reference-scoping"]
---

# ADR 009: OKF Concepts, Projects, and Tenant Nomenclature Refactoring

## Context

* **Author:** Antigravity Agent
* **Date Proposed:** 2026-10-16
* **Deciders:** Core platform team
* **Supersedes:** `adr-002-simplified-directory-topology`, `adr-007-repository-context-and-reference-scoping`

### 1. The Problem

Over earlier development phases, Gyrus terminology accumulated ambiguities and historical baggage:

1. **"Documents" vs. Open Knowledge Framework (OKF)**: In OKF, knowledge records are conceptual units of thought, architectural decisions, specifications, guidelines, and standards. Naming them "documents" or "docs" implies static office files rather than relational, graph-connected knowledge nodes ("concepts"). Furthermore, `.gyrus/docs/` collided semantically with traditional user documentation.
2. **"Workspaces" Ambiguity**: The term "workspace" conflicted directly with developer container workspaces (such as VS Code DevContainers and GitHub Codespaces mounts at `/workspaces/<repo>`), causing path parsing edge cases. Additionally, contemporary agentic frameworks standardize on "projects" as the unit of scoped application development.
3. **`owner_group` vs. `tenant`**: The YAML configuration field `default_owner_group` and frontmatter field `owner_group` were overly verbose and did not convey multi-tenant organization boundaries as clearly as `tenant`.
4. **Automation Hooks Context Contamination**: Shell execution hooks (`gyrus hook pre-invocation` / `stop`) injected command strings directly into developer terminals, causing context overload and unintended loops without aiding agent reasoning.
5. **CLI Shell Completion**: Cobra's default `gyrus completion` command was unnecessary for the scripted CLI and CI/CD agent invocation patterns of Gyrus.

### 2. Objectives & Constraints

**Goals:**
- Shift storage topology from `.gyrus/docs/<tenant>/workspaces/<name>/` to `.gyrus/concepts/<tenant>/projects/<name>/`.
- Rename core domain types, methods, tables, and columns from `Document` / `DocumentEdge` to `Concept` / `ConceptEdge`.
- Standardize on `tenant` in `.gyrus.yaml` and OKF frontmatter, replacing `default_owner_group` and `owner_group`.
- Standardize on `project` in `.gyrus.yaml` and CLI flags (`--project`), replacing `workspace`.
- Eliminate `gyrus hook` and disable default `gyrus completion`.
- Retain backwards-compatible URI fallbacks in MCP while promoting `memory://concept/{id}`.

---

## Decision

We execute a comprehensive, clean nomenclature refactoring across Gyrus:

### 1. Storage Topology Refactoring
The filesystem directory structure is updated:
```
.gyrus/
├── config.yaml / .gyrus.yaml
├── index.db
├── schemas/
└── concepts/
    └── <tenant>/
        ├── projects/
        │   └── <project-name>/
        │       └── <concept-id>.md
        └── reference/
            └── <concept-id>.md
```

### 2. Domain & Schema Harmonization
- **Core Structs**: `gyrus.Concept`, `gyrus.ConceptEdge`, `gyrus.ConceptStore`, `gyrus.ConceptPatch`.
- **OKF Frontmatter**: `tenant: <tenant-name>` is the mandatory field (validated by OKF validator).
- **Configuration**:
  ```yaml
  tenant: root
  project: gyrus
  storage:
    provider: localfs
    root: ./.gyrus
  ```
- **CLI Commands & Flags**:
  - `gyrus create --tenant <t> --project <p>`
  - `gyrus search --tenant <t> --project <p>`
  - `gyrus config init --tenant <t>`
- **Database Schema**:
  - Tables: `concepts`, `fts_concepts`, `concept_edges`.
  - Columns: `tenant`, `project`, `from_concept_id`, `to_concept_id`.

### 3. Removal of Hooks & Completion
- All `gyrus hook` subcommands and automation hooks are removed.
- Cobra's default completion command is disabled via `CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true}`.

---

## Consequences

### Positive
- **Architectural Clarity**: High-level knowledge records are consistently referred to as `concepts`, clearly distinguished from implementation code files.
- **Agent Alignment**: The term `project` aligns cleanly with developer workspaces and AI agent frameworks, eliminating confusion with container mount paths.
- **Clean Configuration**: `.gyrus.yaml` is concise (`tenant: root`, `project: gyrus`).
- **Reduced Context Overhead**: Eliminating shell hooks ensures agents interact with Gyrus through clean MCP tools and standard CLI queries.

### Superseded Records
- `adr-002-simplified-directory-topology`: Replaced directory layout `.gyrus/docs/` with `.gyrus/concepts/`.
- `adr-007-repository-context-and-reference-scoping`: Replaced scoping path `.gyrus/docs/<tenant>/workspaces/<name>/` with `.gyrus/concepts/<tenant>/projects/<name>/`.

