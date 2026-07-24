#!/usr/bin/env bash
# Gyrus Agent Skill Installer Script
# Usage: bash install-skill.sh [version]
# Examples:
#   bash install-skill.sh          (Installs latest release)
#   bash install-skill.sh v0.1.0   (Installs specific pinned version)

set -euo pipefail

TARGET_DIR=".agents/skills/gyrus"
REPO="armckinney/gyrus"
VERSION="${1:-${GYRUS_SKILL_VERSION:-latest}}"

echo "=> Installing Gyrus Agent Skill (${VERSION}) into ${TARGET_DIR}..."

# Create target directory
mkdir -p .agents/skills

if [ "${VERSION}" = "latest" ]; then
  TARBALL_URL="https://github.com/${REPO}/releases/latest/download/gyrus-skill.tar.gz"
else
  TARBALL_URL="https://github.com/${REPO}/releases/download/${VERSION}/gyrus-skill_${VERSION}.tar.gz"
fi

if curl -sI "${TARBALL_URL}" | grep -qE "200 OK|302 Found|301 Moved"; then
  echo "=> Downloading ${TARBALL_URL}..."
  curl -sSL "${TARBALL_URL}" | tar -xz -C .agents/skills/
else
  echo "=> Download URL not available. Falling back to downloading main branch skill..."
  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT
  curl -sSL "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" | tar -xz -C "$TMP_DIR"
  mkdir -p "${TARGET_DIR}"
  cp -r "$TMP_DIR/gyrus-main/skills/gyrus/"* "${TARGET_DIR}/"
fi

chmod +x "${TARGET_DIR}/scripts/"*.sh 2>/dev/null || true

echo "=> Gyrus Agent Skill successfully installed at ${TARGET_DIR}!"
