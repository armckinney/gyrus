# Specification 01: JSON Contract Schema

Every document in Gyrus is stored as a structured JSON object (or represented as Markdown with a YAML frontmatter envelope). The Gyrus Core Engine validates all documents before saving them to prevent schema decay and formatting drift.

---

## 1. JSON Contract Envelope

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

---

## 2. Metadata Schema Validation Rules

The Gyrus Core Engine enforces the following validation boundaries:

* **`id` (string, required):** Unique alphanumeric ID identifying the document (e.g. `adr-2026-001`). Matches regex: `^[a-z0-9-_]+$`.
* **`title` (string, required):** A short human-readable title.
* **`category` (string, required):** Broad category. Enforced enums: `[architecture, business-logic, product, operations, technical]`.
* **`type` (string, required):** Specific document type mapping to a structural CLI template:
  * `adr`: Architecture Decision Record.
  * `prd`: Product Requirements Document (PRD).
  * `guide`: Onboarding tutorials, setup steps, or user guides.
  * `improvement-proposal`: Technical plans, design proposals, and implementation specs.
  * `release-note`: Release documentation, version changes, and deprecation notices.
  * `specification`: Technical specifications, architecture layout, protocols, or component designs.
  * `standards`: Process, workflow, and engineering practice standards.
  * `technical-reference`: API endpoints, CLI syntax, configuration tables, or integration lookup reference.
  * `product`: Product overview, landing pages, and documentation hubs.
  * `glossary`: Defined business/technical terms and system vocabulary.
  * `freeform`: Catch-all for general unstructured engineering context.
* **`format` (string, required):** Payload format. Enforced enums: `[markdown, json, yaml]`.
* **`owner_group` (string, required):** Identifies the team owning the document. Used for access filters.
* **`version` (integer, required):** Auto-incrementing revision index. Used for optimistic locking.
* **`status` (string, required):** State machine lifecycle status (e.g. `proposed`, `accepted`, `superseded`). Validated by Gyrus engine transition filters.
* **`last_modified_by` (string, auto-populated):** Author username/agent identifier.
* **`last_updated` (string, auto-populated):** ISO 8601 server/system timestamp.
* **`tags` (array of strings, optional):** Descriptive tags.
* **`dependencies` (array of strings, optional):** List of document IDs this file references.
* **`content` (string, required):** Freeform payload containing raw document text (e.g. Markdown).
