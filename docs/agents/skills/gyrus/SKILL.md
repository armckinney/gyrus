---
name: gyrus
description: Gyrus Unified Context & Memory Engine agent skill. Use to search, retrieve, create, update, link, and suggest relevant OKF codebase context for tasks.
---

# Gyrus Agent Skill Specification & CLI Reference

This skill equips AI agents to interact directly with Gyrus codebase memory via the `gyrus` CLI executable.

---

## 💡 Core Agent Guidelines

1. **Before Modifying Code:** Always run `gyrus suggest-context --prompt "<task description>"` or `gyrus search --query "<keyword>"` to read relevant ADRs and technical contracts.
2. **Machine Parsing:** Pass global `--json` flag to receive structured JSON envelopes instead of terminal formatted text.
3. **ID Naming Rule:** Document IDs MUST match the lower-case alphanumeric pattern `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
4. **Exit Codes Protocol:**
   - `0`: Success
   - `1`: Frontmatter schema or ID pattern validation error
   - `2`: Illegal state machine transition error (e.g., `accepted` ➔ `proposed`)
   - `3`: Unauthorized / owner group permission error
   - `4`: Optimistic concurrency lock conflict (`--expected-version` mismatch)
   - `5`: File/Record not found or storage read/write error

---

## 📋 Allowed Document Enums

- **Categories (`--category`):** `architecture`, `business-logic`, `product`, `operations`, `technical`
- **Document Types (`--type`):** `adr`, `prd`, `guide`, `specification`, `standards`, `technical-reference`, `product`, `release-note`, `improvement-proposal`, `glossary`, `freeform`
- **Relationship Types (`--rel-type`):** `depends_on`, `supersedes`, `implements`, `mitigates`

---

## 🛠️ Complete CLI Command Reference

### 1. `gyrus suggest-context`
Linearizes top relevant documents matching a task prompt (Recommended first step).

```bash
gyrus suggest-context --prompt "<task or problem description>" [--max-tokens 4000] [--json]
```

### 2. `gyrus search`
Executes FTS5 lexical keyword search across documents.

```bash
gyrus search --query "<keyword>" [--category "<category>"] [--type "<type>"] [--status "<status>"] [--tag "<tag>"] [--max-results 10] [--json]
```

### 3. `gyrus get`
Retrieves a single document by ID.

```bash
gyrus get <document-id> [--json]
```

### 4. `gyrus create`
Creates a new OKF contract document.

```bash
gyrus create \
  --id "<id>" \
  --title "<title>" \
  --category "<category>" \
  --type "<type>" \
  --owner-group "<owner_group>" \
  [--status "<draft|proposed|active>"] \
  [--tags "tag1,tag2"] \
  [--dependencies "dep-id-1,dep-id-2"] \
  [--content "<markdown body>"] \
  [--content-file "<path/to/file.md>"] \
  [--json]
```

### 5. `gyrus update`
Patches metadata fields or content of an existing document.

```bash
gyrus update <id> \
  [--title "<new-title>"] \
  [--status "<new-status>"] \
  [--content "<new-markdown-body>"] \
  [--expected-version <current-version>] \
  [--json]
```

### 6. `gyrus link` / `gyrus unlink`
Creates or removes a directed relationship edge between two documents.

```bash
# Create edge
gyrus link <from-id> <to-id> [--rel-type "depends_on|supersedes|implements|mitigates"]

# Delete edge
gyrus unlink <from-id> <to-id> [--rel-type "depends_on"]
```

### 7. `gyrus sync`
Re-indexes filesystem documents, updates SQLite FTS5 indexes, and extracts dependency links.

```bash
gyrus sync [--json]
```

### 8. `gyrus validate`
Validates an OKF Markdown file or JSON envelope schema without saving changes.

```bash
gyrus validate <path-to-file.md>
```

### 9. `gyrus schema`
Prints structural frontmatter template for a document type.

```bash
gyrus schema <doc-type>
```
