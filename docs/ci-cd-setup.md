# CI/CD Configuration Guide

## Overview

This project uses GitHub Actions for CI/CD with the following workflows:

- **CI** (`.github/workflows/ci.yml`): Runs on every PR and push to `main`
- **Release** (`.github/workflows/release.yml`): Runs when a tag is pushed

## File Structure

```
.github/
└── workflows/
    ├── ci.yml          # Continuous Integration
    └── release.yml     # Release builds and Homebrew updates
```

## CI Workflow

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**
1. **Test**: Runs `go test` with race detection and coverage
2. **Lint**: Runs `go vet` and `golangci-lint`
3. **Build**: Runs the build

**Go Version:** 1.24

## Release Workflow

**Triggers:**
- Push tags that match `v*`, for example `v0.1.0` and `v1.0.0`

**Jobs:**

### 1. Build Binaries
- Builds for 5 platforms in parallel:
  - `darwin/amd64` (Intel Mac)
  - `darwin/arm64` (Apple Silicon)
  - `linux/amd64`
  - `linux/arm64`
  - `windows/amd64`
- Uploads artifacts

### 2. Create Release
- Downloads all built artifacts
- Calculates SHA256 checksums for macOS binaries
- Creates GitHub Release with:
  - All binaries attached
  - Auto-generated release notes
- Uploads checksums for Homebrew

### 3. Update Homebrew
- Checks out the `hcuong-me/homebrew-tap` repository using the `TAP_GITHUB_TOKEN` secret
- Updates formula with new version and SHA256 hashes
- Commits and pushes changes

## Required Setup

### 1. Create Homebrew Tap Repository

Create a new public repository:
- **Name:** `homebrew-tap`
- **URL:** `github.com/hcuong-me/homebrew-tap`
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
4. Name: `TAP_GITHUB_TOKEN`
5. Value: Your personal access token
6. Click "Add secret"

### 3. Enable Actions

Make sure that GitHub Actions is enabled:
1. Go to repository Settings → Actions → General
2. Under "Actions permissions", select:
   - "Allow all actions and reusable workflows"

## Creating a Release

### Step 1: Prepare the Release

Commit all changes. Then run the local tests:

```bash
# Run tests
make test

# Build locally
make build
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
3. Watch the workflow until all jobs complete:
   - Build (5 parallel builds)
   - Release (creates GitHub Release)
   - Update Homebrew (updates formula)

### Step 4: Verify Release

**Check GitHub Release:**
- Go to Releases page
- Verify all 5 binaries are attached
- Check release notes

**Check Homebrew:**
- Go to the `hcuong-me/homebrew-tap` repository
- Verify `siyuan-cli.rb` was updated
- Check the commit message: "siyuan-cli: update to v0.1.0"

**Test Installation:**

```bash
# Update Homebrew
brew update

# Install
brew install hcuong-me/tap/siyuan-cli

# Verify the installed command without a server request
siyuan-cli tools
```

## Troubleshooting

### Workflow Fails at "Update Homebrew"

**Error:** `Repository not found` or authentication failed

**Solution:**
- Verify the `TAP_GITHUB_TOKEN` secret is set
- Make sure that the token has `repo` scope
- Confirm the `hcuong-me/homebrew-tap` repository exists and is accessible to the token

### Binary Not Found in Release

**Error:** Missing artifacts

**Solution:**
- Check the "Build Binaries" job logs
- Verify all matrix builds completed successfully
- Check artifact upload/download steps

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

# Run the release workflow (requires the TAP_GITHUB_TOKEN secret and a
# tag-push event payload, for example: act push --eventpath tag-push.json)
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
