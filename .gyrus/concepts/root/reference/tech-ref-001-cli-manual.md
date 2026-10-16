---
title: Gyrus CLI Subcommand Reference Manual
id: tech-ref-001-cli-manual
type: technical-reference
format: markdown
category: technical
version: 5
status: active
last_updated: 2026-10-16T18:59:08Z
last_modified_by: ""
tenant: root
---

# Gyrus CLI Command Reference Manual

This manual provides detailed command signatures, flags, arguments, and exit code specifications for the standalone `gyrus` command-line executable.

---

## 1. Global Flags

The following flags apply to all `gyrus` CLI subcommands:

| Flag | Type | Description |
| :--- | :--- | :--- |
| `--json` | Boolean | Output command results as formatted JSON payloads. |
| `--verbose` | Boolean | Enable verbose debug logging output. |
| `-h`, `--help` | Boolean | Display help text for any command. |

---

## 2. Programmatic Exit Code Table

Gyrus CLI commands return deterministic exit codes so shell scripts and AI agents can handle and self-heal errors automatically:

| Exit Code | Constant Name | Cause & Description |
| :---: | :--- | :--- |
| **`0`** | `ExitSuccess` | Command completed successfully with zero errors. |
| **`1`** | `ExitValidationError` | Frontmatter schema validation or ID regex pattern violation (`^[a-z0-9-_]+$`). |
| **`2`** | `ExitTransitionError` | Illegal lifecycle state transition attempt (e.g. `accepted` ➔ `proposed`). |
| **`3`** | `ExitAuthError` | Unauthorized access or owner group permission restriction. |
| **`4`** | `ExitConcurrencyError` | Concurrency lock mismatch (`--expected-version` does not match current version). |
| **`5`** | `ExitStorageError` | Storage read/write error or missing file/database record. |

---

## 3. Subcommand Reference

### 1. `gyrus config init` & `gyrus config show`
Initializes Gyrus workspace or global configuration, and displays resolved active settings.

```bash
# Initialize workspace configuration (.gyrus.yaml in current repository)
gyrus config init [-p <profile>] [-o <owner_group>] [-f]

# Initialize global user configuration (~/.gyrus.yaml across all repositories)
gyrus config init -g [-p <profile>] [-o <owner_group>] [-f]

# Display resolved configuration and active storage root
gyrus config show [--json]
```

### 2. `gyrus create`
Creates a new Open Knowledge Format (OKF) contract document.

```bash
gyrus create \
  --id "<id>" \
  --title "<title>" \
  --category "<category>" \
  --type "<type>" \
  --tenant "<tenant>" \
  [--status "<status>"] \
  [--tags "tag1,tag2"] \
  [--content "<markdown body>"] \
  [--content-file "<path-to-file>"]
```

### 3. `gyrus get <id>`
Retrieves an OKF document payload by ID.

```bash
# Output as raw Markdown
gyrus get root/reference/adr-001-storage

# Output as JSON envelope
gyrus get root/reference/adr-001-storage --json
```

### 4. `gyrus update <id>`
Updates an existing document fields or Markdown body content.

```bash
gyrus update root/reference/adr-001-storage \
  [--title "<new-title>"] \
  [--status "<new-status>"] \
  [--content "<new-body>"] \
  [--expected-version 1]
```

### 5. `gyrus link <from-id> <to-id>`
Creates a directed relationship edge between two documents.

```bash
gyrus link root/reference/adr-001-storage root/reference/prd-002-context --rel-type "depends_on"
```

### 6. `gyrus unlink <from-id> <to-id>`
Removes a directed relationship edge between two documents.

```bash
gyrus unlink root/reference/adr-001-storage root/reference/prd-002-context --rel-type "depends_on"
```

### 7. `gyrus archive <id>`
Archives (deletes) a document payload from storage and search index.

```bash
gyrus archive root/reference/adr-001-storage [--json]
```

### 7. `gyrus sync`
Validates the authoritative collection or rebuilds the selected database index.

```bash
gyrus sync
```

### 8. `gyrus validate <file-path>`
Validates an OKF Markdown file or JSON envelope schema without saving.

```bash
gyrus validate docs/gyrus/okf/platform/reference/adr-001.md
```

### 9. `gyrus search`
Executes an FTS5 lexical keyword search query over documents.

```bash
gyrus search \
  [--query "<search-text>"] \
  [--category "<category>"] \
  [--type "<type>"] \
  [--status "<status>"] \
  [--tag "<tag>"] \
  [--max-results 10]
```

### 10. `gyrus suggest-context`
Linearizes top context documents matching an agent prompt context.

```bash
gyrus suggest-context --prompt "How is full-text search implemented?" [--max-tokens 4000]
```

### 11. `gyrus schema <doc-type>`
Prints structural frontmatter schema and template for a document type.

```bash
gyrus schema adr
```

### 12. `gyrus config show`
Displays the resolved configuration hierarchy, source metadata, and resolved storage root.

```bash
gyrus config show [--json]
```

### 13. `gyrus mcp serve`
Starts the embedded Model Context Protocol (MCP) stdio server for GUI IDE integration.

```bash
gyrus mcp serve
```

## OKF provider retrieval and maintenance

File-backed profiles now generate OKF index, graph, and search. The same defaults apply when secondary provider fields are omitted. `gyrus config show` displays effective selections. Explicit database/vector providers retain their search semantics.

OKF keyword search returns the top ten matches; `gyrus search --query "" --scope all --json` browses the complete matching metadata collection. Keyword terms use case-insensitive Unicode substring matching with OR semantics; FTS expression syntax and stemming are not interpreted by OKF.

Link/unlink persist source-owned {concept, rel} links with qualified concept names. Missing targets remain valid. Neighbors/traversal operate on persisted relationships. Version/Git/object-precondition conflicts map to exit code 4. Sync scans and validates the authoritative provider; IndexedFiles counts validated concepts. Invalid-record reports remain visible in text/JSON and produce validation exit code 1.

Before switching a SQLite graph, use the packaged gyrus-migrate-sqlite-to-okf skill to back up and export database-only edges, prove qualified tuple parity, change all three provider fields, and restart MCP processes. There is no gyrus migrate command, and sync alone cannot export database-only edges.

## Qualified-name arguments

Create --id remains a local slug; tenant, scope, and project determine its canonical name. Existing-concept get/update/archive/link/unlink arguments must be fully qualified: root/reference/adr-example or root/projects/gyrus/ticket-003. There is no short-ID fallback. JSON concept envelopes and mutation references include name and local id. Text mutation/search output identifies concepts by name. Creation and update no longer accept --dependencies; manage relationships with link/unlink. The --rel-type flag retains its name; persisted link fields are concept and rel.
