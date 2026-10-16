# Agent-led qualified-link and SQLite migration

## Tools and boundaries

Use installed Gyrus for validating the converted representation and verifying cutover. Use available file tools for inventory, backups, byte comparisons, and structured frontmatter edits. No Go checkout/compiler, Perl, Python, Node, or bundled program is required. Shell examples are conveniences; native equivalents are valid.

For a SQLite graph, use an available SQLite CLI, database connector, or GUI with consistent backup and read-only query/export capabilities. Detect it rather than assuming it exists. If none is available, report SQLite export as blocked and leave live data/configuration unchanged; explain that a SQLite-capable tool is required. Do not assume graph edges can be reconstructed from Markdown. A shell by itself cannot safely parse arbitrary YAML or SQLite. Do not install tooling without applicable authorization.

A SQLite graph plus file-backed concept storage is supported here. If concept bodies exist only inside SQLite, or the storage is Git/blob/PostgreSQL, stop and report that a separate storage export/provider-aware procedure is needed. This procedure does not move paths or repair metadata/location mismatches.

## 1. Inventory and preflight without source writes

1. Locate actual config, storage root, concepts directory, and graph database. Use `gyrus config show --json` only if the current runtime can initialize; otherwise inspect the config directly. An old identity database must not be opened by the new Gyrus runtime. Account for environment overrides, config-relative paths, and separate graph/index DSNs.
2. Stop competing CLI/MCP/UI writers for the migration window. Establish exclusive access; if writers cannot be quiesced, do not apply. Resolve actual paths and keep staging/backups outside the collection. Reject symlinks or unexpected canonical paths rather than following them.
3. Inventory every canonical concept file directly, excluding schemas/backups. Do not rely on search limits. Record source bytes/checksum, exact body bytes after the closing frontmatter delimiter, metadata, status, revision, and canonical name derived from its relative path. Filename/local id, tenant, and path must agree. Allow repeated local IDs across scopes; reject duplicate names or malformed identity.
4. Read frontmatter structurally, preserving unrelated metadata, multiline values, body bytes, and newline style. Do not use line-based substitutions to parse YAML collections. If available file tools cannot preserve/validate this representation reliably, stop before writes and identify the missing capability.
5. If SQLite is present, inspect its actual table/column schema and export the full graph from a read-only connection. For the known Gyrus schema, the relevant columns are below. Do not execute unverified queries against a different schema, discard unknown tables, or interpret a failed export as an empty graph.

With `sqlite3`, these are read-only inspection examples (use your resolved path):

```sh
sqlite3 -readonly "$database" '.schema concept_edges'
sqlite3 -readonly "$database" 'PRAGMA table_info(concept_edges);'
sqlite3 -readonly -json "$database" 'SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges ORDER BY from_concept_id, to_concept_id, relationship_type;'
```

Use equivalent structured export if `-json` is unavailable. Retain raw export evidence, including legacy audit columns if needed for diagnosis; discard audit columns from the converted relationship contract.

## 2. Resolve identities and plan the complete edge set

Build an inventory mapping each local ID to all existing qualified names. Resolve names for document links and both SQLite edge endpoints by these rules:

- A valid qualified name remains exact and case-sensitive. Missing qualified targets are permitted; a missing source is an error.
- An old short ID resolves automatically only when exactly one existing concept has that local ID.
- Ambiguous or missing short IDs require an explicit user-provided mapping to a valid qualified name. Record mapping decisions in the migration report. Use source/edge-specific mappings where the same legacy ID refers to different scopes; never guess from the source tenant/project.
- Reject malformed/unsafe names, conflicting old/new link keys, empty fields, unsupported relations, unknown link properties, malformed dependencies, and database edges without a source concept. Audit keys created_by/created_at are recognized legacy fields to remove.

For each source, merge these tuples into a deduplicated set `(qualified source, qualified target, relation)`:

| Input | Converted link |
| --- | --- |
| `dependencies: [<id>]` | `{concept: <resolved-name>, rel: depends_on}` |
| `links: [{target_id: <id>, rel_type: <rel>, ...audit}]` | `{concept: <resolved-name>, rel: <rel>}` |
| `links: [{concept: <name>, rel: <rel>}]` | Same minimal pair after validation |
| SQLite graph row | Same pair stored on its resolved source document |

Accept only depends_on, implements, supersedes, mitigates. Retain self-links, cycles, and qualified missing targets; store no inverse edges. Reconcile the entire inventory before any live writes. Present planned changed sources, normalized edge set, mappings, and unresolved items. Unresolved items block apply, not unrelated inspection.

