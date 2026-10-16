---
title: Retire SQLite providers and explicit synchronization
id: adr-012-retire-sqlite-providers-and-sync
type: adr
format: ""
category: architecture
version: 1
status: accepted
last_updated: 2026-10-16T19:41:58Z
last_modified_by: ""
tenant: root
---

# Retire SQLite providers and explicit synchronization

## Context

File-backed profiles now use authoritative OKF scans and source-owned typed links. Keeping SQLite as a second index/graph/search system introduces duplicated state and requires explicit synchronization. SQLite graph unlink can be reversed by reindexing links that remain in Markdown. Search limits and published MCP templates also need a consistent public contract.

## Decision

Remove SQLite index, graph, and search providers, their aliases, and CLI/MCP sync. Keep internal/persistence/db SQLite support as infrastructure for a future cache provider; no cache is implemented by this change. Preserve PostgreSQL index/graph/search and move the vector profile to OKF index/graph and lexical fallback.

Use SearchQuery across search providers so CLI --max-results, MCP max_results, and SDK MaxResults reach retrieval. Keyword requests default to ten; empty-query browsing is unlimited unless positively capped. PostgreSQL storage/index/search share one implementation with scope/project filtering. Context expansion reads the selected graph, including database-only dependencies.

Render prepare-adr from the active shared ADR schema, preserving custom template content and canonical frontmatter order. Qualified concept names and minimal concept/rel links remain unchanged.

## Consequences

Legacy SQLite configurations fail explicitly. Export existing edges with the gyrus-migrate-sqlite-to-okf skill before selecting OKF providers; retain databases and backups. There is no public reindex command. Validate files individually and use uncapped empty-query search to validate the collection. Live PostgreSQL integration is opt-in through GYRUS_TEST_POSTGRES_URL targeting an isolated test database.
