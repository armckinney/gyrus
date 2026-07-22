---
id: prd-001-specification-roadmap
title: Gyrus Specification Implementation Roadmap & TODOs
category: technical
type: prd
format: ""
owner_group: armckinney
version: 2
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

- [ ] **GitHub Actions Automated CLI Binary Release**: Set up GitHub Actions workflow (`.github/workflows/release.yml`) to cross-compile, package (`tar.gz`/`zip`), generate SHA-256 checksums, and publish standalone `gyrus` CLI executables across Linux (`amd64`/`arm64`), macOS (`amd64`/`arm64`), and Windows (`amd64`) on git tag pushes (`v*`).
- [ ] **GitHub Actions MCP Container Image Publishing**: Build and publish multi-arch Docker container images (`linux/amd64`, `linux/arm64`) for the Gyrus MCP server to GitHub Container Registry (`ghcr.io/armckinney/gyrus:latest` and version tags) via GitHub Actions (`.github/workflows/docker-mcp.yml`).
- [ ] **Create Official Project Logo & Visual Assets**: Design and create an official Gyrus logo and visual branding assets for the GitHub repository, root README, documentation, and web app.

---

## 🌟 Phase 2: Version 1.0 Release Scope

Target: Expand data providers, transport interfaces, context hygiene, and embed the Web UI visual dashboard.

### 2.1 Additional Storage & Search Provider Drivers
- [ ] **Git Storage Driver (`git`)**: Direct remote Git repository persistence via GitHub/Bitbucket APIs without requiring local workspace clones.
- [ ] **Cloud Object Storage Drivers (`s3`, `blob`)**: AWS S3, Azure Blob Storage, and Google Cloud Storage drivers for cloud-native OKF document bundles.
- [ ] **PostgreSQL Index & Storage Driver (`postgres`)**: Centralized PostgreSQL database backend for enterprise deployments.
- [ ] **PostgreSQL FTS Search Engine (`postgres_fts`)**: PostgreSQL `tsvector` and `tsquery` full-text search engine.
- [ ] **Vector Embedding Search Driver (`vector`)**: Semantic vector search provider (using pgvector or local embeddings) for hybrid BM25 keyword + vector context retrieval.

### 2.2 Transport & Networking Enhancements
- [ ] **MCP SSE/HTTP Listener Mode (`gyrus mcp serve --transport sse`)**: Server-Sent Events (SSE) and HTTP listener mode for remote MCP server consumption over network endpoints.

### 2.3 Web UI & Visualization Surface
- [ ] **Embedded Web Dashboard (`gyrus ui`)**: Embedded single-page application (SPA) for visual graph topology exploration, ADR browsing, and document editing.
- [ ] **Interactive Dependency Graph Visualizer**: D3.js or Cytoscape.js interactive node-edge graph visualization of document links (`depends_on`, `supersedes`, `implements`).
- [ ] **AI Context Retrieval & Search Chatbot Agent**: Embedded conversational AI agent in the Web UI for natural language query answering, interactive context retrieval, multi-document synthesis (integrating `gyrus suggest-context`), and guided contract/ADR drafting.

### 2.4 Context Hygiene & Governance
- [ ] **Stale & Low-Quality Context Cleanup**: Automated staleness detection, decay/quality scoring, garbage collection routines, and `deprecated`/`archived` state sweeps.

---

## 🔮 Phase 3: Future & Enterprise Extensions

Target: Multi-tenant enterprise deployment packaging, RBAC, and multi-language SDK bindings.

- [ ] **Helm Chart / Kubernetes Deployment**: Official Helm chart for hosting central Gyrus memory services in Kubernetes clusters.
- [ ] **gRPC Core SDK Endpoint**: High-performance gRPC service definitions for multi-language Core SDK bindings (Python, TypeScript).
- [ ] **Owner-Group Access Control (RBAC)**: Fine-grained Role-Based Access Control enforcing read/write permissions per `owner_group`.
- [ ] **Authentication Tokens & OAuth2**: API token validation for HTTP/SSE MCP servers and centralized team instances.
