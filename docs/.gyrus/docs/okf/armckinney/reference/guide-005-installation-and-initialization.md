---
id: guide-005-installation-and-initialization
title: Gyrus Installation & Workspace Initialization Guide
category: technical
type: guide
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-26T21:28:03Z
tags:
    - installation
    - initialization
    - setup
    - mcp
---

---
id: guide-005-installation-and-initialization
title: Gyrus Installation & Workspace Initialization Guide
category: technical
type: guide
format: markdown
owner_group: armckinney
version: 1
status: active
tags:
  - installation
  - initialization
  - setup
  - mcp
  - agent-skills
dependencies:
  - prd-001-specification-roadmap
  - adr-001-context-storage-architecture
---

# Gyrus Installation & Workspace Initialization Guide

This guide covers installing the Gyrus CLI binary on your machine and initializing your workspace with custom configurations, AI agent skills, and Model Context Protocol (MCP) server registrations.

---

## 📥 1. Installing the Gyrus CLI

Choose one of the following methods to install the `gyrus` executable into your system `$PATH`:

### Option A: One-Liner Installer Script (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash
```
This automatically detects your OS (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`), downloads the latest binary release from GitHub, and places `gyrus` into `/usr/local/bin` (or `~/.local/bin`).

### Option B: Build from Source via `go install`
```bash
go install github.com/armckinney/gyrus/cmd/gyrus@latest
```

### Option C: Manual Binary Release Download
Download the pre-compiled binary tarball for your platform directly from [Gyrus GitHub Releases](https://github.com/armckinney/gyrus/releases/latest).

---

## 🚀 2. Workspace Initialization (`gyrus init`)

The `gyrus init` command is the master entrypoint for initializing a workspace. By default, running `gyrus init` without flags performs a complete setup:

```bash
gyrus init
```

### What `gyrus init` does automatically:
1. **Creates Storage Roots:** Initializes `docs/.gyrus/docs/` and SQLite indexer databases (`index.db`).
2. **Generates Configuration:** Writes a `.gyrus.yaml` file in the workspace root.
3. **Equips Agent Skills:** Installs Gyrus Agent Skill files (`SKILL.md`, `scripts/verify.sh`, `references/`) into `.agents/skills/gyrus/`.
4. **Registers MCP Servers:** Auto-detects and registers the stdio MCP server (`gyrus mcp serve`) in configuration files for Cursor/Antigravity, Claude Desktop, OpenAI Codex, and GitHub Copilot.

---

## 🎛️ 3. Advanced Customization & Flag Options

### 3.1 Custom-Tailored Agent Tool Selection
If you only use specific AI agent tools (e.g. Claude Desktop or Cursor), you can restrict MCP registration and skill equipping to your preferred platform:

```bash
# Register MCP and equip skills ONLY for Claude Desktop / Claude Code
gyrus init --mcp-target claude --skill-target claude

# Target Cursor / Antigravity only
gyrus init --mcp-target antigravity

# Target GitHub Copilot / VS Code only
gyrus init --mcp-target copilot

# Target OpenAI Codex only
gyrus init --mcp-target codex
```

### 3.2 CLI-Only & Headless Environments
For CI/CD pipelines, Docker containers, or headless server instances where you do not need MCP server registration or prompt skills:

```bash
# Skip MCP server registration
gyrus init --no-mcp

# Skip Agent Skill equipping
gyrus init --no-skill

# Complete headless CLI-only initialization
gyrus init --no-mcp --no-skill
```

### 3.3 Storage Profile Matrix Selection (`--profile`)
Specify a pre-configured storage & search profile using the `-p` or `--profile` flag:

```bash
# LocalFS Storage + SQLite FTS5 (Default)
gyrus init --profile local

# Remote Git Repository Persistence
gyrus init --profile git

# Cloud Object Storage (AWS S3, Azure Blob, Google Cloud Storage)
gyrus init --profile blob

# PostgreSQL Enterprise Database Backend
gyrus init --profile postgres

```
