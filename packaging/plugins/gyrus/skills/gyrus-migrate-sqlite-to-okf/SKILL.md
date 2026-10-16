---
name: gyrus-migrate-sqlite-to-okf
description: Perform agent-led conversion of legacy Gyrus relationships and SQLite graph edges into qualified, minimal OKF links. Works with installed Gyrus and file/database tools, without Go or repository source. Excludes PostgreSQL exports and storage-layout conversions.
---

# Migrate relationships to qualified OKF links

Read [the migration procedure](references/sqlite-to-okf.md) before modifying concepts or configuration. Perform inventory, staged edits, validation, and cutover with available file tools, installed Gyrus commands, and a SQLite-capable tool when a database graph exists. There are no bundled Go helpers, source imports, or required language runtimes. Do not substitute shell regex rewriting for structured YAML conversion.

Preserve exact document body bytes and lifecycle statuses. Consolidate dependencies, legacy typed links, and database-only edges into source-owned `{concept, rel}` links. Qualified names are `tenant/reference/id` or `tenant/projects/project/id`; retain local frontmatter id. There is no migration or sync command/tool, and SQLite is no longer a runtime index/graph/search provider. Inspect/export the legacy database with a SQLite-capable tool; do not initialize it through Gyrus.

Preflight the complete collection and graph before applying changes. Use explicit mappings for unresolved/ambiguous identities, never guessed tenant/project scope. Build and validate the converted collection in a separate staging directory. Back up originals and a consistent SQLite snapshot before cutover. Verify normalized edge-set parity, unchanged bodies/statuses, and an unchanged second conversion.

This procedure applies to localfs concept collections and SQLite graphs, including collections already using OKF. Git/blob mutation requires a separate provider-aware workflow preserving remote conflict/conditional-write protections; do not treat a remote collection as a local directory.

On failure retain originals, staging, mapping decisions, backups, and an applied-file journal. Report where migration stopped. Do not discard unresolved edges, overwrite concurrent edits, delete old databases/backups, or restore snapshots over subsequent changes. Restart long-running MCP processes after verified cutover.
