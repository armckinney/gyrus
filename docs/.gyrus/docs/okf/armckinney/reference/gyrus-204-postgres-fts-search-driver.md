---
id: gyrus-204-postgres-fts-search-driver
title: 'GYRUS-204: PostgreSQL Full-Text Search Engine'
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: completed
last_modified_by: ""
last_updated: 2026-07-23T21:23:12Z
tags:
    - phase-2.1
    - search-provider
    - postgres-fts
---

# GYRUS-204: PostgreSQL Full-Text Search Engine

## 1. Overview & Objective
Implement a native PostgreSQL `SearchProvider` under `internal/provider/postgres` utilizing PostgreSQL `tsvector` and `tsquery` full-text search capabilities.

## 2. Requirements & Constraints
- Must implement `gyrus.SearchProvider` interface (`Search(ctx, query, filter)`).
- Must construct GIN indexes over `tsvector` generated columns combining document title, content, and frontmatter tags.
- Configurable text search dictionaries (defaulting to `'english'`) and relevance ranking using `ts_rank_cd()`.

## 3. Key Test Verification
- Search test suite in `internal/provider/postgres/fts_test.go`.
- Verification of keyword search relevance scoring and filtering.
