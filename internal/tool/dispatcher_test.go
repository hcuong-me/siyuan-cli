package tool

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestToolsCatalog(t *testing.T) {
	var output bytes.Buffer
	if code := Execute([]string{"tools"}, strings.NewReader("not read"), &output); code != 0 {
		t.Fatalf("Execute(tools) code = %d", code)
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK || response.Tool != "tools" {
		t.Fatalf("catalog response = %s, error = %v", output.String(), err)
	}
}

func TestHelpFlag(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		var output bytes.Buffer
		if code := Execute([]string{flag}, strings.NewReader("not read"), &output); code != 0 {
			t.Fatalf("Execute(%s) code = %d", flag, code)
		}
		var response Response
		if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK || response.Tool != "tools" {
			t.Fatalf("help response = %s, error = %v", output.String(), err)
		}
	}
}

func TestUnknownToolEnvelope(t *testing.T) {
	var output bytes.Buffer
	if code := Execute([]string{"missing"}, strings.NewReader("{}"), &output); code == 0 {
		t.Fatal("Execute(missing) succeeded")
	}
	assertErrorJSON(t, output.Bytes(), InvalidRequest)
}

func TestDispatcher(t *testing.T) {
	var output bytes.Buffer
	if code := Execute([]string{"status", "extra"}, strings.NewReader("{}"), &output); code == 0 {
		t.Fatal("Execute with extra argument succeeded")
	}
	assertErrorJSON(t, output.Bytes(), InvalidRequest)

	output.Reset()
	if code := Execute([]string{"status"}, strings.NewReader("not json"), &output); code == 0 {
		t.Fatal("Execute with invalid JSON succeeded")
	}
	assertErrorJSON(t, output.Bytes(), InvalidRequest)
}

func TestStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/system/version" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":"3.2.0"}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"version","input":{}}`
	if code := Execute([]string{"status"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("Execute(status) code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK {
		t.Fatalf("response = %s, error = %v", output.String(), err)
	}
}

func TestContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/notebook/lsNotebooks" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n1","name":"Notes"}]}}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"list_notebooks","input":{}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("Execute(context) code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK {
		t.Fatalf("response = %s, error = %v", output.String(), err)
	}
}

func TestSelectorResolution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n1","name":"Notes"}]}}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"resolve_notebook","input":{"selector":"Notes"}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("resolve code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK {
		t.Fatalf("response = %s, error = %v", output.String(), err)
	}
}

func TestAmbiguousDocumentReturnsCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n","name":"Notes"}]}}`))
		case "/api/filetree/getIDsByHPath":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":["doc-b","doc-a"]}`))
		default:
			t.Fatalf("unexpected path = %s", request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"read_document","input":{"notebook":"n","path":"/Notes/Duplicate"}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code == 0 {
		t.Fatalf("ambiguous document succeeded: %s", output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Code != AmbiguousSelector || len(response.Error.Candidates) != 2 {
		t.Fatalf("response = %#v", response)
	}
	if response.Error.Candidates[0].ID != "doc-a" || response.Error.Candidates[1].ID != "doc-b" {
		t.Fatalf("candidates = %#v", response.Error.Candidates)
	}
}

func TestSearchDocumentsResolvesNotebookPrefixedHPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/filetree/searchDocs":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[{"box":"n","hPath":"Notes/Doc","path":"/doc.sy"}]}`))
		case "/api/filetree/getIDsByHPath":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["path"] == "/Notes/Doc" {
				_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[]}`))
				return
			}
			if body["path"] != "/Doc" {
				t.Fatalf("path = %q, want /Doc fallback", body["path"])
			}
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":["doc-id"]}`))
		default:
			t.Fatalf("unexpected path = %s", request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"search_documents","input":{"query":"doc"}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("search code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %#v", response.Data)
	}
	documents, ok := data["documents"].([]any)
	if !ok || len(documents) != 1 {
		t.Fatalf("documents = %#v", data["documents"])
	}
	first, ok := documents[0].(map[string]any)
	if !ok || first["id"] != "doc-id" {
		t.Fatalf("first document = %#v", documents[0])
	}
}

func TestSQLReadOnlyGate(t *testing.T) {
	oldURL, hadURL := os.LookupEnv("SIYUAN_BASE_URL")
	oldToken, hadToken := os.LookupEnv("SIYUAN_TOKEN")
	defer func() {
		if hadURL {
			_ = os.Setenv("SIYUAN_BASE_URL", oldURL)
		} else {
			_ = os.Unsetenv("SIYUAN_BASE_URL")
		}
		if hadToken {
			_ = os.Setenv("SIYUAN_TOKEN", oldToken)
		} else {
			_ = os.Unsetenv("SIYUAN_TOKEN")
		}
	}()
	_ = os.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"query","input":{"statement":"DELETE FROM blocks"}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code == 0 {
		t.Fatal("unsafe query succeeded")
	}
	assertErrorJSON(t, output.Bytes(), InvalidRequest)
}

func TestCLISubprocess(t *testing.T) {
	command := exec.Command("go", "run", "../../cmd/siyuan", "tools")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("tools subprocess failed: %v", err)
	}
	var response Response
	if err := json.Unmarshal(output, &response); err != nil || !response.OK {
		t.Fatalf("tools output = %s, error = %v", output, err)
	}

	command = exec.Command("go", "run", "../../cmd/siyuan", "status")
	command.Stdin = strings.NewReader("not-json")
	output, err = command.Output()
	if err == nil {
		t.Fatal("malformed request subprocess succeeded")
	}
	assertErrorJSON(t, output, InvalidRequest)
}

func TestListDocumentsResolvesNotebookNameAndCounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"nb-1","name":"Notes"}]}}`))
		case "/api/filetree/listDocsByPath":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["path"] == "/" {
				_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"files":[{"id":"doc-a","name":"a","path":"/a.sy","subFileCount":1}]}}`))
				return
			}
			if body["path"] == "/a" {
				_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"files":[{"id":"doc-b","name":"b","path":"/a/b.sy","subFileCount":0}]}}`))
				return
			}
			t.Fatalf("unexpected listDocsByPath path = %q", body["path"])
		default:
			t.Fatalf("unexpected path = %s", request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"list_documents","input":{"notebook":"Notes"}}`
	if code := Execute([]string{"context"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("list code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data := response.Data.(map[string]any)
	if data["count"] != float64(2) {
		t.Fatalf("count = %#v, want 2", data["count"])
	}
	notebook := data["notebook"].(map[string]any)
	if notebook["id"] != "nb-1" {
		t.Fatalf("notebook = %#v", notebook)
	}
}

func TestPreviewDocumentBadIDIsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/getBlockKramdown" {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"code":-1,"msg":"not found","data":null}`))
			return
		}
		t.Fatalf("unexpected path = %s", request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"preview_document","input":{"document_id":"missing"}}`
	if code := Execute([]string{"export"}, strings.NewReader(request), &output); code == 0 {
		t.Fatalf("bad ID preview succeeded: %s", output.String())
	}
	assertErrorJSON(t, output.Bytes(), NotFound)
}

func assertErrorJSON(t *testing.T, data []byte, code string) {
	t.Helper()
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("stdout is not JSON: %v; %s", err, data)
	}
	if response.OK || response.Error == nil || response.Error.Code != code {
		t.Fatalf("response = %#v", response)
	}
}
