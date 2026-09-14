---
name: gyrus-mcp
description: Gyrus Unified Context & Memory Engine MCP skill. Use native Model Context Protocol (MCP) tools and resources to search, retrieve, create, update, link, and suggest relevant OKF codebase context for tasks.
applyTo:
  - "**"
---

# Gyrus MCP Agent Skill Specification

This skill equips MCP-enabled AI agents (Cursor, Claude Desktop, OpenAI Codex, GitHub Copilot, Windsurf) to interact directly with Gyrus codebase memory via native Model Context Protocol (MCP) tool calls and resource endpoints.

---

## 💡 Core Agent Guidelines

1. **Before Modifying Code:** Always invoke `gyrus_suggest_context({ prompt: "<task description>" })` or `gyrus_search({ query: "<keyword>" })` to read relevant ADRs and technical contracts.
2. **ID Naming Rule:** Document IDs MUST match the lower-case pattern `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
3. **Linkage & Dependency Graph:** When creating or mutating contracts, link dependent documents using `gyrus_link_documents({ from_id: "...", to_id: "...", rel_type: "depends_on" })`.

---

## 🛠️ Complete MCP Tool Reference

### 1. `gyrus_suggest_context`
Linearizes top relevant documents matching a task prompt (Recommended first step before writing code).
- **Arguments:** `{ "prompt": string, "limit"?: int }`

### 2. `gyrus_search`
Executes FTS5 lexical keyword search across codebase context documents.
- **Arguments:** `{ "query": string, "limit"?: int }`

### 3. `gyrus_get_document`
Retrieves a single document payload by ID.
- **Arguments:** `{ "id": string }`

### 4. `gyrus_create_document`
Creates a new OKF contract document in context storage.
- **Arguments:** `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }`

### 5. `gyrus_update_document`
Patches metadata fields or content of an existing contract document.
- **Arguments:** `{ "id": string, "title"?: string, "status"?: string, "content"?: string }`

### 6. `gyrus_link_documents`
Creates a directed relationship edge between two documents (`depends_on`, `supersedes`, `implements`).
- **Arguments:** `{ "from_id": string, "to_id": string, "rel_type": string }`

### 7. `gyrus_sync`
Re-indexes filesystem documents and extracts dependency links into the FTS database.
- **Arguments:** `{}`

---

## 🔌 MCP Resource Endpoints

- `gyrus://documents/{id}`: Direct document payload lookup.
- `gyrus://schema/{type}`: OKF document template schema definition.
- `gyrus://graph/topology`: Complete dependency graph topology.

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- 🔌 **[MCP Setup Guide](references/mcp-setup.md)**
- 🗄️ **[Storage Providers](references/storage-providers.md)**
