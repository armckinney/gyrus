---
title: Open Contract Schema & Frontmatter Validation Rules
id: spec-001-contract-schema
type: specification
format: markdown
category: technical
version: 5
status: active
last_updated: 2026-10-16T18:32:57Z
last_modified_by: ""
tenant: root
---

# Gyrus concept schema contract

This specification describes the implemented Gyrus contract for Markdown concepts with YAML frontmatter and supported JSON concept envelopes. Typed relationship metadata, required fields, lifecycle rules, and tenant/project topology are Gyrus extensions to [upstream Open Knowledge Format](https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md). Upstream OKF expresses concept relationships through ordinary Markdown links; it does not mandate the Gyrus-specific schema below.

## Metadata fields and serialization order

Gyrus YAML writers emit the following fields in this order when present. Readers accept other key orders; order is a presentation rule, not a validation requirement. Optional fields are not all emitted when empty.

| Field | Type | Implemented contract |
| --- | --- | --- |
| `title` | string | Required and nonempty. |
| `id` | string | Required local slug matching `^[a-z0-9-_]+$`, also subject to safe-path validation. It contains no tenant, scope, slash, or extension. |
| `type` | string | Required; one of the supported concept types below. |
| `format` | string | Optional. Markdown parsing defaults an absent/empty value to `markdown`; JSON parsing defaults it to `json`. The structural validator does not restrict explicit values to an enum. |
| `category` | string | Required; one of the supported categories below. |
| `version` | integer | Required and at least one in a validated concept. Normal file-backed creation and patching manage revisions. |
| `status` | string | Required and nonempty. Structural validation does not enforce an initial-status enum; lifecycle updates have separate transition rules. |
| `last_updated` | timestamp | Optional typed `time.Time` metadata. JSON uses RFC 3339-compatible timestamps; YAML uses the timestamp forms accepted by the YAML decoder. File-backed writes assign UTC update timestamps. |
| `last_modified_by` | string | Optional attribution metadata. It is distinct from concept identity and relationship semantics. |
| `tenant` | string | Required, nonempty, safe path segment; must match the tenant in a stored concept's canonical location. |
| `tags` | list of strings | Optional; empty collections may be omitted by normal serialization. |
| `links` | list of mappings | Optional; each entry has exactly `concept` and `rel`, as defined below. Empty collections may be omitted. |

Additional fields follow these ordered fields. The supported optional `immutable` boolean is emitted when true and follows `links` when both are present. The ordering utility retains additional keys and comments when reordering existing YAML. Normal concept deserialization/serialization uses the typed `Concept` struct: unrecognized top-level metadata is ignored on read and is not retained on a subsequent normal write. YAML comments are not retained by that typed round trip. This differs from the stricter rejection of unknown keys inside individual links.

Categories are `architecture`, `business-logic`, `product`, `operations`, and `technical`.

Concept types are `adr`, `prd`, `guide`, `improvement-proposal`, `release-note`, `specification`, `standards`, `technical-reference`, `product`, `glossary`, and `freeform`.

A complete Markdown example follows. Optional fields are shown to demonstrate the canonical order; they are not all required.

```markdown
---
title: Example decision
id: adr-example
type: adr
format: markdown
category: architecture
version: 1
status: proposed
last_updated: 2026-10-16T00:00:00Z
last_modified_by: developer
tenant: root
tags: [architecture]
links:
  - concept: root/reference/spec-004-core-engine
    rel: depends_on
immutable: true
---

# Example decision

Decision content belongs in the Markdown body.
```

## Local IDs, qualified names, and storage locations

The local `id` remains in frontmatter. The canonical `name` is a case-sensitive, collection-relative identifier with exactly one of these forms:

| Logical location | Qualified name | Canonical Markdown path relative to storage root |
| --- | --- | --- |
| Shared reference | `root/reference/adr-example` | `concepts/root/reference/adr-example.md` |
| Project concept | `root/projects/gyrus/ticket-003` | `concepts/root/projects/gyrus/ticket-003.md` |

Names exclude a leading slash, the storage `concepts/` prefix, and the file extension. Git locates its collection under `.gyrus/concepts/` and also supports JSON concept files. Blob keys may add a configured prefix before `concepts/`. These physical prefixes are not part of the qualified name.

