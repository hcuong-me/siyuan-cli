package tool

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// CanonicalRequest returns the stable request form shared by preview and apply.
func CanonicalRequest(tool string, request Request) ([]byte, error) {
	var input any
	decoder := json.NewDecoder(bytes.NewReader(request.Input))
	decoder.UseNumber()
	if err := decoder.Decode(&input); err != nil {
		return nil, fmt.Errorf("decode input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("input must contain one JSON value")
		}
		return nil, fmt.Errorf("read input: %w", err)
	}
	return json.Marshal(struct {
		Version   string `json:"version"`
		Tool      string `json:"tool"`
		Operation string `json:"operation"`
		Input     any    `json:"input"`
	}{ProtocolVersion, tool, request.Operation, input})
}

// ConfirmationToken returns a deterministic token for the request and target state.
func ConfirmationToken(tool string, request Request, targets []Target) (string, error) {
	canonical, err := CanonicalRequest(tool, request)
	if err != nil {
		return "", err
	}
	ordered := append([]Target(nil), targets...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ID == ordered[j].ID {
			return ordered[i].Fingerprint < ordered[j].Fingerprint
		}
		return ordered[i].ID < ordered[j].ID
	})
	payload, err := json.Marshal(struct {
		Request []byte   `json:"request"`
		Targets []Target `json:"targets"`
	}{canonical, ordered})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}
