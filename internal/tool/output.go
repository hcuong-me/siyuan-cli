package tool

import (
	"encoding/json"
	"io"
)

// WriteResponse writes one JSON object and a newline.
func WriteResponse(writer io.Writer, response Response) error {
	return json.NewEncoder(writer).Encode(response)
}

// RecoverResponse converts a recovered panic into the protocol envelope.
func RecoverResponse(tool string, _ any) Response {
	return errorResponse(tool, RemoteError, "tool execution stopped unexpectedly", "retry the request; if it fails again, inspect the server logs", true)
}
