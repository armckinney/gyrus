---
id: guide-004-test-phase-2-1
title: Phase 2.1 Manual Validation & Integration Test Guide
category: technical
type: guide
format: markdown
owner_group: armckinney
version: 4
status: active
tags:
  - testing
  - phase-2.1
  - storage-drivers
  - search-providers
  - init
  - skills
dependencies:
  - prd-001-specification-roadmap
  - gyrus-201-git-storage-driver
  - gyrus-202-cloud-blob-storage-driver
  - gyrus-203-postgres-storage-index-driver
  - gyrus-204-postgres-fts-search-driver
  - gyrus-205-vector-hybrid-search-driver
---

# Phase 2.1 Manual Validation & Integration Testing Guide

This guide provides step-by-step instructions to manually test and validate workspace initialization (`../gyrus init`), both agent skills (`gyrus-cli` and `gyrus-mcp`), and each of the five storage and search provider drivers implemented in **Phase 2.1**:

0. **Workspace Initialization & Agent Skill Equipping (`../gyrus init`)**
1. **Git Remote Storage Driver (`git`)** (`GYRUS-201`)
2. **Cloud Blob Storage Driver (`blob`)** (`GYRUS-202`)
3. **PostgreSQL Storage & Index Driver (`postgres`)** (`GYRUS-203`)
4. **PostgreSQL Full-Text Search Engine (`postgres_fts`)** (`GYRUS-204`)
5. **Semantic Vector & Hybrid Search Driver (`vector`)** (`GYRUS-205`)

---

## 🧪 0. Validating Workspace Initialization & Skill Equipping (`../gyrus init`)

Tests master workspace initialization, containerized & local stdio MCP registration, global setup (`--global`), agent skill equipping (`gyrus-cli` and `gyrus-mcp`), and profile selection.

### 0.1 Full Workspace Initialization Test
```bash
# 1. Create a clean test directory
mkdir -p gyrus-init-test && cd gyrus-init-test

# 2. Run master workspace initialization (Containerized Stdio mode by default)
../gyrus init

# 3. Verify workspace root directory layout (zero docs/ folder eagerly created)
ls -la

# 4. Verify .gyrus.yaml configuration file created
cat .gyrus.yaml

# 5. Verify agent skills equipped (.agents/skills/gyrus-cli and .agents/skills/gyrus-mcp)
ls -la .agents/skills/gyrus-cli/SKILL.md .agents/skills/gyrus-mcp/SKILL.md

# 6. Verify stdio MCP servers registered for target platforms
cat .antigravity/mcp.json   # Google Antigravity
cat .claude/mcp.json        # Claude Code & Desktop
cat .codex/mcp.json         # OpenAI Codex
cat .vscode/mcp.json        # GitHub Copilot / VS Code (includes mcpServers & servers blocks)
```

### 0.2 Local Binary & Global Setup Modes
```bash
# Register MCP servers using local binary execution mode instead of Docker
../gyrus init --mcp-mode local

# Register MCP servers globally in user home (~) across all agent tools
../gyrus init --global
```

### 0.3 Tool-Targeted MCP & Skill Initialization Test
```bash
# Target Google Antigravity only
../gyrus init --mcp-target antigravity

# Target Claude Desktop / Claude Code only
../gyrus init --mcp-target claude --skill-target claude

# Target GitHub Copilot / VS Code only
../gyrus init --mcp-target copilot

# Target OpenAI Codex only
../gyrus init --mcp-target codex
```

### 0.4 Selective Component & Headless Initialization Test
```bash
# Skip generating .gyrus.yaml config file (only equip MCP & skills)
../gyrus init --no-config

# Skip MCP server registration for headless / server / CI environments
../gyrus init --no-mcp

# Skip agent skills
../gyrus init --no-skill

# Equip ONLY MCP servers (skip config generation and skills)
../gyrus init --no-config --no-skill
```

### 0.5 Validating `gyrus-cli` Agent Skill (Terminal Agents)
Tests agent execution via terminal CLI subcommands:

