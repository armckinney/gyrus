#!/usr/bin/env bash
# Gyrus Agent Skill Installation Helper
# Checks if `gyrus` CLI is available in PATH. If missing, attempts installation.

set -euo pipefail

if command -v gyrus >/dev/null 2>&1; then
  echo "=> gyrus CLI is already installed at: $(command -v gyrus)"
  gyrus help | head -n 3 || true
  exit 0
fi

echo "=> gyrus CLI not found in PATH. Checking for workspace binary..."

if [ -f "./gyrus" ]; then
  echo "=> Local binary ./gyrus found in workspace."
  ./gyrus help | head -n 3 || true
  exit 0
fi

echo "=> Attempting automated installation..."

if command -v go >/dev/null 2>&1; then
  echo "=> Building gyrus from source via go install..."
  go install github.com/armckinney/gyrus/cmd/gyrus@latest
  echo "=> Successfully installed gyrus via go install!"
else
  echo "=> Downloading gyrus installer script..."
  curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash
fi
