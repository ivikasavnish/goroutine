# Release Process for Portless

This document describes how to create a new release of Portless with Homebrew support for macOS and Linux (AMD64 and ARM64 architectures).

## Prerequisites

- GitHub account with push access to the repository
- Repository with GitHub Actions enabled
- Git installed locally

## Release Steps

### 1. Prepare for Release

Ensure all changes are merged to the main branch:

```bash
git checkout main
git pull origin main
```

### 2. Run Tests

```bash
# Run all tests
make test

# Test building for all platforms and creating release archives
make release

# Test creating checksums
make release-checksums

# Clean up test artifacts
make release-clean

# Verify the binaries work
./bin/portless-linux-amd64 version
./bin/portless-darwin-amd64 version
```

### 3. Create and Push a Git Tag

Choose a version number following semantic versioning (e.g., v1.0.1, v1.1.0, v2.0.0):

```bash
VERSION=v1.0.1  # Change this to your desired version

# Create an annotated tag
git tag -a $VERSION -m "Release $VERSION"

# Push the tag to trigger the release workflow
git push origin $VERSION
```

### 4. Monitor GitHub Actions

1. Go to: https://github.com/ivikasavnish/goroutine/actions
2. Click on the "Release" workflow
3. Watch the build process complete
4. The workflow will:
   - Build binaries for all platforms (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64)
   - Create tar.gz archives
   - Generate SHA256 checksums
   - Create a GitHub release
   - Upload all artifacts

### 5. Update Homebrew Formula

Once the release is complete, update the Homebrew formula with the actual checksums:

```bash
# Run the update script with your version
./homebrew/update_formula.sh v1.0.1

# Verify the formula was updated
cat homebrew/portless.rb
```

### 6. Commit and Push Formula Update

```bash
git add homebrew/portless.rb
git commit -m "Update Homebrew formula for $VERSION"
git push origin main
```

### 7. Test the Installation

Test that users can install via Homebrew:

```bash
# Test local formula
brew install --build-from-source homebrew/portless.rb

# Verify installation
portless version

# Clean up
brew uninstall portless
```

## Creating a Homebrew Tap (One-Time Setup)

To make installation easier for users, you can create a Homebrew tap:

### Step 1: Create Tap Repository

1. Create a new GitHub repository named `homebrew-tap`
2. Clone it locally:
   ```bash
   git clone https://github.com/ivikasavnish/homebrew-tap.git
   cd homebrew-tap
   ```

### Step 2: Add Formula

```bash
mkdir -p Formula
cp path/to/goroutine/homebrew/portless.rb Formula/
git add Formula/portless.rb
git commit -m "Add portless formula"
git push
```

### Step 3: Users Can Now Install

```bash
brew tap ivikasavnish/tap
brew install portless
```

## Updating the Tap After Each Release

After each release and formula update:

```bash
cd homebrew-tap
cp path/to/goroutine/homebrew/portless.rb Formula/
git add Formula/portless.rb
git commit -m "Update portless to $VERSION"
git push
```

## Troubleshooting

### Build Fails

- Check that all dependencies are properly vendored
- Ensure Go version in workflow matches `go.mod`
- Check GitHub Actions logs for specific errors

### Formula Installation Fails

```bash
# Validate the formula
brew audit --strict homebrew/portless.rb

# Test with verbose output
brew install --build-from-source --verbose homebrew/portless.rb

# Check formula info
brew info homebrew/portless.rb
```

### SHA256 Mismatch

If checksums don't match:
1. Re-download the release artifacts
2. Verify checksums.txt in the release
3. Re-run `./homebrew/update_formula.sh`

## Manual Release (Alternative)

If you need to create a release manually, you can use the Makefile targets:

### Using Makefile (Recommended for local testing):
```bash
# Set the version (with v prefix)
export VERSION=v1.0.1

# Create release archives for all platforms
make release

# Generate checksums
make release-checksums

# The artifacts will be in the dist/ directory:
# - portless-darwin-amd64.tar.gz
# - portless-darwin-arm64.tar.gz
# - portless-linux-amd64.tar.gz
# - portless-linux-arm64.tar.gz
# - checksums.txt

# Clean up when done
make release-clean
```

### Manual build (if Make is not available):
```bash
VERSION=v1.0.1
mkdir -p dist
GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$VERSION" -o dist/portless-darwin-amd64 ./cmd/portless
GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$VERSION" -o dist/portless-darwin-arm64 ./cmd/portless
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$VERSION" -o dist/portless-linux-amd64 ./cmd/portless
GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$VERSION" -o dist/portless-linux-arm64 ./cmd/portless
```

### Create archives:
```bash
cd dist
for file in portless-*; do
  tar czf "${file}.tar.gz" "$file"
done
```

### Generate checksums:
```bash
sha256sum *.tar.gz > checksums.txt
```

### Create GitHub Release:
1. Go to https://github.com/ivikasavnish/goroutine/releases/new
2. Select the tag
3. Add release notes
4. Upload all `*.tar.gz` files and `checksums.txt`
5. Publish the release

## Version Guidelines

Follow semantic versioning:
- **Major (v2.0.0)**: Breaking changes
- **Minor (v1.1.0)**: New features, backwards compatible
- **Patch (v1.0.1)**: Bug fixes, backwards compatible

## Release Checklist

- [ ] All tests pass
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated
- [ ] Version is tagged
- [ ] GitHub release is created
- [ ] Binaries are built and uploaded
- [ ] Homebrew formula is updated
- [ ] Formula is tested locally
- [ ] Tap is updated (if applicable)
- [ ] Release announcement is made
