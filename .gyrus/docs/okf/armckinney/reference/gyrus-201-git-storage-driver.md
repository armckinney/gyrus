---
id: gyrus-201-git-storage-driver
title: 'GYRUS-201: Git Remote Storage Driver'
category: technical
type: specification
format: ""
owner_group: armckinney
version: 1
status: completed
last_modified_by: ""
last_updated: 2026-07-23T21:23:12Z
tags:
    - phase-2.1
    - storage-driver
    - git
---

# GYRUS-201: Git Remote Storage Driver

## 1. Overview & Objective
Implement a remote Git repository `DocumentStore` provider under `internal/provider/git` using `go-git` (`github.com/go-git/go-git/v5`). Enables Gyrus to read, create, update, delete, and commit OKF document files directly to any remote Git repository (GitHub, GitLab, Bitbucket, Azure Repos, self-hosted) over HTTPS or SSH without requiring local workspace clones or git binary executions.

## 2. Requirements & Constraints
- Must implement `gyrus.DocumentStore` interface (`Create`, `Get`, `Update`, `Delete`, `Archive`).
- Must support standard Git credential auto-discovery (`git-credential-provider`, `GIT_AUTH_TOKEN`, `GITHUB_TOKEN`, standard SSH keys `~/.ssh/id_*`).
- Must commit OKF document changes with structured commit messages including author identity and document version.
- Must execute in-memory with zero local `git` binary dependency.

## 3. Key Test Verification
- Unit test in `internal/provider/git/git_test.go` verifying in-memory Git repo CRUD operations.
- Integration test verifying commit creation and document retrieval over Git transport.
