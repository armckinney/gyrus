---
title: GYRUS-401 MCP Server & Tools
id: ticket-gyrus-401
type: freeform
format: ""
category: product
version: 2
status: archived
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

> Historical record. This document describes earlier work and is not a current setup or capability contract. Use the current configuration reference and architecture index.

# [GYRUS-401] MCP Stdio Server & Memory Tools

> **Status:** `COMPLETED`
> **Phase:** Phase 4 - MCP Server & Open Skill Format Adapters
> **Owner:** Antigravity Agent
> **Dependencies:** [GYRUS-303]



---

## 1. Description
Implement the Model Context Protocol (MCP) server running over `stdio` in `internal/mcp`. Expose standard memory tools: `memory.suggest_context`, `memory.search`, `memory.get`, `memory.create`, `memory.update`, `memory.link`, `memory.unlink`, `memory.validate`.

---

## 2. Acceptance Criteria
- [ ] Implement `gyrus mcp serve` stdio transport handler.
- [ ] Expose all specified `memory.*` tools matching Specification 03 parameters.
- [ ] Connect MCP tool invocations to the underlying Gyrus Core SDK.

---

## 3. Implementation Tasks
1. Create MCP server wrapper in `internal/mcp/server.go`.
2. Implement tool definitions in `internal/mcp/tools.go`.
3. Add JSON-RPC 2.0 integration tests verifying tool invocation over stdio.

---

## 4. Verification & Testing Commands
```bash
go test ./internal/mcp/... -v
```
