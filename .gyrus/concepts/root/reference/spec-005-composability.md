---
title: Provider Composability, Test Profiles & Capability Matrix
id: spec-005-composability
type: specification
format: markdown
category: technical
version: 7
status: active
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

# Supported provider stacks

Gyrus supports three complete stacks, validated before resource construction:

| Profile | Storage | Index | Graph | Search |
| --- | --- | --- | --- | --- |
| local | localfs | okf | okf | okf |
| git | git | okf | okf | okf |
| postgres | postgres | postgres | postgres | postgres_fts |

Storage defaults to localfs. Omitted secondary providers follow storage. Hybrid combinations and cloud/blob providers are unsupported; removed names have no aliases. SQLite remains internal infrastructure reserved for future cache work, not a configurable index or cache provider.

The application constructs one storage instance and injects it into the stack. OKF adapters scan canonical local/Git concepts and source-owned links. PostgreSQL uses database-backed storage/index/graph and native FTS. Both expose qualified names, explicit project/reference scope, and consistent query limits. Existing-concept arguments require `tenant/reference/id` or `tenant/projects/project/id`.

Git sessions select `direct` or `branching`. Branching isolates all knowledge operations on a runtime initiative branch; database-backed hybrid indexes are excluded to prevent branch leakage. See the [Git specification](spec-201-git-storage-driver.md) and [standard configurations](tech-ref-002-config-schema.md).
