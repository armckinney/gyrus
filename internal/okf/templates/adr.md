---
id: <unique-dashed-id>
title: <Clean Document Title>
category: architecture
type: adr
format: markdown
owner_group: <team-owning-document>
version: 1
status: proposed
immutable: true
last_modified_by: <author-username>
last_updated: YYYY-MM-DD
tags: []
dependencies: []
---

# Architecture Decision Record: [Short Title of Decision]

## Context

* **Author:** [Fill: Name/Team]
* **Date Proposed:** [Fill: YYYY-MM-DD]
* **Deciders:** [Fill: List decision makers]

### 1. The Problem
[Fill: Describe the context and problem we are trying to solve. What is the driver for this change? What architectural constraints or concerns are we addressing?]

### 2. Objectives & Constraints
[Fill: What are the key goals and constraints (performance, cost, security, timeline) that must be met?]

---

## Decision

We will:
[Fill: State clearly and concisely the chosen path. Use active voice and avoid ambiguity. E.g., "We will use PostgreSQL as our primary vector store via pgvector."]

### Key Rationale
* **Reason 1:** [Fill in]
* **Reason 2:** [Fill in]

---

## Alternatives Considered

### Alternative 1: [Name]
* **Pros:** [Fill in]
* **Cons:** [Fill in]

### Alternative 2: [Name]
* **Pros:** [Fill in]
* **Cons:** [Fill in]

---

## Consequences

What are the impacts of making this decision?

* **Positive / Gains:**
  * [Fill: E.g., Reduced infrastructure complexity]
* **Negative / Trade-offs:**
  * [Fill: E.g., Increased CPU overhead on the DB instance]
* **Risks:**
  * [Fill: E.g., Upgrades to pgvector might require temporary downtime]

---

## Compliance & Verification

* **Validation:** [Fill: How do we verify this decision is correctly implemented? E.g., CI/CD checks, schema migrations, architectural linter rules.]
