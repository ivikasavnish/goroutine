#!/bin/bash
set -e

# Portless installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | sh

REPO="ivikasavnish/goroutine"
BINARY_NAME="portless"
INSTALL_DIR="/usr/local/bin"
VERSION="${VERSION:-latest}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case $OS in
    darwin|linux)
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

PLATFORM="${OS}-${ARCH}"

echo "Installing portless for $PLATFORM..."

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Determine download URL
if [ "$VERSION" = "latest" ]; then
    # For now, we'll build from source since releases aren't set up yet
    echo "Downloading source code..."
    
    # Check if go is installed
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed"
        echo "Please install Go from https://golang.org/dl/ and try again"
        exit 1
    fi
    
    # Check if git is installed
    if ! command -v git &> /dev/null; then
        echo "Error: Git is not installed"
        echo "Please install Git and try again"
        exit 1
    fi
    
    # Clone the repository
    git clone --depth 1 https://github.com/${REPO}.git
    cd goroutine
    
    # Build the binary
    echo "Building portless..."
    go build -o "$BINARY_NAME" ./cmd/portless
    
else
    # Download from releases (future enhancement)
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    echo "Downloading from $DOWNLOAD_URL..."
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$BINARY_NAME"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$BINARY_NAME"
    else
        echo "Error: Neither curl nor wget is available"
        exit 1
    fi
fi

# Make binary executable
chmod +x "$BINARY_NAME"

# Install to system
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Cleanup
cd /
rm -rf "$TMP_DIR"

echo ""
echo "✓ Portless installed successfully!"
echo ""
echo "Get started with:"
echo "  portless proxy start          # Start the proxy server"
echo "  portless myapp npm start       # Run your app with named URL"
echo ""
echo "Access your apps at http://<name>.localhost:1355"
echo ""
