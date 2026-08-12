---
id: prd-002-value-proposition-positioning
title: Gyrus Value Proposition & Strategic Product Positioning
category: technical
type: prd
format: ""
owner_group: armckinney
version: 3
status: active
tags:
  - product-positioning
  - value-proposition
  - competitive-analysis
  - context-control-plane
dependencies: []
---

# Gyrus Value Proposition & Strategic Product Positioning

## 1. Executive Summary

**Gyrus** is a lightweight, provider-neutral **Context Control Plane** and knowledge engine designed for software development teams and AI agentic tools. While alternative market solutions specialize narrowly in workplace search, source-code retrieval, data catalogs, persistent agent memory, or vector databases, Gyrus provides a unified coordination layer where humans and heterogeneous AI tools collaboratively create, govern, resolve, and distribute portable engineering context.

Gyrus is built around the **Open Knowledge Format (OKF)** standard. It acts as a coordination layer between agent clients, documentation platforms, code repositories, storage backends, search engines, and human review workflows—guaranteeing that organizational context remains discoverable, structured, governed, portable, human-readable, and reliable enough to influence critical engineering decisions.

---

## 2. Core Product Philosophy & Defining Pillars

### 2.1 Core Product Philosophy

1. **Context Should Be a Managed Engineering Asset:** Engineering context must be treated with the same discipline as source code and system configuration—possessing explicit ownership, defined scope, known authority, version history, validation rules, review workflows, lifecycle status, and traceable provenance.
2. **Humans and Agents Should Share the Same Canonical Knowledge:** Canonical knowledge must remain portable and human-readable (Markdown + YAML frontmatter contracts). Embeddings, vector indexes, full-text databases, and knowledge graphs are derived projections, not the canonical source of truth.
3. **Context Resolution Is More Than Search:** Traditional search answers *"which documents appear relevant?"* Gyrus context resolution answers *"which information is applicable, current, authoritative, and required for this task within our token budget?"*
4. **Agents Should Propose Knowledge, Not Silently Redefine It:** AI agents can discover patterns and propose updates, but high-authority context requires governed promotion (`Session Observation` ➔ `Candidate Memory` ➔ `Proposed Document` ➔ `Validated Review` ➔ `Published Knowledge`).
5. **The System Must Remain Provider-Neutral:** Git repositories, Confluence wikis, local filesystems, S3 blob stores, PostgreSQL databases, SQLite FTS5, and vector stores are replaceable storage and publishing providers. Gyrus owns the common domain model and context invariants.

### 2.2 Eight Defining Engineering Pillars

1. **Concept Data Contracts (OKF Schema Validation):** Validates structured software contracts (`adr`, `prd`, `specification`, `guide`, `standards`, `glossary`, `improvement-proposal`, `release-note`, `product`, `freeform`) rather than unstructured text dumps.
2. **Standardized Concept Metadata & Knowledge Graphs:** Metadata explicitly captures directional relationships (`depends_on`, `supersedes`, `implements`, `mitigates`). Gyrus parses these links and maintains a directed SQLite knowledge graph for deep edge traversals.
3. **Composable Service Providers:** Clean Go interface abstractions (`DocumentStore`, `IndexStore`, `GraphStore`, `SearchProvider`) decouple core domain rules from storage drivers, supporting zero-CGO local storage (`localfs` + SQLite FTS5) as well as enterprise cloud drivers (Git repos, AWS S3, Azure SA, PostgreSQL).
4. **Unified Memory Persistence:** Serves as a single source of truth across CLI tools (`gyrus`), stdio Model Context Protocol (MCP) servers, agent skills (Open Skill Format), `AGENTS.md`, and web applications.
5. **Composable Core SDK (`pkg/gyrus`):** Exposed as a pure, zero-dependency Go SDK package (`pkg/gyrus`), enabling developers to embed Gyrus directly into custom internal CLI tools, AI pipelines, background workers, or web app backends.
6. **Token-Budgeted Context Linearization (`gyrus suggest-context`):** Combines BM25 keyword relevance, graph depth traversal, and configurable token budget constraints to synthesize a single, non-redundant context payload.
7. **Lifecycle State Machine & Immutability Governance:** Enforces strict state transitions per document type (e.g. ADRs move `proposed` ➔ `accepted` ➔ `deprecated`/`superseded`) and runtime content immutability for decision logs (`immutable: true`).
8. **Embedded Templates (`go:embed`) & Custom Overrides:** Pre-packages 11 standard Markdown templates inside the compiled binary via `go:embed`, with support for custom project overrides via `.gyrus.yaml` (`schemas_path`).