## 3. Back up and stage

Create a new private backup directory outside `concepts/`; copy the entire concepts collection and exact config. For SQLite use a database tool's consistent backup API, including committed WAL data. With sqlite3, open the source read-only and use `.backup` to a fresh destination through the tool's correctly quoted argument mechanism. Do not copy just index.db while WAL state may exist. Run `PRAGMA integrity_check;` against the snapshot and require `ok`.

Record source checksums and the normalized live database edge set, then work from the snapshot. Retain old database and sidecars untouched. Create a separate staging storage root containing copies of the collection and any schema overrides. Never put backups or staging beneath live concepts where Gyrus would index them.

Apply file edits only in staging:

- Merge the complete planned links, remove dependencies and recognized legacy link/audit keys, and persist exactly concept/rel per link. Sort links by concept then rel for stable output.
- Preserve unrelated metadata and all body bytes/statuses, including accepted historical ADR bodies. Do not replay through create/update/link commands: that can alter lifecycle behavior and increment versions per edge.
- Increment document version exactly once for each document whose relationship representation changes, even if it receives multiple edges. Set last_updated to the conversion time in UTC; preserve last_modified_by unless an explicit migration actor is supplied. Leave unchanged documents byte-identical. Do not increment revisions solely to reorder existing valid links.
- Use canonical metadata order for rewritten headers: title, id, type, format, category, version, status, last_updated, last_modified_by, tenant, tags, links; unrelated keys follow. Do not persist derived name/scope/project/filepath.

Give staging its own config using localfs and OKF for index, graph, and search. Make the storage root explicitly point to staging; remove database DSNs there. Isolate environment/home when invoking Gyrus so the live config cannot be selected accidentally. Never initialize with `--force` over the live config.

## 4. Verify staging and idempotency

Before cutover require all of the following:

1. `gyrus validate <staged-file> --json` succeeds for every concept; `gyrus sync --json` succeeds against staging. Inspect resolved config to prove staging-only storage.
2. Every staged body matches its original bytes, every status is unchanged, and all unrelated metadata is preserved. Only planned sources changed, with the expected single revision increment.
3. Read staged links back structurally. Their complete normalized tuple set must equal the planned union of document and SQLite edges, with no extra/missing tuples. Compare sets, not counts. Names must match paths; links contain exactly concept/rel and no audit properties.
4. Repeat the conversion planning against the staged collection and same SQLite snapshot. Require zero planned changes, identical file bytes, and no new revision increments. Do not rewrite just to change timestamps or link ordering.
5. Exercise qualified get, keyword search, complete metadata browsing (`gyrus search --query '' --scope all --max-results 0 --json`), and context suggestions. When an isolated MCP client is available, verify incoming/outgoing neighbors and traversal against staged links. Report unperformed interface checks explicitly.
6. Verify the backup is readable and recovery works in a disposable copy; never test restore over live files.

## 5. Cut over and recover

Recheck the entire live source/config checksum inventory and live SQLite edge set against preflight. Any change blocks cutover and requires reconciliation. Exclusive writer access must continue through installation and verification; per-file checksum checks alone do not close races.

Install only planned staged files using atomic per-file replacement where supported, preserving permissions. Record each installed file and original checksum in an apply journal; a multi-file install is not a single transaction. Once all concepts are installed, update all three live provider fields together:

```yaml
index:
  provider: okf
graph:
  provider: okf
search:
  provider: okf
```

Retain storage/tenant/project settings and unrelated config. Remove only now-unused index/graph/search database settings. Do not automatically change process environment overrides; report any that would override the cutover. Restart owned or authorized long-running MCP/UI services, then verify resolved providers, sync, complete inventory, graph tuple parity, keyword search, and qualified reads on live data. Retain originals, backup, snapshot, mappings, evidence, and journal after success.

If applying fails, stop and report installed files plus backup/staging paths. With writers still quiesced and no later edits, restore the consistent backed-up concepts/config set and verify it before resuming services. Keep the original database available for the old configuration; the new runtime cannot open an obsolete identity database, so original behavior may require the previous installed release. Never upgrade/downgrade binaries or overwrite later edits implicitly. If later edits exist, reconcile them before recovery.

Final report: source/staged/applied concept counts, changed revisions, normalized edge-set parity, body/status preservation, idempotency, interface checks performed/blocked, database integrity result, backup path, and cutover/recovery outcome.
