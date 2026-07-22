# Gyrus Configuration Reference Manual (`config.yaml`)

This document provides a detailed reference for all configuration options supported in `.gyrus/config.yaml` or `~/.config/gyrus/config.yaml`.

---

## 1. Example Configuration File

```yaml
version: 1
profile: small # test | local_full | small | medium | large

storage:
  provider: localfs # localfs | git | blob | sqlite | postgres
  root: ~/.gyrus/   # Path to OKF bundle storage directory

index:
  provider: sqlite  # okf | sqlite | postgres
  dsn: ~/.gyrus/index.db # SQLite DSN or file path

graph:
  provider: sqlite  # okf | sqlite | postgres

search:
  provider: sqlite_fts5 # none | okf_scan | sqlite_fts5 | postgres_fts
```

---

## 2. Configuration Field Reference

### Root Settings

| Property | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`version`** | Integer | `1` | Configuration schema version identifier. Must be set to `1`. |
| **`profile`** | String | `small` | Preset deployment profile. Options: `test`, `local_full`, `small`, `medium`, `large`. |

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

## 4. Storage Resolution Precedence Hierarchy

Gyrus resolves configuration files using the following strict priority order (highest priority first):

1. **CLI Flag:** `--storage-path <path>` passed explicitly on CLI invocation.
2. **Environment Variable:** `GYRUS_STORAGE_PATH` environment variable.
3. **Repository Project Config:** `.gyrus.yaml`, `.gyrus.yml`, `.gyrus/config.yaml`, or `.gyrus/config.yml` located in the current working directory or parent directories. Relative `storage.root` paths resolve relative to the repository root containing the config file.
4. **User Home Config:** `~/.config/gyrus/config.yaml`, `~/.config/gyrus/config.yml`, `~/.gyrus.yaml`, or `~/.gyrus.yml` in user's home directory.
5. **Default Fallback:** `~/.gyrus/` directory.
