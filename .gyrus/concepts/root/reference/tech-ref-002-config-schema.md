---
title: Gyrus YAML Configuration Reference & Schema
id: tech-ref-002-config-schema
type: technical-reference
format: markdown
category: technical
version: 3
status: active
last_updated: 2026-10-16T18:59:08Z
last_modified_by: ""
tenant: root
---

# Gyrus configuration reference

Use `gyrus config init --profile local` to generate a project configuration, or add `--global` for the user-wide configuration. Profiles are local, git, blob, s3, azure, gcs, postgres, and vector. Inspect effective normalized provider selections with `gyrus config show`.

```yaml
tenant: root
project: my-project
storage:
  provider: localfs
  root: .gyrus
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
```

## Provider settings

| Setting | Choices and behavior |
| --- | --- |
| storage.provider | localfs (default), git, blob, s3, azure_blob, gcs, postgres |
| storage.root | local root relative to the configuration file; default .gyrus; blob stores use their configured prefix/root |
| index.provider | okf, sqlite, postgres; file-backed default okf |
| index.dsn | Optional SQLite index path; unused with OKF |
| graph.provider | okf, sqlite, postgres; omitted value follows effective index |
| search.provider | okf, sqlite/sqlite_fts5, postgres_fts, vector; omitted value follows effective index |
| tenant | Default tenant; root when omitted |
| project | Explicit project scope; paths determine stored concept scope |
| ui.client / ui.command | Existing UI agent selection and optional command |

Git uses git.repo_url and git.branch. Blob uses blob.bucket_url and optional blob.prefix. Cloud profiles use s3.bucket_name/region, azure_blob.storage_account/container_name, or gcs.bucket_name. PostgreSQL uses postgres.connection_string. Vector uses its embedding provider/model/endpoint settings. Credentials continue to use existing SDK/environment discovery.

Aliases: aws_s3 → s3; azure → azure_blob; gcp/gcp_gcs → gcs; search sqlite → sqlite_fts5.

## Defaults and constraints

File-backed profiles default to OKF index/graph/search. Explicit SQLite index selects SQLite graph and FTS5 search when omitted. Explicit PostgreSQL index selects PostgreSQL graph/FTS. Explicit vector search with omitted index selects SQLite. The PostgreSQL storage profile remains postgres/postgres/postgres_fts.

OKF index requires OKF graph/search. SQLite index supports SQLite or OKF graph and FTS5/vector/OKF search. PostgreSQL index supports PostgreSQL or OKF graph and PostgreSQL FTS/OKF search. All OKF components require localfs/Git/blob storage; PostgreSQL storage has no scanner in v1. Incompatible/unknown settings fail before I/O.

## OKF behavior

No database or DSN is required. Queries scan canonical concepts, reject invalid/duplicate records, and derive scope from paths. Empty queries browse all matching metadata; keyword queries return the top ten results. Search matches lowercase Unicode keyword substrings in ID/title/tags/body with OR semantics, exact metadata filters, project/reference prioritization, and deterministic ID ties. FTS operators/stemming are not supported.

Graph mutations persist typed frontmatter links alongside dependencies. Cooperating local processes use interprocess locks, Git pushes reject conflicts, and cloud writes/deletes require ETag/generation conditions. Unsupported conditional writes fail rather than overwriting unconditionally.

Before changing an existing SQLite configuration, run the packaged gyrus-migrate-sqlite-to-okf workflow to preserve database-only graph edges. Keep backups and the old database, verify tuple/audit parity, then switch all three fields and restart MCP processes.

## Configuration discovery

Explicit GYRUS_CONFIG and GYRUS_PROJECT selectors are supported by the loader. Otherwise project configuration is discovered from the working directory upward, then user ~/.gyrus.yaml, then built-in defaults. Project configuration replaces global configuration rather than merging it. Relative local storage roots resolve from their configuration file's directory. Removed lifecycle hook settings are not part of the active schema.
