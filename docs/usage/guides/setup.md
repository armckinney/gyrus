# Gyrus Setup & Onboarding Guide

This guide walks human developers and team leads through initializing Gyrus, organizing team context documents, and enabling AI assistants to read and maintain codebase memory.

---

## 1. Setting Up Workspace Configuration

To configure Gyrus for your repository, create a `.gyrus/config.yaml` file in the root of your project:

```yaml
version: 1
profile: small # test | local_full | small | medium | large

storage:
  provider: localfs
  root: ./.gyrus/docs

index:
  provider: sqlite
  dsn: ./.gyrus/index.db

graph:
  provider: sqlite

search:
  provider: sqlite_fts5
```

Add `.gyrus/index.db` to your `.gitignore` so local SQLite databases are not committed to version control:

```text
# .gitignore
.gyrus/index.db
```

---

## 2. Bootstrapping Contract Documents

Gyrus uses 11 standardized document types under the **Open Knowledge Format (OKF)**:

| Document Type | Recommended Use Case |
| :--- | :--- |
| **`adr`** | Architecture Design Records documenting key decisions. |
| **`prd`** | Product Requirement Documents defining features and goals. |
| **`guide`** | Developer guides and onboarding walkthroughs. |
| **`improvement-proposal`** | System enhancement proposals. |
| **`release-note`** | Version release summaries and changelogs. |
| **`specification`** | Detailed technical specifications. |
| **`standards`** | Team coding standards and conventions. |
| **`technical-reference`** | Module API and data model references. |
| **`product`** | High-level product overview documents. |
| **`glossary`** | Domain taxonomy and definitions. |
| **`freeform`** | Notes and unformatted engineering docs. |

To create a new ADR from a template:

```bash
gyrus schema adr > adr-001-my-feature.md
# Edit frontmatter and content, then register in Gyrus:
gyrus sync
```

---

## 3. Configuring Terminal CLI Agents (Claude Code, Aider)

For terminal-based agents, copy the Open Skill Format definition from `skills/gyrus/SKILL.md` into your agent skills folder (`.agents/skills/gyrus/SKILL.md`).

Agents will automatically run `gyrus suggest-context --prompt "<task>"` to retrieve precise documentation context before modifying code.

---

## 4. Re-indexing and Maintaining Memory

Whenever team members create or edit Markdown documents by hand:

```bash
gyrus sync
```

`gyrus sync` scans your storage directory, calculates SHA-256 file checksums, updates SQLite FTS5 indexes, and extracts document linkage edges automatically.
