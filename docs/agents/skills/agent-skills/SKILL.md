---
name: Agent Skills
description: Rules and guidelines for creating and structuring reusable agent skills in this repository.
applyTo:
  - docs/agents/skills/**
---

# Agent Skills Standard

All instruction files are structured as modular **skills** to ensure reusability, isolation, and tooling compatibility.

## Structure

Each skill must live in a dedicated subfolder under `docs/agents/skills/` and have a `SKILL.md` entry point:

```text
docs/agents/skills/<skill-name>/
└── SKILL.md
```

## Writing `SKILL.md`

Every `SKILL.md` file must adhere to these guidelines:
1. **YAML Frontmatter**: Start the file with a YAML frontmatter block containing metadata:
   ```yaml
   ---
   name: Skill Name
   description: 1-2 sentence summary of when and why this skill is active.
   applyTo:
     - path/to/target/**
   ---
   ```
   * `name`: The human-readable name of the skill.
   * `description`: The purpose of this skill.
   * `applyTo`: YAML list of glob patterns mapping this skill to specific workspace paths for Copilot. Use `**` for repository-wide.
2. **Markdown formatting**: Use standard Markdown format for headings, bullet points, and code samples.
