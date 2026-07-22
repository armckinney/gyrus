# Gyrus: Unified Context & Memory Engine

> High-performance, local-first context management engine and knowledge graph for software development teams and AI agents.

---

## 🌟 Overview

**Gyrus** is an open-source, local-first memory engine designed to bridge the gap between human engineering teams and AI coding assistants. Built in Go, Gyrus organizes codebase context, Architecture Design Records (ADRs), specifications, PRDs, and guidelines using the **Open Knowledge Format (OKF)** standard.

Gyrus delivers zero-dependency embedded search (powered by CGO-free SQLite FTS5), incremental filesystem synchronization, document state-machine validation, a standalone CLI (`gyrus`), and an embedded Model Context Protocol (MCP) server for GUI IDEs (Cursor, Claude Desktop, Copilot).

---

## 🚀 Getting Started

### 1. Installation

Install the pre-compiled `gyrus` binary automatically across Linux, macOS, and Windows (Git Bash/WSL):

```bash
curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash
```

*Alternatively, build from source using Go 1.25+:*

```bash
git clone https://github.com/armckinney/gyrus.git
cd gyrus
make build
```


This compiles the standalone `gyrus` executable into the workspace root.

### 2. Initialize Gyrus Storage

Initialize Gyrus in your repository workspace:

```bash
./gyrus init
```

By default, Gyrus resolves storage path hierarchy in the following order:
1. `--storage-path` CLI flag
2. `GYRUS_STORAGE_PATH` environment variable
3. `.gyrus/config.yaml` local config file
4. `~/.config/gyrus/config.yaml` user config file
5. `~/.gyrus/` default application directory

### 3. Create your first OKF Document

Create an Architecture Design Record (ADR):

```bash
./gyrus create \
  --id "adr-001-storage-engine" \
  --title "Use Embedded SQLite FTS5 for Gyrus Search" \
  --category "architecture" \
  --type "adr" \
  --owner-group "platform" \
  --content "We choose CGO-free SQLite FTS5 for zero-dependency local keyword search."
```

### 4. Search and Suggest Context

Search across your contract documents:

```bash
./gyrus search --query "SQLite search"
```

Suggest linearized context matching an agent prompt:

```bash
./gyrus suggest-context --prompt "How is local search implemented in Gyrus?"
```

---

## 📚 Documentation Sitemap

- 🏛️ **[System Architecture](file:///workspaces/gyrus/docs/usage/architecture.md):** Complete guide to the Gyrus Core SDK, Provider Framework, OKF directory topology, and state machines.
- ⚙️ **[Configuration Reference](file:///workspaces/gyrus/docs/usage/reference/config.md):** Comprehensive reference for all `.gyrus/config.yaml` options, profiles, and path precedence.
- 🛠️ **[CLI Reference Manual](file:///workspaces/gyrus/docs/usage/reference/cli.md):** Detailed argument and flag reference for all 11 `gyrus` CLI subcommands and exit codes.
- 🔌 **[MCP Server Reference](file:///workspaces/gyrus/docs/usage/reference/mcp.md):** Setup guide for Cursor, Claude Desktop, and VS Code with MCP tool signatures.
- 📖 **[Setup & Onboarding Guide](file:///workspaces/gyrus/docs/usage/guides/setup.md):** Step-by-step tutorial for setting up team context and configuring agent skills.


---

## 🛠️ Developer Verification

Run unit and integration test suites:

```bash
make test
```
