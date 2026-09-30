---
id: tech-ref-004-client-lifecycle-matrix
title: AI Agent Client Lifecycle Matrix (Install & Uninstall)
category: technical
type: technical-reference
format: ""
owner_group: root
version: 1
status: active
last_modified_by: ""
last_updated: 2026-09-30T23:21:27Z
tags:
    - client
    - mcp
    - lifecycle
    - matrix
    - install
    - uninstall
---

# Technical Reference: AI Agent Client Lifecycle Matrix (Install & Uninstall)

## 1. Overview & Scope

This technical reference defines the standard lifecycle operations—including plugin bundle equipping, Model Context Protocol (MCP) server configuration, environment variable injection, unregistration, and clean process teardown—for all AI agent clients supported by Gyrus:
1. **Google Antigravity**
2. **Claude Code (CLI & Desktop)**
3. **OpenAI Codex**
4. **GitHub Copilot (VS Code & CLI)**

---

## 2. Installation Process Matrix

| Client Platform | Recommended Gyrus CLI Command | Native Tool CLI Command (Alternative / Fallback) | Target Configuration File Paths | Plugin / Rule Bundle Location | Injected Variables & Args |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Google Antigravity** | `gyrus client install --target antigravity` *(add `-g` for global)* | `agy plugin install .agents/plugins/gyrus`<br>*(or `agy mcp add gyrus -- gyrus mcp serve`)* | **Workspace:** `.antigravity/mcp.json`<br>**Global:** `~/.gemini/config/mcp_config.json` | `.agents/plugins/gyrus/`<br>*(Global: `~/.gemini/config/plugins/gyrus/`)* | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_WORKSPACE="${workspaceFolder}"` |
| **Claude Code** | `gyrus client install --target claude` *(add `-g` for global)* | `claude mcp add gyrus -- gyrus mcp serve` | **Workspace CLI:** `<workspace>/.mcp.json`<br>**Global CLI:** `~/.claude.json`<br>**Desktop:** `claude_desktop_config.json` *(macOS/Linux/Win)* | *None* *(Claude natively discovers MCP via `.mcp.json`)* | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_WORKSPACE="${workspaceFolder}"` |
| **OpenAI Codex** | `gyrus client install --target codex` *(add `-g` for global)* | Manual JSON injection under `mcpServers.gyrus` | **Workspace:** `.codex/mcp.json` or `.codex/config.json`<br>**Global:** `~/.codex/config.json` | `.agents/plugins/gyrus/`<br>*(Global: `~/.agents/plugins/gyrus/`)* | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_WORKSPACE="${workspaceFolder}"` |
| **GitHub Copilot (VS Code)** | `gyrus client install --target copilot` *(add `-g` for global)* | Manual JSON injection in `.vscode/mcp.json` | **Workspace:** `.vscode/mcp.json`<br>**Global User:**<br>• Linux: `~/.config/Code/User/settings.json`<br>• macOS: `~/Library/Application Support/Code/User/settings.json`<br>• Win: `%APPDATA%\Code\User\settings.json` | `.agents/plugins/gyrus/`<br>*(Global: `~/.agents/plugins/gyrus/`)* | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_WORKSPACE="${workspaceFolder}"`<br>Populates both `mcpServers` and `servers` *(type: `stdio`)* |

---

## 3. Uninstallation Process Matrix

