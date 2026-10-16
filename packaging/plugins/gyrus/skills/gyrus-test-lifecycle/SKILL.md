---
name: gyrus-test-lifecycle
description: Perform agent-led functional acceptance testing of an installed Gyrus release through public CLI commands, MCP tools, and the Web UI. Use for lifecycle verification and installation troubleshooting without source code or development tooling.
---

# Test an installed Gyrus lifecycle

Follow [the consumer test procedure](references/consumer-lifecycle.md), running commands and inspecting responses yourself. Use the installed `gyrus` executable and available terminal, file-editing, MCP-client, and browser/HTTP tools. There is no bundled runner or requirement for Perl, Python, Node, Go, Make, repository source, or development tests. Do not build, upgrade, or replace the executable during verification.

Create a private temporary workspace, home, and localfs/OKF configuration. Isolate inherited GYRUS environment variables. Start only your own MCP/UI services against that configuration; do not use the user's connected production MCP server or live collection. Record fixture names and owned process handles for cleanup.

Discover syntax with installed command help and MCP tool schemas. Follow the procedure's assertions; a successful command alone is not proof of the expected behavior. If a required surface or client capability is unavailable, report it as blocked and the overall run as incomplete. Do not substitute source tests or silently skip assertions. Verify the localfs/OKF profile only; remote providers require separate tests.

Keep a report outside the disposable workspace with executable/version when available, resolved providers, each check's expected/actual result, exit status or MCP/HTTP error, failures/blockers, and artifact locations. Report full success only when all checks pass. Stop owned services in all outcomes. On success archive owned concepts and remove only your workspace/home; on failure preserve fixtures and logs for diagnosis. Never weaken an assertion to obtain a pass.
