---
id: spec-002-cli-interface
title: Cobra CLI Executable & Command Line Interface Specs
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Specification 02: CLI Interface & Authentication

The CLI tool (`gyrus`) is a compiled standalone Go binary that serves as the execution gateway. It handles configuration management, local caching, developer/agent authentication, and direct communication with the Gyrus Core Engine.

---

## 1. CLI Command Reference

### `gyrus init`
* **Description:** Bootstraps Gyrus configuration in the target environment.
* **Flags:**
  * `--storage-path <path>` (optional): Custom storage directory (default: `~/.gyrus/` or `./.gyrus/` or `docs/gyrus/`).
  * `--profile <local|team|enterprise>` (optional): Sets backend profile (default: `local`).

### `gyrus create`
* **Description:** Creates a new OKF document. Metadata is passed as CLI flags; content is passed inline or via a file.
* **Flags:**
  * `--id <id>` (required): Unique document ID.
  * `--title <title>` (required): Human-readable title.
  * `--type <type>` (required): Document type (e.g. `adr`, `prd`, `specification`).
  * `--category <category>` (required): Category (`architecture`, `technical`, etc.).
  * `--status <status>` (required): Initial state (e.g. `proposed`, `draft`).
  * `--owner-group <group>` (required): Owning group (e.g. `engineering`).
  * `--tags <tag1,tag2>` (optional): Comma-separated list of tags.
  * `--dependencies <id1,id2>` (optional): Comma-separated list of dependent document IDs.
  * `--content <markdown>` (optional): Inline Markdown text body.
  * `--content-file <path>` (optional): Path to a file containing Markdown text body.
  * `--json` (optional): Output JSON result envelope to `stdout`.

### `gyrus get`
* **Description:** Retrieves a document by ID.
* **Flags:**
  * `--id <id>` (required): Document ID.
  * `--json` (optional): Output full JSON envelope instead of raw Markdown content body.

### `gyrus update`
* **Description:** Updates an existing document envelope or content with optimistic concurrency control.
* **Flags:**
  * `--id <id>` (required): Target document ID.
  * `--expected-version <v>` (required): Expected current version integer for lock verification.
  * `--status <status>` (optional): New status (must satisfy transition rules).
  * `--reason <text>` (required): Audit log reason for update.
  * `--content <markdown>` (optional): Updated inline content body.
  * `--content-file <path>` (optional): Path to updated content file.
  * `--json` (optional): Output JSON result envelope.

### `gyrus link` / `gyrus unlink`
* **Description:** Manages document relationships in the active storage backend.
* **Flags:**
  * `--from <id>` (required): Source document ID.
  * `--to <id>` (required): Target document ID.
  * `--rel <type>` (required): Relationship type (`supersedes`, `depends_on`, `implements`, `mitigates`).

### `gyrus search`
* **Description:** Performs full-text lexical search and metadata filtering over indexed documents.
* **Flags:**
  * `--query <search-string>` (required): Search term.
  * `--type <type>` (optional): Filter by document type.
  * `--category <category>` (optional): Filter by category.
  * `--status <status>` (optional): Filter by status.
  * `--tag <tag>` (optional): Filter by tag.
  * `--json` (optional): Output JSON search results.

### `gyrus suggest-context`
* **Description:** Assembles a linearized context package relevant to an input prompt or user request.
* **Flags:**
  * `--request <prompt>` (required): Coding task or prompt context.
  * `--max-results <num>` (optional): Maximum document count (default: 5).
  * `--json` (optional): Output JSON package.

### `gyrus sync`
* **Description:** Triggers an explicit re-index scan over the configured storage directory to synchronize SQLite metadata indexes with files on disk.

### `gyrus validate`
* **Description:** Validates document frontmatter and state-machine transitions without persisting changes.
* **Flags:**
  * `--file <path>` (optional): Validate a specific Markdown file.
  * `--all` (optional): Validate all documents in the active workspace.

### `gyrus schema`
* **Description:** Prints the recommended structural Markdown template for a specified type (e.g. `gyrus schema adr`).

### `gyrus mcp serve`
* **Description:** Launches the local Model Context Protocol (MCP) server over standard input/output (`stdio`).

---

## 2. Programmatic Exit Codes

The `gyrus` CLI returns deterministic exit codes to enable self-healing agent automation:

| Code | Status Name | Description |
| :--- | :--- | :--- |
| **`0`** | `SUCCESS` | Operation completed successfully. |
| **`1`** | `VALIDATION_ERROR` | Document metadata, OKF envelope, or schema validation failed. |
| **`2`** | `TRANSITION_ERROR` | Requested lifecycle state update violates state-machine rules. |
| **`3`** | `AUTH_ERROR` | Authentication failed or user lacks group permissions. |
| **`4`** | `CONCURRENCY_CONFLICT` | Optimistic lock check failed (`expected-version` mismatch). |
| **`5`** | `STORAGE_ERROR` | File I/O failure or database connection error. |

---

## 3. Authentication & Session Caching

When configured in **Team/Enterprise Profiles**, `gyrus` uses Entra ID (Azure AD) tokens:
1. **Azure CLI Handshake:** Executes `az account get-access-token --resource <API_CLIENT_ID>`.
2. **Device Code Fallback:** Prompts device login via `https://microsoft.com/devicelogin`.
3. **Session Cache:** Saves session JWT to `~/.config/gyrus/token.json`.
4. **Offline Read Cache:** Caches retrieved contracts to `~/.cache/gyrus/` for offline access.
