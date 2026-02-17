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

# Determine download URL and strategy
if [ "$VERSION" = "latest" ]; then
    # Try to get the latest release from GitHub API
    echo "Fetching latest release..."
    
    if command -v curl &> /dev/null; then
        LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    
    if [ -n "$LATEST_VERSION" ]; then
        VERSION="$LATEST_VERSION"
        echo "Found latest version: $VERSION"
    else
        echo "Could not fetch latest release. Will build from source."
        VERSION="source"
    fi
fi

# Download or build
if [ "$VERSION" = "source" ]; then
    echo "Building from source..."
    
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
    # Download from releases
    ARCHIVE_NAME="${BINARY_NAME}-${PLATFORM}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    echo "Downloading from $DOWNLOAD_URL..."
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$ARCHIVE_NAME"
    else
        echo "Error: Neither curl nor wget is available"
        exit 1
    fi
    
    # Extract the archive
    echo "Extracting..."
    tar xzf "$ARCHIVE_NAME"
    
    # Find the binary (it will be named with platform)
    BINARY_FILE="${BINARY_NAME}-${PLATFORM}"
    if [ ! -f "$BINARY_FILE" ]; then
        echo "Error: Binary $BINARY_FILE not found in archive"
        exit 1
    fi
    
    # Rename to standard name
    mv "$BINARY_FILE" "$BINARY_NAME"
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
