---
id: standards-002-testing-strategy
title: Gyrus Testing Strategy & QA Standards
category: technical
type: standards
format: markdown
owner_group: armckinney
version: 4
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

## 🏛️ Test Harness Architecture & Layers

```mermaid
graph TD
    subgraph "Layer 1: Unit & Package Tests (In-Memory, Microsecond Feedback)"
        U1["Domain & OKF Parser Unit Tests (internal/domain/okf)"]
        U2["Lifecycle State Machine Unit Tests (internal/domain/lifecycle)"]
        U3["Format Serialization Unit Tests (internal/format)"]
        U4["Provider Algorithmic Logic (BM25, RRF Math)"]
    end

    subgraph "Layer 2: Provider Infrastructure Matrix (Integration)"
        P1["Storage: LocalFS, Git, Postgres, Blob Mock"]
        P2["Index & Search: SQLite FTS5, Postgres FTS, Vector Embeddings"]
        P3["Graph: SQLite & Postgres Directed Edge Stores"]
    end

    subgraph "Layer 3: CLI Binary & Protocol Contract (Product E2E)"
        C1["Subcommands: suggest-context, search, get, create, update, link, archive, sync, init"]
        C2["Standardized Exit Code Protocol: 0, 1, 2, 3, 4, 5"]
        C3["Stdio Stream Isolation: stderr logs / stdout pure JSON"]
    end

    subgraph "Layer 4: MCP Protocol & Transport (JSON-RPC 2.0 Process)"
        M1["Stdio Subprocess Handshake & Request/Response"]
        M2["10 MCP Tools Verification over JSON-RPC 2.0"]
        M3["MCP Resources & Prompt Templates"]
        M4["Docker Container Interactive MCP Serving"]
    end

    subgraph "Layer 5: Product QA, Agent Workflows & Hardening"
        Q1["Deterministic Snapshot Linearization (Golden Markdown Fixtures)"]
        Q2["Token Budgeting & Precision Truncation (--max-tokens)"]
        Q3["Agent Skills & Autonomous AI Agent E2E Scenarios (agy, copilot, claude)"]
        Q4["Race Detection (-race), Crash Consistency & 10k Scale Benchmark"]
    end

    Layer 1 --> Layer 2
    Layer 2 --> Layer 3
    Layer 3 --> Layer 4
    Layer 4 --> Layer 5
```

---

## ⚖️ Unit Tests vs. Product E2E Tests

To ensure the system functions properly as a user and AI agent expect, Gyrus maintains a strict separation of concerns between **Unit Tests** and **Product / E2E Integration Tests**:

| Dimension | Unit Tests (In-Code / Package Level) | Product / E2E Tests (Black Box / Manual User Sim) |
| :--- | :--- | :--- |
| **Execution Surface** | Internal Go packages (`internal/**_test.go`) | Compiled `./gyrus` binary executable and MCP subprocess (`tests/e2e/`, `tests/cli/`, `tests/mcp/`) |
| **Primary Goal** | Validate isolated logic, mathematical formulas, regexes, and data transformations. | Validate the product as a human developer or AI agent interacts with it manually. |
| **I/O & Environment** | In-memory mocks, zero external process overhead, microsecond runtimes. | Real filesystem, real CLI flags, real process exit codes, real stdio JSON-RPC streaming. |
| **Example Scenarios** | - YAML unmarshaling edge cases<br>- Document ID regex validation (`^[a-z0-9-_]+$`)<br>- State machine transition truth tables<br>- BM25 score calculation | - `gyrus init --profile postgres` creates `.gyrus.yaml` and `.agents/skills/`<br>- `gyrus update adr-001 --status proposed` exits with code `2`<br>- Spawning `gyrus mcp serve` and calling `suggest_context` over JSON-RPC stdin/stdout<br>- Multi-step onboarding workflow from init ➔ create ➔ link ➔ suggest-context |

---

## 📝 In-Code Test Documentation Standard

To ensure tests remain immediately comprehensible to human contributors and AI pair programmers, **every test function in this repository must include a structured header docstring**:

```go
// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test | Provider Integration | Unit Test
// [Purpose]: Plain English explanation of what contract or user journey is tested.
// [Execution Surface]: Compiled ./gyrus CLI binary | Stdio MCP Subprocess | In-Memory Package
// [Assertions]: Explicit list of verified outcomes (e.g. Exit Code 2, pure JSON on stdout).
// -----------------------------------------------------------------------------
func TestExample(t *testing.T) { ... }
```

---

## ☁️ Cloud Infrastructure & Integration Authentication Standards

Integration tests targeting external cloud and database providers follow standardized authentication and emulation patterns:

