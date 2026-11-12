---
title: Gyrus Testing Strategy & QA Standards
id: standards-002-testing-strategy
type: standards
format: markdown
category: technical
version: 11
status: active
last_updated: 2026-11-12T05:49:34Z
last_modified_by: ""
tenant: root
tags:
    - testing
    - qa
    - standards
    - integration
    - mcp
    - cli
    - agents
    - phase-2.4
links:
    - concept: root/reference/prd-001-specification-roadmap
      rel: depends_on
    - concept: root/reference/spec-001-contract-schema
      rel: depends_on
    - concept: root/reference/spec-002-cli-interface
      rel: depends_on
    - concept: root/reference/spec-003-mcp-interface
      rel: depends_on
    - concept: root/reference/spec-004-core-engine
      rel: depends_on
    - concept: root/reference/standards-001-doc-strategy
      rel: depends_on
---

# Testing strategy and QA standards

This standard describes how contributors verify Gyrus at package, provider, product, protocol, and agent-workflow boundaries. Package tests provide fast feedback; CLI/MCP subprocess tests verify the experience exposed to humans and agents. Test names and executed assertions establish coverage, not a checklist alone.

## Test architecture and responsibilities

```mermaid
flowchart TD
    Domain["Domain, OKF parsing, lifecycle, serialization"] --> Providers["Localfs, Git, PostgreSQL and query adapters"]
    Providers --> CLI["CLI binary, arguments, output and exit codes"]
    Providers --> MCP["MCP tools, resources, prompts and stdio protocol"]
    Providers --> UI["Local UI rendering, search and graph APIs"]
    CLI --> Product["Installation, plugin packaging and consumer workflows"]
    MCP --> Product
    Product --> Agent["Opt-in agent evaluation and live infrastructure"]
```

| Level | Execution surface | Main evidence |
| --- | --- | --- |
| Package | `internal/config`, `internal/domain`, `internal/format`, persistence packages | Strict configuration, qualified identities, parsing, lifecycle guards, serialization, locks |
| Provider | `internal/provider` and `tests/integration` | Durable concepts/schemas, source-owned links, search/filter limits, version and push conflicts |
| CLI | `internal/cli`, `tests/cli` | Built binary, real flags/files, JSON/text output, process exit codes |
| MCP | `internal/mcp`, `tests/mcp` | Handshake/discovery, tool results, resources/prompts, runtime branch forwarding and isolation |
| UI | `internal/ui` | HTTP handlers, templates, document rendering, graph scope and deterministic ordering |
| Distribution/agents | `tests/install`, `tests/plugins`, `tests/skills` | Installer/build checks, plugin layout, prompt-routing fixtures, explicitly enabled live evaluation |

Use isolated temporary roots and local bare repositories for default tests. PostgreSQL unit tests use mocks; a live database is separately enabled. A successful default run can include skipped external tests; report those skips rather than implying that every host or environment was exercised.

## Repository commands

The [Makefile](../../../../Makefile) defines the contributor entry points:

| Command | Actual behavior | Prerequisites / effects |
| --- | --- | --- |
| `make test` | Runs `go test ./...` with `GYRUS_PROJECT` and `GYRUS_CONFIG` cleared | Includes package tests and ordinary CLI/MCP/integration tests; external gates still apply |
| `make lint` | Lists files reported by `gofmt -l .`, then runs `go vet ./...` | Formatting output requires attention; this target does not itself fail solely because files were listed |
| `make fmt` | Runs `go fmt ./...` | Rewrites source formatting |
| `make build` | Builds `gyrus` with `-buildvcs=false` | Writes the root binary; use a temporary output path when preserving it matters |
| `make test-integration` | Builds, then runs `go test ./tests/...` with `RUN_INTEGRATION_TESTS=1` and `RUN_DOCKER_TESTS=1` | Enables Docker execution when available; does not supply Git credentials or a database URL |
| `make all` | Runs clean, fmt, lint, test, build | Removes build artifacts and formats source before testing |
| `make install` | Builds and symlinks the binary into `/usr/local/bin` | Changes the local installation |

For a focused change, run its affected packages first, then the required repository checks. Examples:

```bash
# Standard regression run, isolated from project-selection environment variables
make test

# Git/configuration and provider-wiring checks
go test ./internal/config ./internal/provider/... ./internal/app

# Actual public CLI and stdio MCP contracts
go test ./tests/cli ./tests/mcp -count=1

# Concurrency checks around shared storage and live branch switching
go test -race ./internal/provider/storage/git ./internal/mcp ./internal/app

# Broader race run when the change touches more subsystems
go test -race ./...

# Static checks and a non-mutating formatting report
go vet ./...
gofmt -l cmd internal pkg tests

# Existing local persistence/search/graph benchmarks
go test ./tests/integration -run '^$' -bench . -benchmem
```

CI runs `go vet ./...`, explicitly fails on a nonempty formatting report, and runs `go test ./... -v`. See the [CI workflow](../../../../.github/workflows/wf-ci.yml). Benchmarks report latency and allocation measurements; they do not establish an undocumented fixed latency guarantee.

## Environment gates and prerequisites

