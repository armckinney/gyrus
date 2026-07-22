---
id: ip-005-proposal-evaluation
title: Context Engine Proposals Comparative Evaluation
category: architecture
type: improvement-proposal
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Architecture Comparison: Proposals 1, 2, and 3

This document compares the core architectural designs of the three context manager proposals in minimal words, and documents what was sourced from each proposal to create the final **Gyrus** specification.

---

## 1. Individual Architecture Summaries

### Proposal 1 (AGY)
* **Storage:** Centralized Azure Cosmos DB (JSON contracts with `/owner_group` partition keys).
* **Interface:** CLI-first tool (`ucl`) + server-side Web Wiki Portal (serving dynamic `/llms.txt` maps).
* **Security:** User-delegated Azure AD / Entra ID JWT validation at the API Gateway.
* **History:** Dedicated `document_history` container in Cosmos DB storing full snapshots.

### Proposal 2 (ChatGPT)
* **Storage:** Progressive multi-tier scaling: Local Filesystem (Markdown/YAML) ➔ Postgres DB ➔ OpenSearch.
* **Interface:** Single Go binary (`agent-memory`) embedding CLI, MCP stdio server, and local Go+HTMX web server.
* **Relationships:** Simple SQL/SQLite edge table (`document_edges` mapping connections).
* **Validation:** Core Engine state-machine schema and transition checks (e.g. `accepted` ──► `superseded`).

### Proposal 3 (Claude)
* **Storage:** Dedicated dual-layer Graph-RAG (pgvector/Qdrant vector store + Neo4j/LightRAG graph database).
* **Interface:** MCP-first server (stdio/SSE) as primary agent gateway + Next.js browser portal + PR CI/CD injector.
* **Relationships:** Native property graph traversal (multi-hop lineage checks).
* **Retrieval:** Vector similarity lookup ➔ Graph expansion ➔ Cross-encoder reranker ➔ Unified Markdown stream.

---

## 2. Core Architectural Differences

| Architectural Layer | Proposal 1 (AGY) | Proposal 2 (ChatGPT) | Proposal 3 (Claude) |
| :--- | :--- | :--- | :--- |
| **Durable Store** | Cosmos DB (Central Cloud) | Local Filesystem ➔ Postgres | Vector & Graph Databases |
| **Ecosystem footprint** | CLI + API Gateway Docker | Single Go binary (`agent-memory`) | Multi-service API cluster |
| **Agent Interface** | CLI-First | CLI + MCP Server | MCP-First (stdio/SSE) |
| **Relational Model** | Metadata dependencies list | Relational SQL Edge Table | Native Graph DB (Neo4j) |
| **Validation Point** | API Gateway Filter | Core Engine State Machine | Ingestion parsing |
| **Search Engine** | Cosmos DB index | SQLite FTS ➔ Postgres FTS ➔ OpenSearch | Vector DB cosine similarity |
| **Auth & Security** | User-Delegated Entra ID JWT | Local session config / RBAC | API token keys |

---

## 3. Provenance & Decisions Sourced for Gyrus (Final Spec)

The final Gyrus specification synthesizes key design choices from all three proposals:

### Sourced from Proposal 1 (AGY)
* **Storage Provider Interface:** The swappable Go `StorageProvider` Repository Pattern interface, allowing the backend to scale from filesystem to databases without client rewrites.
* **User-Delegated Auth:** Inheriting local developer Entra ID/Azure CLI tokens (`az account get-access-token`) to maintain auditable, group-level security.
* **Dynamic `/llms.txt` Sitemap:** An AI-optimized plain-text index endpoint served by the web portal for remote agents.

### Sourced from Proposal 2 (ChatGPT)
* **Local-First Storage:** Storing documents as Markdown + YAML frontmatter to support local offline usage and CI/CD testing.
* **Relational SQLite Edges:** Using a SQL `document_edges` table rather than provisioning a dedicated graph database for the local profile.
* **Core Validation Engine:** Enforcing metadata schema validation and lifecycle state transitions (e.g. proposed ──► accepted) directly at the engine boundary.
* **Go CLI Single Binary:** Packaging the core CLI client and local MCP server into a single Go-compiled binary.

### Sourced from Proposal 3 (Claude)
* **MCP Interface SDK:** Exposing Gyrus capabilities (search, read, write) via standard Model Context Protocol tools, resources, and prompts.
* **Standard Templates:** Structural Markdown designs for Architecture Decision Records (ADRs), Tech Plans, and Guides.

### Core Exclusions (Deviations from Proposals)
* **No Local Web Wiki Server:** Excluded Proposal 2's local Go+HTMX web server and Proposal 3's local Next.js dashboard. Standard local Markdown viewers (Obsidian, VS Code) act as the local wiki.
* **No Graph or Vector DBs for MVP:** Dropped Neo4j and Qdrant from the initial local deployment footprint to eliminate infrastructure hosting overhead.
