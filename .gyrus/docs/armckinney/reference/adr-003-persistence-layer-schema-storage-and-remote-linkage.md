---
id: adr-003-persistence-layer-schema-storage-and-remote-linkage
title: Persistence Layer Schema Storage & Remote Linkage
category: architecture
type: adr
format: markdown
owner_group: armckinney
version: 1
status: accepted
last_modified_by: antigravity
last_updated: 2026-09-09T00:27:00Z
tags:
  - adr
  - architecture
  - persistence
  - schemas
  - phase-2-5
dependencies:
  - adr-001-oop-core-refactoring-and-package-reorganization
  - adr-002-simplified-directory-topology
---

# ADR 003: Persistence Layer Schema Storage & Remote Linkage

## Context

In earlier iterations of Gyrus, document schema templates (`<doc-type>.md`) were either compiled directly into the binary via Go embed or resolved via an arbitrary local configuration property (`schemas_path` in `.gyrus.yaml`).

This architecture presented critical limitations in enterprise, multi-agent, and remote cloud deployments:
1. **Schema Drift Across Clients & Agents**: When multiple developers or autonomous AI agents interacted with remote repositories (Git, AWS S3, Azure Blob, Google Cloud Storage, or PostgreSQL), local schema definitions frequently diverged, resulting in inconsistent document envelopes and validation failures.
2. **Lack of Remote Schema Management**: There was no standard programmatic interface to upload, update, list, or delete custom schema templates in remote persistence stores without manually modifying local files.
3. **Arbitrary Directory Locations**: The `schemas_path` setting allowed unpredictable custom directory paths, complicating cross-workspace synchronization and cloud storage key mapping.
4. **Agent Tooling Gap**: AI agents using the Model Context Protocol (MCP) lacked tools to inspect available schema types, persist custom schema templates, or query schema resources.

## Decision

We decide to formalize schema storage directly within the Gyrus persistence layer:

1. **`gyrus.SchemaStore` Interface**:
   Introduce a dedicated `SchemaStore` interface in `pkg/gyrus` implemented by all four storage provider drivers:
   ```go
   type SchemaStore interface {
       GetSchema(ctx context.Context, docType string) (string, error)
       SaveSchema(ctx context.Context, docType string, content string) error
       ListSchemas(ctx context.Context) ([]string, error)
       DeleteSchema(ctx context.Context, docType string) error
   }
   ```

2. **100% Storage Provider Coverage**:
   - **`localfs`**: Persists schema templates under `<rootDir>/schemas/<docType>.md` (resolving to `.gyrus/schemas/<docType>.md`).
   - **`blob` (AWS S3, Azure Blob, GCP GCS)**: Persists schemas under `.gyrus/schemas/<docType>.md` in cloud bucket containers.
   - **`git`**: Commits schema files directly to `.gyrus/schemas/<docType>.md` within the Git repository tree.
   - **`postgres`**: Persists schemas in a dedicated relational table:
     ```sql
     CREATE TABLE IF NOT EXISTS schemas (
         id TEXT PRIMARY KEY,
         content TEXT NOT NULL,
         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
     );
     ```

3. **Strict Path & Identifier Validation**:
   - Enforce document type regex validation (`^[a-z0-9-_]+$`) across all schema operations to prevent directory traversal attacks (`..`).
   - Validate that stored templates contain valid YAML frontmatter envelopes matching the target document type.

4. **Hierarchical Precedence & Explicit Failure Resolution**:
   When resolving a schema template for `<doc-type>`, the Lifecycle Engine applies strict two-tier resolution:
   1. Active Persistence Layer (`SchemaStore.GetSchema`)
   2. Pre-compiled Binary Embedded Templates (`okf.GetTemplate`)
   3. **Fail Explicitly**: If `<doc-type>` is not found in either persisted or embedded schemas, return an explicit `schema template not found for document type '<doc-type>'` error (Exit Code 5 in CLI; `isError: true` in MCP). Generic fallback scaffolding is eliminated to guarantee schema contract QA and prevent silent typo proliferation.

5. **Deprecation and Elimination of `schemas_path`**:
   The arbitrary `schemas_path` setting in `.gyrus.yaml` is obsolete and removed. Config files ignore the key without warning, and newly generated configurations omit it.

6. **CLI & MCP Tooling**:
   - CLI: `gyrus schema` offers explicit subcommands `get`, `set`, `list`, and `delete` with `--json` formatting, while preserving 100% backward compatibility for `gyrus schema <doc-type>`.
   - MCP: Exposes tools `gyrus_schema_get`, `gyrus_schema_set`, `gyrus_schema_list`, `gyrus_schema_delete`, and dynamic resources `memory://schema/{document_type}` and `memory://schemas`.

## Consequences

### Positive
- **Unified Remote Persistence**: All schemas live alongside documents in the active persistence backend.
- **Zero Schema Drift**: All engineering agents and developers referencing the same storage backend share identical schema validation contracts.
- **Full Backward Compatibility**: Legacy CLI commands and embedded templates continue to function seamlessly.
- **100% Provider Coverage**: Supported uniformly across LocalFS, S3, Azure Blob, GCS, Git, and PostgreSQL.

### Compliance & Verification
- Unit and integration tests cover CRUD operations across all four storage providers.
- Subcommand E2E tests verify CLI execution and MCP resource reading over Stdio.
