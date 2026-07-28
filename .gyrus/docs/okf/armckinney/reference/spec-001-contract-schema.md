---
id: spec-001-contract-schema
title: Open Contract Schema & Frontmatter Validation Rules
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Specification 01: Open Knowledge Format (OKF) Contract Schema

Every document managed by Gyrus adheres to the **Open Knowledge Format (OKF)**—a structured envelope schema stored as JSON or represented as a Markdown file with a YAML frontmatter header. The Gyrus Core Engine validates all documents against this schema before persisting them to prevent formatting decay and schema drift.

---

## 1. JSON & YAML Frontmatter Contract Envelopes

### A. JSON Representation (API / Database Envelope)
```json
{
  "id": "adr-2026-001",
  "title": "Use SQLite Edge Tables for Gyrus Relationships",
  "category": "architecture",
  "type": "adr",
  "format": "markdown",
  "owner_group": "platform-engineering",
  "version": 1,
  "status": "accepted",
  "last_modified_by": "developer-joe",
  "last_updated": "2026-07-14T14:20:00Z",
  "tags": ["storage", "sqlite", "relationships"],
  "dependencies": ["prd-context-manager"],
  "content": "# Use SQLite Edge Tables for Gyrus Relationships\n\n## Context\nTo support fast multi-hop traversals without the overhead of Neo4j..."
}
```

### B. Markdown with YAML Frontmatter Representation (Local Storage)
```markdown
---
id: adr-2026-001
title: Use SQLite Edge Tables for Gyrus Relationships
category: architecture
type: adr
format: markdown
owner_group: platform-engineering
version: 1
status: accepted
last_modified_by: developer-joe
last_updated: 2026-07-14T14:20:00Z
tags:
  - storage
  - sqlite
  - relationships
dependencies:
  - prd-context-manager
---

# Use SQLite Edge Tables for Gyrus Relationships

## Context
To support fast multi-hop traversals without the overhead of Neo4j...
```

---

## 2. Metadata Schema Validation Rules

The Gyrus Core Engine enforces strict validation boundaries for all OKF metadata attributes:

| Attribute | Data Type | Required | Enforced Rules & Pattern Constraints |
| :--- | :--- | :--- | :--- |
| **`id`** | `string` | **Yes** | Alphanumeric identifier matching regex `^[a-z0-9-_]+$`. Must be unique across the context library. |
| **`title`** | `string` | **Yes** | Human-readable title (1 to 200 characters). |
| **`category`** | `string` | **Yes** | Broad category classification. Enforced enum: `[architecture, business-logic, product, operations, technical]`. |
| **`type`** | `string` | **Yes** | Specific document type. Enforced enum: `[adr, prd, guide, improvement-proposal, release-note, specification, standards, technical-reference, product, glossary, freeform]`. |
| **`format`** | `string` | **Yes** | Content payload encoding. Enforced enum: `[markdown, json, yaml]`. Default: `markdown`. |
| **`owner_group`** | `string` | **Yes** | Owning team/group identifier for multi-tenant access control (e.g. `engineering`). |
| **`version`** | `integer` | **Yes** | Auto-incrementing revision index (>= 1). Used for optimistic concurrency control. |
| **`status`** | `string` | **Yes** | Lifecycle state. Must conform to legal transitions for the document's `type` (see Spec 04). |
| **`last_modified_by`** | `string` | **Auto** | Username or agent identity that authored the latest update. |
| **`last_updated`** | `string` | **Auto** | ISO 8601 timestamp (`YYYY-MM-DDTHH:MM:SSZ`) generated at write execution. |
| **`tags`** | `array[string]`| No | List of descriptive keyword tags for search indexing and categorization. |
| **`dependencies`** | `array[string]`| No | List of target document IDs referenced by this file. Used to build edge graphs in OKF mode. |
| **`content`** | `string` | **Yes** | Main body payload. Text string containing raw Markdown or structured content. |

---

## 3. OKF Bundle Directory Topology

When Gyrus operates in **Open Knowledge Format (OKF)** storage mode, documents are organized in an opinionated, multi-layered directory bundle structure:

```text
okf/                            <-- Flexible top-level directory root
└── <team>/                     <-- Security and tenant boundary (N-layers)
      ├── reference/            <-- Global reference documents shared across projects
      │     ├── adrs/           <-- Global Architecture Decision Records (e.g. adr-001.md)
      │     ├── standards/      <-- Global engineering and workflow standards
      │     ├── specifications/ <-- System specifications and core contracts
      │     └── prds/           <-- Global product requirement documents
      │
      └── workspaces/           <-- Repository and project-scoped context
            └── <repo-x>/       <-- Specific workspace instructions (e.g. repo-x specs, runbooks)
                  ├── context.md
                  └── notes/
```

### Key Topology Principles:
1. **Security & Ownership Boundary:** The `<team>` subfolder isolates access groups (matching `owner_group` metadata).
2. **Global Reference Layer (`reference/`):** Contains company-wide or organization-wide specifications, standards, and ADRs.
3. **Workspace Scoping (`workspaces/<repo-x>/`):** Holds project-specific context files that link back to global reference documents via `dependencies: [...]` frontmatter references.
