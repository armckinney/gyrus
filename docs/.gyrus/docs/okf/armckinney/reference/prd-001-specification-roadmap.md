---
id: prd-001-specification-roadmap
title: Gyrus Specification Implementation Roadmap & TODOs
category: technical
type: prd
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus Specification Implementation Roadmap (`docs/TODO.md`)

This document tracks all outstanding features, architectural enhancements, and planned provider drivers needed to achieve 100% full specification compliance beyond the MVP release.

---

## 📋 Outstanding Specification Features

### 1. CI/CD, Container & Deployment Packaging

- [ ] **GitHub Actions Automated CLI Binary Release**: Set up GitHub Actions workflow to cross-compile, package (tar.gz/zip), generate SHA-256 checksums, and publish standalone `gyrus` CLI executables across Linux (`amd64`/`arm64`), macOS (`amd64`/`arm64`), and Windows (`amd64`) for GitHub releases on git tag pushes.
- [ ] **GitHub Actions MCP Container Image Publishing**: Build and publish multi-arch Docker container images (`linux/amd64`, `linux/arm64`) for the Gyrus MCP server to GitHub Container Registry (`ghcr.io/armckinney/gyrus:latest` and version tags) via GitHub Actions.
- [ ] **Helm Chart / Kubernetes Deployment**: Official Helm chart for hosting Gyrus central memory service in Kubernetes clusters.

---

### 2. Additional Provider Drivers

- [ ] **Git Storage Driver (`git`)**: Direct remote Git repository persistence via GitHub/Bitbucket APIs without requiring local workspace clones.
- [ ] **Cloud Object Storage Driver (`blob`)**: AWS S3, Azure Blob Storage, and Google Cloud Storage drivers for cloud-native OKF document bundles.
- [ ] **PostgreSQL Index & Storage Driver (`postgres`)**: Centralized PostgreSQL database backend for enterprise deployments.
- [ ] **PostgreSQL FTS Search Engine (`postgres_fts`)**: PostgreSQL `tsvector` and `tsquery` full-text search engine.
- [ ] **Vector Embedding Search Driver (`vector`)**: Semantic vector search provider (using pgvector or local embeddings) for hybrid keyword + vector context retrieval.

---

### 3. Transport & Networking Enhancements

- [ ] **MCP SSE/HTTP Transport (`gyrus mcp serve --transport sse`)**: Add Server-Sent Events (SSE) and HTTP listener mode for remote MCP server consumption over network endpoints.
- [ ] **GRPC Core SDK Endpoint**: Expose high-performance gRPC service definition for multi-language Core SDK bindings (Python, TypeScript).

---

### 4. Web UI & Visualization (Specification 04)

- [ ] **Embedded Web Dashboard (`gyrus ui`)**: Embedded single-page application (SPA) for visual graph topology exploration, ADR browsing, and document editing.
- [ ] **Interactive Dependency Graph Visualizer**: D3.js or Cytoscape.js interactive node-edge graph visualization of document links (`depends_on`, `supersedes`, `implements`).

---

### 5. Enterprise Security & Multi-Tenancy

- [ ] **Owner-Group Access Control (RBAC)**: Fine-grained Role-Based Access Control enforcing read/write permissions per `owner_group`.
- [ ] **Authentication Tokens & OAuth2**: API token validation for HTTP/SSE MCP servers and centralized team instances.

---

### 6. Context Hygiene & Document Lifecycle Cleanup

- [ ] **Stale & Low-Quality Context Cleanup**: Design and implement strategies for identifying, reviewing, and cleaning up old, superseded, or low-quality context and concept documents (e.g., automated staleness detection, decay/quality scoring, garbage collection routines, and `deprecated`/`archived` state sweeps).

---

### 7. Branding & Visual Identity

- [ ] **Create Official Project Logo**: Design and create an official Gyrus logo and visual branding assets for the GitHub repository, root README, documentation, and web app.


