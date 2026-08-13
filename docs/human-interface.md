# Human interface

`siyuan-cli` runs two ways from the same binary. The agent path reads one JSON request on stdin and prints one JSON envelope. The human path takes a tool, an operation, and `--flag value` arguments, and prints readable text.

The router sends `siyuan-cli <tool> <operation> [flags]` to the human path whenever the second argument is a cataloged operation of the first. A tool name alone with JSON piped on stdin delegates to the agent path. On an interactive terminal it prints the tool's operation list instead. Every other invocation delegates unchanged to the agent path.

```bash
# Human path: flags become the request
siyuan-cli context read_document --notebook notebook-id --path /Projects/Roadmap

# Agent path: one JSON request on stdin
printf '%s' '{"version":"v1","operation":"read_document","input":{"notebook":"notebook-id","path":"/Projects/Roadmap"}}' | siyuan-cli context
```

Set the server address and token before an operational request:

```bash
export SIYUAN_BASE_URL="http://127.0.0.1:6806"
export SIYUAN_TOKEN="your-token"
```

## Discovery

- `siyuan-cli` or `siyuan-cli -h` (or `--help`) prints short usage.
- `siyuan-cli tools` prints the agent tool catalog as JSON (takes no stdin).
- `siyuan-cli <tool> -h` lists the operations of one tool.
- `siyuan-cli <tool> <operation> -h` lists that operation's flags, required and optional, with types.

## Flags

Flags come from the operation schema in the catalog. Required flags are marked in `-h` output. `--mode`, `--confirmation_token`, `--json`, and `--yes` are reserved and never part of an operation's input.

- `--mode preview|apply` — for side-effecting operations. Preview is the default.
- `--confirmation_token` — the token returned by preview, for `--mode apply`.
- `--yes` — preview, then apply automatically without prompting.
- `--json` — print exactly one raw agent JSON envelope. For side effects, omitted mode defaults to `preview`. `--json --yes` and `--json --mode apply` are rejected before any side effect.

Field values follow the schema type: strings as-is, integers as numbers, objects as JSON:

```bash
siyuan-cli context search_blocks --query roadmap --page 2
siyuan-cli write set_attributes --block_id block-id --attributes '{"custom":"value"}'
```

## Side effects

Side-effecting operations always preview first. Preview resolves the target, validates the request, prints the proposed effect, and shows the confirmation token. It never mutates SiYuan.

On a terminal, the CLI prompts `Apply? [y/N]` after the preview and applies only on `y`. On a non-interactive stdin it refuses to apply unless `--yes` is given. `--yes` still previews first and reuses the previewed confirmation token.

```bash
siyuan-cli write update_block --block_id block-id --content "New text"
# Preview output, then: Apply? [y/N]
```

Read-only operations run immediately.

Markdown and HTML exports write to validated local destinations. PDF and DOCX exports pass `output_path` to SiYuan as the server-side `savePath`. Local destination or upload-source failures use `LOCAL_IO_ERROR`. SiYuan failures use `REMOTE_ERROR`.

## Output

The human path prints readable text or tables. `--json` prints the exact agent envelope:

```bash
siyuan-cli status version
# Server version: 3.7.3

siyuan-cli status version --json
# {"version":"v1","tool":"status","ok":true,"data":{"version":"3.7.3"},...}
```

Failures print the message, the fix, and the error code:

```text
Error: document does not exist
Fix: check notebook and path
Code: NOT_FOUND
```

See [agent-tool-contract.md](agent-tool-contract.md) for the protocol.
