# Quality

Run `go test ./...`, `go vet ./...`, and `make lint-arch` before review.

Keep all tool output in the response envelope. Validate new operation input in `internal/tool/schema.go`. A side effect must preview without an HTTP mutation and require a confirmation token before apply.

Add tests for malformed requests, envelope errors, previews, applies, and stale tokens when you add a state-changing operation.

Every declared input field is type-checked at the protocol boundary. Present `null` is invalid, required strings are non-empty, and integer fields must be finite, integral, and in range. Unknown fields are rejected.

Every side-effect preview must contain concrete targets and non-empty `irreversible_effects`. A missing or unreadable precondition returns `PRECONDITION_UNAVAILABLE` without a token. Apply re-resolves targets and returns `CONFIRMATION_STALE` before mutation when state changed or disappeared.
