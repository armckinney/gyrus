# [GYRUS-101] Core SDK Public Interfaces & Domain Models

> **Status:** `NOT STARTED`
> **Phase:** Phase 1 - Core SDK & OKF Domain Model
> **Owner:** Unassigned
> **Dependencies:** None

---

## 1. Description
Establish the public Go package `pkg/gyrus` and define the core domain models (`Document`, `DocumentRef`, `DocumentPatch`, `SearchResult`, `DocumentEdge`, `SearchFilter`) and core provider interfaces (`DocumentStore`, `IndexStore`, `GraphStore`, `SearchProvider`).

---

## 2. Acceptance Criteria
- [ ] Public Go package `pkg/gyrus` created with zero circular dependencies.
- [ ] `Document` struct includes all OKF metadata fields (`ID`, `Title`, `Category`, `Type`, `Format`, `OwnerGroup`, `Version`, `Status`, `LastModifiedBy`, `LastUpdated`, `Tags`, `Dependencies`, `Content`).
- [ ] Core provider Go interfaces (`DocumentStore`, `IndexStore`, `GraphStore`, `SearchProvider`) defined matching Specification 04.

---

## 3. Implementation Tasks
1. Initialize Go package structure under `pkg/gyrus/types.go` and `pkg/gyrus/interfaces.go`.
2. Define JSON and YAML struct tags for all domain structs.
3. Write unit tests for struct serialization/deserialization.

---

## 4. Verification & Testing Commands
```bash
go test ./pkg/gyrus/... -v
```
