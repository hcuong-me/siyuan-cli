# siyuan-cli agent guide

This repository contains a Go CLI for the SiYuan Note HTTP API. Keep this file operational. Read the linked docs before changing a command path or its API mapping.

## 1. Start here

- [Architecture](docs/architecture.md): package boundaries, request flow, and exceptions.
- [Development](docs/development.md): setup, build, test, lint, and live integration checks.
- [Quality](docs/quality.md): checks and rules for safe changes.
- [Tool design](docs/command-design.md): invocation paths and side-effect safety.
- [Agent tool contract](docs/agent-tool-contract.md): request envelope, preview/apply, and error semantics.
- [Human interface](docs/human-interface.md): flag routing and interactive prompts.
- [SiYuan API reference](docs/siyuan-api.md): endpoint request and response details.

## 2. Commands

Run these from the repository root.

```bash
make build             # writes dist/siyuan-cli
make test              # unit and package tests; disables live integration tests
make test-integration  # may call a real SiYuan instance
make lint              # runs golangci-lint
make lint-arch         # runs repository dependency and quality checks
go run ./cmd/siyuan tools
```

`make test-integration` needs all of the following in the invoking shell:

```bash
export SIYUAN_INTEGRATION_TEST=1
export SIYUAN_TOKEN="your-token"
# Optional. The client defaults to http://127.0.0.1:6806.
export SIYUAN_BASE_URL="http://127.0.0.1:6806"
```

It currently runs a live document-tree request against a fixed notebook ID. Do not run it against a SiYuan instance you do not intend to query. [Development](docs/development.md#live-integration-tests) has the details.

## 3. Configuration and safety

- `SIYUAN_TOKEN` is required. Never put it in source, fixtures, docs, shell history, or commits.
- `SIYUAN_BASE_URL` is optional and defaults to `http://127.0.0.1:6806`.
- Commands send authenticated HTTP requests with `Authorization: Token <token>`.
- Run `siyuan-cli tools` before testing a live request. Preview every side effect before apply.
- Preserve the JSON response envelope when extending a tool.

> Sources: `internal/config/config.go`, `internal/siyuan/client.go`, `internal/tool/dispatcher.go`.

## 4. Code map

| Path | Responsibility |
| --- | --- |
| `cmd/siyuan/` | Process entry point. |
| `internal/tool/` | JSON protocol, schemas, dispatcher, previews, and handlers. |
| `internal/logic/` | Reusable operations and transformations. |
| `internal/siyuan/` | SiYuan API methods and shared HTTP client. |
| `internal/config/` | Environment configuration. |
| `internal/utils/output/` | Legacy utility code. Do not use it for tools. |
| `scripts/` | Repository lint checks invoked by `make lint-arch`. |

## 5. Adding or changing a tool

1. Add the operation schema in `internal/tool/schema.go`.
2. Put API-specific request and response handling in `internal/siyuan/`.
3. Put behavior that combines API calls, caching, or transformations in `internal/logic/`.
4. Keep the tool handler focused on validated JSON input and response data.
5. Add preview and apply coverage for every side effect.
6. Update the contract docs ([agent-tool-contract.md](docs/agent-tool-contract.md), [siyuan-api.md](docs/siyuan-api.md)) and add focused tests. Use `make test` and `make lint-arch` before handoff.

`status` creates and calls `siyuan.Client` directly because it only reads server state. Keep complex behavior in a dedicated tool handler.

> Sources: `internal/tool/schema.go`, `internal/tool/dispatcher.go`, `internal/siyuan/notebook.go`.

## 6. Tests and documentation

- Keep ordinary tests independent of a live SiYuan server. `make test` sets `SIYUAN_INTEGRATION_TEST=0`.
- Gate a real-server test with `SIYUAN_INTEGRATION_TEST=1`; make it skip when configuration is unavailable.
- Update [command-design.md](docs/command-design.md) when a tool's invocation or safety behavior changes.
- Keep docs factual. Cite the source file and line range for architecture claims.

> Sources: `Makefile:15-23`, `internal/siyuan/document_test.go:132-139`.

## 7. Release and CI

CI runs tests, vet, golangci-lint, and a build on pushes and pull requests to `main`. Tags matching `v*` start the release workflow. See [ci-cd-setup.md](docs/ci-cd-setup.md) for the release steps, Homebrew tap, and required secrets.

> Sources: `.github/workflows/ci.yml:3-66`, `.github/workflows/release.yml:3-176`.
