---
id: guide-004-mcp-server-setup
title: Gyrus MCP Server Setup Guide
category: technical
type: guide
format: ""
owner_group: root
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

## 2. Configuration Snippet (Native Binary Standard)

Install the pre-compiled `gyrus` binary locally via:
```bash
curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash
```

### Automated Setup via Gyrus CLI
You can install and register MCP servers automatically with a single command:
```bash
gyrus client install --target antigravity   # or claude, codex, copilot
```
To cleanly remove the plugin and unregister MCP configurations:
```bash
gyrus client uninstall --target antigravity
```

---

### Manual Configuration Snippet
Add the following stdio server entry to your client configuration:

```json
{
  "mcpServers": {
    "gyrus": {
      "command": "gyrus",
      "args": [
        "mcp",
        "serve"
      ],
      "env": {
        "GYRUS_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

> *Note for Claude Desktop:* Replace `${workspaceFolder}` with the absolute path to your repository (e.g. `/Users/yourname/projects/my-repo`) or set `GYRUS_WORKSPACE` in your environment.

---

## 3. Server Flags & Workspace Overrides

You can also pass explicit workspace and configuration flags directly when launching `mcp serve`:

```bash
# Explicit workspace directory
gyrus mcp serve --workspace /path/to/project

# Explicit configuration file
gyrus mcp serve --config /path/to/.gyrus.yaml
```

