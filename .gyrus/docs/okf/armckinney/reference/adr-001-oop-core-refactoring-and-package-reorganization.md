---
id: adr-001-oop-core-refactoring-and-package-reorganization
title: OOP Core Refactoring & Package Reorganization
category: architecture
type: adr
format: ""
owner_group: armckinney
version: 1
status: proposed
last_modified_by: ""
last_updated: 2026-08-12T03:06:59Z
tags:
    - adr
    - refactoring
    - architecture
    - phase-2-2
---

# ADR 001: OOP Core Refactoring & Package Reorganization

## Context

As Gyrus expanded across Phase 1 (MVP) and Phase 2.1 (adding Git, Cloud Blob, PostgreSQL, and Vector storage/search drivers), the codebase accumulated several technical debt items and anti-patterns:

1. **Implicit Global Side-Effects**: Cobra CLI commands in `internal/cli/commands/` registered themselves into global state via Go package `init()` functions. This made unit testing difficult and obscured dependency graphs.
2. **Duplicated Engine & Provider Initialization**: Both CLI command handlers and MCP tool handlers independently initialized config files, provider factories, and lifecycle engines.
3. **Scattered Root Directory**: Distribution scripts (`install.sh`), agent skills (`skills/`), container definitions (`Dockerfile`), and scratch scripts (`patch.go`) were scattered in the repository root.
4. **Lack of Layered Architecture**: Domain logic (`lifecycle`, `okf`), transport layers (`cli`, `mcp`), and infrastructure drivers (`provider`) were flatly organized under `internal/`.

## Decision

We decide to execute a comprehensive OOP core refactoring and package reorganization aligned 1:1 with the Gyrus System Architecture:

1. **Consolidate Packaging Artifacts**: Move `skills/` and `install.sh` into standard `packaging/` (`packaging/skills/`, `packaging/install.sh`). Delete obsolete root scratch script `patch.go` and eliminate root `Dockerfile` using GoReleaser native container builds or `packaging/Dockerfile`.
2. **Decompose Providers by Architectural Capability**: Reorganize `internal/provider/` into explicit capability packages matching the system architecture:
   - `internal/provider/storage/`: Concept Persistence Service (`localfs`, `git`, `blob`, `sqlite`, `postgres`)
   - `internal/provider/search/`: Search Service (`fts5`, `postgres_fts`, `vector`)
   - `internal/provider/index/`: Concept Index Service (`okf_indexer`, `sqlite_indexer`, `postgres_indexer`)
   - `internal/provider/graph/`: Knowledge Graph Service (`okf_graph`, `sqlite_graph`, `postgres_graph`)
3. **Consolidate Storage Formats (`OKF` vs `Relational`)**: Introduce `internal/format/` (`format.OKF` for Markdown/frontmatter backends vs `format.Relational` for SQL backends) to eliminate redundant Markdown/YAML serialization across file and object storage drivers.
4. **Adopt Layered Clean Architecture**: Reorganize `internal/` into explicit layer packages (`internal/transport/cli`, `internal/transport/mcp`, `internal/domain/lifecycle`, `internal/domain/okf`, `internal/provider`, `internal/app`).
5. **DRY Transport Adapters with Dependency Injection**: Centralize 100% of engine lifecycle management in `internal/app/` (`app.App`) and make `transport/cli` and `transport/mcp` minimal, thin transport wrappers over `app.Engine()`. Replace all Cobra `init()` side-effects with explicit `New<Command>Cmd(app *app.App)` constructors.

## Consequences

### Positive
- **Elimination of Global State**: Testing CLI commands and MCP adapters can be performed cleanly using mock application dependencies without global side-effects.
- **Strict 1:1 Architectural Alignment**: Code layout mirrors the system architecture diagram in `README.md`.
- **Zero Duplication**: Shared OKF serializer for file backends; 100% DRY engine interface across CLI subcommands and MCP tools.
- **Idiomatic Go Packaging**: Standard `packaging/` layout and clean root workspace.

### Trade-offs / Work
- Internal package import paths will be updated across all Go files.
- CI workflow definitions (`wf-release.yml`) and GoReleaser hooks (`.goreleaser.yaml`) need path adjustments.
