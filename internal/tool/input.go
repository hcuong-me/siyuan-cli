package tool

import (
	"encoding/json"
	"fmt"
	"io"
)

// DecodeRequest reads exactly one strict JSON request.
func DecodeRequest(reader io.Reader) (Request, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var request Request
	if err := decoder.Decode(&request); err != nil {
		return Request{}, fmt.Errorf("decode request: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Request{}, fmt.Errorf("request must contain one JSON object")
		}
		return Request{}, fmt.Errorf("read request: %w", err)
	}
	return request, nil
}
