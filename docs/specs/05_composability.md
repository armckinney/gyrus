# Specification 05: Composability, Provider Framework, & Profiles

To future-proof Gyrus, all storage, indexing, knowledge graph, and search operations are fully decoupled into composable **Provider Interfaces**. This allows developers to bootstrap locally with zero external dependencies and seamlessly scale to large enterprise infrastructures.

---

## 1. Provider Interface Taxonomy

The **Gyrus Core SDK** composes five distinct provider abstractions:

1. **`StorageProvider`:** Manages physical file or database record persistence (`localfs`, Git Repo via GitHub/Bitbucket APIs, S3 / Azure SA Blob Storage, SQLite, PostgreSQL).
2. **`IndexProvider`:** Manages structured document metadata indexing (OKF frontmatter parser, SQLite, PostgreSQL).
3. **`KnowledgeGraphProvider`:** Manages document relationships and edge traversals (OKF `dependencies` links, SQLite edge tables, PostgreSQL edge tables).
4. **`SearchProvider`:** Manages full-text lexical or semantic context retrieval (None / OKF scan, SQLite FTS5, PostgreSQL full-text search).
5. **`DocumentPersistenceService`:** Manages physical file layout and bundle serialization (`localfs` OKF bundle topology).

---

## 2. Progressive Provider Profiles Matrix

Gyrus defines five pre-configured deployment profiles tailored to specific execution environments:

| Profile | Status | Storage Provider | Index Provider | Graph Provider | Search Provider | Target Use Case & Environment |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **`Test`** | `IMPLEMENTED` | `localfs` | OKF / SQLite | OKF / SQLite | SQLite FTS5 | Unit tests, test fixtures, local mock pipelines. |
| **`Local Full`** | `IMPLEMENTED` | `localfs` / SQLite | SQLite | SQLite edge tables | SQLite FTS5 | Local application development, standalone CLI & MCP. |
| **`Small`** *(Default)* | `IMPLEMENTED` | `localfs` | OKF | OKF links / metadata | SQLite FTS5 / OKF scan | Personal and small team repositories. |
| **`Medium`** | `PLANNED` | Blob storage (S3/Azure) | PostgreSQL | PostgreSQL edge tables | PostgreSQL full-text | Team deployments and shared service backend instances. |
| **`Large`** | `PLANNED` | PostgreSQL | PostgreSQL | PostgreSQL edge tables | PostgreSQL FTS | Centralized platform enterprise service supporting multi-tenant access. |

---

## 3. Provider Capability Matrix

| Provider Type | Driver Name | Status | Description |
| :--- | :--- | :--- | :--- |
| **Storage** | `localfs` | `IMPLEMENTED` | Local filesystem storage for OKF Markdown directory bundles. |
| **Storage** | `git` / `blob` / `postgres` | `PLANNED` | Remote Git, Cloud S3/Azure Blob, and PostgreSQL database storage. |
| **Index** | `sqlite` / `okf` | `IMPLEMENTED` | Embedded SQLite `documents_index` and direct YAML frontmatter validation. |
| **Index** | `postgres` | `PLANNED` | PostgreSQL metadata indexer. |
| **Graph** | `sqlite` / `okf` | `IMPLEMENTED` | SQLite `document_edges` graph store and frontmatter dependency extraction. |
| **Graph** | `postgres` | `PLANNED` | PostgreSQL relationship edge tables. |
| **Search** | `sqlite_fts5` / `okf_scan` | `IMPLEMENTED` | CGO-free SQLite FTS5 full-text keyword search and filesystem scan filter. |
| **Search** | `postgres_fts` | `PLANNED` | PostgreSQL full-text search engine. |



---

## 3. Storage Path Resolution Hierarchy

When Gyrus operates in **Small / Local Profiles**, it determines its storage root directory using the following precedence (highest priority first):

