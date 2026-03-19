#!/usr/bin/env bash
# install.sh — download and install the latest gateway-cli binary
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/VincentK1991/mcp-gateway-cli/main/install.sh | bash
#
# Options (env vars):
#   INSTALL_DIR   — install location (default: /usr/local/bin)
#   VERSION       — specific version tag to install (default: latest)

set -euo pipefail

REPO="VincentK1991/mcp-gateway-cli"
BINARY="gateway-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── detect OS ────────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# ── detect architecture ───────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# ── resolve version ───────────────────────────────────────────────────────────
if [ -z "${VERSION:-}" ]; then
  echo "Fetching latest release tag..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version." >&2
  exit 1
fi

echo "Installing gateway-cli ${VERSION} (${OS}/${ARCH}) → ${INSTALL_DIR}"

# ── download ──────────────────────────────────────────────────────────────────
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "${BASE_URL}/${ARCHIVE}"       -o "${TMP}/${ARCHIVE}"
curl -fsSL "${BASE_URL}/checksums.txt"    -o "${TMP}/checksums.txt"

# ── verify checksum ───────────────────────────────────────────────────────────
cd "$TMP"
if command -v sha256sum &>/dev/null; then
  grep "${ARCHIVE}" checksums.txt | sha256sum --check --status
elif command -v shasum &>/dev/null; then
  grep "${ARCHIVE}" checksums.txt | shasum -a 256 --check --status
else
  echo "Warning: no sha256 tool found, skipping checksum verification." >&2
fi
cd - >/dev/null

# ── extract and install ───────────────────────────────────────────────────────
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  sudo chmod +x "${INSTALL_DIR}/${BINARY}"
else
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "gateway-cli ${VERSION} installed successfully!"
echo "Run: gateway-cli --help"