1. **Local Sidecars & Emulators (Default / Zero-Config CI)**:
   - **PostgreSQL**: Tested against local/containerized PostgreSQL (`localhost:5432` / Docker service `postgres`).
   - **Ollama / Vector**: Tested against local DevContainer Ollama sidecar (`http://ollama:11434` or local instance).
   - **Cloud Blob Storage (S3 / GCS / Azure)**: Default integration tests run against `gocloud.dev/blob/fileblob` or local S3-compatible emulators (e.g. MinIO/Azurite).
2. **Real Cloud Provider Integration (Opt-In / Environment Flags)**:
   - Real cloud driver tests activate only when provider-specific credentials are provided in the environment (e.g. `AWS_ACCESS_KEY_ID`, `AZURE_STORAGE_KEY`, `GOOGLE_APPLICATION_CREDENTIALS`), automatically skipping when credentials are not configured.

---

## 1. Domain & Contract Governance (`internal/domain/{okf,lifecycle}`, `internal/format`)
*Contract validation tests verifying frontmatter schemas, state machines, immutability, and optimistic concurrency locks.*

- [x] **Schema & Frontmatter Validation**:
  - Valid OKF YAML frontmatter parsing and attribute normalization.
  - Enforcement of document ID lower-case alphanumeric pattern (`^[a-z0-9-_]+$`).
  - Mandatory frontmatter fields check (`id`, `title`, `category`, `type`, `owner_group`, `status`).
  - Rejection of unknown document types or invalid categories.
- [x] **State Machine & Lifecycle Transitions**:
  - Valid lifecycle paths (`draft` ➔ `proposed` ➔ `accepted`/`active` ➔ `deprecated`/`superseded`/`archived`).
  - Strict blocking of illegal transitions (e.g. `accepted` ➔ `proposed`, `archived` ➔ `draft`) returning **Exit Code 2**.
- [x] **Immutability Enforcement**:
  - Immutability of accepted ADRs, accepted improvement proposals, and custom templates with `immutable: true`.
  - Mutation blocking ensuring immutable documents cannot be updated without creating superseding documents.
- [x] **Optimistic Concurrency Control**:
  - Document version auto-incrementing on updates (`version` integer).
  - Version collision detection when `--expected-version` mismatches current storage version (**Exit Code 4**).

---

## 2. Infrastructure-Backed Multi-Provider Suite (`internal/provider/{storage,index,search,graph}`)
*Integration tests verifying real backing infrastructure across storage, search engines, vector embeddings, and graph databases.*

- [x] **Real Infrastructure Storage Drivers (`storage`)**:
  - `localfs`: Full filesystem CRUD, nested directory auto-creation (`.gyrus/docs/...`), atomic writes, file permissions.
  - `git`: `go-git` in-memory and local repository commit creation, remote sync, clean working tree assertions, branch isolation.
  - `postgres`: Real PostgreSQL backend testing via Docker container (`pgx/v5`), transaction isolation, reconnection pooling.
  - `blob`: Cloud Object Storage (S3, Azure Blob, GCS) via `gocloud.dev/blob` mock server, chunked writes, key prefixing.
- [x] **Real Infrastructure Search & Index Drivers (`search`, `index`)**:
  - `sqlite` / `fts5`: FTS5 full-text indexing, BM25 relevance ranking, special character sanitization (`"`, `'`, `-`, `*`), full re-index on `gyrus sync`.
  - `postgres_fts`: `tsvector` generation, `tsquery` execution, relevance scoring.
  - `vector`: Semantic vector embeddings (Local Ollama sidecar, OpenAI mock), Reciprocal Rank Fusion (RRF) score blending, distance thresholding.
- [x] **Real Infrastructure Graph Drivers (`graph`)**:
  - Directed edge creation (`depends_on`, `supersedes`, `implements`, `mitigates`) in `sqlite` & `postgres`.
  - Circular dependency detection and resolution.
  - Cascade unlink and orphaned link cleanup on document archiving.

---

## 3. CLI Subcommand Suite & Exit Code Protocol (`internal/cli`, `cmd/gyrus`)
*End-to-end CLI binary tests validating command outputs, flag behaviors, structured JSON envelopes, and exit code contracts.*

- [x] **Subcommand Verification**:
  - `gyrus suggest-context`: Prompt relevance ranking, `--max-tokens` budgeting & line truncation, empty prompt handling.
  - `gyrus search`: Keyword search filtering (`--category`, `--type`, `--status`, `--tag`, `--max-results`), structured `--json` output envelopes.
  - CRUD & Topology: `get`, `create`, `update`, `link`, `unlink`, `archive`, `sync`, `validate`, `schema`.
  - Installation & Setup: `gyrus init` (verifying profile matrices `--profile local|git|blob|postgres|vector`, `--owner-group`, `--mcp-target`, `--mcp-mode`, `--skill-target`).
