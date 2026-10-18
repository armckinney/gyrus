---
title: Defer vector search to a future backend
id: adr-013-defer-vector-search
type: adr
format: ""
category: architecture
version: 1
status: accepted
last_updated: 2026-10-17T23:57:59Z
last_modified_by: ""
tenant: root
---

# Defer vector search to a future backend

## Context

The in-process vector prototype duplicates authoritative concept state and couples basic operations to embedding-service availability. Running MCP processes can return archived concepts from stale embedding snapshots. Keeping vectorization synchronized and operationally isolated requires a backend lifecycle that is outside the current lexical-retrieval scope.

## Decision

Remove vector search, embedding clients, the vector profile/configuration, associated tests, and the optional Ollama development sidecar. Keep OKF keyword search for file-backed storage and native PostgreSQL full-text search. Keep coding-agent integrations, PostgreSQL tsvector, and shared SQLite persistence for future cache work.

Reject legacy vector settings and unsupported profiles explicitly before provider initialization or configuration writes. Do not introduce compatibility aliases or automatic configuration rewriting. Preserve qualified concept identity, typed links, current keyword defaults, and explicit result-limit passthrough.

Semantic search and vectorization are deferred to a future backend design, including explicit ownership, freshness, operational failure isolation, and integration contracts. This replaces the vector-retention choice in ADR-012; its SQLite/sync retirement decision remains in effect.

## Consequences

Basic concept operations need no embedding service. Existing vector configurations must select OKF or PostgreSQL FTS and remove the vector section. Users retain their concepts unchanged. Future semantic-search work starts from a new design rather than retaining an unsupported runtime surface. Documentation, consumer checks, and configuration validation change alongside the implementation.