| Client Platform | Recommended Gyrus CLI Command | Native Tool CLI Command | Configuration Cleanup Steps | Orphaned Artifacts & Cache Removal |
| :--- | :--- | :--- | :--- | :--- |
| **Google Antigravity** | `gyrus client uninstall --target antigravity` *(add `-g` for global)* | `agy plugin uninstall gyrus`<br>`agy mcp remove gyrus`<br>`agy mcp remove gyrus_gyrus` | • Remove `"gyrus"` from `.antigravity/mcp.json`<br>• Remove `"gyrus"` from `~/.gemini/config/mcp_config.json`<br>• Prune `mcp(gyrus_gyrus/*)` from `~/.gemini/config/config.json` | • Delete `.agents/plugins/gyrus/` (or `~/.gemini/config/plugins/gyrus/`)<br>• Purge runtime cache: `~/.gemini/antigravity/plugin_data/gyrus` |
| **Claude Code** | `gyrus client uninstall --target claude` *(add `-g` for global)* | `claude mcp remove gyrus` | • Delete `mcpServers.gyrus` from workspace `.mcp.json`<br>• Delete `mcpServers.gyrus` from `~/.claude.json`<br>• Remove `gyrus` from `claude_desktop_config.json` | *None* |
| **OpenAI Codex** | `gyrus client uninstall --target codex` *(add `-g` for global)* | Manual removal from JSON config | • Remove `mcpServers.gyrus` block from `.codex/mcp.json`<br>• Remove `mcpServers.gyrus` from `~/.codex/config.json` | • Delete `.agents/plugins/gyrus/` |
| **GitHub Copilot (VS Code)** | `gyrus client uninstall --target copilot` *(add `-g` for global)* | Manual removal from JSON config | • Remove `gyrus` from `.vscode/mcp.json` (`mcpServers` and `servers`)<br>• Remove `gyrus` from VS Code User `settings.json` | • Delete `.agents/plugins/gyrus/` |
| **All Platforms (Process)** | `pkill -f "gyrus mcp serve"` | `kill $(pgrep -f "gyrus mcp serve")` | Ensure stdio pipes are gracefully closed | Terminate background orphan processes |

---

## 4. Verification Matrix

| Client Platform | Verification Method | Expected Success Output |
| :--- | :--- | :--- |
| **Google Antigravity** | Run: `agy plugin list`<br>Run: `agy mcp list` | **Install:** `gyrus` appears in plugin list; `gyrus` or `gyrus_gyrus` listed with active status.<br>**Uninstall:** Neither `gyrus` plugin nor MCP server appears in output. |
| **Claude Code** | Run: `claude mcp list` | **Install:** `gyrus` appears in configured MCP servers.<br>**Uninstall:** `No MCP servers configured` (or list excludes `gyrus`). |
| **OpenAI Codex** | Inspect `.codex/mcp.json` or `~/.codex/config.json` | **Install:** Valid JSON with `mcpServers.gyrus.command == "gyrus"`.<br>**Uninstall:** File either deleted or valid JSON without `gyrus` key. |
| **GitHub Copilot (VS Code)** | Command Palette: `MCP: List Servers`<br>*(or inspect `.vscode/mcp.json`)* | **Install:** Server `gyrus` active and connected.<br>**Uninstall:** `gyrus` absent from MCP server list and `.vscode/mcp.json`. |
| **SQLite Index (Gyrus Engine)** | Run: `gyrus search --query ""` | **Install:** Returns count of all synchronized living and immutable documents in the workspace. |

---

## 5. Configuration & Schema Specifications

### Standard MCP Stdio Definition
```json
{
  "mcpServers": {
    "gyrus": {
      "command": "gyrus",
      "args": ["mcp", "serve"],
      "env": {
        "GYRUS_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

### VS Code Copilot Stdio Definition
```json
{
  "servers": {
    "gyrus": {
      "type": "stdio",
      "command": "gyrus",
      "args": ["mcp", "serve"],
      "env": {
        "GYRUS_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

---

## 6. Process & Daemon Management Protocol

When unregistering or reinstalling Gyrus across agents, lingering background instances of `gyrus mcp serve` can maintain locks on the SQLite index store (`.gyrus/index.db`) or graph edge tables.

* **List running processes:** `pgrep -fl "gyrus mcp serve"`
* **Graceful termination:** `pkill -TERM -f "gyrus mcp serve"`
* **Force termination:** `pkill -KILL -f "gyrus mcp serve"`
