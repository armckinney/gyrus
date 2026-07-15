# AI Agent Memory Tool Architecture & Specification

## 1. Overview

This project defines a generalized, configurable context and memory management tool for AI coding agents such as Claude Code, GitHub Copilot, Antigravity, and similar agentic development environments.

The goal is to provide agents with a controlled way to discover, retrieve, create, update, validate, and navigate durable project context before executing user requests.

The system should prevent agents from storing arbitrary unstructured context by enforcing strong data contracts for supported document types such as ADRs, improvement proposals, runbooks, decision logs, repo notes, and other structured memory records.

The system should also support human consumption through readable document formats and a graphical interface for browsing, searching, editing, validating, and navigating memory documents.

---

## 2. Core Design Principles

### 2.1 Contracts Must Be Enforced by the Application

Data contracts must not rely on agent prompts, Skills, or instructions alone.

Prompts and Skills can instruct agents to use the memory tool, but only the memory application should be allowed to persist memory.

The enforcement boundary is the core memory engine.

Agents may request memory operations, but the tool must validate and reject invalid writes.

```text
Correct:
Agent -> CLI/MCP -> Core Memory Engine -> Backend

Incorrect:
Agent -> Skill prompt convention -> direct file edit / arbitrary context write
```

### 2.2 One Source of Truth

There should be exactly one durable source of truth for memory documents.

Search indexes, graph indexes, embedding indexes, caches, and UI projections should be rebuildable derived state.

```text
Source of truth:
  durable document store

Derived state:
  metadata index
  full-text search index
  semantic/vector index
  relationship graph
  cache
```

### 2.3 Shared Core, Multiple Interfaces

The CLI, MCP server, and GUI should all call the same core memory engine.

```text
CLI        -> Core Memory Engine -> Backend
MCP Server -> Core Memory Engine -> Backend
Web UI     -> Core Memory Engine -> Backend
```

The core engine owns:

* Schema validation
* Document lifecycle rules
* Metadata enforcement
* Storage operations
* Index updates
* Search behavior
* Relationship management
* Policy enforcement
* Versioning rules

### 2.4 Agent Instructions Are Guidance, Not Enforcement

Agent Skills, `AGENTS.md`, Copilot instructions, or similar context files should explain when and how agents should use the tool.

They should not be treated as the data integrity mechanism.

Example instruction:

```markdown
Before making architecture, platform, or implementation changes, call the memory tool to retrieve relevant ADRs, proposals, repo notes, and known constraints.

Do not create or edit memory documents directly. Use `agent-memory create`, `agent-memory update`, or the MCP memory tools.
```

---

## 3. Recommended Implementation Language

The implementation should be written in Go.

Go is appropriate because this tool behaves like infrastructure:

* Cross-platform CLI distribution
* Single binary installation
* Fast startup time
* Strong typing
* Good filesystem support
* Good server support
* Good fit for local-first and service-backed deployment models
* Suitable for embedding CLI, MCP server, and web UI in one binary

The preferred binary name is:

```text
agent-memory
```

---

## 4. High-Level Architecture

```text
                    +----------------------+
                    |  Agent Instructions  |
                    |  Skill / AGENTS.md   |
                    +----------+-----------+
                               |
                               v
+-------------+        +--------------+        +----------------------+
| Human / CI  | -----> |     CLI      | -----> |                      |
+-------------+        +--------------+        |                      |
                                               |                      |
+-------------+        +--------------+        |  Core Memory Engine  |
| MCP Client  | -----> | MCP Server   | -----> |                      |
+-------------+        +--------------+        |                      |
                                               |                      |
+-------------+        +--------------+        |                      |
| Human User  | -----> | Web UI       | -----> |                      |
+-------------+        +--------------+        +----------+-----------+
                                                          |
                                                          v
                                               +----------------------+
                                               |      Backend(s)      |
                                               +----------------------+
```

---

## 5. Preferred Binary Shape

Use one Go binary with multiple subcommands.

