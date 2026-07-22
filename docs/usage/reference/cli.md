# Gyrus CLI Command Reference Manual

This manual provides detailed command signatures, flags, arguments, and exit code specifications for the standalone `gyrus` command-line executable.

---

## 1. Global Flags

The following flags apply to all `gyrus` CLI subcommands:

| Flag | Type | Description |
| :--- | :--- | :--- |
| `--storage-path <path>` | String | Path to storage root directory (overrides `GYRUS_STORAGE_PATH` env). |
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

### 1. `gyrus init`
Initializes Gyrus storage directory and configuration.

```bash
gyrus init [--storage-path <path>]
```

### 2. `gyrus create`
Creates a new Open Knowledge Format (OKF) contract document.

```bash
gyrus create \
  --id "<id>" \
  --title "<title>" \
  --category "<category>" \
  --type "<type>" \
  --owner-group "<owner_group>" \
  [--status "<status>"] \
  [--tags "tag1,tag2"] \
  [--dependencies "dep-1,dep-2"] \
  [--content "<markdown body>"] \
  [--content-file "<path-to-file>"]
```

### 3. `gyrus get <id>`
Retrieves an OKF document payload by ID.

```bash
# Output as raw Markdown
gyrus get adr-001-storage

# Output as JSON envelope
gyrus get adr-001-storage --json
```

### 4. `gyrus update <id>`
Updates an existing document fields or Markdown body content.

```bash
gyrus update adr-001-storage \
  [--title "<new-title>"] \
  [--status "<new-status>"] \
  [--content "<new-body>"] \
  [--expected-version 1]
```

### 5. `gyrus link <from-id> <to-id>`
Creates a directed relationship edge between two documents.

```bash
gyrus link adr-001-storage prd-002-context --rel-type "depends_on"
```

### 6. `gyrus unlink <from-id> <to-id>`
Removes a directed relationship edge between two documents.

```bash
gyrus unlink adr-001-storage prd-002-context --rel-type "depends_on"
```

### 7. `gyrus sync`
Re-indexes filesystem documents, updates SQLite FTS5 indexes, and extracts dependencies automatically.

```bash
gyrus sync [--storage-path <path>]
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

### 12. `gyrus mcp serve`
Starts the embedded Model Context Protocol (MCP) stdio server for GUI IDE integration.

```bash
gyrus mcp serve [--storage-path <path>]
```
