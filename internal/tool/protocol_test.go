package tool

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeRequest(t *testing.T) {
	request, err := DecodeRequest(strings.NewReader(`{"version":"v1","operation":"version","input":{}}`))
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}
	if request.Operation != "version" {
		t.Fatalf("operation = %q, want version", request.Operation)
	}

	for _, input := range []string{"", `{`, `{"version":"v1","operation":"version","input":{},"extra":true}`, `{} {}`} {
		if _, err := DecodeRequest(strings.NewReader(input)); err == nil {
			t.Fatalf("DecodeRequest(%q) succeeded", input)
		}
	}
}

func TestSchemaValidation(t *testing.T) {
	valid := Request{Version: ProtocolVersion, Operation: "create_document", Input: json.RawMessage(`{"notebook":"n","path":"p","content":"x"}`), Mode: "preview"}
	if err := ValidateRequest("write", valid); err != nil {
		t.Fatalf("ValidateRequest() error = %v", err)
	}

	for _, request := range []Request{
		{Version: "v2", Operation: "version", Input: json.RawMessage(`{}`)},
		{Version: ProtocolVersion, Operation: "missing", Input: json.RawMessage(`{}`)},
		{Version: ProtocolVersion, Operation: "create_document", Input: json.RawMessage(`{"notebook":"n"}`), Mode: "preview"},
		{Version: ProtocolVersion, Operation: "create_document", Input: json.RawMessage(`{"notebook":"n","path":"p","content":"x","extra":true}`), Mode: "preview"},
		{Version: ProtocolVersion, Operation: "create_document", Input: json.RawMessage(`{"notebook":"n","path":"p","content":"x"}`)},
	} {
		if err := ValidateRequest("write", request); err == nil {
			t.Fatalf("ValidateRequest(%+v) succeeded", request)
		}
	}
}

func TestSearchDocumentsIsUnpaginated(t *testing.T) {
	schema := Catalog()["context"].Operations["search_documents"]
	if _, ok := schema.Input["page"]; ok {
		t.Fatal("search_documents still advertises page")
	}
	if _, ok := schema.Input["size"]; ok {
		t.Fatal("search_documents still advertises size")
	}
	request := Request{Version: ProtocolVersion, Operation: "search_documents", Input: json.RawMessage(`{"query":"roadmap","page":1}`)}
	if err := ValidateRequest("context", request); err == nil {
		t.Fatal("pagination field was accepted")
	}
}

func TestCatalogExamplesAndRiskMetadata(t *testing.T) {
	for toolName, schema := range Catalog() {
		for operationName, operation := range schema.Operations {
			if strings.TrimSpace(operation.Description) == "" || strings.TrimSpace(operation.Risk) == "" {
				t.Errorf("%s.%s has incomplete description/risk metadata", toolName, operationName)
			}
			if operation.Example == nil || operation.Example["operation"] != operationName {
				t.Errorf("%s.%s has no representative example: %#v", toolName, operationName, operation.Example)
			}
			if operation.SideEffect && operation.Example["mode"] != "preview" {
				t.Errorf("%s.%s example does not show preview mode", toolName, operationName)
			}
		}
	}
}

func TestCatalogApplicabilityAndExamplesAreRepresentative(t *testing.T) {
	catalog := Catalog()
	genericRead := "Use for read-only inspection; it does not change SiYuan."
	genericWrite := "Use when the requested change is intentional and the resolved target is known."
	boundaried := map[string]string{
		"list_documents":      "context",
		"search_documents":    "context",
		"query":               "context",
		"create_document":     "write",
		"remove_notebook":     "organize",
		"restore_snapshot":    "maintain",
		"export_markdown":     "export",
		"clean_unused_assets": "maintain",
	}
	for operation, toolName := range boundaried {
		op := catalog[toolName].Operations[operation]
		if op.Applicability == genericRead || op.Applicability == genericWrite || strings.TrimSpace(op.Applicability) == "" {
			t.Errorf("%s.%s uses generic applicability", toolName, operation)
		}
	}
	notebook := exampleInput(catalog["context"].Operations["list_documents"].Input)["notebook"].(string)
	if !strings.HasPrefix(notebook, "20") || len(notebook) != 22 {
		t.Errorf("notebook example %q is not a representative SiYuan ID", notebook)
	}
	blockID := exampleInput(catalog["context"].Operations["read_block"].Input)["block_id"].(string)
	if !strings.HasPrefix(blockID, "20") || len(blockID) != 22 {
		t.Errorf("block_id example %q is not a representative SiYuan ID", blockID)
	}
}

