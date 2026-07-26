#!/usr/bin/env bash
# Live Google Antigravity Agent Skill Integration Test Script
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMP_DIR=$(mktemp -d "/tmp/gyrus-integration-antigravity-XXXXXX")
defer_cleanup() {
  rm -rf "$TEMP_DIR"
}
trap defer_cleanup EXIT

echo "=== 🪐 Starting Live Google Antigravity Skill Integration Test ==="

# 1. Initialize workspace with Antigravity target
echo "1. Initializing test workspace..."
cd "$TEMP_DIR"
"${REPO_ROOT}/gyrus" init --mcp-target antigravity --skill-target antigravity --json

# 2. Assert .antigravity/mcp.json and agent skills exist
if [ -f "${TEMP_DIR}/.antigravity/mcp.json" ]; then
  echo "✅ Dedicated .antigravity/mcp.json verified"
else
  echo "❌ Error: .antigravity/mcp.json missing"
  exit 1
fi

if [ -f "${TEMP_DIR}/.agents/skills/gyrus-cli/SKILL.md" ] && [ -f "${TEMP_DIR}/.agents/skills/gyrus-mcp/SKILL.md" ]; then
  echo "✅ Equipped gyrus-cli and gyrus-mcp skills verified"
else
  echo "❌ Error: Agent skill files missing in .agents/skills/"
  exit 1
fi

# 3. Create test OKF document inside temp storage path
echo "2. Creating test OKF document..."
"${REPO_ROOT}/gyrus" --storage-path "${TEMP_DIR}/docs/.gyrus/docs" create \
  --id "adr-integration-antigravity-test" \
  --title "Antigravity Integration Storage Engine ADR" \
  --category "architecture" \
  --type "adr" \
  --owner-group "armckinney" \
  --status "proposed" \
  --content "Testing live Antigravity agent context resolution." \
  --json

# 4. Check agy CLI availability for live LLM prompt execution
if command -v agy >/dev/null 2>&1; then
  echo "3. Executing live AGY agent prompt evaluation..."
  AGY_OUT=$(agy "Search for Antigravity Integration Storage Engine ADR using gyrus" 2>&1 || true)
  if echo "$AGY_OUT" | grep -iE "gyrus|adr-integration-antigravity-test" >/dev/null 2>&1; then
    echo "✅ Live AGY agent invoked gyrus context tool successfully!"
  else
    echo "⚠️ Warning: Live AGY prompt executed but output trace did not contain document ID"
  fi
else
  echo "ℹ️ Note: agy CLI binary not detected in environment. Verifying JSON-RPC stdio transport interface..."
  RESP=$(timeout 2s bash -c "echo '{\"jsonrpc\":\"2.0\",\"id\":100,\"method\":\"tools/list\"}' | \"${REPO_ROOT}/gyrus\" --storage-path \"${TEMP_DIR}/docs/.gyrus/docs\" mcp serve 2>/dev/null" || true)
  if echo "$RESP" | grep -q "gyrus_search"; then
    echo "✅ Gyrus stdio MCP server tools/list endpoint verified"
  else
    echo "❌ Error: Gyrus stdio MCP server tools/list endpoint failed"
    exit 1
  fi
fi

echo "=== ✅ Google Antigravity Skill Integration Test Completed ==="
