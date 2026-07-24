# Gyrus Storage & Search Provider Configuration Guide

Gyrus supports pluggable storage and indexing providers configured via `.gyrus.yaml` in the workspace root.

## ⚙️ Configuration File (`.gyrus.yaml`)

```yaml
storage_provider: localfs  # Options: localfs, git, blob
index_provider: sqlite      # Options: sqlite, postgres
search_provider: sqlite     # Options: sqlite, postgres_fts, vector

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: armckinney
```

---

## 🗄️ Storage Providers

### 1. Local Filesystem (`localfs`) - Default
Stores OKF Markdown documents directly in the repository filesystem (`docs/.gyrus/docs/okf/<owner_group>/reference/<id>.md`).
- **Zero Infra:** No external databases required.
- **Git Native:** Files are committed directly into version control.

### 2. Git Remote Driver (`git`)
Direct remote Git repository persistence via `go-git`.
```yaml
storage_provider: git
git:
  repo_url: https://github.com/my-org/my-docs.git
  branch: main
```

### 3. Cloud Blob Storage (`blob`)
Cloud-native object storage for AWS S3, Azure Blob, and Google Cloud Storage.
```yaml
storage_provider: blob
blob:
  bucket_url: s3://my-gyrus-bucket?region=us-east-1
  prefix: docs
```

### 4. PostgreSQL Enterprise Driver (`postgres`)
Centralized database backend for multi-tenant enterprise deployments.
```yaml
storage_provider: postgres
index_provider: postgres
postgres:
  connection_string: postgres://user:password@localhost:5432/gyrus?sslmode=disable
```

---

## 🔍 Search & Vector Providers

### Vector & Hybrid Search (`vector`)
```yaml
search_provider: vector
vector:
  embedding_provider: ollama  # Options: ollama, openai
  model: nomic-embed-text
  ollama_endpoint: http://localhost:11434
```
