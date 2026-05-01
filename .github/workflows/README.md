# CI/CD Documentation

## Workflows

### CI (`ci.yml`)

Runs on every push and PR:
- **Test**: `go test ./... -v -race`
- **Lint**: `go vet ./...` + `golangci-lint`
- **Build**: Verify build succeeds

### Release (`release.yml`)

Triggers on tag push (`v*`):
1. Build for 5 platforms
2. Create GitHub Release with artifacts
3. Calculate SHA256 for macOS binaries
4. Update Homebrew formula

## Creating a Release

```bash
# 1. Tag the release
git tag v0.1.0

# 2. Push the tag
git push origin v0.1.0

# 3. GitHub Actions will:
#    - Build binaries for all platforms
#    - Create GitHub Release
#    - Update Homebrew formula
```

## Installing via Homebrew

```bash
# Add tap
brew tap hcuong-me/siyuan-cli

# Install
brew install siyuan-cli

# Or in one line
brew install hcuong-me/siyuan-cli/siyuan-cli
```

## Required Secrets

- `HOMEBREW_TAP_TOKEN`: Personal access token with `repo` scope for updating Homebrew tap

## Manual Build

```bash
# Local development build
go build -o dist/siyuan ./cmd/siyuan

# With version
go build -ldflags="-X siyuan/internal/cmd.version=0.1.0" -o dist/siyuan ./cmd/siyuan
```
