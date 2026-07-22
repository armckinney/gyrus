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
