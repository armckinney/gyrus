---
title: Pluggable Storage & Index Provider Architecture Proposal
id: ip-004-provider-architecture
type: improvement-proposal
format: ""
category: architecture
version: 2
status: abandoned
last_updated: 2026-11-12T04:38:04Z
last_modified_by: ""
tenant: root
---

> Historical record. This document describes earlier work and is not a current setup or capability contract. Use the current configuration reference and architecture index.

# Gyrus providers

| Profile        | Storage      | Index         | Graph                  | Search               | Best for                                   |
| -------------- | ------------ | ------------- | ---------------------- | -------------------- | ------------------------------------------ |
| **Test**       | `localfs`    | OKF or SQLite | OKF or SQLite          | SQLite FTS5 optional | Unit tests, fixtures, CI                   |
| **Local Full** | SQLite       | SQLite        | SQLite edge tables     | SQLite FTS5          | Local app, deterministic integration tests |
| **Small**      | Git or blob  | OKF           | OKF links/metadata     | OKF/local scan       | Personal/team repo, low volume             |
| **Medium**     | Blob storage | PostgreSQL    | PostgreSQL edge tables | PostgreSQL full-text | Team/service deployment                    |
| **Large**      | PostgreSQL   | PostgreSQL    | PostgreSQL             | PostgreSQL FTS       | Centralized platform/service               |
