# TODO

## Remediate Local Configuration
- [x] Codex/claude fix install plugin
  - Just move install to client native + instructions (validate plugin install methods)
  - Remove `gyrus client` commands
- [x] (MVP) Add OKF backend for index/search/
  - Should link/unlink persist in doc metadata always or just in OKF graphg
- [x] Update config defaults to okf backend

## Remediate MVP Multi-Repo Configuration
- [ ] Fix Git Storage Provider
- [x] Remove sqlite from index/graph/search providers and remove `gyrus sync`
  - Keep shared SQLite persistence for a future cache provider.
- [ ] Update profile options to logical implementations (local, git, postgres)
- [ ] Update docs
  - Update guides / root readme
  - Simplify value add doc
  - Update Config / Setup Docs (include standard consigs)
  - Rework ADRs to better scoping

### Cleanup
- [x] What are the `gyrus hook` commands for? (Removed hook commands and automation hooks)
- [x] Remove `gyrus completion` functionality does not fit the scripts/cicd cli usage pattern
- [x] Rename "documents" / "docs" to concepts for OKF standard
- [x] Rename "workspaces" to "projects" to match similar agentic terminology
- [x] Update `owner_group` to `tenant` and config rename
- [x] Remove Vector Search Provider and config; defer semantic search to a future backend
- [ ] Codex Refactor  
  - [x] rename/remove "setup" package as appropriate



### MCP Server Cache
- [ ] Implement Cache Providers: `none` (default), `sqlite`, `redis`

## Enable Team Configuration

### Split out Componenets
- [ ] Split out all components from repo:
  - sdk
  - cli
  - mcp
  - ui

### Python MCP Server
- [ ] Implement thin python wrapper that can run on Databricks Apps using FastMCP
  - OR gRPC Core SDK Endpoint: High-performance gRPC service definitions for multi-language Core SDK bindings (Python)

### Web App
- [ ] Full Web App Conversational AI Context Retrieval Chatbot Agent
  - Build on Core SDK
- [ ] display gyrus schemas in ui with navbar icon

## Enable Enterprise Configuration
TODO
