# siyuan-cli

CLI for SiYuan Note written in Go. Search notes, read documents, update content, and export results from the terminal. Built for seamless integration with AI agents, automation workflows, and developer tooling.

```bash
siyuan-cli search --content "roadmap" --json
```

Before running commands, set `SIYUAN_BASE_URL` and `SIYUAN_TOKEN`.

- **AI-Agent Ready**: JSON output, structured responses, tool-calling friendly
- **Search & Query**: Full-text search, SQL queries, tag-based filtering
- **Content Management**: Read/update documents and blocks as Markdown
- **Export**: Markdown, HTML, PDF, DOCX for sharing and processing

See [API Reference](docs/siyuan-api.md) for complete SiYuan API documentation.

## Why use it

**For Power Users:**
- Search across all notes instantly with full-text search
- Query your database directly with SQL (SELECT only)
- Export notes to Markdown, HTML, PDF, DOCX for sharing
- Automate workflows with shell scripts and pipes

**For Developers:**
- Pipe note content to other CLI tools (`grep`, `awk`, `jq`)
- Integrate SiYuan into your development workflow
- Version control your notes with git
- Batch operations on notebooks, documents, and blocks

**Key Features:**
- 13 command groups covering all SiYuan operations
- JSON output for programmatic access (`--json`)
- Human-readable paths or IDs supported
- Cross-platform: macOS, Linux, Windows

**For AI Agents & Automation:**
- Structured JSON output for LLM tool calling
- Read-only SQL queries for safe data exploration
- Full CRUD operations on documents, blocks, and attributes
- Shell-scriptable for CI/CD pipelines
- No API credentials exposed in command history (via env vars)

## Install

### Go install

```bash
go install github.com/hcuong-me/siyuan-cli/cmd/siyuan@latest
siyuan-cli --help
```

### Pre-built binary

Download from [GitHub Releases](https://github.com/hcuong-me/siyuan-cli/releases).

### Homebrew (macOS)

```bash
brew tap hcuong-me/siyuan-cli
brew install siyuan-cli
```

Or in one line:

```bash
brew install hcuong-me/siyuan-cli/siyuan-cli
```

## Requirements

- Go >= 1.24 (for `go install`) or a pre-built binary
- A reachable SiYuan instance
- A valid SiYuan API token

## Configuration

The CLI reads two environment variables:

```bash
export SIYUAN_BASE_URL="http://127.0.0.1:6806"
export SIYUAN_TOKEN="your-token"
```

If either variable is missing, commands fail with a readable error.

## Quick Start

```bash
siyuan-cli system version
siyuan-cli search --content "project alpha"
siyuan-cli doc get --path /notebook/doc.sy
siyuan-cli notebook list
siyuan-cli sql query --statement "SELECT * FROM blocks LIMIT 5"
siyuan-cli tag list
```

## Commands

The CLI provides 13 command groups:

| Group | Description |
|-------|-------------|
| `system` | Version, time, boot status |
| `notebook` | Create, list, open, close, rename notebooks |
| `doc` | Create, read, update, move documents |
| `block` | Get, update, insert, remove content blocks |
| `search` | Full-text search blocks and documents |
| `sql` | Execute read-only SQL queries |
| `tag` | Manage tags and tag-based queries |
| `attr` | Get/set block attributes (metadata) |
| `export` | Export to Markdown, HTML, PDF, DOCX |
| `template` | List, render templates |
| `snapshot` | Repository snapshots (backups) |
| `asset` | Upload and manage assets |
| `file` | Raw filesystem operations |

See [Command Design Document](docs/command-design.md) for complete command reference with examples.

### Global flags

| Flag | Description |
| --- | --- |
| `-j, --json` | Output in JSON format |
| `-h, --help` | Show help |
| `-v, --version` | Show version |

## Common Workflows

Search, inspect, then export:

```bash
siyuan-cli search --content "roadmap" --json
siyuan-cli doc get --path /notebook/doc.sy
siyuan-cli export markdown --id 20260316120000-abc123 --output ./export.md
```

Attach metadata to a block:

```bash
siyuan-cli attr set --id blk-1 --key review-status --value done
siyuan-cli attr get --id blk-1 --json
```

Create a backup before major changes:

```bash
siyuan-cli snapshot create --memo "before-reorganization"
siyuan-cli notebook remove --id old-notebook --yes
```

## Development

```bash
go build -o dist/siyuan-cli ./cmd/siyuan
go test ./... -v -cover
go vet ./...
```

## Project Structure

```
.
├── cmd/siyuan/         # Main entry point
├── internal/
│   ├── cmd/            # CLI command handlers
│   ├── logic/          # Business logic layer
│   ├── siyuan/         # API client layer
│   ├── config/         # Configuration management
│   └── utils/output    # Output formatting utilities
└── docs/               # Documentation
```

## Architecture

Three-layer architecture:

1. **API Layer** (`internal/siyuan/`): Direct HTTP client for SiYuan API
2. **Logic Layer** (`internal/logic/`): Business logic and transformations
3. **Command Layer** (`internal/cmd/`): Cobra CLI command handlers

## Release

Tag a release to trigger GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This builds binaries for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, and windows/amd64, then uploads them to GitHub Releases.

## Troubleshooting

**Missing environment variables** — Confirm `SIYUAN_BASE_URL` and `SIYUAN_TOKEN` are exported in the same shell session.

**Need IDs for follow-up commands** — Re-run the read/search command with `--json`.

**A remove or restore command failed** — Re-run with `--yes` only after confirming the target ID or path is correct.

## Contributing

See [CI/CD Setup Guide](docs/CI_CD_SETUP.md) for configuring GitHub Actions and releases.

Contributions are welcome. Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Run tests (`go test ./...`)
4. Run linter (`go vet ./...`)
5. Commit your changes (`git commit -am 'Add new feature'`)
6. Push to the branch (`git push origin feature/my-feature`)
7. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related

- [SiYuan](https://github.com/siyuan-note/siyuan) - The note-taking app
- [SiYuan API Documentation](https://github.com/siyuan-note/siyuan/blob/master/API.md)