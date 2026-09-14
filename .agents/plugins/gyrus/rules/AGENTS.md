# Gyrus Context Control Plane & Memory Engine Rules

This document instructs AI agents (Antigravity, Claude Code, GitHub Copilot, OpenAI Codex) on interacting with the Gyrus Context Control Plane.

---

## 💡 Core Agent Guidelines

1. **Context Resolution Before Code Modification**:
   - Before editing or implementing code, ALWAYS invoke `gyrus suggest-context --prompt "<task description>"` or MCP tool `gyrus_suggest_context({ prompt: "<task description>" })`.
   - Linearized context prioritizes applicable Architecture Design Records (ADRs), Product Requirements Documents (PRDs), and technical specifications within prompt token budgets.

2. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (`prd`, `specification`, `guide`, `standards`, `glossary`): Represent current active system specifications. Maintain and update these documents as implementations evolve.
   - 📜 **Immutable Decision Logs** (`adr`, `improvement-proposal`, `release-note`): Historical records. Once `accepted`, `approved`, or `active`, they are strictly immutable (`immutable: true`). When architectural decisions change:
     1. Create a NEW ADR (`gyrus create`).
     2. Link the new ADR to supersede the old one (`gyrus link <new-id> <old-id> --rel-type supersedes`).
     3. Update the superseded ADR status to `superseded` (`gyrus update <old-id> --status superseded`).

3. **OKF Contract Schema Compliance**:
   - All document IDs MUST match lower-case regex `^[a-z0-9-_]+$` (e.g., `adr-001-storage-engine`).
   - Every contract document requires YAML frontmatter defining `id`, `title`, `category`, `type`, `owner_group`, `version`, `status`.

4. **Programmatic Exit Codes Protocol**:
   - `0`: Success
   - `1`: Schema / ID validation error
   - `2`: Illegal lifecycle state transition
   - `3`: Permission / authentication error
   - `4`: Concurrency lock mismatch
   - `5`: Record / Storage backend error
