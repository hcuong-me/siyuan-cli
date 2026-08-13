package tool

import (
	"encoding/json"
	"fmt"
	"sort"
)

const (
	// InvalidRequest identifies malformed or schema-invalid input.
	InvalidRequest = "INVALID_REQUEST"
	// MissingConfig identifies absent SiYuan connection configuration.
	MissingConfig = "MISSING_CONFIG"
	// NotFound identifies a missing selected resource.
	NotFound = "NOT_FOUND"
	// AmbiguousSelector identifies a selector with multiple matches.
	AmbiguousSelector = "AMBIGUOUS_SELECTOR"
	// ConfirmationRequired identifies an apply request without a preview token.
	ConfirmationRequired = "CONFIRMATION_REQUIRED"
	// ConfirmationStale identifies a request or target that changed after preview.
	ConfirmationStale = "CONFIRMATION_STALE"
	// Conflict identifies a remote concurrent-change conflict.
	Conflict = "CONFLICT"
	// RemoteError identifies a SiYuan server or transport failure.
	RemoteError = "REMOTE_ERROR"
	// LocalIOError identifies a local file read or write failure.
	LocalIOError = "LOCAL_IO_ERROR"
	// PreconditionUnavailable identifies a side effect that cannot be safely
	// previewed because its current state is not inspectable.
	PreconditionUnavailable = "PRECONDITION_UNAVAILABLE"
)

// Error gives an agent enough information to repair or retry a request.
type Error struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Fix        string      `json:"fix"`
	Retryable  bool        `json:"retryable"`
	Candidates []Candidate `json:"candidates,omitempty"`
}

// Candidate identifies one stable selector match. Name and Path are optional
// because different SiYuan endpoints expose different amounts of context.
type Candidate struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// Error implements error so strict accessors can return protocol-native
// invalid-request errors without silently coercing malformed values.
func (e Error) Error() string { return e.Message }

func invalidInputTypeError(field, expected string) *Error {
	return &Error{
		Code:      InvalidRequest,
		Message:   fmt.Sprintf("input.%s must be a JSON %s", field, expected),
		Fix:       "send input values that match the operation schema",
		Retryable: false,
	}
}

// normalizeError fills safe, stable metadata without changing the existing
// error envelope. Candidates are sorted so callers receive deterministic JSON
// regardless of endpoint iteration order.
func normalizeError(toolError *Error) *Error {
	if toolError == nil {
		return nil
	}
	normalized := *toolError
	if len(normalized.Candidates) > 0 {
		normalized.Candidates = append([]Candidate(nil), normalized.Candidates...)
		sort.SliceStable(normalized.Candidates, func(i, j int) bool {
			left, right := normalized.Candidates[i], normalized.Candidates[j]
			if left.ID != right.ID {
				return left.ID < right.ID
			}
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.Path < right.Path
		})
	}
	return &normalized
}

// MarshalJSON keeps metadata additive for errors constructed directly by
// handlers (rather than through errorResponse).
func (e Error) MarshalJSON() ([]byte, error) {
	normalized := normalizeError(&e)
	type errorAlias Error
	return json.Marshal(errorAlias(*normalized))
}

func errorResponse(tool, code, message, fix string, retryable bool) Response {
	toolError := normalizeError(&Error{
		Code:      code,
		Message:   message,
		Fix:       fix,
		Retryable: retryable,
	})
	return Response{
		Version:     ProtocolVersion,
		Tool:        tool,
		OK:          false,
		Summary:     message,
		Warnings:    []Warning{},
		NextActions: defaultNextActions(tool, "", code),
		Error:       toolError,
	}
}

func defaultNextActions(tool, operation, code string) []NextAction {
	switch code {
	case InvalidRequest:
		return []NextAction{{Tool: "tools", Reason: "inspect the operation schema and correct the request"}}
	case ConfirmationRequired, ConfirmationStale:
		if operation != "" {
			return []NextAction{{Tool: tool, Operation: operation, Reason: "run the same operation with mode preview and use its confirmation token"}}
		}
	case AmbiguousSelector:
		if operation != "" {
			return []NextAction{{Tool: tool, Operation: operation, Reason: "select one candidate ID, then run the operation again"}}
		}
	case NotFound:
		if operation != "" {
			return []NextAction{{Tool: tool, Operation: operation, Reason: "check the selector or path and retry"}}
		}
	case PreconditionUnavailable, LocalIOError:
		if operation != "" {
			return []NextAction{{Tool: tool, Operation: operation, Reason: "fix the reported precondition, then run preview again"}}
		}
	case Conflict:
		if operation != "" {
			return []NextAction{{Tool: tool, Operation: operation, Reason: "choose a distinct target name or path, then run preview again"}}
		}
	}
	return []NextAction{}
}
