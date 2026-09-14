---
name: gyrus
description: Gyrus Unified Context Control Plane & Memory Engine agent skill. Discovers, retrieves, creates, and maintains OKF codebase context (ADRs, PRDs, specs) using native MCP tools (primary) or CLI subcommands (fallback).
applyTo:
  - "**"
---

# Gyrus Agent Skill: Context Control Plane & Memory Engine

This skill equips AI coding assistants (Google Antigravity, GitHub Copilot, OpenAI Codex, Claude) to discover, synthesize, create, and maintain codebase context through the **Gyrus Context Control Plane**.

---

## 💡 Core Agent Guidelines

1. **Context First Before Editing Code**:
   - Always resolve relevant architectural and design context before modifying codebase implementation.
   - **Primary Interface (Native MCP):** Call `gyrus_suggest_context({ prompt: "<task description>" })` or `gyrus_search({ query: "<keywords>" })`.
   - **Fallback Interface (CLI):** If MCP tools are unavailable or in shell-only environments, run `gyrus suggest-context --prompt "<task description>"` or `gyrus search --query "<keywords>" --json`.
2. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (`prd`, `specification`, `guide`, `standards`, `glossary`, `product`, `technical-reference`): Capture the active system state. Agents MUST update living specs and standards when implementation or design evolves.
   - 📜 **Immutable Decision Logs** (`adr`, `improvement-proposal`, `release-note`): Historical snapshots. Once accepted or published (`status: accepted` / `status: active`), they are strictly immutable (`immutable: true`). When architectural decisions change:
     1. Propose a NEW ADR (`gyrus_create_document` or `gyrus create`).
     2. Link the new ADR to supersede the old one (`gyrus_link_documents({ from_id: new_id, to_id: old_id, rel_type: "supersedes" })` or `gyrus link`).
     3. Update the old ADR status to `superseded` (`gyrus_update_document` or `gyrus update`).
3. **OKF Contract Schema Compliance**:
   - Document IDs MUST match lower-case alphanumeric pattern `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
   - Every contract requires YAML frontmatter defining `id`, `title`, `category`, `type`, `owner_group`, `version`, `status`.

---

## 🛠️ Primary Interface: Model Context Protocol (MCP) Tools

Use native MCP tools whenever available:

### 1. `gyrus_suggest_context`
Synthesizes linearized top-ranking context matching a task prompt within token budgets.
- **Parameters:** `{ "prompt": string, "limit"?: int }`

### 2. `gyrus_search`
Executes FTS5 lexical keyword search across codebase context documents.
- **Parameters:** `{ "query": string, "limit"?: int }`

### 3. `gyrus_get_document`
Retrieves a single OKF document payload by ID.
- **Parameters:** `{ "id": string }`

### 4. `gyrus_create_document`
Creates a new OKF contract document in context storage.
- **Parameters:** `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }`

### 5. `gyrus_update_document`
Patches metadata fields or content of an existing document.
- **Parameters:** `{ "id": string, "title"?: string, "status"?: string, "content"?: string }`

### 6. `gyrus_link_documents`
Creates a directed relationship edge between two documents (`depends_on`, `supersedes`, `implements`, `mitigates`).
- **Parameters:** `{ "from_id": string, "to_id": string, "rel_type": string }`

### 7. `gyrus_sync`
Re-indexes storage documents into the SQLite FTS5 database and updates dependency edges.
- **Parameters:** `{}`

### 🔌 MCP Resource Endpoints
- `gyrus://documents/{id}`: Document payload lookup.
- `gyrus://schema/{type}`: OKF document template schema definition.
- `gyrus://graph/topology`: Complete dependency graph topology.

---

## 💻 Fallback & Admin Interface: Gyrus CLI

When running in shell-only environments, CI pipelines, or if MCP is unavailable, use the `gyrus` CLI executable:

| Command | Purpose | Example |
| :--- | :--- | :--- |
| `gyrus suggest-context` | Synthesize context for task | `gyrus suggest-context --prompt "<task>" --json` |
| `gyrus search` | FTS5 keyword search | `gyrus search --query "<term>" --json` |
| `gyrus get` | Retrieve document by ID | `gyrus get <id> --json` |
| `gyrus create` | Create new OKF document | `gyrus create --id "<id>" --title "<title>" --type "<type>" --category "<cat>" --owner-group "<grp>" --content "<md>"` |
| `gyrus update` | Update document fields/body | `gyrus update <id> --status "<status>" --expected-version <v>` |
| `gyrus link` | Create relationship edge | `gyrus link <from-id> <to-id> --rel-type supersedes` |
| `gyrus sync` | Re-index context database | `gyrus sync` |
| `gyrus validate` | Validate OKF schema | `gyrus validate <file.md>` |
| `gyrus schema` | Print template schema | `gyrus schema <doc-type>` |

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md):** Complete frontmatter schema & state machine specifications.
- 🔌 **[MCP Setup Guide](references/mcp-setup.md):** Registration guide for Cursor, Claude Desktop, and VS Code / Copilot.
- 🗄️ **[Storage Providers](references/storage-providers.md):** Configuration guide for LocalFS, Git, Cloud Blob, and PostgreSQL.

