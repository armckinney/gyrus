---
title: Delegate HTTPS Git authentication to credential helpers
id: adr-016-git-credential-helpers
type: adr
format: markdown
category: architecture
version: 2
status: accepted
last_updated: 2026-11-12T05:07:17Z
tenant: root
immutable: true
---

# Architecture Decision Record: Delegate HTTPS Git authentication to credential helpers

## Context

* **Author:** Codex, implementing the repository owner's approved plan.
* **Date Proposed:** 2026-11-11
* **Deciders:** Repository owner.

### 1. The Problem

Git hosts require different credential usernames and token forms. Hardcoded usernames and unrelated token-variable fallbacks bypass the user's credential manager.

### 2. Objectives & Constraints

Reuse configured credentials, prevent interactive prompts in agent sessions, preserve cancellation, and keep secrets out of diagnostics.

---

## Decision

We will use noninteractive, cancellable `git credential fill` for HTTPS and preserve the returned username/password. Retain `go-git` transport and in-memory storage. SSH respects URL users and known hosts.

### Key Rationale

* **Host compatibility:** The helper supplies host-specific credentials.
* **Bounded responsibility:** Only credential lookup requires system Git; transport remains in the provider.

---

## Alternatives Considered

### Alternative 1: Implement host-specific token handling

* **Pros:** No configured helper is needed for supported token variables.
* **Cons:** Duplicates host policy and requires token-specific usernames.

### Alternative 2: Replace the provider with shell Git

* **Pros:** Uses system Git for every operation.
* **Cons:** Broadens the change beyond authentication and replaces the existing in-memory implementation.

---

## Consequences

* **Positive / Gains:** Private remotes reuse existing credential-manager configuration.
* **Negative / Trade-offs:** HTTPS requires the Git executable; private access also requires stored helper credentials.
* **Risks:** Missing credentials permit anonymous transport for public remotes; private remotes fail. Passwords must never be placed in configuration URLs or diagnostic output.

---

## Compliance & Verification

* **Validation:** Tests verify helper protocol, noninteractive settings, cancellation, username preservation, and redaction. Live GitHub validation used the configured credential manager.