---

## 3. Comprehensive Market Landscape Positioning & Categorized Comparisons

Gyrus is positioned across both **commercial marketplace products** and **architectural solution approaches**.

### 3.1 Market Category Classification

Alternatives in the AI context space generally specialize in one of six product categories:
1. **Governed Data & Semantic Definitions:** Atlan Context Engineering Studio
2. **Enterprise Workplace Search & Knowledge Platforms:** Glean, Notion AI
3. **Source Code Retrieval & Intelligence Engines:** Augment Context Engine, Sourcegraph
4. **Persistent Agent Memory Platforms:** Mem0
5. **Temporal Knowledge Graphs & Graph Infrastructure:** Zep, Graphiti, TrustGraph
6. **Lightweight Instruction & File Frameworks:** Agent Skills, Doc MCPs (e.g. Confluence MCP), Repo-Local Files (`.cursorrules`, `AGENTS.md`)

---

### 3.2 Master Market & Solution Comparison Matrix

| Solution / Product | Product Category / Design Center | Key Strengths | Gaps & Limitations | Why Choose Gyrus Context Control Plane Instead? |
| :--- | :--- | :--- | :--- | :--- |
| **Atlan Context Engineering Studio** | Governed Business & Data Context Layer | Versioned Context Repos, BI semantic models, data lineage, evaluation benchmarks | Specialized for data warehouse/catalog assets; lacks software architecture, ADRs, runbooks, and repo context | Choose Gyrus when context extends beyond data governance into architecture, ADRs, PRDs, code conventions, runbooks, and agent memory without requiring an enterprise data catalog foundation. |
| **Glean** | Enterprise Search & Workplace Knowledge Platform | Broad SaaS application connectors, enterprise permissions search, managed AI assistants | Managed SaaS platform; lacks context-as-code, repository ownership, portable Markdown artifacts, and offline capability | Choose Gyrus when you need context-as-code, repository ownership, portable Markdown artifacts, custom lifecycle rules, offline execution, and freedom from a managed workplace-AI SaaS platform. |
| **Augment Context Engine** | Source-Code Retrieval for Coding Agents | Deep semantic code indexing, cross-repository code retrieval, MCP integration | Focuses on selecting source code snippets; lacks artifact lifecycle, ADR governance, or team knowledge contracts | Choose Gyrus when context includes complete plans, decisions, standards, runbooks, and memories in addition to code selection. Augment retrieves code; Gyrus governs knowledge authority. |
| **Sourcegraph** | Large-Scale Code Intelligence & Navigation | Multi-repository code search, symbol navigation, AST syntax understanding | Focuses on code search & symbol definitions; does not manage authority or lifecycle of engineering decisions and specs | Choose Gyrus when code intelligence is only one provider within a broader knowledge ecosystem. Gyrus manages the lifecycle, authority, and contracts of knowledge rather than specializing in source-code AST parsing. |
| **Mem0** | Persistent Personal Agent Memory Platform | Long-term memory extraction across users, sessions, and chat applications | Breaks knowledge into unstructured facts/triples; loses narrative, human reviewability, and artifact contracts | Choose Gyrus when knowledge must remain coherent, reviewable, human-readable, and authoritative. Gyrus separates temporary memory observations from approved canonical knowledge artifacts. |
| **Zep / Graphiti** | Temporal Knowledge Graph Engine for Agents | Dynamic entity relationship modeling, changing facts, temporal retrieval | Derived graph is the primary store; harder for humans to directly inspect, review, or version-control in Git | Choose Gyrus when the knowledge graph should be a derived index rather than the canonical representation. Humans and agents work with portable artifacts while graph relationships support retrieval behind the scenes. |
| **TrustGraph & Graph Frameworks** | Self-Hosted Graph Construction Toolkit | Flexible infrastructure primitives for building custom graph context pipelines | Unopinionated toolkit; does not define artifact contracts, scope inheritance, lifecycle states, or review workflows | Choose Gyrus when you want an opinionated, ready-to-use context product rather than a complex infrastructure toolkit. Gyrus defines artifact types, scopes, lifecycle, validation, and review workflows. |
| **Documentation Platforms & MCPs** *(e.g. Confluence)* | Enterprise Documentation Platform & Wiki MCP | Human-readable wiki pages, permissions, version history, familiar editing | Page-centric retrieval; lacks task-specific context resolution, scope inheritance, or contract schemas | Choose Gyrus when Confluence should remain a human publishing surface without dictating the entire context architecture. Gyrus adds provider-neutral contracts, resolution, and governance. |
| **Agent Skills Only** *(e.g. `SKILL.md`)* | Instruction-Driven Behavioral Prompt Adapter | Portable instructions, easy to prototype, low operational overhead | Prompt instructions only; lacks persistent storage, stable identities, search index, authority rules, or enforcement | Choose Gyrus to convert recommended agent instructions into an enforceable Go core engine, storage persistence layer, and unified multi-agent context hub. |
| **Agent Skills + Doc MCP** | Instruction-Driven Doc Retrieval Workflow | Low-cost combination of behavioral guidance and wiki persistence | Relies on LLMs to consistently follow search & update instructions; lacks deterministic resolution or state machines | Choose Gyrus to eliminate instruction drift. Gyrus converts search & validation conventions into a deterministic Go context resolution engine (`suggest-context`). |
| **Repository-Local Files Only** *(e.g. `.cursorrules`, `AGENTS.md`)* | Repo-Local Static File Hints | Version-controlled, portable, close to code, reviewable via PRs | Fragmented per repo; lacks global discovery, cross-repo context inheritance, or central indexing | Choose Gyrus to combine local repository ownership with global context inheritance, search indexing, and unified multi-agent persistence. |
| **Search, Vector, or Graph DB Primitives** *(e.g. Pinecone, Chroma)* | Search & Retrieval Infrastructure Primitive | High-dimensional semantic similarity, vector indexing, scalable retrieval | Infrastructure primitives; lack artifact contracts, human-readable canonical sources, scope inheritance, or approval workflows | Choose Gyrus to use vector/search/graph databases as replaceable derived indexes beneath a governed, contract-native context control plane. |

