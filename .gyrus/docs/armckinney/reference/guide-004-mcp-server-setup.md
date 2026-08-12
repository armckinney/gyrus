---
id: guide-004-mcp-server-setup
title: Gyrus MCP Server Setup Guide
category: technical
type: guide
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus MCP Server Setup Guide

The Model Context Protocol (MCP) server configuration format is **identical** across **Antigravity CLI**, **GitHub Copilot**, **Claude Desktop**, and **OpenAI Codex**. You only need to add the standard `mcpServers` block to your tool's respective configuration file.

---

## 1. Tool Configuration File Locations

Add the configuration snippet below to the appropriate path for your tool:

| AI Tool / Client | Target Configuration File Path |
| :--- | :--- |
| **Google Antigravity CLI (AGY)** | `~/.gemini/antigravity-cli/mcp.json` or `<workspace>/mcp.json` |
| **GitHub Copilot / VS Code** | `<workspace>/.vscode/mcp.json` |
| **Claude Desktop** | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`<br>Windows: `%APPDATA%\Claude\claude_desktop_config.json` |
| **OpenAI Codex / Custom MCP Clients** | `<workspace>/.codex/mcp.json` or client configuration |

---

## 2. Configuration Snippets

### Option A: Containerized Docker GHCR (Recommended - Zero Installation)

No binary installation required. Docker automatically pulls `ghcr.io/armckinney/gyrus:latest` and starts the stdio MCP server:

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

> *Note for Claude Desktop:* Replace `${workspaceFolder}` with the absolute path to your repository (e.g. `/Users/yourname/projects/my-repo`).

---

### Option B: Local Native Binary

If you have installed the `gyrus` binary locally via `curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash`:

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