```text
agent-memory init
agent-memory create
agent-memory get
agent-memory update
agent-memory delete
agent-memory validate
agent-memory search
agent-memory suggest-context
agent-memory index
agent-memory serve
agent-memory mcp serve
```

The CLI, MCP server, and web UI should be different interfaces over the same core package.

---

## 6. Recommended Repository Structure

```text
agent-memory/
  cmd/
    agent-memory/
      main.go

  internal/
    cli/
      commands/

    mcp/
      server/
      tools/
      resources/
      prompts/

    ui/
      handlers/
      templates/
      static/

    memory/
      service.go
      documents.go
      lifecycle.go
      validation.go
      search.go
      graph.go

    schema/
      loader.go
      validator.go

    storage/
      filesystem/
      postgres/
      s3/

    index/
      sqlite/
      postgres/
      opensearch/

    graph/
      sqlite/
      postgres/
      neo4j/

    policy/
      permissions.go
      lifecycle.go

    config/
      loader.go
      profiles.go

  schemas/
    adr.schema.json
    proposal.schema.json
    runbook.schema.json
    repo-note.schema.json
    decision-log.schema.json

  docs/
    examples/
    architecture.md
    configuration.md

  skills/
    memory-management/
      SKILL.md

  examples/
    config.local.yaml
    config.team.yaml
    config.enterprise.yaml
```

---

## 7. Interface Model

### 7.1 CLI Interface

The CLI is the primary human, CI, and shell-agent interface.

It should be stable, programmatic, and safe to wrap from MCP.

CLI requirements:

* Support `--json` output for every command
* Support structured input via `--input file.json` or stdin
* Use stable exit codes
* Return machine-readable validation errors
* Avoid interactive prompts in agent/MCP mode
* Support idempotency keys for create/update operations where useful
* Support optimistic locking for updates
* Support clear permission and policy failures
* Never silently persist invalid documents

Example command:

```bash
agent-memory create \
  --type adr \
  --input ./adr.json \
  --json
```

Example validation failure:

```json
{
  "ok": false,
  "error_code": "VALIDATION_FAILED",
  "errors": [
    {
      "path": "metadata.status",
      "message": "Required field missing"
    }
  ]
}
```

### 7.2 MCP Interface

The MCP server should expose an agent-native interface to the same core memory engine.

Preferred model:

```text
MCP Client -> agent-memory mcp serve -> Core Memory Engine -> Backend
```

The MCP server should not duplicate business logic.

Recommended MCP tools:

```text
memory.suggest_context
memory.search
memory.get
memory.create
memory.update
memory.delete
memory.validate
memory.link
```

Recommended MCP resources:

```text
memory://doc/{id}
memory://schema/{document_type}
memory://repo/{repo}/context
memory://adr
memory://proposal
memory://runbook
memory://decision-log
```

Recommended MCP prompts:

```text
prepare-adr
review-implementation-against-memory
summarize-relevant-context
propose-memory-update
```

### 7.3 Skill / AGENTS.md Interface

Skills and instruction files should tell agents when to use memory.

They should not define the true data contract.

Example:

```markdown
# Memory Management

Use this workflow before:
- Starting implementation work
- Making architecture changes
- Modifying platform behavior
- Writing ADRs or proposals
- Answering questions about historical decisions

Procedure:
1. Call `agent-memory suggest-context` or the MCP `memory.suggest_context` tool.
2. Read returned ADRs, proposals, repo notes, and runbooks.
3. Do not directly edit memory files.
4. Use the memory tool for create/update/delete operations.
5. If validation fails, fix the structured fields before continuing.
```

---

## 8. Core Memory Engine

The core memory engine is the actual product.

It should be independent of CLI, MCP, and UI concerns.

Responsibilities:

* Load configuration
* Resolve backend adapters
* Load schemas
* Validate documents
* Enforce required metadata
* Enforce document lifecycle rules
* Create documents
* Read documents
* Update documents
* Delete or archive documents
* Search documents
* Suggest relevant context
* Manage links and relationships
* Maintain indexes
* Support migrations
* Support audit/versioning behavior

