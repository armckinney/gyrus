---
title: Team Onboarding & Context Setup Guide
id: guide-002-setup-onboarding
type: guide
format: markdown
category: technical
version: 4
status: active
last_updated: 2026-10-16T19:41:57Z
last_modified_by: ""
tenant: root
---

# Gyrus setup and OKF onboarding

Install the Gyrus binary, then initialize the repository:

```bash
gyrus config init --profile local --tenant root --project my-project
gyrus config show
```

The local profile uses localfs storage and OKF index/graph/search. There is no index database to initialize. Git, blob, s3, azure, and gcs profiles also default to OKF. PostgreSQL remains database-backed. Vector uses OKF index/graph and lexical fallback with its configured embedding service. Set the corresponding remote connection settings and use the existing credential discovery for remote storage.

## Create and connect concepts

```bash
gyrus create --id guide-onboarding --title "Onboarding" --type guide --category technical --tenant root --status active --content "Project setup instructions."
gyrus link root/projects/my-project/guide-onboarding root/reference/adr-storage --rel-type implements
gyrus search --query onboarding --json
gyrus search --query "" --scope all --json
```

Targets may be unresolved. OKF links persist in source frontmatter and survive restarts. Obsolete dependency fields require explicit conversion. Keyword search defaults to ten results; --max-results overrides that default and empty-query browsing returns all matching metadata unless capped. There is no sync command/tool; OKF queries read the collection directly. Use gyrus validate on individual files.

Concepts live under concepts/<tenant>/reference/ or concepts/<tenant>/projects/<project>/ below the storage root. Scope/project derive from paths. Schema templates live outside the collection. Do not put repository README files or schema templates inside the concept collection.

## Existing SQLite installations

Use the packaged gyrus-migrate-sqlite-to-okf skill before upgrading SQLite-backed configurations; the runtime no longer provides SQLite index/graph/search. Pause writers; inventory the complete graph and concept collection; make consistent database, concept, and config backups; export source-owned typed links; verify normalized edge tuples and preserved bodies/statuses; switch all three providers to OKF; restart MCP processes. Keep the original database and backups. This workflow does not export PostgreSQL or convert legacy blob layouts.

## Agent integrations

Install the canonical gyrus plugin through the client's supported plugin mechanism. It includes SDLC context guidance, OKF schema references, repository indexing guidance, and the SQLite-to-OKF migration skill. CLI/MCP tool names remain stable.

## Failures

Resolve invalid concepts or duplicate IDs before querying. Conflicting local versions, Git pushes, or cloud object preconditions must be reconciled before retrying. An unsupported conditional-write driver is an explicit error. Multi-source graph mutations report already-applied sources; do not assume a collection-wide transaction or blindly restore backups over later edits.
