---
title: Deferred semantic search backend
id: spec-205-vector-hybrid-search-driver
type: specification
format: markdown
category: technical
version: 3
status: deprecated
last_updated: 2026-10-18T00:00:28Z
last_modified_by: ""
tenant: root
tags:
    - phase-2.1
    - search-provider
    - vector
    - ollama
    - openai
---

# Deferred semantic search backend

The in-process vector/hybrid search prototype has been retired. This deprecated specification is retained as context for future backend planning, not as a description of a supported Gyrus provider.

## Current implementation

Supported search providers are OKF keyword scanning and PostgreSQL full-text search. PostgreSQL tsvector remains part of its native full-text implementation. There is no runtime vector contract or compatibility handling; configuration uses shared provider/profile validation and standard YAML decoding.

## Future work

Semantic retrieval, vectorization, embedding lifecycle, and service integration are deferred to a future backend design. That work must define ownership, freshness, failure isolation, and integration contracts before implementation. No embedding API, model default, or vector configuration is promised by this specification.

See the accepted decision root/reference/adr-013-defer-vector-search for the removal rationale. Historical ADR bodies are retained unchanged.
