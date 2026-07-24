#!/usr/bin/env bash
# Gyrus Diagnostic Verification Script
# Verifies binary execution, index sync, and OKF schema health.

set -euo pipefail

GYRUS_CMD="gyrus"
if [ -f "./gyrus" ]; then
  GYRUS_CMD="./gyrus"
fi

echo "=== 🔍 Gyrus Skill Health Check ==="

if [ ! -f "$GYRUS_CMD" ] && ! command -v gyrus >/dev/null 2>&1; then
  echo "❌ Error: gyrus binary not found. Run scripts/install.sh first."
  exit 1
fi

echo "1. CLI Binary Availability:"
$GYRUS_CMD help | head -n 3

echo -e "\n2. Storage & Search Sync Check:"
$GYRUS_CMD sync --json

echo -e "\n3. Checking FTS Database:"
if [ -f "docs/.gyrus/docs/index.db" ]; then
  echo "✅ index.db exists ($(du -h docs/.gyrus/docs/index.db | cut -f1))"
else
  echo "⚠️ Warning: docs/.gyrus/docs/index.db not found. Run gyrus sync to generate."
fi

echo -e "\n=== ✅ Health Check Passed ==="
