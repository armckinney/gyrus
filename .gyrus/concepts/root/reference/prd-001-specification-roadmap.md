---
title: Gyrus Specification Implementation Roadmap
id: prd-001-specification-roadmap
type: prd
format: markdown
category: technical
version: 20
status: active
last_updated: 2026-11-12T05:49:34Z
last_modified_by: ""
tenant: root
tags:
    - roadmap
    - release-phases
    - specification
    - terraform
    - iac
    - infrastructure
---

# Product roadmap

This document is the authoritative delivery roadmap for Gyrus. The task content formerly maintained in root `TODO.md` is kept here so requirements, architecture context, and completion status can be maintained together. The root file now directs contributors to this document.

Gyrus's current foundation supports local filesystem/OKF, shared Git/OKF, and PostgreSQL/native FTS through CLI, local stdio MCP, and the existing local UI. The [architecture drawing](../../../../drw_arch-gyrus.png) and [architecture guide](guide-001-system-architecture.md) describe the progression toward hosted team and enterprise intelligence services.

## 1. Remediate MVP multi-repository configuration

Outcome: a usable shared Git knowledge repository, clear storage-profile choices, and cohesive current documentation.

- [x] Fix the Git storage provider, including configured trunk, direct/branching commit strategies, runtime initiative selection, and credential-helper authentication.
- [x] Update profiles to the supported `local`, `git`, and `postgres` complete stacks.
- [x] Update documentation:
  - Guides and root README.
  - Value proposition and positioning.
  - Configuration/setup documentation with standard configurations.
  - Focused ADRs using the established template.

Git acceptance includes local regression tests and opt-in hosted validation. Both strategies, independent initiatives, schema/link/query isolation, conflict recovery, and fresh-session behavior must be verified. Documentation remains maintained as review comments are addressed; the completed remediation does not close all broader cleanup work automatically.

### Cleanup carried forward

- [ ] Complete the Codex refactor.
  - [x] Rename/remove the `setup` package as appropriate.
- [ ] Ensure documentation stays current across specifications, guides, standards, references, and archived records.

## 2. Stability

Outcome: predictable consistency and resource use as more clients share knowledge.

- [ ] Implement cache providers: `none` (default), `sqlite`, and `redis`.
- [ ] Implement stronger write locking / two-phase commit, initially for PostgreSQL and later Redis-backed coordination.

Current local/Git stacks have no selectable persistent cache provider; they derive ephemeral OKF snapshots. PostgreSQL already has row/version checks, but those checks are not a distributed two-phase commit system. Cache freshness, lock ownership/TTL, failure recovery, and retry/idempotency behavior need explicit design before these tasks are considered complete.

The drawing asks whether the hosted MCP server can manage transaction queues or whether another coordination mechanism is needed. Resolve that with the storage/transaction design rather than silently choosing one from the diagram.

## 3. Platform MVP

Outcome: shared hosted access for teams through MCP and a document-exploration UI, backed by the same knowledge contracts.

### Hosted MCP on Databricks Apps

- [ ] Implement hosted Python MCP on Databricks Apps.
  - [ ] Add the Python/gRPC integration.
  - [ ] Build/package the MCP Databricks App.
  - [ ] Update the agent plugin's MCP configuration for hosted access.

The existing implementation serves local stdio MCP. Hosted transport, deployment, authentication, and the bridge to the core SDK need their own interfaces and acceptance tests. A PostgreSQL-compatible Lakebase target can use the database concept model; configuring a database alone does not implement the hosted MCP application.

### Hosted UI web application

- [ ] Implement the hosted document-viewer web application on Databricks Apps.
  - [ ] Build on the core SDK integration.
  - [ ] Display Gyrus schemas in the UI with a navigation icon.

Keep the distinction between the existing local UI and this hosted platform deployment. Specify document access, filtering, schema browsing, and identity before claiming hosted feature parity.

## 4. Value add: Intelligence Platform

Outcome: connect the maintained knowledge foundation to evaluated intelligence processing and agent-assisted exploration.

- [ ] Enable context intelligence through Lakebase synchronization and context processing.
- [ ] Enable agentic UI integration.
  - [ ] Integrate the UI application's Lakebase state.
  - [ ] Add Mosaic integration for agent/model access.

The drawing includes OLAP persistence in Lakehouse/Iceberg and candidate GraphRAG, knowledge-graph, evaluation, and rule-processing pipelines. These are design candidates for the intelligence work, not current runtime guarantees. Define which outputs are advisory, which may propose knowledge changes, and which require human review before updating shared knowledge.

## 5. Scalability

Outcome: decouple interactive request handling from ingestion and storage execution.

- [ ] Decouple the data layer through an execution queue and ingestion/processing services.

The enterprise drawing proposes a control plane with hosted MCP/UI, PostgreSQL index/graph/search and Redis cache/locks, plus a data plane with an execution queue, transaction processor, document storage, analytical persistence, and intelligence pipelines.

Before implementation, resolve:

- Whether one generic service API should serve MCP and the UI.
- On-behalf-of OAuth propagation, authorization boundaries, and worker identity.
- Queue delivery/idempotency, transaction boundaries, lock TTL/recovery, and read-after-write behavior.
- Cache/index freshness when documents are updated asynchronously.
- Operational visibility and recovery when downstream processing fails.

## Delivery and verification rules

Keep task status here. Update the owning specification/reference when a contract changes, the guide when a workflow changes, and the relevant ADR when a durable choice changes. Preserve archived project tickets as historical execution records rather than a competing current roadmap.

Use [testing standards](standards-002-testing-strategy.md) and the [provider validation guide](guide-004-test-phase-2-1.md) for executable acceptance. A task is complete when its intended behavior and documentation are implemented and its relevant checks pass. No dates or guarantees for hosted/distributed capabilities are implied by their presence in this roadmap.
