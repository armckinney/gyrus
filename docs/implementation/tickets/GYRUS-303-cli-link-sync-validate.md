# [GYRUS-303] CLI Relationship & Maintenance Commands (link, unlink, sync, validate)

> **Status:** `NOT STARTED`
> **Phase:** Phase 3 - Gyrus CLI Executable
> **Owner:** Unassigned
> **Dependencies:** [GYRUS-302], [GYRUS-203]

---

## 1. Description
Implement relationship management and maintenance CLI subcommands: `gyrus link`, `gyrus unlink`, `gyrus sync`, and `gyrus validate`.

---

## 2. Acceptance Criteria
- [ ] `gyrus link` / `gyrus unlink`: Creates/deletes relationship edges between documents.
- [ ] `gyrus sync`: Triggers incremental index sync over the target storage directory.
- [ ] `gyrus validate`: Performs frontmatter and transition validation without saving changes.

---

## 3. Implementation Tasks
1. Create `internal/cli/commands/link.go`, `sync.go`, `validate.go`.
2. Connect handlers to `GraphStore`, `IndexStore.Sync()`, and `okf.Validator`.
3. Write unit tests for relationship modifications and validation checks.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/cli/commands/... -v
```
