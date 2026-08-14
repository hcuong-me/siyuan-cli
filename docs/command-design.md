# Tool design

The executable exposes six goal-oriented tools plus the `tools` discovery command. `status` and `context` read data. `write`, `organize`, `export`, and `maintain` preview side effects before apply.

Two invocation styles share one request and response core:

- The agent path accepts one top-level tool name and reads one JSON request from stdin. It writes one JSON response and no other output.
- The human path routes `siyuan-cli <tool> <operation> --flag value` through the same tool layer when the second argument is a cataloged operation of the first. It builds the JSON request from the schema, renders the response as text, and supports `--json`, `--yes`, and an interactive apply prompt. Every other invocation delegates unchanged to the agent path.

See [human-interface.md](human-interface.md) for the flag syntax, [agent-tool-contract.md](agent-tool-contract.md) for protocol details, and [legacy-command-baseline.md](legacy-command-baseline.md) for the retired command mapping.

The catalog is the request contract. `search_documents` is unpaginated and rejects `page` and `size`. Side-effect previews expose stable target IDs, fingerprints, and deterministic irreversible-effect identifiers. Human JSON output is preview-only for side effects. Explicit apply remains an agent-path operation with a token.

Notebook and tag selectors accept an ID or an exact name and bind to a stable identity before any API call. Apply re-resolves every selector so the mutation binds to the stable identity that preview fingerprinted.

The tables below map each operation to its invocation and its behavior. Side-effecting operations require `mode: preview`, then `mode: apply` with the returned confirmation token; read operations run immediately.

## status

| Invocation | Description |
|------------|-------------|
| `siyuan-cli status version` | Get SiYuan kernel version |
| `siyuan-cli status time` | Get current system time |
| `siyuan-cli status boot_progress` | Get kernel boot progress |

## context

| Invocation | Description |
|------------|-------------|
| `siyuan-cli context list_notebooks` | List all notebooks |
| `siyuan-cli context resolve_notebook --selector <id-or-name>` | Resolve a notebook selector to a stable identity |
| `siyuan-cli context list_documents --notebook <id-or-name> [--max_depth <n>]` | List a notebook's document tree |
| `siyuan-cli context read_document --notebook <id-or-name> --path <path>` | Read a document as Markdown |
| `siyuan-cli context read_block --block_id <id>` | Get a block's kramdown |
| `siyuan-cli context list_block_children --block_id <id>` | List a block's child blocks |
| `siyuan-cli context get_attributes --block_id <id>` | Get block attributes |
| `siyuan-cli context list_bookmarks` | List bookmark labels |
| `siyuan-cli context list_tags` | List all tags |
| `siyuan-cli context search_blocks --query <keyword> [--page <n> --size <n>]` | Full-text search blocks |
| `siyuan-cli context search_documents --query <keyword> [--notebook <id-or-name> --path <path>]` | Search documents; complete unpaginated result set |
| `siyuan-cli context search_tags --query <keyword>` | Search tags |
| `siyuan-cli context query --statement <sql>` | Execute a SQL query (SELECT only) |

## write

All operations in this tool are side effects: preview, then apply.

| Invocation | Description |
|------------|-------------|
| `siyuan-cli write create_document --notebook <id-or-name> --path <path> --content <md>` | Create a document from Markdown |
| `siyuan-cli write update_document --notebook <id-or-name> --path <path> --content <md>` | Replace a document's content (replace semantics) |
| `siyuan-cli write update_block --block_id <id> --content <md>` | Replace a block's content |
| `siyuan-cli write append_block --parent_id <id> --content <md>` | Insert a block at the end of a parent |
| `siyuan-cli write insert_after_block --previous_id <id> --content <md>` | Insert a block after another block |
| `siyuan-cli write set_attribute --block_id <id> --key <k> --value <v>` | Set one attribute |
| `siyuan-cli write set_attributes --block_id <id> --attributes '<json>'` | Set multiple attributes (JSON object) |
| `siyuan-cli write reset_attribute --block_id <id> --key <k>` | Reset an attribute |

## organize

All operations in this tool are side effects: preview, then apply.

| Invocation | Description |
|------------|-------------|
| `siyuan-cli organize create_notebook --name <name>` | Create a notebook |
| `siyuan-cli organize rename_notebook --notebook <id-or-name> --name <name>` | Rename a notebook |
| `siyuan-cli organize remove_notebook --notebook <id-or-name>` | Remove a notebook (destructive) |
| `siyuan-cli organize open_notebook --notebook <id-or-name>` | Open a closed notebook |
| `siyuan-cli organize close_notebook --notebook <id-or-name>` | Close an open notebook |
| `siyuan-cli organize remove_document --notebook <id-or-name> --path <path>` | Remove a document (destructive) |
| `siyuan-cli organize delete_block --block_id <id>` | Delete a block and its children (destructive) |
| `siyuan-cli organize move_block --block_id <id> --parent_id <id> [--previous_id <id>]` | Move a block |
| `siyuan-cli organize rename_tag --tag <label> --name <label>` | Rename a tag |
| `siyuan-cli organize remove_tag --tag <label>` | Remove a tag (destructive) |

## export

| Invocation | Description |
|------------|-------------|
| `siyuan-cli export preview_document --document_id <id>` | Read. Preview a document; a bad ID returns `NOT_FOUND` |
| `siyuan-cli export export_markdown --document_id <id> --output_path <path>` | Side effect. Export Markdown to a local file |
| `siyuan-cli export export_html --document_id <id> --output_path <path>` | Side effect. Export HTML to a local file |
| `siyuan-cli export export_pdf --document_id <id> --output_path <path>` | Side effect. Export PDF; `output_path` is the server-side `savePath` |
| `siyuan-cli export export_docx --document_id <id> --output_path <path>` | Side effect. Export DOCX; `output_path` is the server-side `savePath` |

## maintain

Mutation operations in this tool are side effects: preview, then apply.

| Invocation | Description |
|------------|-------------|
| `siyuan-cli maintain list_templates` | List templates |
| `siyuan-cli maintain get_template --path <path>` | Get template content |
| `siyuan-cli maintain render_template --document_id <id> --path <path>` | Render a template to a document |
| `siyuan-cli maintain remove_template --path <path>` | Remove a template (destructive) |
| `siyuan-cli maintain list_snapshots` | List repository snapshots |
| `siyuan-cli maintain create_snapshot --name <name>` | Create a snapshot |
| `siyuan-cli maintain restore_snapshot --snapshot_id <id>` | Restore repository state (destructive) |
| `siyuan-cli maintain upload_asset --path <local-file>` | Upload a local file as an asset |
| `siyuan-cli maintain list_unused_assets` | List unused assets |
| `siyuan-cli maintain clean_unused_assets` | Remove unused assets (destructive) |
| `siyuan-cli maintain read_tree --path <path>` | List directory contents |
| `siyuan-cli maintain read_file --path <path>` | Read file content |
| `siyuan-cli maintain write_file --path <path> --content <text>` | Write file content |
| `siyuan-cli maintain make_directory --path <path>` | Create a directory |
| `siyuan-cli maintain remove_file --path <path>` | Remove a file or directory (destructive) |
| `siyuan-cli maintain rename_file --old_path <path> --new_path <path>` | Rename or move a file |

## Summary by tool

| Tool | Operations |
|------|------------|
| status | 3 |
| context | 13 |
| write | 8 |
| organize | 10 |
| export | 5 |
| maintain | 16 |
