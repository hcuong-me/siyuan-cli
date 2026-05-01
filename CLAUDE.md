# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Start

```bash
make build              # Build CLI binary to dist/siyuan-cli
go run ./cmd/siyuan --help    # Run locally
```
## Development Commands

```bash
make build              # Build binary
make install            # Install globally
make test               # Run tests with race detection
make test-coverage      # Generate HTML coverage report
make clean              # Clean build artifacts
golangci-lint run ./... # Lint
```

## Project Structure

```
cmd/siyuan/         # Entry point (main.go)
internal/
  siyuan/           # API client layer
  logic/            # Business logic with caching
  cmd/              # Cobra CLI handlers
  config/           # Env config (SIYUAN_BASE_URL, SIYUAN_TOKEN)
  utils/output/     # JSON/table formatting
docs/               # API docs, command reference
```

## Adding a Command

1. `internal/siyuan/{group}.go` - API methods
2. `internal/logic/{group}.go` - Business logic
3. `internal/cmd/{group}/` - CLI commands
4. Register in `internal/cmd/root.go`

## Configuration

Required env vars:
- `SIYUAN_BASE_URL` (default: http://127.0.0.1:6806)
- `SIYUAN_TOKEN` (from SiYuan Settings > About)

## Release

```bash
git tag v0.1.0
git push origin v0.1.0   # Triggers GitHub Actions release
```
