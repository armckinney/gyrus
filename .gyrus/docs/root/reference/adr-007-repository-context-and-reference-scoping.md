---
id: adr-007-repository-context-and-reference-scoping
title: Repository-Focused Context and Reference Scoping (Phase 2.8)
category: architecture
type: adr
format: markdown
owner_group: root
version: 1
status: accepted
immutable: true
last_modified_by: antigravity
last_updated: 2026-10-02T00:00:00Z
tags: ["scoping", "workspace", "reference", "context", "phase-2.8"]
dependencies: ["adr-002-simplified-directory-topology", "adr-005-global-configuration-and-precedence-hierarchy", "adr-006-native-mcp-workspace-discovery-and-lifecycle-cli-refactor"]
---

# ADR 007: Repository-Focused Context and Reference Scoping (Phase 2.8)

## Context

* **Author:** Antigravity Agent
* **Date Proposed:** 2026-10-02
* **Deciders:** Core platform team

### 1. The Problem

Prior to Phase 2.8, Gyrus operated with a single flat namespace across all documents in a storage root. All documents in `.gyrus/docs/<owner_group>/workspaces/main/` and `.gyrus/docs/<owner_group>/reference/` were treated equivalently during search and context suggestion.

This created several problems:

1. **No Workspace Isolation**: When multiple teams or repositories shared the same Gyrus storage root, search results and suggested context included documents from all workspaces indiscriminately, creating noise for agents working in a specific project.

2. **No Reference Promotion**: Shared architectural decisions, standards, and cross-cutting specifications had no mechanism to be treated as a higher-fidelity "reference" layer that any workspace could link to.

3. **Hardcoded `main` Workspace**: The storage topology used a hardcoded `workspaces/main` subdirectory, preventing teams from operating distinct named workspaces within the same tenant.

4. **No Explicit Workspace Configuration**: There was no user-facing mechanism to associate a repository's `.gyrus.yaml` with a named workspace, forcing users to rely on filesystem conventions.

5. **DevContainer Path Ambiguity**: The `/workspaces/` path prefix is shared with DevContainer and GitHub Codespace environment mount paths (e.g., `/workspaces/gyrus`). Naive substring matching on `/workspaces/` would misidentify the host project root as a Gyrus workspace name.

### 2. Objectives & Constraints

**Goals:**
- Provide explicit `workspace: <name>` config in `.gyrus.yaml` to associate a repo with a named workspace.
- Prioritize workspace-scoped documents in search and `suggest-context` output, while seamlessly expanding to reference layer documents via dependency traversal.
- Support a global `~/.gyrus.yaml` with no `workspace:` set for users who want tenant-wide reference search.
- Preserve OKF frontmatter purity — `scope` and `workspace` must be derived from storage path topology, not required in YAML frontmatter.

**Constraints:**
- **No directory auto-detection**: Workspace name must be explicitly configured in `.gyrus.yaml`. No heuristics based on directory names or repo names.
- **Path-based topology only**: Scope is derived purely from the OKF storage path (`docs/<tenant>/workspaces/<name>/` vs `docs/<tenant>/reference/`), not from document metadata.
- **DevContainer safety**: Scope parsing must anchor on `/docs/` to avoid false-positive `/workspaces/` matches from container mount paths.

---

## Decision

We adopt a **two-tier document scoping model** with **explicit workspace configuration** and **path-anchored scope derivation**.

### 1. Storage Topology (unchanged from ADR-002, formalized here)

```
.gyrus/
└── docs/
    └── <tenant>/               # owner_group (e.g. "root", "platform")
        ├── reference/          # Shared org-wide documents (scope=reference)
        │   ├── adr-001.md
        │   └── spec-001.md
        └── workspaces/
            ├── <workspace-a>/  # Project-scoped documents (scope=workspace)
            │   ├── ticket-001.md
            │   └── prd-local.md
            └── <workspace-b>/  # Another project's documents
                └── ticket-002.md
```

### 2. Explicit Workspace Configuration

Users add `workspace: <name>` to their repository `.gyrus.yaml`:

```yaml
# .gyrus.yaml in a project repository
storage:
  provider: localfs
  root: .gyrus
workspace: my-project          # <-- explicit workspace scoping
default_owner_group: platform
```

A global `~/.gyrus.yaml` without `workspace:` operates in **reference/all mode**, returning results across all tenants for tenant-wide navigation.

### 3. Scope Derivation Logic

Scope is computed at storage/index time by parsing the document's filesystem path. The parser anchors on `/docs/` to avoid the DevContainer `/workspaces/` prefix collision:

```
Path: /repo/.gyrus/docs/root/workspaces/gyrus/ticket-001.md
      ─────────────────────┬────────────────────────────────
                           │ Anchor on /docs/
      sub = "root/workspaces/gyrus/ticket-001.md"
      parts = ["root", "workspaces", "gyrus", "ticket-001.md"]
      parts[1] == "workspaces" → scope="workspace", workspace="gyrus"

Path: /repo/.gyrus/docs/root/reference/adr-001.md
      sub = "root/reference/adr-001.md"
      parts[1] == "reference" → scope="reference", workspace=""
```

Fallback chain if no `/docs/` anchor is found:
1. If path contains `/reference/` → `scope=reference`
2. If path contains `/workspaces/<name>/` via `LastIndex` → `scope=workspace, workspace=<name>`
3. Default → `scope=reference`

### 4. Two-Tier Search Prioritization

When a workspace is active (either via config or `--workspace` flag), search results are ranked:

| Condition | Score Bonus | Label |
|-----------|-------------|-------|
| `workspace == active_workspace` | +60 | Workspace Document |
| `scope == "reference"` | +20 | Reference Document |
| Other workspace | +0 | Excluded (unless `--scope all`) |

### 5. Dependency Graph Expansion in `suggest-context`

`SuggestContextWithFilter` performs two-pass expansion:
1. **Pass 1**: Retrieve workspace-scoped documents matching the prompt.
2. **Pass 2**: Walk `doc.Dependencies` of workspace documents; include any referenced `reference`-scope documents even if they didn't directly match the search query.

This ensures agents always receive the shared architecture and standards docs that a workspace ticket depends on.

### 6. CLI & MCP Surface

New flags added to `create`, `search`, and `suggest-context`:

| Flag | Command | Description |
|------|---------|-------------|
| `--scope <workspace\|reference\|all>` | search, suggest-context, create | Override scope filter |
| `--workspace <name>` | search, suggest-context, create | Override active workspace |
| `--max-docs <n>` | suggest-context | Max documents in context output |

MCP tools `gyrus_search` and `gyrus_suggest_context` gain `scope` and `workspace` parameters.

### 7. Workspace Migration (gyrus → this repo)

The existing `workspaces/main` directory in this repository is migrated to `workspaces/gyrus` to align with the explicit workspace name `gyrus` set in `.gyrus.yaml`.

---

## Alternatives Considered

### Alternative 1: Auto-detect workspace from repository directory name
* **Pros**: Zero configuration.
* **Cons**: Fragile; fails when directories are renamed, cloned with different names, or mounted in containers at paths like `/workspaces/project`. Explicitly rejected by user requirement.

### Alternative 2: Require `scope` / `workspace` frontmatter fields
* **Pros**: Explicit per-document control.
* **Cons**: Pollutes the OKF contract schema; breaks existing documents; creates schema migration burden. Rejected — path topology is sufficient.

### Alternative 3: Separate Gyrus databases per workspace
* **Pros**: Hard isolation.
* **Cons**: Eliminates cross-workspace reference queries; complicates dependency traversal; increases operational overhead.

---

## Consequences

### Positive / Gains
* **Agent Context Quality**: AI agents operating in a repository receive workspace-specific documents first, with reference docs surfaced only when relevant via dependency links.
* **Multi-repo Support**: Multiple projects can coexist in a shared Gyrus installation (`~/.gyrus/`) with clean workspace isolation.
* **Zero Schema Migration**: No OKF frontmatter changes required; scope is derived transparently from existing path topology.
* **DevContainer Safety**: Robust `/docs/`-anchored parsing eliminates false positives from container mount paths.
* **Global Reference Navigation**: Users/agents with a global `~/.gyrus.yaml` (no `workspace:`) get full reference-layer search across all tenants.

### Negative / Trade-offs
* **Requires Explicit Config**: Teams adopting workspace scoping must add `workspace: <name>` to their `.gyrus.yaml`. There is no silent auto-detection.
* **Documents Without Scope**: Legacy documents stored outside the `reference/` or `workspaces/<name>/` topology paths default to `scope=reference`, which may be unexpected for some setups.

### Risks
* **Index Staleness**: Existing SQLite indexes built before this ADR will not have `scope` or `workspace` columns. The schema migration in `initSchema()` handles this via `ALTER TABLE ... ADD COLUMN` guards, but teams must run `gyrus sync` after upgrading.

---

## Compliance & Verification

* **Unit Tests**: `TestLocalfsStoreWorkspaceScoping`, `TestSQLiteIndexerWorkspaceScopingAndPrioritization`, `TestEngineSuggestContextWithWorkspaceScopingAndDependencyExpansion`, `TestCLIScopedSearchAndSuggestContext` — all pass.
* **Full Test Suite**: `go test ./...` — all packages pass.
* **Integration**: `gyrus sync` on the `gyrus` repository re-indexes 52 documents with correct scope and workspace derivation.
* **Config Verification**: Repository `.gyrus.yaml` sets `workspace: gyrus`; workspace directory is `.gyrus/docs/root/workspaces/gyrus/`.