Suggested Go interfaces:

```go
type DocumentStore interface {
    Create(ctx context.Context, doc Document) (DocumentRef, error)
    Get(ctx context.Context, id string) (Document, error)
    Update(ctx context.Context, id string, patch DocumentPatch) (DocumentRef, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter DocumentFilter) ([]DocumentSummary, error)
}

type IndexStore interface {
    Index(ctx context.Context, doc Document) error
    Remove(ctx context.Context, id string) error
    Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)
}

type GraphStore interface {
    UpsertEdges(ctx context.Context, edges []DocumentEdge) error
    Neighbors(ctx context.Context, id string, filter EdgeFilter) ([]DocumentEdge, error)
    Traverse(ctx context.Context, query GraphQuery) ([]GraphPath, error)
}
```

---

## 9. Document Model

Each memory document should contain:

* Stable ID
* Document type
* Schema version
* Title
* Status
* Owners
* Tags
* Scope
* Created timestamp
* Updated timestamp
* Related documents
* Supersession metadata
* Visibility/sensitivity metadata
* Body content

Example ADR metadata:

```yaml
id: adr-2026-001
type: adr
schema_version: 1
title: Use MCP and CLI for Agent Memory Integration
status: accepted
owners:
  - platform-engineering
scope:
  repos:
    - ml-platform/*
  systems:
    - agent-memory
tags:
  - agents
  - context-management
  - mcp
  - cli
created_at: 2026-07-13
updated_at: 2026-07-13
supersedes: []
related:
  proposals:
    - aip-2026-001
visibility: internal
sensitivity: low
```

Recommended document types:

```text
adr
proposal
runbook
repo_note
decision_log
project_brief
system_context
integration_note
known_issue
migration_note
```

---

## 10. Validation Strategy

Use two validation layers:

### 10.1 Go Types

Go structs provide internal correctness and developer ergonomics.

```go
type ADR struct {
    ID            string   `json:"id" yaml:"id"`
    Type          string   `json:"type" yaml:"type"`
    SchemaVersion int      `json:"schema_version" yaml:"schema_version"`
    Title         string   `json:"title" yaml:"title"`
    Status        string   `json:"status" yaml:"status"`
    Owners        []string `json:"owners" yaml:"owners"`
    Tags          []string `json:"tags" yaml:"tags"`
    CreatedAt     string   `json:"created_at" yaml:"created_at"`
    UpdatedAt     string   `json:"updated_at" yaml:"updated_at"`
    Body          string   `json:"body" yaml:"-"`
}
```

### 10.2 JSON Schema

JSON Schema provides the external contract for document types.

Benefits:

* Versionable contracts
* Extensible document types
* Schema-driven validation
* Schema documentation
* Cross-language compatibility
* Easier external integrations

The system should reject documents that fail schema validation.

---

## 11. Backend Architecture

The backend should be configurable but opinionated.

The recommended model is:

```text
Document Store:
  durable source of truth

Metadata Index:
  fast filtering and listing

Search Index:
  keyword and optional semantic retrieval

Graph Store:
  relationships among documents, systems, repos, teams, and decisions
```

Search indexes and graph stores should be rebuildable from the source of truth.

---

## 12. Backend Profiles

Rather than requiring users to design their own backend stack from scratch, provide blessed profiles.

### 12.1 Local Profile

Best for local development, repo-local memory, and MVP adoption.

```text
Document store:
  filesystem

Document format:
  Markdown with YAML frontmatter

Metadata index:
  SQLite

Full-text search:
  SQLite FTS

Graph:
  SQLite edge tables

UI:
  local Go + HTMX server
```

Example layout:

```text
.agent-memory/
  memory/
    adr/
      adr-2026-001.md
    proposals/
      aip-2026-001.md
    runbooks/
    repo-notes/

  schemas/
    adr.schema.json
    proposal.schema.json

  index/
    memory.sqlite

  config.yaml
```