---

### 3.3 Evaluation of Structural Solution Alternatives

Beyond commercial products, engineering organizations typically evaluate six structural architectural approaches. Gyrus fulfills the specific gaps of each:

#### Alternative 1: Agent Skills Only (`SKILL.md`)
* **Strengths:** Lightweight, portable, easy to prototype, zero infra overhead, compatible with multiple agent clients.
* **Gaps:** Provides prompt instructions, not shared state or enforcement. Lacks persistent storage, stable artifact IDs, search indexes, authority rules, scope inheritance, supersession, or review workflows.
* **Gyrus Fulfillment:** Gyrus uses Agent Skills as an integration transport, while providing the durable state, validation, retrieval, lifecycle, and governance capabilities beneath them.

#### Alternative 2: Documentation Platform or Documentation MCP (Confluence)
* **Strengths:** Human-readable pages, rich search, organization permissions, existing adoption, familiar editing.
* **Gaps:** Focuses on storing and retrieving wiki pages. Lacks task-specific context resolution, scope inheritance, repository-local distribution, agent memory promotion, contract schemas, or multi-agent uniformity.
* **Gyrus Fulfillment:** Gyrus uses documentation platforms as replaceable publishing providers while maintaining a provider-independent artifact model and context resolution layer.

#### Alternative 3: Agent Skills and Documentation MCP
* **Strengths:** Combines behavioral guidance with wiki persistence. Low cost, human-readable, good for early experimentation.
* **Gaps:** Relies on LLMs remembering to search, interpret, merge, and validate context correctly. Lacks deterministic context resolution, scope precedence, or state machine enforcement.
* **Gyrus Fulfillment:** Gyrus replaces prompt-driven conventions with deterministic system invariants (`context.resolve` / `gyrus suggest-context`), ensuring context is resolved consistently regardless of the calling LLM or client.

#### Alternative 4: Agent Memory Platform (Mem0)
* **Strengths:** Persistent memory across sessions, automated fact extraction, personalized user state.
* **Gaps:** Treats knowledge as isolated key-value facts or message episodes. Weakens narrative structure, authority, reviewability, and human comprehension of engineering architecture.
* **Gyrus Fulfillment:** Gyrus maintains coherent OKF Markdown documents as canonical sources of truth, using memory systems as derived indexes while enforcing a governed promotion workflow (`Observation` ➔ `Candidate Memory` ➔ `Published Document`).

#### Alternative 5: Repository-Local Files Only (`.cursorrules`, `AGENTS.md`)
* **Strengths:** Portable, version-controlled, close to code, reviewable in PRs, offline-ready.
* **Gaps:** Isolated per repository. Cannot handle organization-wide standards, cross-repo discovery, scope inheritance, or central search indexing.
* **Gyrus Fulfillment:** Gyrus treats repositories as first-class providers and materialization targets while resolving broader organizational context (`Organization` ➔ `Domain` ➔ `System` ➔ `Repository` ➔ `Task`).

