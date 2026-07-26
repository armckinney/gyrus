# Gyrus MCP Server Setup Guide

Gyrus provides a Model Context Protocol (MCP) server running over `stdio` transport. It enables AI coding assistants (Cursor, Claude Desktop, GitHub Copilot, VS Code) to directly access codebase memory, search ADRs, and resolve task context.

## 🛠️ MCP Tool Definitions

Gyrus exposes 7 native MCP tools:
1. `gyrus_suggest_context`: Linearizes top relevant documents within token budget (`prompt`, `max_tokens`).
2. `gyrus_search`: FTS5 keyword search across documents (`query`, `category`, `type`, `status`).
3. `gyrus_get`: Retrieves a document by ID (`id`).
4. `gyrus_create`: Creates a new OKF document (`id`, `title`, `category`, `type`, `owner_group`, `content`).
5. `gyrus_update`: Updates document metadata or content (`id`, `title`, `status`, `content`).
6. `gyrus_link`: Creates a directional link edge (`from_id`, `to_id`, `rel_type`).
7. `gyrus_sync`: Re-indexes filesystem documents and updates graph edges.

---

## ⚙️ IDE Configuration Files

### 1. Cursor (`.cursor/mcp.json`)
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

### 2. Claude Desktop (`claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "/usr/local/bin/gyrus",
      "args": ["mcp", "serve"]
    }
  }
}
```

### 3. VS Code / Copilot (`.vscode/mcp.json`)
```json
{
  "inputs": [],
  "servers": {
    "gyrus": {
      "type": "stdio",
      "command": "gyrus",
      "args": ["mcp", "serve"]
    }
  }
}
```