### 12.2 Team Profile

Best for shared team memory and concurrent access.

```text
Document store:
  Postgres

Metadata index:
  Postgres

Full-text search:
  Postgres full-text search

Graph:
  Postgres edge tables

UI:
  shared Go + HTMX web app

Optional:
  object storage for large artifacts
```

### 12.3 Enterprise Profile

Best for large-scale, cross-organization memory.

```text
Document store:
  Postgres

Metadata/index:
  Postgres

Search:
  OpenSearch or similar search backend

Semantic search:
  vector index or embedding-backed search

Graph:
  SQL edge tables initially
  optional graph database later

Artifacts:
  object storage

UI:
  shared authenticated web app

Additional concerns:
  RBAC
  audit trails
  central policy
  multi-tenant boundaries
```

---

## 13. Graph Strategy

Do not start with a dedicated graph database.

Start with an edge table.

```text
document_edges
  from_document_id
  to_document_id
  relationship_type
  confidence
  created_by
  created_at
```

Example relationships:

```text
adr supersedes adr
proposal resulted_in adr
repo implements adr
service depends_on service
team owns system
doc references doc
runbook mitigates known_issue
```

A graph database should only be introduced after SQL edge tables are insufficient.

Valid reasons to introduce a graph database:

* Complex multi-hop traversal
* Dependency impact analysis
* Large-scale graph visualization
* Ownership and system topology analysis
* Centrality or clustering analysis
* Cross-repo architectural mapping

---

## 14. Search Strategy

The system should support layered search.

### 14.1 Metadata Filtering

Examples:

```text
type = adr
status = accepted
owner = platform-engineering
repo = ml-platform/foo
tag = databricks
updated_after = 2026-01-01
```

### 14.2 Keyword Search

Use full-text search for document titles, metadata, and body content.

Local profile:

```text
SQLite FTS
```

Team profile:

```text
Postgres full-text search
```

Enterprise profile:

```text
OpenSearch or similar
```

### 14.3 Semantic Search

Semantic search should be optional and pluggable.

The core memory contract should not depend on embeddings.

Possible providers:

```text
OpenAI-compatible embeddings
local embedding service
Databricks model serving
enterprise embedding provider
disabled
```

Semantic search should augment, not replace, metadata and keyword search.

---

## 15. Configuration Strategy

Use both YAML and environment variables.

### 15.1 YAML

YAML should define durable application configuration:

* Backend profile
* Storage type
* Index type
* Search providers
* Graph provider
* Schema paths
* Document type mappings
* UI settings
* Policy settings

### 15.2 `.env`

Environment variables should hold secrets and local overrides:

* Config path
* Database DSN
* Search service URL
* Tokens
* Credentials
* Local machine overrides

Do not put the whole backend design in `.env`.

Example `config.yaml`:

```yaml
version: 1

profile: local

storage:
  type: filesystem
  root: ./.agent-memory/memory
  format: markdown_yaml

index:
  type: sqlite
  path: ./.agent-memory/index/memory.sqlite
  full_text: true

search:
  lexical:
    enabled: true
    provider: sqlite_fts5
  semantic:
    enabled: false
    provider: none

graph:
  enabled: true
  provider: sqlite_edges

schemas:
  root: ./.agent-memory/schemas
  document_types:
    adr:
      schema: adr.schema.json
      path: adr/
    proposal:
      schema: proposal.schema.json
      path: proposals/
    runbook:
      schema: runbook.schema.json
      path: runbooks/
    repo_note:
      schema: repo-note.schema.json
      path: repo-notes/

ui:
  enabled: true
  host: 127.0.0.1
  port: 7357

policy:
  require_owners: true
  require_status: true
  allow_direct_file_edits: false
  require_change_reason_on_update: true
```

Example `.env`:

```bash
AGENT_MEMORY_CONFIG=./.agent-memory/config.yaml
AGENT_MEMORY_POSTGRES_DSN=
AGENT_MEMORY_OPENSEARCH_URL=
AGENT_MEMORY_OPENSEARCH_TOKEN=
```

