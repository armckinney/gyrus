---
id: <unique-dashed-id>
title: <Clean Document Title>
category: architecture
type: adr
format: markdown
owner_group: <team-owning-document>
version: 1
status: proposed
last_modified_by: <author-username>
last_updated: YYYY-MM-DD
tags: []
dependencies: []
---

# Architecture Decision Record: [Short Title of Decision]

## Context

* **Author:** [TODO: Name/Team]
* **Date Proposed:** [TODO: YYYY-MM-DD]
* **Deciders:** [TODO: List decision makers]

### 1. The Problem
[TODO: Describe the context and problem we are trying to solve. What is the driver for this change? What architectural constraints or concerns are we addressing?]

### 2. Objectives & Constraints
[TODO: What are the key goals and constraints (performance, cost, security, timeline) that must be met?]

---

## Decision

We will:
[TODO: State clearly and concisely the chosen path. Use active voice and avoid ambiguity. E.g., "We will use PostgreSQL as our primary vector store via pgvector."]

### Key Rationale
* **Reason 1:** [TODO]
* **Reason 2:** [TODO]

---

## Alternatives Considered

### Alternative 1: [Name]
* **Pros:** [TODO]
* **Cons:** [TODO]

### Alternative 2: [Name]
* **Pros:** [TODO]
* **Cons:** [TODO]

---

## Consequences

What are the impacts of making this decision?

* **Positive / Gains:**
  * [TODO: E.g., Reduced infrastructure complexity]
* **Negative / Trade-offs:**
  * [TODO: E.g., Increased CPU overhead on the DB instance]
* **Risks:**
  * [TODO: E.g., Upgrades to pgvector might require temporary downtime]

---

## Compliance & Verification

* **Validation:** [TODO: How do we verify this decision is correctly implemented? E.g., CI/CD checks, schema migrations, architectural linter rules.]
