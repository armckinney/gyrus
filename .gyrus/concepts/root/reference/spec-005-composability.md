---
title: Provider Composability, Test Profiles & Capability Matrix
id: spec-005-composability
type: specification
format: markdown
category: technical
version: 3
status: active
last_updated: 2026-10-16T16:57:36Z
last_modified_by: ""
tenant: root
---

# Provider composition and profiles

Provider capability is defined by the source factories and config resolver, not historical profile matrices.

## Implemented providers

| Role | Providers |
| --- | --- |
| Storage | localfs, git, blob, s3, azure_blob, gcs, postgres |
| Index | okf, sqlite, postgres |
| Graph | okf, sqlite, postgres |
| Search | okf, sqlite_fts5 (sqlite alias), postgres_fts, vector |

Storage aliases: aws_s3 maps to s3, azure to azure_blob, gcp/gcp_gcs to gcs. OKF requires scanner-capable storage: localfs, Git, or blob/cloud aliases. PostgreSQL storage remains database-backed.

## Generated profiles

| Profile | Storage | Index | Graph | Search |
| --- | --- | --- | --- | --- |
| local | localfs | okf | okf | okf |
| git | git | okf | okf | okf |
| blob | blob | okf | okf | okf |
| s3 | s3 | okf | okf | okf |
| azure | azure_blob | okf | okf | okf |
| gcs | gcs | okf | okf | okf |
| postgres | postgres | postgres | postgres | postgres_fts |
| vector | localfs | sqlite | sqlite | vector |

## Omitted fields and compatible combinations

Omitted storage uses localfs. File-backed storage defaults index to okf. Graph defaults to the effective index. Search defaults according to index: okf, sqlite_fts5, or postgres_fts. Explicit vector search with omitted index chooses SQLite. PostgreSQL storage defaults all secondary providers to PostgreSQL.

An OKF index requires OKF graph/search. SQLite indexes support SQLite or OKF graph, and SQLite FTS5/vector/OKF search. PostgreSQL indexes support PostgreSQL or OKF graph and PostgreSQL FTS/OKF search, subject to scanner-capable storage for every OKF component. PostgreSQL storage requires postgres/postgres/postgres_fts. Unknown or incompatible providers fail before resource creation.

## Retrieval and portability

OKF uses direct collection scans without persistent caches/indexes. Keyword search defaults to ten results; metadata-only browsing is uncapped unless the SDK request supplies a limit. SQLite and PostgreSQL retain their own retrieval semantics. Scope/project derive from canonical concept paths.

Typed source-owned links make graph relationships portable. Before switching an existing SQLite graph, use the packaged gyrus-migrate-okf skill to back up, export, and verify every database-only edge. PostgreSQL migration and legacy blob-layout conversion are outside v1.

## Qualified concepts and minimal links

Creation retains a local id and returns its qualified name. Address existing concepts with tenant/reference/id or tenant/projects/project/id across SDK, CLI, MCP, and UI. API envelopes include name; YAML frontmatter does not persist it. Relationships use only links with concept and rel. Dependency arrays, old link keys, per-link creation metadata, and short lookup IDs are rejected. Local IDs may repeat across tenants/projects. Incoming neighbors provide backlinks; relationship types and general cycle-safe graph traversal remain unchanged. Use the migration skill for legacy concepts or SQLite graph exports; retain backups and rebuild obsolete identity databases explicitly. Restart cached MCP processes after upgrading.
