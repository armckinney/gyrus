---
title: 'GYRUS-204: PostgreSQL Full-Text Search Engine'
id: spec-204-postgres-fts-search-driver
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
    - search-provider
    - postgres-fts
---

# PostgreSQL full-text search contract

The `postgres_fts` adapter implements `SearchProvider.Search(context.Context, SearchQuery)` for the complete PostgreSQL stack. Storage/index DDL creates generated `search_vector` values and a GIN index. Title, body, and tags contribute weighted English-language vectors.

Queries use PostgreSQL English text search and `ts_rank_cd` ranking, with metadata and qualified project/reference filters. The dictionary is fixed by the implementation, not a configurable profile option. Positive result limits cap matching results; keyword queries default to ten and metadata-only browsing is unlimited unless capped.

Storage, index, and search share the PostgreSQL query implementation. Tests cover filtering, limits, native FTS generation, and scan/decode failures; optional live database validation uses `GYRUS_TEST_POSTGRES_URL`. Vector/hybrid semantic search remains deferred.
