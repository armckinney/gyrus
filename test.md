# Gyrus Local Testing Commands Walkthrough

Copy and paste these commands sequentially into your terminal to test every subcommand and feature of Gyrus.

---

## 1. Setup & Initialization

Rebuild the latest binary and initialize workspace storage:

```bash
# Rebuild binary
make build

# Initialize Gyrus storage
./gyrus init
```

---

## 2. Document Schema Inspection (`gyrus schema`)

Inspect the frontmatter template for ADRs and PRDs:

```bash
# Print ADR schema template
./gyrus schema adr

# Print PRD schema template
./gyrus schema prd
```

---

## 3. Creating Documents (`gyrus create`)

Create sample architecture and product contract documents:

```bash
# Create Architecture Design Record (ADR)
./gyrus create \
  --id "adr-001-storage-engine" \
  --title "Use Embedded SQLite FTS5 for Gyrus Search" \
  --category "architecture" \
  --type "adr" \
  --owner-group "platform" \
  --status "accepted" \
  --tags "sqlite,fts5,cgo-free,architecture" \
  --content "## Context & Problem Statement
Gyrus requires zero-dependency embedded full-text search (FTS5) and edge graph storage.

## Decision
We selected modernc.org/sqlite as the pure Go driver engine.

## Consequences
- Zero CGO dependencies for cross-platform releases.
- Native FTS5 lexical search."

# Create Technical Specification
./gyrus create \
  --id "spec-002-cli-framework" \
  --title "Cobra CLI Command Routing Architecture" \
  --category "technical" \
  --type "specification" \
  --owner-group "platform" \
  --status "active" \
  --tags "cli,cobra,exit-codes" \
  --dependencies "adr-001-storage-engine" \
  --content "## Specification
Defines CLI subcommands and programmatic exit code mappings (0..5)."
```

---

## 4. Reading Documents (`gyrus get`)

Retrieve raw Markdown or structured JSON envelopes:

```bash
# Fetch raw Markdown payload
./gyrus get adr-001-storage-engine

# Fetch structured JSON envelope
./gyrus get adr-001-storage-engine --json
```

---

## 5. Updating Documents (`gyrus update`)

Patch document title, status, or body content with version lock:

```bash
# Update document status and content
./gyrus update adr-001-storage-engine \
  --title "Use Embedded SQLite FTS5 Engine (Updated)" \
  --status "active" \
  --expected-version 1
```

---

## 6. Linking Document Dependencies (`gyrus link` / `gyrus unlink`)

Manage directed relationship edges between documents:

```bash
# Link spec-002-cli-framework -> adr-001-storage-engine
./gyrus link spec-002-cli-framework adr-001-storage-engine --rel-type "depends_on"

# Unlink relationship
./gyrus unlink spec-002-cli-framework adr-001-storage-engine --rel-type "depends_on"
```

---

## 7. Validating Documents (`gyrus validate`)

Validate OKF frontmatter schema rules on disk without saving:

```bash
./gyrus validate .gyrus/docs/okf/platform/reference/adr-001-storage-engine.md
```

---

## 8. Indexing Storage Directory (`gyrus sync`)

Re-index filesystem documents, calculate SHA-256 checksums, and extract dependency links:

```bash
# Human readable output
./gyrus sync

# Structured JSON report
./gyrus sync --json
```

---

## 9. Full-Text Search (`gyrus search`)

Query indexed contract documents using FTS5 lexical matching:

```bash
# Basic keyword search
./gyrus search --query "SQLite search"

# Filter by category and document type
./gyrus search --query "Cobra" --category "technical" --type "specification"

# JSON output
./gyrus search --query "FTS5" --json
```

---

## 10. Context Linearization (`gyrus suggest-context`)

Linearize relevant context matching an AI agent prompt:

```bash
# Suggest context for an architectural prompt
./gyrus suggest-context --prompt "How does Gyrus implement zero-dependency FTS search?"

# Return context as JSON array
./gyrus suggest-context --prompt "Cobra CLI exit codes" --json
```

---

## 11. Archiving Documents (`gyrus archive`)

Archive (delete) a document from storage and search index:

```bash
# Human-readable output
./gyrus archive spec-002-cli-framework

# Programmatic JSON response
./gyrus archive adr-001-storage-engine --json
```

---

## 12. Testing Embedded MCP Stdio Server (`gyrus mcp serve`)

Test launching the stdio Model Context Protocol (MCP) server:

```bash
# Test MCP stdio server startup (press Ctrl+C to stop)
./gyrus mcp serve
```

---

## 13. Testing Docker Containerized MCP Execution

Test zero-install containerized MCP server execution via Docker:

```bash
# Build local docker image
docker build -t gyrus:latest .

# Run containerized MCP server over stdio
docker run -i --rm -v "$(pwd):/workspace" gyrus:latest mcp serve --storage-path /workspace
```
