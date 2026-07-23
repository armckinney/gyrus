---
id: ip-003-claude-proposal
title: Claude Agent Skill Integration Proposal
category: architecture
type: improvement-proposal
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Architecture Proposal: Unified Context Creation & Query Engine

This document outlines the end-to-end technical architecture, component breakdown, and technology selections for a Unified Context Creation and Query Engine. This platform is designed to serve as a single source of truth for both autonomous AI agents (e.g., Claude Code, GitHub Copilot, Antigravity) and human developers during the engineering lifecycle.

## 1. System Overview & Core Topology

The engine leverages a dual-layer Graph-RAG architecture to store and manage highly relational product and engineering documents (such as PRDs, Technical Plans, and Architecture Decision Records). By mapping both the semantic meaning and the structural lineage of these documents, it guarantees high-precision context retrieval, preventing agents from violating historical engineering constraints.

[ Developer IDE / Browser ]
  └── Clients: Claude Code / Antigravity / Human Web UI
        │
        ▼ (JSON-RPC 2.0 via stdio/SSE or REST)
[ MCP Server & Core Engine Layer ]
  ├── Cores: Ingestion Worker & Query Resolver
  └── Middleware: Semantic Router & Graph Linker
        │
        ├── (Vector Queries + Metadata) ──► [ Vector Store: pgvector / Qdrant ]
        └── (Graph/Lineage Traversals) ───► [ Graph Store: Neo4j / LightRAG ]

---

## 2. Platform Component Breakdown ("The Products")

To fully realize this platform as an internal developer service, the following key sub-systems must be built:

* **MCP Server Execution Interface:** The primary communication gateway for AI agents. Implements the Model Context Protocol (MCP) using a standard input/output (`stdio`) or Server-Sent Events (`SSE`) protocol to natively expose search tools and document resources to agent hosts.
* **Document Ingestion & Parsing Engine:** A backend pipeline that watches documentation directories or handles manual uploads. It performs hierarchical Markdown chunking (splitting sections structurally by `# H1`, `## H2`, and `### H3` headers) and uses entity extraction to isolate system entities, active statuses, and cross-document dependencies.
* **Hybrid Query Resolver (The Orchestrator):** The central routing service that handles context requests. It executes parallel lookups across vector and graph database layers, feeds raw results through a cross-encoder reranker, and linearizes the final text into clean Markdown context packages for the agent.
* **Developer CLI Tool:** A local terminal utility used by human engineers to query the context engine directly, update document status, check structural lineages, or initialize configuration files within local repositories.
* **Human-in-the-Loop Web Portal:** An interactive, browser-based dashboard allowing developers to search documents visually, view cross-document dependency lineage graphs, and override AI-extracted connections or update lifecycle statuses (e.g., marking an ADR as `Superseded` or `Deprecated`).
* **CI/CD Context Injector (Git Integrations):** A GitHub Action or Git webhook service that automatically runs context queries based on the files altered in a Pull Request, leaving automated summaries of historical constraints as comments for human code reviewers.

---

## 3. Technology Stack Recommendations

The platform components will utilize a modern, highly efficient infrastructure footprint designed for structural metadata management:

| Layer / Component | Technology Choice | Key Rationale |
| :--- | :--- | :--- |
| **Vector Database** | PostgreSQL with `pgvector` (or Qdrant) | Provides high-performance vector embeddings matched with rigid, fast metadata filtering (e.g., filtering queries by document type, author, or active status). |
| **Graph Database** | Neo4j (or LightRAG framework) | Natively maps explicit document relationships (e.g., `SUPERSEDES`, `MOTIVATES`, `IMPLEMENTS`) to allow fast multi-hop lineage traversals. |
| **Interface Layer (Agents)** | Model Context Protocol (MCP) SDK | Universal open standard natively supported by next-gen IDE tools and coding agents, minimizing client-side integration friction. |
| **Context Reranking** | Cohere Rerank or `bge-reranker-large` | Ensures that long, verbose prose documents are precisely prioritized so only high-signal text chunks enter the limited LLM context window. |
| **Human Web UI** | Next.js with Tailwind CSS & Cytoscape.js | Enables fast server-rendered searches along with highly interactive network graph visualizations for human understanding. |

---

## 4. End-to-End Query Lifecycle

1.  **Trigger:** An agent client or developer encounters a component and requests context via the engine interface.
2.  **Vector Search:** The query text is vectorized and a cosine similarity lookup is performed against the vector database to discover the top semantically relevant documentation chunks.
3.  **Graph Expansion:** The engine extracts the `document_id` values from the top vector matches and queries the graph database to traverse related items (such as the motivating PRD or a newer ADR that supersedes it).
4.  **Reranking & Assembly:** The combined data is deduplicated, sorted by relevance via the cross-encoder, and compiled into a single, cohesive Markdown document streamed directly back to the requesting client.
