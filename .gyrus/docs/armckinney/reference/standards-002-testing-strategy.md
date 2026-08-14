---
id: standards-002-testing-strategy
title: Gyrus Testing Strategy & QA Standards
category: technical
type: standards
format: markdown
owner_group: armckinney
version: 1
status: active
tags:
  - testing
  - qa
  - standards
  - integration
  - mcp
  - cli
  - agents
  - phase-2.4
dependencies:
  - prd-001-specification-roadmap
  - standards-001-doc-strategy
  - spec-001-contract-schema
  - spec-002-cli-interface
  - spec-003-mcp-interface
  - spec-004-core-engine
---

# Gyrus Testing Strategy & QA Standards

> **Testing Methodology & Scope**:
> This document establishes the single source of truth for the **Integration, End-to-End (E2E), Protocol, and Product Quality Assurance (QA) standards** for Gyrus. Because Gyrus is developed agentically and serves as foundational context and memory infrastructure for AI agents, this suite guarantees full-system capability, agent interoperability, and long-term stability.
> 
> *Note on Unit Tests*: These product and integration test suites do **not** replace package-level unit tests. Granular unit tests scattered throughout `internal/` packages continue to be maintained and executed in conjunction with this suite for rapid local development and isolated logic verification.
> 
> *(Future roadmap features such as `gyrus init config`/`client` subcommands, Agent Plugin Packaging `plugin.json`, `~/.gyrus.yaml` global config, `gyrus ui`, and remote schema storage have their target test requirements tracked in `prd-001-specification-roadmap.md` and are implemented during their respective releases).*

---

## 1. Domain & Contract Governance (`internal/domain/{okf,lifecycle}`, `internal/format`)
*Contract validation tests verifying frontmatter schemas, state machines, immutability, and optimistic concurrency locks.*

- [ ] **Schema & Frontmatter Validation**:
  - Valid OKF YAML frontmatter parsing and attribute normalization.
  - Enforcement of document ID lower-case alphanumeric pattern (`^[a-z0-9-_]+$`).
  - Mandatory frontmatter fields check (`id`, `title`, `category`, `type`, `owner_group`, `status`).
  - Rejection of unknown document types or invalid categories.
- [ ] **State Machine & Lifecycle Transitions**:
  - Valid lifecycle paths (`draft` ➔ `proposed` ➔ `accepted`/`active` ➔ `deprecated`/`superseded`/`archived`).
  - Strict blocking of illegal transitions (e.g. `accepted` ➔ `proposed`, `archived` ➔ `draft`) returning **Exit Code 2**.
- [ ] **Immutability Enforcement**:
  - Immutability of accepted ADRs, accepted improvement proposals, and custom templates with `immutable: true`.
  - Mutation blocking ensuring immutable documents cannot be updated without creating superseding documents.
- [ ] **Optimistic Concurrency Control**:
  - Document version auto-incrementing on updates (`version` integer).
  - Version collision detection when `--expected-version` mismatches current storage version (**Exit Code 4**).

---

## 2. Infrastructure-Backed Multi-Provider Suite (`internal/provider/{storage,index,search,graph}`)
*Integration tests verifying real backing infrastructure across storage, search engines, vector embeddings, and graph databases.*

- [ ] **Real Infrastructure Storage Drivers (`storage`)**:
  - `localfs`: Full filesystem CRUD, nested directory auto-creation (`.gyrus/docs/...`), atomic writes, file permissions.
  - `git`: `go-git` in-memory and local repository commit creation, remote sync, clean working tree assertions, branch isolation.
  - `postgres`: Real PostgreSQL backend testing via Docker container (`pgx/v5`), transaction isolation, reconnection pooling.
  - `blob`: Cloud Object Storage (S3, Azure Blob, GCS) via `gocloud.dev/blob` mock server, chunked writes, key prefixing.
- [ ] **Real Infrastructure Search & Index Drivers (`search`, `index`)**:
  - `sqlite` / `fts5`: FTS5 full-text indexing, BM25 relevance ranking, special character sanitization (`"`, `'`, `-`, `*`), full re-index on `gyrus sync`.
  - `postgres_fts`: `tsvector` generation, `tsquery` execution, relevance scoring.
  - `vector`: Semantic vector embeddings (Local Ollama sidecar, OpenAI mock), Reciprocal Rank Fusion (RRF) score blending, distance thresholding.
- [ ] **Real Infrastructure Graph Drivers (`graph`)**:
  - Directed edge creation (`depends_on`, `supersedes`, `implements`, `mitigates`) in `sqlite` & `postgres`.
  - Circular dependency detection and resolution.
  - Cascade unlink and orphaned link cleanup on document archiving.

---

## 3. CLI Subcommand Suite & Exit Code Protocol (`internal/cli`, `cmd/gyrus`)
*End-to-end CLI binary tests validating command outputs, flag behaviors, structured JSON envelopes, and exit code contracts.*

