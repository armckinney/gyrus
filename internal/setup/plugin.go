package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClientTarget identifies an AI agent platform for plugin and MCP installation.
type ClientTarget string

const (
	ClientTargetAntigravity ClientTarget = "antigravity"
	ClientTargetClaude      ClientTarget = "claude"
	ClientTargetCodex       ClientTarget = "codex"
	ClientTargetCopilot     ClientTarget = "copilot"
)

// ValidateClientTarget validates whether the target string is a supported client target.
func ValidateClientTarget(target string) (ClientTarget, error) {
	switch ClientTarget(target) {
	case ClientTargetAntigravity, ClientTargetClaude, ClientTargetCodex, ClientTargetCopilot:
		return ClientTarget(target), nil
	default:
		return "", fmt.Errorf("invalid client target '%s'. Valid targets: antigravity, claude, codex, copilot", target)
	}
}

// PluginInstallOptions defines parameters for equipping the Gyrus Agent Plugin.
type PluginInstallOptions struct {
	TargetDir      string
	WorkspaceDir   string
	Mode           MCPMode
	ContainerImage string
	BinaryCmd      string
}

// InstallAgentPlugin installs the complete Gyrus Agent Plugin package conforming to both
// the Agent Plugins Standard 1.0.0 and Google Antigravity plugin loader.
func InstallAgentPlugin(opts PluginInstallOptions) ([]string, error) {
	if opts.TargetDir == "" {
		opts.TargetDir = filepath.Join(".agents", "plugins", "gyrus")
	}
	if opts.BinaryCmd == "" {
		opts.BinaryCmd = "gyrus"
	}
	if opts.ContainerImage == "" {
		opts.ContainerImage = "ghcr.io/armckinney/gyrus:latest"
	}
	if opts.Mode == "" {
		opts.Mode = MCPModeContainer
	}
	if opts.WorkspaceDir == "" {
		opts.WorkspaceDir = "."
	}

	var createdFiles []string

	// 1. Create directory structure
	rulesDir := filepath.Join(opts.TargetDir, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating plugin rules dir: %w", err)
	}

	// 2. Write plugin.json manifest (Agent Plugins 1.0.0 & Antigravity compatible)
	pluginManifest := map[string]interface{}{
		"$schema":     "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json",
		"name":        "gyrus",
		"version":     "1.0.0",
		"description": "Gyrus: Unified Context Control Plane & Memory Engine for AI Agents",
		"author": map[string]string{
			"name": "Andrew McKinney",
		},
		"repository": "https://github.com/armckinney/gyrus",
		"license":    "Apache-2.0",
		"keywords": []string{
			"gyrus",
			"context-engine",
			"okf",
			"mcp",
			"agent-skills",
			"memory",
		},
	}
	manifestBytes, err := json.MarshalIndent(pluginManifest, "", "  ")
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(opts.TargetDir, "plugin.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, manifestPath)

	// Resolve MCP command and arguments
	var mcpCommand string
	var mcpArgs []string
	if opts.Mode == MCPModeContainer {
		mcpCommand = "docker"
		mcpArgs = []string{"run", "-i", "--rm", "-v", opts.WorkspaceDir + ":/workspace", "-w", "/workspace", opts.ContainerImage, "mcp", "serve"}
	} else {
		mcpCommand = opts.BinaryCmd
		mcpArgs = []string{"mcp", "serve"}
	}

	// 3. Write mcp.json (Agent Plugins 1.0.0 MCP schema compliant)
	standardMCP := map[string]interface{}{
		"$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
		"mcpServers": map[string]interface{}{
			"gyrus": map[string]interface{}{
				"type":    "stdio",
				"command": mcpCommand,
				"args":    mcpArgs,
			},
		},
	}
	mcpBytes, err := json.MarshalIndent(standardMCP, "", "  ")
	if err != nil {
		return nil, err
	}
	mcpPath := filepath.Join(opts.TargetDir, "mcp.json")
	if err := os.WriteFile(mcpPath, mcpBytes, 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, mcpPath)

	// 4. Write mcp_config.json (Antigravity format)
	antigravityMCP := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"gyrus": map[string]interface{}{
				"command": mcpCommand,
				"args":    mcpArgs,
			},
		},
	}
	antigravityMCPBytes, err := json.MarshalIndent(antigravityMCP, "", "  ")
	if err != nil {
		return nil, err
	}
	antigravityMCPPath := filepath.Join(opts.TargetDir, "mcp_config.json")
	if err := os.WriteFile(antigravityMCPPath, antigravityMCPBytes, 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, antigravityMCPPath)

	// 5. Write rules/AGENTS.md
	rulesPath := filepath.Join(rulesDir, "AGENTS.md")
	if err := os.WriteFile(rulesPath, []byte(embeddedPluginRules), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, rulesPath)

	// 6. Write skills/gyrus
	skillDir := filepath.Join(opts.TargetDir, "skills", "gyrus")
	refDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating skill dir %s: %w", refDir, err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(embeddedSkillMD), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, skillFile)

	schemasRef := filepath.Join(refDir, "okf-schemas.md")
	if err := os.WriteFile(schemasRef, []byte(embeddedSchemasRef), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, schemasRef)

	mcpRef := filepath.Join(refDir, "mcp-setup.md")
	if err := os.WriteFile(mcpRef, []byte(embeddedMcpRef), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, mcpRef)

	storageRef := filepath.Join(refDir, "storage-providers.md")
	if err := os.WriteFile(storageRef, []byte(embeddedStorageRef), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, storageRef)

	return createdFiles, nil
}

