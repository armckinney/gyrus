# [GYRUS-201] Localfs Storage Provider & Path Resolution

> **Status:** `COMPLETED`
> **Phase:** Phase 2 - Local Storage & SQLite Provider Engines
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-102]



---

## 1. Description
Implement the `localfs` storage provider in `internal/provider/localfs`. Handle storage root resolution matching the precedence hierarchy (`--storage-path`, `GYRUS_STORAGE_PATH`, `.gyrus/config.yaml`, `~/.config/gyrus/config.yaml`, `~/.gyrus/`), and write files following the OKF bundle directory topology.

---

## 2. Acceptance Criteria
- [ ] Implement `gyrus.DocumentStore` for `localfs`.
- [ ] Implement storage path resolution hierarchy (`CLI flag` ➔ `ENV` ➔ `Local Config` ➔ `User Config` ➔ `Default`).
- [ ] Persist documents into the OKF bundle directory structure (`okf/<team>/reference/` and `okf/<team>/workspaces/<repo-x>/`).
- [ ] Support atomic file writes and file reads.

---

## 3. Implementation Tasks
1. Create `internal/provider/localfs/store.go`.
2. Implement directory resolution in `internal/provider/localfs/resolver.go`.
3. Add unit tests verifying file creation, reading, and path precedence.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/provider/localfs/... -v
```