---

## 16. Human GUI

A human-facing GUI is a first-class requirement.

The GUI should allow users to:

* Browse documents
* Search memory
* Filter by metadata
* View document relationships
* Edit documents through validated forms
* Review validation errors
* View document history
* View schemas
* Navigate ADR/proposal/runbook relationships
* Trigger reindexing
* Inspect backend health

Recommended implementation:

```text
Go server
HTMX frontend
Server-rendered templates
Minimal JavaScript
```

Preferred command:

```bash
agent-memory serve
```

Recommended routes:

```text
/documents
/documents/:id
/documents/:id/edit
/documents/:id/history
/search
/graph
/schemas
/validation
/admin/index
/admin/health
```

The UI should use the same core memory engine as the CLI and MCP server.

---

## 17. Persistence Strategy

### 17.1 V1 Recommended Backend

Start with:

```text
Filesystem Markdown/YAML
+ SQLite metadata index
+ SQLite FTS search
+ SQLite edge table graph
```

Why:

* Simple
* Portable
* Human-readable
* Git-reviewable
* Easy to adopt
* Works locally
* Works in CI
* Does not require services
* Easy to migrate later

### 17.2 V2 Recommended Backend

Add:

```text
Postgres document store
+ Postgres metadata index
+ Postgres full-text search
+ Postgres edge tables
```

Why:

* Shared team usage
* Concurrent writes
* Better permissions
* Centralized audit
* Shared UI
* Better service deployment

### 17.3 V3 Recommended Backend

Add:

```text
Postgres source of truth
+ OpenSearch search/vector index
+ optional graph database
+ object storage
```

Why:

* Large-scale search
* Semantic retrieval
* Multi-team usage
* Large artifacts
* Relationship-heavy queries

---

## 18. Document Lifecycle Rules

Each document type should define allowed lifecycle states.

Example ADR states:

```text
proposed
accepted
superseded
deprecated
rejected
```

Example proposal states:

```text
draft
reviewing
approved
rejected
implemented
abandoned
```

Updates should require:

* Valid schema
* Existing document ID
* Expected version or revision
* Change reason
* Valid lifecycle transition
* Updated timestamp
* Optional reviewer or approver metadata

Example invalid transitions:

```text
accepted -> draft
superseded -> accepted
rejected -> implemented
```

---

## 19. Versioning and Audit

The system should support versioning.

Local profile:

```text
Git history
document revision metadata
optional local changelog
```

Team/enterprise profile:

```text
audit table
document version table
created_by
updated_by
change_reason
timestamp
previous_version
new_version
```

Update operations should support optimistic locking.

Example:

```json
{
  "id": "adr-2026-001",
  "expected_version": "3",
  "patch": {
    "status": "superseded"
  },
  "change_reason": "Replaced by ADR-2026-009"
}
```

---

## 20. Security and Policy

The system should eventually support:

* Read/write permissions
* Document visibility
* Sensitivity labels
* Owner enforcement
* Allowed document types
* Backend access control
* Audit logging
* Policy-controlled deletion
* Soft delete/archive
* Secret scanning or sensitive content checks

Recommended metadata:

```yaml
visibility: internal
sensitivity: low
owners:
  - platform-engineering
```

Deletion should usually be soft-delete or archive, not physical removal.

---

## 21. Recommended MVP Scope

The MVP should include:

```text
Go binary:
  agent-memory

Interfaces:
  CLI
  MCP server
  local web UI

Backend:
  filesystem document store
  SQLite index
  SQLite FTS
  SQLite edge table

Document types:
  adr
  proposal
  runbook
  repo_note
  decision_log

Validation:
  JSON Schema
  required metadata
  lifecycle states

Agent integration:
  AGENTS.md instructions
  optional Skill package
  MCP tools

Commands:
  init
  create
  get
  update
  validate
  search
  suggest-context
  index
  serve
  mcp serve
```

