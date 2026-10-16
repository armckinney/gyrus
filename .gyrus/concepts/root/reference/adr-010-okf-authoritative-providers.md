---
title: Authoritative OKF scans and durable typed links
id: adr-010-okf-authoritative-providers
type: adr
format: markdown
category: architecture
version: 4
status: accepted
last_updated: 2026-10-16T16:47:49Z
last_modified_by: ""
tenant: root
links:
    - concept: root/reference/ip-006-okf-providers
      rel: implements
---

# Authoritative OKF scans and durable typed links

## Context
Database-only graph mutations are not portable with stored concepts. Existing factories and profiles require databases even for file-backed collections.

## Decision
Provide independent OKF index, graph, and search adapters over a shared scanner-capable store. Persist source-owned typed links, retaining dependencies as depends_on edges. Queries rebuild ephemeral snapshots; no durable index or cache is introduced. Empty queries enumerate metadata, keyword queries default to ten results. Use explicit provider resolution and compatibility checks before I/O.

## Consistency
Lock local read-modify-write operations across processes, reject conflicting Git pushes, and use cloud generation/ETag conditions. Atomicity is per concept; cross-source batches report partial application. Missing targets remain valid, duplicate IDs and invalid concepts fail queries.

## Consequences
Collections remain portable and immediately reflect stored changes. Scans trade scalability for operational simplicity. Database providers remain explicit alternatives. SQLite edges must be exported before cutover; PostgreSQL migration is outside v1.
