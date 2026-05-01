# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

Build the CLI:
```bash
go build -o dist/siyuan ./cmd/siyuan
```

Or use make:
```bash
make build
```

Run all tests:
```bash
go test ./...
```

Run tests for a specific package:
```bash
go test ./internal/siyuan/...
go test ./internal/logic/...
```

Run tests with coverage:
```bash
go test ./... -cover
```

Lint:
```bash
go vet ./...
```

Run the CLI locally:
```bash
go run ./cmd/siyuan --help
go run ./cmd/siyuan notebook list
```

## Architecture

Three-layer architecture for command implementation:

1. **API Layer** (`internal/siyuan/`): HTTP client for SiYuan API
   - `client.go`: Base HTTP client with `Post()`, `Get()` methods
   - Each API endpoint has its own file (e.g., `notebook.go`, `doc.go`)
   - Types are defined alongside their API calls
   - Authentication via `Authorization: Token {token}` header

2. **Logic Layer** (`internal/logic/`): Business logic and transformations
   - One file per command group (e.g., `notebook.go`, `doc.go`)
   - Each Logic struct embeds `*siyuan.Client`
   - Handles caching (e.g., notebook list cache with expiry)
   - Converts API types to internal types
   - Validation functions (e.g., `ValidateNotebookName()`)

3. **Command Layer** (`internal/cmd/`): Cobra CLI handlers
   - Each command group is a package folder (e.g., `cmd/notebook/`)
   - `root.go` registers all command groups
   - Commands return errors; `Execute()` in root.go handles exit codes
   - Use `cmd.Flag("json").Changed` to check if JSON output requested

## Patterns

Adding a new command group:

1. **API layer**: Create `internal/siyuan/{group}.go` with API methods
2. **Logic layer**: Create `internal/logic/{group}.go` with business logic
3. **Command layer**: Create `internal/cmd/{group}/{group}.go` with Cobra commands
4. **Register**: Add import and `root.AddCommand()` in `internal/cmd/root.go`

Command implementation template:

```go
func NewCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "verb [args]",
        Short: "Description",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            l, err := logic.NewLogic()
            if err != nil {
                return err
            }

            result, err := l.Method(cmd.Context(), args[0])
            if err != nil {
                return err
            }

            if cmd.Flag("json").Changed {
                return output.AsJSON(result)
            }

            // Table or plain text output
            output.AsTable([]string{"COL1", "COL2"}, rows)
            return nil
        },
    }
}
```

Destructive operations require `--yes` flag:

```go
var yes bool
cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
if !yes {
    return fmt.Errorf("use --yes to confirm removal")
}
```

## Configuration

Environment variables required:
- `SIYUAN_BASE_URL`: API endpoint (default: `http://127.0.0.1:6806`)
- `SIYUAN_TOKEN`: API token from Settings > About in SiYuan

## Output Utilities

`internal/utils/output/` provides:
- `AsJSON(data)` - Format as indented JSON
- `AsTable(headers, rows)` - Format as table
- `Printf()`, `Println()` - Direct output
- `Error(err)` - Print to stderr

## Testing

Tests use `httptest` to mock SiYuan API. See `internal/siyuan/*_test.go` for examples.

Key patterns:
- Create mock server with expected responses
- Use `client.SetHTTPClient()` to inject test client
- Test both success and error paths

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- `github.com/olekukonko/tablewriter`: Table formatting

Go 1.24+ required.
