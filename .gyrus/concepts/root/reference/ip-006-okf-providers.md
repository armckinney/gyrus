---
title: Direct-scan OKF providers and durable relationships
id: ip-006-okf-providers
type: improvement-proposal
format: markdown
category: architecture
version: 6
status: implemented
last_updated: 2026-10-16T16:47:49Z
last_modified_by: ""
tenant: root
links:
    - concept: root/reference/spec-004-core-engine
      rel: implements
---

# Direct-scan OKF providers

## Motivation
Make portable OKF concepts authoritative for metadata, graph relationships, and keyword retrieval without a secondary database.

## Approved design
Implement separate OKF index, graph, and search providers sharing storage scans. Persist typed links alongside legacy dependencies. Default file-backed profiles to OKF; return all metadata-only matches and ten keyword matches. Use interprocess file locking, Git push conflict detection, and cloud object preconditions. Migrate this repository's SQLite edges before cutover and ship a SQLite migration skill.

## Validation
Verify provider contracts, restart persistence, scan errors, deterministic search and graph behavior, concurrent writers, SQLite edge parity, CLI/MCP/UI consumers, and platform builds. Source code and tests are authoritative for implementation.