Tenant/project components reject empty values, `.`/`..`, directory separators, control characters, unsafe filename characters, trailing dots/spaces, and reserved device names. Local IDs additionally obey the slug regex. Names are validated rather than silently cleaned or normalized.

Location-aware file-backed reads check that the filename matches the local ID and the path tenant matches metadata, then derive `name`, `scope`, `project`, and `filepath`. YAML serialization excludes all four derived/location fields. JSON API envelopes can expose them; persisted database JSON metadata omits the derived `name` while retaining the logical scope/project needed to reconstruct it.

Creation accepts a local ID plus tenant/scope/project and returns a `ConceptRef` containing both local `id` and qualified `name`, with version, status, and last-updated metadata. The application engine applies its configured project when no explicit reference scope/project is supplied. The shared path builder also retains the existing `gyrus` project fallback for product/business-logic concepts when no project is supplied and reference scope is not explicit. An explicit nonempty project determines project placement.

Operations addressing an existing concept require its qualified name: SDK lookups/mutations, CLI get/update/archive/link/unlink arguments, MCP `id`/`from_id`/`to_id`/`start_id` arguments, graph endpoints, and UI concept routes. Argument names are retained; short-ID values are rejected. Creation's `id` argument remains a local slug. Search envelopes expose local `id` and derived `name`; Markdown payloads retain the local frontmatter ID. Local IDs may repeat across tenants/projects; canonical names must be unique within the collection. Moving a concept changes its name without automatically rewriting existing links. Cross-collection identifiers are outside this contract.

## Directed relationships

A persisted link has exactly this shape:

```yaml
links:
  - concept: root/reference/spec-004-core-engine
    rel: depends_on
```

Both fields must be nonempty strings. `concept` must be a qualified name. `rel` must be one of the following directed assertions:

| Stored relation from A to B | Meaning |
| --- | --- |
| `depends_on` | A depends on B. Context suggestions use this relation to expand project context. |
| `implements` | A realizes or implements the requirement, proposal, or specification represented by B. |
| `supersedes` | A replaces B. |
| `mitigates` | A addresses or reduces the concern represented by B. |

Source identity comes from the owning concept's qualified name. Link entries do not repeat the source and carry no creation attribution or timestamp. The top-level `dependencies` field is rejected even when empty/null. Old `target_id`/`rel_type` keys, per-link `created_by`/`created_at`, and any other unknown link keys are rejected by Markdown/JSON format readers. This does not remove the CLI `--rel-type` flag or MCP `rel_type` argument: those interface names still select the relationship written as `rel`.

Graph normalization deduplicates source/target/relation tuples; it does not require the stored list itself to have unique entries. Different relationship types between the same endpoints remain distinct. Incoming queries expose backlinks to the original directed assertion without storing inverse copies. There are no parent/child types or DAG constraints. General cycles and self-links are permitted; traversal is cycle-safe.

### Provider-specific persistence

With the OKF graph provider, reads derive edges from source-owned links in validated concept snapshots. A write requires an existing source, permits a qualified missing target, and patches only the source's link metadata using its expected version. Duplicate link and absent-unlink operations are no-ops and do not increment the version. Multi-source batches preflight all sources, then apply updates in qualified-name order; later failures report the sources already updated. They are not a cross-concept transaction.

Explicit database graph providers persist edges in their database tables rather than patching source frontmatter. Their existence constraints differ: PostgreSQL foreign keys require both endpoint concepts, while SQLite graph tables do not impose those foreign keys. Source-frontmatter ownership, per-source revision updates, and no-op revision behavior above describe the OKF provider, not every database adapter. Public edge responses retain the field names `from_concept_id`, `to_concept_id`, and `relationship_type`, with qualified endpoint values and no creation metadata.

## Validation boundaries, revisions, and lifecycle

Markdown/JSON format readers parse metadata and reject obsolete or malformed relationship shapes. Parsing alone is not full structural validation. `okf.Validate` checks required values, supported category/type, version, safe identifiers, qualified link targets, and relationship vocabulary. The standalone `gyrus validate <file>` validates its payload; it does not establish that the file belongs at a canonical collection path, that an ID is unique in a collection, or that a link target exists. Location-aware storage reads and collection scans perform those additional identity/location checks where applicable.

