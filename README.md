# Gyrus: Unified Context & Memory Engine

> High-performance, local-first context management engine and knowledge graph for software development teams and AI agents.

---

## 🌟 Overview

**Gyrus** is an open-source, local-first memory engine designed to bridge the gap between human engineering teams and AI coding assistants. Built in Go, Gyrus organizes codebase context, Architecture Design Records (ADRs), specifications, PRDs, and guidelines using the **Open Knowledge Format (OKF)** standard.

Gyrus delivers zero-dependency embedded search (powered by CGO-free SQLite FTS5), incremental filesystem synchronization, document state-machine validation, a standalone CLI (`gyrus`), and an embedded Model Context Protocol (MCP) server for GUI IDEs (Cursor, Claude Desktop, Copilot).

---

## 🏛️ High-Level System Architecture

```mermaid
graph LR
    subgraph Clients ["Callers & Clients"]
        CICD["CICD & Scripting"]
        AgentTools["Agentic Tool(s)"]
        Humans["Human Users"]
    end

    subgraph Interface ["Adapters & Application Surfaces"]
        Skill["Skill (Open Skill Format)"]
        CLI["Gyrus CLI"]
        MCP["Gyrus MCP Server"]
        Apps["Applications\ni.e. Serverless Content Navigator Web App\n- Implements AI Chatbot"]
    end

    subgraph Core ["Gyrus Core SDK"]
        CoreSDK["Gyrus Core SDK\n\n# Applies (MVP Provider):\n- Concept Storage Format (OKF)\n- Concept Index Provider (OKF)\n- Knowledge Graph Provider (OKF)\n- Search Provider (None)\n- Concept Persistence Service (localfs)\n- Concept Types (OKF)\n\n# Implements:\n- OKF Standard Metadata & Concepts"]
    end

    subgraph Services ["Core Services & Provider Drivers"]
        IndexSvc["Concept Index Service\n\nProviders:\n- OKF\n- SQLite\n- PostgreSQL"]
        GraphSvc["Knowledge Graph Service\n\nProviders:\n- OKF\n- SQLite\n- PostgreSQL"]
        SearchSvc["Search Service\n\nProviders:\n- SQLite FTS\n- PostgreSQL full-text search"]
        PersistSvc["Concept Persistence Service\n\nProviders:\n- localfs\n- SQLite\n- Git Repo (GitHub, Bitbucket)\n- Blob Storage (S3, Azure SA)\n- PostgreSQL"]
    end

    subgraph Storage ["Concept Storage Format"]
        StorageFmt["Concept Storage Format\n\nProviders:\n- OKF\n- Relational Table"]
    end

    %% Client and Adapter Relations
    CICD -->|invokes| CLI
    AgentTools -->|invokes| Skill
    Skill -->|invokes| CLI
    CLI -->|implements| CoreSDK
    AgentTools -->|OR invokes| MCP
    MCP -->|implements| CoreSDK
    Humans -->|invokes| Apps
    Apps -->|Implements| CoreSDK

    %% Core SDK Services Relations
    CoreSDK -->|instantiates| IndexSvc
    CoreSDK -->|instantiates| GraphSvc
    CoreSDK -->|instantiates| SearchSvc
    CoreSDK -->|instantiates| PersistSvc

    %% Persistence to Storage Format
    PersistSvc -->|instantiates| StorageFmt
```

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
3. `.gyrus.yaml` / `.gyrus/config.yaml` local config file
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

- 🏛️ **[System Architecture](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/guide-001-system-architecture.md):** Complete guide to the Gyrus Core SDK, Provider Framework, OKF directory topology, and state machines.
- ⚙️ **[Configuration Reference](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/tech-ref-002-config-schema.md):** Comprehensive reference for all `.gyrus.yaml` options, profiles, and path precedence.
- 🛠️ **[CLI Reference Manual](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/tech-ref-001-cli-manual.md):** Detailed argument and flag reference for all 11 `gyrus` CLI subcommands and exit codes.
- 🔌 **[MCP Server Setup Guide](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/guide-004-mcp-server-setup.md):** Native and Docker containerized MCP stdio server setup for Cursor, Claude Desktop, and VS Code.
- 🤖 **[Agent Skills Setup Guide](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/guide-003-agent-skills-setup.md):** Instructions for copying `.agents/skills/gyrus/SKILL.md` into code repositories for Claude Code and terminal agents.
- 📋 **[Specification Implementation Roadmap](file:///workspaces/gyrus/docs/.gyrus/docs/okf/armckinney/reference/prd-001-specification-roadmap.md):** Remaining planned feature roadmap including cloud providers, PostgreSQL, vector search, and web dashboard.

---

## 🛠️ Developer Verification

Run unit and integration test suites:

```bash
make test
```
