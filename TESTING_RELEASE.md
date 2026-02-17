# Testing the Release Workflow

This document provides instructions for testing the entire Homebrew release workflow after merging this PR.

## Prerequisites

- [ ] This PR has been merged to main branch
- [ ] You have push access to the repository
- [ ] GitHub Actions is enabled

## Test Steps

### 1. Create a Test Release

```bash
# Checkout main and ensure it's up to date
git checkout main
git pull origin main

# Create a test tag (use the next appropriate version)
git tag -a v1.0.1 -m "Test release v1.0.1"

# Push the tag to trigger the release workflow
git push origin v1.0.1
```

### 2. Monitor GitHub Actions

1. Go to: https://github.com/ivikasavnish/goroutine/actions
2. Click on the "Release" workflow run
3. Wait for it to complete (should take 2-5 minutes)
4. Check for any errors in the logs

**Expected Results:**
- ✅ All build steps succeed
- ✅ Binaries are created for all 4 platforms
- ✅ Archives are created (*.tar.gz)
- ✅ Checksums file is generated
- ✅ GitHub release is created

### 3. Verify the GitHub Release

1. Go to: https://github.com/ivikasavnish/goroutine/releases
2. Find the v1.0.1 release
3. Verify the following files are present:
   - [ ] portless-darwin-amd64.tar.gz
   - [ ] portless-darwin-arm64.tar.gz
   - [ ] portless-linux-amd64.tar.gz
   - [ ] portless-linux-arm64.tar.gz
   - [ ] checksums.txt

### 4. Update the Homebrew Formula

```bash
# Run the update script
./homebrew/update_formula.sh v1.0.1

# Verify the formula was updated
cat homebrew/portless.rb | grep sha256

# Commit the updated formula
git add homebrew/portless.rb
git commit -m "Update Homebrew formula for v1.0.1"
git push origin main
```

**Expected Results:**
- ✅ Script downloads checksums successfully
- ✅ Formula file is updated with real SHA256 values
- ✅ No placeholder values remain

### 5. Test Local Formula Installation

**On macOS (if available):**
```bash
# Install from local formula
brew install --build-from-source homebrew/portless.rb

# Verify installation
portless version
# Expected: portless version 1.0.1

# Test basic functionality
portless help

# Uninstall
brew uninstall portless
```

**On Linux (if available):**
```bash
# Same steps as macOS
brew install --build-from-source homebrew/portless.rb
portless version
portless help
brew uninstall portless
```

### 6. Test Install Script

```bash
# Test the install script
curl -fsSL https://raw.githubusercontent.com/ivikasavnish/goroutine/main/install.sh | VERSION=v1.0.1 sh

# Verify installation
portless version

# Uninstall
sudo rm /usr/local/bin/portless
```

### 7. Create Homebrew Tap (Optional)

If you want to make installation easier:

```bash
# Create a new repository: homebrew-tap
# Then:
cd homebrew-tap
mkdir -p Formula
cp /path/to/goroutine/homebrew/portless.rb Formula/
git add Formula/portless.rb
git commit -m "Add portless formula v1.0.1"
git push

# Test tap installation
brew tap ivikasavnish/tap
brew install portless
portless version
```

## Troubleshooting

### Build Fails in GitHub Actions

**Check:**
- Go version compatibility (should be 1.22)
- Module dependencies (go.mod and go.sum are up to date)
- Build command syntax
- GitHub Actions logs for specific errors

**Fix:**
- Update workflow file if needed
- Delete the tag: `git tag -d v1.0.1 && git push origin :refs/tags/v1.0.1`
- Fix the issue
- Create a new tag

### Checksums Don't Match

**Check:**
- Download checksums.txt from the release
- Manually verify: `sha256sum downloaded-file.tar.gz`
- Compare with checksums.txt

**Fix:**
- Re-run the update_formula.sh script
- Ensure you're using the correct version number

### Formula Installation Fails

**Check:**
```bash
# Validate formula
brew audit --strict homebrew/portless.rb

# Test with verbose output
brew install --build-from-source --verbose homebrew/portless.rb
```

**Common Issues:**
- Binary name mismatch (must match platform name in archive)
- URL 404 (verify release exists)
- Checksum mismatch (re-run update_formula.sh)

### Install Script Fails

**Check:**
- Release exists: curl -I https://github.com/ivikasavnish/goroutine/releases/download/v1.0.1/checksums.txt
- Archive format is correct (tar.gz with proper structure)
- Script has correct permissions (should be executable)

## Success Criteria

- [ ] GitHub Actions workflow completes without errors
- [ ] All 4 platform binaries are built and released
- [ ] Checksums file is generated correctly
- [ ] Formula can be updated with update_formula.sh
- [ ] Formula installs successfully on at least one platform
- [ ] Install script works correctly
- [ ] portless version shows correct version number
- [ ] Basic commands work (portless help, portless version)

## Next Steps After Successful Test

1. **Document the process:** Ensure RELEASE_PROCESS.md is accurate
2. **Update main README:** Add installation badge if desired
3. **Create Homebrew tap:** For easier installation
4. **Announce the release:** Share with users
5. **Monitor issues:** Watch for installation problems

## Rolling Back a Release

If you need to delete a release:

```bash
# Delete the tag locally and remotely
git tag -d v1.0.1
git push origin :refs/tags/v1.0.1

# Delete the GitHub release manually from the web interface
# Go to: https://github.com/ivikasavnish/goroutine/releases
# Click "Delete" on the release
```

## Notes

- Test releases can be marked as "Pre-release" in GitHub to avoid confusing users
- Consider using a test tag like `v1.0.1-rc1` for initial testing
- Document any issues found during testing in GitHub Issues
