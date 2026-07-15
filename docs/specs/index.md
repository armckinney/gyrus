# Gyrus: Unified Context & Memory Engine
## Architecture Specification Index

Gyrus is a modular, local-first context and memory management system designed to serve as the single source of truth for both autonomous AI agents (such as Claude Code, GitHub Copilot, and Antigravity) and human developers.

---

## 1. Core Architecture Blueprint

Gyrus uses a local-first repository storage model with a SQLite-backed metadata index and edge-table relationship graph.

```
+---------------------------------------------------------------------------------+
|                                 CLIENT TIER                                     |
|  [ Claude Code / CLI Agents ]   [ Cursor / GUI Agents ]   [ Human Developers ]  |
|       (Shell Execution)           (Local MCP Server)       (Obsidian/VS Code)   |
+──────────────────────────┬─────────────────┬──────────────────────┬─────────────+
                           │                 │                      │
                           ▼                 ▼                      ▼
+────────────────────────────────────────────────────────────────────────────────-+
|                          SHARED CORE ENGINE (Go Binary)                         |
|  * Enforces JSON metadata contracts & schemas                                  |
|  * Enforces lifecycle transition rules (e.g. proposed -> accepted)             |
|  * Manages SQL Edge-Table relationship graphs                                   |
+────────────────────────────────────────┬────────────────────────────────────────+
                                         │ (Repository Pattern Interface)
                                         ▼
+────────────────────────────────────────────────────────────────────────────────-+
|                            STORAGE ADAPTERS (Swappable)                         |
|     [ Local File System + SQLite ]          [ Azure Cosmos DB + AI Search ]     |
|       (Zero-dependency Bootstrapping)             (Scale-out Enterprise Production)  |
+---------------------------------------------------------------------------------+
```

---

## 2. Modular Specifications

1. **[01. JSON Contract Schema](file:///Users/armck/git/wiki/docs/specs/01_contract_schema.md):** Defines metadata envelope rules, tags, and standard document types.
2. **[02. CLI Interface & Auth Flow](file:///Users/armck/git/wiki/docs/specs/02_cli_interface.md):** Outlines the `gyrus` command-line subcommands, caching, and token validation.
3. **[03. MCP Interface Specification](file:///Users/armck/git/wiki/docs/specs/03_mcp_interface.md):** Formulates tools, resources, and prompt templates exposed natively to AI editors.
4. **[04. Core Engine & Validation](file:///Users/armck/git/wiki/docs/specs/04_core_engine.md):** Details Go interfaces, SQLite edge relationship mapping, and state lifecycle validation rules.
5. **[05. Composability & Portability Shims](file:///Users/armck/git/wiki/docs/specs/05_composability.md):** Abstraction repository patterns and cross-agent compatibility guidelines.
6. **[06. Documentation Strategy](file:///Users/armck/git/wiki/docs/documentation-strategy.md):** Outlines documentation organization, SDLC workflows, user lifecycles, and template directories.
