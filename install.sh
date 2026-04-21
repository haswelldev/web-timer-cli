#!/usr/bin/env sh
set -e

REPO="haswelldev/web-timer-cli"
INSTALL_DIR="/usr/local/bin"
BIN_NAME="web-timer-cli"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  GOOS="linux" ;;
  darwin) GOOS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) GOARCH="amd64" ;;
  arm64|aarch64) GOARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

ASSET="${BIN_NAME}-${GOOS}-${GOARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading ${ASSET}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "/tmp/${BIN_NAME}"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "/tmp/${BIN_NAME}" "$URL"
else
  echo "Error: curl or wget is required"
  exit 1
fi

chmod +x "/tmp/${BIN_NAME}"

# Install (may need sudo)
if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (sudo required)"
  sudo mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
fi

echo "Installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"
echo "Run: ${BIN_NAME} [room-id-or-url]"

# Uninstall instructions
echo ""
echo "To uninstall:"
echo "  sudo rm ${INSTALL_DIR}/${BIN_NAME}"
