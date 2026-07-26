package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// SkillTarget identifies an AI agent platform for skill installation.
type SkillTarget string

const (
	SkillTargetAll         SkillTarget = "all"
	SkillTargetAntigravity SkillTarget = "antigravity"
	SkillTargetClaude      SkillTarget = "claude"
	SkillTargetCodex       SkillTarget = "codex"
	SkillTargetCopilot     SkillTarget = "copilot"
)

// InstallAgentSkill equips the target workspace directory with both Gyrus CLI and MCP Agent Skill files.
func InstallAgentSkill(workspaceDir string, target SkillTarget) ([]string, error) {
	skillTypes := []struct {
		folderName string
		content    string
	}{
		{folderName: "gyrus-cli", content: embeddedSkillMD},
		{folderName: "gyrus-mcp", content: embeddedMCPSkillMD},
	}

	var createdFiles []string

	for _, st := range skillTypes {
		skillDirs := []string{
			filepath.Join(workspaceDir, ".agents", "skills", st.folderName),
			filepath.Join(workspaceDir, "skills", st.folderName),
		}

		for _, dir := range skillDirs {
			refDir := filepath.Join(dir, "references")

			if err := os.MkdirAll(refDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create skill references dir %s: %w", refDir, err)
			}

			// Write SKILL.md
			skillFile := filepath.Join(dir, "SKILL.md")
			if err := os.WriteFile(skillFile, []byte(st.content), 0644); err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, skillFile)

			// Write references/okf-schemas.md
			schemasRef := filepath.Join(refDir, "okf-schemas.md")
			if err := os.WriteFile(schemasRef, []byte(embeddedSchemasRef), 0644); err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, schemasRef)

			// Write references/mcp-setup.md
			mcpRef := filepath.Join(refDir, "mcp-setup.md")
			if err := os.WriteFile(mcpRef, []byte(embeddedMcpRef), 0644); err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, mcpRef)

			// Write references/storage-providers.md
			storageRef := filepath.Join(refDir, "storage-providers.md")
			if err := os.WriteFile(storageRef, []byte(embeddedStorageRef), 0644); err != nil {
				return nil, err
			}
			createdFiles = append(createdFiles, storageRef)
		}
	}

	return createdFiles, nil
}

const embeddedSkillMD = `---
name: gyrus-cli
description: Gyrus Unified Context & Memory Engine CLI agent skill. Use to search, retrieve, create, update, link, and suggest relevant OKF codebase context for tasks via gyrus CLI subcommands.
applyTo:
  - "**"
---

# Gyrus Agent Skill Specification & CLI Reference

This skill equips AI agents (Antigravity, Claude Code, Codex, GitHub Copilot) to interact directly with Gyrus codebase memory via the ` + "`" + `gyrus` + "`" + ` CLI executable.

---

## 💡 Core Agent Guidelines

1. **Before Modifying Code:** Always run ` + "`" + `gyrus suggest-context --prompt "<task description>"` + "`" + ` or ` + "`" + `gyrus search --query "<keyword>"` + "`" + ` to read relevant ADRs and technical contracts.
2. **Machine Parsing:** Pass global ` + "`" + `--json` + "`" + ` flag to receive structured JSON envelopes instead of terminal formatted text.
3. **ID Naming Rule:** Document IDs MUST match lower-case pattern ` + "`" + `^[a-z0-9-_]+$` + "`" + ` (e.g., ` + "`" + `adr-001-storage-engine` + "`" + `).
4. **Exit Codes Protocol:** ` + "`" + `0` + "`" + `: Success, ` + "`" + `1` + "`" + `: Schema/ID validation error, ` + "`" + `2` + "`" + `: Illegal state transition, ` + "`" + `3` + "`" + `: Permission error, ` + "`" + `4` + "`" + `: Lock conflict, ` + "`" + `5` + "`" + `: Record/Storage error.

---

## 🛠️ Complete CLI Command Reference

### 1. ` + "`" + `gyrus suggest-context` + "`" + `
Linearizes top relevant documents matching a task prompt (Recommended first step).

### 2. ` + "`" + `gyrus search` + "`" + `
Executes FTS5 lexical keyword search across documents.

### 3. ` + "`" + `gyrus get` + "`" + `
Retrieves a single document by ID.

### 4. ` + "`" + `gyrus create` + "`" + `
Creates a new OKF contract document.

### 5. ` + "`" + `gyrus update` + "`" + `
Patches metadata fields or content of an existing document.

### 6. ` + "`" + `gyrus link` + "`" + ` / ` + "`" + `gyrus unlink` + "`" + `
Creates or removes a directed relationship edge between two documents.

### 7. ` + "`" + `gyrus sync` + "`" + `
Re-indexes filesystem documents and extracts dependency links.

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- 🔌 **[MCP Setup Guide](references/mcp-setup.md)**
- 🗄️ **[Storage Providers](references/storage-providers.md)**
`

