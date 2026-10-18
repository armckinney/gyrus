# Consumer lifecycle procedure

Run these checks through the installed executable. Examples use POSIX shell syntax; use equivalent native terminal/file tools on other platforms. Inspect JSON with your own tool capabilities; no jq or language runtime is required. Record actual returned versions, qualified names, storage root, and file paths rather than reconstructing them for mutations.

## 1. Isolate and initialize

Resolve the executable before changing directory. Make a private temporary directory with separate workspace, home, and evidence directories. Record the caller's directory and environment. Run every command/service in the private workspace with a private HOME (and USERPROFILE where relevant), removing inherited `GYRUS_*` variables first. Set `GYRUS_PROJECT` to the private workspace. After initialization set `GYRUS_CONFIG` to the absolute generated config path. Keep evidence outside the disposable workspace/home.

```sh
gyrus --help
gyrus config init --profile local --tenant lifecycle --project demo --json
gyrus config show --json
gyrus search --query '' --scope all --max-results 0 --json
```

Require localfs storage inside the private workspace, OKF index/graph/search, and zero initial concepts. Require that CLI help and MCP tools expose no sync command/tool. SQLite index/graph/search configuration must fail explicitly before data writes. Save the generated configuration for a final byte comparison. Do not copy customer configuration or data. Inspect `--help` for create/update/search/schema/link/unlink/archive/validate, `mcp serve`, and `ui serve` as needed.

In separate disposable directories, verify an unknown profile fails without creating/replacing configuration. Verify an unknown search provider fails before storage access. Use copies of private fixture configurations for these rejection checks; keep the main test configuration unchanged.

## 2. Schema and interface interoperability

List embedded schemas and save `schema get guide --json`. Set a private override with `schema set guide --content '<valid guide template with a unique marker>' --json` and read it back.

Start `gyrus mcp serve --config <private-config>` using an available MCP client or a terminal with persistent stdin/stdout. Initialize the session, send the initialized notification, and discover `tools/list`. For raw stdio, send newline-delimited JSON-RPC messages with unique request IDs; inspect matching responses and MCP `isError`, not just process exit status. Keep stdin open until checks finish. Do not register or reconfigure the user's MCP client.

Require create/get/update/archive/search/suggest-context/link/unlink/neighbors/traverse and schema get/set/list/delete tools. Use tool schemas for exact inputs. MCP must read the CLI override; MCP schema_set must be visible through CLI schema_get and listed as persisted; deleting the override must restore the original embedded template. Request the prepare-adr prompt and verify it uses the active ADR template, minimal links, canonical metadata order, and no obsolete dependencies field; repeat with a private ADR schema override.

## 3. Identity, CRUD, filters, and conflicts

Create a reference guide through CLI:

```sh
gyrus create --id guide-reference --title 'Lifecycle reference' --type guide --category technical --tenant lifecycle --scope reference --status draft --tags consumer,okf --content referenceonlymarker --json
```

Through MCP create a draft project guide with id `guide-project`, tenant `lifecycle`, scope `project`, project `demo`, and body `projectonlymarker`. Require returned names `lifecycle/reference/guide-reference` and `lifecycle/projects/demo/guide-project`. Fresh CLI must get the MCP creation; running MCP must get the CLI creation. New concepts start at version 1.

Create `guide-reference` again in project demo and in tenant other; all three qualified names must resolve independently. Short-ID get, malformed local ID, and duplicate qualified creation must fail without changing an existing document. Use the installed CLI error contract: general validation 1, invalid lifecycle state 2, version conflict 4, not found 5. Check error payloads as well as codes.

Update the reference to active with content `updatedonlymarker` and expected version 1; update the project through MCP. Require version 2. Stale CLI and MCP updates using version 1 must fail without changing content/title. Active-to-draft must fail with status unchanged.

Search for `updatedonlymarker` plus exact tag `okf`: find the reference; tag `ok` must not match. Create a document in another project with unique body `foreignonlymarker`; default scope must exclude it and explicit project/scope must find it. Tenant other filtering must isolate its repeated ID. Use fresh CLI processes throughout to test persisted state.

## 4. Links, traversal, and context

Use the two reference/project documents plus the demo copy of `guide-reference`:

| Source | Target | Relationship |
| --- | --- | --- |
| project guide | reference guide | depends_on |
| project guide | lifecycle/reference/future-concept (absent) | implements |
| reference guide | demo copy | supersedes |
| demo copy | project guide | mitigates |
| project guide | itself | mitigates |

