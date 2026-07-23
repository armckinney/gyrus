---
id: ip-004-provider-architecture
title: Pluggable Storage & Index Provider Architecture Proposal
category: architecture
type: improvement-proposal
format: ""
owner_group: armckinney
version: 1
status: active
last_modified_by: ""
last_updated: 2026-07-22T06:38:56Z
---

# Gyrus providers

| Profile        | Storage      | Index         | Graph                  | Search               | Best for                                   |
| -------------- | ------------ | ------------- | ---------------------- | -------------------- | ------------------------------------------ |
| **Test**       | `localfs`    | OKF or SQLite | OKF or SQLite          | SQLite FTS5 optional | Unit tests, fixtures, CI                   |
| **Local Full** | SQLite       | SQLite        | SQLite edge tables     | SQLite FTS5          | Local app, deterministic integration tests |
| **Small**      | Git or blob  | OKF           | OKF links/metadata     | OKF/local scan       | Personal/team repo, low volume             |
| **Medium**     | Blob storage | PostgreSQL    | PostgreSQL edge tables | PostgreSQL full-text | Team/service deployment                    |
| **Large**      | PostgreSQL   | PostgreSQL    | PostgreSQL             | PostgreSQL FTS       | Centralized platform/service               |
