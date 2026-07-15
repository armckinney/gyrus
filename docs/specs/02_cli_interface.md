# Specification 02: CLI Interface & Authentication

The CLI tool (`gyrus`) is a compiled standalone Go binary that serves as the execution gateway. It handles local caching, authenticates the developer, and communicates directly with the Core Memory Engine.

---

## 1. CLI Command reference

* **`gyrus init`**
  * *Behavior:* Bootstraps the current folder with a `.gyrus/` directory, default configurations, and local JSON schemas.
* **`gyrus get --id <doc-id> [--json]`**
  * *Behavior:* Reads the document from storage. Prints the Markdown body to `stdout` (or the complete JSON envelope if `--json` is set).
* **`gyrus create --type <type> --input <json-file> [--json]`**
  * *Behavior:* Passes payload to the Core Engine. Validates the schema locally. Returns validation status and the document ID.
* **`gyrus update --id <doc-id> --input <patch-file> --expected-version <version> --reason "<text>"`**
  * *Behavior:* Performs an optimistic lock check. If versions match, applies patch modifications, increments version, and updates database index.
* **`gyrus validate [--all | --input <file>]`**
  * *Behavior:* Evaluates document metadata contracts and checks lifecycle state transitions, outputting syntax warnings or errors.
* **`gyrus search --query "<search-string>"`**
  * *Behavior:* Runs FTS (Full-Text Search) over local SQLite tables and outputs a list of matching document IDs.
* **`gyrus suggest-context --request "<prompt>" [--max-results <num>]`**
  * *Behavior:* Performs semantic filtering based on keywords and edge tags to retrieve context packages relevant to the target task.
* **`gyrus schema <type>`**
  * *Behavior:* Retrieves and prints the recommended structural markdown template for the requested type (e.g. `gyrus schema adr`).
* **`gyrus mcp serve`**
  * *Behavior:* Launches the local Model Context Protocol (MCP) server over standard input/output (`stdio`) for AI client integration.

---

## 2. Authentication & Local Caching

When configured in **Team/Enterprise (Cloud) Profile**, `gyrus` utilizes Azure AD / Entra ID for access:
1. **Azure CLI Handshake:** `gyrus` executes `az account get-access-token --resource <API_CLIENT_ID>` to retrieve a JWT token.
2. **Device Code Fallback:** If Azure CLI is missing, the CLI requests a code from Entra ID and prints it:
   `Open https://microsoft.com/devicelogin and enter code: ABCD-EFGH`
3. **Session Cache:** Once retrieved, the CLI saves the token to `~/.config/gyrus/token.json` along with its expiration timestamp.
4. **Offline Cache:** Caches retrieved documents to `~/.cache/gyrus/` for offline read operations.