func TestSchemaRejectsWrongDeclaredTypes(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		operation string
		input     string
		valid     bool
	}{
		{name: "required string", tool: "write", operation: "create_document", input: `{"notebook":"n","path":"p","content":"x"}`, valid: true},
		{name: "required string null", tool: "write", operation: "create_document", input: `{"notebook":null,"path":"p","content":"x"}`},
		{name: "required string number", tool: "write", operation: "create_document", input: `{"notebook":1,"path":"p","content":"x"}`},
		{name: "required string array", tool: "write", operation: "create_document", input: `{"notebook":[],"path":"p","content":"x"}`},
		{name: "required string object", tool: "write", operation: "create_document", input: `{"notebook":{},"path":"p","content":"x"}`},
		{name: "required string empty", tool: "write", operation: "create_document", input: `{"notebook":"","path":"p","content":"x"}`},
		{name: "optional string empty", tool: "context", operation: "search_documents", input: `{"query":"q","notebook":"","path":""}`, valid: true},
		{name: "optional string null", tool: "context", operation: "search_documents", input: `{"query":"q","notebook":null}`},
		{name: "integer", tool: "context", operation: "search_blocks", input: `{"query":"q","page":1}`, valid: true},
		{name: "integral decimal", tool: "context", operation: "search_blocks", input: `{"query":"q","page":1.0}`, valid: true},
		{name: "integral exponent", tool: "context", operation: "search_blocks", input: `{"query":"q","page":1e2}`, valid: true},
		{name: "fractional integer", tool: "context", operation: "search_blocks", input: `{"query":"q","page":1.5}`},
		{name: "overflow integer", tool: "context", operation: "search_blocks", input: `{"query":"q","page":9223372036854775808}`},
		{name: "underflow integer", tool: "context", operation: "search_blocks", input: `{"query":"q","page":-9223372036854775809}`},
		{name: "non-finite integer", tool: "context", operation: "search_blocks", input: `{"query":"q","page":NaN}`},
		{name: "integer string", tool: "context", operation: "search_blocks", input: `{"query":"q","page":"1"}`},
		{name: "integer null", tool: "context", operation: "search_blocks", input: `{"query":"q","page":null}`},
		{name: "object", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":{"key":"value"}}`, valid: true},
		{name: "object empty", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":{}}`, valid: true},
		{name: "object member number", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":{"key":1}}`},
		{name: "object null", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":null}`},
		{name: "object array", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":[]}`},
		{name: "object scalar", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":"{}"}`},
		{name: "malformed object", tool: "write", operation: "set_attributes", input: `{"block_id":"b","attributes":{"key":}}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := Request{Version: ProtocolVersion, Operation: test.operation, Input: json.RawMessage(test.input), Mode: "preview"}
			err := ValidateRequest(test.tool, request)
			if test.valid && err != nil {
				t.Fatalf("ValidateRequest() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatalf("ValidateRequest() accepted %s", test.input)
			}
		})
	}
}

func TestStrictTypedAccessors(t *testing.T) {
	input, err := decodeInput(Request{Input: json.RawMessage(`{"name":"n","count":1.0,"attrs":{}}`)})
	if err != nil {
		t.Fatalf("decodeInput() error = %v", err)
	}
	if value, err := stringInputStrict(input, "name"); err != nil || value != "n" {
		t.Fatalf("stringInputStrict() = %q, %v", value, err)
	}
	if value, err := integerInputStrict(input, "count"); err != nil || value != 1 {
		t.Fatalf("integerInputStrict() = %d, %v", value, err)
	}
	if value, err := objectInputStrict(input, "attrs"); err != nil || value == nil {
		t.Fatalf("objectInputStrict() = %#v, %v", value, err)
	}
	for name, want := range map[string]string{"name": "string", "count": "integer", "attrs": "object"} {
		var value any
		var err error
		switch want {
		case "string":
			value, err = stringInputStrict(map[string]any{name: 1}, name)
		case "integer":
			value, err = integerInputStrict(map[string]any{name: "1"}, name)
		case "object":
			value, err = objectInputStrict(map[string]any{name: []any{}}, name)
		}
		if err == nil {
			t.Errorf("%s accessor accepted wrong type with value %#v", want, value)
		}
		toolError, ok := err.(*Error)
		if !ok || toolError.Code != InvalidRequest {
			t.Errorf("%s accessor error = %#v, want INVALID_REQUEST", want, err)
		}
	}
}

func TestMalformedCreateDocumentRejectedBeforeDispatch(t *testing.T) {
	var output bytes.Buffer
	request := `{"version":"v1","operation":"create_document","input":{"notebook":null,"path":"p","content":"x"},"mode":"preview"}`
	if code := Execute([]string{"write"}, strings.NewReader(request), &output); code == 0 {
		t.Fatalf("malformed create_document request succeeded: %s", output.String())
	}
	assertErrorJSON(t, output.Bytes(), InvalidRequest)
}

func TestErrorCandidatesAreStableAndAdditive(t *testing.T) {
	response := errorResponse("context", AmbiguousSelector, "multiple matches", "choose one", false)
	response.Error.Candidates = []Candidate{{ID: "b", Name: "Beta"}, {ID: "a", Path: "/A"}}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Error == nil || decoded.Error.Code != AmbiguousSelector {
		t.Fatalf("error metadata = %#v", decoded.Error)
	}
	if got := decoded.Error.Candidates; len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" || got[0].Path != "/A" || got[1].Name != "Beta" {
		t.Fatalf("candidates = %#v", got)
	}
}

func TestResponseEnvelope(t *testing.T) {
	var output bytes.Buffer
	if err := WriteResponse(&output, success("status", "server version", map[string]string{"version": "x"})); err != nil {
		t.Fatalf("WriteResponse() error = %v", err)
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if !response.OK || response.Error != nil || response.Tool != "status" {
		t.Fatalf("response = %+v", response)
	}
}

func TestErrorEnvelope(t *testing.T) {
	response := errorResponse("context", InvalidRequest, "bad input", "fix input", false)
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("error response is not JSON: %v", err)
	}
	errorData := decoded["error"].(map[string]any)
	for _, name := range []string{"code", "message", "fix", "retryable"} {
		if _, ok := errorData[name]; !ok {
			t.Errorf("missing error.%s", name)
		}
	}
}

func TestPreviewDoesNotApply(t *testing.T) {
	request := Request{Version: ProtocolVersion, Operation: "delete_block", Input: json.RawMessage(`{"block_id":"b"}`), Mode: "preview"}
	confirmation, toolError := Preflight("organize", request, []Target{{ID: "b", Fingerprint: "one"}})
	if toolError != nil || confirmation == nil || confirmation.Token == "" {
		t.Fatalf("Preflight(preview) = %#v, %#v", confirmation, toolError)
	}
}

func TestConfirmationToken(t *testing.T) {
	request := Request{Version: ProtocolVersion, Operation: "delete_block", Input: json.RawMessage(`{"block_id":"b"}`), Mode: "preview"}
	preview, toolError := Preflight("organize", request, []Target{{ID: "b", Fingerprint: "one"}})
	if toolError != nil {
		t.Fatalf("preview error = %#v", toolError)
	}
	request.Mode = "apply"
	request.ConfirmationToken = preview.Token
	if _, toolError = Preflight("organize", request, []Target{{ID: "b", Fingerprint: "one"}}); toolError != nil {
		t.Fatalf("apply error = %#v", toolError)
	}
}

func TestStaleConfirmation(t *testing.T) {
	request := Request{Version: ProtocolVersion, Operation: "delete_block", Input: json.RawMessage(`{"block_id":"b"}`), Mode: "preview"}
	preview, _ := Preflight("organize", request, []Target{{ID: "b", Fingerprint: "one"}})
	request.Mode = "apply"
	request.ConfirmationToken = preview.Token
	if _, toolError := Preflight("organize", request, []Target{{ID: "b", Fingerprint: "two"}}); toolError == nil || toolError.Code != ConfirmationStale {
		t.Fatalf("stale Preflight() error = %#v", toolError)
	}
}

func TestPanicEnvelope(t *testing.T) {
	response := RecoverResponse("context", "boom")
	if response.OK || response.Error == nil || response.Error.Code != RemoteError {
		t.Fatalf("RecoverResponse() = %#v", response)
	}
}
