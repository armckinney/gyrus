---
title: Gyrus Value Proposition & Strategic Product Positioning
id: prd-002-value-proposition-positioning
type: prd
format: ""
category: technical
version: 6
status: active
last_updated: 2026-11-12T05:49:34Z
tenant: root
tags:
    - product-positioning
    - value-proposition
    - competitive-analysis
    - context-control-plane
links: []
---

# Gyrus Intelligence Platform: value proposition and positioning

## Strategic value

**Gyrus helps teams turn engineering knowledge into a durable, shared asset that humans and heterogeneous AI agents can discover, apply, review, and improve across repositories and sessions.** Its context-control foundation joins knowledge creation, structured relationships, lifecycle governance, and task-oriented retrieval behind consistent interfaces.

The highest-value outcome is organizational continuity: a decision made during one initiative can inform another repository, a new teammate, or a different agent harness without being reconstructed from scattered conversations. As teams adopt more agents, Gyrus provides a common knowledge model and review workflow that can survive changes in tools, models, and storage choices.

## The user problem

Engineering intent is spread across code, tickets, PR discussions, chat, guides, and individual memory. Source code records implementation, but it does not reliably explain every constraint, rejected alternative, operational lesson, or dependency behind that implementation. Agent sessions can produce useful insights that disappear when their conversations end.

Without a maintained knowledge layer, teams repeatedly rediscover decisions, provide inconsistent context to different tools, and struggle to distinguish a proposed observation from an accepted decision. Search can find text, but a useful engineering workflow also needs identity, scope, relationships, review, and continued maintenance.

Gyrus focuses on that coordination: representing knowledge clearly, connecting it to related concepts, exposing it through common interfaces, and making updates part of engineering work.

## Highest-value capabilities and outcomes

| Value | What Gyrus provides today | Outcome for the team |
| --- | --- | --- |
| Shared organizational memory | Durable concepts with qualified tenant/project/reference identity | Knowledge remains available beyond an individual session or code repository |
| Human/agent interoperability | The same lifecycle engine through CLI, local MCP, local UI, and Go contracts | Different tools can consume and maintain the same collection |
| Traceable engineering intent | ADRs, specifications, PRDs, standards, guides, and typed relationships | Decisions and requirements can be connected to the work they inform |
| Governed knowledge changes | Type-specific lifecycle, protected finalized bodies, versions, and Git initiative review | Teams can review knowledge updates and supersede decisions explicitly |
| Applicable context | Project/reference filtering, keyword/FTS retrieval, and dependency-aware context assembly | Agents receive selected related concepts rather than an undifferentiated document collection |
| Portable, inspectable knowledge | Markdown/frontmatter bundles for local/Git, explicit links, and a common database concept model | People can inspect knowledge and retain a stable model across supported storage implementations |
| Practical adoption | Local, shared Git, and PostgreSQL complete stacks with embedded templates | Teams can start with a repository and choose shared infrastructure when needed |

These are enabling capabilities, not guarantees of model correctness, automatic compliance, or reduced cost. Their value comes from teams maintaining relevant knowledge and using review and context-selection workflows consistently.

## Product principles

### 1. Manage context with engineering discipline

Knowledge needs a clear identity, scope, lifecycle, relationships, and revision history. A proposed ADR should be distinguishable from an accepted one; an obsolete decision should link to its replacement. The knowledge collection is maintained alongside the product rather than treated as a one-time documentation export.

### 2. Give humans and agents the same canonical knowledge

Humans should be able to read and review what agents retrieve and change. Structured Markdown and explicit relationships make the collection understandable without depending on one chat history or agent-specific private memory format. The local and Git stacks keep concepts directly in OKF files; PostgreSQL represents the same public concept model in database records.

### 3. Use relationships and scope to improve task context

Project knowledge belongs to an explicit project; shared references can support several projects. `depends_on`, `implements`, `supersedes`, and `mitigates` make engineering intent navigable. Current suggestions combine scoped retrieval with dependency expansion and concept-count limits. More advanced semantic reasoning, authority arbitration, and intelligence pipelines remain product evolution rather than implicit retrieval guarantees.

### 4. Make review a workflow choice

Direct Git commits support autonomous knowledge maintenance where the team accepts that model. Initiative branches collect related changes across concepts and schemas for human PR review before merging to trunk. Branch selection belongs to the runtime session so parallel initiatives can remain independent. Gyrus supplies storage and session controls; humans own review/merge policy.

