package tool

// Preflight returns a confirmation for preview, or validates one for apply.
func Preflight(tool string, request Request, targets []Target) (*Confirmation, *Error) {
	token, err := ConfirmationToken(tool, request, targets)
	if err != nil {
		return nil, &Error{Code: InvalidRequest, Message: "input cannot be canonicalized", Fix: "send JSON input that matches the operation schema", Retryable: false}
	}

	switch request.Mode {
	case "preview":
		return &Confirmation{Token: token, Note: "This token records deliberate intent and unchanged target state. It is not permission control and has no expiry."}, nil
	case "apply":
		if request.ConfirmationToken == "" {
			return nil, &Error{Code: ConfirmationRequired, Message: "apply requires the confirmation token from preview", Fix: "run the same request with mode preview, then send its token with mode apply", Retryable: true}
		}
		if request.ConfirmationToken != token {
			return nil, &Error{Code: ConfirmationStale, Message: "the request or target state changed after preview", Fix: "run preview again and use the new token", Retryable: true}
		}
		return nil, nil
	default:
		return nil, &Error{Code: InvalidRequest, Message: "side-effecting operations require mode preview or apply", Fix: "set mode to preview or apply", Retryable: false}
	}
}