Scanner-backed OKF index/search/graph queries fail if the authoritative collection contains malformed concepts or duplicate qualified names. Scans enumerate canonical concepts and exclude schemas/backups outside that collection. Sync reports per-record errors; the CLI exits with validation code 1, and MCP marks the report as a tool error. A scan report can contain validated-record counts alongside errors; it is not a guarantee of a usable partial query result.

Normal file-backed patches increment the version and update `last_updated`, including accepted metadata-only patches. A positive expected version enables the version check; zero means no caller-supplied version check. File-backed providers retain local interprocess locks, Git conflict detection, or conditional blob mutations. Link mutations made through the OKF graph provider supply the source's observed version. The structural schema does not itself impose these storage concurrency mechanisms.

Lifecycle updates through the engine validate these transitions; same-state updates are accepted:

| Type | Allowed changes |
| --- | --- |
| ADR | `proposed` → `accepted`/`rejected`; `accepted` → `superseded`/`deprecated` |
| Improvement proposal | `draft` → `reviewing`/`abandoned`; `reviewing` → `approved`/`rejected`/`abandoned`; `approved` → `implemented`/`abandoned` |
| Other supported types | `draft` → `active`/`archived`; `active` → `deprecated`/`archived`; `deprecated` → `archived` |

ADR, improvement-proposal, and release-note types protect finalized content. Without explicit `immutable`, locked statuses are `accepted`, `rejected`, `superseded`, `deprecated`, `approved`, `abandoned`, `implemented`, `published`, and `archived`. With `immutable: true`, any status other than `draft` or `proposed` locks the body. The mutation guard applies to changed body content; metadata-only updates remain permitted. Status/type validation and lifecycle transitions are separate from initial structural validation; templates do not add a stricter runtime initial-status enum.

Normal Markdown reading/writing trims surrounding body whitespace, and serialization emits delimiters plus a trailing newline. Metadata-only edits do not replace the logical content field, but normal serialization is not a byte-preserving edit of arbitrary Markdown whitespace. This differs from the dedicated conversion's exact body-byte preservation guarantee.

## Templates and legacy conversion

Persisted schema templates override embedded templates for schema retrieval. Template validation checks parseability and matching type when supplied; template placeholders/custom headings do not replace the code-defined concept validator or add new supported concept types.

`migration.ConvertLinks`, owned by the `gyrus-migrate-okf` skill under `packaging/plugins/gyrus/skills/gyrus-migrate-okf/scripts/migration`, is a source-checkout local conversion path, separate from strict runtime readers. It inventories the entire collection, merges legacy dependencies as `depends_on` with old/new typed links, deduplicates qualified tuples, and removes link creation metadata. It resolves short targets only when unique; ambiguous or absent short targets require explicit `NameMappings`. Already-qualified missing targets remain valid for OKF.

Applying changed records requires a backup directory outside the concept collection and quiescent writers. Conversion backs up concepts and the supplied configuration, preserves exact body bytes and statuses, increments each changed source revision once, and verifies resulting tuple/body/status/revision parity. A repeat conversion plans zero updates. The SQLite export helper separately preserves database-only relationships as the same minimal qualified links and ignores old edge creation metadata. Old database identity schemas require explicit export/rebuild, not automatic runtime reuse; sync alone performs neither conversion nor graph export.

## Implementation references

- [Public concept, patch, reference, and relationship types](../../../../pkg/gyrus/types.go)
- [Structural validation](../../../../internal/domain/okf/validator.go), [qualified names](../../../../internal/domain/okf/name.go), and [collection locations/scans](../../../../internal/domain/okf/collection.go)
- [Frontmatter order](../../../../internal/format/frontmatter.go), [relationship-shape checks](../../../../internal/format/relationships.go), and [Markdown serialization](../../../../internal/format/okf.go)
- [Lifecycle transitions](../../../../internal/domain/lifecycle/engine.go) and [file-backed patch rules](../../../../internal/domain/lifecycle/patch.go)
- [OKF graph persistence](../../../../internal/provider/graph/okf/graph.go) and [legacy conversion](../../../../packaging/plugins/gyrus/skills/gyrus-migrate-okf/scripts/migration/links.go)
