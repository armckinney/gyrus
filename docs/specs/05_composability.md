# Specification 05: Composability, Portability, & Backend Profiles

To future-proof Gyrus, the storage, search, and relationship backends are fully decoupled. This allows developers to bootstrap with zero external dependencies and scale to large cloud infrastructures later.

---

## 1. The Storage Provider Interface (Repository Pattern)

The API Gateway and CLI connect to storage adapters through an interface wrapper:

```go
type StorageProvider interface {
    GetDocument(ctx context.Context, id string, userGroups []string) (*Document, error)
    SaveDocument(ctx context.Context, doc *Document) error
    SearchDocuments(ctx context.Context, query string, userGroups []string) ([]SearchResult, error)
    GetHistory(ctx context.Context, id string, userGroups []string) ([]HistorySnapshot, error)
}
```

This guarantees that swapping backend drivers has **zero impact** on CLI commands or agent tools.

---

## 2. Progressive Backend Profiles

Gyrus is distributed with three pre-configured deployment profiles:

### A. Local Profile (Default / Bootstrapping)
Designed for local development, repo-local memory, and offline usage. Has **zero external service dependencies**.
* **Document Store:** Local filesystem (JSON contracts or Markdown files with YAML frontmatter).
* **Metadata Index:** SQLite.
* **Full-Text Search:** SQLite FTS5 extension.
* **Graph Relationships:** SQLite `document_edges` table.
* **Human Interface:** Local files opened directly in Obsidian or IDE editors.

### B. Team Profile
Designed for shared team workspaces and concurrent agent edits.
* **Document Store:** PostgreSQL.
* **Metadata Index:** PostgreSQL.
* **Full-Text Search:** PostgreSQL Full-Text Search.
* **Graph Relationships:** PostgreSQL `document_edges` table.
* **Human Interface:** Shared Web Portal (Go+HTMX or Python) reading Postgres.

### C. Enterprise Profile
Designed for multi-tenant organizations requiring fine-grained RBAC and semantic search.
* **Document Store:** PostgreSQL.
* **Metadata Index:** PostgreSQL.
* **Full-Text Search:** Azure AI Search or OpenSearch (Hybrid Lexical + Vector search).
* **Graph Relationships:** PostgreSQL edge tables (with optional Graph DB upgrade path).
* **Human Interface:** Authenticated Web Wiki Portal with access controls.

---

## 3. Agent Integration Shims

To allow multiple coding assistants to query Gyrus, the system provides three portability paths:

* **Claude Code (CLI-First):**
  * Sits in the developer's execution path (e.g. `/usr/local/bin/gyrus`).
  * The agent detects `gyrus` automatically via terminal command checks and executes `gyrus read` or `gyrus search` via shell commands.
* **Cursor / Windsurf / VS Code Copilot (GUI IDEs):**
  * Cursor runs Gyrus natively using the embedded local MCP server (`gyrus mcp serve`).
  * The client connects via stdio, exposing Gyrus tools directly to the IDE's prompt interface.
* **Enterprise CI/CD Pipelines:**
  * Runs `gyrus validate` during pull request steps to block builds if agents try to merge code that violates documented architectural constraints or uses deprecated API contracts.
