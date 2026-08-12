---
id: prd-001-specification-roadmap
title: Gyrus Specification Implementation Roadmap & TODOs
category: technical
type: prd
format: ""
owner_group: armckinney
version: 8
status: active
tags:
  - roadmap
  - release-phases
  - specification
dependencies: []
---

# Gyrus Specification Implementation Roadmap

This document serves as the central master roadmap for Gyrus. All planned features, architectural enhancements, and provider drivers are organized into three sequential release phases: **MVP Wrap-Up**, **Version 1.0**, and **Future Extensions**.

---

## 🚀 Phase 1: MVP Wrap-Up (Immediate Release)

Target: Complete core automated release workflows and visual branding to finalize the initial open-source MVP.

- [x] **GitHub Actions Automated CLI Binary Release**: Set up GitHub Actions workflow (`.github/workflows/wf-release.yml`) to cross-compile, package (`tar.gz`/`zip`), generate SHA-256 checksums, and publish standalone `gyrus` CLI executables across Linux (`amd64`/`arm64`), macOS (`amd64`/`arm64`), and Windows (`amd64`) via Verge version bumping (`v*`).
- [x] **GitHub Actions MCP Container Image Publishing**: Build and publish multi-arch Docker container images (`linux/amd64`, `linux/arm64`) for the Gyrus MCP server to GitHub Container Registry (`ghcr.io/armckinney/gyrus:latest` and version tags) via GitHub Actions (`.github/workflows/wf-release.yml`).
- [x] **Create Official Project Logo & Visual Assets**: Design and create an official Gyrus logo and visual branding assets for the GitHub repository, root README, documentation, and web app.

---

## 🌟 Phase 2: Version 1.0 Release Scope

Target: Expand data providers, transport interfaces, context hygiene, test suite modernization, agent plugin distribution, global config, and embedded visualization surface.

### 2.1 Additional Storage & Search Provider Drivers
- [x] **Git Storage Driver (`git`)**: Direct remote Git repository persistence via `go-git` (`GYRUS-201`) without requiring local workspace clones.
- [x] **Cloud Object Storage Drivers (`s3`, `blob`)**: AWS S3, Azure Blob Storage, and Google Cloud Storage drivers (`GYRUS-202`) via `gocloud.dev/blob` for cloud-native OKF document bundles.
- [x] **PostgreSQL Index & Storage Driver (`postgres`)**: Centralized PostgreSQL database backend (`GYRUS-203`) for enterprise deployments using `pgx/v5`.
- [x] **PostgreSQL FTS Search Engine (`postgres_fts`)**: Native PostgreSQL `tsvector` and `tsquery` full-text search engine (`GYRUS-204`).
- [x] **Vector Embedding Search Driver (`vector`)**: Semantic vector search provider (`GYRUS-205`) supporting Local Ollama, OpenAI embeddings, and Reciprocal Rank Fusion (RRF) hybrid search.
- [x] **DevContainer Ollama Sidecar**: Implement DevContainer Docker Compose sidecar service (`ollama/ollama`) with pre-configured embedding models (`nomic-embed-text`) for zero-setup local vector search testing.

### 2.2 Core Refactoring & OOP Architecture
- [x] **Refactor repo to OOP structure and cleanup**: Refactor Go core packages into clean layered OOP architectures (`internal/domain/okf`, `internal/domain/lifecycle`, `internal/format`, `internal/provider/{storage,search,index,graph}`, `internal/app`, `internal/cli`, `internal/mcp`), consolidate provider factories, add shared connection pools (`internal/provider/db`), and streamline packaging.

### 2.3 Transport & Networking
- [x] **Standard Stdio MCP Transport (`gyrus mcp serve`)**: High-performance, zero-latency Stdio (standard I/O) process transport for local agent integration (`agy`, Claude Code, Cursor, Copilot) and containerized MCP execution (`docker run -i ghcr.io/armckinney/gyrus:latest`).