1. **CLI Flag:** `--storage-path <path>` passed explicitly on command invocation.
2. **Environment Variable:** `GYRUS_STORAGE_PATH` environment variable.
3. **Repository Project Config File:** `.gyrus.yaml`, `.gyrus.yml`, `.gyrus/config.yaml`, or `.gyrus/config.yml` located in the current working directory or any parent repository directory. Relative paths in `storage.root` resolve relative to the directory containing the config file.
4. **User Home Config File:** `~/.config/gyrus/config.yaml`, `~/.config/gyrus/config.yml`, `~/.gyrus.yaml`, or `~/.gyrus.yml`.
5. **Default Fallback:** `~/.gyrus/` (global application storage) or `./.gyrus/` (project-local directory).


---

## 4. Client Integration Interfaces & Web Architecture

Gyrus supports four primary client entrypoints:

```
                  ┌───────────────────────────────┐
                  │          Agentic Tools        │
                  └───────┬───────────────┬───────┘
                          │               │
                          ▼               ▼
                  ┌───────────────┐ ┌─────────────┐
                  │ Skill Adapter │ │ MCP Server  │
                  │ (Open Skill)  │ │ (stdio/SSE) │
                  └───────┬───────┘ └───────┬─────┘
                          │ (CLI)           │ (Go SDK)
                          ▼                 │
                  ┌───────────────┐         │
                  │   Gyrus CLI   │◄────────┤
                  └───────┬───────┘         │
                          │ (Go SDK)        │
                          ▼                 ▼
                  ┌───────────────────────────────┐
                  │        Gyrus Core SDK         │
                  └───────────────────────────────┘
```

1. **Direct Core SDK (Applications):** Go applications import `pkg/gyrus` directly to interact with context programmatically.
2. **Gyrus CLI (Scripts & Agents):** CLI agents (like Claude Code) execute terminal commands (`gyrus search`, `gyrus get`, `gyrus create`).
3. **Skill Adapter (Open Skill Format):** Enables agents using the Open Skill Format standard to wrap and invoke the `gyrus` CLI binary transparently.
4. **Gyrus MCP Server (GUI IDEs):** IDE clients (Cursor, Windsurf, Copilot) connect to `gyrus mcp serve` via stdio/SSE to execute Gyrus tools natively.
5. **Serverless Content Navigator Web App (Human Web Portal):** For non-CLI human users, a serverless web application provides an interactive documentation browser with an embedded **AI Chatbot**. It reads directly from persistence services (Blob Storage or Document Databases) to present structured context without requiring local CLI tools.

---

## 5. Technical Stack & Configuration Schema

### A. Official Go Third-Party Libraries
To maintain consistency and enable seamless CGO-free cross-compilation, Gyrus mandate the following Go libraries:

| Purpose | Selected Library | Key Benefit |
| :--- | :--- | :--- |
| **SQLite Driver** | `modernc.org/sqlite` | Pure Go implementation of SQLite; requires **zero CGO** for Linux, macOS, and Windows cross-compilation via GoReleaser. |
| **CLI Framework** | `github.com/spf13/cobra` | Industry-standard Go CLI routing engine for command flags and subcommands. |
| **MCP Server SDK** | `github.com/mark3labs/mcp-go` | Standard Model Context Protocol (MCP) server library for stdio/SSE transports. |
| **YAML Parser** | `gopkg.in/yaml.v3` | Fast frontmatter and application configuration parser. |

### B. Standard `config.yaml` Schema
```yaml
version: 1
profile: small # test | local_full | small | medium | large

storage:
  provider: localfs # localfs | git | blob | sqlite | postgres
  root: ~/.gyrus/   # Storage path resolution root

index:
  provider: sqlite  # okf | sqlite | postgres
  dsn: ~/.gyrus/index.db

graph:
  provider: sqlite  # okf | sqlite | postgres

search:
  provider: sqlite_fts5 # none | okf_scan | sqlite_fts5 | postgres_fts
```