const embeddedMCPSkillMD = `---
name: gyrus-mcp
description: Gyrus Unified Context & Memory Engine MCP skill. Use native Model Context Protocol (MCP) tools and resources to search, retrieve, create, update, link, and suggest relevant OKF codebase context for tasks.
applyTo:
  - "**"
---

# Gyrus MCP Agent Skill Specification

This skill equips MCP-enabled AI agents (Cursor, Claude Desktop, OpenAI Codex, GitHub Copilot, Windsurf) to interact directly with Gyrus codebase memory via native Model Context Protocol (MCP) tool calls and resource endpoints.

---

## 💡 Core Agent Guidelines

1. **Before Modifying Code:** Always invoke ` + "`" + `gyrus_suggest_context({ prompt: "<task description>" })` + "`" + ` or ` + "`" + `gyrus_search({ query: "<keyword>" })` + "`" + ` to read relevant ADRs and technical contracts.
2. **ID Naming Rule:** Document IDs MUST match the lower-case pattern ` + "`" + `^[a-z0-9-_]+$` + "`" + ` (e.g., ` + "`" + `adr-001-storage-engine` + "`" + `).
3. **Linkage & Dependency Graph:** When creating or mutating contracts, link dependent documents using ` + "`" + `gyrus_link_documents({ from_id: "...", to_id: "...", rel_type: "depends_on" })` + "`" + `.

---

## 🛠️ Complete MCP Tool Reference

### 1. ` + "`" + `gyrus_suggest_context` + "`" + `
Linearizes top relevant documents matching a task prompt (Recommended first step before writing code).
- **Arguments:** ` + "`" + `{ "prompt": string, "limit"?: int }` + "`" + `

### 2. ` + "`" + `gyrus_search` + "`" + `
Executes FTS5 lexical keyword search across codebase context documents.
- **Arguments:** ` + "`" + `{ "query": string, "limit"?: int }` + "`" + `

### 3. ` + "`" + `gyrus_get_document` + "`" + `
Retrieves a single document payload by ID.
- **Arguments:** ` + "`" + `{ "id": string }` + "`" + `

### 4. ` + "`" + `gyrus_create_document` + "`" + `
Creates a new OKF contract document in context storage.
- **Arguments:** ` + "`" + `{ "id": string, "title": string, "category": string, "type": string, "owner_group": string, "status": string, "content": string }` + "`" + `

### 5. ` + "`" + `gyrus_update_document` + "`" + `
Patches metadata fields or content of an existing contract document.
- **Arguments:** ` + "`" + `{ "id": string, "title"?: string, "status"?: string, "content"?: string }` + "`" + `

### 6. ` + "`" + `gyrus_link_documents` + "`" + `
Creates a directed relationship edge between two documents (` + "`" + `depends_on` + "`" + `, ` + "`" + `supersedes` + "`" + `, ` + "`" + `implements` + "`" + `).
- **Arguments:** ` + "`" + `{ "from_id": string, "to_id": string, "rel_type": string }` + "`" + `

### 7. ` + "`" + `gyrus_sync` + "`" + `
Re-indexes filesystem documents and extracts dependency links into the FTS database.
- **Arguments:** ` + "`" + `{}` + "`" + `

---

## 🔌 MCP Resource Endpoints

- ` + "`" + `gyrus://documents/{id}` + "`" + `: Direct document payload lookup.
- ` + "`" + `gyrus://schema/{type}` + "`" + `: OKF document template schema definition.
- ` + "`" + `gyrus://graph/topology` + "`" + `: Complete dependency graph topology.

---

## 📚 Skill Reference Guides

- 📄 **[OKF Schemas Reference](references/okf-schemas.md)**
- 🔌 **[MCP Setup Guide](references/mcp-setup.md)**
- 🗄️ **[Storage Providers](references/storage-providers.md)**
`

const embeddedSchemasRef = `# OKF Document Schemas Reference
See docs/.gyrus/schemas for complete template definitions.
`

const embeddedMcpRef = `# Gyrus MCP Setup Reference
Registers stdio server: gyrus mcp serve
`

const embeddedStorageRef = `# Gyrus Storage Providers Reference
Supports localfs, git, blob (S3/Azure/GCS), and postgres backends via .gyrus.yaml.
`
