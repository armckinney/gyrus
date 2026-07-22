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

| Profile | Storage Provider | Index Provider | Graph Provider | Search Provider | Target Use Case & Environment |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`Test`** | `localfs` | OKF or SQLite | OKF or SQLite | SQLite FTS5 (Optional) | Unit tests, test fixtures, local mock pipelines, fast CI validation runs. |
| **`Local Full`** | SQLite | SQLite | SQLite edge tables | SQLite FTS5 | Local application development, deterministic local integration tests. |
| **`Small`** *(MVP Default)* | Git repo / `localfs` | OKF | OKF links / metadata | OKF / local scan | Personal and small team repositories; zero-dependency Git storage. |
| **`Medium`** | Blob storage (S3/Azure) | PostgreSQL | PostgreSQL edge tables | PostgreSQL full-text | Team deployments and shared service backend instances. |
| **`Large`** | PostgreSQL | PostgreSQL | PostgreSQL edge tables | PostgreSQL FTS | Centralized platform enterprise service supporting multi-tenant access. |

---

## 3. Storage Path Resolution Hierarchy

When Gyrus operates in **Small / Local Profiles**, it determines its storage root directory using the following precedence (highest priority first):

1. **CLI Flag:** `--storage-path <path>` passed explicitly on command invocation.
2. **Environment Variable:** `GYRUS_STORAGE_PATH` environment variable.
3. **Project Config File:** `storage.root` property defined in `.gyrus/config.yaml`.
4. **User Config File:** `storage.root` property defined in `~/.config/gyrus/config.yaml`.
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