### 2.4 Test Suite Modernization & Stabilization
- [ ] **Modernize Testing Suite & Coverage**: Upgrade test harness across core packages (`internal/domain`, `internal/format`, `internal/provider`, `internal/cli`, `internal/mcp`), adding comprehensive mock providers, end-to-end integration matrices, and benchmark tests.
- [ ] **Automated Agent Integration Test Framework**: Automated test suites verifying that AI agents (`agy`, Claude Code, GitHub Copilot, Cursor) actually invoke Gyrus CLI subcommands and MCP tools correctly when prompted with realistic engineering tasks.

### 2.5 Persistence Layer Schema Storage & Remote Linkage
- [ ] **Persistence Layer Schema Storage & Remote Linkage**: Implement core interface (CLI and MCP commands) for storing OKF contract schemas directly in the persistence layer (`storage_provider`), with schemas stored remotely and linked to the active storage provider via enforced locations (`.gyrus/schemas/`).

### 2.6 Agent Plugin Packaging, Distribution & Init Revamp
- [ ] **Revamp `gyrus init` CLI Entrypoint**: Reorganize `gyrus init` into explicit subcommands:
  - `gyrus init config`: Interactive or flag-driven creation wizard for `.gyrus.yaml` workspace configuration (setting storage backend `localfs`, `git`, `blob`, `postgres`, `vector`, default owner group, and search provider).
  - `gyrus init client`: Explicit client distribution installer that installs Agent Plugins and equips Gyrus skills/MCP servers across target agent tools (`agy cli`, `copilot/codex cli`, `claude cli`).
- [ ] **Agent Plugin Packaging**: Package Gyrus skills, subagents, and MCP tools into official Agent Plugins supporting both the [Agent Plugins Standard](https://agent-plugins.org/) format (Google Developers: https://developers.googleblog.com/agent-plugins-package-your-skills-tools-and-more/) and full compatibility with `agy cli`, `copilot/codex cli`, and `claude cli` via `plugin.json` for zero-config agent discovery, distribution, and runtime sidecar loading.
- [ ] **Interactive Demo Showcase in README**: Add interactive demo recording/GIF showcase to `README.md` highlighting `gyrus init`, `gyrus suggest-context`, and agent MCP workflows.

### 2.7 Global Configuration (`~/.gyrus.yaml`)
- [ ] **Global Config Support (`~/.gyrus.yaml`)**: Support user-wide global configuration files in user home (`~/.gyrus.yaml` and `~/.config/gyrus/config.yaml`) for setting user-level defaults across workspace boundaries.

### 2.8 Context Hygiene & Governance
- [ ] **Stale & Low-Quality Context Cleanup**: Automated staleness detection, decay/quality scoring, garbage collection routines, and `deprecated`/`archived` state sweeps.

### 2.9 Web UI & Interactive Visualization Surface
- [ ] **Embedded Web Dashboard (`gyrus ui`)**: Embedded single-page application (SPA) for visual graph topology exploration, ADR browsing, and document editing.
- [ ] **Interactive Dependency Graph Visualizer**: D3.js or Cytoscape.js interactive node-edge graph visualization of document links (`depends_on`, `supersedes`, `implements`).

---

## 🔮 Phase 3: Future & Enterprise Extensions

Target: Multi-tenant enterprise capabilities, full web app conversational chatbot agent, RBAC, and multi-language SDK bindings.

- [ ] **Full Web App Conversational AI Context Retrieval Chatbot Agent**: Embedded conversational AI agent in the Web UI for natural language query answering, interactive context retrieval, multi-document synthesis (integrating `gyrus suggest-context`), and guided contract/ADR drafting.
- [ ] **gRPC Core SDK Endpoint**: High-performance gRPC service definitions for multi-language Core SDK bindings (Python, TypeScript).
- [ ] **Owner-Group Access Control (RBAC)**: Fine-grained Role-Based Access Control enforcing read/write permissions per `owner_group`.
