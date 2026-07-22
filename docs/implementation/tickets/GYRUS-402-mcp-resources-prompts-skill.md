# [GYRUS-402] MCP Resources, Prompts & Open Skill Format Shim

> **Status:** `NOT STARTED`
> **Phase:** Phase 4 - MCP Server & Open Skill Format Adapters
> **Owner:** Unassigned
> **Dependencies:** [GYRUS-401]

---

## 1. Description
Implement MCP resources (`memory://doc/{id}`, `memory://schema/{type}`, etc.) and MCP prompts (`prepare-adr`, `review-implementation-against-memory`, `propose-memory-update`), and create the Open Skill Format definition wrapping the `gyrus` CLI binary.

---

## 2. Acceptance Criteria
- [ ] Implement MCP resource handlers in `internal/mcp/resources.go`.
- [ ] Implement MCP prompt templates in `internal/mcp/prompts.go`.
- [ ] Create `skills/gyrus/SKILL.md` exposing the Open Skill Format wrapper for CLI agents.

---

## 3. Implementation Tasks
1. Register resource URIs and prompt templates in the MCP server.
2. Draft `skills/gyrus/SKILL.md` with instructions for agent CLI invocation.
3. Write test cases validating resource fetching and prompt rendering.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/mcp/... -v
```
