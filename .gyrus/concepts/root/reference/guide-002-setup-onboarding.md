---
title: Team Onboarding & Context Setup Guide
id: guide-002-setup-onboarding
type: guide
format: markdown
category: technical
version: 6
status: active
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

# Set up a knowledge collection

Choose `local` for knowledge committed with one code repository, `git` for a shared remote knowledge repository, or `postgres` for shared database storage and full-text search. Only these complete stacks are supported.

```bash
gyrus config init --profile local --tenant root --project my-project
gyrus config show
```

Use `--profile git` or `--profile postgres` for the other stacks, then edit connection settings. Add `--global` to write user-wide defaults; a repository configuration replaces them entirely. See [standard configurations](tech-ref-002-config-schema.md).

## Create and connect knowledge

```bash
gyrus create --id guide-first --title "First guide" --category technical --type guide --tenant root --scope project --project my-project --status active --content "Shared engineering knowledge."
gyrus get root/projects/my-project/guide-first
gyrus link root/projects/my-project/guide-first root/reference/shared-standard --rel-type depends_on
gyrus search --query engineering
gyrus suggest-context --prompt "Plan the next change"
```

Creation uses a local ID; subsequent operations require the qualified name. Relationships persist as source-owned `links` containing `concept` and `rel`. Missing targets are allowed; incoming neighbors supply backlinks.

## Shared Git and human review

Configure `git.repo_url`, `git.trunk_branch`, and `git.commit_strategy`. Direct writes go to trunk. For human review choose `branching`, then supply an initiative:

```bash
gyrus search --query engineering --git-branch initiative/my-change
gyrus mcp serve --git-branch initiative/my-change
```

All changes in that session collect on the initiative branch. A second server can use a different branch. Use `gyrus_git_switch` to change an active MCP selection; restart requires selecting the branch again. Open and merge a PR through your Git host. Gyrus performs no automatic merges. Configure Git Credential Manager for HTTPS or an SSH agent/key and known hosts for SSH. See the [Git walkthrough](guide-006-git-initiatives.md).

## Upgrading

Replace old configurations with one supported stack. Removed cloud providers, hybrid stacks, `git.branch`, and unknown YAML fields now fail explicitly. Export or relocate existing data separately before changing stores; no automatic migration is provided. The packaged legacy SQLite migration skill remains available when an old installation needs its database-only relationships exported.

Install the agent plugin using the [installation guide](guide-005-installation-and-initialization.md). Re-run `config show` and restart MCP servers after upgrading.
