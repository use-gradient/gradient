#!/usr/bin/env sh
# Install gradient CLI (curl -fsSL https://... | sh)
# Placeholder: set GRADIENT_INSTALL_URL when hosting releases.

set -e

GITHUB_REPO="${GRADIENT_REPO:-usegradient/gradient}"
# Placeholder until releases are hosted; user can override with GRADIENT_INSTALL_URL
BASE_URL="${GRADIENT_INSTALL_URL:-https://github.com/${GITHUB_REPO}/releases/latest/download}"

OS=$(uname -s)
ARCH=$(uname -m)
case "$OS" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac
case "$ARCH" in
  x86_64)  ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

FILENAME="gradient_${OS}_${ARCH}"
URL="${BASE_URL}/${FILENAME}"
BIN_DEST="${GRADIENT_BIN:-/usr/local/bin/gradient}"

echo "Downloading gradient from $URL"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$FILENAME" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$FILENAME" "$URL"
else
  echo "Need curl or wget to download"; exit 1
fi

chmod +x "$FILENAME"
mkdir -p "$(dirname "$BIN_DEST")"
mv "$FILENAME" "$BIN_DEST"
echo "Installed gradient to $BIN_DEST"
echo "Run 'gradient auth login' to set your API key."