- [ ] **Subcommand Verification**:
  - `gyrus suggest-context`: Prompt relevance ranking, `--max-tokens` budgeting & line truncation, empty prompt handling.
  - `gyrus search`: Keyword search filtering (`--category`, `--type`, `--status`, `--tag`, `--max-results`), structured `--json` output envelopes.
  - CRUD & Topology: `get`, `create`, `update`, `link`, `unlink`, `archive`, `sync`, `validate`, `schema`.
  - Installation & Setup: `gyrus init` (verifying profile matrices `--profile local|git|blob|postgres|vector`, `--owner-group`, `--mcp-target`, `--mcp-mode`, `--skill-target`).
- [ ] **Exit Code Protocol Enforcement**:
  - `0`: Success
  - `1`: Frontmatter schema / validation / ID error
  - `2`: Illegal State Machine transition error
  - `3`: Unauthorized / owner group permission error
  - `4`: Optimistic concurrency lock conflict (`--expected-version` mismatch)
  - `5`: File/Record not found or storage read/write error
- [ ] **Stdio Stream Isolation**:
  - Guarantee that logs, warnings, and informational banners route strictly to `stderr`, keeping `stdout` completely clean for JSON/JSON-RPC consumers.

---

## 4. MCP Transport & Server Integration (`internal/mcp`)
*Protocol and transport tests verifying JSON-RPC 2.0 message handling, stdio streaming, and MCP tool/resource interfaces.*

- [ ] **MCP Serving from CLI (`gyrus mcp serve`)**:
  - Stdio JSON-RPC 2.0 message parsing, initialization handshake, request/response encoding over stdin/stdout.
- [ ] **MCP Container Serving (`docker run -i ghcr.io/armckinney/gyrus:latest`)**:
  - Running containerized MCP server in interactive stdio mode, verifying stdio pipe persistence, signal handling (SIGTERM, SIGINT).
- [ ] **MCP Tool Suite Integration**:
  - Complete verification of all 10 implemented MCP tools (`suggest_context`, `search`, `get`, `create`, `update`, `link`, `unlink`, `archive`, `sync`, `validate`).
- [ ] **MCP Resources & Prompts**:
  - Resource fetching (`gyrus://document/{id}`, `gyrus://schemas`) and prompt template rendering (`gyrus-context-prompt`).

---

## 5. Installation, Workspace Setup & E2E Agent Integration (`internal/setup`, `tests/skills/`)
*End-to-end agent product tests verifying workspace provisioning, skill distribution, and autonomous agent tool usage.*

- [ ] **Installation & Workspace Setup (`gyrus init`)**:
  - CLI binary execution & installation verification.
  - Workspace setup generating `.gyrus.yaml` configuration.
  - Skill/Plugin equipping across target agent tools (`agy cli`, `copilot & codex cli`, `claude code` target paths).
- [ ] **End-to-End Agent Integration Harness**:
  - E2E integration test harness (`tests/skills/agent_eval_test.go` and `e2e_antigravity_test.go`) verifying that target AI agents (`agy cli`, `copilot cli`, `claude code`) properly discover, invoke, and consume Gyrus CLI subcommands and MCP tools to resolve engineering context tasks.
- [ ] **Agent Skill Specification Conformance**:
  - Automated schema and frontmatter linting for `.agents/skills/gyrus/SKILL.md` to ensure zero discovery warnings across all agent loaders.

---

## 6. Context Linearization & Snapshot Regression Testing
*Product quality assurance tests ensuring context generation outputs remain deterministic, clean, and token-accurate for LLMs.*

- [ ] **Suggest-Context Golden Snapshot Tests**:
  - Regression testing across diverse multi-document repositories ensuring linearized Markdown context matches golden snapshot fixtures.
- [ ] **Token Budgeting & Precision Truncation**:
  - Strict validation that `--max-tokens` budgets correctly truncate document bodies without corrupting frontmatter boundaries or truncating mid-character.

---

## 7. Reliability, Concurrency & Security Hardening
*System stability tests verifying data race safety, crash recovery, path sanitization, and scale performance under heavy load.*

- [ ] **Data Race Detector (`-race`)**:
  - Parallel Go test suite execution verifying zero data races under concurrent CLI/MCP operations.
- [ ] **Crash Consistency & Transaction Rollback**:
  - Verifying that interrupted sync operations or database write failures rollback cleanly without corrupting the SQLite index or leaving orphan graph links.
- [ ] **Database Migration & Upgrade Compatibility**:
  - Backward compatibility tests ensuring existing SQLite databases migrate cleanly across version upgrades without data loss.
- [ ] **Path Traversal Security**:
  - Rejection of directory traversal attacks (e.g. `../../etc/passwd` injection attempts).
- [ ] **Credential Masking**:
  - Guarantee that storage secrets (S3 keys, DB passwords) are never leaked in logs or `--json` outputs.
- [ ] **Scalability & Load Benchmarks**:
  - Benchmark test harness scaling to 10,000+ OKF documents, asserting indexing time < 2s and query latency < 50ms.
