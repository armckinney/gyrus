---
title: Git initiatives and human review
id: guide-006-git-initiatives
type: guide
format: markdown
category: technical
version: 1
status: active
last_updated: 2026-11-12T04:33:26Z
tenant: root
---

# Git initiatives and human review

## Configure the shared knowledge repository

Create a normal Git repository, not a GitHub Wiki service. All knowledge lives beneath `.gyrus/concepts` and `.gyrus/schemas`. Initialize `main` with an initial commit or a direct-mode Gyrus write before using branching.

Configure each code repository with the same `git.repo_url` and an explicit project name. Use the complete Git/OKF configuration in the [configuration reference](tech-ref-002-config-schema.md). Set `commit_strategy: branching` for required review or `direct` for trunk writes.

## Configure authentication

For HTTPS, install Git and configure Git Credential Manager or another Git credential helper. Authenticate to the host outside Gyrus first. Gyrus invokes `git credential fill` with terminal/GUI prompts disabled and uses the returned username/password unchanged. This supports the host-specific usernames required by different token types. `GIT_AUTH_TOKEN` and `GITHUB_TOKEN` do not override helper credentials.

For GitHub, follow [credential-manager guidance](https://docs.github.com/en/get-started/git-basics/caching-your-github-credentials-in-git). For Bitbucket, configure the matching username and token through the helper using [Atlassian's API-token instructions](https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/). Bitbucket was not live-tested in this delivery.

SSH remotes respect their username and use an available agent or standard unencrypted `~/.ssh/id_ed25519`, `id_rsa`, or `id_ecdsa` keys. Populate known hosts; host verification remains enabled. Custom SSH configurations and encrypted key-file prompting are not added by this delivery.

## Work on an initiative

```bash
gyrus git status --git-branch initiative/payments
gyrus search --query payments --git-branch initiative/payments
gyrus mcp serve --git-branch initiative/payments
```

Supply the branch on each independent CLI invocation. The MCP server retains it for the process lifetime. The branch is never saved to `.gyrus.yaml`; restarting requires selecting it again. A missing initiative reads trunk until its first mutation pushes the new branch.

Switch a running MCP server:

```json
{"strategy":"branching","branch":"initiative/payments"}
```

Pass this to `gyrus_git_switch`; inspect `gyrus_git_status`. Return to trunk using `{"strategy":"direct"}`. Selection failures leave the current branch unchanged. Knowledge operations in flight finish before a switch becomes active.

Start a second MCP process with `--git-branch initiative/billing` for parallel work. Switching the first process leaves the second on its initiative. Project scoping and branch selection are independent: a branch may contain changes to several documents/projects for one initiative.

## Review and merge

Open a PR from the initiative into `trunk_branch` through GitHub or your Git host. Review all concepts, schemas, and links together. Merge externally, then switch to direct to read the approved trunk state. Choose a fresh initiative or manually integrate trunk changes before continuing an existing one.

Gyrus does not open PRs, merge, rebase, resolve divergence, or force-push. A rejected push returns an error; stale/conflicting changes return a conflict. Retry after reading the current selected branch and resolving the conflict.

## Troubleshooting

A required-selection error means the configured strategy is branching but no runtime branch was supplied. An empty remote cannot serve as the base of an initiative: initialize trunk first. Authentication errors require checking helper credentials or SSH access outside Gyrus. An existing initiative does not automatically receive new trunk changes. `git.branch`, `base_branch`, and `review_branch` are unsupported configuration keys.