- [x] **Exit Code Protocol Enforcement**:
  - `0`: Success
  - `1`: Frontmatter schema / validation / ID error
  - `2`: Illegal State Machine transition error
  - `3`: Unauthorized / owner group permission error
  - `4`: Optimistic concurrency lock conflict (`--expected-version` mismatch)
  - `5`: File/Record not found or storage read/write error
- [x] **Stdio Stream Isolation**:
  - Guarantee that logs, warnings, and informational banners route strictly to `stderr`, keeping `stdout` completely clean for JSON/JSON-RPC consumers.

---

## 4. MCP Transport & Server Integration (`internal/mcp`)
*Protocol and transport tests verifying JSON-RPC 2.0 message handling, stdio streaming, and MCP tool/resource interfaces.*

- [x] **MCP Serving from CLI (`gyrus mcp serve`)**:
  - Stdio JSON-RPC 2.0 message parsing, initialization handshake, request/response encoding over stdin/stdout.
- [x] **MCP Container Serving (`docker run -i ghcr.io/armckinney/gyrus:latest`)**:
  - Running containerized MCP server in interactive stdio mode, verifying stdio pipe persistence, signal handling (SIGTERM, SIGINT).
- [x] **MCP Tool Suite Integration**:
  - Complete verification of all 11 implemented MCP tools (`gyrus_create`, `gyrus_get`, `gyrus_update`, `gyrus_archive`, `gyrus_search`, `gyrus_suggest_context`, `gyrus_link`, `gyrus_unlink`, `gyrus_neighbors`, `gyrus_traverse`, `gyrus_sync`).
- [x] **MCP Resources & Prompts**:
  - Resource fetching (`gyrus://documents`, `gyrus://schemas`) and prompt template rendering (`architect_review`, etc.).

---

## 5. Installation, Workspace Setup & E2E Agent Integration (`internal/setup`, `tests/skills/`)
*End-to-end agent product tests verifying workspace provisioning, skill distribution, and autonomous agent tool usage.*

- [x] **Installation & Workspace Setup (`gyrus init`)**:
  - CLI binary execution & installation verification.
  - Workspace setup generating `.gyrus.yaml` configuration.
  - Skill/Plugin equipping across target agent tools (`agy cli`, `copilot & codex cli`, `claude code` target paths).
- [x] **End-to-End Agent Integration Harness**:
  - E2E integration test harness (`tests/skills/agent_eval_test.go` and `e2e_antigravity_test.go`) verifying that target AI agents (`agy cli`, `copilot cli`, `claude code`) properly discover, invoke, and consume Gyrus CLI subcommands and MCP tools to resolve engineering context tasks.
- [x] **Agent Skill Specification Conformance**:
  - Automated schema and frontmatter linting for `.agents/skills/gyrus/SKILL.md` to ensure zero discovery warnings across all agent loaders.

---

## 6. Context Linearization & Snapshot Regression Testing
*Product quality assurance tests ensuring context generation outputs remain deterministic, clean, and token-accurate for LLMs.*

- [x] **Suggest-Context Golden Snapshot Tests**:
  - Regression testing across diverse multi-document repositories ensuring linearized Markdown context matches golden snapshot fixtures.
- [x] **Token Budgeting & Precision Truncation**:
  - Strict validation that `--max-tokens` budgets correctly truncate document bodies without corrupting frontmatter boundaries or truncating mid-character.

---

## 7. Reliability, Concurrency & Security Hardening
*System stability tests verifying data race safety, crash recovery, path sanitization, and scale performance under heavy load.*

- [x] **Data Race Detector (`-race`)**:
  - Parallel Go test suite execution verifying zero data races under concurrent CLI/MCP operations.
- [x] **Crash Consistency & Transaction Rollback**:
  - Verifying that interrupted sync operations or database write failures rollback cleanly without corrupting the SQLite index or leaving orphan graph links.
- [x] **Database Migration & Upgrade Compatibility**:
  - Backward compatibility tests ensuring existing SQLite databases migrate cleanly across version upgrades without data loss.
- [x] **Path Traversal Security**:
  - Rejection of directory traversal attacks (e.g. `../../etc/passwd` injection attempts).
- [x] **Credential Masking**:
  - Guarantee that storage secrets (S3 keys, DB passwords) are never leaked in logs or `--json` outputs.
- [x] **Scalability & Load Benchmarks**:
  - Benchmark test harness scaling to documents, asserting indexing time < 2s and query latency < 50ms.
