// Package tool defines the agent-oriented SiYuan command protocol.
package tool

import "encoding/json"

// ProtocolVersion is the current JSON tool protocol version.
const ProtocolVersion = "v1"

// Request is one operational tool request read from standard input.
type Request struct {
	Version           string          `json:"version"`
	Operation         string          `json:"operation"`
	Input             json.RawMessage `json:"input"`
	Mode              string          `json:"mode,omitempty"`
	ConfirmationToken string          `json:"confirmation_token,omitempty"`
}

// Response is the only stdout format for tool success and failure.
type Response struct {
	Version     string       `json:"version"`
	Tool        string       `json:"tool"`
	OK          bool         `json:"ok"`
	Data        any          `json:"data,omitempty"`
	Summary     string       `json:"summary"`
	Warnings    []Warning    `json:"warnings"`
	NextActions []NextAction `json:"next_actions"`
	Error       *Error       `json:"error"`
}

// Warning identifies a non-fatal condition in a response.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NextAction suggests one concrete follow-up request.
type NextAction struct {
	Tool      string `json:"tool"`
	Operation string `json:"operation"`
	Reason    string `json:"reason"`
}

// Preview describes the effect proposed by a side-effecting operation.
type Preview struct {
	Targets             []Target `json:"targets"`
	Changes             any      `json:"changes"`
	IrreversibleEffects []string `json:"irreversible_effects"`
}

// Confirmation binds an apply request to the previewed target state.
type Confirmation struct {
	Token string `json:"token"`
	Note  string `json:"note"`
}

// Target is a resolved item and the fingerprint used for confirmation.
type Target struct {
	ID          string            `json:"id"`
	Fingerprint string            `json:"fingerprint"`
	Details     map[string]string `json:"details,omitempty"`
}

func success(tool, summary string, data any) Response {
	return Response{
		Version:     ProtocolVersion,
		Tool:        tool,
		OK:          true,
		Data:        data,
		Summary:     summary,
		Warnings:    []Warning{},
		NextActions: []NextAction{},
		Error:       nil,
	}
}
