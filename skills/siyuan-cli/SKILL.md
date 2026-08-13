---
name: siyuan-cli
description: Use the local siyuan-cli tools for SiYuan note workflows.
---

# siyuan-cli

Run `siyuan-cli tools` first. It returns the current schemas for `status`, `context`, `write`, `organize`, `export`, and `maintain`.

## Agent path (JSON on stdin)

Each operational request is one JSON object on stdin:

```json
{"version":"v1","operation":"read_document","input":{"notebook":"notebook-id","path":"/Projects/Roadmap"}}
```

Run it with the matching tool:

```bash
printf '%s' '<request-json>' | siyuan-cli context
```

For a mutation, send `"mode":"preview"` first. Read `data.confirmation.token`, then resend the same request with `"mode":"apply"` and `"confirmation_token":"<token>"`. Do not apply a mutation without reviewing its preview.

Every response is JSON. On failure, inspect `error.code`, `error.message`, `error.fix`, and `error.retryable`. Set `SIYUAN_TOKEN` and, when required, `SIYUAN_BASE_URL` before an operational request.

## Human path (flags)

The same operations accept flags. `siyuan-cli <tool> <operation> -h` prints the schema-driven flag list.

```bash
# Read a document
siyuan-cli context read_document --notebook notebook-id --path /Projects/Roadmap

# Search blocks
siyuan-cli context search_blocks --query roadmap

# List notebooks
siyuan-cli context list_notebooks

# Server state
siyuan-cli status version
```

Side-effecting operations preview first and prompt `Apply? [y/N]`. `--yes` previews and applies in one step:

```bash
siyuan-cli write update_block --block_id block-id --content "New text"
siyuan-cli write update_block --block_id block-id --content "New text" --yes
```

`--json` prints exactly one raw agent envelope. For side effects it defaults to preview and never applies; `--json --yes` and `--json --mode apply` are rejected:

```bash
siyuan-cli context list_notebooks --json
```

`context.search_documents` is unpaginated and returns every result with a count. Do not send `page` or `size`.

See [human-interface.md](references/human-interface.md) for the flag syntax and [agent-tool-contract.md](references/agent-tool-contract.md) for the agent protocol.
