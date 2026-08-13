package siyuan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"siyuan/internal/config"
)

func adapterTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := NewWithConfig(&config.Config{BaseURL: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	return client
}

func TestRemoveDocumentByID_SendsIDPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/filetree/removeDocByID" {
			t.Errorf("path = %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if len(body) != 1 || body["id"] != "doc-id" {
			t.Errorf("request body = %#v, want only id=doc-id", body)
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
	}))
	defer server.Close()

	if err := adapterTestClient(t, server).RemoveDocumentByID(context.Background(), "doc-id"); err != nil {
		t.Fatalf("RemoveDocumentByID() error = %v", err)
	}
}

func TestCreateDocumentWithMarkdown_AcceptsStringAndObjectResponses(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		wantID string
	}{
		{name: "legacy string", data: `"legacy-id"`, wantID: "legacy-id"},
		{name: "document object", data: `{"id":"object-id","box":"notebook-id","path":"/object-id.sy","hPath":"/Notes/Object"}`, wantID: "object-id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/api/filetree/createDocWithMd" {
					t.Errorf("path = %s", request.URL.Path)
				}
				_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":` + test.data + `}`))
			}))
			defer server.Close()

			got, err := adapterTestClient(t, server).CreateDocumentWithMarkdown(context.Background(), "notebook-id", "/Notes/Object", "# Object")
			if err != nil {
				t.Fatalf("CreateDocumentWithMarkdown() error = %v", err)
			}
			if got != test.wantID {
				t.Fatalf("CreateDocumentWithMarkdown() = %q, want %q", got, test.wantID)
			}
			if test.name == "document object" {
				result, err := adapterTestClient(t, server).CreateDocumentWithMarkdownResult(context.Background(), "notebook-id", "/Notes/Object", "# Object")
				if err != nil || result.ID != "object-id" || result.Box != "notebook-id" || result.Path != "/object-id.sy" || result.HPath != "/Notes/Object" {
					t.Fatalf("CreateDocumentWithMarkdownResult() = %#v, error = %v", result, err)
				}
			}
		})
	}
}

func TestPutFile_SendsMultipartFields(t *testing.T) {
	expected := []struct {
		path    string
		content string
		isDir   string
	}{
		{path: "/data/file.md", content: "hello\nworld", isDir: "false"},
		{path: "/data/new-dir", content: "", isDir: "true"},
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if calls >= len(expected) {
			t.Errorf("unexpected extra request")
			return
		}
		if request.URL.Path != "/api/file/putFile" {
			t.Errorf("path = %s", request.URL.Path)
		}
		contentType := request.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
			t.Errorf("Content-Type = %q, want multipart/form-data", contentType)
		}
		if strings.Contains(contentType, "application/json") {
			t.Errorf("Content-Type unexpectedly advertises JSON: %q", contentType)
		}
		if got := request.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm() error = %v", err)
		}
		want := expected[calls]
		if got := request.FormValue("path"); got != want.path {
			t.Errorf("path form value = %q, want %q", got, want.path)
		}
		if got := request.FormValue("content"); got != want.content {
			t.Errorf("content form value = %q, want %q", got, want.content)
		}
		if got := request.FormValue("isDir"); got != want.isDir {
			t.Errorf("isDir form value = %q, want %q", got, want.isDir)
		}
		calls++
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
	}))
	defer server.Close()

	client := adapterTestClient(t, server)
	for _, test := range expected {
		isDir := test.isDir == "true"
		if err := client.PutFile(context.Background(), test.path, test.content, isDir); err != nil {
			t.Fatalf("PutFile(%q) error = %v", test.path, err)
		}
	}
	if calls != len(expected) {
		t.Fatalf("server calls = %d, want %d", calls, len(expected))
	}
}
