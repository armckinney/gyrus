# Gyrus Storage & Search Provider Configuration Guide

Gyrus supports pluggable storage and indexing providers configured via `.gyrus.yaml` in the workspace root.

## ⚙️ Configuration File (`.gyrus.yaml`)

```yaml
storage_provider: localfs  # Options: localfs, git, blob
index_provider: sqlite      # Options: sqlite, postgres
search_provider: sqlite     # Options: sqlite, postgres_fts, vector

storage_root: .gyrus/docs
schemas_path: .gyrus/schemas
default_owner_group: armckinney
```

---

## 🗄️ Storage Providers

### 1. Local Filesystem (`localfs`) - Default
Stores OKF Markdown documents directly in the repository filesystem (`.gyrus/docs/okf/<owner_group>/reference/<id>.md`).
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

### 3. Dedicated Cloud Object Storage Drivers (`azure_blob`, `s3`, `gcs`, `blob`)
Cloud-native object storage for Azure Blob Storage, AWS S3, Google Cloud Storage, or generic blob endpoints.

#### Azure Blob Storage (`azure_blob` / `azure`):
```yaml
storage_provider: azure_blob
azure_blob:
  storage_account: "myaccountname"
  container_name: "my-gyrus-container"
```

#### AWS S3 Storage (`s3` / `aws_s3`):
```yaml
storage_provider: s3
s3:
  bucket_name: "my-gyrus-bucket"
  region: "us-east-1"
```

#### Google Cloud Storage (`gcs` / `gcp`):
```yaml
storage_provider: gcs
gcs:
  bucket_name: "my-gyrus-gcs-bucket"
```

> [!IMPORTANT]
> **Required Cloud Storage Permissions & Access Settings**
> Depending on your cloud provider and bucket/container access settings, users and AI agents **MUST have explicit Data Plane Read & Write permissions** assigned:
> - **Azure Blob Storage (`azure_blob`):** Standard Azure management roles (*Owner*, *Contributor*) do **NOT** grant data access. Your Entra ID user or Managed Identity (`az login`) **MUST be assigned the `Storage Blob Data Contributor` or `Storage Blob Data Owner` role** on the storage account/container. Alternatively, export `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` (or `AZURE_STORAGE_SAS_TOKEN`).
> - **AWS S3 (`s3`):** IAM policy must grant `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject`, and `s3:ListBucket` permissions on `arn:aws:s3:::<bucket_name>/*`.
> - **Google Cloud Storage (`gcs`):** Service account / user identity must be assigned `roles/storage.objectAdmin` or `roles/storage.objectUser`.

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
