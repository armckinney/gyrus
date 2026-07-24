#!/usr/bin/env bash
# Gyrus Repository Bootstrapper Script
# Initializes .gyrus.yaml configuration and initial OKF directory structure.

set -euo pipefail

GYRUS_CMD="gyrus"
if [ -f "./gyrus" ]; then
  GYRUS_CMD="./gyrus"
fi

echo "=> Initializing Gyrus storage & configuration..."

# Create .gyrus.yaml if missing
if [ ! -f ".gyrus.yaml" ]; then
  cat <<'EOF' > .gyrus.yaml
storage_provider: localfs
index_provider: sqlite
storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: root
EOF
  echo "=> Created default .gyrus.yaml"
fi

# Create default directories
mkdir -p docs/.gyrus/docs/okf/root/reference
mkdir -p docs/.gyrus/schemas

# Run sync
$GYRUS_CMD sync

echo "=> Gyrus repository setup complete!"