| Variable | Current effect |
| --- | --- |
| `RUN_DOCKER_TESTS` or `RUN_INTEGRATION_TESTS` | A nonempty value enables the live Docker test in `tests/install`; missing Docker or an inaccessible daemon still causes a skip |
| `GYRUS_TEST_POSTGRES_URL` | Enables `TestIntegration_PostgresLive` against the supplied isolated PostgreSQL database |
| `GYRUS_TEST_GIT_URL` | Enables `TestGitLiveRemote` against the supplied disposable Git remote |
| `GYRUS_TEST_GIT_RUN_ID` | Optional unique concept ID used by the live Git fixture; otherwise generated per run |
| `RUN_AGENT_EVAL` or `RUN_E2E_TESTS` | A nonempty value enables the Antigravity evaluation script; its actual path depends on installed agent tooling |
| `GYRUS_PROJECT`, `GYRUS_CONFIG` | Select runtime configuration; Makefile test targets clear them to avoid accidental fixture routing |

These gates are independent: `make test-integration` does not automatically enable live Git/PostgreSQL or live agent evaluation. The existing agent script validates the plugin and attempts an `agy` prompt when available; without `agy` it checks MCP transport. Treat that as an integration smoke test, not a quality score for autonomous reasoning. Inspect the script's output to report which path ran.

## Live infrastructure runbooks

### Git

Use a disposable repository and preconfigured Git credential-manager or SSH access. HTTPS credentials are obtained through noninteractive `git credential fill`; do not put tokens in command arguments or remote URLs.

```bash
GYRUS_TEST_GIT_URL=https://github.com/OWNER/test-wiki.git \
  go test ./tests/integration -run '^TestGitLiveRemote$' -v -count=1
```

The test validates direct trunk writes, two initiative branches, schema/link/search isolation, idempotent schema saves, fresh sessions, and return to unchanged trunk. It leaves fixtures and branches for inspection. Use the [provider validation guide](guide-004-test-phase-2-1.md) for external PR review/merge and fresh-trunk verification. GitHub results do not imply live Bitbucket coverage.

### PostgreSQL

Use a dedicated database that can receive fixture concepts, edges, schema initialization, and history writes.

```bash
GYRUS_TEST_POSTGRES_URL='postgres://USER:PASSWORD@HOST:5432/gyrus_test?sslmode=require' \
  go test ./tests/integration -run '^TestIntegration_PostgresLive$' -v -count=1
```

For a local disposable database, use its actual host/port and appropriate SSL setting. The integration test exercises storage, graph, native FTS, and query filters. Mocked provider tests independently verify expected-version conflicts, rollback, and query/decode failures. Do not infer live transaction or host behavior solely from mock results.

### Docker and agent workflows

```bash
# Builds the binary and enables the existing Docker integration gate
make test-integration

# Build first: the evaluation script invokes the root binary
make build
RUN_AGENT_EVAL=1 go test ./tests/skills -run '^TestAntigravityIntegration$' -v -count=1
```

Docker requires the executable and a running accessible daemon. Live agent evaluation requires the relevant installed/authenticated harness for a real prompt run. It can invoke external services; keep it outside the ordinary regression suite.

## Contract acceptance by subsystem

- **Domain/configuration:** Required metadata, qualified names, scope/project derivation, invalid paths, minimal typed links, complete provider stacks, strict YAML errors, and lifecycle/body-immutability rules.
- **Storage/query:** Local atomic replacement and locks; Git commit history, selected branches and rejected-push recovery; PostgreSQL version checks/rollback; malformed collections, duplicate identities, exact filters, and result limits.
- **Graph/context:** Directional relations, backlinks, cycles/self-links, provider-specific missing-target behavior, deterministic traversal, and project/dependency/reference ordering. `max_concepts` caps selected context. CLI `--max-tokens` is currently accepted but not enforced; there is no current token-truncation acceptance claim.
- **CLI:** Required flags, positional qualified names, configuration generation, text/JSON payloads, exit codes, runtime branch arguments, and absence of retired commands. Schema validation is a local file operation.
- **MCP:** Initialization and discovery, actual `gyrus_*` tools, text/JSON-text results, `memory://` resources, active ADR templates, storage error propagation, and independent Git sessions. Switches must wait for in-flight knowledge requests and never save configuration.
- **UI:** Document/search rendering, fixed visible graph sources, display-only missing/filtered targets, edge/node deduplication, stable ordering, and HTTP errors without partial graph payloads.
- **Distribution:** Canonical plugin manifests/skills/rules, installer syntax, binary build behavior, and client-specific packaging. Test behavior through installed public interfaces when validating consumer workflows.

Keep test purpose and execution surface clear in names and comments. For substantial product/integration tests, the established header format records `[Test Level]`, `[Purpose]`, `[Execution Surface]`, and `[Assertions]`. Do not mark unexecuted scenarios complete or copy stale assertions into documentation.

## Documentation verification and reporting

Validate concept frontmatter, local ID/filename agreement, qualified links, relative Markdown links, example configurations, and explicit distinctions between current behavior and the [roadmap](prd-001-specification-roadmap.md). Prefer representative CLI/MCP contract checks over adding tests that merely duplicate prose.

Record the command, result, relevant skips, and any external target used in the change review or testing guide. Routine implementation work does not require a release note. New backends, caching, hosted services, or transaction guarantees need their own implemented acceptance evidence.
