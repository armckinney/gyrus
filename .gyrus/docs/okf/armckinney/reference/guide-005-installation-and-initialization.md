---
id: guide-005-installation-and-initialization
title: Gyrus Installation & Workspace Initialization Guide
category: technical
type: guide
format: markdown
owner_group: armckinney
version: 2
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

This guide covers installing the Gyrus CLI binary on your machine and initializing your workspace or user home with custom configurations, AI agent skills, and Model Context Protocol (MCP) server registrations.

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
1. **Generates Configuration:** Writes a `.gyrus.yaml` file in the workspace root (`storage_root: .gyrus/docs`). Storage directories are created **lazily** on document write (`gyrus create`), keeping workspace root 100% clean on `init`.
2. **Equips Agent Skills:** Installs Gyrus Agent Skill files (`SKILL.md`, `references/`) exclusively into `.agents/skills/gyrus-cli/` and `.agents/skills/gyrus-mcp/`.
3. **Registers MCP Servers (Containerized Stdio by Default):** Non-destructively merges Gyrus MCP stdio server configurations into agent tool JSON files for Google Antigravity, Claude, OpenAI Codex, and GitHub Copilot.

---

## 🎛️ 3. Advanced Customization & Flag Options

### 3.1 MCP Execution Modes (`--mcp-mode`)
Gyrus supports containerized stdio execution (default) and local binary execution:

```bash
# Containerized Stdio (Default) - Runs via Docker without local binary dependencies
gyrus init --mcp-mode container --mcp-container-image ghcr.io/armckinney/gyrus:latest

# Local Binary Execution - Uses locally installed 'gyrus' executable
gyrus init --mcp-mode local
```

### 3.2 Global User Home MCP Setup (`--global` / `-g`)
Register MCP server configurations globally in your user home directory (`~`) across all agent platforms instead of (or in addition to) workspace-local config files:

```bash
gyrus init --global
```

### 3.3 Target Platform Selection (`--mcp-target` & `--skill-target`)
Selectively register MCP servers or equip skills for specific AI agent platforms:

| Platform Target | Flag Example | Generated MCP Configuration File |
| :--- | :--- | :--- |
| **Google Antigravity** | `gyrus init --mcp-target antigravity` | `.antigravity/mcp.json` |
| **Claude Desktop / Code** | `gyrus init --mcp-target claude` | `.claude/mcp.json` & `~/.config/Claude/claude_desktop_config.json` |
| **OpenAI Codex** | `gyrus init --mcp-target codex` | `.codex/mcp.json` |
| **GitHub Copilot / VS Code** | `gyrus init --mcp-target copilot` | `.vscode/mcp.json` |
| **All Active Agents** | `gyrus init --mcp-target all` | *Registers across all 4 platforms* |

### 3.4 Selective Component Initialization (`--no-config`, `--no-mcp`, `--no-skill`)
You can selectively skip setup components if you only want to equip skills or MCP configs without generating a `.gyrus.yaml` file:

```bash
# Skip generating .gyrus.yaml (only equip MCP configs and agent skills)
gyrus init --no-config

# Skip MCP server registration
gyrus init --no-mcp

# Skip Agent Skill equipping
gyrus init --no-skill

# Equip ONLY MCP servers (skip config generation and agent skills)
gyrus init --no-config --no-skill
```

### 3.5 Storage Profile Selection (`--profile`)
Specify a pre-configured storage & search profile using `--profile` (`-p`):

```bash
gyrus init --profile local     # LocalFS Storage + SQLite FTS5 (Default)
gyrus init --profile git       # Remote Git Repository Persistence
gyrus init --profile blob      # Cloud Object Storage (S3, Azure Blob, GCS)
gyrus init --profile postgres  # PostgreSQL Database & FTS Search Backend
```

---

## 🔒 4. Safe Non-Destructive Config Merging

Gyrus unmarshals existing `mcpServers` JSON files and non-destructively merges the `"gyrus"` server definition. All pre-existing user servers (e.g. `sqlite`, `github`, `fetch`) remain completely untouched.
