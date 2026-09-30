# Gyrus Storage & Search Provider Configuration Guide

Gyrus supports pluggable storage and indexing providers configured via `.gyrus.yaml` in the workspace root or `~/.gyrus.yaml` in the user's home directory.

## ⚙️ Configuration File (`.gyrus.yaml`)

```yaml
storage:
  provider: localfs  # Options: localfs, git, blob, s3, azure_blob, gcs, postgres
  root: .gyrus

index:
  provider: sqlite   # Options: sqlite, postgres

graph:
  provider: sqlite   # Options: sqlite, postgres

search:
  provider: sqlite   # Options: sqlite, sqlite_fts5, postgres_fts, vector

default_owner_group: root
```

---

## 🗄️ Storage Providers

### 1. Local Filesystem (`localfs`) - Default
Stores OKF Markdown documents directly in the repository filesystem (`.gyrus/docs/<owner_group>/reference/<id>.md`).
- **Zero Infra:** No external databases required.
- **Git Native:** Files are committed directly into version control.

### 2. Git Remote Driver (`git`)
Direct remote Git repository persistence via `go-git`.
```yaml
storage:
  provider: git
  root: .gyrus
git:
  repo_url: https://github.com/my-org/my-docs.git
  branch: main
```

### 3. Dedicated Cloud Object Storage Drivers (`azure_blob`, `s3`, `gcs`, `blob`)
Cloud-native object storage for Azure Blob Storage, AWS S3, Google Cloud Storage, or generic blob endpoints.

#### Azure Blob Storage (`azure_blob` / `azure`):
```yaml
storage:
  provider: azure_blob
  root: .gyrus
azure_blob:
  storage_account: "myaccountname"
  container_name: "my-gyrus-container"
```

#### AWS S3 Storage (`s3` / `aws_s3`):
```yaml
storage:
  provider: s3
  root: .gyrus
s3:
  bucket_name: "my-gyrus-bucket"
  region: "us-east-1"
```

#### Google Cloud Storage (`gcs` / `gcp`):
```yaml
storage:
  provider: gcs
  root: .gyrus
gcs:
  bucket_name: "my-gyrus-gcs-bucket"
```

### 4. PostgreSQL Enterprise Driver (`postgres`)
Centralized database backend for multi-tenant enterprise deployments.
```yaml
storage:
  provider: postgres
  root: .gyrus
index:
  provider: postgres
graph:
  provider: postgres
search:
  provider: postgres_fts
postgres:
  connection_string: postgres://user:password@localhost:5432/gyrus?sslmode=disable
```

---

## 🔍 Search & Vector Providers

### Vector & Hybrid Search (`vector`)
```yaml
search:
  provider: vector
vector:
  embedding_provider: ollama  # Options: ollama, openai
  model: nomic-embed-text
  ollama_endpoint: http://localhost:11434
```
