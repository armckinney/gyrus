---
id: guide-003-agent-skills-setup
title: Gyrus Agent Plugins & Skills Setup Guide
category: technical
type: guide
format: markdown
owner_group: armckinney
version: 2
status: active
last_modified_by: antigravity
last_updated: 2026-09-13T23:58:30Z
tags:
  - plugins
  - agent-skills
  - agent-plugins-standard
  - mcp
  - setup
dependencies:
  - prd-001-specification-roadmap
  - adr-004-agent-plugin-packaging-distribution-and-init-revamp
---

# Gyrus Agent Plugins & Skills Setup Guide

This guide explains how to install and distribute the **Gyrus Agent Plugin** and skills so AI coding assistants (**Google Antigravity CLI**, **GitHub Copilot**, **OpenAI Codex**, and **Claude Desktop / Code**) can discover, read, and maintain your codebase memory.

---

## 1. Unified Agent Plugin Packaging

Rather than managing loose, fragmented skill scripts, Gyrus packages all agent assets into an integrated bundle compliant with the [Agent Plugins Standard 1.0.0](https://agent-plugins.org/) and compatible with Google Antigravity and Gemini CLI:

```text
docs/agents/plugins/gyrus/
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

## 2. Installing via `gyrus init client`

Install the plugin and register the stdio MCP server for your specific AI coding assistant using `gyrus init client`:

```bash
# For Google Antigravity CLI
gyrus init client --target antigravity

# For GitHub Copilot / VS Code
gyrus init client --target copilot

# For OpenAI Codex
gyrus init client --target codex

# For Claude Desktop / Claude Code
gyrus init client --target claude
```

### Supported Flags

| Flag | Shorthand | Description | Default |
| :--- | :--- | :--- | :--- |
| `--target` | `-t` | **Required.** Client target: `antigravity`, `claude`, `codex`, or `copilot`. | None |
| `--mode` | `-m` | MCP execution mode: `stdio` (local binary) or `docker` (containerized). | `stdio` |
| `--mcp-container-image` | | Docker container image tag for containerized mode. | `ghcr.io/armckinney/gyrus:latest` |
| `--global` | `-g` | Install into user home runtime directory rather than repository workspace. | `false` |
| `--plugin-dir` | | Custom target directory for extracting the plugin bundle. | Derived per target |

---

## 3. Tool Integration Matrix

| AI Tool / Harness | Target Identifier | Installed Artifacts | Discovery Mechanism |
| :--- | :--- | :--- | :--- |
| **Google Antigravity CLI (AGY)** | `antigravity` | `docs/agents/plugins/gyrus/` + `.antigravity/mcp.json` | Reads plugin manifest, rules (`rules/AGENTS.md`), skills, and MCP configuration. |
| **GitHub Copilot / VS Code** | `copilot` | `docs/agents/plugins/gyrus/` + `.vscode/mcp.json` | VS Code auto-detects `.vscode/mcp.json`; Copilot reads rules and skills in workspace. |
| **OpenAI Codex** | `codex` | `docs/agents/plugins/gyrus/` + `.codex/mcp.json` | Reads `.codex/mcp.json` and agent skills from plugin bundle. |
| **Claude Desktop / Code** | `claude` | `.claude/mcp.json` (or `~/.config/Claude/claude_desktop_config.json` with `-g`) | Claude connects to stdio MCP server; reads rules/skills if referenced. |

---

## 4. Document Mutability & Lifecycle Rules

AI Agents must adhere to these rules when interacting with Gyrus codebase memory:

1. 🌿 **Living Documents (`prd`, `specification`, `guide`, `standards`, `glossary`, `product`, `technical-reference`, `freeform`):** Represent the **current active state** of the system. Agents MUST actively update living specifications, guides, and standards whenever codebase implementation or architecture changes.
2. 📜 **Immutable Decision Logs (`adr`, `improvement-proposal`, `release-note`):** Represent **historical snapshots**. Once accepted or published (`status: accepted` / `status: active`), agents MUST NOT modify historical ADRs or proposals. When design choices change:
   - Create a NEW ADR or proposal (`gyrus create`).
   - Link the new document to supersede the old one (`gyrus link <new-id> <old-id> --rel-type supersedes`).
   - Update the old document status to `superseded` (`gyrus update <old-id> --status superseded`).
3. ⚙️ **Custom Immutable Templates (`immutable: true`):** Users can mark custom document types (e.g. `security-audit`, `compliance-report`, `incident-postmortem`) as immutable by adding `immutable: true` in the document frontmatter header. The Gyrus Core Engine will enforce content immutability once the document exits `draft`/`proposed` status.
