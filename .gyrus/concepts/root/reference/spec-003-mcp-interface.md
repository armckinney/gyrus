---
title: Embedded Model Context Protocol Server Specification
id: spec-003-mcp-interface
type: specification
format: ""
category: technical
version: 4
status: active
last_updated: 2026-11-12T05:49:34Z
last_modified_by: ""
tenant: root
---

# Model Context Protocol interface specification

Gyrus exposes the lifecycle engine through a local stdio MCP server. Tools, resources, and prompts share the configured store, identity rules, schema overrides, and active Git session. This specification records the implemented interface so an integrator can build a client without relying on historical tool names or proposed endpoints.

## Startup, transport, and discovery

```bash
gyrus mcp serve
gyrus mcp serve --project /path/to/code-repo
gyrus mcp serve --config /path/to/.gyrus.yaml --git-branch initiative/payments
```

`--config` selects an explicit configuration file; otherwise `--project` selects a project directory, followed by normal discovery. The plugin starts `gyrus mcp serve` and supplies `GYRUS_PROJECT` from the client workspace. MCP traffic is JSON-RPC over stdin/stdout; clients initialize and then discover the installed tool/resource/prompt schemas. Operational diagnostics must not be mixed into protocol stdout.

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"example-client","version":"1.0"}}}
```

After a successful handshake, send `notifications/initialized` and use `tools/list`, `resources/list`, `resources/templates/list`, or `prompts/list`. The example protocol version is used by repository tests; clients should negotiate a mutually supported version. This implementation serves local stdio, not a hosted HTTP/SSE MCP endpoint.

## Common parameter and identity rules

- Existing-concept `id`, `from_id`, `to_id`, and `start_id` values are full names such as `root/reference/adr-storage` or `root/projects/payments/guide-deploy`.
- `gyrus_create.id` is a local slug, combined with tenant/scope/project to derive its name. MCP creation defaults tenant to `root`; project context can come from the configured engine. Scope is `project` or `reference` for creation, and additionally `all` for retrieval.
- Categories are `architecture`, `business-logic`, `product`, `operations`, and `technical`. Supported concept types are defined by the [concept contract](spec-001-contract-schema.md).
- A positive `expected_version` enables optimistic version checking. Omitting it, or using zero, supplies no caller version precondition. Writes still use the selected provider's conflict protections.
- Omitted optional update strings are not patched. The current handler also treats empty strings as absent, so it does not clear string fields through an empty-string argument.

## Document tools

| Tool | Required parameters | Optional parameters and behavior | Success payload |
| --- | --- | --- | --- |
| `gyrus_create` | `id`, `title`, `category`, `type` (strings) | `tenant`, `status`, `scope`, `project`, `content` (strings); status defaults to `proposed` for ADRs and `draft` otherwise | JSON text containing a `ConceptRef` |
| `gyrus_get` | `id` (qualified-name string) | None | Markdown with frontmatter and body |
| `gyrus_update` | `id` (qualified-name string) | `title`, `status`, `content`, `tags` (comma-separated string), `expected_version` (number) | JSON text containing a `ConceptRef` |
| `gyrus_archive` | `id` (qualified-name string) | None | Text confirmation |

A `ConceptRef` contains `name`, local `id`, `version`, `status`, and `last_updated`. Concept retrieval preserves the local frontmatter ID; derived identity is represented by the requested name and storage location. Archive invokes the selected store's archive operation: localfs removes the file, while Git and PostgreSQL retain the concept with status `archived`. It does not promise cascading relationship cleanup. Lifecycle and body-immutability failures are tool errors.

Example call:

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gyrus_update","arguments":{"id":"root/projects/payments/guide-deploy","title":"Reviewed deployment guide","expected_version":2}}}
```

## Search and context tools

| Tool | Required parameters | Optional parameters | Success payload |
| --- | --- | --- | --- |
| `gyrus_search` | None | `query`, `category`, `type`, `status`, `tag`, `tenant`, `scope`, `project` (strings); `max_results` (number) | JSON text array of search results |
| `gyrus_suggest_context` | `prompt` (string) | `scope`, `project` (strings); `max_concepts` (number, default 5 for omitted/nonpositive values) | Formatted context text |

Search results contain `concept`, numeric `score`, and optional `match_reason`. Positive `max_results` caps results; nonpositive/omitted values select ten for keywords and unlimited metadata-only browsing. Local/Git use case-insensitive OKF keyword scanning; PostgreSQL uses native English FTS. Filters and qualified project/reference context apply through the selected provider.

Context suggestion searches for the prompt, prioritizes project concepts, expands `depends_on` targets through the graph, deduplicates by qualified name, and fills the remaining slots with shared references. The final concept count is capped by `max_concepts`. It does not expose a token-budget argument or automatically decide which engineering statement is authoritative.

## Graph tools

