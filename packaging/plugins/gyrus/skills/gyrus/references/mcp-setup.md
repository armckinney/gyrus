# Gyrus MCP Server Setup Guide

Gyrus provides a Model Context Protocol (MCP) server running over `stdio` transport. It enables AI coding assistants (Google Antigravity, Claude, OpenAI Codex, GitHub Copilot / VS Code) to directly access codebase memory, search ADRs, and resolve task context.

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

## 🚀 Automatic MCP Registration (`gyrus init`)

Run `gyrus init` to non-destructively register Gyrus MCP servers across target agent tools:

```bash
# Containerized Stdio (Default)
gyrus init --mcp-mode container

# Local Binary Mode
gyrus init --mcp-mode local

# Global Registration (~/ user home)
gyrus init --global
```

---

## ⚙️ Target Configuration Matrix

### 1. Google Antigravity (`.antigravity/mcp.json`)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "${PWD}:/workspace", "-w", "/workspace", "ghcr.io/armckinney/gyrus:latest", "mcp", "serve"]
    }
  }
}
```

### 2. Claude Desktop & Code (`.claude/mcp.json` & `~/.config/Claude/claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "${PWD}:/workspace", "-w", "/workspace", "ghcr.io/armckinney/gyrus:latest", "mcp", "serve"]
    }
  }
}
```

### 3. OpenAI Codex (`.codex/mcp.json`)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "${PWD}:/workspace", "-w", "/workspace", "ghcr.io/armckinney/gyrus:latest", "mcp", "serve"]
    }
  }
}
```

### 4. GitHub Copilot / VS Code (`.vscode/mcp.json`)
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "${PWD}:/workspace", "-w", "/workspace", "ghcr.io/armckinney/gyrus:latest", "mcp", "serve"]
    }
  }
}
```