### 5. Keep contracts consistent across supported implementations

Core concept, lifecycle, relationship, and query contracts are separate from storage mechanics. The current supported choices are three complete stacks, not arbitrary provider combinations. This gives users clear operating models while keeping the platform's knowledge representation and interfaces coherent.

## Primary workflows

1. **Start an initiative:** Retrieve project knowledge and linked standards/decisions before planning a change.
2. **Capture intent:** Record requirements, alternatives, decisions, and guides as the work progresses; retain implementation details in source code.
3. **Review knowledge:** Validate concept structure and lifecycle, collect related Git changes on an initiative branch when review is required, and review them with the implementation.
4. **Reuse across boundaries:** A teammate or another agent harness retrieves the same qualified knowledge from a different repository/session.
5. **Maintain continuity:** Update living contracts and guides, supersede historical decisions when choices change, and preserve links explaining that change.

For example, a payments initiative can depend on shared authentication standards and an accepted storage decision. A billing initiative can reuse those same references without copying them into a second private memory store. Each initiative can propose its own guides or decision updates, with a human reviewing the combined knowledge diff before it becomes trunk knowledge.

## Deployment and operating value

| Choose | When it fits | Operating implications |
| --- | --- | --- |
| local | Knowledge is maintained with one code repository | No separate database is required for OKF query/index/graph behavior; version control is managed with the repository |
| git | Multiple repositories share a wiki and Git review/history | Gyrus authenticates to a remote and uses direct or runtime initiative commits; sessions coordinate through normal Git conflict checks |
| postgres | A team needs shared database storage and native full-text search | The team operates a PostgreSQL service and its credentials; storage/index/graph/FTS use the complete database stack |

Local use does not require an embedding service or hosted Gyrus service. Remote Git and PostgreSQL use the endpoints and access controls configured by the team. The host agent/model may have its own data handling requirements; choosing local knowledge storage alone does not establish that entire agent workflow's privacy properties.

## Position within an engineering toolchain

| Existing capability | Gyrus's complementary role |
| --- | --- |
| Code repositories and code intelligence | Retain the rationale, constraints, standards, and relationships surrounding implementation |
| Document authoring and wikis | Add qualified identity, concept lifecycle, typed links, and shared CLI/MCP access to curated engineering knowledge |
| Search and databases | Coordinate the concept model and workflow while using the selected backend for storage/retrieval |
| Agent skills and prompts | Give instructions a maintained knowledge collection to query and update across sessions |
| Agent-specific memory | Provide a shared, inspectable collection with explicit human/agent review choices |

The product's distinctive focus is the common engineering-knowledge workflow. It is not necessary to replace every existing authoring, code-search, model-hosting, or analytics tool to gain that value. Specific external-product comparisons need a separately dated assessment using current sources.

## Intelligence-platform direction

The [architecture blueprint](guide-001-system-architecture.md) extends this foundation toward hosted team MCP/UI, Lakebase-backed knowledge/application state, agent chat through a model gateway, analytical storage, and context-intelligence processing. The enterprise drawing separates interactive coordination from queued execution and downstream intelligence pipelines.

The strategic aim is a feedback loop: engineering work produces knowledge, maintained knowledge improves subsequent work, and evaluated intelligence helps teams identify missing relationships, stale context, and opportunities to improve their knowledge. Human validation and the authority of published knowledge need explicit policy as that loop becomes more automated.

Hosted Databricks applications, persistent caches/Redis coordination, distributed write processing, GraphRAG/evaluation pipelines, and governed promotion of automatically inferred knowledge are roadmap work. The present runtime does not implement the full organization/domain/system/task inheritance model or an automatic observation-to-approved-knowledge pipeline from earlier proposals. The [roadmap](prd-001-specification-roadmap.md) tracks the next steps.

## Success criteria

Evaluate the platform through concrete workflows: can users find the applicable project/reference knowledge, share it across harnesses, trace decisions through relationships, review an initiative's knowledge changes, and retrieve the approved result from another session? Validate consistency and failure behavior before attributing reasoning quality or business outcomes to the platform.

See [setup](guide-002-setup-onboarding.md), [configuration examples](tech-ref-002-config-schema.md), [documentation strategy](standards-001-doc-strategy.md), and [testing standards](standards-002-testing-strategy.md).
