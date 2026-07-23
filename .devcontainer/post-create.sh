#!/usr/bin/env bash
# .devcontainer/post-create.sh
set -euo pipefail

# Ensure target directories exist
mkdir -p .agents
mkdir -p .github

echo "Initializing AI agent rule symlinks pointing to docs/agents/AGENTS.md..."

# Create symbolic links relative to their respective directories
ln -sf ../docs/agents/AGENTS.md .agents/AGENTS.md
ln -sf ../docs/agents/AGENTS.md .github/copilot-instructions.md

echo "AI agent symlinks configured successfully!"
