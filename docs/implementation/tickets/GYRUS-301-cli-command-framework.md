# [GYRUS-301] CLI Command Framework & Exit Code Mapper

> **Status:** `NOT STARTED`
> **Phase:** Phase 3 - Gyrus CLI Executable
> **Owner:** Unassigned
> **Dependencies:** [GYRUS-101], [GYRUS-201]

---

## 1. Description
Build the `cmd/gyrus` executable CLI framework (using Cobra/Urfave CLI or standard Go flag handling) and implement deterministic exit code mapping matching Specification 02.

---

## 2. Acceptance Criteria
- [ ] Implement `cmd/gyrus/main.go` entrypoint.
- [ ] Map engine errors to exit codes: `0` (Success), `1` (Validation Error), `2` (Transition Error), `3` (Auth Error), `4` (Concurrency Conflict), `5` (Storage Error).
- [ ] Format output as human-readable text or JSON (`--json` flag).

---

## 3. Implementation Tasks
1. Set up CLI command router in `internal/cli`.
2. Create exit code mapper `internal/cli/exitcodes.go`.
3. Add CLI integration unit tests verifying exit codes.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/cli/... -v
```
