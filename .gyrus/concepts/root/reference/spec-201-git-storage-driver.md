---
title: 'GYRUS-201: Git Remote Storage Driver'
id: spec-201-git-storage-driver
type: specification
format: ""
category: technical
version: 2
status: active
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
tags:
    - phase-2.1
    - storage-driver
    - git
---

# Git storage contract

The Git provider implements concept storage, authoritative scanning, and schema persistence using `go-git` with in-memory checkouts. The system Git executable is used only for HTTPS credential discovery.

## Strategies and consistency

`direct` reads/writes `trunk_branch` (default `main`). `branching` requires an initiative branch supplied at runtime; initiative names are never configuration fields. A session without a selection rejects knowledge operations. Trunk must exist for branching. A missing initiative reads the current trunk snapshot and becomes a remote branch on its first mutation. Once created, the initiative is reused without automatic trunk integration.

Concept CRUD, archive, source-owned links, schema CRUD, and OKF scans all use the selected branch. Each operation reloads confirmed remote state. Writes use normal non-forced pushes; stale versions and competing rejected pushes are conflicts. A failed push cannot leak unconfirmed data into later reads. Identical schema saves succeed without creating commits. Cancellation reaches credential lookup, remote listing, cloning, and pushing.

A session switch prepares and validates a candidate checkout before activation. Failed switches preserve the previous selection. Store operations share a mutex; MCP additionally pins whole tool/resource/prompt requests while switching waits. Different MCP processes hold independent selections and never write them to config.

## Authentication and layout

HTTPS uses noninteractive `git credential fill`, preserving helper-provided username and password. No token-variable fallback or token-bearing URL is supported. SSH respects the URL username, uses the SSH agent or standard key files, and verifies known hosts. Remote errors redact credentials.

Concepts live under `.gyrus/concepts`; schemas under `.gyrus/schemas`. `storage.root` does not relocate Git content. Qualified names remain the public identity.

## Interfaces and review

`--git-branch <name>` selects branching for CLI calls and MCP startup. CLI `git status` and MCP `gyrus_git_status` report strategy, trunk, active branch, and selection-required state. MCP `gyrus_git_switch` changes only the running server session. Branching requires a distinct valid initiative name; direct rejects an initiative argument.

Gyrus creates no PRs and performs no merges/rebases. Humans review and merge through their host. After a merge, select direct to read trunk, choose a new initiative, or integrate trunk into an existing initiative externally.

## Verification

Local bare-remotes cover lifecycle, schema idempotency, branch creation/isolation, version/push conflicts, authentication, and cancellation. CLI and MCP tests verify runtime selection, independent sessions, restart behavior, and unchanged configuration. The opt-in `TestGitLiveRemote` uses `GYRUS_TEST_GIT_URL` pointing to a disposable repository. See the [validation guide](guide-004-test-phase-2-1.md).