| Tool | Required parameters | Optional parameters | Success payload |
| --- | --- | --- | --- |
| `gyrus_link` / `gyrus_unlink` | `from_id`, `to_id` (qualified-name strings) | `rel_type` (string, default `depends_on`) | Text confirmation |
| `gyrus_neighbors` | `id` (qualified-name string) | `direction` (`incoming`, `outgoing`, `both`; omitted defaults to both); `rel_type` | JSON text array of edges |
| `gyrus_traverse` | `start_id` (qualified-name string) | `max_depth` (number; default 2 for omitted/nonpositive values) | JSON text array of graph paths |

Relations are `depends_on`, `implements`, `supersedes`, and `mitigates`. Edge payloads use `from_concept_id`, `to_concept_id`, and `relationship_type`; a graph path contains `nodes` and `edges`. Traversal is cycle-safe; these relationships do not impose a DAG.

OKF link mutations persist minimal `{concept, rel}` entries in the source concept. Qualified missing targets are permitted, while a missing source is an error. PostgreSQL graph edges use database persistence and require existing endpoints. Incoming neighbors derive backlinks without writing inverse assertions. Multi-source batches have per-source consistency and report partial application when later sources fail; they are not cross-concept transactions.

## Schema tools

| Tool | Required parameters | Success payload |
| --- | --- | --- |
| `gyrus_schema_get` | `type` (string) | Active Markdown schema template |
| `gyrus_schema_list` | None | JSON text array of `{type, source, location?}` |
| `gyrus_schema_set` | `type`, `content` (strings) | JSON text with `status`, `type`, `location` |
| `gyrus_schema_delete` | `type` (string) | JSON text confirming deletion |

Persisted overrides take precedence over embedded templates. Only an absent override permits fallback; storage, cancellation, and required-selection failures propagate. Deleting an override can expose the embedded template again. Schema content must be parseable and match the supplied type where present; custom headings do not add new runtime concept types or replace structural validation.

## Git session tools

| Tool | Required parameters | Optional parameters | Success payload |
| --- | --- | --- | --- |
| `gyrus_git_status` | None | None | JSON text status |
| `gyrus_git_switch` | `strategy` (`direct` or `branching`) | `branch` (required for branching, omitted for direct) | JSON text status after successful switch |

Status contains `strategy`, `trunk_branch`, `active_branch`, and `selection_required`. Both tools return errors on non-Git stores. `direct` reads/writes the configured trunk; `branching` uses a valid runtime initiative branch distinct from trunk. No branch is saved to configuration.

Each stdio server owns its selection. `--git-branch` selects branching at startup. A branching server can start without a selection, expose discovery/status/switch, and reject knowledge operations until selected. Switches wait for in-flight tool/resource/prompt requests. Candidate preparation must succeed before activation; failure retains the old branch. A missing initiative reads trunk until its first mutation creates the remote branch. Existing branches are reused without automatic trunk integration.

Example switch arguments:

```json
{"strategy":"branching","branch":"initiative/payments"}
```

Return to trunk with `{"strategy":"direct"}`. Restarting requires runtime selection again. Humans create, review, and merge PRs through their Git host; the provider does not perform those actions.

## Resources and prompts

| Resource URI | MIME type | Behavior |
| --- | --- | --- |
| `memory://concept/{id}` | `text/markdown` | Read a complete concept using its qualified name |
| `memory://schema/{concept_type}` | `text/markdown` | Read the active persisted or embedded template |
| `memory://schemas` | `application/json` | List schema types and their source/location |

For example, `memory://concept/root/reference/adr-storage` addresses a reference concept. Resource failures are returned as MCP resource errors rather than successful partial content.

`prepare-adr` is the implemented prompt. Its required `title` argument is safely serialized into frontmatter and substituted into the active ADR template. It uses the same schema resolution as tools/resources. No additional implementation-review or memory-update prompt is currently registered.

## Results, errors, and compatibility

Tool results use MCP `content` items of type `text`. Structured payloads described above are JSON encoded inside that text; `gyrus_get`, schema retrieval, suggestions, and mutation confirmations return their specified textual forms. Do not assume all successful tools return the same JSON object. A tool operation failure returns `isError: true`; protocol-level failures remain JSON-RPC errors. CLI numeric exit codes are not MCP result codes.

Stale versions, rejected competing Git pushes, invalid lifecycle transitions, missing concepts, malformed collections, schema failures, and invalid selections must remain visible. Do not substitute embedded schemas for remote failures or silently return partial invalid collections. Authentication uses the configured provider: Git credential helpers/SSH or the PostgreSQL connection string.

The public tools use `gyrus_*` names. Historical `memory.*` tools, sync/validation tools, extra prompts, and hosted transport sketches do not define this interface. See [MCP consumer reference](tech-ref-003-mcp-manual.md), [Git behavior](spec-201-git-storage-driver.md), and [contract schema](spec-001-contract-schema.md).
