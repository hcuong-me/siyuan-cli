# Tool design

The executable exposes six goal-oriented tools plus the `tools` discovery command. `status` and `context` read data. `write`, `organize`, `export`, and `maintain` preview side effects before apply.

Two invocation styles share one request and response core:

- The agent path accepts one top-level tool name and reads one JSON request from stdin. It writes one JSON response and no other output.
- The human path routes `siyuan-cli <tool> <operation> --flag value` through the same tool layer when the second argument is a cataloged operation of the first. It builds the JSON request from the schema, renders the response as text, and supports `--json`, `--yes`, and an interactive apply prompt. Every other invocation delegates unchanged to the agent path.

See [human-interface.md](human-interface.md) for the flag syntax, [agent-tool-contract.md](agent-tool-contract.md) for protocol details, and [legacy-command-baseline.md](legacy-command-baseline.md) for the retired command mapping.

The catalog is the request contract. `search_documents` is unpaginated and rejects `page` and `size`. Side-effect previews expose stable target IDs, fingerprints, and deterministic irreversible-effect identifiers. Human JSON output is preview-only for side effects. Explicit apply remains an agent-path operation with a token.

Notebook and tag selectors accept an ID or an exact name and bind to a stable identity before any API call. Apply re-resolves every selector so the mutation binds to the stable identity that preview fingerprinted. `list_documents` returns the resolved notebook, a node count, and the document tree. `read_document`, `update_document`, and `remove_document` resolve the notebook the same way.
