# siyuan-cli

`siyuan-cli` is a tool runner for SiYuan Note. It exposes the same operations through two interfaces: a JSON-over-stdin protocol for agents and a flag syntax for humans. Every response is one JSON object on the agent path. The human path prints readable text or tables.

## Install

```bash
brew install hcuong-me/tap/siyuan-cli
```

Or build from source with Go 1.24+:

```bash
make build   # writes dist/siyuan-cli
```

Set the server address and token before you run an operational tool:

```bash
export SIYUAN_BASE_URL="http://127.0.0.1:6806"
export SIYUAN_TOKEN="your-token"
```

## Human path

```bash
siyuan-cli context read_document --notebook notebook-id --path /Projects/Roadmap
siyuan-cli status version
```

Side-effecting operations preview first and prompt `Apply? [y/N]`. `--yes` previews and applies without prompting. `--json` prints exactly one raw agent envelope. For side effects it is preview-only, so `--json --yes` and `--json --mode apply` are rejected. Use `-h` for top-level usage, `<tool> -h` for an operation list, or `<tool> <operation> -h` for schema-driven flag help.

## Agent path

Discover the contract without stdin:

```bash
siyuan-cli tools
```

Each operational request is one JSON object on stdin:

```bash
printf '%s' '{"version":"v1","operation":"read_document","input":{"notebook":"notebook-id","path":"/Projects/Roadmap"}}' | siyuan-cli context
```

Preview a mutation. Apply uses the returned confirmation token with the same input:

```bash
printf '%s' '{"version":"v1","operation":"update_block","input":{"block_id":"block-id","content":"New text"},"mode":"preview"}' | siyuan-cli write
```

The seven tools are `tools`, `status`, `context`, `write`, `organize`, `export`, and `maintain`.

`context.search_documents` returns the complete, unpaginated result set with a count. Use returned stable IDs and canonical paths for follow-up operations. `page` and `size` are not accepted fields.

## Documentation

- [Architecture](docs/architecture.md) — package boundaries, request flow, and exceptions
- [Tool design](docs/command-design.md) — invocation paths and side-effect safety
- [Human interface](docs/human-interface.md) — flag syntax and interactive prompts
- [Agent tool contract](docs/agent-tool-contract.md) — request and response formats
- [SiYuan API reference](docs/siyuan-api.md) — endpoint request and response details
- [Legacy command baseline](docs/legacy-command-baseline.md) — source-derived mapping of the retired command groups

## Development

```bash
go test ./...
go vet ./...
make lint-arch
```

The real-server integration test remains opt-in with `SIYUAN_INTEGRATION_TEST=1`.
