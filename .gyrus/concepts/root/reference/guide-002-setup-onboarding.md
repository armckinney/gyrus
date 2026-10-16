---
title: Team Onboarding & Context Setup Guide
id: guide-002-setup-onboarding
type: guide
format: markdown
category: technical
version: 2
status: active
last_updated: 2026-10-16T06:40:16Z
last_modified_by: ""
tenant: root
---

# Gyrus setup and OKF onboarding

Install the Gyrus binary, then initialize the repository:

```bash
gyrus config init --profile local --tenant root --project my-project
gyrus config show
```

The local profile uses localfs storage and OKF index/graph/search. There is no index database to initialize. Git, blob, s3, azure, and gcs profiles also default to OKF. PostgreSQL and vector profiles keep their database-backed compositions. Set the corresponding remote connection settings and use the existing credential discovery for remote storage.

## Create and connect concepts

```bash
gyrus create --id guide-onboarding --title "Onboarding" --type guide --category technical --tenant root --status active --content "Project setup instructions."
gyrus link guide-onboarding adr-storage --rel-type implements
gyrus search --query onboarding --json
gyrus search --query "" --scope all --json
gyrus sync --json
```

Targets may be unresolved. OKF links persist in source frontmatter, survive restarts, and are merged with legacy dependencies. Keyword search defaults to ten results; empty-query browsing returns all matching metadata. Sync validates source concepts; its indexed count means validated concepts, not database rows.

Concepts live under concepts/<tenant>/reference/ or concepts/<tenant>/projects/<project>/ below the storage root. Scope/project derive from paths. Schema templates live outside the collection. Do not put repository README files or schema templates inside the concept collection.

## Existing SQLite installations

Use the packaged gyrus-migrate-okf skill before changing secondary providers. Pause writers; inventory the complete SQLite graph and concept collection; make consistent database, concept, and config backups; export source-owned typed links; compare edge tuples and audit metadata; switch all three providers together; restart MCP processes. Keep the original database and backups. This first migration workflow does not export PostgreSQL or convert legacy blob layouts.

## Agent integrations

Install the canonical gyrus plugin through the client's supported plugin mechanism. It includes SDLC context guidance, OKF schema references, repository indexing guidance, and the SQLite-to-OKF migration skill. CLI/MCP tool names remain stable.

## Failures

Resolve invalid concepts or duplicate IDs before querying. Conflicting local versions, Git pushes, or cloud object preconditions must be reconciled before retrying. An unsupported conditional-write driver is an explicit error. Multi-source graph mutations report already-applied sources; do not assume a collection-wide transaction or blindly restore backups over later edits.
