# [GYRUS-103] Lifecycle State-Machine Validation Engine

> **Status:** `NOT STARTED`
> **Phase:** Phase 1 - Core SDK & OKF Domain Model
> **Owner:** Unassigned
> **Dependencies:** [GYRUS-101], [GYRUS-102]

---

## 1. Description
Build the state-machine transition validation engine in `internal/lifecycle`. Enforce state transitions for `adr`, `improvement-proposal`, and general document types matching Specification 04.

---

## 2. Acceptance Criteria
- [ ] Enforce ADR transitions: `proposed` ➔ `accepted` ➔ `superseded` / `deprecated` / `rejected`.
- [ ] Enforce Improvement Proposal transitions: `draft` ➔ `reviewing` ➔ `approved` ➔ `implemented` / `abandoned` / `rejected`.
- [ ] Enforce General Doc Type transitions: `draft` ➔ `active` ➔ `deprecated` ➔ `archived`.
- [ ] Reject illegal state transitions (e.g. `accepted` ➔ `proposed` or `superseded` ➔ `accepted`) with `TRANSITION_ERROR`.

---

## 3. Implementation Tasks
1. Implement state transition maps in `internal/lifecycle/engine.go`.
2. Expose `ValidateTransition(docType, currentStatus, newStatus) error`.
3. Add unit tests for every valid and invalid transition combination.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/lifecycle/... -v
```
