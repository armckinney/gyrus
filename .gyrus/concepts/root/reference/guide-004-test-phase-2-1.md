---
title: Storage-provider and lexical-search verification
id: guide-004-test-phase-2-1
type: guide
format: markdown
category: technical
version: 12
status: active
last_updated: 2026-11-12T05:50:25Z
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
    - concept: root/reference/spec-203-postgres-storage-index-driver
      rel: depends_on
    - concept: root/reference/spec-204-postgres-fts-search-driver
      rel: depends_on
---

# Validate supported storage stacks

Run default tests with isolated local fixtures:

```bash
go test ./...
go test -race ./internal/provider/storage/git ./internal/mcp ./internal/app
go vet ./...
```

The three supported stacks are localfs/OKF, Git/OKF, and PostgreSQL/PostgreSQL/FTS. Use the [standard configurations](tech-ref-002-config-schema.md). Default Git tests use local bare repositories and require no host credentials. PostgreSQL integration uses the existing `GYRUS_TEST_POSTGRES_URL` opt-in fixture when available; inspect the integration test for its exact environment contract.

## Live GitHub validation

Configure Git Credential Manager with read/write access to a disposable repository. Gyrus invokes `git credential fill` without interactive prompts. Do not put tokens in command arguments or configuration URLs.

```bash
GYRUS_TEST_GIT_URL=https://github.com/OWNER/test-wiki.git \
  go test ./tests/integration -run '^TestGitLiveRemote$' -v -count=1
```

The test creates a unique trunk concept, two initiative branches, typed links, and branch-specific schemas. It verifies search, idempotent schema saves, fresh-session reads, and return to unchanged trunk. It leaves fixtures for inspection. To choose a fixture ID, set `GYRUS_TEST_GIT_RUN_ID` to a unique valid concept ID.

Open a PR for the logged `/a` branch against trunk, review its concept/schema/link changes, and merge using GitHub. Confirm a fresh direct-mode Gyrus read sees the reviewed title and link. The `/b` branch remains independent; integrate trunk changes externally if continuing it. Gyrus neither opens nor merges PRs.

GitHub validation used private [armckinney/test-wiki](https://github.com/armckinney/test-wiki), including both strategies, independent initiatives, schemas, links, search, and fresh sessions. [Validation PR #1](https://github.com/armckinney/test-wiki/pull/1) was created and merged externally through GitHub; a subsequent direct Gyrus read observed the merged title, link, and schema override. This is provider-validation evidence, not a separate release milestone. Bitbucket was not live-tested; HTTPS delegates to the same host-aware credential manager.
