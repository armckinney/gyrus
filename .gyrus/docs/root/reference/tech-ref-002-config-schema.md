---
id: tech-ref-002-config-schema
title: Gyrus YAML Configuration Reference & Schema
category: technical
type: technical-reference
format: ""
owner_group: root
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus Configuration Reference Manual (`.gyrus.yaml`)

This document provides a detailed reference for all configuration options supported in workspace `.gyrus.yaml` or user-wide global `~/.gyrus.yaml`.

---

## 1. Example Configuration File

```yaml
storage:
  provider: localfs # localfs | git | blob | s3 | azure_blob | gcs | postgres
  root: .gyrus      # Path to storage directory (relative to config file)

index:
  provider: sqlite  # sqlite | postgres
  dsn: .gyrus/index.db

graph:
  provider: sqlite  # sqlite | postgres

search:
  provider: sqlite_fts5 # sqlite | sqlite_fts5 | postgres_fts | vector

default_owner_group: root
```

---

## 2. Configuration Field Reference

### Root Settings

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`version`** | Integer | `1` | Configuration schema version identifier. Must be set to `1`. |
| **`profile`** | String | `small` | Preset deployment profile. Options: `test`, `local_full`, `small`, `medium`, `large`. |
| **`schemas_path`** | String | `""` | *(Obsolete / Deprecated)* Schemas are now stored directly in the persistence layer under `.gyrus/schemas/`. |



---

### Storage Settings (`storage`)

Controls document payload persistence.

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`storage.provider`** | String | `localfs` | Driver for document persistence. Options: `localfs`, `git`, `blob`, `sqlite`, `postgres`. |
| **`storage.root`** | String | `~/.gyrus/` | Target filesystem directory path for storing OKF Markdown bundles. Supports `~` home directory expansion. |

---

### Index Settings (`index`)

Controls structured metadata indexing.

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`index.provider`** | String | `sqlite` | Metadata indexing engine. Options: `okf` (direct frontmatter scan), `sqlite`, `postgres`. |
| **`index.dsn`** | String | `~/.gyrus/index.db` | Data Source Name (DSN) or database file path for the metadata index database. |

---

### Knowledge Graph Settings (`graph`)

Controls document relationship edge traversals (`depends_on`, `supersedes`, `implements`, `mitigates`).

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`graph.provider`** | String | `sqlite` | Driver for relationship edge queries. Options: `okf` (frontmatter dependencies), `sqlite` (edge tables), `postgres`. |

---

### Search Settings (`search`)

Controls full-text lexical and semantic search indexing.

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`search.provider`** | String | `sqlite_fts5` | Full-text search engine. Options: `none`, `okf_scan`, `sqlite_fts5`, `postgres_fts`. |

## 3. Provider Capability Matrix

The table below indicates which provider drivers are **Implemented** versus **Planned**:

| Provider Type | Driver Name | Status | Description & Use Case |
| :--- | :--- | :--- | :--- |
| **Storage** | `localfs` | `IMPLEMENTED` | Local filesystem storage supporting OKF directory bundles. |
| **Storage** | `git` | `PLANNED` | Remote Git repository persistence via GitHub/Bitbucket APIs. |
| **Storage** | `blob` | `PLANNED` | Cloud object storage (AWS S3, Azure Blob Storage, GCP Bucket). |
| **Storage** | `postgres` | `PLANNED` | Centralized relational database document storage. |
| **Index** | `sqlite` | `IMPLEMENTED` | Embedded SQLite `documents_index` metadata table (CGO-free). |
| **Index** | `okf` | `IMPLEMENTED` | Direct YAML frontmatter schema parsing and validation. |
| **Index** | `postgres` | `PLANNED` | PostgreSQL document index backend. |
| **Graph** | `sqlite` | `IMPLEMENTED` | Embedded SQLite `document_edges` table with traversal & BFS support. |
| **Graph** | `okf` | `IMPLEMENTED` | Frontmatter dependency links extraction (`dependencies: [...]`). |
| **Graph** | `postgres` | `PLANNED` | PostgreSQL edge relationship tables. |
| **Search** | `sqlite_fts5` | `IMPLEMENTED` | FTS5 full-text lexical keyword search engine with ranking. |
| **Search** | `okf_scan` | `IMPLEMENTED` | Direct filesystem scan matching metadata filters. |
| **Search** | `postgres_fts`| `PLANNED` | PostgreSQL full-text search query engine. |


---

## 4. Storage & Configuration Precedence Hierarchy

Gyrus resolves configuration files using a strictly file-driven 3-tier priority order with full override semantics (highest priority first):

1. **Repository Workspace Config:** `.gyrus.yaml` located in current working directory or any parent repository directory. If present, workspace configuration completely overrides global configuration. Relative `storage.root` paths resolve relative to the directory containing `.gyrus.yaml`.
2. **Global User Config:** `~/.gyrus.yaml` in the user's home directory. Applies user-wide defaults across all repositories when no workspace configuration is found. Relative `storage.root` paths resolve relative to `$HOME`.
3. **Default Fallback:** Built-in defaults utilizing LocalFS storage driver rooted at `~/.gyrus/`.
