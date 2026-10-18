---
title: Provider Composability, Test Profiles & Capability Matrix
id: spec-005-composability
type: specification
format: markdown
category: technical
version: 6
status: active
last_updated: 2026-10-17T23:57:58Z
last_modified_by: ""
tenant: root
---

# Provider composition and profiles

Provider capability is defined by the source factories and config resolver, not historical profile matrices.

## Implemented providers

| Role | Providers |
| --- | --- |
| Storage | localfs, git, blob, s3, azure_blob, gcs, postgres |
| Index | okf, postgres |
| Graph | okf, postgres |
| Search | okf, postgres_fts |

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

## Omitted fields and compatible combinations

Omitted storage uses localfs. File-backed storage defaults index to okf. Graph defaults to the effective index. Search defaults according to index: okf or postgres_fts. PostgreSQL storage defaults all secondary providers to PostgreSQL. Unsupported profiles return an error before configuration writes. Provider names use shared validation; unrecognized YAML sections follow standard decoding behavior.

An OKF index requires OKF graph/search. PostgreSQL indexes support PostgreSQL or OKF graph and PostgreSQL FTS/OKF search, subject to scanner-capable storage for every OKF component. PostgreSQL storage requires postgres/postgres/postgres_fts. Unknown or incompatible providers fail before resource creation. Shared SQLite persistence remains internal infrastructure for future cache work.

## Retrieval and portability

OKF uses direct collection scans without persistent caches/indexes. Keyword search defaults to ten results; positive SDK MaxResults, CLI --max-results, and MCP max_results values override that limit. Metadata-only browsing is uncapped unless a positive limit is supplied. PostgreSQL uses the same limits and scope/project filter contract with native full-text matching. Scope/project derive from canonical concept identity. The sync command/tool has been removed.

Typed source-owned links make graph relationships portable. Before switching an existing SQLite graph, use the packaged gyrus-migrate-sqlite-to-okf skill to back up, export, and verify every database-only edge. PostgreSQL migration and legacy blob-layout conversion are outside v1.

## Qualified concepts and minimal links

Creation retains a local id and returns its qualified name. Address existing concepts with tenant/reference/id or tenant/projects/project/id across SDK, CLI, MCP, and UI. API envelopes include name; YAML frontmatter does not persist it. Relationships use only links with concept and rel. Dependency arrays, old link keys, per-link creation metadata, and short lookup IDs are rejected. Local IDs may repeat across tenants/projects. Incoming neighbors provide backlinks; relationship types and general cycle-safe graph traversal remain unchanged. Use the migration skill for legacy concepts or SQLite graph exports; retain backups and rebuild obsolete identity databases explicitly. Restart cached MCP processes after upgrading.
