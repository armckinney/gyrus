---
id: guide-003-agent-skills-setup
title: Gyrus Agent Plugin & Client Setup Guide
category: technical
type: guide
format: markdown
owner_group: root
version: 3
status: active
last_modified_by: antigravity
last_updated: 2026-10-02T05:51:00Z
tags:
  - plugins
  - agent-skills
  - agent-plugins-standard
  - mcp
  - setup
  - client
dependencies:
  - prd-001-specification-roadmap
  - adr-004-agent-plugin-packaging-distribution-and-init-revamp
---

# Gyrus Agent Plugin & Client Setup Guide

This guide explains how to install and distribute the **Gyrus Agent Plugin** and Model Context Protocol (MCP) servers so AI coding assistants (**Google Antigravity CLI**, **GitHub Copilot**, **OpenAI Codex**, and **Claude Desktop / Code**) can discover, read, and maintain your codebase memory.

---

## 1. Unified Agent Plugin Packaging

Rather than managing loose, fragmented skill scripts, Gyrus packages all agent assets into an integrated bundle compliant with the [Agent Plugins Standard 1.0.0](https://agent-plugins.org/) and compatible with Google Antigravity, Gemini CLI, VS Code, and Claude:

```text
.agents/plugins/gyrus/
├── plugin.json          # Agent Plugins Standard 1.0.0 manifest (identity, capabilities)
├── mcp.json             # Agent Plugins Standard stdio MCP configuration
├── mcp_config.json      # Google Antigravity / Gemini CLI MCP configuration
├── rules/
│   └── AGENTS.md        # Context Control Plane rules & instructions
└── skills/
    └── gyrus/
        ├── SKILL.md     # Unified Agent Skill (MCP tools primary, CLI subcommands fallback)
        └── references/
            ├── okf-schemas.md
            ├── mcp-setup.md
            └── storage-providers.md
```

---

## 2. Automated Installation via `gyrus client install`

Install the plugin and register the stdio MCP server for your specific AI coding assistant using `gyrus client install`:

```bash
# For Google Antigravity CLI
gyrus client install --target antigravity

# For GitHub Copilot / VS Code
gyrus client install --target copilot

# For OpenAI Codex
gyrus client install --target codex

# For Claude Desktop / Claude Code
gyrus client install --target claude
```

### Supported Flags

| Flag | Shorthand | Description | Default |
| :--- | :--- | :--- | :--- |
| `--target` | `-t` | **Required.** Client target: `antigravity`, `claude`, `codex`, or `copilot`. | None |
| `--mode` | `-m` | MCP execution mode: `local` (local binary) or `container` (containerized Docker). | `local` |
| `--mcp-container-image` | | Docker container image tag for containerized mode. | `ghcr.io/armckinney/gyrus:latest` |
| `--global` | `-g` | Install into user home runtime directory rather than repository workspace. | `false` |
| `--plugin-dir` | | Custom target directory for extracting the plugin bundle. | Derived per target |

---

## 3. Uninstallation via `gyrus client uninstall`

To cleanly remove the plugin bundle and unregister client MCP configurations:

```bash
# Clean uninstall for a specific client
gyrus client uninstall --target antigravity

# Globally unregister from user home directory
gyrus client uninstall --target antigravity --global
```

---

## 4. Tool Integration Matrix & File Locations

| AI Tool / Harness | Target Identifier | Installed Artifacts | Target Configuration File Path | Discovery Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Google Antigravity CLI (AGY)** | `antigravity` | `.agents/plugins/gyrus/` *(Global: `~/.gemini/config/plugins/gyrus/`)* | Discovered directly from plugin `mcp_config.json` | Reads plugin manifest, rules (`rules/AGENTS.md`), skills, and starts plugin MCP server. |
| **GitHub Copilot / VS Code** | `copilot` | `.agents/plugins/gyrus/` + `.vscode/mcp.json` | `<workspace>/.vscode/mcp.json`<br>*(Global: `Code/User/settings.json`)* | VS Code auto-detects `.vscode/mcp.json`; Copilot reads rules and skills in workspace. |
| **OpenAI Codex** | `codex` | `.agents/plugins/gyrus/` + `.codex/mcp.json` | `<workspace>/.codex/mcp.json`<br>*(Global: `~/.codex/config.json`)* | Reads `.codex/mcp.json` and agent skills from plugin bundle. |
| **Claude Desktop / Code** | `claude` | `<workspace>/.mcp.json` | `<workspace>/.mcp.json` (or `~/.claude.json`)<br>Desktop: `claude_desktop_config.json` | Claude connects to stdio MCP server; reads rules/skills if referenced. |

> [!NOTE]
> For **Google Antigravity**, MCP is loaded natively from `.agents/plugins/gyrus/mcp_config.json` when the plugin is active. A separate `.antigravity/mcp.json` is not required and should not be created, preventing duplicate server process launches.

---

## 5. Manual Configuration Snippets (Standalone MCP)

If your environment cannot run `gyrus client install`, you can manually add the stdio server entry to your client configuration:

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

> *Note for Claude Desktop:* Replace `${workspaceFolder}` with the absolute path to your repository root (e.g. `/Users/yourname/projects/my-repo`) or export `GYRUS_WORKSPACE` in your shell environment.

### Server Flags & Workspace Overrides

You can pass explicit workspace or configuration file flags directly when running `gyrus mcp serve`:

```bash
# Explicit workspace directory
gyrus mcp serve --workspace /path/to/project

# Explicit configuration file path
gyrus mcp serve --config /path/to/.gyrus.yaml
```

---

## 6. Document Mutability & Lifecycle Rules

AI Agents must adhere to these rules when interacting with Gyrus codebase memory:

1. 🌿 **Living Documents (`prd`, `specification`, `guide`, `standards`, `glossary`, `product`, `technical-reference`, `freeform`):** Represent the **current active state** of the system. Agents MUST actively update living specifications, guides, and standards whenever codebase implementation or architecture changes.
2. 📜 **Immutable Decision Logs (`adr`, `improvement-proposal`, `release-note`):** Represent **historical snapshots**. Once accepted or published (`status: accepted` / `status: active`), agents MUST NOT modify historical ADRs or proposals. When design choices change:
   - Create a NEW ADR or proposal (`gyrus create`).
   - Link the new document to supersede the old one (`gyrus link <new-id> <old-id> --rel-type supersedes`).
   - Update the old document status to `superseded` (`gyrus update <old-id> --status superseded`).
3. ⚙️ **Custom Immutable Templates (`immutable: true`):** Users can mark custom document types (e.g. `security-audit`, `compliance-report`, `incident-postmortem`) as immutable by adding `immutable: true` in the document frontmatter header. The Gyrus Core Engine will enforce content immutability once the document exits `draft`/`proposed` status.
