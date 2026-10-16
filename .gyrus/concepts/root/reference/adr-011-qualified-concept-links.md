---
title: Qualified concept identity and minimal typed links
id: adr-011-qualified-concept-links
type: adr
format: markdown
category: architecture
version: 3
status: accepted
last_updated: 2026-10-16T16:56:27Z
last_modified_by: ""
tenant: root
links:
    - concept: root/reference/spec-001-contract-schema
      rel: depends_on
---

# Qualified concept identity and minimal typed links

## Context

Local IDs collide across tenants and projects. Dependency arrays duplicate typed relationships, while per-link creation metadata adds context without helping resolution.

## Decision

Retain a local frontmatter id and derive a canonical collection-relative name from tenant and scope: tenant/reference/id or tenant/projects/project/id. Require that name for existing-concept operations across every provider and interface. Creation accepts a local ID and location and returns both identities. Store relationships only as links containing concept and rel. Remove dependency arrays and edge creation metadata. Keep depends_on, implements, supersedes, and mitigates as directed semantic assertions. Derive incoming views; permit general cycles, self-links, and qualified missing targets where supported by the backend.

## Consequences

Local IDs can repeat across scopes, but qualified names are unique. Strict runtime readers reject obsolete fields and short lookup values. A separate inventory-based converter requires explicit mappings when legacy targets are ambiguous or absent, backs up data, preserves bodies and statuses, and verifies idempotency. Old database schemas require explicit export/rebuild. Moves change names without automatic link rewriting; cross-collection references are deferred. Typed metadata is a Gyrus extension to upstream OKF. Document-level revisions and conditional-write protection remain intact.
