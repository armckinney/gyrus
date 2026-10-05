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

- **`gyrus_suggest_context`**: `{ "prompt": string, "limit"?: int }`
- **`gyrus_search`**: `{ "query": string, "limit"?: int }`
- **`gyrus_get_document`**: `{ "id": string }`
- **`gyrus_create_document`**: `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }`
- **`gyrus_update_document`**: `{ "id": string, "title"?: string, "status"?: string, "content"?: string }`
- **`gyrus_link_documents`**: `{ "from_id": string, "to_id": string, "rel_type": string }`
- **`gyrus_sync`**: `{}`

---

## 💻 Fallback & Admin Interface: Gyrus CLI

- `gyrus suggest-context --prompt "<task>" --json`
- `gyrus search --query "<query>" --json`
- `gyrus get <id> --json`
- `gyrus create --id "<id>" --title "<title>" --category "<cat>" --type "<type>" --owner-group "<grp>" --content "<md>"`
- `gyrus update <id> --status "<status>" --expected-version <v>`
- `gyrus link <from-id> <to-id> --rel-type <type>`
- `gyrus sync`
