---
title: Cobra CLI Executable & Command Line Interface Specs
id: spec-002-cli-interface
type: specification
format: markdown
category: technical
version: 5
status: active
last_updated: 2026-11-12T05:49:34Z
last_modified_by: ""
tenant: root
---

# CLI interface specification

The compiled `gyrus` command exposes configuration, concept lifecycle, search/context, relationships, schemas, Git session status, local MCP, and the local UI. Commands delegate to the shared lifecycle engine and selected complete storage stack. This specification describes current arguments, defaults, outputs, and failure behavior; the [CLI reference](tech-ref-001-cli-manual.md) provides companion examples.

## Invocation, configuration, and identity

```text
gyrus [global flags] <command> [arguments] [command flags]
```

| Global flag | Type / default | Behavior |
| --- | --- | --- |
| `--json` | Boolean, false | Request the command's machine-readable output |
| `--verbose` | Boolean, false | Existing verbose/debug option |
| `--git-branch <name>` | String, unset | Runtime initiative selection for knowledge commands and MCP startup; selects `branching` and never writes configuration |
| `-h`, `--help` | Boolean | Display command help |

Commands use normal configuration discovery; `GYRUS_CONFIG` and `GYRUS_PROJECT` can explicitly select configuration. MCP/UI server commands also have their own selectors. Project files replace global settings rather than merging them. Local/Git/PostgreSQL complete stacks are validated before resource creation. See the [configuration reference](tech-ref-002-config-schema.md) for discovery, strict loading, and standard examples.

Creation takes a local slug with tenant/scope/project. Existing-concept arguments take full names such as `root/reference/adr-storage` or `root/projects/payments/guide-deploy`. Short-ID lookup is unsupported. Safe identifiers, supported types/categories, and minimal source-owned links follow the [concept contract](spec-001-contract-schema.md).

## Configuration commands

### `gyrus config init`

Generate `.gyrus.yaml` in the current directory, or `~/.gyrus.yaml` with `--global`.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-p`, `--profile` | `local` | `local`, `git`, or `postgres` |
| `-t`, `--tenant` | `root` | Default tenant |
| `-P`, `--project` | `default` | Default project |
| `-g`, `--global` | false | Write the global file |
| `-f`, `--force` | false | Permit replacement of an existing global file |

The current project writer replaces the target configuration file; the global writer checks `--force` before replacement. Unsupported profiles fail before writes. Git generation defaults to trunk `main` and strategy `direct`; remote/database connection settings need to be supplied for actual use.

`--json` returns a configuration result with status, file path, profile, storage, tenant, project, and scope. Plain output describes the generated configuration.

### `gyrus config show`

Display the effective normalized configuration and source. JSON contains `source`, `source_path`, `storage_root`, and `config`. This reports saved configuration; use Git status for the current runtime branch.

## Concept lifecycle commands

### `gyrus create`

Required flags: `--id`, `--title`, `--category`, and `--type`.

| Optional flag | Behavior |
| --- | --- |
| `--tenant` | Configured tenant or `root` |
| `--scope` | Explicit `project` or `reference` scope |
| `--project` | Explicit project; engine defaults can supply the configured project |
| `--status` | Defaults to `proposed` for ADRs, `draft` for other types |
| `--tags` | Comma-separated tags |
| `--content` | Inline Markdown body |
| `--content-file` | Body from a file, taking precedence over inline content |

```bash
gyrus create --id guide-deploy --title "Deployment guide" \
  --category technical --type guide --tenant root \
  --scope project --project payments --status active \
  --content-file ./deployment-body.md --json
```

Plain output identifies the qualified name, version, and status. JSON returns a `ConceptRef`: `name`, local `id`, `version`, `status`, `last_updated`.

### `gyrus get <qualified-name>`

Retrieve a concept as Markdown with frontmatter/body, or a full JSON `Concept` with `--json`. The JSON envelope includes derived `name`, `scope`, project where applicable, storage filepath where available, metadata, and `content`; Markdown keeps the local frontmatter ID.

### `gyrus update <qualified-name>`

Optional patch flags are `--title`, `--status`, `--tags`, `--content`, `--content-file`, and `--expected-version` (integer, default 0). Only supplied fields are patched. A content file takes precedence over inline content. Positive expected versions must match; zero supplies no caller version precondition. Updates still apply provider conflict checks and lifecycle/body-immutability guards.

```bash
gyrus update root/projects/payments/guide-deploy \
  --title "Reviewed deployment guide" --expected-version 2 --json
