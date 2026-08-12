---
id: guide-001-system-architecture
title: Gyrus Architecture Blueprint & Domain Topology
category: architecture
type: guide
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus System Architecture

This document details the architectural design, provider abstractions, filesystem layout, and state machine lifecycle engines of **Gyrus: Unified Context & Memory Engine**.

---

## 1. System Blueprint & Component Topology

Gyrus is architected as a modular single Go binary delivering four primary runtime layers:

```mermaid
graph LR
    subgraph Clients ["Callers & Clients"]
        CICD["CICD & Scripting"]
        AgentTools["Agentic Tool(s)"]
        Humans["Human Users"]
    end

    subgraph Interface ["Adapters & Application Surfaces"]
        Skill["Skill (Open Skill Format)"]
        CLI["Gyrus CLI"]
        MCP["Gyrus MCP Server"]
        Apps["Applications\ni.e. Serverless Content Navigator Web App\n- Implements AI Chatbot"]
    end

    subgraph Core ["Gyrus Core SDK"]
        CoreSDK["Gyrus Core SDK\n\n# Applies (MVP Provider):\n- Concept Storage Format (OKF)\n- Concept Index Provider (OKF)\n- Knowledge Graph Provider (OKF)\n- Search Provider (None)\n- Concept Persistence Service (localfs)\n- Concept Types (OKF)\n\n# Implements:\n- OKF Standard Metadata & Concepts"]
    end

    subgraph Services ["Core Services & Provider Drivers"]
        IndexSvc["Concept Index Service\n\nProviders:\n- OKF\n- SQLite\n- PostgreSQL"]
        GraphSvc["Knowledge Graph Service\n\nProviders:\n- OKF\n- SQLite\n- PostgreSQL"]
        SearchSvc["Search Service\n\nProviders:\n- SQLite FTS\n- PostgreSQL full-text search"]
        PersistSvc["Concept Persistence Service\n\nProviders:\n- localfs\n- SQLite\n- Git Repo (GitHub, Bitbucket)\n- Blob Storage (S3, Azure SA)\n- PostgreSQL"]
    end

    subgraph Storage ["Concept Storage Format"]
        StorageFmt["Concept Storage Format\n\nProviders:\n- OKF\n- Relational Table"]
    end

    %% Client and Adapter Relations
    CICD -->|invokes| CLI
    AgentTools -->|invokes| Skill
    Skill -->|invokes| CLI
    CLI -->|implements| CoreSDK
    AgentTools -->|OR invokes| MCP
    MCP -->|implements| CoreSDK
    Humans -->|invokes| Apps
    Apps -->|Implements| CoreSDK

    %% Core SDK Services Relations
    CoreSDK -->|instantiates| IndexSvc
    CoreSDK -->|instantiates| GraphSvc
    CoreSDK -->|instantiates| SearchSvc
    CoreSDK -->|instantiates| PersistSvc

    %% Persistence to Storage Format
    PersistSvc -->|instantiates| StorageFmt
```

### 🔄 Core SDK Subsystem Execution Flows

```mermaid
graph LR
    subgraph Ingestion ["Ingestion Flow (create / update / sync)"]
        direction LR
        P1["1. OKF Parser and Lifecycle"] -->|Validate Schema and State| P2["2. DocumentStore"]
        P2 -->|Persist Payload| P3["3. IndexStore and Search"]
        P3 -->|Index FTS5 and Extract Links| P4["4. GraphStore"]
    end
```

```mermaid
graph LR
    subgraph Retrieval ["Context Discovery Flow (suggest-context / search)"]
        direction LR
        Q1["1. SearchProvider (FTS5)"] -->|Lexical Match| Q2["2. GraphStore"]
        Q2 -->|Traverse Edges| Q3["3. DocumentStore"]
        Q3 -->|Hydrate Payload| Q4["4. Context Linearizer"]
    end
```
---

## 2. Core Components

### A. Core SDK (`pkg/gyrus`)
The public domain package exposing:
- `Document`: Struct holding OKF metadata fields (`id`, `title`, `category`, `type`, `format`, `owner_group`, `version`, `status`, `last_modified_by`, `last_updated`, `tags`, `dependencies`) and Markdown body content.
- `DocumentStore`: Interface for CRUD document operations.
- `IndexStore`: Interface for metadata indexing and search retrieval.
- `GraphStore`: Interface for document relationship edges (`depends_on`, `supersedes`, `implements`, `mitigates`).
- `SearchProvider`: Interface for full-text lexical and metadata filtering.

### B. Storage & Index Providers (`internal/provider`)
- **`localfs`:** Handles atomic file operations over local filesystem storage directories.
- **`sqlite`:** Uses `modernc.org/sqlite` (pure Go, CGO-free) to manage SQLite DDL migrations (`documents_index`, `documents_fts`, `document_edges`), FTS5 lexical queries, and edge traversals.

### C. Open Knowledge Format (OKF) Parser (`internal/okf`)
Extracts YAML frontmatter headers from Markdown files, enforces schema validation rules (such as ID regex `^[a-z0-9-_]+$`), and serializes documents back to Markdown.

### D. Lifecycle State Machine Engine (`internal/lifecycle`)
Enforces valid state transitions:
- **ADR (`adr`):** `proposed` ➔ `accepted` ➔ `superseded` / `deprecated`
- **Improvement Proposal (`improvement-proposal`):** `draft` ➔ `reviewing` ➔ `approved` ➔ `implemented`
- **General Documents (`prd`, `guide`, `specification`, etc.):** `draft` ➔ `active` ➔ `deprecated` ➔ `archived`

---

## 3. OKF Bundle Directory Topology

When storing documents locally, Gyrus arranges files into a structured directory hierarchy under the storage root:

```text
<storage-root>/
├── config.yaml
├── index.db
└── okf/
    └── <owner_group>/ (Security boundary)
          ├── reference/ (Global ref docs: ADRs, Standards, Specs)
          │     ├── adr-001-storage.md
          │     └── spec-002-schema.md
          └── workspaces/<repo-x>/ (Repo-scoped context)
                └── prd-003-context-engine.md
```

---

## 4. Storage Path Precedence Hierarchy

Gyrus resolves its root storage location in the following order of precedence:

1. **`--storage-path <path>`** CLI flag argument.
2. **`GYRUS_STORAGE_PATH`** environment variable.
3. **Repository Project Config File:** `.gyrus.yaml`, `.gyrus.yml`, `.gyrus/config.yaml`, or `.gyrus/config.yml` in current working directory or any parent repository directory.
4. **User Home Config File:** `~/.config/gyrus/config.yaml`, `~/.config/gyrus/config.yml`, `~/.gyrus.yaml`, or `~/.gyrus.yml`.
5. **`~/.gyrus/`** default application directory.
