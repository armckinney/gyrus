---
title: Gyrus YAML Configuration Reference & Schema
id: tech-ref-002-config-schema
type: technical-reference
format: markdown
category: technical
version: 6
status: active
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

# Configuration reference

Generate `.gyrus.yaml` with `gyrus config init --profile local|git|postgres`. Add `--global` for `~/.gyrus.yaml`; inspect with `gyrus config show`. Unknown profiles fail before writing files.

## Supported stacks

| Profile | Storage | Index | Graph | Search | Use when |
| --- | --- | --- | --- | --- | --- |
| local | localfs | okf | okf | okf | Knowledge lives alongside one code repository |
| git | git | okf | okf | okf | Several code repositories share one remote knowledge repository |
| postgres | postgres | postgres | postgres | postgres_fts | A team needs shared database-backed storage and native full-text search |

Only these complete stacks are supported. Omitting secondary providers selects the stack appropriate to storage. Hybrid combinations, cloud/blob providers, and their aliases are rejected before opening resources. No cache provider is implemented.

## Standard configurations

Local (default):

```yaml
tenant: root
project: my-project
storage:
  provider: localfs
  root: .gyrus
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
```

Git, direct commits (default Git strategy):

```yaml
tenant: root
project: my-project
storage:
  provider: git
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
git:
  repo_url: https://github.com/OWNER/knowledge.git
  trunk_branch: main
  commit_strategy: direct
```

Git, initiative branches with human review: use the same Git configuration with this section:

```yaml
git:
  repo_url: https://github.com/OWNER/knowledge.git
  trunk_branch: main
  commit_strategy: branching
```

Select the initiative at runtime:

```bash
gyrus search --query architecture --git-branch initiative/payments
gyrus mcp serve --git-branch initiative/payments
```

PostgreSQL:

```yaml
tenant: root
project: my-project
storage:
  provider: postgres
index:
  provider: postgres
graph:
  provider: postgres
search:
  provider: postgres_fts
postgres:
  connection_string: "postgres://USER:PASSWORD@HOST:5432/gyrus?sslmode=require"
```

The generated PostgreSQL template targets the development-container database; replace its connection string for your deployment. Protect configuration files containing database credentials.

## Git branch contract

`trunk_branch` defaults to `main`; `commit_strategy` defaults to `direct` and accepts only `direct` or `branching`. Direct mode reads and writes trunk. `--git-branch` selects branching for that invocation regardless of the configured default. The initiative must differ from trunk.

Branch names are never persisted in configuration. Each MCP server remembers its own active selection. `gyrus_git_switch` accepts `strategy` and a `branch` required for branching; selecting direct omits `branch`. `gyrus_git_status` and `gyrus git status` report `strategy`, `trunk_branch`, `active_branch`, and `selection_required`.

A branching MCP server may start without a selection; knowledge tools/resources/prompts fail until an initiative is selected. A missing initiative initially reads trunk and becomes a remote branch on its first mutation. Trunk must exist for branching. Direct mode can initialize an empty remote. Existing initiative branches are reused without integrating newer trunk changes. Humans create/review/merge PRs and integrate trunk changes through their Git host. Gyrus never automatically merges, rebases, or force-pushes.

All concepts, schemas, searches, and links use the selected branch. Git layout is `.gyrus/concepts/<qualified-name>.md` and `.gyrus/schemas/<type>.md`. `storage.root` applies to local filesystem storage, not the remote Git layout. HTTPS credentials come from noninteractive `git credential fill`; SSH uses an agent or standard unencrypted key files with known-host verification. Passwords must not appear in `repo_url`. `GIT_AUTH_TOKEN` and `GITHUB_TOKEN` are not provider authentication inputs.

## Other settings and discovery

`tenant` defaults to `root`; `project` selects project context. Concept paths determine stored scope. `ui.client` and `ui.command` configure the existing local UI agent. `schemas_path` remains a recognized configuration field; schema persistence uses the selected storage provider.

Explicit `GYRUS_CONFIG` selects a file. `GYRUS_PROJECT` selects a project directory. Otherwise search upward for `.gyrus.yaml`; the existing container-discovery fallback may select a project under `/workspaces`, followed by global configuration and defaults. Project files replace global configuration entirely. Relative local roots resolve from their configuration file's directory. Present malformed files, unknown keys, and multiple YAML documents are errors rather than fallback triggers.

## Breaking changes

Replace `git.branch` with `git.trunk_branch`. Remove `base_branch`, `review_branch`, cloud/blob settings, provider aliases, and obsolete configuration fields. Choose one supported stack; move old cloud data to the new store separately. Replace token environment variables with Git credential-manager setup. Restart MCP servers after upgrading and supply initiative names at runtime. No legacy aliases or automatic migrations are provided.

OKF reads concepts and typed links directly; there is no persistent index database or sync command. Keyword search defaults to ten results; metadata-only browsing is unlimited unless capped. PostgreSQL supplies native FTS with the same result-limit contract. See the [contract specification](spec-001-contract-schema.md) for qualified identities and links.
