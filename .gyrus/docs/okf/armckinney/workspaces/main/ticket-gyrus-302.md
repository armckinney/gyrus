---
id: ticket-gyrus-302
title: GYRUS-302 CLI Core CRUD Commands
category: product
type: freeform
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:43:33Z
---

# [GYRUS-302] CLI Core Mutation Commands (init, create, get, update)

> **Status:** `COMPLETED`
> **Phase:** Phase 3 - Gyrus CLI Executable
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-301], [GYRUS-201]



---

## 1. Description
Implement the core document CRUD CLI subcommands: `gyrus init`, `gyrus create`, `gyrus get`, and `gyrus update`.

---

## 2. Acceptance Criteria
- [ ] `gyrus init`: Bootstraps configuration and local directory setup.
- [ ] `gyrus create`: Accepts metadata flags (`--type`, `--title`, `--category`, `--status`, `--owner-group`, `--tags`, `--dependencies`) and content body (`--content` or `--content-file`).
- [ ] `gyrus get`: Retrieves document text or JSON envelope (`--json`).
- [ ] `gyrus update`: Applies patches with optimistic lock verification (`--expected-version`).

---

## 3. Implementation Tasks
1. Create `internal/cli/commands/init.go`, `create.go`, `get.go`, `update.go`.
2. Connect handlers to Gyrus Core SDK services.
3. Write unit tests for CRUD command invocations and flag parsing.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/cli/commands/... -v
```