---

## 22. Recommended Initial CLI Commands

```bash
agent-memory init

agent-memory create \
  --type adr \
  --input ./adr.json \
  --json

agent-memory get \
  --id adr-2026-001 \
  --json

agent-memory update \
  --id adr-2026-001 \
  --input ./patch.json \
  --json

agent-memory validate \
  --all \
  --json

agent-memory search \
  --query "databricks auth scopes" \
  --json

agent-memory suggest-context \
  --request "Update the Databricks deployment workflow" \
  --repo "." \
  --json

agent-memory index rebuild

agent-memory serve

agent-memory mcp serve
```

---

## 23. Recommended MCP Tools

### `memory.suggest_context`

Input:

```json
{
  "request": "string",
  "repo": "string",
  "path": "string",
  "max_results": 10,
  "doc_types": ["adr", "proposal", "runbook"]
}
```

Output:

```json
{
  "summary": "string",
  "documents": [
    {
      "id": "adr-2026-001",
      "type": "adr",
      "title": "Use Auth Scopes for Databricks Security Boundaries",
      "status": "accepted",
      "relevance_reason": "Matches repo and deployment workflow context",
      "uri": "memory://doc/adr-2026-001",
      "last_updated": "2026-07-13"
    }
  ]
}
```

### `memory.create`

Input:

```json
{
  "doc_type": "adr",
  "title": "string",
  "metadata": {},
  "body": "string"
}
```

Output:

```json
{
  "id": "adr-2026-001",
  "uri": "memory://doc/adr-2026-001",
  "validation_status": "passed"
}
```

### `memory.update`

Input:

```json
{
  "id": "adr-2026-001",
  "expected_version": "3",
  "patch": {},
  "change_reason": "string"
}
```

Output:

```json
{
  "id": "adr-2026-001",
  "new_version": "4",
  "validation_status": "passed"
}
```

### `memory.validate`

Input:

```json
{
  "doc_type": "adr",
  "metadata": {},
  "body": "string"
}
```

Output:

```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

---

## 24. Design Decision Summary

### Chosen

```text
Language:
  Go

Primary artifact:
  Single `agent-memory` binary

Interfaces:
  CLI
  MCP server
  Go + HTMX web UI

Contract enforcement:
  Core memory engine

Agent guidance:
  Skill / AGENTS.md / Copilot instructions

V1 backend:
  Filesystem Markdown/YAML + SQLite

V2 backend:
  Postgres

V3 backend:
  Postgres + OpenSearch + optional graph database

Config:
  YAML for durable config
  .env for secrets/local overrides

Graph:
  Start with SQL edge tables
  Add graph DB only when justified

Search:
  Metadata + full-text first
  Semantic search optional later
```

### Explicitly Avoid

```text
Do not rely on Skills/prompts for data enforcement.
Do not allow agents to hand-edit memory documents as the normal write path.
Do not introduce graph databases before SQL edge tables prove insufficient.
Do not require service dependencies for the MVP.
Do not put all configuration in .env.
Do not bury business logic inside CLI command handlers.
Do not make search indexes the source of truth.
```

---

## 25. Final Architecture Recommendation

Build `agent-memory` as a Go-based single binary that provides:

```text
CLI:
  for humans, CI, and shell-based agents

MCP server:
  for agent-native integrations

Web UI:
  for human browsing, editing, validation, and navigation

Core memory engine:
  for contracts, lifecycle, storage, search, graph, and policy

Configurable backends:
  starting with local filesystem + SQLite
  expanding to Postgres and search services later
```

The final target model is:

```text
Skill / AGENTS.md
  -> instructs agents when to use memory

CLI
  -> stable human and automation interface

MCP
  -> structured agent-native interface

Web UI
  -> human consumption and management

Core Memory Engine
  -> actual enforcement layer

Backend
  -> durable source of truth plus rebuildable indexes
```

The most important architectural rule:

```text
Agents may request memory operations, but only the memory tool may persist memory.
```
