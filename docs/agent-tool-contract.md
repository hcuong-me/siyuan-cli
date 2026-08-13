# Agent tool contract

Run `siyuan-cli tools` to discover the current schemas. It takes no stdin.

All other tools read exactly one request:

```json
{"version":"v1","operation":"read_document","input":{"notebook":"notebook-id","path":"/Projects/Roadmap"}}
```

The response always has `version`, `tool`, `ok`, `summary`, `warnings`, `next_actions`, and `error`. A failed response has an error with `code`, `message`, `fix`, and `retryable`. Selector errors can also include sorted `candidates` with stable `id` values and optional `name` and `path`.

Read-only operations run immediately. Side-effecting operations require `mode` to be `preview` or `apply`. Preview returns a confirmation token and target fingerprints. Apply must use that token with the same request. Preview issues a token only when every target has a concrete, re-resolvable state precondition. When state cannot be inspected, the operation fails closed with `PRECONDITION_UNAVAILABLE`.

Notebook and tag selectors accept an ID or an exact name. Every selector binds to a stable identity before any API call: no match is `NOT_FOUND` with a fix, and multiple matches are `AMBIGUOUS_SELECTOR` with sorted `candidates`. Apply re-resolves the selector so the mutation binds to the same stable identity that preview fingerprinted. A disappeared or renamed target is `CONFIRMATION_STALE` before any mutation.

The token records deliberate intent and unchanged request targets. It is not an authorization credential. It has no expiry, but it is stale when the request or its target identity or content changes. A disappeared target is also `CONFIRMATION_STALE`. Apply runs only on a fresh check.

`preview.irreversible_effects` contains deterministic operation identifiers such as `delete_document:<id>` and `overwrite_local_file:<canonical-path>`. Export preview validates the owner of the destination: Markdown and HTML are local files. PDF and DOCX receive a server-side `savePath`. Preview never writes a local file or asks SiYuan to generate PDF or DOCX output.

`context.search_documents` is intentionally unpaginated. It returns the complete result set and a `count`; `page` and `size` are not valid fields.