```bash
# 1. Verify gyrus-cli frontmatter (name: gyrus-cli)
head -n 6 .agents/skills/gyrus-cli/SKILL.md

# 2. Test suggest-context CLI subcommand
../gyrus suggest-context --prompt "architecture standards" --json

# 3. Test FTS keyword search CLI subcommand
../gyrus search --query "storage engine" --json

# 4. Test creating a contract document via CLI subcommand (lazily creates docs/ directory on write)
../gyrus create \
  --id "adr-cli-skill-test" \
  --title "CLI Skill Test ADR" \
  --category "architecture" \
  --type "adr" \
  --owner-group "armckinney" \
  --status "proposed" \
  --content "Testing CLI skill execution."
```

### 0.6 Validating `gyrus-mcp` Agent Skill (MCP-Native Agents)
Tests agent execution via native MCP tools and JSON-RPC stdio protocol calls:

```bash
# 1. Verify gyrus-mcp frontmatter (name: gyrus-mcp)
head -n 6 .agents/skills/gyrus-mcp/SKILL.md

# 2. Verify stdio MCP server responds to JSON-RPC initialization request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ../gyrus mcp serve

# 3. Test listing registered native MCP tools (gyrus_suggest_context, gyrus_search, gyrus_get_document, gyrus_create_document, gyrus_update_document, gyrus_link_documents, gyrus_sync)
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ../gyrus mcp serve

# 4. Test listing registered MCP resources (gyrus://documents/{id}, gyrus://schema/{type}, gyrus://graph/topology)
echo '{"jsonrpc":"2.0","id":3,"method":"resources/list"}' | ../gyrus mcp serve
```

### 0.7 Automated Skill & Integration Test Suite (`tests/skills/`)
```bash
# Fast, offline Unit & Static Skill Analysis Tests (skips live integration)
make test

# Explicit Live Google Antigravity Integration Test
make test-integration
```

---

## 🧪 1. Validating the Git Remote Storage Driver (`GYRUS-201`)

The Git storage driver provides direct remote Git repository persistence via `go-git` without requiring local workspace clones.

### 1.1 Configuration (`.gyrus.yaml`)
Initialize with Git profile or update `.gyrus.yaml`:

```bash
../gyrus init --profile git
```

```yaml
storage_provider: git
git:
  repo_url: "https://github.com/<your-org>/<your-repo>.git"
  branch: "main"
```

### 1.2 Authentication Setup
Set your authentication token or ensure SSH key availability:

```bash
# Token authentication (GitHub, GitLab, Bitbucket)
export GIT_AUTH_TOKEN="ghp_your_github_personal_access_token"

# Or ensure standard SSH key is available
ls -la ~/.ssh/id_ed25519 ~/.ssh/id_rsa
```

### 1.3 Execution & Verification Steps
```bash
# 1. Create a document directly on remote Git
../gyrus create \
  --id "adr-git-driver-test" \
  --title "Git Driver Test ADR" \
  --category "architecture" \
  --type "adr" \
  --owner-group "armckinney" \
  --status "proposed" \
  --content "Testing remote Git storage driver execution."

# 2. Retrieve document over Git transport
../gyrus get adr-git-driver-test --json

# 3. Verify in GitHub / GitLab web interface that a commit was created with message:
# "gyrus: create adr-git-driver-test (v1)"
```

---

## 🧪 2. Validating the Azure Blob Storage Driver (`GYRUS-202`)

The Azure Blob Storage driver (`azure_blob` / `azure`) provides cloud-native persistence directly in Azure Storage containers using `gocloud.dev/blob/azureblob`.

### 2.1 Configuration (`.gyrus.yaml`)
Initialize with Azure profile or update `.gyrus.yaml`:

```bash
../gyrus init --profile azure
```

```yaml
storage_provider: azure_blob
index_provider: sqlite
search_provider: sqlite

storage_root: .gyrus/docs
schemas_path: .gyrus/schemas
default_owner_group: armckinney

azure_blob:
  storage_account: "myaccountname"
  container_name: "my-gyrus-container"
```

### 2.2 Azure Authentication Setup
Set your Azure Storage credentials via environment variables or Azure CLI:

