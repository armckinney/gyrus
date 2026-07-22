# Gyrus: Unified Context & Memory Engine
## Architecture Specification Index

Gyrus is a modular context and memory management system designed to serve as the single source of truth for both autonomous AI agents (such as Claude Code, Cursor, GitHub Copilot, and Antigravity) and human developers.

---

## 1. Core Architecture Blueprint

The system is centered around the **Gyrus Core SDK**—a composable Go framework exposing OKF APIs over pluggable storage, indexing, knowledge graph, and search providers.

```
+-----------------------------------------------------------------------------------------------------------------+
|                                                   CLIENT TIER                                                   |
|  [ Applications ]       [ CI/CD & Scripts ]        [ Agentic Tool(s) ]                      [ Human Users ]     |
|   (Direct SDK)              (CLI execution)     ┌───────┴──────┐                            ┌───────┴────────┐  |
|                                                 ▼              ▼                            ▼                ▼  |
|                                             [ Skill ]    [ MCP Server ]                [ Git Repo ]  [ Web App ]|
|                                            (Open Skill)   (stdio/SSE)                   (Direct)     (Chatbot)  |
+─────────────────────────────┬───────────────────┬──────────────┬────────────────────────────┬───────────────────+
                              │                   │              │                            │
                              ▼                   ▼              ▼                            │
+─────────────────────────────────────────────────────────────────────────────────────────+   │
|                                    GYRUS CORE SDK                                       |   │
|  * Applies (MVP): Storage (OKF/localfs), Index (OKF), Graph (OKF), Search (None)        |   │
|  * Implements: OKF APIs & Validation State Machines                                     |   │
|  * Provider Interfaces: StorageProvider, IndexProvider, KnowledgeGraphProvider, Search  |   │
+─────────────────────────────────────────────────────────┬───────────────────────────────+   │
                                                          │ (Structures & Writes)             │ (Reads)
                                                          ▼                                   ▼
+─────────────────────────────────────────────────────────────────────────────────────────────────────────────────+
|                                    OKF BUNDLE TOPOLOGY (Git Repository)                                         |
|  okf/ (Flexible top layer)                                                                                       |
|  ├── <team>/ (Security boundary N-layers)                                                                        |
|  │     ├── reference/ (Global reference doc types: ADRs, Standards, Specs...)                                   |
|  │     └── workspaces/ (Repo scoping, points back to reference docs)                                           |
|  │           └── <repo-x>/ (Workspace-specific instructions & context)                                          |
+-----------------------------------------------------------------------------------------------------------------+
```

---

## 2. Modular Specifications

1. **[01. Open Knowledge Format (OKF) Contract Schema & Topology](01_contract_schema.md):** Defines metadata envelope rules, Open Knowledge Format (OKF) attributes, tags, standard document types, and the **OKF Bundle Topology**.
2. **[02. CLI Interface & Programmatic Exit Codes](02_cli_interface.md):** Outlines the `gyrus` command-line reference, flag signatures, `gyrus sync` indexing, relationship linking, and exit code definitions.
3. **[03. MCP Interface & Skill Specifications](03_mcp_interface.md):** Formulates tools, resources, prompt templates, and Open Skill Format integrations exposed natively to AI editors (Cursor, Claude Desktop, Copilot).
4. **[04. Core Engine SDK & SQLite DDL](04_core_engine.md):** Details the **Gyrus Core SDK** architecture, complete SQLite DDL (`documents_index`, `documents_fts`, `document_edges`), relationship extraction rules, and comprehensive state lifecycle matrices for all 11 document types.
5. **[05. Composability, Provider Framework & Profiles](05_composability.md):** Defines the **5 Provider Interfaces** (`StorageProvider`, `IndexProvider`, `KnowledgeGraphProvider`, `SearchProvider`, `DocumentPersistenceService`) and **5 Deployment Profiles** (`Test`, `Local Full`, `Small`, `Medium`, `Large`), storage path resolution, and Serverless Content Navigator Web App architecture.
6. **[06. Documentation Strategy](../documentation-strategy.md):** Outlines documentation organization, SDLC workflows, user lifecycles, and template directories.


