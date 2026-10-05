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
	NoHooks        bool
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
		opts.Mode = MCPModeLocal
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
				"env": map[string]string{
					"GYRUS_WORKSPACE": "${workspaceFolder}",
				},
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
				"env": map[string]string{
					"GYRUS_WORKSPACE": "${workspaceFolder}",
				},
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

	// 5. Write hooks.json (Antigravity & Agent Plugins lifecycle hooks)
	if !opts.NoHooks {
		hooksPayload := map[string]interface{}{
			"gyrus-automation": map[string]interface{}{
				"PreInvocation": []map[string]interface{}{
					{
						"type":    "command",
						"command": opts.BinaryCmd + " hook pre-invocation",
						"timeout": 15,
					},
				},
				"Stop": []map[string]interface{}{
					{
						"type":    "command",
						"command": opts.BinaryCmd + " hook stop",
						"timeout": 15,
					},
				},
			},
		}
		hooksBytes, err := json.MarshalIndent(hooksPayload, "", "  ")
		if err != nil {
			return nil, err
		}
		hooksPath := filepath.Join(opts.TargetDir, "hooks.json")
		if err := os.WriteFile(hooksPath, hooksBytes, 0644); err != nil {
			return nil, err
		}
		createdFiles = append(createdFiles, hooksPath)
	}

	// 6. Write rules/AGENTS.md
	rulesPath := filepath.Join(rulesDir, "AGENTS.md")
	if err := os.WriteFile(rulesPath, []byte(embeddedPluginRules), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, rulesPath)

	// 7. Write skills/gyrus
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

	// 8. Write skills/gyrus-index
	indexSkillDir := filepath.Join(opts.TargetDir, "skills", "gyrus-index")
	if err := os.MkdirAll(indexSkillDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating skill dir %s: %w", indexSkillDir, err)
	}

	indexSkillFile := filepath.Join(indexSkillDir, "SKILL.md")
	if err := os.WriteFile(indexSkillFile, []byte(embeddedIndexSkillMD), 0644); err != nil {
		return nil, err
	}
	createdFiles = append(createdFiles, indexSkillFile)

	return createdFiles, nil
}

