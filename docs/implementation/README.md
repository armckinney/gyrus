# Gyrus Implementation Plan

This directory contains the top-level implementation plan and ticket breakdown for building **Gyrus: Unified Context & Memory Engine**.

---

## 1. Implementation Overview

The Gyrus codebase is built as a single Go module delivering the **Gyrus Core SDK** (`pkg/gyrus`), the standalone **Gyrus CLI** (`cmd/gyrus`), the **MCP Server Adapter** (`gyrus mcp serve`), and the **Open Skill Format Shim**.

---

## 2. Implementation Phases & Status Matrix

| Phase | Phase Name | Status | Showcase Product Milestone |
| :--- | :--- | :--- | :--- |
| **Phase 0** | Repository Cleanup & Project Scaffolding | `COMPLETED` | Clean Go foundation with working build pipeline |
| **Phase 1** | Core SDK & OKF Domain Model | `COMPLETED` | Reusable `pkg/gyrus` Go Library for OKF & state machine |
| **Phase 2** | Local Storage & SQLite Provider Engines | `COMPLETED` | Embedded Storage, FTS5 Search & Relationship Engine |
| **Phase 3** | Gyrus CLI Executable (`cmd/gyrus`) | `NOT STARTED` | Standalone `gyrus` Terminal Binary for Humans & CLI Agents |
| **Phase 4** | MCP Server & Open Skill Format Adapters | `NOT STARTED` | Native IDE Integration Server (Cursor, Claude Desktop, Copilot) |
| **Phase 5** | Test Profiles, CI/CD Pipeline & Packaging | `NOT STARTED` | Multi-Platform Release Binaries (Linux, macOS, Windows) |

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 0: Repository Cleanup & Project Scaffolding                               │
│  - Remove legacy `template-go` HTTP boilerplate (cmd/api, internal/handlers)     │
│  - Initialize official `github.com/armckinney/gyrus` Go module & directories    │
│  ★ SHOWCASE PRODUCT: Clean Go project foundation with working build pipeline   │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 1: Core SDK & OKF Domain Model                                           │
│  - Go module public interfaces & domain models (pkg/gyrus)                      │
│  - Open Knowledge Format (OKF) parser, YAML frontmatter & schema validator      │
│  - Lifecycle state-transition engines for all 11 document types                 │
│  ★ SHOWCASE PRODUCT: Reusable `pkg/gyrus` Go Library for OKF validation & states │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 2: Local Storage & SQLite Provider Engines                                │
│  - Localfs document store & Storage Path Resolution Hierarchy                  │
│  - SQLite DDL migrations (documents_index, documents_fts, document_edges)       │
│  - FTS5 search provider, edge-table graph traverser, and incremental Sync()     │
│  ★ SHOWCASE PRODUCT: Embedded Storage, FTS5 Search & Relationship Engine        │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 3: Gyrus CLI Executable (`cmd/gyrus`)                                     │
│  - CLI command parsing (create, get, update, link, search, suggest, sync, etc.) │
│  - Inline & file-based metadata/content flag handling                           │
│  - Programmatic exit code mapping (0..5) & terminal output formatting           │
│  ★ SHOWCASE PRODUCT: Standalone `gyrus` Terminal Binary for Humans & CLI Agents │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 4: MCP Server & Open Skill Format Adapters                                │
│  - Embedded stdio/SSE Model Context Protocol (MCP) server                       │
│  - Exposed MCP Tools (suggest_context, search, get, create, update, link)       │
│  - MCP Resources & Prompts + Open Skill Format CLI wrapper                      │
│  ★ SHOWCASE PRODUCT: Native IDE Integration Server (Cursor, Claude Desktop, etc) │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ PHASE 5: Test Profiles, CI/CD Pipeline & Packaging                              │
│  - Test Profile & Local Full Profile test suites                                │
│  - Makefile targets, .goreleaser.yaml binary compilation, & integration tests   │
│  ★ SHOWCASE PRODUCT: Multi-Platform Release Binaries (Linux, macOS, Windows)    │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Master Ticket Index

All individual task tickets are maintained in the [`tickets/`](file:///workspaces/gyrus/docs/implementation/tickets) directory. Each ticket contains a persistent status indicator (`NOT STARTED`, `IN PROGRESS`, `COMPLETED`).

| Ticket ID | Phase | Title | Status |
| :--- | :--- | :--- | :--- |
| **[GYRUS-001](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-001-repo-cleanup-bootstrap.md)** | Phase 0 | Repository Cleanup & Go Module Scaffolding | `COMPLETED` |
| **[GYRUS-101](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-101-core-sdk-interfaces.md)** | Phase 1 | Core SDK Public Interfaces & Domain Models | `COMPLETED` |
| **[GYRUS-102](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-102-okf-envelope-parser.md)** | Phase 1 | OKF Envelope Parser & Schema Validator | `COMPLETED` |
| **[GYRUS-103](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-103-lifecycle-state-machine.md)** | Phase 1 | Lifecycle State-Machine Validation Engine | `COMPLETED` |
| **[GYRUS-201](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-201-localfs-storage-provider.md)** | Phase 2 | Localfs Storage Provider & Path Resolution | `COMPLETED` |
| **[GYRUS-202](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-202-sqlite-index-fts-provider.md)** | Phase 2 | SQLite DDL, Indexing & FTS5 Search Engine | `COMPLETED` |
| **[GYRUS-203](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-203-sqlite-graph-sync.md)** | Phase 2 | SQLite Edge Graph & Incremental File Sync | `COMPLETED` |
| **[GYRUS-301](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-301-cli-command-framework.md)** | Phase 3 | CLI Command Framework & Exit Code Mapper | `NOT STARTED` |
| **[GYRUS-302](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-302-cli-core-crud.md)** | Phase 3 | CLI Core Mutation Commands (`init`, `create`, `get`, `update`) | `NOT STARTED` |
| **[GYRUS-303](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-303-cli-link-sync-validate.md)** | Phase 3 | CLI Relationship & Maintenance Commands (`link`, `unlink`, `sync`, `validate`) | `NOT STARTED` |
| **[GYRUS-304](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-304-cli-search-suggest.md)** | Phase 3 | CLI Context Discovery Commands (`search`, `suggest-context`, `schema`) | `NOT STARTED` |
| **[GYRUS-401](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-401-mcp-server-tools.md)** | Phase 4 | MCP Stdio Server & Memory Tools | `NOT STARTED` |
| **[GYRUS-402](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-402-mcp-resources-prompts-skill.md)** | Phase 4 | MCP Resources, Prompts & Open Skill Format Shim | `NOT STARTED` |
| **[GYRUS-501](file:///workspaces/gyrus/docs/implementation/tickets/GYRUS-501-testing-packaging.md)** | Phase 5 | Profile Verification Tests & GoReleaser Packaging | `NOT STARTED` |

---

## 4. Status Tracking Guidelines

* **Updating Status:** When working on a ticket, update the `> **Status:**` tag at the top of the ticket markdown file (`NOT STARTED` ➔ `IN PROGRESS` ➔ `COMPLETED`).
* **Verification Required:** A ticket cannot be marked as `COMPLETED` until all test cases and verification commands specified in the ticket have passed.
