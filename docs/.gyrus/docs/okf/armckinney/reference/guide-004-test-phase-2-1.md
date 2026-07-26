---
id: guide-004-test-phase-2-1
title: Phase 2.1 Manual Validation & Integration Test Guide
category: technical
type: guide
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-24T21:04:45Z
tags:
    - testing
    - phase-2.1
    - manual-validation
---

---
id: guide-004-test-phase-2-1
title: Phase 2.1 Manual Validation & Integration Test Guide
category: technical
type: guide
format: markdown
owner_group: armckinney
version: 1
status: active
tags:
  - testing
  - phase-2.1
  - storage-drivers
  - search-providers
dependencies:
  - prd-001-specification-roadmap
  - gyrus-201-git-storage-driver
  - gyrus-202-cloud-blob-storage-driver
  - gyrus-203-postgres-storage-index-driver
  - gyrus-204-postgres-fts-search-driver
  - gyrus-205-vector-hybrid-search-driver
---

# Phase 2.1 Manual Validation & Integration Testing Guide

This guide provides step-by-step instructions to manually test and validate each of the five storage and search provider drivers implemented in **Phase 2.1**:

1. **Git Remote Storage Driver (`git`)** (`GYRUS-201`)
2. **Cloud Blob Storage Driver (`blob`)** (`GYRUS-202`)
3. **PostgreSQL Storage & Index Driver (`postgres`)** (`GYRUS-203`)
4. **PostgreSQL Full-Text Search Engine (`postgres_fts`)** (`GYRUS-204`)
5. **Semantic Vector & Hybrid Search Driver (`vector`)** (`GYRUS-205`)

---

## 🧪 1. Validating the Git Remote Storage Driver (`GYRUS-201`)

The Git storage driver provides direct remote Git repository persistence via `go-git` without requiring local workspace clones.

### 1.1 Configuration (`.gyrus.yaml`)
Update `.gyrus.yaml` in your workspace root:

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
./gyrus create \
  --id "adr-git-driver-test" \
  --title "Git Driver Test ADR" \
  --category "architecture" \
  --type "adr" \
  --owner-group "armckinney" \
  --status "proposed" \
  --content "Testing remote Git storage driver execution."

# 2. Retrieve document over Git transport
./gyrus get adr-git-driver-test --json

# 3. Verify in GitHub / GitLab web interface that a commit was created with message:
# "gyrus: create adr-git-driver-test (v1)"
```

---

## 🧪 2. Validating the Cloud Blob Storage Driver (`GYRUS-202`)

The Cloud Blob Storage driver uses `gocloud.dev/blob` to store OKF documents across AWS S3 (`s3://`), Azure Blob Storage (`azblob://`), Google Cloud Storage (`gs://`), or local filesystem buckets (`file://`).

### 2.1 Local Test (using `fileblob` bucket)
```bash
# Create a local test bucket folder
mkdir -p /tmp/gyrus-blob-bucket
```

Update `.gyrus.yaml`:

```yaml
storage_provider: blob
blob:
  bucket_url: "file:///tmp/gyrus-blob-bucket"
  prefix: "docs"
```

### 2.2 Execution & Verification Steps
```bash
# 1. Create a document in blob storage
./gyrus create \
  --id "prd-blob-driver-test" \
  --title "Blob Driver Test PRD" \
  --category "product" \
  --type "prd" \
  --owner-group "armckinney" \
  --status "draft" \
  --content "Testing cloud blob storage driver."

# 2. Inspect blob key hierarchy on disk
ls -la /tmp/gyrus-blob-bucket/docs/armckinney/product/prd-blob-driver-test.md

# 3. Retrieve document from blob store
./gyrus get prd-blob-driver-test --json
```

### 2.3 Cloud S3 / Azure / GCS Test (Optional)
```yaml
# AWS S3 (uses default AWS credentials / AWS_PROFILE / IAM Role)
storage_provider: blob
blob:
  bucket_url: "s3://my-gyrus-bucket?region=us-east-1"
  prefix: "gyrus-docs"
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
Update `.gyrus.yaml`:

```yaml
storage_provider: postgres
index_provider: postgres
postgres:
  connection_string: "postgres://postgres:postgres@localhost:5432/gyrus?sslmode=disable"
```

### 3.3 Execution & Verification Steps
```bash
# 1. Run sync to execute DDL migrations and populate PostgreSQL schema
./gyrus sync --json

# 2. Create a document in PostgreSQL
./gyrus create \
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
./gyrus search --query "enterprise storage" --json

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

Update `.gyrus.yaml`:

```yaml
search_provider: vector
vector:
  embedding_provider: ollama
  model: nomic-embed-text
  ollama_endpoint: "http://localhost:11434"
```

### 5.2 Cloud Mode (OpenAI)
```yaml
search_provider: vector
vector:
  embedding_provider: openai
  model: text-embedding-3-small
```
```bash
export OPENAI_API_KEY="sk-your-openai-api-key"
```

### 5.3 Execution & Verification Steps
```bash
# 1. Perform semantic context resolution for a concept prompt
./gyrus suggest-context --prompt "how to store relational SQL records" --json

# 2. Verify that vector search matches semantically related documents (e.g. spec-postgres-driver-test)
# even if the exact words "relational" or "SQL" do not appear in the document title!
```

---

## 📋 Quick Health Check Script

You can also run automated unit test verification across all provider drivers:

```bash
# Run unit & integration tests across all provider packages
go test ./internal/provider/... -v
```
