---
id: ticket-gyrus-001
title: GYRUS-001 Repository Cleanup & Bootstrap
category: product
type: freeform
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:43:33Z
---

# [GYRUS-001] Repository Cleanup & Go Module Scaffolding

> **Status:** `COMPLETED`
> **Phase:** Phase 0 - Repository Cleanup & Scaffolding
> **Owner:** Antigravity Agent
> **Dependencies:** None



---

## 1. Description
Clean up the legacy `template-go` boilerplate code (`cmd/api/`, `internal/handlers`, `internal/repository`, `internal/server`, `internal/database`, `internal/middleware`), initialize the official Go module `github.com/armckinney/gyrus` (or `gyrus`), and set up the target package directory structure (`pkg/gyrus`, `internal/provider`, `internal/okf`, `internal/lifecycle`, `internal/cli`, `internal/mcp`, `cmd/gyrus`).

---

## 2. Acceptance Criteria
- [ ] Legacy template files removed (`cmd/api/`, `example.com/template-go` references in `internal/`).
- [ ] Go module initialized/renamed cleanly to `github.com/armckinney/gyrus` with `go.mod` file created.
- [ ] Directory structure scaffolded: `pkg/gyrus`, `cmd/gyrus`, `internal/okf`, `internal/lifecycle`, `internal/provider`, `internal/cli`, `internal/mcp`.
- [ ] `go vet ./...` and `go test ./...` run cleanly with zero errors.

---

## 3. Implementation Tasks
1. Remove stale `cmd/api/` and `internal/` template files.
2. Initialize `go.mod` for `github.com/armckinney/gyrus`.
3. Add core dependencies: `go get github.com/spf13/cobra modernc.org/sqlite github.com/mark3labs/mcp-go gopkg.in/yaml.v3`.
4. Create target directory folders and initial package doc files (`doc.go`).
5. Update `Makefile` build target to `go build -o gyrus cmd/gyrus/main.go`.

---

## 4. Verification & Testing Commands
```bash
go vet ./...
go mod tidy
```
