# CI/CD Configuration Guide

## Overview

This project uses GitHub Actions for CI/CD with the following workflows:

- **CI** (`.github/workflows/ci.yml`): Runs on every PR and push to `main`
- **Release** (`.github/workflows/release.yml`): Runs when a tag is pushed

## File Structure

```
.github/
├── workflows/
│   ├── ci.yml          # Continuous Integration
│   ├── release.yml     # Release builds and Homebrew updates
│   └── README.md       # Documentation
```

## CI Workflow

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**
1. **Test**: Runs `go test` with race detection and coverage
2. **Lint**: Runs `go vet` and `golangci-lint`
3. **Build**: Verifies the build succeeds

**Go Version:** 1.24

## Release Workflow

**Triggers:**
- Push tags matching `v*` (e.g., `v0.1.0`, `v1.0.0`)

**Jobs:**

### 1. Build Binaries
- Builds for 5 platforms in parallel:
  - `darwin/amd64` (Intel Mac)
  - `darwin/arm64` (Apple Silicon)
  - `linux/amd64`
  - `linux/arm64`
  - `windows/amd64`
- Injects version from git tag
- Uploads artifacts

### 2. Create Release
- Downloads all built artifacts
- Calculates SHA256 checksums for macOS binaries
- Creates GitHub Release with:
  - All binaries attached
  - Auto-generated release notes
- Uploads checksums for Homebrew

### 3. Update Homebrew
- Checks out `hcuong-me/homebrew-siyuan-cli` tap repo
- Updates formula with new version and SHA256 hashes
- Commits and pushes changes

## Required Setup

### 1. Create Homebrew Tap Repository

Create a new public repository:
- **Name:** `homebrew-siyuan-cli`
- **URL:** `github.com/hcuong-me/homebrew-siyuan-cli`
- **Visibility:** Public (required for Homebrew taps)

Initialize with a README (optional).

### 2. Add GitHub Secret

The release workflow needs a Personal Access Token to update the Homebrew tap:

**Step 1: Generate Token**
1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scopes:
   - `repo` (full control of private repositories)
4. Generate and copy the token

**Step 2: Add to Repository Secrets**
1. Go to your `siyuan-cli` repository
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `HOMEBREW_TAP_TOKEN`
5. Value: Your personal access token
6. Click "Add secret"

### 3. Enable Actions

Ensure GitHub Actions are enabled:
1. Go to repository Settings → Actions → General
2. Under "Actions permissions", select:
   - "Allow all actions and reusable workflows"

## Creating a Release

### Step 1: Prepare the Release

Ensure all changes are committed and tests pass locally:

```bash
# Run tests
go test ./...

# Build locally
go build -o dist/siyuan ./cmd/siyuan
```

### Step 2: Tag the Release

Use semantic versioning:

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push tag to trigger release
git push origin v0.1.0
```

### Step 3: Monitor the Workflow

1. Go to Actions tab in your repository
2. Click on the "Release" workflow run
3. Wait for all jobs to complete:
   - Build (5 parallel builds)
   - Release (creates GitHub Release)
   - Update Homebrew (updates formula)

### Step 4: Verify Release

**Check GitHub Release:**
- Go to Releases page
- Verify all 5 binaries are attached
- Check release notes

**Check Homebrew:**
- Go to `hcuong-me/homebrew-siyuan-cli` repository
- Verify `siyuan-cli.rb` was updated
- Check the commit message: "Bump siyuan-cli to v0.1.0"

**Test Installation:**

```bash
# Update Homebrew
brew update

# Install
brew install hcuong-me/siyuan-cli/siyuan-cli

# Verify version
siyuan-cli --version
```

## Troubleshooting

### Workflow Fails at "Update Homebrew"

**Error:** `Repository not found` or authentication failed

**Solution:**
- Verify `HOMEBREW_TAP_TOKEN` secret is set correctly
- Ensure the token has `repo` scope
- Confirm `homebrew-siyuan-cli` repository exists and is public

### Binary Not Found in Release

**Error:** Missing artifacts

**Solution:**
- Check the "Build Binaries" job logs
- Verify all matrix builds completed successfully
- Check artifact upload/download steps

### Version Shows as "dev"

**Error:** `siyuan-cli version dev` instead of actual version

**Solution:**
- Ensure you're building with ldflags: `-X siyuan/internal/cmd.version=VERSION`
- Check the release workflow's build step includes this flag

### Formula SHA Mismatch

**Error:** SHA256 mismatch when installing via Homebrew

**Solution:**
- Check that SHA256 values in formula match the downloaded binaries
- Verify the `shasum` command worked correctly in the workflow
- Re-run the release workflow if needed

## Manual Testing

Test the CI workflows locally with [act](https://github.com/nektos/act):

```bash
# Install act
brew install act

# Run CI workflow
act push

# Run release workflow (requires secrets)
act push --eventpath .github/test-events/tag-push.json
```

## Workflow Modification

To modify workflows:

1. Edit files in `.github/workflows/`
2. Test changes on a branch
3. Create a PR to review changes
4. Merge to `main`

**Common modifications:**

**Add new build target:**
Edit `.github/workflows/release.yml`:
```yaml
strategy:
  matrix:
    include:
      - os: freebsd
        arch: amd64
        ext: ''
```

**Change Go version:**
Edit both workflow files:
```yaml
with:
  go-version: '1.25'
```

**Add test coverage threshold:**
Edit `.github/workflows/ci.yml`:
```yaml
- name: Check coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    if (( $(echo "$coverage < 70" | bc -l) )); then
      echo "Coverage below 70%"
      exit 1
    fi
```
