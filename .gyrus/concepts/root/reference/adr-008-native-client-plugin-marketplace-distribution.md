---
id: adr-008-native-client-plugin-marketplace-distribution
title: Native Client Plugin Marketplace Distribution Architecture
category: architecture
type: adr
format: markdown
tenant: root
version: 1
status: accepted
last_modified_by: antigravity
last_updated: 2026-10-16T01:24:00Z
tags:
    - adr
    - architecture
    - plugins
    - marketplace
    - codex
    - antigravity
    - copilot
    - claude-code
dependencies:
    - adr-004-agent-plugin-packaging-distribution-and-init-revamp
    - adr-006-native-mcp-workspace-discovery-and-lifecycle-cli-refactor
---

# ADR 008: Native Client Plugin Marketplace Distribution Architecture

## Context

Prior to this architectural shift, Gyrus distributed its agent capabilities through the CLI command `gyrus client install/uninstall` and internal file-provisioning routines (`internal/setup`). This implementation generated local client configuration files (such as `.vscode/mcp.json`, `.claude/mcp.json`, `.codex/mcp.json`, and `.agents/plugins/gyrus/`) directly inside developer workspaces.

This legacy approach presented several critical challenges:
1. **Developer Experience Friction**: Users were forced to install the Gyrus binary first, manually run CLI provisioning commands per client harness, and deal with file clutter written into their repositories.
2. **Ecosystem Inversion**: Modern agent harnesses (OpenAI Codex, Google Antigravity, GitHub Copilot, Anthropic Claude Code) have established first-class native plugin and remote marketplace distribution protocols (`codex plugin add ...`, `agy plugin install ...`, `claude plugin install ...`). A bespoke CLI provisioning tool bypassed these standard ecosystem registries and prevented automatic updates, permissions management, and harness sandboxing.
3. **Repository Directory Duplication**: Maintaining multiple copies of the plugin manifests and assets (e.g. duplicating `.agents/plugins/gyrus` alongside `packaging/plugins/gyrus`) created synchronization debt and violated the single-source-of-truth principle.

## Decision

We migrate Gyrus from CLI-based client provisioning to native harness-level plugin and marketplace distribution:

1. **Remove `gyrus client` CLI Commands and `internal/setup` Package**:
   - The CLI subcommands `gyrus client install` and `gyrus client uninstall` are eliminated.
   - The `internal/setup` package is removed entirely. Profile generation for `.gyrus.yaml` is retained cleanly within `internal/config/profiles.go` and surfaced exclusively via `gyrus config init`.
2. **Establish GitHub as the Remote Marketplace Registry**:
   - Standardize marketplace identification under `"name": "armckinney"` so users can address the plugin canonically as `gyrus@armckinney`.
   - Provide standard marketplace manifests at repository roots:
     - OpenAI Codex: `.agents/plugins/marketplace.json` (`source: {"source": "local", "path": "./packaging/plugins/gyrus"}`)
     - GitHub Copilot: `.github/plugin/marketplace.json` (`source: "./packaging/plugins/gyrus"`)
     - Anthropic Claude Code: `.claude-plugin/marketplace.json` (`source: "./packaging/plugins/gyrus"`)
3. **Canonical Plugin Bundle in `packaging/plugins/gyrus`**:
   - Retain `packaging/plugins/gyrus` as the single canonical plugin bundle containing:
     - Standard Agent Plugins 1.0 manifest: `plugin.json`
     - Claude Code manifests: `.claude-plugin/plugin.json`, `.mcp.json`
     - Codex manifest: `.codex-plugin/plugin.json`
     - MCP server configuration: `mcp.json` and `mcp_config.json` with `"env": {"GYRUS_WORKSPACE": "${workspaceFolder}"}`
     - System rules (`rules/AGENTS.md`) and skills (`skills/gyrus/SKILL.md`, `skills/gyrus-index/SKILL.md`).
4. **Antigravity Repository Discovery**:
   - Define `.agents/plugins.json` referencing `{"entries": [{"path": "packaging/plugins"}]}` to enable local Antigravity plugin resolution without duplicating asset directories.
   - Support remote installation via `agy plugin install https://github.com/armckinney/gyrus/tree/main/packaging/plugins/gyrus`.

## Consequences

### Positive
- **Streamlined User Onboarding**: Users install Gyrus using their harness's native idioms (`codex plugin marketplace add armckinney/gyrus && codex plugin add gyrus@armckinney`, `claude plugin marketplace add armckinney/gyrus && claude plugin install gyrus@armckinney`, etc.).
- **Zero Workspace Litter**: No unexpected `.vscode/`, `.claude/`, or duplicate `.agents/` configuration trees are injected into user projects.
- **Single Source of Truth**: All plugin metadata, MCP definitions, rules, and skills live strictly in `packaging/plugins/gyrus`.
- **Architectural Purity**: Gyrus CLI strictly manages OKF knowledge graph operations, search, and configuration; it no longer acts as a client harness installer.

### Negative
- **Manual Binary Prerequisite**: Users must ensure the `gyrus` CLI binary is installed and present on `PATH` for the stdio MCP server spawned by the harness plugins.