Create edges through both CLI `link <qualified-from> <qualified-to> --rel-type <rel> --json` and MCP `gyrus_link`. Unsupported relationships must fail. Repeating an existing link must not increment the source version. Outgoing neighbors must contain the stored edges; incoming neighbors on the reference must derive the depends_on backlink. No inverse edges should be stored.

MCP traversal from the project guide at depth 3 must finish despite the cycle/self-link and expose four unique nodes and five unique edges, including the absent target. Both MCP suggest_context and `gyrus suggest-context --prompt projectonlymarker --max-concepts 2 --json` must include project and reference. Use this unique body marker: a qualified-name prompt can match unrelated documents sharing tenant/project tokens.

## 5. UI, direct edits, and restart

Start an owned `gyrus ui serve --config <private-config> --host 127.0.0.1 --port <available-port>`. Use a browser or available HTTP tool, bypassing proxies for loopback. Poll readiness with a bounded deadline. Require successful HTTP responses and meaningful content at `/`, `/concepts`, `/graph`, `/search?q=updatedonlymarker`, `/concepts/lifecycle/reference/guide-reference`, and `/api/graph`. Graph node IDs must be qualified and include the missing target; graph edges must match the five stored edges. Concept page must render the body marker.

Add a private graph boundary fixture: a visible reference concept points to a concept in another project, which points onward to a third name. The UI graph must show the visible source's edge and the immediate target as a ghost; it must not show the hidden source's outgoing edge or body/title metadata. Repeat the same HTTP request and require byte-identical node/edge ordering. Include these fixtures in archive cleanup. Perform this check separately from the five-edge baseline assertions.

Get the reference's returned filepath and confirm it is inside private storage. Replace only body marker `updatedonlymarker` with `externaleditmarker` using file-edit tools. Without sync/restart, require running MCP search, fresh CLI get, and the UI concept page to reflect it. `gyrus validate <filepath> --json` must pass. Restart only the owned MCP process and verify the graph persists.

## 6. Concurrency, lifecycle rules, and limits

- Create a fresh guide, then start two independent CLI update processes with different titles and expected version 1 before awaiting either. Require exactly one success and one conflict (4), final version 2, and the successful title. Run against private localfs only.
- Create a draft improvement-proposal. Reject a jump to implemented, then accept draft → reviewing → approved → implemented using current versions.
- Create an ADR as proposed, then accept it. Body change must fail (2); a metadata-only title update must succeed and preserve the accepted body.
- Create 12 guides containing `keywordlimitmarker`. CLI and MCP default keyword search must return ten. CLI `--max-results 3` and MCP `max_results: 3` must return three; requesting 20 through each must return all twelve. Empty-query browsing with `--scope all --max-results 0` must return the complete owned inventory. MCP empty-query browsing with scope all must also be complete; a positive limit must cap it. Do not infer completeness from a capped keyword query.

## 7. Malformed input and persisted contract

Create exactly one owned malformed `.md` under private `concepts/lifecycle/reference/`. CLI search must return nonzero with the malformed file in its error; MCP search must set isError and expose it. Search must fail explicitly rather than silently return partial results. Remove only the malformed fixture, then require query recovery.

Inspect all stored fixture frontmatter. Present fields must follow this order; unrelated extension keys follow these keys:

`title, id, type, format, category, version, status, last_updated, last_modified_by, tenant, tags, links`

Require local frontmatter id and links with exactly `{concept, rel}` and fully qualified targets. No dependencies, target_id, rel_type, or per-link created_by/created_at; no persisted derived name/scope/project/filepath. Document-level version and modification metadata remain valid. Public graph edges must also omit link creation metadata.

## 8. Unlink, archive, cleanup, and report

Unlink the project→reference depends_on edge through MCP. Require its absence through graph queries and a fresh read; repeating CLI unlink must be a version no-op. Restart the owned MCP server and ensure the removed edge stays absent. Archive every owned concept, alternating CLI/MCP; each subsequent get must fail with not-found (5). Require empty all-scope search and empty UI graph. Require no SQLite/database files and byte-identical initialized configuration.

Stop/reap only owned services. Preserve the evidence report; remove private workspace/home only on complete success. Otherwise retain them and record the failed assertion, last successful phase, process cleanup outcome, and recovery paths. Mark unavailable MCP/browser/concurrency capabilities blocked and overall verification incomplete; never reuse a live service to fill the gap.
