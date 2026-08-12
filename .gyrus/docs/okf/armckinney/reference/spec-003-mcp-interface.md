---
id: spec-003-mcp-interface
title: Embedded Model Context Protocol Server Specification
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Specification 03: MCP Interface Specification

The Gyrus MCP Server is launched locally via `gyrus mcp serve` and exposes standard resources, tools, and prompts to AI agent clients (such as Cursor, Claude Desktop, and VS Code Copilot).

---

## 1. Exposed MCP Tools

* **`memory.suggest_context`**
  * *Parameters:*
    * `request` (string, required): The coding task or prompt description.
    * `max_results` (integer, optional): Maximum document count (default: 5).
  * *Returns:* Linearized text summary of relevant context documents.

* **`memory.search`**
  * *Parameters:*
    * `query` (string, required): Search term.
    * `category` (string, optional): Filter by category.
    * `doc_type` (string, optional): Filter by document type.
    * `status` (string, optional): Filter by status.
    * `tag` (string, optional): Filter by tag.
  * *Returns:* List of matching document headers and metadata summaries.

* **`memory.get`**
  * *Parameters:*
    * `id` (string, required): Target document ID.
  * *Returns:* Full JSON envelope and Markdown content body.

* **`memory.create`**
  * *Parameters:*
    * `id` (string, required): Unique document ID.
    * `title` (string, required): Human-readable title.
    * `doc_type` (string, required): Document type (e.g. `adr`, `prd`, `specification`).
    * `category` (string, required): Category (`architecture`, `technical`, etc.).
    * `status` (string, required): Initial state (e.g. `proposed`, `draft`).
    * `owner_group` (string, required): Owning group (e.g. `engineering`).
    * `tags` (array[string], optional): List of tags.
    * `dependencies` (array[string], optional): List of dependent document IDs.
    * `body` (string, required): Markdown text body content.
  * *Returns:* Created document ID, version, and validation status.

* **`memory.update`**
  * *Parameters:*
    * `id` (string, required): Target document ID.
    * `expected_version` (integer, required): Expected current version for lock check.
    * `status` (string, optional): New status string (must be a valid transition).
    * `change_reason` (string, required): Audit log reason.
    * `body` (string, optional): Updated Markdown text body content.
  * *Returns:* Status and new version index.

* **`memory.link` / `memory.unlink`**
  * *Parameters:*
    * `from_id` (string, required): Source document ID.
    * `to_id` (string, required): Target document ID.
    * `relationship_type` (string, required): Enum: `[supersedes, depends_on, implements, mitigates]`.
  * *Returns:* Execution result status.

* **`memory.validate`**
  * *Parameters:*
    * `doc_type` (string, required): Document type.
    * `metadata` (object, required): OKF metadata dictionary.
    * `body` (string, required): Markdown body content.
  * *Returns:* Validity boolean and validation error messages.

---

## 2. Exposed MCP Resources

Agents can fetch documents or indexes directly via resource URIs:

* **`memory://doc/{id}`:** Reads the complete document content and frontmatter envelope.
* **`memory://schema/{document_type}`:** Retrieves the JSON Schema validation contract and template for a document type.
* **`memory://repo/context`:** Fetches the active repository context summary.
* **`memory://adr`:** Lists all active Architecture Decision Records.
* **`memory://prd`:** Lists all active Product Requirements Documents.
* **`memory://specification`:** Lists all active Technical Specifications.

---

## 3. Exposed MCP Prompts

Standard prompts exposed to align agent behaviors:

* **`prepare-adr`:** Prompts the agent to scaffold a new Architecture Decision Record using Gyrus templates.
* **`review-implementation-against-memory`:** Prompts the agent to compare its proposed code changes against retrieved Gyrus documents to catch constraint violations.
* **`propose-memory-update`:** Instructs the agent on how to structure a patch to update context when adding new features.
