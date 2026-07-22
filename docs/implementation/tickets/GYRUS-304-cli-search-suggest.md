# [GYRUS-304] CLI Context Discovery Commands (search, suggest-context, schema)

> **Status:** `COMPLETED`
> **Phase:** Phase 3 - Gyrus CLI Executable
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-301], [GYRUS-302]


---

## 1. Description
Implement CLI query and context discovery subcommands: `gyrus search`, `gyrus suggest-context`, and `gyrus schema`.

---

## 2. Acceptance Criteria
- [ ] `gyrus search`: Executes FTS5 search with `--query`, `--type`, `--category`, `--status`, `--tag`, `--owner-group`, and `--json` formatting.
- [ ] `gyrus suggest-context`: Linearizes top context documents matching prompt context.
- [ ] `gyrus schema`: Prints structural Markdown templates for requested doc types (e.g. `adr`, `prd`, `specification`).

---

## 3. Implementation Tasks
1. Create `internal/cli/commands/search.go`, `suggest.go`, `schema.go`.
2. Connect `suggest-context` to FTS search and edge traversal ranking.
3. Write test cases validating output text and JSON payloads.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/cli/commands/... -v
```
