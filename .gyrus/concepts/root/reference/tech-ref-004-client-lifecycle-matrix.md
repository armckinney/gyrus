---
id: tech-ref-004-client-lifecycle-matrix
title: AI Agent Client Lifecycle Matrix (Install & Uninstall)
category: technical
type: technical-reference
format: ""
tenant: root
version: 2
status: active
last_modified_by: antigravity
last_updated: 2026-10-16T01:25:00Z
tags:
    - client
    - mcp
    - lifecycle
    - matrix
    - install
    - uninstall
    - marketplace
dependencies:
    - adr-008-native-client-plugin-marketplace-distribution
---

# Technical Reference: AI Agent Client Lifecycle Matrix (Install & Uninstall)

## 1. Overview & Scope

This technical reference defines the standard lifecycle operations—including plugin bundle discovery, Model Context Protocol (MCP) server configuration, environment variable injection, marketplace registration, and clean process teardown—for all AI agent clients supported by Gyrus:
1. **Google Antigravity**
2. **Claude Code (CLI & Desktop)**
3. **OpenAI Codex**
4. **GitHub Copilot (VS Code & CLI)**

---

## 2. Installation Process Matrix

| Client Platform | Native Marketplace / Installation Command | Repository Marketplace Manifest | Canonical Plugin Bundle | Injected Variables & Args |
| :--- | :--- | :--- | :--- | :--- |
| **OpenAI Codex** | `codex plugin marketplace add armckinney/gyrus`<br>`codex plugin add gyrus@armckinney` | `.agents/plugins/marketplace.json` | `packaging/plugins/gyrus/` | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_PROJECT="${workspaceFolder}"` |
| **Google Antigravity** | `agy plugin install https://github.com/armckinney/gyrus/tree/main/packaging/plugins/gyrus`<br>*(Local repo: `.agents/plugins.json`)* | `.agents/plugins.json` *(local repository)* | `packaging/plugins/gyrus/` | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_PROJECT="${workspaceFolder}"` |
| **GitHub Copilot** | `copilot plugin marketplace add armckinney/gyrus`<br>`copilot plugin install gyrus@armckinney` | `.github/plugin/marketplace.json` | `packaging/plugins/gyrus/` | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_PROJECT="${workspaceFolder}"` |
| **Claude Code** | `claude plugin marketplace add armckinney/gyrus`<br>`claude plugin install gyrus@armckinney` | `.claude-plugin/marketplace.json` | `packaging/plugins/gyrus/` | `args`: `["mcp", "serve"]`<br>`env`: `GYRUS_PROJECT="${workspaceFolder}"` |

---

## 3. Uninstallation Process Matrix

| Client Platform | Native Uninstallation Command | Configuration Cleanup Steps |
| :--- | :--- | :--- |
| **OpenAI Codex** | `codex plugin remove gyrus@armckinney` | Removes plugin entry from client registry. |
| **Google Antigravity** | `agy plugin uninstall gyrus` | Removes installed plugin from `~/.gemini/config/plugins/gyrus` and purges runtime cache. |
| **GitHub Copilot** | `copilot plugin uninstall gyrus@armckinney` | Removes plugin and associated MCP server entry. |
| **Claude Code** | `claude plugin uninstall gyrus@armckinney` | Removes plugin entry from Claude Code registry. |
| **All Platforms (Process)** | `pkill -f "gyrus mcp serve"` | Terminates lingering background stdio server processes. |

---

## 4. Verification Matrix

| Client Platform | Verification Method | Expected Success Output |
| :--- | :--- | :--- |
| **Google Antigravity** | Run: `agy plugin list`<br>Run: `agy mcp list` | **Install:** `gyrus` appears in plugin list with active status.<br>**Uninstall:** Neither `gyrus` plugin nor MCP server appears in output. |
| **Claude Code** | Run: `claude plugin list` | **Install:** `gyrus` appears in installed plugins.<br>**Uninstall:** List excludes `gyrus`. |
| **OpenAI Codex** | Run: `codex plugin list` | **Install:** `gyrus@armckinney` is listed as installed.<br>**Uninstall:** Excludes `gyrus@armckinney`. |
| **GitHub Copilot** | Run: `copilot plugin list` | **Install:** Server `gyrus` active and connected.<br>**Uninstall:** `gyrus` absent from plugin list. |
| **SQLite Index (Gyrus Engine)** | Run: `gyrus search --query ""` | Returns count of all synchronized living and immutable documents in the workspace. |

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
        "GYRUS_PROJECT": "${workspaceFolder}"
      }
    }
  }
}
```

### Agent Plugins Standard 1.0 Definition (`mcp.json`)
```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
  "mcpServers": {
    "gyrus": {
      "type": "stdio",
      "command": "gyrus",
      "args": [
        "mcp",
        "serve"
      ],
      "env": {
        "GYRUS_PROJECT": "${workspaceFolder}"
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
