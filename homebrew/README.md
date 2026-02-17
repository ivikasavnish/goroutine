# Homebrew Installation for Portless

This directory contains the Homebrew formula and related scripts for installing `portless` on macOS and Linux (both AMD64 and ARM64 architectures).

## Quick Install

### Option 1: Direct Install (When releases are available)
```bash
brew install ivikasavnish/tap/portless
```

### Option 2: Install from Formula File
```bash
brew install --build-from-source homebrew/portless.rb
```

### Option 3: Using install script (builds from source)
```bash
curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | sh
```

## For Maintainers

### Creating a Release

1. **Tag a new version:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Build binaries for all platforms (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64)
   - Create a GitHub release
   - Upload binaries and checksums

3. **Update the Homebrew formula:**
   ```bash
   ./homebrew/update_formula.sh v1.0.0
   ```
   This script will:
   - Download the checksums from the release
   - Update `homebrew/portless.rb` with the correct version and SHA256 checksums

4. **Test the formula locally:**
   ```bash
   brew install --build-from-source homebrew/portless.rb
   portless version
   brew uninstall portless
   ```

5. **Publish to Homebrew Tap (Optional):**
   
   To make installation easier for users, create a Homebrew tap repository:
   
   a. Create a new repository: `homebrew-tap`
   
   b. Copy the formula:
   ```bash
   mkdir -p Formula
   cp homebrew/portless.rb Formula/
   git add Formula/portless.rb
   git commit -m "Add portless formula"
   git push
   ```
   
   c. Users can then install with:
   ```bash
   brew tap ivikasavnish/tap
   brew install portless
   ```

## Supported Platforms

The formula supports the following platforms:
- **macOS (Darwin)**
  - Intel (AMD64)
  - Apple Silicon (ARM64)
- **Linux**
  - AMD64
  - ARM64

## Architecture Details

The formula uses conditional installation based on the platform:
- Automatically detects the OS and CPU architecture
- Downloads the correct binary for the platform
- Installs to the Homebrew binary directory

## Troubleshooting

### Formula validation
```bash
brew audit --strict homebrew/portless.rb
```

### Test installation
```bash
brew install --build-from-source --verbose homebrew/portless.rb
```

### Check formula info
```bash
brew info homebrew/portless.rb
```

## Development

### Testing the GitHub Actions workflow locally

You can test the release workflow locally using [act](https://github.com/nektos/act):
```bash
act -j build
```

### Manual build for all platforms
```bash
make build-all
```

This will create binaries in the `bin/` directory for all supported platforms.
