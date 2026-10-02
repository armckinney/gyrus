---
id: guide-004-test-phase-2-1
title: Phase 2.1 Manual Validation & Integration Test Guide
category: technical
type: guide
format: markdown
owner_group: root
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
  - spec-201-git-storage-driver
  - spec-202-cloud-blob-storage-driver
  - spec-203-postgres-storage-index-driver
  - spec-204-postgres-fts-search-driver
  - spec-205-vector-hybrid-search-driver
---

# Phase 2.1 Manual Validation & Integration Testing Guide

This guide provides step-by-step instructions to manually test and validate workspace initialization (`../gyrus init`), both agent skills (`gyrus-cli` and `gyrus-mcp`), and each of the five storage and search provider drivers implemented in **Phase 2.1**:

0. **Workspace Configuration & Agent Plugin Setup (`gyrus config init` & `gyrus client install`)**
1. **Git Remote Storage Driver (`git`)** (`spec-201-git-storage-driver`)
2. **Cloud Blob Storage Driver (`blob`)** (`spec-202-cloud-blob-storage-driver`)
3. **PostgreSQL Storage & Index Driver (`postgres`)** (`spec-203-postgres-storage-index-driver`)
4. **PostgreSQL Full-Text Search Engine (`postgres_fts`)** (`spec-204-postgres-fts-search-driver`)
5. **Semantic Vector & Hybrid Search Driver (`vector`)** (`spec-205-vector-hybrid-search-driver`)

---

## 🧪 0. Validating Workspace Configuration & Agent Setup (`gyrus config init` & `gyrus client install`)

> [!NOTE]
> Workspace configuration and agent installation are decoupled into explicit subcommands: `gyrus config init` for repository configuration and `gyrus client install` for agent plugin installation.

### 0.1 Full Workspace Configuration Test
```bash
# 1. Create a clean test directory
mkdir -p gyrus-init-test && cd gyrus-init-test

# 2. Run workspace configuration initialization
../gyrus config init

# 3. Verify workspace root directory layout
ls -la

# 4. Verify .gyrus.yaml configuration file created
cat .gyrus.yaml

# 5. Equip Agent Plugin for AI coding assistants
../gyrus client install --target antigravity
../gyrus client install --target copilot

# 6. Verify agent plugin and MCP configurations created
ls -la .agents/plugins/gyrus/plugin.json .agents/plugins/gyrus/skills/gyrus/SKILL.md
cat .vscode/mcp.json
```

### 0.2 Global Configuration Mode
```bash
# Initialize global configuration in user home directory (~)
../gyrus config init --global

# Equip client plugin globally
../gyrus client install --target antigravity --global
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
  --owner-group "root" \
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
  --owner-group "root" \
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

storage_root: ""
schemas_path: schemas
default_owner_group: root

azure_blob:
  storage_account: "stgyrusdev"
  container_name: "docs"
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
  --owner-group "root" \
  --status "draft" \
  --content "Testing Azure Blob Storage driver execution."

# 2. Retrieve document from Azure Blob Store
../gyrus get prd-azure-blob-test --json

# 3. Verify in Azure Portal / Azure CLI that object key was created:
# my-gyrus-container/.gyrus/docs/okf/root/workspaces/main/prd-azure-blob-test.md
az storage blob list --account-name stgyrusdev --container-name docs --output table
```

---

## 🧪 3. Validating the PostgreSQL Enterprise Storage & FTS Driver (`GYRUS-203` & `GYRUS-204`)

The PostgreSQL driver provides centralized enterprise storage (`DocumentStore`, `IndexStore`, `GraphStore`) and native `tsvector`/`tsquery` full-text search (`search_provider: postgres_fts`) using `pgx/v5`.

### 3.1 DevContainer Automatic Sidecar
The DevContainer setup ([`.devcontainer/docker-compose.yaml`](file:///workspaces/gyrus/.devcontainer/docker-compose.yaml)) includes PostgreSQL 16 as an automatic sidecar container:
- **Hostname inside DevContainer:** `postgres` (or `localhost`)
- **Port:** `5432`
- **Database / User / Password:** `gyrus` / `postgres` / `postgres`

*(If running outside a DevContainer, launch PostgreSQL via Docker: `docker run -d --name gyrus-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=gyrus -p 5432:5432 postgres:16-alpine`)*

### 3.2 Configuration (`.gyrus.yaml`)
Initialize with PostgreSQL profile:

```bash
../gyrus init --profile postgres
```

```yaml
storage_provider: postgres
index_provider: postgres
search_provider: postgres_fts

storage_root: .gyrus/docs
schemas_path: .gyrus/schemas
default_owner_group: root

postgres:
  connection_string: "postgres://postgres:postgres@postgres:5432/gyrus?sslmode=disable"
```

### 3.3 Execution & Verification Steps
```bash
# 1. Run sync to execute DDL migrations and initialize PostgreSQL schema
../gyrus sync --json

# 2. Create a contract document directly in PostgreSQL
../gyrus create \
  --id "spec-postgres-driver-test" \
  --title "PostgreSQL Driver Test Spec" \
  --category "technical" \
  --type "specification" \
  --owner-group "root" \
  --status "active" \
  --content "Testing PostgreSQL enterprise storage backend and full-text search."

# 3. Retrieve document envelope from PostgreSQL
../gyrus get spec-postgres-driver-test --json

# 4. Perform native PostgreSQL tsvector FTS keyword search
../gyrus search --query "enterprise storage" --json
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
