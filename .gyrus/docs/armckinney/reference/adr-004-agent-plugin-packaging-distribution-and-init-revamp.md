---
id: adr-004-agent-plugin-packaging-distribution-and-init-revamp
title: Agent Plugin Packaging, Distribution & Init Revamp
category: architecture
type: adr
format: markdown
owner_group: armckinney
version: 1
status: accepted
last_modified_by: antigravity
last_updated: 2026-09-13T23:58:00Z
tags:
  - adr
  - architecture
  - plugins
  - agent-plugins-standard
  - mcp
  - cli
  - phase-2-6
dependencies:
  - adr-001-oop-core-refactoring-and-package-reorganization
  - adr-002-simplified-directory-topology
  - adr-003-persistence-layer-schema-storage-and-remote-linkage
---

# ADR 004: Agent Plugin Packaging, Distribution & Init Revamp

## Context

In earlier phases of Gyrus development, AI agent integrations were distributed as loose, standalone "agent skills" located under `packaging/skills/` and deployed to `.agents/skills/` or `.gemini/antigravity/skills/`. MCP server configurations were installed via an ad-hoc registry routine.

This approach presented several critical architectural drawbacks:
1. **Fragmented Agent Artifacts**: Skills, MCP configurations, context prompts, and behavioral instructions were managed as disconnected artifacts rather than a coherent, versioned bundle.
2. **Ecosystem Divergence**: Modern AI coding assistants and agent environments are adopting structured plugin standards. The [Agent Plugins Standard 1.0.0](https://agent-plugins.org/) defines an open, client-agnostic manifest specification (`plugin.json`, `mcp.json`), whereas Google Antigravity, GitHub Copilot, OpenAI Codex, and Claude Desktop each expect specific directory layouts, manifest conventions, or configuration keys.
3. **Monolithic & Ambiguous `gyrus init`**: The previous `gyrus init` command was an implicit, overloaded operation that conflated repository workspace initialization (`.gyrus.yaml`, directory creation, storage configuration) with client tool registration. Moreover, its `--target all` flag blindly mutated configuration directories across the filesystem, leading to unexpected side effects in multi-agent environments.
4. **Maintenance Overhead of Legacy Skills**: Loose standalone skill definitions (`gyrus-cli`, `gyrus-mcp`) duplicate documentation and configuration, diverging from the embedded agent plugin bundle.

## Decision

We decide to formalize Gyrus agent integration into a standardized, multi-client **Agent Plugin** bundle and revamp the initialization workflow into explicit, decoupled subcommands:

### 1. Dual-Compatible Agent Plugin Bundle
We adopt the Agent Plugins Standard 1.0.0 specification while maintaining full compatibility with Google Antigravity and major AI runtime clients.

Canonical plugin assets reside in `packaging/plugins/gyrus/` and are embedded directly into the Go binary (`internal/setup/plugin_embed.go`):
- **`plugin.json`**: Agent Plugins Standard 1.0.0 manifest referencing `$schema: https://agent-plugins.org/schemas/1.0.0/plugin.schema.json`, defining plugin identity, description, author ("Andrew McKinney"), capabilities, and interfaces.
- **`mcp.json`**: Agent Plugins Standard 1.0.0 MCP manifest referencing `$schema: https://agent-plugins.org/schemas/1.0.0/mcp.schema.json`, configuring the standard I/O (`stdio`) `gyrus mcp` server executable and arguments.
- **`mcp_config.json`**: Antigravity/Gemini CLI client MCP configuration format.
- **`rules/AGENTS.md`**: Context control plane rules instructing agents on memory retrieval, search, and context management protocols.
- **`skills/`**: Packaged unified skill conforming to the Open Skill Format:
  - `skills/gyrus/SKILL.md`: Comprehensive agent skill establishing native MCP tools as the primary interface, CLI subcommands as the resilient fallback and administration interface, and OKF contract governance rules.

### 2. Complete Deprecation of Standalone Legacy Skills
- All loose legacy skill packaging directories (`packaging/skills/`, `docs/agents/skills/gyrus-cli`, `docs/agents/skills/gyrus-mcp`) are removed.
- Workspace developer guidance meta-skills (such as `docs/agents/skills/agent-skills`) remain intact as repository development guides.
- All agent distribution now occurs exclusively through the unified Gyrus plugin.

### 3. Decoupled, Explicit CLI Initialization Workflow
The bare, monolithic `gyrus init` command is eliminated. Running `gyrus init` without a subcommand yields an informative error instructing the user to run explicit subcommands.

Two explicit subcommands are introduced:

#### a. `gyrus init config`
Initializes workspace or global repository configuration:
- Creates `.gyrus.yaml` with the chosen profile (`default`, `remote-git`, `remote-s3`, `remote-postgres`, etc.).
- Configures default owner groups and storage paths.
- Establishes the `.gyrus/docs/<owner_group>/` directory topology.
- Flags: `--profile` (`-p`), `--owner-group` (`-o`), `--storage-path`.

#### b. `gyrus init client`
Installs the Gyrus agent plugin and configures MCP server integration for a specific AI coding assistant runtime:
- Requires an explicit `--target` (`-t`): `antigravity`, `claude`, `codex`, or `copilot`. The ambiguous `all` target is permanently removed.
- Supported modes: `--mode` (`stdio` or `docker`).
- Scope control: `--global` (`-g`) to install into user home runtime paths, or workspace local installation by default.
- Customizable destination: `--plugin-dir` for custom target installations.

## Consequences

### Positive
- **Standardized Packaging**: Complies with the emerging Agent Plugins Standard 1.0.0 specification while maintaining full operational compatibility with Google Antigravity, GitHub Copilot, OpenAI Codex, and Claude.
- **Single Source of Truth**: Agent skills, MCP configurations, and system rules are bundled and versioned together as one canonical plugin, embedded at compile-time.
- **Predictable & Safe Initialization**: Splitting repository configuration from client tool installation eliminates unintentional filesystem mutations, guarantees explicit client selection, and streamlines CI/CD environments.
- **Zero Legacy Drag**: Eliminates outdated standalone skills and reduces cognitive overhead for developers and AI agents.

### Negative / Breaking Changes
- Scripts or workflows invoking bare `gyrus init` or `gyrus init --target all` will fail with an explicit validation error and must be updated to invoke `gyrus init config` and `gyrus init client --target <target>`.
- Any external dependencies expecting loose skills in `packaging/skills/` must consume the plugin bundle under `packaging/plugins/gyrus/` instead.

