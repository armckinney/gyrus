---
id: adr-006-native-mcp-workspace-discovery-and-lifecycle-cli-refactor
title: Native MCP Architecture, Multi-Tier Workspace Discovery, and Client Lifecycle CLI Refactor
category: architecture
type: adr
format: markdown
owner_group: root
version: 2
status: accepted
last_modified_by: ""
last_updated: 2026-09-30T22:42:04Z
---

# ADR 006: Native MCP Architecture, Multi-Tier Workspace Discovery, and Client Lifecycle CLI Refactor

## Context

Prior to Phase 2.8, Gyrus distributed its Model Context Protocol (MCP) server integration with support for both containerized stdio (`docker run ... ghcr.io/armckinney/gyrus:latest mcp serve`) and local binary execution (`gyrus mcp serve`), with `--mode container` as the default in `gyrus init client`.

In practice, this architecture suffered from critical flaws:
1. **The Circular Installation Paradox**: To configure the containerized MCP server, developers were already required to install the `gyrus` binary locally to run `gyrus init client`. Having the CLI instruct AI agents to ignore the installed native binary and spawn Docker was redundant and counterproductive.
2. **Container-in-Container / DevContainer Breakage**: Modern software teams frequently develop inside DevContainers and GitHub Codespaces. Running containerized stdio MCP inside a DevContainer requires Docker-in-Docker (DinD), which causes host-versus-container volume mount collisions (`-v ${workspaceFolder}:/workspace`), socket permission failures, and empty directory mounts.
3. **File Ownership & Virtualized Paths**: Docker containers running as root created files owned by `root:root` on Linux host filesystems. Furthermore, paths virtualized inside the container (`/workspace/...`) broke clickable markdown file links on developer host IDEs.
4. **Agent Plugin Working Directory Disconnect**: When modern agent plugin runners (Antigravity, VS Code, Claude) launch MCP servers from plugin cache directories (`~/.gemini/config/plugins/...`), Gyrus's configuration resolution (relying purely on `os.Getwd()`) walked up the plugin directory and failed to find the project's `.gyrus.yaml`, silently falling back to empty user-home defaults.
5. **CLI Namespace Ambiguity**: Grouping setup tasks under `gyrus init` created ambiguity (`gyrus init config` vs `gyrus init client`). Configuration initialization naturally belongs under `gyrus config`, while agent client management belongs under a dedicated `gyrus client` lifecycle group (`install`, `uninstall`).

## Decision

We decide to:

### 1. Standardize Exclusively on Native Binary MCP Execution
We retire the containerized local stdio MCP execution mode. For all local and DevContainer development, AI agent tools execute the native, zero-dependency static Go binary (`gyrus mcp serve`). Docker container images remain dedicated to CI/CD automation pipelines and centralized cloud services (`gyrus serve`).

### 2. Implement Multi-Tier Workspace & Configuration Discovery
Configuration resolution in `internal/config/loader.go` now prioritizes an explicit hierarchy:
1. **Explicit Config Override**: `GYRUS_CONFIG` environment variable or `--config` CLI flag.
2. **Explicit Workspace Override**: `GYRUS_WORKSPACE` environment variable or `--workspace` CLI flag (searches upward for `.gyrus.yaml`).
3. **Current Working Directory**: `os.Getwd()` upward directory traversal.
4. **DevContainer Auto-Detection**: If running inside containerized environments (`/.dockerenv`), inspect active container workspaces (`/workspaces/*`).
5. **Global Fallback**: `~/.gyrus.yaml` in user home directory.
6. **Default Application Root**: `~/.gyrus/`.

### 3. Inject Workspace Context into Generated Agent Configurations
When configuring agent clients, `mcp.json` and `mcp_config.json` automatically include:
```json
{
  "mcpServers": {
    "gyrus": {
      "type": "stdio",
      "command": "gyrus",
      "args": ["mcp", "serve"],
      "env": {
        "GYRUS_WORKSPACE": "${workspaceFolder}"
      }
    }
  }
}
```

### 4. Refactor Setup Commands into Domain Lifecycles
We restructure CLI commands into cohesive domain namespaces:
* **`gyrus config init`**: Replaces `gyrus init config` as the canonical command for generating `.gyrus.yaml`.
* **`gyrus client install`**: Replaces `gyrus init client` as the canonical command for equipping agent plugins and registering MCP servers. Automatically executes an initial index sync so context is primed immediately.
* **`gyrus client uninstall`**: Cleanly removes the agent plugin bundle, unregisters MCP server definitions from target AI tools, and deregisters with the Antigravity plugin registry (`agy plugin uninstall gyrus`).
* **Backward Compatibility**: `gyrus init config` and `gyrus init client` are retained as transparent aliases forwarding to the new commands with deprecation notices.

## Consequences

### Positive
* **Zero Container Friction**: Eliminates Docker daemon requirements, volume mount mismatches, root file ownership issues, and DinD breakage in DevContainers.
* **Instant MCP Startup**: Server spin-up drops from 1-2 seconds (container pull/start) to <15 milliseconds (native execution).
* **Guaranteed Context Resolution**: Agents immediately resolve workspace documents regardless of process launch directories.
* **Symmetric Lifecycle Management**: Teams can both cleanly install and completely uninstall Gyrus integrations across Antigravity, Claude, Codex, and Copilot.
* **Cleaner CLI Ergonomics**: Logical command organization aligns with developer expectations (`config show/init`, `client install/uninstall`).

### Negative / Migration
* Users running in environments without `gyrus` in `$PATH` must install the binary via `curl -sSL ... | bash` or pass an explicit binary path.
