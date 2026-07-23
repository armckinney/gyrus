---
id: guide-003-agent-skills-setup
title: Gyrus Agent Skills Setup Guide
category: technical
type: guide
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus Agent Skills Setup Guide

This guide explains how to install the Gyrus Agent Skill so AI coding assistants (**Antigravity CLI**, **GitHub Copilot**, **Claude**, and **Codex**) can read and maintain your codebase memory.

---

## 1. Quick Skill Installation

Copy `.agents/skills/gyrus/SKILL.md` into your repository:

```bash
mkdir -p .agents/skills/gyrus
curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/.agents/skills/gyrus/SKILL.md -o .agents/skills/gyrus/SKILL.md
```

---

## 2. Tool Integration Matrix

Because the Open Skill Format uses a standardized `.agents/skills/` directory structure, skill loading is identical for auto-discovering tools:

| AI Tool / Harness | Discovery Mechanism | Registration Instructions |
| :--- | :--- | :--- |
| **Google Antigravity CLI (AGY)** | Auto-discovered | Automatically reads `.agents/skills/gyrus/SKILL.md` in workspace or `~/.gemini/antigravity-cli/skills/`. |
| **Claude Code CLI** | Auto-discovered | Automatically reads `.agents/skills/gyrus/SKILL.md` in workspace. |
| **GitHub Copilot** | Instruction file link | Add reference to `.github/copilot-instructions.md`:<br>`Consult Gyrus via gyrus suggest-context before modifying code.` |
| **OpenAI Codex & Custom Agents** | Shell / Prompt Harness | Run `CONTEXT=$(gyrus suggest-context --prompt "<task>")` before prompt submission. |

---

## 3. Document Mutability & Lifecycle Rules

AI Agents must adhere to these rules when interacting with Gyrus codebase memory:

1. 🌿 **Living Documents (`prd`, `specification`, `guide`, `standards`, `glossary`, `product`, `technical-reference`, `freeform`):** Represent the **current active state** of the system. Agents MUST actively update living specifications, guides, and standards whenever codebase implementation or architecture changes.
2. 📜 **Immutable Decision Logs (`adr`, `improvement-proposal`, `release-note`):** Represent **historical snapshots**. Once accepted or published (`status: accepted` / `status: active`), agents MUST NOT modify historical ADRs or proposals. When design choices change:
   - Create a NEW ADR or proposal (`gyrus create`).
   - Link the new document to supersede the old one (`gyrus link <new-id> <old-id> --rel-type supersedes`).
   - Update the old document status to `superseded` (`gyrus update <old-id> --status superseded`).
3. ⚙️ **Custom Immutable Templates (`immutable: true`):** Users can mark custom document types (e.g. `security-audit`, `compliance-report`, `incident-postmortem`) as immutable by adding `immutable: true` in the document frontmatter header. The Gyrus Core Engine will enforce content immutability once the document exits `draft`/`proposed` status.
