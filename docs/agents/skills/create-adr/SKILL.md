---
name: create-adr
description: Template and instructions to create a new Architecture Design Record (ADR) in the repository.
applyTo:
  - docs/architecture-design-records/**
---

# Create ADR

Use this skill when you need to document a new architectural or design decision. It provides the required template and naming standards.

## Location & Naming

* **Directory**: `docs/architecture-design-records/`
* **Filename**: `adr-00X-name-of-decision.md` (e.g., `adr-001-use-architecture-design-records.md`)
  * Use 3-digit zero-padded numbers for `00X`.
  * Use lowercase kebab-case for the name.

## ADR Template

Use this Markdown skeleton for the file:

```markdown
# ADR 00X: Title of Decision

* **Status**: Proposed
* **Date**: YYYY-MM-DD
* **Author**: [Name/Agent]

## Context

Describe the background, the problem we are trying to solve, and the forces/constraints at play.

## Decision

State the decision clearly, including chosen options, technical choices, and the specific rationale for why this option was preferred.

## Consequences

Detail the impact of this decision (both positive and negative):
* What becomes easier?
* What becomes harder or requires extra work?
* What are the trade-offs?
```
