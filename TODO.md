# TODO

## Remediate Local Configuration
- [x] Codex/claude fix install plugin
  - Just move install to client native + instructions (validate plugin install methods)
  - Remove `gyrus client` commands
- [] (MVP) Add OKF backend for index/search/
  - Should link/unlink persist in doc metadata always or just in OKF graphg
- [] Update config defaults to okf backend

## Remediate MVP Multi-Repo Configuration
- [] Fix Git Storage Provider
- [] Fix `gyrus sync` to use set Storage Provider and not just local sync
- [] Remove sqlite from index/graph/search providers
- [] Update docs
  - Update guides / root readme
  - Simplify value add doc
  - Update Config / Setup Docs (include standard consigs)
  - Rework ADRs to better scoping

### Cleanup
- [] What are the `gyrus hook` commands for?
- [] Remove `gyrus completion` functionality does not fit the scripts/cicd cli usage pattern
- [] Rename "documents" / "docs" to concepts for OKF standard
- [] Rename "workspaces" to "projects" to match similar agentic terminology
- [] Update `owner_group` to `tenant` and config rename
- [] Codex Refactor  
  - [x] rename/remove "setup" package as appropriate

### MCP Server Cache
- [] Implement Cache Providers: `none` (default), `sqlite`, `redis`

## Enable Team Configuration

### Split out Componenets
- [] Split out all components from repo:
  - sdk
  - cli
  - mcp
  - ui

### Python MCP Server
- [] Implement thin python wrapper that can run on Databricks Apps using FastMCP
  - OR gRPC Core SDK Endpoint: High-performance gRPC service definitions for multi-language Core SDK bindings (Python)

### Web App
- [] Full Web App Conversational AI Context Retrieval Chatbot Agent
  - Build on Core SDK
- [] display gyrus schemas in ui with navbar icon

## Enable Enterprise Configuration
TODO
