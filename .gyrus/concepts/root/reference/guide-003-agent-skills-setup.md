---
title: Gyrus Agent Plugin & Client Setup Guide
id: guide-003-agent-skills-setup
type: guide
format: markdown
category: technical
version: 7
status: active
last_updated: 2026-10-16T16:57:36Z
last_modified_by: antigravity
tenant: root
tags:
    - plugins
    - agent-skills
    - agent-plugins-standard
    - mcp
    - setup
    - marketplace
links:
    - concept: root/reference/adr-008-native-client-plugin-marketplace-distribution
      rel: depends_on
    - concept: root/reference/prd-001-specification-roadmap
      rel: depends_on
---

# Gyrus Agent Plugin & Client Setup Guide

This guide explains how to install and distribute the **Gyrus Agent Plugin** and Model Context Protocol (MCP) servers using native harness marketplace commands across **OpenAI Codex**, **Google Antigravity**, **GitHub Copilot**, and **Claude Code**.

---

## 1. Unified Agent Plugin Packaging

Rather than managing loose, fragmented skill scripts or manually generating configuration files in client directories, Gyrus packages all agent assets into a single canonical plugin bundle located at `packaging/plugins/gyrus/`:

```text
packaging/plugins/gyrus/
├── plugin.json               # Agent Plugins Standard 1.0.0 manifest (identity, capabilities)
├── mcp.json                  # Agent Plugins Standard stdio MCP configuration
├── mcp_config.json           # Google Antigravity MCP configuration
├── .mcp.json                 # Claude Code MCP configuration
├── .claude-plugin/
│   └── plugin.json           # Claude Code plugin manifest
├── .codex-plugin/
│   └── plugin.json           # OpenAI Codex plugin manifest
├── rules/
│   └── AGENTS.md             # Context Control Plane rules & instructions
└── skills/
    ├── gyrus/
    │   └── SKILL.md          # Unified Agent Skill (MCP tools primary, CLI subcommands fallback)
    └── gyrus-index/
        └── SKILL.md          # Repository onboarding and indexing skill
```

---

## 2. Native Plugin Installation via Remote Marketplaces

Install Gyrus directly into your AI coding assistant using native harness plugin commands:

### OpenAI Codex CLI

```bash
# Add the remote GitHub repository marketplace
codex plugin marketplace add armckinney/gyrus

# Install the gyrus plugin
codex plugin add gyrus@armckinney
```

### Google Antigravity (AGY)

```bash
# Install directly from the repository plugin directory
agy plugin install https://github.com/armckinney/gyrus/tree/main/packaging/plugins/gyrus
```

For local repository development, Antigravity discovers workspace plugins via `.agents/plugins.json`:
```json
{
  "entries": [
    {
      "path": "packaging/plugins"
    }
  ]
}
```

### GitHub Copilot CLI

```bash
# Add marketplace and install plugin
copilot plugin marketplace add armckinney/gyrus
copilot plugin install gyrus@armckinney
```

### Anthropic Claude Code

```bash
# Add marketplace and install plugin
claude plugin marketplace add armckinney/gyrus
claude plugin install gyrus@armckinney
```

---

## 3. Remote Marketplace Catalogs

Gyrus declares standard marketplace catalogs at repository root:
- **Claude Code**: `.claude-plugin/marketplace.json` (`"name": "armckinney"`)
- **OpenAI Codex**: `.agents/plugins/marketplace.json` (`"name": "armckinney"`)
- **GitHub Copilot**: `.github/plugin/marketplace.json` (`"name": "armckinney"`)

These manifests allow client harnesses to discover and install `gyrus@armckinney` without manual configuration editing.

---

## 4. MCP Stdio Server Integration

All plugin manifestations configure the native Gyrus stdio Model Context Protocol (MCP) server:

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

> [!NOTE]
> Ensure the `gyrus` executable is installed and available on your system `PATH` so your AI assistant's harness can launch `gyrus mcp serve` automatically upon session initialization.

## Qualified concepts and minimal links

Creation retains a local id and returns its qualified name. Address existing concepts with tenant/reference/id or tenant/projects/project/id across SDK, CLI, MCP, and UI. API envelopes include name; YAML frontmatter does not persist it. Relationships use only links with concept and rel. Dependency arrays, old link keys, per-link creation metadata, and short lookup IDs are rejected. Local IDs may repeat across tenants/projects. Incoming neighbors provide backlinks; relationship types and general cycle-safe graph traversal remain unchanged. Use the migration skill for legacy concepts or SQLite graph exports; retain backups and rebuild obsolete identity databases explicitly. Restart cached MCP processes after upgrading.