```bash
# Option A: Storage Account & Access Key
export AZURE_STORAGE_ACCOUNT="myaccountname"
export AZURE_STORAGE_KEY="your_azure_storage_account_key"

# Option B: Shared Access Signature (SAS) Token
export AZURE_STORAGE_ACCOUNT="myaccountname"
export AZURE_STORAGE_SAS_TOKEN="sv=2020-08-04&ss=b&srt=sco&..."

# Option C: Azure Identity / Azure CLI Login
az login
```

### 2.3 Execution & Verification Steps
```bash
# 1. Create a document directly in your Azure Blob Storage container
../gyrus create \
  --id "prd-azure-blob-test" \
  --title "Azure Blob Driver Test PRD" \
  --category "product" \
  --type "prd" \
  --owner-group "armckinney" \
  --status "draft" \
  --content "Testing Azure Blob Storage driver execution."

# 2. Retrieve document from Azure Blob Store
../gyrus get prd-azure-blob-test --json

# 3. Verify in Azure Portal / Azure CLI that object key was created:
# my-gyrus-container/.gyrus/docs/okf/armckinney/workspaces/main/prd-azure-blob-test.md
az storage blob list --account-name myaccountname --container-name my-gyrus-container --output table
```

---

## 🧪 3. Validating the PostgreSQL Enterprise Storage Driver (`GYRUS-203`)

The PostgreSQL driver provides centralized database storage (`DocumentStore`, `IndexStore`, `GraphStore`) using `pgx/v5`.

### 3.1 Start a Local PostgreSQL Container
```bash
docker run -d \
  --name gyrus-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gyrus \
  -p 5432:5432 \
  postgres:16
```

### 3.2 Configuration (`.gyrus.yaml`)
Initialize with postgres profile:

```bash
../gyrus init --profile postgres
```

### 3.3 Execution & Verification Steps
```bash
# 1. Run sync to execute DDL migrations and populate PostgreSQL schema
../gyrus sync --json

# 2. Create a document in PostgreSQL
../gyrus create \
  --id "spec-postgres-driver-test" \
  --title "PostgreSQL Driver Test Spec" \
  --category "technical" \
  --type "specification" \
  --owner-group "armckinney" \
  --status "active" \
  --content "Testing PostgreSQL enterprise storage backend."

# 3. Query PostgreSQL tables directly
docker exec -it gyrus-postgres psql -U postgres -d gyrus -c "SELECT id, title, category, status FROM documents;"
docker exec -it gyrus-postgres psql -U postgres -d gyrus -c "SELECT from_id, to_id, rel_type FROM document_edges;"
```

---

## 🧪 4. Validating the PostgreSQL Full-Text Search Engine (`GYRUS-204`)

Tests native PostgreSQL `tsvector` and `tsquery` full-text search capabilities.

### 4.1 Configuration (`.gyrus.yaml`)
```yaml
storage_provider: postgres
index_provider: postgres
search_provider: postgres_fts
postgres:
  connection_string: "postgres://postgres:postgres@localhost:5432/gyrus?sslmode=disable"
```

### 4.2 Execution & Verification Steps
```bash
# 1. Execute FTS keyword search
../gyrus search --query "enterprise storage" --json

# 2. Verify search output returns relevant documents with PostgreSQL ts_rank_cd() scores.
```

---

## 🧪 5. Validating Semantic Vector & Hybrid Search (`GYRUS-205`)

Tests semantic vector similarity search and Reciprocal Rank Fusion (RRF) hybrid search.

### 5.1 Local Zero-Infra Mode (Local Ollama)
```bash
# 1. Start Ollama with an embedding model (e.g. nomic-embed-text)
ollama pull nomic-embed-text
```

Initialize with vector profile:

```bash
../gyrus init --profile vector
```

### 5.2 Execution & Verification Steps
```bash
# 1. Perform semantic context resolution for a concept prompt
../gyrus suggest-context --prompt "how to store relational SQL records" --json

# 2. Verify that vector search matches semantically related documents (e.g. spec-postgres-driver-test)
# even if the exact words "relational" or "SQL" do not appear in the document title!
```

---

## 📋 Quick Health Check Script

You can also run automated unit test verification across all provider packages and setup routines:

```bash
# Run unit & static skill tests across all packages
make test

# Run explicit live integration tests
make test-integration
```