#### Alternative 6: Build Directly on Search, Vector, or Knowledge Graph DBs (Pinecone, Chroma, Neo4j)
* **Strengths:** High-dimensional semantic search, flexible queries, scalable retrieval primitives.
* **Gaps:** Database primitives, not a complete context product. Lack artifact contracts, human-readable canonical representations, scope inheritance, approval workflows, or document lifecycle.
* **Gyrus Fulfillment:** Gyrus uses vector, graph, and full-text databases as replaceable indexing projections while maintaining Markdown contracts as the portable source of truth.

---

## 4. Operational, Security, Financial & Architectural Drivers

### 🔒 4.1 Data Privacy, Air-Gapped Security & Zero SaaS Exfiltration
* **Zero Third-Party Lock-In:** Cloud MCPs and SaaS vector databases require transmitting proprietary IP, internal architectural specs, and security vulnerability reports to external SaaS clouds.
* **Air-Gapped Readiness:** Gyrus runs 100% locally with zero required external API calls or telemetry. Memory indices stay strictly inside your enterprise security perimeter (local disk, private Git repos, or internal cloud storage).

### ⏳ 4.2 Deterministic Reproducibility & Contract Versioning
* **Version-Locked Memory:** External wikis and SaaS memory stores drift continuously out of band. Prompting an agent today vs. next month yields inconsistent code generation if someone edited an unversioned wiki page.
* **Standardized Contract Audit Trail:** Gyrus documents natively maintain envelope metadata (`version: 1`, `last_updated`, `last_modified_by`). Whether stored in Git, S3, or PostgreSQL, Gyrus provides deterministic contract versioning and auditability.

### 🔄 4.3 Multi-Agent & Multi-IDE Interoperability
* **Tool-Agnostic Context Hub:** Engineering teams use diverse AI tools (Cursor, Claude Code, GitHub Copilot, custom scripts).
* **Single Source of Truth:** Gyrus simultaneously exposes stdio MCP, agent skills (Open Skill Format), CLI subcommands, and Go SDK bindings so all developers and AI agents query the exact same context engine.

### 💰 4.4 Optional Zero-Infra Deployment & Scalable TCO
* **Mandatory Cloud Costs for Competitors:** Enterprise search platforms and vector DBs enforce database cluster provisioning, API usage fees, and per-seat SaaS subscriptions.
* **Optional Zero-Infra Mode:** Gyrus gives teams the choice to run completely zero-infra locally (`localfs` + embedded SQLite FTS5) with zero setup costs, OR compose enterprise cloud backends (S3, Postgres, Git repos) when scaling infrastructure.

---

## 5. Build-versus-Buy Rationale

Organizations should not rebuild commodity infrastructure already provided by specialized vendors.

### 5.1 What Existing Tools Handle
* **Enterprise Connectors:** Glean, Confluence, Jira APIs
* **Source Code Intelligence:** Sourcegraph, Augment, Tree-sitter
* **Vector Search & Embeddings:** pgvector, Pinecone, Chroma
* **Relational Storage:** PostgreSQL, SQLite

### 5.2 What Gyrus Provides (The Missing Coordination Layer)
* **Canonical Artifact Contracts:** Standardized OKF frontmatter schemas (`adr`, `prd`, `spec`)
* **Context Resolution Engine:** Determining applicable, authoritative context within token budgets (`context.resolve`)
* **Hierarchical Scope & Inheritance:** `Organization` ➔ `Domain` ➔ `System` ➔ `Repository` ➔ `Task`
* **Governed Promotion Workflows:** Converting temporary observations into approved organizational context
* **Lifecycle State Machines & Immutability:** Enforcing status transitions (`proposed` ➔ `accepted`) and `immutable: true` locks
* **Provider Neutrality:** Materializing context across local disk, Git, S3, and Postgres without changing API contracts
* **Multi-Tool Compatibility Adapters:** CLI, stdio MCP, Agent Skills, `AGENTS.md`, and Go SDK (`pkg/gyrus`)

---

## 6. Strategic Summary & Recommended Positioning

Gyrus should be positioned as:

> **The Context Control Plane that enables humans and heterogeneous AI agents to collaboratively create, discover, resolve, govern, and distribute portable engineering context.**

By coordinating existing storage and search technologies around an open, contract-native artifact model, Gyrus allows engineering organizations to plug any AI agent into their codebase once—giving it a reliable, governed way to understand existing architecture, follow team standards, record decisions, and share knowledge across both human teammates and AI tools.
