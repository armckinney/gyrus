---
title: Gyrus Model Context Protocol Reference Manual
id: tech-ref-003-mcp-manual
type: technical-reference
format: markdown
category: technical
version: 5
status: active
last_updated: 2026-10-16T19:41:57Z
last_modified_by: ""
tenant: root
---

# Gyrus MCP reference

Run `gyrus mcp serve` over stdio using the installed Gyrus binary. The canonical plugin configures this command and passes project discovery settings. Restart the MCP process after provider configuration changes; a running application caches its engine.

## Tools

| Tool | Inputs |
| --- | --- |
| gyrus_create | Required id/title/category/type; optional content/status/tenant/project/scope |
| gyrus_get | id |
| gyrus_update | id; optional title/status/content/tags/expected_version |
| gyrus_archive | id |
| gyrus_search | Optional query/category/type/status/tag/tenant/project/scope/max_results |
| gyrus_suggest_context | prompt; optional max_concepts/project/scope |
| gyrus_link / gyrus_unlink | from_id, to_id; optional rel_type (defaults depends_on) |
| gyrus_neighbors | id; optional direction/rel_type |
| gyrus_traverse | start_id; optional max_depth (tool default 2) |
| gyrus_schema_get / gyrus_schema_delete | type |
| gyrus_schema_list | No arguments |
| gyrus_schema_set | type, content |

Valid relationships are depends_on, supersedes, implements, and mitigates. Neighbor directions are incoming, outgoing, and both. Scope values are project, reference, and all.

## OKF retrieval and graph behavior

File-backed profiles default to OKF index/graph/search. Keyword queries use case-insensitive Unicode substring terms with OR semantics and return ten matches by default. Positive max_results values pass to the selected provider and override the default. Empty-query metadata browsing returns all matching concepts unless capped. PostgreSQL applies scope/project filters with native full-text matching; vector uses OKF index/graph and lexical fallback. SQLite providers have been removed.

Link/unlink persist typed source frontmatter links and use qualified concept names with concept and rel only. Targets may be unresolved. Metadata-only links remain editable on immutable concepts without changing the body. Mutations reject version/Git/object-precondition conflicts, and multi-source batches report any already-applied sources.

The sync tool/command has been removed. Invalid concepts and duplicate qualified names fail queries with tool errors rather than hiding partial results. Validate individual Markdown files through the CLI and verify complete collection browsing through empty-query search.

Use the packaged gyrus-migrate-sqlite-to-okf skill to back up, export, and verify legacy database-only edges before choosing OKF providers. Change all three provider fields and restart the server afterward; the runtime no longer opens SQLite provider configurations.

## Resources and prompts

Concept resources use `memory://concept/{id}` with a qualified name and return Markdown. Schema resources expose active persisted templates or embedded fallbacks. The prepare-adr prompt uses that same active ADR template and serializes the supplied title safely in canonical metadata order, without obsolete dependencies. Use tool/resource discovery for the precise installed schema.

## Qualified-name contract

The id arguments of gyrus_get, gyrus_update, and gyrus_archive, from_id/to_id of gyrus_link and gyrus_unlink, id of gyrus_neighbors, and start_id of gyrus_traverse require complete names such as root/reference/adr-example or root/projects/gyrus/ticket-003. Argument names are retained; short values are rejected. gyrus_create still takes a local id with tenant/scope/project and returns name plus local id. Search concept envelopes include name. Markdown resources keep a local frontmatter id and omit derived name; use memory://concept/root/reference/adr-example to address one. rel_type remains a tool argument; persisted links contain only concept and rel. Incoming queries derive backlinks without writing inverse records.
