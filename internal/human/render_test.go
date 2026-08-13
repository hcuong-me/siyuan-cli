package human

import (
	"strings"
	"testing"

	"siyuan/internal/tool"
)

func TestRenderStatusVersion(t *testing.T) {
	response := tool.Response{Version: "v1", Tool: "status", OK: true, Data: map[string]string{"version": "3.2.0"}, Summary: "server version"}
	text := Render(response, "version")
	if !strings.Contains(text, "3.2.0") {
		t.Fatalf("text = %q", text)
	}
}

func TestRenderError(t *testing.T) {
	response := tool.Response{
		Version: "v1", Tool: "context", OK: false,
		Error: &tool.Error{Code: tool.NotFound, Message: "document does not exist", Fix: "check notebook and path", Retryable: false},
	}
	text := Render(response, "read_document")
	for _, want := range []string{"Error:", "document does not exist", "Fix:", "check notebook and path", "NOT_FOUND"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text = %q, missing %q", text, want)
		}
	}
}

func TestRenderReadDocument(t *testing.T) {
	response := tool.Response{Version: "v1", Tool: "context", OK: true, Data: map[string]any{
		"id": "d", "notebook": "n", "path": "/Projects/Roadmap",
		"document": map[string]any{"hPath": "/Projects/Roadmap", "content": "Hello"},
	}, Summary: "document content"}
	text := Render(response, "read_document")
	for _, want := range []string{"d", "/Projects/Roadmap", "Hello"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text = %q, missing %q", text, want)
		}
	}
}

func TestRenderPreview(t *testing.T) {
	response := tool.Response{Version: "v1", Tool: "write", OK: true, Data: map[string]any{
		"preview": map[string]any{
			"targets":              []any{map[string]any{"id": "block:b", "fingerprint": "abc"}},
			"changes":              map[string]any{"block_id": "b", "content": "new"},
			"irreversible_effects": []any{},
		},
		"confirmation": map[string]any{"token": "tok", "note": "record deliberate intent"},
	}, Summary: "preview ready"}
	text := Render(response, "update_block")
	for _, want := range []string{"Preview", "block:b", "block_id", "tok"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text = %q, missing %q", text, want)
		}
	}
}

func TestRenderSearchDocuments(t *testing.T) {
	response := tool.Response{Version: "v1", Tool: "context", OK: true, Data: map[string]any{
		"count": 1,
		"documents": []any{
			map[string]any{"box": "n", "hPath": "/Projects/Roadmap", "path": "/Projects/Roadmap"},
		},
	}, Summary: "document search results"}
	text := Render(response, "search_documents")
	if !strings.Contains(text, "1 documents") || !strings.Contains(text, "/Projects/Roadmap") {
		t.Fatalf("text = %q", text)
	}
}

func TestRenderFallbackMap(t *testing.T) {
	response := tool.Response{Version: "v1", Tool: "write", OK: true, Data: map[string]string{"id": "b"}, Summary: "change applied"}
	text := Render(response, "update_block")
	if !strings.Contains(text, "id: b") {
		t.Fatalf("text = %q", text)
	}
}
