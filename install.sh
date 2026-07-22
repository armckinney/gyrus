#!/usr/bin/env bash
set -e

# Gyrus Installer Script
# Usage: curl -sSL https://raw.githubusercontent.com/armckinney/gyrus/main/install.sh | bash

REPO="armckinney/gyrus"
BINARY="gyrus"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: ${ARCH}"
    exit 1
    ;;
esac

case "${OS}" in
  linux)
    OS="linux"
    FORMAT="tar.gz"
    ;;
  darwin)
    OS="darwin"
    FORMAT="tar.gz"
    ;;
  mingw*|msys*|cygwin*)
    OS="windows"
    FORMAT="zip"
    ;;
  *)
    echo "Unsupported operating system: ${OS}"
    exit 1
    ;;
esac

echo "Detecting latest release for ${REPO}..."
LATEST_RELEASE=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_RELEASE}" ]; then
  echo "Error: Unable to fetch latest release version from GitHub API."
  exit 1
fi

VERSION="${LATEST_RELEASE#v}"
FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.${FORMAT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${FILENAME}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading ${BINARY} v${VERSION} for ${OS}/${ARCH}..."
curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${FILENAME}"

cd "${TMP_DIR}"
if [ "${FORMAT}" = "tar.gz" ]; then
  tar -xzf "${FILENAME}"
else
  unzip -q "${FILENAME}"
fi

INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "${INSTALL_DIR}"
fi

echo "Installing ${BINARY} to ${INSTALL_DIR}..."
mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "✓ Successfully installed ${BINARY} v${VERSION} to ${INSTALL_DIR}/${BINARY}!"
echo "Run '${BINARY} --help' to get started."
