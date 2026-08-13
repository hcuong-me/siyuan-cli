# Development

Run the checks before you submit a change:

```bash
make test
go vet ./...
make lint
make lint-arch
```

Inspect the live contract without a server request:

```bash
go run ./cmd/siyuan tools
```

## Live integration tests

Set `SIYUAN_TOKEN` before testing an operational request. `make test-integration` runs the real-server document-tree test only when `SIYUAN_INTEGRATION_TEST=1`. Set `SIYUAN_BASE_URL` when the server is not at `http://127.0.0.1:6806`.
