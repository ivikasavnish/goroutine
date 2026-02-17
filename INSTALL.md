# Quick Start Guide - Portless Installation

This guide provides quick installation instructions for Portless on macOS and Linux.

## Supported Platforms

✅ macOS (Intel and Apple Silicon)  
✅ Linux (AMD64 and ARM64)

## Installation Methods

### Method 1: Homebrew (Recommended)

**After the tap is set up:**

```bash
brew tap ivikasavnish/tap
brew install portless
```

**Or install directly from formula:**

```bash
brew install --build-from-source https://raw.githubusercontent.com/ivikasavnish/goroutine/main/homebrew/portless.rb
```

### Method 2: Install Script

This method automatically detects your OS and architecture:

```bash
curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | sh
```

### Method 3: Download Binary from Release

1. Go to [Releases](https://github.com/ivikasavnish/goroutine/releases)
2. Download the appropriate binary for your platform:
   - macOS Intel: `portless-darwin-amd64.tar.gz`
   - macOS Apple Silicon: `portless-darwin-arm64.tar.gz`
   - Linux AMD64: `portless-linux-amd64.tar.gz`
   - Linux ARM64: `portless-linux-arm64.tar.gz`
3. Extract and install:

```bash
# Download (replace URL with actual release)
curl -LO https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-darwin-arm64.tar.gz

# Extract
tar xzf portless-darwin-arm64.tar.gz

# Make executable
chmod +x portless-darwin-arm64

# Move to PATH
sudo mv portless-darwin-arm64 /usr/local/bin/portless
```

### Method 4: Build from Source

Requires Go 1.22 or later:

```bash
git clone https://github.com/ivikasavnish/goroutine.git
cd goroutine
make install
```

## Verify Installation

```bash
portless version
```

Expected output:
```
portless version 1.0.0
```

## Quick Usage

```bash
# Start the proxy server
portless proxy start

# Run your app with a named URL
portless myapp npm start
# Access at http://myapp.localhost:1355

# Run another service
portless api go run main.go
# Access at http://api.localhost:1355

# Check proxy status
portless proxy status

# Stop the proxy
portless proxy stop
```

## Next Steps

- Read the [full documentation](PORTLESS.md)
- Check out [usage examples](example/)
- Join the community discussions

## Troubleshooting

### Permission Denied

If you get permission errors on Linux/macOS:

```bash
sudo install portless /usr/local/bin/
```

### Command Not Found

Make sure `/usr/local/bin` is in your PATH:

```bash
echo $PATH
```

If not, add to your shell profile (~/.bashrc, ~/.zshrc, etc.):

```bash
export PATH="/usr/local/bin:$PATH"
```

### Brew Installation Issues

```bash
# Update Homebrew
brew update

# Try with verbose output
brew install --verbose portless
```

## Uninstall

### Homebrew:
```bash
brew uninstall portless
```

### Manual:
```bash
sudo rm /usr/local/bin/portless
```

## Support

- [Report Issues](https://github.com/ivikasavnish/goroutine/issues)
- [View Documentation](README.md)
- [Release Notes](CHANGELOG.md)
