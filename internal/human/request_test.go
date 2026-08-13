package human

import (
	"encoding/json"
	"strings"
	"testing"

	"siyuan/internal/tool"
)

func TestBuildRequestFromFlags(t *testing.T) {
	catalog := tool.Catalog()
	cases := []struct {
		name      string
		tool      string
		operation string
		args      []string
		wantInput string
	}{
		{
			name:      "required flags only",
			tool:      "context",
			operation: "read_document",
			args:      []string{"--notebook", "n", "--path", "/Projects/Roadmap"},
			wantInput: `{"notebook":"n","path":"/Projects/Roadmap"}`,
		},
		{
			name:      "empty input for no-field operations",
			tool:      "context",
			operation: "list_notebooks",
			wantInput: `{}`,
		},
		{
			name:      "integer flag",
			tool:      "context",
			operation: "search_blocks",
			args:      []string{"--query", "roadmap", "--page", "2"},
			wantInput: `{"page":2,"query":"roadmap"}`,
		},
		{
			name:      "object flag",
			tool:      "write",
			operation: "set_attributes",
			args:      []string{"--block_id", "b", "--attributes", `{"custom":"x"}`},
			wantInput: `{"attributes":{"custom":"x"},"block_id":"b"}`,
		},
		{
			name:      "reserved flags map onto the envelope",
			tool:      "write",
			operation: "update_block",
			args:      []string{"--block_id", "b", "--content", "x", "--mode", "preview", "--json", "--yes"},
			wantInput: `{"block_id":"b","content":"x"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := buildRequest(tc.tool, tc.operation, catalog[tc.tool].Operations[tc.operation], tc.args)
			if err != nil {
				t.Fatalf("buildRequest() error = %v", err)
			}
			if parsed.request.Version != tool.ProtocolVersion || parsed.request.Operation != tc.operation {
				t.Fatalf("request = %+v", parsed.request)
			}
			if got := string(parsed.request.Input); got != tc.wantInput {
				t.Fatalf("input = %s, want %s", got, tc.wantInput)
			}
			if tc.name == "reserved flags map onto the envelope" {
				if parsed.request.Mode != "preview" || !parsed.json || !parsed.yes {
					t.Fatalf("parsed = %+v", parsed)
				}
			}
		})
	}
}

func TestBuildRequestErrors(t *testing.T) {
	catalog := tool.Catalog()
	cases := []struct {
		name      string
		tool      string
		operation string
		args      []string
		want      string
	}{
		{"missing required flag", "context", "read_document", []string{"--notebook", "n"}, "missing required flag --path"},
		{"unknown flag", "context", "read_document", []string{"--nope", "x"}, "unknown flag --nope"},
		{"positional argument", "context", "read_document", []string{"--notebook", "n", "--path", "/p", "extra"}, "unexpected argument"},
		{"missing value", "context", "read_document", []string{"--notebook"}, "requires a value"},
		{"bad integer", "context", "search_blocks", []string{"--query", "q", "--page", "x"}, "expects an integer"},
		{"bad object", "write", "set_attributes", []string{"--block_id", "b", "--attributes", "nope"}, "expects a JSON object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildRequest(tc.tool, tc.operation, catalog[tc.tool].Operations[tc.operation], tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("buildRequest() error = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestRequestMatchesAgentEnvelope(t *testing.T) {
	catalog := tool.Catalog()
	parsed, err := buildRequest("context", "read_document", catalog["context"].Operations["read_document"],
		[]string{"--notebook", "n", "--path", "/p"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(parsed.request)
	if err != nil {
		t.Fatal(err)
	}
	agent := `{"version":"v1","operation":"read_document","input":{"notebook":"n","path":"/p"}}`
	var got, want map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(agent), &want); err != nil {
		t.Fatal(err)
	}
	if got["version"] != want["version"] || got["operation"] != want["operation"] {
		t.Fatalf("request = %s, want %s", encoded, agent)
	}
	gotInput, _ := json.Marshal(got["input"])
	wantInput, _ := json.Marshal(want["input"])
	if string(gotInput) != string(wantInput) {
		t.Fatalf("input = %s, want %s", gotInput, wantInput)
	}
}
