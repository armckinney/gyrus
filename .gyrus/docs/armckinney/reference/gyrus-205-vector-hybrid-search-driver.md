---
id: gyrus-205-vector-hybrid-search-driver
title: 'GYRUS-205: Semantic Vector Embedding & Hybrid Search Driver'
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
    - vector
    - ollama
    - openai
---

# GYRUS-205: Semantic Vector Embedding & Hybrid Search Driver

## 1. Overview & Objective
Implement a semantic vector embedding search provider under `internal/provider/vector` supporting hybrid lexical + vector search using Reciprocal Rank Fusion (RRF).

## 2. Requirements & Constraints
- Must implement `gyrus.SearchProvider` interface.
- Provide a pluggable `EmbeddingProvider` interface (`Embed(ctx, text) ([]float32, error)`).
- Implement built-in client for **Local Ollama** (e.g. `http://localhost:11434/api/embeddings` using `nomic-embed-text`) for zero-infra local execution.
- Implement built-in client for **OpenAI** (`text-embedding-3-small`).
- Use `pgvector` or local cosine similarity to score vector distances.
- Merge BM25 lexical search results with vector similarity ranks using Reciprocal Rank Fusion (RRF).

## 3. Key Test Verification
- Vector search test suite in `internal/provider/vector/vector_test.go`.
- Mock embedding provider tests verifying RRF hybrid rank fusion score calculations.
