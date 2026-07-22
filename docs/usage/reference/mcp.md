# Gyrus MCP Server Reference Manual

This manual provides setup configurations, tool definitions, resource URIs, and prompts for integrating the Gyrus embedded Model Context Protocol (MCP) server into GUI IDEs (Cursor, Claude Desktop, VS Code).

---

## 1. IDE Launch Configurations

### A. Cursor (`.cursor/mcp.json`)

Add the following block to `.cursor/mcp.json` in your workspace:

```json
{
  "mcpServers": {
    "gyrus": {
      "command": "gyrus",
      "args": ["mcp", "serve", "--storage-path", "/absolute/path/to/your/workspace"]
    }
  }
}
```

### B. Claude Desktop (`claude_desktop_config.json`)

Add the following block to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gyrus": {
      "command": "/usr/local/bin/gyrus",
      "args": ["mcp", "serve", "--storage-path", "/Users/yourname/projects/gyrus-docs"]
    }
  }
}
```

---

## 2. MCP Tools Reference

When connected, Gyrus exposes five native memory tools to the AI assistant:

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

---

## 3. MCP Resources & Prompts

### Resources
- **URI:** `memory://doc/{id}`  
  **MIME Type:** `text/markdown`  
  **Description:** Reads raw Markdown content of an OKF document.

### Prompts
- **Prompt Name:** `prepare-adr`  
  **Arguments:** `title` (string, required)  
  **Description:** Generates an Architecture Design Record (ADR) template.
