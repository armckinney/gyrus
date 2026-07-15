# Unified Context Layer (UCL) Specification Index

This directory contains the modular design specifications for the Unified Context Layer (UCL), a CosmosDB-First architecture designed for cross-repository human and machine context sharing.

## Modular Specifications

1. **[01. JSON Contract Specification](file:///Users/armck/git/wiki/docs/specs/01_contract_schema.md):** Defines the metadata schema envelope for all context documents.
2. **[02. CLI Interface & Authentication](file:///Users/armck/git/wiki/docs/specs/02_cli_interface.md):** Outlines the `ucl` CLI subcommands, caching, and Azure AD / Device Code auth pipelines.
3. **[03. Human Web Wiki & Sitemap](file:///Users/armck/git/wiki/docs/specs/03_web_wiki.md):** Details the Go/Python server rendering Markdown/JSON, and the dynamic `/llms.txt` crawler endpoint.
4. **[04. Architectural Trade-offs](file:///Users/armck/git/wiki/docs/specs/04_tradeoffs.md):** Matrices explaining the reasoning behind CosmosDB vs. Git, CLI vs. MCP, and Go/HTMX vs. React.
5. **[05. Composability & Portability](file:///Users/armck/git/wiki/docs/specs/05_composability.md):** The Repository Pattern for swappable database backends and agent shims (Claude Code, Cursor/MCP).
6. **[06. Security Architecture](file:///Users/armck/git/wiki/docs/specs/06_security.md):** Explains Entra ID AuthN, ABAC AuthZ query isolation, input shielding, and edit auditing.
7. **[07. Reliability & Fault Tolerance](file:///Users/armck/git/wiki/docs/specs/07_reliability.md):** Resiliency mechanisms including local offline caching, backoffs, and defensive agent exit coding.
8. **[08. Technical Specifications & Endpoints](file:///Users/armck/git/wiki/docs/specs/08_technical_specs.md):** Cosmos DB partitioning keys, collection layouts, and REST API endpoints.
9. **[09. Implementation Roadmap](file:///Users/armck/git/wiki/docs/specs/09_roadmap.md):** Complete WBS showing Epics and JIRA tickets for development engineering teams.
10. **[10. Product Mapping & Repository Layout](file:///Users/armck/git/wiki/docs/specs/10_product_mapping.md):** Product naming, monorepo design benefits, and project folder structures.

---

## Core System Blueprint

```
             [ User-Delegated Session / Azure AD ]
                             │
     ┌───────────────────────┴───────────────────────┐
     ▼                                               ▼
[ Human Web Portal ]                           [ `ucl` CLI Query Tool ]
(Go+HTMX or Python)                            (Client-agnostic terminal utility)
     │                                               │
     ├─► Exposes dynamic `/llms.txt` sitemap         ├─► Local Offline Cache (~/.cache/ucl/)
     │                                               │
     └───────────────┬───────────────────────────────┘
                     ▼ (HTTP REST API with OAuth2 JWT)
            [ API Gateway / Proxy ]
                     │
         [ Security & AuthZ Filter ] ◄── Validates JSON contracts & JWT claims
                     │
         [ Swappable Storage Interface ]
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
  [ Cosmos DB Driver ]   [ Git Storage Driver ]
  (Active Production)    (Future Swappable Option)
```
# Specification 01: JSON Contract Schema

Every document in the Unified Context Layer is stored as a JSON object adhering to a standard metadata envelope. The CLI and API Gateway validate these fields against a JSON Schema to prevent formatting drifts.

## The JSON Contract Payload Example

```json
{
  "id": "billing-rules",
  "title": "Billing Logic & Subscriptions",
  "category": "business-logic",
  "type": "adr",
  "format": "markdown",
  "owner_group": "group-a",
  "version": 3,
  "last_modified_by": "agent-claude-joe",
  "last_updated": "2026-07-08T00:40:00Z",
  "tags": ["pricing", "stripe", "finance"],
  "dependencies": ["database-schema-v1"],
  "content": "# Billing Rules\n\nAll subscriptions default to a 14-day trial period unless explicitly overridden.\n\n## 1. Tax Calculation\n- EU customers are taxed at the regional VAT rate.\n- US customers are taxed based on zip code."
}
```

## Schema Fields & Validation Rules

* **`id` (string, required):** Unique alphanumeric ID identifying the document (e.g. `user-auth-policy`). Regex enforced: `^[a-z0-9-_]+$`.
* **`title` (string, required):** A short human-readable title.
* **`category` (string, required):** Determines document grouping. Must match enum: `[architecture, business-logic, api-contract, operations]`.
* **`type` (string, optional):** The specific document sub-type. Enforces standard structures via CLI templates. Allowed standard types:
  * `adr`: Architecture Decision Record.
  * `product-spec`: Product Requirements Document / Specification.
  * `guide`: Onboarding docs, runbooks, configuration procedures.
  * `tech-plan`: Tech Specs, Improvement Proposals, and Implementation plans.
  * `api-contract`: Interface schemas, OpenAPI definitions, endpoints.
  * `db-schema`: Relational/document database schemas, tables, indexes, keys.
  * `env-spec`: Environment variables, secret requirements, deployment configurations.
  * `freeform`: Freeform context documents (catch-all for vital unstructured details).
* **`format` (string, required):** Declares payload type. Must match enum: `[markdown, json, yaml]`.
* **`owner_group` (string, required):** Mapping target for user access groups (e.g. `group-a`). Used as the Cosmos DB partitioning key.
* **`version` (integer, required):** Auto-incrementing transaction index for edit tracing.
* **`last_modified_by` (string, auto-populated):** Email or service principal ID of the last editor. Overridden by API gateway JWT claims.
* **`last_updated` (string, auto-populated):** ISO 8601 timestamp of write execution. Overridden by API gateway server time.
* **`tags` (array of strings, optional):** Index keys for search filtering.
* **`dependencies` (array of strings, optional):** Lists dependent document IDs.
* **`content` (string, required):** Freeform payload containing raw document text. Recommended to format as standard Markdown.
# Specification 02: CLI Interface & Authentication

The CLI tool (`ucl`) serves as the universal developer and agent entry point to query the global database. By relying on shell executions, agents avoid importing heavy database client libraries or suffering prompt token bloat.

## CLI Command Interface

* **`ucl search "<query>"`**
  * *Behavior:* Queries the REST search endpoint `/api/v1/search` with keyword/tag parameters and outputs matching document lists.
* **`ucl read <doc-id>`**
  * *Behavior:* Retrieves full document payload and prints the raw `content` field to standard output. Supports fallback caching.
* **`ucl write <doc-id> <json-file-or-string>`**
  * *Behavior:* Validates metadata format locally, validates authorization, and posts JSON to `/api/v1/documents/{id}`.
* **`ucl history <doc-id>`**
  * *Behavior:* Queries `/api/v1/documents/{id}/history` and prints edit trails.
* **`ucl schema <type>`**
  * *Behavior:* Outputs recommended best-practice document templates and schemas for a specific document `type` (e.g. `ucl schema adr`).
  * *Purpose:* Provides immediate guidance to developers and agents on how to structure a document before creating or modifying it. (Non-enforceable guidance, purely advisory).

---

## Authentication Flow (AuthN)

To eliminate complex credentials management, `ucl` leverages the user's existing security context:

```
[ ucl login ] ──► Check Azure CLI Session (az account get-access-token)
      │
      ├── (Token Valid) ─────► Write to local session cache (~/.config/ucl/token.json)
      │
      └── (Token Missing) ──► Trigger OAuth 2.0 Device Code Flow
                                - Print URL & Verification Code
                                - Poll /token endpoint for JWT
```

### 1. Azure CLI Integration (Primary)
The CLI runs:
```bash
az account get-access-token --resource <API_CLIENT_ID>
```
If logged in, it retrieves the OAuth2 JWT token.

### 2. Device Code Flow (Fallback)
If Azure CLI is unavailable:
* CLI requests a device code from Microsoft Entra ID.
* Outputs authentication instructions to `stderr`:
  `Please navigate to https://microsoft.com/devicelogin and enter code: ABCD-EFGH`
* User validates access, and CLI caches token.

### 3. Session Caching
The CLI caches session data in a JSON file at `~/.config/ucl/token.json` containing the JWT and refresh metadata.
# Specification 03: Human Web Wiki & Sitemap

The human web portal provides developers with an interactive, searchable portal to navigate, update, and audit documentation contracts without database tool complexities.

---

## 1. Web Portal Architecture (Go+HTMX or Python)

* **Server-Side Rendered:** Uses lightweight HTML templates coupled with HTMX for low latency transitions.
* **Format-Agnostic rendering:** 
  * Checks document's `format` field.
  * Markdown (`format: "markdown"`): Server compiles Markdown to HTML on the fly.
  * JSON / YAML (`format: "json"`): Server highlights syntax blocks.
* **Audit Dashboard:** Displays version timelines and editor identities for audit control.

---

## 2. Dynamic Sitemap `/llms.txt`

The Web Wiki hosts a text/plain sitemap endpoint at the root domain (`/llms.txt`) designed for machine consumption.

### Dynamic Generation Logic
1. **Request Received:** Crawler requests `https://wiki.company.internal/llms.txt`.
2. **Access Evaluation:** Server reads authorization headers. If authorized, extracts the user's/agent's Entra ID group list.
3. **Partition-Restricted Query:** The server queries Cosmos DB, restricting lookups strictly to the user's authorized `/owner_group` partitions to prevent cross-partition query scans:
   ```sql
   SELECT c.id, c.title, c.category, c.tags FROM c WHERE c.owner_group IN ('group-a', 'group-b')
   ```
4. **Markdown Formatting:** Compiles output as token-dense plain text:
   ```markdown
   # Company Wiki Index (Scoped to Group A/B)
   
   ## business-logic
   - [billing-rules](/api/v1/documents/billing-rules): Billing logic and tax rules.
   
   ## architecture
   - [db-schema](/api/v1/documents/db-schema): Database layout specs.
   ```
5. **Efficiency Payload:** This allows remote agents (Slack bots, support assistants) to map the entire library structure in a single network call.
# Specification 04: Architectural Trade-offs & Decisions

The following design matrices highlight the key tradeoffs evaluated when arriving at this architecture, emphasizing performance, security, and developer friction.

---

## 1. Storage Backend: CosmosDB-First vs. Git-Wiki

| Dimension | Git-Wiki (Submodules & Symlinks) | CosmosDB-First (Selected) |
| :--- | :--- | :--- |
| **Pros** | - Offline human access.<br>- Standard Git branching and PRs.<br>- Low hosting costs. | - Single source of truth.<br>- High-frequency concurrent agent writes.<br>- Zero Git workspace clutter. |
| **Cons** | - Workspace configuration overhead.<br>- Stale clones if developers forget to pull.<br>- GitHub/GitLab API rate limits. | - Requires database infrastructure provisioning.<br>- Lacks Git branch history (replaced by DB versioning). |
| **Performance** | **Poor Latency:** Pulling/syncing remote repositories takes seconds. Local reads are microsecond disk I/O, but synchronization is slow. | **High Performance:** Sub-10ms read/write latency. Scalable indexing and single-partition queries eliminate network limits. |

---

## 2. Client Access Protocol: CLI Engine (`ctx`) vs. Model Context Protocol (MCP) vs. Skills

| Dimension | Model Context Protocol (MCP) | Agent-Framework Skills | CLI Utility (`ctx`) (Selected) |
| :--- | :--- | :--- | :--- |
| **Pros** | - Open standard.<br>- GUI IDEs (Cursor) support natively. | - Deep framework integration.<br>- Native UI components. | - Universal compatibility.<br>- Zero startup prompt bloat.<br>- Enforces strict validation. |
| **Cons** | - High startup handshakes.<br>- Lack of CLI-only agent support. | - Framework/client-locked.<br>- Higher agent flexibility (prone to hallucinations). | - Requires terminal execution.<br>- Requires local installation. |
| **Performance** | **Poor Token Latency:** Startup handshake injects full tool schemas, bloating prompts and raising token costs on every turn. | **Fast Execution:** Runs directly inside the agent's Python/Go run runtime. | **High Token Efficiency:** Small, targeted shell execution. Agents retrieve only relevant files, cutting tokens by up to 90%. |

---

## 3. Access Control: User-Delegated vs. Centralized Admin Key

| Dimension | Central Admin API Key | User-Delegated AD Auth (Selected) |
| :--- | :--- | :--- |
| **Pros** | - Single credential config.<br>- Simplest CLI coding setup. | - Inherited permissions (AD/Entra ID).<br>- Dynamic group partitioning.<br>- High auditing footprint. |
| **Cons** | - Critical security leak vector.<br>- Unauditable (all agents share one key). | - Session tokens must be cached locally. |
| **Performance** | **Fast Auth:** No lookup overhead; direct token match. | **Low Overhead:** Small JWT validation overhead at the Gateway (sub-1ms using cached public key certs). |

---

## 4. Frontend Rendering: Server-Side Jinja/HTMX vs. Single Page App (React)

| Dimension | React / Next.js / Blazor | Go+HTMX / Python (Selected) |
| :--- | :--- | :--- |
| **Pros** | - High interactivity.<br>- Advanced stateful graphs. | - Single binary deployment.<br>- Agnostic to payload schema changes.<br>- Fast, dynamic Markdown-to-HTML rendering. |
| **Cons** | - Complex build pipelines.<br>- Breaks easily on contract updates. | - Basic browser styling. |
| **Performance** | **High TTI:** Bundle downloads and hydration slow down load times over VPN or remote networks. | **Sub-100ms Load Times:** Tiny HTML payloads stream natively, offering instantaneous load times. |
# Specification 05: Composability & Portability

The UCL is designed with strict boundaries to ensure storage drivers can be swapped and agent clients can easily consume it.

---

## 1. Storage Abstraction (Repository Pattern)

To future-proof the storage layer, the API Gateway does not interact directly with Cosmos DB SDKs. Instead, it utilizes a decoupled interface. In Go or Python, this contract enforces swappability:

```go
type StorageProvider interface {
    GetDocument(ctx context.Context, id string, userGroups []string) (*Document, error)
    SaveDocument(ctx context.Context, doc *Document) error
    SearchDocuments(ctx context.Context, query string, userGroups []string) ([]SearchResult, error)
    GetHistory(ctx context.Context, id string, userGroups []string) ([]HistorySnapshot, error)
}
```

### Swapping to a Git-Wiki Filesystem Backend
Should the organization decide to migrate from Cosmos DB to Git:
1. **Develop Git Storage Driver:** An engineer implements the `StorageProvider` interface using a Git library (like `go-git`).
   * `SaveDocument`: Writes the JSON contract directly into a local clone of the wiki repository in the server environment, executes `git add`, `git commit`, and pushes to the central GitLab/GitHub host.
   * `SearchDocuments`: Searches files locally in the clone using grep/ripgrep or a local embedded key-value cache (e.g., SQLite) generated on file-change hooks.
2. **Switch Configuration:** Swapping the backend requires changing a single environment variable on the API Gateway (e.g., `STORAGE_PROVIDER=git` instead of `STORAGE_PROVIDER=cosmosdb`).
3. **Zero Client Impact:** Because the `ctx` CLI and the Web Portal talk only to the API Gateway's REST endpoints, **clients are completely insulated from this change**. They require no updates, re-installations, or credential changes.

---

## 2. CLI Portability Across AI Agent Clients

The `ctx` CLI is designed to compile as a lightweight standalone binary, making it extremely easy to run across various AI clients:

* **CLI-First Agents (Claude Code, Antigravity CLI):**
  * Sits in the developer's PATH (e.g. `/usr/local/bin/ctx`). 
  * The agent discovers `ctx` automatically via standard shell path checks and executes commands directly (e.g., `ctx read <doc-id>`).
* **GUI-First Editor Agents (Cursor, Windsurf, VS Code Copilot):**
  * We provide a standard **MCP Shim Server**. The shim server is a tiny wrapper (Node.js/Go) that runs locally on the developer's machine and exposes the `ctx` CLI commands as Model Context Protocol tools.
  * This allows Cursor or Claude Desktop to connect via MCP without changing the underlying `ctx` logic.
* **Custom Enterprise Agents (CrewAI, LangChain, Autogen):**
  * Because `ctx` is a standard terminal command, Python and C# agent SDKs can register it instantly using standard `ShellTool` or `CommandLineTool` class wrappers, avoiding custom database drivers in code.
# Specification 06: Security Architecture & Access Control

To safeguard sensitive corporate business logic and schemas from unauthorized access and malicious modification (by either compromised developer systems or rogue AI agents), the system implements the following security posture:

---

## 1. Authentication (AuthN)
All client communication with the API Gateway requires a valid JSON Web Token (JWT) issued by Microsoft Entra ID (formerly Azure AD).
* **JWT Claims Extraction:** The API Gateway decodes the JWT and validates the following claims on every request:
  * `oid` / `sub`: The unique identifier of the calling user/agent.
  * `groups`: The list of Azure AD Security Groups the user belongs to.
* **Token Expiration:** The `ctx` CLI validates local token expiration in `~/.config/ctx/token.json` and automatically triggers a silent refresh if needed.

---

## 2. Authorization & Data Isolation (AuthZ)
We enforce a strict **Attribute-Based Access Control (ABAC)** filter inside the API Gateway middleware:
1. **Write Access Checks:** When a write request (`POST /api/v1/documents/{id}`) is received:
   * The API Gateway extracts the user's `groups` list.
   * It inspects the document payload's `owner_group` metadata field.
   * If the `owner_group` (e.g. `group-a`) is not present in the user's JWT `groups` list, the API aborts transaction execution and returns `403 Forbidden`.
2. **Read Access Checks:** When fetching or searching documents:
   * The API Gateway rewrites database query parameters. Instead of searching globally, it dynamically appends:
     `SELECT * FROM c WHERE c.owner_group IN (<user_groups_list>)`
   * This guarantees that users and agents cannot discover or read document data belonging to other security groups.

---

## 3. Security Boundary: Shielding against Malicious Agent Updates
AI agents have the potential to introduce malicious injections, invalid schemas, or formatted text exploits. The API Gateway mitigates this via three structural filters:
1. **Local and Remote Schema Validation:** The API Gateway validates all inputs against strict JSON Schema (Draft 7) metadata rules. If fields (like `owner_group` or `id`) contain NoSQL syntax or illegal characters, the document is rejected.
2. **Write Immutability (Modified By):** The API Gateway overrides the client's payload `last_modified_by` and `last_updated` fields with security-context claims from the JWT. An agent cannot spoof who performed the write.
3. **Audit History Log:** Every write transaction commits the prior version snapshot to the immutable `document_history` container, allowing security administrators to audit and roll back agent changes instantly.
# Specification 07: Reliability & Fault Tolerance

Since agents rely on context to perform critical development and deployment steps, the context layer must remain resilient to network disconnects, API throttling, and unexpected server failures.

---

## 1. Local Read-Through Cache (Offline Support)
To prevent agents from crashing when developers are working offline (e.g. on airplanes, trains, or during VPN outages):
* **Cache Directory:** The `ucl` CLI maintains a local cache directory at `~/.cache/ucl/`.
* **Write-Through logic:** When `ucl read <doc-id>` runs, the CLI caches the returned document contract.
* **Offline Fallback:** If the API Gateway is unreachable, the CLI attempts to read from the local cache. If found, it prints the document to `stdout` along with a warning log to stderr: 
  `[WARNING] API Gateway unreachable. Serving cached version from YYYY-MM-DD.`
* This allows local developer agents to continue reading guidelines and schemas even when offline.

---

## 2. Defensive CLI Error Handling (Self-Healing Agents)
If an API request fails, the CLI prevents agents from entering infinite retry loops or attempting to write code to "fix" the CLI by returning strict exit codes:
* **Standard Exit Codes:**
  * `0`: Success.
  * `1`: Validation/Schema error (instructs the agent: *your document metadata formatting is wrong*).
  * `2`: Authentication error (instructs the agent: *user needs to run `ucl login`*).
  * `3`: Network / Server error (instructs the agent: *gateway is offline, retry later*).
* **Stack Trace Shielding:** The CLI intercepts raw stack traces or internal DB error logs and outputs brief, human-and-agent readable messages (e.g., `Error: ID 'billing-rules' not found.`). This keeps noisy stack traces out of the agent's context window.

---

## 3. API Gateway Rate-Limit & Backoff Resilience
* **Exponential Backoff:** The `ucl` CLI implements standard exponential backoff with jitter when encountering HTTP `429 Too Many Requests` or `503 Service Unavailable` statuses, mitigating client stampedes when databases scale up.
# Specification 08: Technical Specifications & Endpoints

This document establishes the technical layout parameters for Azure Cosmos DB and the HTTP API Gateway routing contracts.

---

## 1. Cosmos DB Partitioning Key & Structure

* **Database Name:** `unified_context`
* **Primary Container Name:** `documents` (Stores active document contracts).
  * **Partition Key:** `/owner_group`
  * **Secondary Indexes:** Range index on `/category` and `/tags`.
* **History Container Name:** `document_history` (Stores historic snapshots for auditability).
  * **Partition Key:** `/document_id`
  * **Why `/document_id`?** Audit trails are queried on a per-document basis. Grouping history by `document_id` ensures that retrieving a file's log is a single-partition query.
  * **History Storage Model:** To avoid complex diff algorithms, each update to a document in the `documents` container triggers a transaction that appends a complete copy of the previous document state to the `document_history` container, labeled with `version` and `modified_by`.

---

## 2. Dynamic Schema Templates (Initial Scope)

The API Gateway endpoint `/api/v1/schemas/{type}` serves pre-configured Markdown content templates for developers and agents. The initial set of templates includes:

### A. `adr` (Architecture Decision Record)
```markdown
# ADR: [Decision Title]

* **Status:** [Proposed | Accepted | Superseded]
* **Date:** YYYY-MM-DD
* **Deciders:** [Name, Name]

## 1. Context
[What is the context, problem statement, and forces at play?]

## 2. Decision
[What is the selected option and why? What were the alternatives evaluated?]

## 3. Consequences
[What is the impact of this choice? What are the new technical debts or changes?]
```

### B. `product-spec` (Product Requirements Document)
```markdown
# Product Spec: [Feature Title]

* **Category:** Product Requirements
* **Target Audience:** [e.g. End Users, Internal Ops]

## 1. Overview & Objective
[What problem does this solve and what are the key business metrics we expect to move?]

## 2. User Stories & Workflows
- **As a** [role], **I want to** [action], **so that** [benefit].

## 3. Scope & Exclusions
- **In-Scope:** [Detail features]
- **Out-of-Scope:** [Detail postponed items]
```

### C. `guide` (Onboarding, Runbooks, Tasks)
```markdown
# Guide: [Title / Topic]

* **Prerequisites:** [e.g., Azure CLI, Node v18]

## 1. Setup Instructions
[Step-by-step setup guides]

## 2. Verification Steps
[How to run tests or checks to verify the setup worked]

## 3. Common Troubleshooting Issues
- **Issue:** [Problem description]
  * **Fix:** [Resolution step]
```

### D. `tech-plan` (Tech Specs / Implementation Proposals)
```markdown
# Tech Plan: [System / Feature Name]

* **Author:** [Name]
* **Dependencies:** [doc-id-1, doc-id-2]

## 1. Problem Statement & Scope
[Briefly define what needs to be engineered]

## 2. Proposed Architecture & System Design
[Data schemas, flowcharts, class structures, endpoints]

## 3. Implementation Phases
* **Phase 1:** [Boilerplate/database]
* **Phase 2:** [APIs/CLI validation]

## 4. Rollout & Rollback Strategy
[How do we release this safely? What is the rollback trigger?]
```

### E. `api-contract` (API Endpoint Contracts)
```markdown
# API Contract: [Service Name / Endpoint Path]

* **Base URL:** `https://api.company.com/v1`
* **Authentication:** [e.g. JWT Bearer token, Admin scope]

## 1. Endpoint: [METHOD] `/path`
[Description of what the endpoint does]

### Request Payload (JSON Schema)
```json
{
  "param": "type"
}
```

### Response Codes
* **200 OK:** [Success payload structure]
* **400 Bad Request:** [Validation error response]
* **401 Unauthorized:** [Auth error response]
```

### F. `db-schema` (Database Layout & Indices)
```markdown
# DB Schema: [Database / Table Name]

* **Database Engine:** [e.g. PostgreSQL, Cosmos DB]
* **Container/Table Name:** `[name]`
* **Partition/Primary Key:** `[key]`

## 1. Fields & Columns
| Field Name | Data Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | Primary Key | Unique ID |

## 2. Indexes & Foreign Keys
- **Primary Index:** `/owner_group`
- **Foreign Key:** `user_id` references `users(id)`
```

### G. `env-spec` (Environment & Deployment Settings)
```markdown
# Env Spec: [Service / Repository Name]

* **Applicable Environments:** [Development | Staging | Production]

## 1. Variables Definition
| Variable Name | Required | Default Value | Description |
| :--- | :--- | :--- | :--- |
| `DATABASE_URL` | Yes | N/A | Connection string for Postgres |
| `PORT` | No | `8080` | Web server listening port |

## 2. Secrets Management
[Where are these secrets stored? Key Vault, Doppler, GitHub Secrets?]
```

### H. `freeform` (General Context)
```markdown
# [Context Title]

* **Tags:** [e.g. legacy, operational-tips]

[Freeform markdown text. Record any vital engineering context or tribal knowledge here.]
```

---

## 3. API Gateway Interface Specs (REST OpenAPI)

The API Gateway mediates requests from both the Web Portal and the `ucl` CLI:

* `GET /api/v1/documents`
  * **Query Parameters:** `category` (string, optional), `tags` (array, optional)
  * **Headers:** `Authorization: Bearer <JWT>`
  * **Returns:** Array of document metadata JSONs (excluding the heavy `content` field).
* `GET /api/v1/documents/{id}`
  * **Headers:** `Authorization: Bearer <JWT>`
  * **Returns:** Full JSON contract object including the payload `content`.
* `POST /api/v1/documents/{id}`
  * **Headers:** `Authorization: Bearer <JWT>`
  * **Body:** JSON Contract payload.
  * **Validation:** Enforces JSON Schema (Draft 7). Increments `version`, updates audit fields (`last_modified_by`, `last_updated`). Writes previous document snapshot to `document_history` container.
* `GET /api/v1/search`
  * **Query Parameters:** `q` (string, search query)
  * **Returns:** Array of ranked document matches (IDs and snippets).
* `GET /api/v1/schemas/{type}`
  * **Returns:** The template schema and structural guidelines for the requested document `type` (e.g. `adr`). Used by `ucl schema <type>` for offline/online advisory rules.
* `GET /llms.txt`
  * **Returns:** Plain-text sitemap summarizing all accessible metadata contracts for the user's active session.
# Specification 09: Implementation Milestones & Tickets

This document serves as the project backlog blueprint. Development teams can use these specs to cut JIRA or GitHub tickets directly.

---

## Milestone 1: Database & API Gateway Provisioning (Epic: UCL-DB)

* **Ticket UCL-101: Provision Cosmos DB Containers**
  * *Description:* Set up the `unified_context` database. Create `documents` container (partition key `/owner_group`) and `document_history` container (partition key `/document_id`).
* **Ticket UCL-102: Setup API Gateway Boilerplate & AD Auth**
  * *Description:* Scaffold the Go or Python API server. Integrate Azure AD / OAuth 2.0 middleware. Validate JWT tokens and map user groups to database access controls.
* **Ticket UCL-103: Implement Document CRUD Endpoints & History Log**
  * *Description:* Build endpoints for document retrieval, updates, and creation. Implement JSON validation middleware against the Document Contract schema and automated snapshotting into `document_history`.
* **Ticket UCL-104: Design Swappable Storage Interface Layer**
  * *Description:* Create the `StorageProvider` abstraction interface. Implement the `CosmosDBProvider` concrete driver as the default active storage adapter.
* **Ticket UCL-105: Implement API Gateway Access Control Middleware**
  * *Description:* Code the ABAC security middleware verifying user's group claims against document `owner_group` properties, and query rewriting to isolate reads.
* **Ticket UCL-106: Implement Schema Template API Endpoints**
  * *Description:* Create the `GET /api/v1/schemas/{type}` endpoints. Store and serve pre-defined markdown and metadata templates (like ADR, Schema contracts) to aid client tools.

---

## Milestone 2: The `ucl` CLI Query Engine (Epic: UCL-CLI)

* **Ticket UCL-201: Scaffold CLI Application & Auth Client**
  * *Description:* Build the CLI CLI scaffolding in Go/Python. Integrate local session caching (save JWT token at `~/.config/ucl/token.json` after running `ucl login` via Azure CLI token exchange).
* **Ticket UCL-202: Implement `ucl search` and `ucl read` Commands**
  * *Description:* Add CLI subcommands to query the API gateway `/api/v1/search` and print raw document content to `stdout`.
* **Ticket UCL-203: Implement `ucl write` with Schema Validation**
  * *Description:* Build `ucl write` subcommand. Ensure the CLI validates metadata schemas locally before submitting the JSON package to the API endpoint.
* **Ticket UCL-204: Implement Local Offline Read Cache**
  * *Description:* Build CLI offline fallback logic reading from and writing to `~/.cache/ucl/` on read queries.
* **Ticket UCL-205: Implement Defensive Exit Coding & Trace shielding**
  * *Description:* Write CLI exception handlers returning strict code groups (1, 2, 3) and formatted CLI-friendly errors to protect agent contexts.
* **Ticket UCL-206: Implement `ucl schema` Template Retrieval**
  * *Description:* Add the `ucl schema <type>` subcommand. Query `/api/v1/schemas/{type}` to print best-practice templates (e.g. for `adr`) directly to standard output to guide agents.

---

## Milestone 3: Dumb Web Portal & AI Sitemap (Epic: UCL-UI)

* **Ticket UCL-301: Develop Go+HTMX / Python Web Dashboard**
  * *Description:* Build a simple responsive page. Include a sidebar listing categories/tags and a reader view.
* **Ticket UCL-302: Build Dynamic Markdown and JSON Renderer**
  * *Description:* Build client/server-side rendering. Compile Markdown into HTML for document views, and syntax-highlight raw JSON files.
* **Ticket UCL-303: Expose dynamic `/llms.txt` Endpoint**
  * *Description:* Write a route handler for `/llms.txt` that pulls active documents, compiles a clean Markdown index, and outputs it as text/plain.
# Specification 10: Product Mapping & Repository Structure

This document outlines the product names, repository topology, and codebase directory layout for the Unified Context Layer (UCL) project suite.

---

## 1. Product Naming & Components

Due to naming collisions with existing open-source agent projects on GitHub (such as `ctxrs/ctx` and `Alegau03/CTX`), we name this ecosystem **Unified Context Layer (UCL)**. The CLI tool and its supporting modules use the prefix **`ucl`** to guarantee namespace safety and prevent command overlap:

| Component | Recommended Product Name | Role / Function | Distribution Format |
| :--- | :--- | :--- | :--- |
| **CLI Tool** | **`ucl`** | Developer & Agent command-line interface. Enforces JSON metadata schemas and executes DB queries. | Standalone Compiled Binary (e.g. via Go or PyInstaller) |
| **API Gateway** | **`ucl-gateway`** | Secure REST proxy mediating between Cosmos DB and clients. Validates Entra ID JWTs and ABAC group permissions. | Docker Container / Azure App Service |
| **Web Portal** | **`ucl-portal`** | Human-facing navigable Wiki (Go+HTMX or Python) showing document hierarchies, markdown, and rendering `/llms.txt`. | Docker Container (often co-located/embedded in the gateway) |
| **MCP Shim** | **`ucl-mcp`** | A lightweight local wrapper server exposing `ucl` CLI commands as Model Context Protocol (MCP) tools for GUI editors. | NPM package or binary wrapper |

---

## 2. Repository Topology: The Monorepo Choice

We **strongly recommend a Monorepo** (named `ucl`) rather than separate repositories.

### Why a Monorepo?
1. **Shared Code Contracts:** The CLI, Gateway, Portal, and MCP Shim must validate the exact same JSON Contract Schema (Specification 01). A monorepo allows sharing validation schemas and structs directly without publishing packages to private registries.
2. **Atomic Versioning:** Changes to the metadata schema (e.g., adding a new document category type) require simultaneous changes to CLI validation, API Gateway endpoints, and the Web Portal UI. In a monorepo, these are committed in a single Pull Request.
3. **Simplified Local Testing:** Developers can spin up the Gateway, Database emulator, and CLI client locally using a single docker-compose or build command.

---

## 3. Directory Layout Blueprint

The proposed structure for the unified **`ucl`** monorepo:

```
ucl/                              <-- Monorepo Root
├── cli/                          <-- `ucl` CLI source code (Go or Python)
│   ├── cmd/                      <-- Entry points (login, search, read, write)
│   └── config/                   <-- Local cache parser (~/.config/ucl/token.json)
│
├── gateway/                      <-- `ucl-gateway` API Server
│   ├── auth/                     <-- Entra ID JWT validation middleware
│   ├── handler/                  <-- HTTP handlers (search, CRUD, sitemap)
│   └── storage/                  <-- StorageProvider implementation drivers
│
├── portal/                       <-- `ucl-portal` web wiki UI (Go+HTMX or Python)
│   ├── templates/                <-- Server-rendered HTML templates
│   └── static/                   <-- Light CSS assets
│
├── mcp/                          <-- `ucl-mcp` wrapper code
│
├── shared/                       <-- Shared schemas, JSON-schemas, types
│   └── contract.json             <-- The central JSON Schema definition
│
├── docs/                         <-- Project Documentation
│   └── specs/                    <-- This modular specifications wiki
│       └── index.md
│
├── go.work / package.json        <-- Workspace configuration
└── docker-compose.yml            <-- Local development environment definition
```