const embeddedPluginRules = `# Gyrus Context Control Plane & Memory Engine Rules

This document instructs AI agents (Antigravity, Claude Code, GitHub Copilot, OpenAI Codex) on interacting with the Gyrus Context Control Plane.

---

## 💡 Core Agent Guidelines

1. **Context Resolution Before Code Modification**:
   - Before editing or implementing code, ALWAYS invoke ` + "`gyrus suggest-context --prompt \"<task description>\"`" + ` or MCP tool ` + "`gyrus_suggest_context({ prompt: \"<task description>\" })`" + `.
   - Linearized context prioritizes applicable Architecture Design Records (ADRs), Product Requirements Documents (PRDs), and technical specifications within prompt token budgets.

2. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (` + "`prd`, `specification`, `guide`, `standards`, `glossary`" + `): Represent current active system specifications. Maintain and update these documents as implementations evolve.
   - 📜 **Immutable Decision Logs** (` + "`adr`, `improvement-proposal`, `release-note`" + `): Historical records. Once ` + "`accepted`, `approved`, or `active`" + `, they are strictly immutable (` + "`immutable: true`" + `). When architectural decisions change:
     1. Create a NEW ADR (` + "`gyrus create`" + `).
     2. Link the new ADR to supersede the old one (` + "`gyrus link <new-id> <old-id> --rel-type supersedes`" + `).
     3. Update the superseded ADR status to ` + "`superseded` (`gyrus update <old-id> --status superseded`)" + `.

3. **OKF Contract Schema Compliance**:
   - All document IDs MUST match lower-case regex ` + "`^[a-z0-9-_]+$`" + ` (e.g., ` + "`adr-001-storage-engine`" + `).
   - Every contract document requires YAML frontmatter defining ` + "`id`, `title`, `category`, `type`, `owner_group`, `version`, `status`" + `.

4. **Programmatic Exit Codes Protocol**:
   - ` + "`0`: Success" + `
   - ` + "`1`: Schema / ID validation error" + `
   - ` + "`2`: Illegal lifecycle state transition" + `
   - ` + "`3`: Permission / authentication error" + `
   - ` + "`4`: Concurrency lock mismatch" + `
   - ` + "`5`: Record / Storage backend error" + `
`

