---
id: adr-002-simplified-directory-topology
title: Simplified Directory Topology (.gyrus/ Root & .gyrus/docs/ Layout)
category: architecture
type: adr
format: markdown
owner_group: armckinney
version: 1
status: accepted
last_modified_by: antigravity
last_updated: 2026-08-12T05:40:00Z
tags: [architecture, storage, topology, gyrus]
---

# ADR 002: Simplified Directory Topology (.gyrus/ Root & .gyrus/docs/ Layout)

## Context
Previously, Gyrus stored documents in a nested hierarchy with a mandatory `okf/` subfolder underneath a designated `docs/` storage root:
`<storage_root>/okf/<owner_group>/<categorySubdir>/<id>.md`

When `.gyrus/docs` was configured as the storage root, document paths resulted in double-nesting (`.gyrus/docs/okf/<owner_group>/...`), while SQLite index files (`index.db`, `index.db-wal`) were placed alongside the `okf/` folder inside `.gyrus/docs/`.

This structure presented two drawbacks:
1. **Redundant Nesting**: Having `okf/` inside `.gyrus/docs/` created unnecessary directory depth (`.gyrus/docs/okf/`).
2. **System Database Co-location**: Storing `index.db` inside `.gyrus/docs/` mixed runtime database artifacts into the documentation root directory.

## Decision
We adopt `.gyrus/` as the primary storage root and simplify the document bundle path:

1. **Storage Root (`.gyrus/`)**: `.gyrus/` serves as the top-level tool directory (similar to `.git/` or `.vscode/`). System runtime databases (`index.db`, `index.db-shm`, `index.db-wal`) and configuration (`config.yaml`) live directly at the root of `.gyrus/`.
2. **Document Collection Root (`.gyrus/docs/`)**: Document bundles sit cleanly inside `.gyrus/docs/<owner_group>/<categorySubdir>/<id>.md`, replacing the redundant `okf/` folder.
3. **Collision Safety**: Because `index.db` sits at `.gyrus/index.db` (one level up), owner group directories inside `.gyrus/docs/` (e.g. `.gyrus/docs/armckinney/`) cannot collide with system engine files.

## Consequences

### Positive
- **Simpler, Intuitive Paths**: Document paths become `.gyrus/docs/<owner_group>/reference/` and `.gyrus/docs/<owner_group>/workspaces/`.
- **Clean System File Isolation**: Database files (`index.db`) are separated from human/agent-readable Markdown files.
- **Unified Across Providers**: Default storage key prefix across local disk, Git, and Cloud Blob drivers becomes `docs/` instead of `okf/`.

### Risks & Mitigations
- **Path Migration**: Existing files located under `.gyrus/docs/okf/` are moved to `.gyrus/docs/`. References in documentation and tests are updated to reflect the new structure.