```

Returns the same mutation-reference shape as create. No CLI `--reason`, `--dependencies`, or arbitrary metadata patch flag is implemented.

### `gyrus archive <qualified-name>`

Invoke the provider's archive operation and remove its index entry. Current behavior differs by store: localfs removes the concept file; Git and PostgreSQL retain the concept with status `archived`. It does not guarantee cascading relationship cleanup. To request a lifecycle transition while retaining a local file, use `update --status archived` where that type's transition rules permit it.

Plain output confirms archiving; JSON is `{ "archived": true, "id": "<qualified-name>" }`. There is no separate public `delete` command.

## Relationship commands

```text
gyrus link <from-name> <to-name> [--rel-type depends_on]
gyrus unlink <from-name> <to-name> [--rel-type depends_on]
```

Both positional endpoints are qualified names. Relationship types are `depends_on` (default), `implements`, `supersedes`, and `mitigates`. Plain output confirms the operation; `--json` currently suppresses that confirmation without emitting a JSON result object. Use exit status to check completion.

OKF writes source-owned `{concept, rel}` links and allows a qualified missing target. PostgreSQL graph persistence requires existing endpoints. Incoming neighbors and general traversal are exposed by MCP/engine interfaces; there is no separate CLI neighbors/traverse command.

## Search and suggested context

### `gyrus search`

| Flag | Meaning |
| --- | --- |
| `--query` | Keyword text; empty/omitted for metadata browsing |
| `--category`, `--type`, `--status`, `--tag`, `--tenant` | Exact metadata filters |
| `--scope` | `project`, `reference`, or `all` |
| `--project` | Explicit project filter |
| `--max-results` | Positive result cap; 0/default means ten keyword matches or uncapped metadata browsing |

Local/Git use direct OKF keyword scanning; PostgreSQL supplies native full-text matching. Search results include concept, score, and optional match reason. `--json` emits the result array. Project defaults and scope control which project/reference concepts are considered.

```bash
gyrus search --query "deployment" --project payments --max-results 5 --json
gyrus search --query '' --scope all --json
```

### `gyrus suggest-context`

Required: `--prompt`. Optional: `--scope`, `--project`, `--category`, `--max-concepts` (default 5; nonpositive values also use 5), and `--max-tokens` (registered default 4000).

The engine searches for the prompt, prioritizes project concepts, expands `depends_on` targets, deduplicates by qualified name, and emits project/reference sections up to the concept limit. Plain output is formatted context; JSON is `{ "context": "..." }`.

**Current limitation:** `--max-tokens` is accepted but not applied by the command. Do not rely on it for truncation or model-budget enforcement. MCP exposes a concept-count limit, not a token limit.

## Schema and local validation

| Invocation | Inputs and behavior |
| --- | --- |
| `schema <type>` or `schema get <type>` | Print the active template; JSON wraps `type` and `template` |
| `schema list` | List types and whether templates are persisted or embedded |
| `schema set <type> [file]` | Save an override; aliases `save` and `put` |
| `schema delete <type>` | Remove a persisted override; embedded fallback may remain |
| `validate <file-path>` | Parse and structurally validate a local Markdown/JSON concept without storing it |

Schema set accepts `--content-file`, an optional positional file, or `--content`; a file takes precedence, with `--content-file` preferred over the positional file. Schema override validation checks parseability and matching type where present. It does not replace the compiled concept validator. Retrieval falls back to embedded templates only for an absent override; storage and branch-selection errors propagate.

`validate` does not accept `--all` or establish collection uniqueness/location or target existence. With `--json`, successful validation currently emits no body. Browse the collection through an uncapped empty search to exercise location-aware scan validation.

## Git session status and server commands

### `gyrus git status`

Report `strategy`, `trunk_branch`, `active_branch`, and `selection_required`, as text or JSON. It errors on non-Git profiles. Supply `--git-branch initiative/name` to select an initiative for that invocation. There is no persistent CLI switch command.

`direct` operates on configured trunk. `branching` requires a valid initiative distinct from trunk; the branch name is never saved. A missing initiative reads trunk until its first write pushes the branch. Trunk must exist for branching. Separate CLI calls repeat the runtime selection. See the [Git guide](guide-006-git-initiatives.md).

### `gyrus mcp serve`

Run the local stdio MCP server. Flags: `-p/--project <directory>`, `-c/--config <file>`, plus global `--git-branch`. Explicit config takes precedence over the project-directory option. Each process owns its Git selection. A branching server may start without an initiative, then use MCP status/switch before knowledge operations. See the [MCP contract](spec-003-mcp-interface.md).

### `gyrus ui serve`

Run the existing local web UI. Flags: `--host` (default `127.0.0.1`), `-p/--port` (8080), `-P/--project <directory>`, and `-c/--config <file>`. The server exposes concept browsing, search, graph, and configured agent/chat interfaces. This is the local UI; the hosted platform in the architecture roadmap is separate work.

## Output, errors, and authentication

Successful commands use the output forms above; `--json` does not imply that every command returns an object. Local MCP uses protocol stdout, not human command output. The CLI router reports initialization/command errors on stderr unless command JSON mode suppresses ordinary error printing; callers must inspect exit status.

| Exit | Meaning in the current mapper |
| --- | --- |
| 0 | Success |
| 1 | Typed OKF validation errors or errors identified as validation failures |
| 2 | Lifecycle-transition or immutability errors |
| 3 | Permission-denied/unauthorized errors recognized by the mapper |
| 4 | Typed conflicts, expected-version mismatches, or recognized concurrency errors |
| 5 | Other command/storage failures |

Not every configuration/argument/transport error has its own typed category. Application initialization failures occur before command execution and exit 1. The mapper does not implement an ownership/RBAC authorization service.

Git HTTPS uses noninteractive configured credential helpers; SSH uses agent/key authentication and known hosts. PostgreSQL uses its connection string. There is no separate device-login flow, identity-token cache, or offline-read cache. Retired `init`, client-install, sync, cloud-provider, and hybrid-stack interfaces remain unsupported; native client plugin managers own installation.