const (
	embeddedSkillMD = `---
name: gyrus
description: Gyrus Unified Context Control Plane & Memory Engine agent skill. Discovers, retrieves, creates, and maintains OKF codebase context (ADRs, PRDs, specs) using native MCP tools (primary) or CLI subcommands (fallback).
applyTo:
  - "**"
---

# Gyrus Agent Skill: Context Control Plane & Memory Engine

This skill equips AI coding assistants (Google Antigravity, GitHub Copilot, OpenAI Codex, Claude) to discover, synthesize, create, and maintain codebase context through the **Gyrus Context Control Plane**.

---

## 💡 Core Agent Guidelines

1. **Context First Before Editing Code**:
   - Always resolve relevant architectural and design context before modifying codebase implementation.
   - **Primary Interface (Native MCP):** Call ` + "`" + `gyrus_suggest_context({ prompt: "<task description>" })` + "`" + ` or ` + "`" + `gyrus_search({ query: "<keywords>" })` + "`" + `.
   - **Fallback Interface (CLI):** If MCP tools are unavailable or in shell-only environments, run ` + "`" + `gyrus suggest-context --prompt "<task description>"` + "`" + ` or ` + "`" + `gyrus search --query "<keywords>" --json` + "`" + `.
2. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (` + "`" + `prd` + "`" + `, ` + "`" + `specification` + "`" + `, ` + "`" + `guide` + "`" + `, ` + "`" + `standards` + "`" + `, ` + "`" + `glossary` + "`" + `, ` + "`" + `product` + "`" + `, ` + "`" + `technical-reference` + "`" + `): Capture active system state. Agents MUST update living specs and standards when implementation or design evolves.
   - 📜 **Immutable Decision Logs** (` + "`" + `adr` + "`" + `, ` + "`" + `improvement-proposal` + "`" + `, ` + "`" + `release-note` + "`" + `): Historical snapshots. Once accepted or published (` + "`" + `status: accepted` + "`" + ` / ` + "`" + `status: active` + "`" + `), they are strictly immutable (` + "`" + `immutable: true` + "`" + `). When architectural decisions change:
     1. Propose a NEW ADR (` + "`" + `gyrus_create_document` + "`" + ` or ` + "`" + `gyrus create` + "`" + `).
     2. Link the new ADR to supersede the old one (` + "`" + `gyrus_link_documents` + "`" + ` or ` + "`" + `gyrus link` + "`" + `).
     3. Update the old ADR status to ` + "`" + `superseded` + "`" + ` (` + "`" + `gyrus_update_document` + "`" + ` or ` + "`" + `gyrus update` + "`" + `).
3. **OKF Contract Schema Compliance**:
   - Document IDs MUST match lower-case pattern ` + "`" + `^[a-z0-9-_]+$` + "`" + ` (e.g., ` + "`" + `adr-001-storage-engine` + "`" + `).
   - Every contract requires YAML frontmatter defining ` + "`" + `id` + "`" + `, ` + "`" + `title` + "`" + `, ` + "`" + `category` + "`" + `, ` + "`" + `type` + "`" + `, ` + "`" + `owner_group` + "`" + `, ` + "`" + `version` + "`" + `, ` + "`" + `status` + "`" + `.

---

## 🛠️ Primary Interface: Model Context Protocol (MCP) Tools

Use native MCP tools whenever available:
- **` + "`" + `gyrus_suggest_context` + "`" + `**: ` + "`" + `{ "prompt": string, "limit"?: int }` + "`" + `
- **` + "`" + `gyrus_search` + "`" + `**: ` + "`" + `{ "query": string, "limit"?: int }` + "`" + `
- **` + "`" + `gyrus_get_document` + "`" + `**: ` + "`" + `{ "id": string }` + "`" + `
- **` + "`" + `gyrus_create_document` + "`" + `**: ` + "`" + `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }` + "`" + `
- **` + "`" + `gyrus_update_document` + "`" + `**: ` + "`" + `{ "id": string, "title"?: string, "status"?: string, "content"?: string }` + "`" + `
- **` + "`" + `gyrus_link_documents` + "`" + `**: ` + "`" + `{ "from_id": string, "to_id": string, "rel_type": string }` + "`" + `
- **` + "`" + `gyrus_sync` + "`" + `**: ` + "`" + `{}` + "`" + `

---

## 💻 Fallback & Admin Interface: Gyrus CLI

When running in shell-only environments, CI pipelines, or if MCP is unavailable, use the ` + "`" + `gyrus` + "`" + ` CLI:
- ` + "`" + `gyrus suggest-context --prompt "<task>" --json` + "`" + `
- ` + "`" + `gyrus search --query "<query>" --json` + "`" + `
- ` + "`" + `gyrus get <id> --json` + "`" + `
- ` + "`" + `gyrus create --id "<id>" --title "<title>" --category "<cat>" --type "<type>" --owner-group "<grp>" --content "<md>"` + "`" + `
- ` + "`" + `gyrus update <id> --status "<status>" --expected-version <v>` + "`" + `
- ` + "`" + `gyrus link <from-id> <to-id> --rel-type <type>` + "`" + `
- ` + "`" + `gyrus sync` + "`" + `

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- 🔌 **[MCP Setup Guide](references/mcp-setup.md)**
- 🗄️ **[Storage Providers](references/storage-providers.md)**
`

	embeddedSchemasRef = `# OKF Document Schemas Reference
See docs/.gyrus/schemas for complete template definitions.
`

	embeddedMcpRef = `# Gyrus MCP Setup Reference
Registers stdio server: gyrus mcp serve
`

	embeddedStorageRef = `# Gyrus Storage Providers Reference
Supports localfs, git, blob (S3/Azure/GCS), and postgres backends via .gyrus.yaml.
`
)
