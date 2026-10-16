---
title: AI Agent Client Lifecycle Matrix (Install & Uninstall)
id: tech-ref-004-client-lifecycle-matrix
type: technical-reference
format: markdown
category: technical
version: 6
status: active
last_updated: 2026-10-16T19:41:57Z
last_modified_by: antigravity
tenant: root
tags:
    - client
    - mcp
    - lifecycle
    - matrix
    - install
    - uninstall
    - marketplace
links:
    - concept: root/reference/adr-008-native-client-plugin-marketplace-distribution
      rel: depends_on
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
| **OKF collection (Gyrus Engine)** | Run: `gyrus search --query "" --scope all --max-results 0 --json` | Returns all matching concepts directly from authoritative storage. |

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

When unregistering or reinstalling Gyrus across agents, stop the relevant owned `gyrus mcp serve` instances so they release resources and reload provider configuration on restart. SQLite index/graph/search providers have been removed; OKF uses short-lived interprocess locks while accessing authoritative concepts.

* **List running processes:** `pgrep -fl "gyrus mcp serve"`
* **Graceful termination:** `pkill -TERM -f "gyrus mcp serve"`
* **Force termination:** `pkill -KILL -f "gyrus mcp serve"`

## Qualified concepts and minimal links

Creation retains a local id and returns its qualified name. Address existing concepts with tenant/reference/id or tenant/projects/project/id across SDK, CLI, MCP, and UI. API envelopes include name; YAML frontmatter does not persist it. Relationships use only links with concept and rel. Dependency arrays, old link keys, per-link creation metadata, and short lookup IDs are rejected. Local IDs may repeat across tenants/projects. Incoming neighbors provide backlinks; relationship types and general cycle-safe graph traversal remain unchanged. Use the migration skill for legacy concepts or SQLite graph exports; retain backups and rebuild obsolete identity databases explicitly. Restart cached MCP processes after upgrading.
