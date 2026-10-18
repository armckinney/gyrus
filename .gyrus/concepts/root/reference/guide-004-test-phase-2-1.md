---
title: Storage-provider and lexical-search verification
id: guide-004-test-phase-2-1
type: guide
format: markdown
category: technical
version: 9
status: active
last_updated: 2026-10-18T00:00:28Z
last_modified_by: ""
tenant: root
tags:
    - testing
    - phase-2.1
    - storage-drivers
    - search-providers
    - init
    - skills
links:
    - concept: root/reference/prd-001-specification-roadmap
      rel: depends_on
    - concept: root/reference/spec-201-git-storage-driver
      rel: depends_on
    - concept: root/reference/spec-202-cloud-blob-storage-driver
      rel: depends_on
    - concept: root/reference/spec-203-postgres-storage-index-driver
      rel: depends_on
    - concept: root/reference/spec-204-postgres-fts-search-driver
      rel: depends_on
---

# Storage-provider and lexical-search verification

Use an installed Gyrus executable. For complete consumer acceptance checks, run the packaged gyrus-test-lifecycle skill; it exercises CLI, MCP, and UI without repository source or development tooling. The local command sequence is in [the local walkthrough](guide-005-local-testing-walkthrough.md).

## Supported compositions

| Profile | Storage | Index / graph / search |
| --- | --- | --- |
| local | localfs | okf / okf / okf |
| git | git | okf / okf / okf |
| blob, s3, azure, gcs | configured object store | okf / okf / okf |
| postgres | PostgreSQL | postgres / postgres / postgres_fts |

Vector search, embedding clients, and the vector profile/configuration have been removed. Semantic search is deferred to future backend work. SQLite index/graph/search providers and the sync command/tool are also removed. File-backed queries read authoritative concepts directly.

## Isolate the test configuration

Use a dedicated temporary project and private home, clearing inherited GYRUS selectors before initialization. Resolve the installed executable before changing directories. Inspect resolved settings with `gyrus config show --json`; verify the storage target before creating fixtures. Use a dedicated test Git repository, bucket/container/prefix, or PostgreSQL database for remote-provider checks.

Initialize the selected profile in its test project:

```bash
gyrus config init --profile local --tenant verification --project demo
gyrus config show --json
```

Supported alternatives are git, blob, s3, azure, gcs, and postgres. Unknown profiles and provider names fail through shared validation before configuration writes or provider initialization. Existing files must remain unchanged on rejected initialization.

## Git

Generate the Git profile with `gyrus config init --profile git`, then supply the test repository and branch in the generated nested configuration:

```yaml
tenant: verification
project: demo
storage:
  provider: git
  root: .gyrus
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
git:
  repo_url: /absolute/path/to/dedicated-test-repository.git
  branch: main
```

Git uses existing credential discovery for authenticated remotes. Create a reference guide using tenant verification and scope reference; read it by verification/reference/guide-provider-test in a fresh process. Confirm the concept exists under .gyrus/concepts/verification/reference/ in the remote branch. Conflicting pushes must fail without exposing unconfirmed content. Do not force-push during verification.

## Blob and cloud storage

Generate blob, s3, azure, or gcs configuration and set its actual bucket/container details. Retain OKF index/graph/search. For Azure the current nested fields are:

```yaml
storage:
  provider: azure_blob
  root: .gyrus
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
azure_blob:
  storage_account: test-account
  container_name: test-container
tenant: verification
project: demo
```

Use the selected provider's credential discovery. Verify canonical concept keys under the configured prefix followed by concepts/verification/reference/guide-provider-test.md. Updates/deletes must use observed ETag/generation preconditions; unsupported conditional mutation must fail explicitly. Use fresh processes to prove persistence and scope resolution.

## PostgreSQL and full-text search

Generate the postgres profile and set postgres.connection_string to a dedicated database. Storage/index/graph/search must resolve to postgres/postgres/postgres_fts. PostgreSQL tsvector full-text matching remains supported and requires no embedding service.

```bash
gyrus create --id guide-provider-test --title "Provider verification" --type guide --category technical --tenant verification --scope reference --status draft --content "providerverificationmarker"
gyrus get verification/reference/guide-provider-test --json
gyrus search --query providerverificationmarker --tenant verification --scope reference --max-results 20 --json
```

Create another fixture in project demo and one in another project. Verify default project-plus-reference results, explicit project/reference scopes, tenant isolation, and uncapped all-scope metadata browsing. Repeated local IDs in different scopes must resolve independently by qualified name.

An update with a stale positive expected version must return a concurrency error, CLI exit code 4, and MCP isError=true. PostgreSQL must roll back without changing the concept or appending history. Database graph links require existing endpoints. Read incoming/outgoing edges through MCP and verify database-only depends_on edges appear in context suggestions.

For repository contributors, live PostgreSQL integration is explicitly opt-in:

```bash
GYRUS_TEST_POSTGRES_URL="<dedicated-test-database-DSN>" go test ./tests/integration -run '^TestIntegration_PostgresLive$' -count=1 -v
```

Without that variable the live test skips. Mocked tests cover filtering, conflict typing, and rollback without a database. The dev-container retains the optional PostgreSQL testing sidecar; no Ollama/embedding sidecar is required.

## Common acceptance checks

- Keyword queries default to ten results. Explicit CLI --max-results and MCP max_results values pass through, including requests above ten. Empty queries browse all matching metadata unless positively capped.
- Existing-concept operations require qualified names. Create accepts a local ID plus tenant/scope/project and returns the qualified name.
- Persist relationships as concept/rel pairs only. Use link/unlink and incoming graph queries rather than dependencies arrays or inverse copies.
- The UI graph expands only original visible sources. Immediate external targets are display-only ghosts; repeated identical requests return deterministically sorted, deduplicated nodes/edges. A failed edge query returns HTTP 500.
- Validate fixture Markdown with gyrus validate. Malformed collections fail queries rather than returning partial results.
- Stop only the services started for this test, remove owned fixtures using the backend's supported lifecycle, and retain diagnostic evidence on failure. Do not delete existing databases/backups as part of testing.

Install the Gyrus plugin through the coding client's native plugin mechanism. Use the packaged gyrus and gyrus-test-lifecycle skills; the former gyrus client commands and separate gyrus-cli/gyrus-mcp skills are not current interfaces. MCP sessions require initialization, an initialized notification, and persistent stdin/stdout; use tool/resource discovery instead of sending unrelated requests to fresh processes.
