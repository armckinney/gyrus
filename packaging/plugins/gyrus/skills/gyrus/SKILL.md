---
name: gyrus
description: Gyrus Unified Context Control Plane & Memory Engine agent skill. Guides developers and agents through the 3-phase SDLC documentation lifecycle (Plan, Implement, Post-Implement) using native MCP tools or CLI commands.
applyTo:
  - "**"
---

# Gyrus Agent Skill: Context Control Plane & SDLC Engine

This skill equips AI coding assistants (Google Antigravity, GitHub Copilot, OpenAI Codex, Claude) to discover, synthesize, create, and maintain architectural context through the **Gyrus Context Control Plane**.

---

## 💡 The Gyrus SDLC Lifecycle

Use this skill whenever you need to resolve architectural context, align with requirements, or record engineering decisions across the 3 SDLC phases:

1. **Phase 1: Plan (Architecture & Design)**:
   - Check existing PRDs and blueprints before proposing structural changes.
   - Propose an Improvement Proposal (`improvement-proposal`) for major system designs.
   - Record immutable Architectural Decision Records (`adr`) for key choices and tradeoffs.
   - Maintain living system specifications (`specification`) and standards (`standards`).
2. **Phase 2: Implement (Build & Interfaces)**:
   - Implement source code to satisfy specifications. Source code is the truth for code; never duplicate code diffs in docs.
   - Update Technical Reference (`technical-reference`) when public APIs, CLI syntax, or configs change.
   - Update Guides (`guide`) when setup runbooks or onboarding flows change.
3. **Phase 3: Post-Implement (Release & Portal)**:
   - Draft Release Notes (`release-note`) for version milestones.
   - Maintain the Product Page (`product`) doc portal.

---

## 🛠️ Primary Interface: Model Context Protocol (MCP) Tools

- **`gyrus_suggest_context`**: `{ "prompt": string, "max_concepts"?: int, "scope"?: string, "project"?: string }`
- **`gyrus_search`**: `{ "query": string, "max_results"?: int, "scope"?: string, "project"?: string, "tenant"?: string }`
- **`gyrus_get`**: `{ "id": string }`
- **`gyrus_create`**: `{ "id": string, "title": string, "category": string, "type": string, "tenant": string, "status": string, "content": string, "scope"?: string, "project"?: string }`
- **`gyrus_update`**: `{ "id": string, "title"?: string, "status"?: string, "content"?: string, "expected_version"?: int }`
- **`gyrus_link`**: `{ "from_id": string, "to_id": string, "rel_type": string }`

---

## 💻 Fallback & Admin Interface: Gyrus CLI

- `gyrus suggest-context --prompt "<task>" --json`
- `gyrus search --query "<query>" --max-results 10 --json`
- `gyrus get <id> --json`
- `gyrus create --id "<id>" --title "<title>" --category "<cat>" --type "<type>" --tenant "<tenant>" --content "<md>"`
- `gyrus update <id> --status "<status>" --expected-version <v>`
- `gyrus link <from-id> <to-id> --rel-type <type>`

Keyword search defaults to ten results. CLI `--max-results` and MCP `max_results` pass through to the selected provider; positive limits also cap metadata browsing. Empty-query browsing is unlimited by default. SQLite index/graph/search providers and the sync command/tool have been removed; file-backed OKF data is read directly.

Supported search providers are `okf` (direct keyword scan) and `postgres_fts` (native PostgreSQL FTS). Vector search, the vector profile, and the `vector:` configuration section are removed and fail explicitly. Use supported profiles; no embedding service is required. PostgreSQL stale updates return concurrency errors, matching CLI exit code 4.

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- **[Installed-release lifecycle verification](../gyrus-test-lifecycle/SKILL.md):** Run agent-led CLI, MCP, and UI checks with isolated consumer fixtures, without repository source or development tooling.


## Qualified identity and relationships

Create takes a local frontmatter id and tenant/scope/project. All existing-concept `id`, `from_id`, `to_id`, and `start_id` arguments require a full name: `root/reference/adr-example` or `root/projects/gyrus/ticket-003`. No short-ID fallback is supported. API envelopes/references expose the derived name; Markdown frontmatter keeps the local id. Use `gyrus get root/reference/adr-example --json` and `gyrus link root/projects/gyrus/ticket-003 root/reference/adr-example --rel-type depends_on`.

Persist links only as `{concept: <qualified-name>, rel: <relationship>}`. Dependencies, old link keys, and per-link creation metadata are obsolete. Use incoming neighbors for backlinks; do not write inverse copies. Retain depends_on, implements, supersedes, and mitigates; do not impose parent/child or DAG rules. Use the packaged migration skill for legacy data. Typed links and Gyrus-required metadata are extensions to upstream OKF.
