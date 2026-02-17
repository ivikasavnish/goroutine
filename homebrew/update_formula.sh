#!/bin/bash
# Script to update Homebrew formula with actual release checksums
# Usage: ./homebrew/update_formula.sh <version>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1
REPO="ivikasavnish/goroutine"
FORMULA_FILE="homebrew/portless.rb"

echo "Updating Homebrew formula for version $VERSION..."

# Download checksums.txt from the release
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
echo "Downloading checksums from $CHECKSUMS_URL..."

curl -fsSL "$CHECKSUMS_URL" -o /tmp/checksums.txt

# Extract checksums for each platform
DARWIN_AMD64_SHA=$(grep "portless-darwin-amd64.tar.gz" /tmp/checksums.txt | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep "portless-darwin-arm64.tar.gz" /tmp/checksums.txt | awk '{print $1}')
LINUX_AMD64_SHA=$(grep "portless-linux-amd64.tar.gz" /tmp/checksums.txt | awk '{print $1}')
LINUX_ARM64_SHA=$(grep "portless-linux-arm64.tar.gz" /tmp/checksums.txt | awk '{print $1}')

echo "Checksums:"
echo "  darwin-amd64: $DARWIN_AMD64_SHA"
echo "  darwin-arm64: $DARWIN_ARM64_SHA"
echo "  linux-amd64:  $LINUX_AMD64_SHA"
echo "  linux-arm64:  $LINUX_ARM64_SHA"

# Create a new formula file with updated values
cat > "$FORMULA_FILE" << EOF
class Portless < Formula
  desc "Named localhost URLs for development servers with ngrok integration"
  homepage "https://github.com/ivikasavnish/goroutine"
  version "${VERSION#v}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/${VERSION}/portless-darwin-arm64.tar.gz"
      sha256 "${DARWIN_ARM64_SHA}"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/${VERSION}/portless-darwin-amd64.tar.gz"
      sha256 "${DARWIN_AMD64_SHA}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/${VERSION}/portless-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM64_SHA}"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/${VERSION}/portless-linux-amd64.tar.gz"
      sha256 "${LINUX_AMD64_SHA}"
    end
  end

  def install
    bin.install "portless-darwin-amd64" => "portless" if OS.mac? && Hardware::CPU.intel?
    bin.install "portless-darwin-arm64" => "portless" if OS.mac? && Hardware::CPU.arm?
    bin.install "portless-linux-amd64" => "portless" if OS.linux? && Hardware::CPU.intel?
    bin.install "portless-linux-arm64" => "portless" if OS.linux? && Hardware::CPU.arm?
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/portless version")
  end
end
EOF

echo "✓ Formula updated successfully!"
echo ""
echo "To install locally for testing:"
echo "  brew install --build-from-source ${FORMULA_FILE}"
echo ""
echo "To add to a Homebrew tap:"
echo "  1. Create a tap repository: homebrew-tap"
echo "  2. Add this formula to: Formula/portless.rb"
echo "  3. Users can then install with: brew install ivikasavnish/tap/portless"
