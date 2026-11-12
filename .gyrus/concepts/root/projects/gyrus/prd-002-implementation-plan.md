---
title: Gyrus Master Implementation Plan (Phases 0 through 5)
id: prd-002-implementation-plan
type: prd
format: ""
category: product
version: 2
status: archived
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

> Historical record. This document describes earlier work and is not a current setup or capability contract. Use the current configuration reference and architecture index.

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
| **Phase 3** | Gyrus CLI Executable (`cmd/gyrus`) | `COMPLETED` | Standalone `gyrus` Terminal Binary for Humans & CLI Agents |
| **Phase 4** | MCP Server & Open Skill Format Adapters | `COMPLETED` | Native IDE Integration Server (Cursor, Claude Desktop, Copilot) |
| **Phase 5** | Test Profiles, CI/CD Pipeline & Packaging | `COMPLETED` | Multi-Platform Release Binaries (Linux, macOS, Windows) |

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

All individual task tickets are maintained in the [`tickets/`](./) directory. Each ticket contains a persistent status indicator (`NOT STARTED`, `IN PROGRESS`, `COMPLETED`).

| Ticket ID | Phase | Title | Status |
| :--- | :--- | :--- | :--- |
| **[GYRUS-001](ticket-gyrus-001.md)** | Phase 0 | Repository Cleanup & Go Module Scaffolding | `COMPLETED` |
| **[GYRUS-101](ticket-gyrus-101.md)** | Phase 1 | Core SDK Public Interfaces & Domain Models | `COMPLETED` |
| **[GYRUS-102](ticket-gyrus-102.md)** | Phase 1 | OKF Envelope Parser & Schema Validator | `COMPLETED` |
| **[GYRUS-103](ticket-gyrus-103.md)** | Phase 1 | Lifecycle State-Machine Validation Engine | `COMPLETED` |
| **[GYRUS-201](ticket-gyrus-201.md)** | Phase 2 | Localfs Storage Provider & Path Resolution | `COMPLETED` |
| **[GYRUS-202](ticket-gyrus-202.md)** | Phase 2 | SQLite DDL, Indexing & FTS5 Search Engine | `COMPLETED` |
| **[GYRUS-203](ticket-gyrus-203.md)** | Phase 2 | SQLite Edge Graph & Incremental File Sync | `COMPLETED` |
| **[GYRUS-301](ticket-gyrus-301.md)** | Phase 3 | CLI Command Framework & Exit Code Mapper | `COMPLETED` |
| **[GYRUS-302](ticket-gyrus-302.md)** | Phase 3 | CLI Core Mutation Commands (`init`, `create`, `get`, `update`) | `COMPLETED` |
| **[GYRUS-303](ticket-gyrus-303.md)** | Phase 3 | CLI Relationship & Maintenance Commands (`link`, `unlink`, `sync`, `validate`) | `COMPLETED` |
| **[GYRUS-304](ticket-gyrus-304.md)** | Phase 3 | CLI Context Discovery Commands (`search`, `suggest-context`, `schema`) | `COMPLETED` |
| **[GYRUS-401](ticket-gyrus-401.md)** | Phase 4 | MCP Stdio Server & Memory Tools | `COMPLETED` |
| **[GYRUS-402](ticket-gyrus-402.md)** | Phase 4 | MCP Resources, Prompts & Open Skill Format Shim | `COMPLETED` |
| **[GYRUS-501](ticket-gyrus-501.md)** | Phase 5 | Profile Verification Tests & GoReleaser Packaging | `COMPLETED` |

---

## 4. Status Tracking Guidelines

* **Updating Status:** When working on a ticket, update the `> **Status:**` tag at the top of the ticket markdown file (`NOT STARTED` ➔ `IN PROGRESS` ➔ `COMPLETED`).
* **Verification Required:** A ticket cannot be marked as `COMPLETED` until all test cases and verification commands specified in the ticket have passed.
