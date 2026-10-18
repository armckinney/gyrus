---
title: Gyrus Local Testing Commands Walkthrough
id: guide-005-local-testing-walkthrough
type: guide
format: markdown
category: technical
version: 4
status: active
last_updated: 2026-10-17T23:59:37Z
last_modified_by: ""
tenant: root
tags:
    - testing
    - manual-testing
    - cli-walkthrough
---

# Local consumer lifecycle walkthrough

Use an installed Gyrus binary in a new terminal/subshell. The packaged gyrus-test-lifecycle skill provides the more extensive CLI/MCP/UI acceptance procedure. This sequence uses localfs with OKF and requires no source checkout, compiler, embedding service, or Make commands.

## Private setup

```bash
gyrus_binary=$(command -v gyrus)
test_directory=$(mktemp -d)
mkdir "$test_directory/home" "$test_directory/project"
cd "$test_directory/project"
export HOME="$test_directory/home"
export USERPROFILE="$test_directory/home"
unset GYRUS_CONFIG GYRUS_PROJECT
export GYRUS_PROJECT="$test_directory/project"
"$gyrus_binary" config init --profile local --tenant root --project demo
export GYRUS_CONFIG="$test_directory/project/.gyrus.yaml"
"$gyrus_binary" config show --json
```

Require localfs storage inside this private project and OKF index/graph/search. Supported profiles are local, git, blob, s3, azure, gcs, and postgres. Unknown profiles fail without replacing configuration; unknown provider names use shared validation. SQLite providers and sync are removed.

## Templates, creation, and qualified reads

```bash
"$gyrus_binary" schema get adr
"$gyrus_binary" schema get guide
"$gyrus_binary" create --id adr-storage --title "Direct OKF retrieval" --category architecture --type adr --tenant root --scope reference --status proposed --tags okf,architecture --content "Use authoritative Markdown and typed concept links."
"$gyrus_binary" create --id guide-retrieval --title "Retrieval guide" --category technical --type guide --tenant root --scope project --project demo --status active --content "retrievalwalkthroughmarker"
"$gyrus_binary" get root/reference/adr-storage --json
"$gyrus_binary" get root/projects/demo/guide-retrieval --json
```

Creation takes a local ID and returns name plus revision. All subsequent operations use that qualified name. Observe returned versions rather than assuming a revision after additional mutations.

## Lifecycle and concurrency

```bash
"$gyrus_binary" update root/reference/adr-storage --status accepted --expected-version 1 --json
"$gyrus_binary" update root/reference/adr-storage --title "Direct OKF retrieval decision" --expected-version 2 --json
```

The accepted ADR body is immutable; metadata remains editable. A stale expected version fails with CLI exit code 4. Invalid transitions or immutable-body edits fail with code 2. PostgreSQL uses the same conflict classification with transaction rollback.

## Directed relationships and context

```bash
"$gyrus_binary" link root/projects/demo/guide-retrieval root/reference/adr-storage --rel-type depends_on --json
"$gyrus_binary" suggest-context --prompt retrievalwalkthroughmarker --max-concepts 2 --json
```

Context includes the project guide and its reference dependency. Supported relations are depends_on, implements, supersedes, and mitigates. OKF links are stored on their source as concept/rel pairs; incoming neighbors derive backlinks. MCP exposes neighbors and traversal; there are no CLI commands with those names.

## Search and validation

```bash
"$gyrus_binary" search --query retrievalwalkthroughmarker --max-results 20 --json
"$gyrus_binary" search --query "" --scope all --max-results 0 --json
"$gyrus_binary" validate .gyrus/concepts/root/reference/adr-storage.md --json
"$gyrus_binary" validate .gyrus/concepts/root/projects/demo/guide-retrieval.md --json
```

Keyword searches default to ten results and accept explicit limits above or below ten. Empty-query browsing is unlimited unless a positive limit is supplied. OKF performs direct keyword scans; PostgreSQL offers native full-text search. Neither requires vectorization. Invalid concepts fail collection queries explicitly.

## UI and MCP

Start gyrus ui serve --config "$GYRUS_CONFIG" --host 127.0.0.1 --port 8080 in a separate terminal with the private environment. Verify /concepts, the qualified concept route, and /api/graph. Visible sources contribute outgoing edges; immediate filtered/missing targets appear as ghosts without further expansion or hidden metadata. Responses sort and deduplicate nodes/edges; edge failures return HTTP 500.

Start gyrus mcp serve --config "$GYRUS_CONFIG" with a client that maintains its initialized stdio session. Discover tools and use max_results for search, max_concepts for context, and expected_version for updates. The prepare-adr prompt uses the active shared ADR template. Stop only the UI/MCP processes started for this private workspace.

## Durable unlink and cleanup

```bash
"$gyrus_binary" unlink root/projects/demo/guide-retrieval root/reference/adr-storage --rel-type depends_on --json
"$gyrus_binary" get root/projects/demo/guide-retrieval --json
"$gyrus_binary" archive root/projects/demo/guide-retrieval --json
"$gyrus_binary" archive root/reference/adr-storage --json
"$gyrus_binary" search --query "" --scope all --max-results 0 --json
```

The fresh read must show the link absent; a repeated unlink is a version no-op. Final browsing must return an empty collection. After stopping owned processes, remove only the recorded private test directory. Keep fixtures and logs for diagnosis when a check fails. The local profile creates no index database.
