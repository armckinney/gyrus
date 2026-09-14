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
   - 🌿 **Living Documents** (`prd`, `specification`, `guide`, `standards`, `glossary`, `product`, `technical-reference`): Capture active system state. Agents MUST update living specs and standards when implementation or design evolves.
   - 📜 **Immutable Decision Logs** (`adr`, `improvement-proposal`, `release-note`): Historical snapshots. Once accepted or published (`status: accepted` / `status: active`), they are strictly immutable (`immutable: true`). When architectural decisions change:
     1. Propose a NEW ADR (`gyrus_create_document` or `gyrus create`).
     2. Link the new ADR to supersede the old one (`gyrus_link_documents` or `gyrus link`).
     3. Update the old ADR status to `superseded` (`gyrus_update_document` or `gyrus update`).
3. **OKF Contract Schema Compliance**:
   - Document IDs MUST match lower-case pattern `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
   - Every contract requires YAML frontmatter defining `id`, `title`, `category`, `type`, `owner_group`, `version`, `status`.

---

## 🛠️ Primary Interface: Model Context Protocol (MCP) Tools

Use native MCP tools whenever available:
- **`gyrus_suggest_context`**: `{ "prompt": string, "limit"?: int }`
- **`gyrus_search`**: `{ "query": string, "limit"?: int }`
- **`gyrus_get_document`**: `{ "id": string }`
- **`gyrus_create_document`**: `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }`
- **`gyrus_update_document`**: `{ "id": string, "title"?: string, "status"?: string, "content"?: string }`
- **`gyrus_link_documents`**: `{ "from_id": string, "to_id": string, "rel_type": string }`
- **`gyrus_sync`**: `{}`

---

## 💻 Fallback & Admin Interface: Gyrus CLI

When running in shell-only environments, CI pipelines, or if MCP is unavailable, use the `gyrus` CLI:
- `gyrus suggest-context --prompt "<task>" --json`
- `gyrus search --query "<query>" --json`
- `gyrus get <id> --json`
- `gyrus create --id "<id>" --title "<title>" --category "<cat>" --type "<type>" --owner-group "<grp>" --content "<md>"`
- `gyrus update <id> --status "<status>" --expected-version <v>`
- `gyrus link <from-id> <to-id> --rel-type <type>`
- `gyrus sync`

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- 🔌 **[MCP Setup Guide](references/mcp-setup.md)**
- 🗄️ **[Storage Providers](references/storage-providers.md)**
