---
name: gyrus-index
description: Gyrus Repository Primer skill. Interactively explores codebase architecture, dependencies, and standards, proposes synthesized baseline living documents (spec, standards, technical reference, onboarding guide), confirms with developer, and indexes them into Gyrus.
applyTo:
  - "**"
---

# Gyrus Repository Primer Skill (`/gyrus-index`)

This skill guides the AI agent to survey a new or existing codebase and prime Gyrus with high-level architectural living documents without duplicating low-level code implementation details.

---

## 📋 The Primer Workflow

When the developer invokes `/gyrus-index`, follow this 5-step interactive procedure:

### Step 1: Survey Codebase Architecture
- Examine the root directory, package manifests (`go.mod`, `package.json`, `Cargo.toml`, etc.), README, build configs, and entry points.
- Identify the project name, primary language, architectural style (e.g. modular monolith, microservices, CLI tool), and external dependencies.

### Step 2: Propose Baseline Living Documents
Formulate a concise, synthesized proposal of 3-4 core living documents for the repository:
1. **`spec-001-architecture.md`** (`type: specification`): System topology, high-level component boundaries, and data flows.
2. **`standards-001-guidelines.md`** (`type: standards`): Engineering quality standards, code organization, testing strategy, and git conventions.
3. **`tech-ref-001-api-cli.md`** (`type: technical-reference`): Public CLI commands, API interfaces, configuration parameters, or technical lookups.
4. **`guide-001-onboarding.md`** (`type: guide`): Developer environment setup, build instructions, and common runbooks.

### Step 3: Developer Review & Confirmation
- Present the synthesized proposal to the developer with a clear overview of what each document will capture.
- Ask for confirmation or adjustments before creating any files.

### Step 4: Scaffold Confirmed Documents
- Once approved, create the documents in `.gyrus/docs/<owner_group>/reference/` using valid OKF YAML frontmatter:
  - `id`, `title`, `category`, `type`, `owner_group`, `version: 1`, `status: active`.
- Ensure documents capture **high-level intent and boundaries**, not line-by-line function implementations.

### Step 5: Index & Synchronize
- Run `gyrus sync` to parse documents, build the SQLite FTS search index, and construct dependency edges.
- Inform the developer that Gyrus is primed and ready to provide architectural context to all AI agents.
