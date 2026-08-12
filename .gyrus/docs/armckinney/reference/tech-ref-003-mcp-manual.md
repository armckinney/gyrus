---
id: tech-ref-003-mcp-manual
title: Gyrus Model Context Protocol Reference Manual
category: technical
type: technical-reference
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus MCP Server Reference Manual

This manual provides the consolidated launch configurations, tool definitions, resource URIs, and prompts for the Gyrus Model Context Protocol (MCP) server.

---

## 1. Tool Configuration File Locations

Add the standard `mcpServers.gyrus` configuration block to your tool's respective path:

| AI Tool / Client | Configuration File Path |
| :--- | :--- |
| **Google Antigravity CLI (AGY)** | `~/.gemini/antigravity-cli/mcp.json` or `<workspace>/mcp.json` |
| **GitHub Copilot / VS Code** | `<workspace>/.vscode/mcp.json` |
| **Claude Desktop** | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows: `%APPDATA%\Claude\claude_desktop_config.json` |
| **OpenAI Codex / Custom MCP Clients** | `<workspace>/.codex/mcp.json` |

---

## 2. Standard MCP Launch Configurations

### Containerized Docker GHCR (Zero-Install)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v", "${workspaceFolder}:/workspace",
        "ghcr.io/armckinney/gyrus:latest",
        "mcp", "serve", "--storage-path", "/workspace"
      ]
    }
  }
}
```

### Native Binary
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "gyrus",
      "args": ["mcp", "serve"]
    }
  }
}
```

---

## 3. MCP Tools Reference

When connected, Gyrus exposes five native memory tools to AI assistants:

### 1. `memory_create`
Creates a new OKF contract document.

* **Parameters:**
  * `id` (string, required): Document ID (matches regex `^[a-z0-9-_]+$`).
  * `title` (string, required): Document title.
  * `category` (string, required): Category (`architecture`, `business-logic`, `product`, `operations`, `technical`).
  * `type` (string, required): Document type (`adr`, `prd`, `guide`, `specification`, etc.).
  * `owner_group` (string, required): Owner group identifier.
  * `status` (string, optional): Status (`draft`, `proposed`, `active`).
  * `content` (string, optional): Markdown body content.

### 2. `memory_get`
Retrieves an OKF document by ID.

* **Parameters:**
  * `id` (string, required): Target document ID.

### 3. `memory_search`
Executes an FTS5 search query across all OKF documents.

* **Parameters:**
  * `query` (string, optional): Keyword query text.
  * `category` (string, optional): Filter by category.
  * `type` (string, optional): Filter by type.

### 4. `memory_suggest_context`
Linearizes top context documents matching an agent prompt description.

* **Parameters:**
  * `prompt` (string, required): Task description or context query.

### 5. `memory_link`
Creates a directed relationship edge between two documents.

* **Parameters:**
  * `from_id` (string, required): Source document ID.
  * `to_id` (string, required): Target document ID.
  * `rel_type` (string, optional): Relationship type (`depends_on`, `supersedes`, `implements`).

### 6. `memory_archive`
Archives (deletes) a document from storage and search index.

* **Parameters:**
  * `id` (string, required): Target document ID to archive.


---

## 4. MCP Resources & Prompts

### Resources
- **URI:** `memory://doc/{id}`  
  **MIME Type:** `text/markdown`  
  **Description:** Reads raw Markdown content of an OKF document.

### Prompts
- **Prompt Name:** `prepare-adr`  
  **Arguments:** `title` (string, required)  
  **Description:** Generates an Architecture Design Record (ADR) template.
