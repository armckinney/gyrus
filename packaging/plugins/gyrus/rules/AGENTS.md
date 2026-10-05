# Gyrus Context Control Plane & SDLC Directives

This document instructs AI agents (Antigravity, Claude Code, GitHub Copilot, OpenAI Codex) on interacting with the Gyrus Context Control Plane and working within the OKF documentation framework.

---

## 💡 Core Architectural Principles

1. **Source Code vs. Gyrus Context Plane**:
   - **Source Code Owns Implementation**: Function logic, routine bug fixes, internal refactoring, and code-level changes belong exclusively in the codebase. Never duplicate routine code edits, method implementations, or commit diffs in Gyrus.
   - **Gyrus Owns Architectural Intent & System Governance**: Gyrus captures the high-level context required to piece together current architecture and guide future changes across the engineering lifecycle.

2. **The Gyrus 3-Phase SDLC Lifecycle (from standards-001)**:
   - **Phase 1: Plan (Architecture & Design)**:
     - `prd`: Business requirements, motivation, success metrics (*what & why*).
     - `improvement-proposal`: Design alternatives, system architecture, testing strategy (*how*).
     - `specification`: Final system blueprints, data models, protocols, component boundaries.
     - `adr`: Short audit records of micro-decisions and tradeoffs (*why this way*).
     - `standards`: Engineering quality rules, linting policies, and git conventions.
   - **Phase 2: Implement (Build & Interfaces)**:
     - *Source Code*: Canonical source of truth for implementation logic.
     - `technical-reference`: Public API endpoints, CLI syntax, config parameters, and technical FAQs.
     - `guide`: Developer onboarding tutorials, runbooks, and troubleshooting walkthroughs.
   - **Phase 3: Post-Implement (Release & Portal)**:
     - `release-note`: Version release tracking, changelogs, migration guides.
     - `product`: Main product homepage portal summarizing value proposition and doc navigation.

3. **Context Resolution Before Architecture Changes**:
   - Before designing new features or making architectural decisions, resolve relevant context:
     - Primary: Native MCP `gyrus_suggest_context({ prompt: "<task>" })` or `gyrus_search`.
     - Fallback (CLI): `gyrus suggest-context --prompt "<task>" --json`.

4. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (`prd`, `specification`, `standards`, `technical-reference`, `guide`, `product`): Capture active system state. Update living specs and standards when architecture or public interfaces evolve.
   - 📜 **Immutable Decision Logs** (`adr`, `improvement-proposal`, `release-note`): Historical snapshots. Once accepted (`status: accepted`), they are strictly immutable. When architectural decisions change:
     1. Propose a NEW ADR (`gyrus create`).
     2. Link the new ADR to supersede the old one (`gyrus link <new-id> <old-id> --rel-type supersedes`).
     3. Update the old ADR status to `superseded` (`gyrus update <old-id> --status superseded`).

5. **OKF Contract Schema Compliance**:
   - Document IDs MUST match lower-case regex `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
   - Every contract document requires YAML frontmatter defining `id`, `title`, `category`, `type`, `owner_group`, `version`, `status`.

6. **Programmatic Exit Codes Protocol**:
   - `0`: Success
   - `1`: Schema / ID validation error
   - `2`: Illegal lifecycle state transition
   - `3`: Permission / authentication error
   - `4`: Concurrency lock mismatch
   - `5`: Record / Storage backend error
