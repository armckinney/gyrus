---
title: 'GYRUS-203: PostgreSQL Enterprise Index & Storage Driver'
id: spec-203-postgres-storage-index-driver
type: specification
format: ""
category: technical
version: 2
status: completed
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
tags:
    - phase-2.1
    - storage-driver
    - postgres
---

# PostgreSQL storage contract

The PostgreSQL profile uses the storage, index, graph, and FTS adapters together. Hybrid file-backed/database stacks are rejected before construction. A configured connection string opens the existing pgx-backed pools and initializes required tables/indexes.

Storage implements `ConceptStore` and `SchemaStore`. Tables include `concepts` (JSONB frontmatter and content), `concept_edges`, `concepts_history`, and `schemas`. Qualified names identify concepts and relationships. Positive expected versions are checked under row locks; stale writes return typed conflicts and roll back without history writes.

The metadata index and graph use PostgreSQL; search uses [native FTS](spec-204-postgres-fts-search-driver.md). Runtime Git branch selection does not apply to this stack. Distributed locking/queues and hosted deployment remain future work.

Provider mocks cover CRUD, version conflicts, schemas, relationships, and queries. `TestIntegration_PostgresLive` is opt-in through `GYRUS_TEST_POSTGRES_URL` pointing to an isolated database. See [configuration examples](tech-ref-002-config-schema.md).
