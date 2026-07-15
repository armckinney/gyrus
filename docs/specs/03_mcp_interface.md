# Specification 03: MCP Interface Specification

The Gyrus MCP Server is launched locally via `gyrus mcp serve` and exposes standard resources, tools, and prompts to AI agent clients.

---

## 1. Exposed MCP Tools

* **`memory.suggest_context`**
  * *Parameters:*
    * `request` (string, required): The coding task or user request prompt.
    * `max_results` (integer, optional): Maximum document limit.
  * *Returns:* Summary of relevant document IDs and their text payloads.
* **`memory.search`**
  * *Parameters:*
    * `query` (string, required): Keyword/lexical query string.
  * *Returns:* List of matching document headers and summaries.
* **`memory.get`**
  * *Parameters:*
    * `id` (string, required): The document ID.
  * *Returns:* Raw content and metadata payload.
* **`memory.create`**
  * *Parameters:*
    * `doc_type` (string, required): e.g. `adr`, `prd`.
    * `title` (string, required): Document title.
    * `metadata` (object, required): Envelope tags and categories.
    * `body` (string, required): Markdown content.
  * *Returns:* Target document ID and validation status.
* **`memory.update`**
  * *Parameters:*
    * `id` (string, required): Target document ID.
    * `expected_version` (integer, required): For optimistic concurrency checks.
    * `patch` (object, required): Fields to modify.
    * `change_reason` (string, required): Audit log justification.
  * *Returns:* Status and new version index.
* **`memory.validate`**
  * *Parameters:*
    * `doc_type` (string, required): target type.
    * `metadata` (object, required): Envelope tags.
    * `body` (string, required): Content block.
  * *Returns:* Boolean validity and error details.

---

## 2. Exposed MCP Resources

Agents can fetch specific documents or index files directly using resource URIs:

* **`memory://doc/{id}`:** Fetches the full document text.
* **`memory://schema/{document_type}`:** Retrieves the JSON Schema validation contract.
* **`memory://repo/context`:** Fetches the local repository's active context summary.
* **`memory://adr`:** Lists all active Architecture Decision Records.
* **`memory://proposal`:** Lists all active Improvement Proposals.
* **`memory://runbook`:** Lists all operations and setup runbooks.

---

## 3. Exposed MCP Prompts

Exposes standard prompts to align agent behaviors:

* **`prepare-adr`:** Guides the agent to scaffold a new Architecture Decision Record using Gyrus templates.
* **`review-implementation-against-memory`:** Prompts the agent to compare its proposed code changes against the retrieved Gyrus context documents to catch constraint violations.
* **`propose-memory-update`:** Instructs the agent on how to structure a patch to update the context library when implementing new features.
