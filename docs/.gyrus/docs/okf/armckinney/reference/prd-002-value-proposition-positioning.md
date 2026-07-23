---
id: prd-002-value-proposition-positioning
title: Gyrus Value Proposition & Strategic Product Positioning
category: technical
type: prd
format: ""
owner_group: armckinney
version: 2
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

**Gyrus** is a lightweight, provider-neutral **Context Control Plane** and knowledge graph engine designed for software development teams and AI agentic tools. While alternative solutions focus narrowly on workplace search, code retrieval, data governance catalogs, or personal agent memory, Gyrus provides a unified layer where humans and heterogeneous AI tools collaboratively create, govern, resolve, and distribute portable engineering context.

Gyrus is built around the **Open Knowledge Format (OKF)** standard. It acts as a coordination layer between agent clients, documentation platforms, code repositories, storage backends, and human review workflows—guaranteeing that organizational context remains discoverable, structured, governed, portable, human-readable, and reliable enough to influence critical engineering decisions.

---

## 2. Core Product Philosophy & Defining Pillars

### 2.1 Core Product Philosophy

1. **Context Should Be a Managed Engineering Asset:** Engineering context must be treated with the same discipline as source code and infrastructure configuration—possessing explicit ownership, defined scope, known authority, version history, validation rules, review workflows, lifecycle status, and traceable provenance.
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

## 3. Market Landscape Positioning & Competitor Differentiation

Gyrus is designed for a distinct purpose in the AI infrastructure ecosystem. The table and analysis below position Gyrus against products across the entire context market:

### 3.1 Comprehensive Market Comparison Matrix

| Alternative Solution | Primary Design Center | Key Strengths | Why Choose Gyrus Context Control Plane Instead? |
| :--- | :--- | :--- | :--- |
| **Atlan Context Engineering Studio** | Governed business and data context | Versioned Context Repos, semantic models, data lineage, BI integration | Choose Gyrus when context extends beyond data governance into architecture, ADRs, PRDs, code conventions, runbooks, and agent memory without requiring an enterprise data catalog foundation. |
| **Glean** | Enterprise workplace search & agent platform | Broad application connectors, enterprise permissions search, managed assistants | Choose Gyrus when you need context-as-code, repository ownership, portable Markdown artifacts, custom lifecycle rules, offline execution, and freedom from a managed workplace-AI SaaS platform. |
| **Augment Context Engine** | Code retrieval for software agents | Deep semantic code indexing, cross-repository code retrieval, MCP integration | Choose Gyrus when context includes complete plans, decisions, standards, runbooks, and memories in addition to code selection. Augment retrieves code; Gyrus governs knowledge authority. |
| **Sourcegraph** | Code intelligence and software navigation | Large-scale code search, symbol navigation, multi-repository code retrieval | Choose Gyrus when code intelligence is only one provider within a broader knowledge ecosystem. Gyrus manages the lifecycle, authority, and contracts of knowledge rather than specializing in source-code AST parsing. |
| **Mem0** | Persistent personal agent memory | Long-term memory across users, sessions, and chat applications | Choose Gyrus when knowledge must remain coherent, reviewable, human-readable, and authoritative. Gyrus separates temporary memory observations from approved canonical knowledge artifacts. |
| **Zep / Graphiti** | Temporal knowledge graphs for agents | Entity relationships, changing facts, provenance, temporal retrieval | Choose Gyrus when the knowledge graph should be a derived index rather than the canonical representation. Humans and agents work with portable artifacts while graph relationships support retrieval behind the scenes. |
| **TrustGraph & Graph Frameworks** | Self-hosted graph ingestion & retrieval toolkits | Flexible infrastructure for building custom graph-based context systems | Choose Gyrus when you want an opinionated, ready-to-use context product rather than a complex infrastructure toolkit. Gyrus defines artifact types, scopes, lifecycle, validation, and review workflows. |
| **Skills + Documentation MCP** *(e.g. Confluence MCP)* | Agent-guided documentation workflows | Low complexity, familiar human interface, reusable prompt instructions | Choose Gyrus when guidelines must become enforceable behavior. Gyrus adds stable artifact identities, contract validation, scope inheritance, state machines, and task-specific context resolution. |
| **Repository-Local Files Only** *(e.g. `.cursorrules`, `AGENTS.md`)* | Local, version-controlled file hints | Portable, close to the code, reviewable via PRs, zero latency | Choose Gyrus when context spans multiple repositories or requires global-to-local inheritance. Gyrus combines local repository ownership with global context inheritance, search indexing, and multi-agent persistence. |

---

### 3.2 Deep-Dive Competitor Analysis

#### A. Gyrus vs. Data & Enterprise Search Platforms (Atlan, Glean)
* **Atlan Context Engineering Studio:** Atlan creates versioned Context Repos for data catalogs, BI metrics, and data lineage. However, it is fundamentally tied to data warehouse governance. Gyrus operates as a developer-first context-as-code control plane covering software engineering architecture, implementation plans, ADRs, runbooks, and repository conventions.
* **Glean:** Glean provides a centralized workplace AI platform with application connectors across Google Drive, Slack, and Jira. While powerful for corporate workplace search, Glean requires sending organizational data to a managed SaaS platform. Gyrus provides git-native, offline-ready, air-gapped context control that developers check directly into repositories or local storage.

#### B. Gyrus vs. Code Retrieval & Code Intelligence Systems (Augment, Sourcegraph)
* **Augment Code Engine & Sourcegraph:** Augment and Sourcegraph specialize in deep source-code indexing, AST symbol resolution, and multi-repo code snippet retrieval for coding LLMs. Gyrus complements these tools by serving as the higher-level knowledge authority—governing *why* decisions were made (ADRs), *what* requirements exist (PRDs), and *how* systems must be built (Standards), leaving code intelligence engines to handle raw source-code lookup.

#### C. Gyrus vs. Agent Memory Platforms & Temporal Graphs (Mem0, Zep, Graphiti)
* **Mem0 & Zep/Graphiti:** Memory platforms extract unstructured key-value facts or temporal graph edges from user chat sessions. However, breaking complex engineering architecture into isolated graph triples loses narrative, authority, and human reviewability. Gyrus maintains coherent, human-readable OKF Markdown documents as the canonical source of truth, deriving graph edges (`depends_on`, `implements`) and memory promotions through governed workflows.

#### D. Gyrus vs. Lightweight Agent Skills & Doc MCPs (`.cursorrules`, Confluence MCP)
* **Raw Agent Skills:** Static prompt files (e.g. `.cursorrules`) rely entirely on LLMs remembering to read unindexed files, resulting in high prompt token bloat and unvalidated edits.
* **Generic Doc MCPs (e.g. Confluence):** Confluence MCPs allow searching wiki pages, but lack task-specific context resolution, scope inheritance, and contract validation.
* **The Gyrus Synergy:** Gyrus uses Agent Skills and MCP as runtime adapters, but enforces context resolution (`suggest-context`), contract validation (`okf.Validate`), and state machines in a pure, fast Go engine.

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
