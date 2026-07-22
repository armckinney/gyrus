---
id: ticket-gyrus-102
title: GYRUS-102 OKF Envelope Parser
category: product
type: freeform
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:43:33Z
---

# [GYRUS-102] OKF Envelope Parser & Schema Validator

> **Status:** `COMPLETED`
> **Phase:** Phase 1 - Core SDK & OKF Domain Model
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-101]



---

## 1. Description
Implement the Open Knowledge Format (OKF) parser and validator in `internal/okf`. The parser reads both JSON envelopes and Markdown files with YAML frontmatter headers, enforcing pattern regexes, mandatory fields, and type enums matching Specification 01.

---

## 2. Acceptance Criteria
- [ ] Parse Markdown files with YAML frontmatter headers into `gyrus.Document`.
- [ ] Parse JSON envelopes into `gyrus.Document`.
- [ ] Validate regex boundaries for `id` (`^[a-z0-9-_]+$`), non-empty `title`, valid `category` enums, valid `type` enums, and non-empty `owner_group`.
- [ ] Output clear validation error messages if fields are missing or invalid.

---

## 3. Implementation Tasks
1. Create `internal/okf/parser.go` using a fast YAML frontmatter parser.
2. Create `internal/okf/validator.go` with field validation functions.
3. Write test cases covering valid OKF documents and invalid/corrupt frontmatter inputs.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/okf/... -v
```
