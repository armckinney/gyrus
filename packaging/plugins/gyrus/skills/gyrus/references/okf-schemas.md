# Gyrus concept contract — quick reference

Use `gyrus schema get <type>` for a template. Code defines validation; custom templates do not change it. Required metadata and typed links are Gyrus extensions to upstream OKF.

## Frontmatter

```yaml
---
title: Storage decision
id: adr-storage
type: adr
format: markdown
category: architecture
version: 1
status: proposed
last_updated: 2026-10-16T00:00:00Z
last_modified_by: developer
tenant: root
tags: [storage]
links:
  - concept: root/reference/spec-004-core-engine
    rel: depends_on
---

# Storage decision
```

| Rule | Contract |
| --- | --- |
| Required | Nonempty `title`, `id`, `type`, `category`, `status`, `tenant`; integer `version >= 1`. |
| Local ID | `^[a-z0-9-_]+$`; equals filename without extension; safe path segment. |
| Optional | `format`, `last_updated`, `last_modified_by`, `tags`, `links`, `immutable`. |
| Write order | As shown above; additional fields (e.g. `immutable`) follow `links`. |
| Categories | `architecture`, `business-logic`, `product`, `operations`, `technical`. |
| Types | `adr`, `prd`, `guide`, `improvement-proposal`, `release-note`, `specification`, `standards`, `technical-reference`, `product`, `glossary`, `freeform`. |

## Identity and links

- **Name:** `tenant/reference/id` or `tenant/projects/project/id`. Case-sensitive, no leading slash/extension. Local IDs may repeat; names cannot.
- **Location:** `concepts/<name>.md`; tenant/filename match metadata. Omit derived `name`, `scope`, `project`, `filepath` from YAML.
- **Create:** local `id` + tenant/scope/project; returns local `id` and qualified `name`.
- **Existing-concept operations:** qualified names only. Moves do not rewrite links.
- **Link:** exactly `{concept: <qualified-name>, rel: <relation>}`. Relations: `depends_on`, `implements`, `supersedes`, `mitigates`. Source is implicit; incoming queries derive backlinks.
- **OKF graph:** cycles/self-links/missing targets allowed; duplicate tuples normalize to one edge.
- **Rejected:** `dependencies`, old link keys `target_id`/`rel_type`, per-link creation metadata, other unknown link keys. CLI `--rel-type` and MCP `rel_type` arguments remain valid.

## Lifecycle

| Type | Allowed transitions |
| --- | --- |
| ADR | proposed → accepted/rejected; accepted → superseded/deprecated |
| Improvement proposal | draft → reviewing/abandoned; reviewing → approved/rejected/abandoned; approved → implemented/abandoned |
| Others | draft → active/archived; active → deprecated/archived; deprecated → archived |

Same-state updates are allowed. Preserve finalized ADR/proposal/release-note bodies. `immutable: true` locks bodies outside draft/proposed; metadata stays editable. Initial status validation checks nonemptiness; updates check transitions. Use link/unlink for relationships and expected versions for updates.

## Persistence and legacy data

Normal typed serialization drops unknown top-level fields/comments and trims surrounding body whitespace.

Use `gyrus-migrate-sqlite-to-okf` for legacy data and database-only edges. SQLite index/graph/search providers and the sync command/tool have been removed. OKF reads the collection directly; validate individual files with `gyrus validate <filepath> --json`.