const embeddedPluginRules = `# Gyrus Context Control Plane & SDLC Directives

This document instructs AI agents (Antigravity, Claude Code, GitHub Copilot, OpenAI Codex) on interacting with the Gyrus Context Control Plane and working within the OKF documentation framework.

---

## 💡 Core Architectural Principles

1. **Source Code vs. Gyrus Context Plane**:
   - **Source Code Owns Implementation**: Function logic, routine bug fixes, internal refactoring, and code-level changes belong exclusively in the codebase. Never duplicate routine code edits, method implementations, or commit diffs in Gyrus.
   - **Gyrus Owns Architectural Intent & System Governance**: Gyrus captures the high-level context required to piece together current architecture and guide future changes across the engineering lifecycle.

2. **The Gyrus 3-Phase SDLC Lifecycle (from standards-001)**:
   - **Phase 1: Plan (Architecture & Design)**:
     - ` + "`prd`" + `: Business requirements, motivation, success metrics (*what & why*).
     - ` + "`improvement-proposal`" + `: Design alternatives, system architecture, testing strategy (*how*).
     - ` + "`specification`" + `: Final system blueprints, data models, protocols, component boundaries.
     - ` + "`adr`" + `: Short audit records of micro-decisions and tradeoffs (*why this way*).
     - ` + "`standards`" + `: Engineering quality rules, linting policies, and git conventions.
   - **Phase 2: Implement (Build & Interfaces)**:
     - *Source Code*: Canonical source of truth for implementation logic.
     - ` + "`technical-reference`" + `: Public API endpoints, CLI syntax, config parameters, and technical FAQs.
     - ` + "`guide`" + `: Developer onboarding tutorials, runbooks, and troubleshooting walkthroughs.
   - **Phase 3: Post-Implement (Release & Portal)**:
     - ` + "`release-note`" + `: Version release tracking, changelogs, migration guides.
     - ` + "`product`" + `: Main product homepage portal summarizing value proposition and doc navigation.

3. **Context Resolution Before Architecture Changes**:
   - Before designing new features or making architectural decisions, resolve relevant context:
     - Primary: Native MCP ` + "`gyrus_suggest_context({ prompt: \"<task>\" })`" + ` or ` + "`gyrus_search`" + `.
     - Fallback (CLI): ` + "`gyrus suggest-context --prompt \"<task>\" --json`" + `.

4. **Document Mutability & Immutability Rules**:
   - 🌿 **Living Documents** (` + "`prd`, `specification`, `standards`, `technical-reference`, `guide`, `product`" + `): Capture active system state. Update living specs and standards when architecture or public interfaces evolve.
   - 📜 **Immutable Decision Logs** (` + "`adr`, `improvement-proposal`, `release-note`" + `): Historical snapshots. Once accepted (` + "`status: accepted`" + `), they are strictly immutable. When architectural decisions change:
     1. Propose a NEW ADR (` + "`gyrus create`" + `).
     2. Link the new ADR to supersede the old one (` + "`gyrus link <new-id> <old-id> --rel-type supersedes`" + `).
     3. Update the old ADR status to ` + "`superseded`" + ` (` + "`gyrus update <old-id> --status superseded`" + `).

5. **OKF Contract Schema Compliance**:
   - Document IDs MUST match lower-case regex ` + "`^[a-z0-9-_]+$`" + ` (e.g., ` + "`adr-001-storage-engine`" + `).
   - Every contract document requires YAML frontmatter defining ` + "`id`, `title`, `category`, `type`, `owner_group`, `version`, `status`" + `.

6. **Programmatic Exit Codes Protocol**:
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
   - Propose an Improvement Proposal (` + "`improvement-proposal`" + `) for major system designs.
   - Record immutable Architectural Decision Records (` + "`adr`" + `) for key choices and tradeoffs.
   - Maintain living system specifications (` + "`specification`" + `) and standards (` + "`standards`" + `).
2. **Phase 2: Implement (Build & Interfaces)**:
   - Implement source code to satisfy specifications. Source code is the truth for code; never duplicate code diffs in docs.
   - Update Technical Reference (` + "`technical-reference`" + `) when public APIs, CLI syntax, or configs change.
   - Update Guides (` + "`guide`" + `) when setup runbooks or onboarding flows change.
3. **Phase 3: Post-Implement (Release & Portal)**:
   - Draft Release Notes (` + "`release-note`" + `) for version milestones.
   - Maintain the Product Page (` + "`product`" + `) doc portal.

---

## 🛠️ Primary Interface: Model Context Protocol (MCP) Tools

- **` + "`" + `gyrus_suggest_context` + "`" + `**: ` + "`" + `{ "prompt": string, "limit"?: int }` + "`" + `
- **` + "`" + `gyrus_search` + "`" + `**: ` + "`" + `{ "query": string, "limit"?: int }` + "`" + `
- **` + "`" + `gyrus_get_document` + "`" + `**: ` + "`" + `{ "id": string }` + "`" + `
- **` + "`" + `gyrus_create_document` + "`" + `**: ` + "`" + `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }` + "`" + `
- **` + "`" + `gyrus_update_document` + "`" + `**: ` + "`" + `{ "id": string, "title"?: string, "status"?: string, "content"?: string }` + "`" + `
- **` + "`" + `gyrus_link_documents` + "`" + `**: ` + "`" + `{ "from_id": string, "to_id": string, "rel_type": string }` + "`" + `
- **` + "`" + `gyrus_sync` + "`" + `**: ` + "`" + `{}` + "`" + `

---

## 💻 Fallback & Admin Interface: Gyrus CLI

- ` + "`" + `gyrus suggest-context --prompt "<task>" --json` + "`" + `
- ` + "`" + `gyrus search --query "<query>" --json` + "`" + `
- ` + "`" + `gyrus get <id> --json` + "`" + `
- ` + "`" + `gyrus create --id "<id>" --title "<title>" --category "<cat>" --type "<type>" --owner-group "<grp>" --content "<md>"` + "`" + `
- ` + "`" + `gyrus update <id> --status "<status>" --expected-version <v>` + "`" + `
- ` + "`" + `gyrus link <from-id> <to-id> --rel-type <type>` + "`" + `
- ` + "`" + `gyrus sync` + "`" + `
`

	embeddedIndexSkillMD = `---
name: gyrus-index
description: Gyrus Repository Primer skill. Interactively explores codebase architecture, dependencies, and standards, proposes synthesized baseline living documents (spec, standards, technical reference, onboarding guide), confirms with developer, and indexes them into Gyrus.
applyTo:
  - "**"
---

# Gyrus Repository Primer Skill (` + "`/gyrus-index`" + `)

This skill guides the AI agent to survey a new or existing codebase and prime Gyrus with high-level architectural living documents without duplicating low-level code implementation details.

---

## 📋 The Primer Workflow

When the developer invokes ` + "`/gyrus-index`" + `, follow this 5-step interactive procedure:

### Step 1: Survey Codebase Architecture
- Examine the root directory, package manifests (` + "`go.mod`" + `, ` + "`package.json`" + `, ` + "`Cargo.toml`" + `, etc.), README, build configs, and entry points.
- Identify the project name, primary language, architectural style (e.g. modular monolith, microservices, CLI tool), and external dependencies.

### Step 2: Propose Baseline Living Documents
Formulate a concise, synthesized proposal of 3-4 core living documents for the repository:
1. **` + "`spec-001-architecture.md`" + `** (` + "`type: specification`" + `): System topology, high-level component boundaries, and data flows.
2. **` + "`standards-001-guidelines.md`" + `** (` + "`type: standards`" + `): Engineering quality standards, code organization, testing strategy, and git conventions.
3. **` + "`tech-ref-001-api-cli.md`" + `** (` + "`type: technical-reference`" + `): Public CLI commands, API interfaces, configuration parameters, or technical lookups.
4. **` + "`guide-001-onboarding.md`" + `** (` + "`type: guide`" + `): Developer environment setup, build instructions, and common runbooks.

### Step 3: Developer Review & Confirmation
- Present the synthesized proposal to the developer with a clear overview of what each document will capture.
- Ask for confirmation or adjustments before creating any files.

### Step 4: Scaffold Confirmed Documents
- Once approved, create the documents in ` + "`.gyrus/docs/<owner_group>/reference/`" + ` using valid OKF YAML frontmatter:
  - ` + "`id`" + `, ` + "`title`" + `, ` + "`category`" + `, ` + "`type`" + `, ` + "`owner_group`" + `, ` + "`version: 1`" + `, ` + "`status: active`" + `.
- Ensure documents capture **high-level intent and boundaries**, not line-by-line function implementations.

### Step 5: Index & Synchronize
- Run ` + "`gyrus sync`" + ` to parse documents, build the SQLite FTS search index, and construct dependency edges.
- Inform the developer that Gyrus is primed and ready to provide architectural context to all AI agents.
`

	embeddedSchemasRef = `# OKF Document Schemas Reference
See docs/.gyrus/schemas for complete template definitions.
`
)
