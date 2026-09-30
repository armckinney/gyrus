---
id: adr-005-global-configuration-and-precedence-hierarchy
title: Global Configuration and Simplified Precedence Hierarchy
category: architecture
type: adr
format: ""
owner_group: root
version: 1
status: accepted
last_modified_by: ""
last_updated: 2026-09-30T21:41:45Z
---

# ADR 005: Global Configuration and Simplified Precedence Hierarchy

## Context

Prior to Phase 2.7, Gyrus configuration resolution was tightly coupled to the localfs storage package (internal/provider/storage/localfs/resolver.go). This created several architectural and user-experience issues:

1. **Storage Driver Coupling**: All packages requiring configuration settings were forced to import a concrete storage driver package (localfs), creating circular coupling and architectural violations.
2. **Schema Divergence**: The repository's .gyrus.yaml file used a nested structure (storage.provider, index.provider), while the setup generator emitted a flat schema (storage_provider, index_provider). The provider factory only read flat fields, causing nested workspace configurations to silently fall through to default values.
3. **Redundant Precedence Complexity**: Configuration resolution attempted a 5-tier search with multiple filename variants (.gyrus.yaml, .gyrus.yml, .gyrus/config.yaml, .gyrus/config.yml, ~/.config/gyrus/config.yaml), combined with CLI flags (--storage-path) and environment variables (GYRUS_STORAGE_PATH). This created unpredictable behavior across workspaces and testing environments.
4. **Lack of User-Wide Defaults**: Users working across multiple repositories had no way to define persistent preferences (default owner group, shared remote connection strings) without manually configuring each workspace.

## Decision

We decide to:

### 1. Extract Dedicated internal/config Package
Extract Config, ResolvedConfig, Load(), and resolution logic out of internal/provider/storage/localfs into an independent internal/config package. Storage drivers now depend only on domain interfaces or explicit storage root strings.

### 2. Standardize on Canonical Nested YAML Schema
Formalize the nested YAML schema as the single canonical configuration format across all profiles and files:
```yaml
storage:
  provider: localfs
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

default_owner_group: root
```
Convenience accessors (StorageProvider(), IndexProvider(), SearchProvider(), GraphProvider(), StorageRoot()) provide type-safe access.

### 3. Simplify Precedence Chain to File-Driven Hierarchy
Eliminate --storage-path and GYRUS_STORAGE_PATH in favor of a strictly file-driven 3-tier precedence chain with full override semantics:
Workspace .gyrus.yaml > Global ~/.gyrus.yaml > Hardcoded Defaults.
If a workspace .gyrus.yaml is present, it completely overrides global configuration. Relative paths in configuration files resolve relative to the directory containing that configuration file.

### 4. Single Canonical File Name
Only .gyrus.yaml (in workspace root or user home ~/.gyrus.yaml) is searched. Deprecated variants (.gyrus.yml, nested .gyrus/config.yaml, and XDG ~/.config/gyrus/) are removed.

### 5. Config Inspection & Initialization CLI Commands
- gyrus config show: Displays resolved configuration and source metadata.
- gyrus init config --global: Writes user-wide configuration to ~/.gyrus.yaml (refusing overwrite without --force).

### 6. Fail-Fast Load-Time Validation
Validate provider names at config load time against recognized provider drivers, emitting descriptive error messages early rather than failing downstream at factory instantiation.

## Consequences

* **Positive / Gains**:
  * Clean separation of configuration from storage implementation.
  * Predictable, deterministic resolution behavior across local and containerized environments.
  * Unified schema between CLI templates and repository configs.
  * Easy multi-repo developer onboarding via ~/.gyrus.yaml.
* **Trade-offs / Breaking Changes**:
  * Removed --storage-path CLI persistent flag (must configure via .gyrus.yaml).
  * Removed GYRUS_STORAGE_PATH environment variable.
  * Removed .gyrus.yml and .gyrus/config.yaml filename variants.
  * Flat configuration keys are replaced with nested keys.
