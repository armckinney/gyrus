---
id: <unique-dashed-id>
title: <Clean Document Title>
category: operations
type: release-note
format: markdown
owner_group: <team-owning-document>
version: 1
status: proposed
last_modified_by: <author-username>
last_updated: YYYY-MM-DD
tags: []
dependencies: []
---

# Release Notes: [Project Name] v[Version]

## 1. Summary & Highlights

* **Release Date:** [Fill: YYYY-MM-DD]
* **Target Version:** `vX.Y.Z`
* **Release Branch:** `release/vX.Y.Z`

### High-Level Summary
[Fill: Provide a brief 1-2 paragraph description of the main focus of this release. E.g., "This release introduces SQLite Edge Database support for offline agent queries, along with stability improvements under network drops."]

---

## 2. Changelog

### 🚀 Features & Enhancements
* **[Feature Name]:** [Fill: Brief description of the new capability and links to specifications or JIRA tickets.]
* **[Enhancement Name]:** [Fill: Brief description.]

### 🐛 Bug Fixes
* **[Fix Name]:** [Fill: Description of the bug resolved, links to issues.]
* **[Fix Name 2]:** [Fill: Description.]

### ⚠️ Deprecations & Breaking Changes
* **Deprecation:** [Fill: Describe deprecated endpoints or variables, when they will be removed.]
* **Breaking Change:** [Fill: Describe breaking modifications and their impact.]

---

## 3. Upgrade / Migration Guide

[Fill: Provide step-by-step instructions if special actions are required to migrate to this version.]

```bash
# Example upgrade steps
npm install @gyrus/client@latest
# Run DB migrations
gyrus db migrate
```
