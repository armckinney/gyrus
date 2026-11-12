<p align="center">
  <img src="internal/ui/static/img/logo.png" alt="Gyrus Logo" width="150" />
</p>

# Gyrus: Intelligence Platform

Gyrus is an intelligence platform that keeps engineering knowledge available across repositories and agent sessions. Maintain decisions, specifications, and guides as structured concepts with explicit relationships, then use the same knowledge through the CLI, local MCP server, or local UI.

Choose how knowledge changes are reviewed: commit directly to a Git trunk branch, or collect an initiative on a branch for human review.

## Choose a profile

| Profile | Implementation | Use when |
| --- | --- | --- |
| `local` | Local filesystem with OKF index, graph, and keyword search | Knowledge belongs alongside one code repository |
| `git` | Remote Git repository with OKF index, graph, and keyword search | Several code repositories share knowledge and Git history/review |
| `postgres` | PostgreSQL storage, index, graph, and native full-text search | A team wants shared database-backed knowledge |

## Get started

Build from source with the Go version specified in `go.mod`:

```bash
git clone https://github.com/armckinney/gyrus.git
cd gyrus
make build
./gyrus --help
```

Place the compiled binary on your `PATH`, then initialize configuration in your code repository:

```bash
gyrus config init --profile local --tenant root --project my-project
gyrus config show
```

Add `--global` to initialize `~/.gyrus.yaml`. A repository `.gyrus.yaml` replaces global configuration entirely. Relative local storage roots resolve from the configuration file's directory. Invalid files, unknown fields, and unsupported provider combinations fail explicitly.

```bash
gyrus create --id first-guide --title "First guide" --category technical \
  --type guide --tenant root --scope project --project my-project \
  --status active --content "Engineering knowledge for this project."
gyrus get root/projects/my-project/first-guide
gyrus search --query engineering
gyrus suggest-context --prompt "Plan the next change"
gyrus mcp serve
```

Creation uses a local ID. Reads, updates, and relationship endpoints use qualified names such as `root/reference/shared-standard` or `root/projects/my-project/first-guide`. Source-owned typed links connect concepts. Keyword searches default to ten results; metadata browsing is unlimited unless capped.

## Shared Git knowledge

Initialize with `gyrus config init --profile git`, then set the remote URL. Code repositories can point to the same knowledge repository while selecting different project contexts.

**Direct commits** read and write trunk:

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
  repo_url: https://github.com/OWNER/wiki.git
  trunk_branch: main
  commit_strategy: direct
```

**Branching commits** collect an initiative for human review. Change the Git section to:

```yaml
git:
  repo_url: https://github.com/OWNER/wiki.git
  trunk_branch: main
  commit_strategy: branching
```

Supply the initiative branch at runtime:

```bash
gyrus search --query architecture --git-branch initiative/payments
gyrus mcp serve --git-branch initiative/payments
```

The branch name is never saved in configuration. Each MCP server remembers its own selection; agents can call `gyrus_git_switch` to change initiatives without restarting. A server started without a required selection accepts Git status/switch tools and rejects knowledge operations until a branch is selected. Separate CLI calls supply `--git-branch` each time. The flag also selects branching when configuration defaults to direct.

A new initiative initially reads trunk and is pushed on its first mutation. Trunk must already exist for branching. Existing initiative branches are reused. All concepts, relationships, schemas, and queries use the selected branch. Humans open, review, and merge PRs through their Git host; Gyrus never automatically merges, rebases, or force-pushes.

HTTPS authentication uses your configured Git credential manager through noninteractive `git credential fill`. SSH uses an agent or standard keys with known-host verification. Do not embed passwords in remote URLs. See the [initiative walkthrough](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/guide-006-git-initiatives.md).

## Agent integration

The agent plugin is packaged under `packaging/plugins/gyrus`. Install it using your client's native plugin manager; see the [installation guide](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/guide-005-installation-and-initialization.md). MCP runs locally over stdio. Each parallel initiative should use its own server process or explicit runtime selection.

## Documentation

This repository uses [armckinney/wiki](https://github.com/armckinney/wiki) as its Gyrus Git storage backend. Its [project documentation](https://github.com/armckinney/wiki/tree/main/.gyrus/concepts/root/projects/gyrus) is scoped to project `gyrus`. The checked-in [configuration](.gyrus.yaml) sends Gyrus writes directly to the wiki repository's `main` branch.

- [Setup and onboarding](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/guide-002-setup-onboarding.md)
- [Configuration reference and standard examples](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/tech-ref-002-config-schema.md)
- [CLI reference](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/tech-ref-001-cli-manual.md)
- [MCP reference](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/tech-ref-003-mcp-manual.md)
- [Architecture index](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/spec-000-overview.md)
- [Value proposition](https://github.com/armckinney/wiki/blob/main/.gyrus/concepts/root/projects/gyrus/prd-002-value-proposition-positioning.md)

See [LICENSE](LICENSE) for licensing terms.
