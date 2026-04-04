#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

BIN_NAME="postmanify"
REPO_URL="https://github.com/hieudoanm/postmanify/raw/refs/heads/master/packages/cli/bin/postmanify"
INSTALL_DIR="/usr/local/bin"

echo "Downloading $BIN_NAME..."
curl -fsSL "$REPO_URL" -o "$BIN_NAME"

echo "Making $BIN_NAME executable..."
chmod +x "$BIN_NAME"

echo "Installing $BIN_NAME to $INSTALL_DIR..."
# Move binary to install directory (might require sudo)
if [ -w "$INSTALL_DIR" ]; then
    mv "$BIN_NAME" "$INSTALL_DIR/"
else
    echo "Privilege escalation required to install to $INSTALL_DIR. Prompting for sudo password..."
    sudo mv "$BIN_NAME" "$INSTALL_DIR/"
fi

echo ""
echo "✅ $BIN_NAME installed successfully to $INSTALL_DIR/$BIN_NAME"
echo "You can now run '$BIN_NAME version' to verify the installation."
