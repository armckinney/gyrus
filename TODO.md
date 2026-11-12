# TODO

## Remediate MVP Multi-Repo Configuration
- [x] Fix Git Storage Provider
- [x] Update profile options to logical implementations (local, git, postgres)
- [x] Update docs
  - Update guides / root readme
  - Simplify value add doc
  - Update Config / Setup Docs (include standard consigs)
  - Rework ADRs to better scoping

### Cleanup
- [ ] Codex Refactor  
  - [x] rename/remove "setup" package as appropriate
- [ ] Ensure Docs are up to date


## Stability
- [ ] Implement Cache Providers: `none` (default), `sqlite`, `redis`
- [ ] Implement write locking / 2 phase commit (Postgres initially, redis later)

## Platform MVP
- [ ] Hosted MCP - Implement python MCP running on Databricks Apps
    - [ ] Python gRPC
    - [ ] Databricks App for MCP
    - [ ] Plugin MCP Update
- [ ] Hosted UI  WebApp - Doc viewer running on Databricks App
   - Build on Core SDK
   - Display gyrus schemas in ui with navbar icon


## Value Add - Intelligence Platform
- [ ] Enable context intelligence - Lakebase Sync / Context Processing
- [ ] Enable Agentic UI Integration
  -  UI App Lakebase
  - Mosaic Integration

## Scalability
- [ ] Data Layer Decoupling - Execution Queue / Ingestion Processing
