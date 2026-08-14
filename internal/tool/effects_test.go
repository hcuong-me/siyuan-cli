package tool

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Setenv("SIYUAN_TOKEN", "test-token")
	t.Setenv("SIYUAN_BASE_URL", "http://127.0.0.1:6806")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"x"},"mode":"apply"}`
	if code := Execute([]string{"write"}, strings.NewReader(request), &output); code == 0 {
		t.Fatal("apply without token succeeded")
	}
	assertErrorJSON(t, output.Bytes(), ConfirmationRequired)
}

func TestOrganize(t *testing.T) {
	t.Setenv("SIYUAN_TOKEN", "test-token")
	t.Setenv("SIYUAN_BASE_URL", "http://127.0.0.1:6806")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"delete_block","input":{"block_id":"b"},"mode":"apply"}`
	if code := Execute([]string{"organize"}, strings.NewReader(request), &output); code == 0 {
		t.Fatal("apply without token succeeded")
	}
	assertErrorJSON(t, output.Bytes(), ConfirmationRequired)
}

func TestMaintain(t *testing.T) {
	t.Setenv("SIYUAN_TOKEN", "test-token")
	t.Setenv("SIYUAN_BASE_URL", "http://127.0.0.1:6806")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"remove_file","input":{"path":"/x"},"mode":"apply"}`
	if code := Execute([]string{"maintain"}, strings.NewReader(request), &output); code == 0 {
		t.Fatal("apply without token succeeded")
	}
	assertErrorJSON(t, output.Bytes(), ConfirmationRequired)
}

func TestExport(t *testing.T) { TestExportPreviewDoesNotGenerate(t) }

func TestMutationPreviewsDoNotCallServer(t *testing.T) {
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/updateBlock" || request.URL.Path == "/api/block/deleteBlock" || request.URL.Path == "/api/export/exportPDF" {
			mutations++
		}
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	requests := []struct{ tool, body string }{
		{"write", `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"x"},"mode":"preview"}`},
		{"organize", `{"version":"v1","operation":"delete_block","input":{"block_id":"b"},"mode":"preview"}`},
		{"export", `{"version":"v1","operation":"export_pdf","input":{"document_id":"d","output_path":"/tmp/x"},"mode":"preview"}`},
		{"maintain", `{"version":"v1","operation":"remove_file","input":{"path":"/x"},"mode":"preview"}`},
	}
	for _, request := range requests {
		var output bytes.Buffer
		if code := Execute([]string{request.tool}, strings.NewReader(request.body), &output); code != 0 {
			t.Fatalf("%s preview code = %d: %s", request.tool, code, output.String())
		}
	}
	if mutations != 0 {
		t.Fatalf("previews sent %d mutation requests", mutations)
	}
}

func TestExportPreviewDoesNotGenerate(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/export/") {
			calls++
		}
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"export_pdf","input":{"document_id":"d","output_path":"/tmp/x.pdf"},"mode":"preview"}`
	if code := Execute([]string{"export"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("preview code = %d: %s", code, output.String())
	}
	if calls != 0 {
		t.Fatalf("preview sent %d export requests", calls)
	}
}

func TestPreviewContainsConcreteIrreversibleEffects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"delete_block","input":{"block_id":"block-42"},"mode":"preview"}`
	if code := Execute([]string{"organize"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("preview code = %d: %s", code, output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %#v", response.Data)
	}
	preview, ok := data["preview"].(map[string]any)
	if !ok {
		t.Fatalf("preview = %#v", data["preview"])
	}
	effects, ok := preview["irreversible_effects"].([]any)
	if !ok || len(effects) == 0 || effects[0] != "delete_block:block-42" {
		t.Fatalf("irreversible_effects = %#v", preview["irreversible_effects"])
	}
}

func TestLocalExportPreconditionFailureIsPreconditionUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"export_markdown","input":{"document_id":"doc-1","output_path":"/definitely/missing-parent/out.md"},"mode":"preview"}`
	if code := Execute([]string{"export"}, strings.NewReader(request), &output); code == 0 {
		t.Fatalf("invalid local destination succeeded: %s", output.String())
	}
	// A preview that cannot inspect the destination issues no token.
	assertErrorJSON(t, output.Bytes(), PreconditionUnavailable)
}

func TestPDFDOCXApplyGenerates(t *testing.T) {
	for _, operation := range []string{"export_pdf", "export_docx"} {
		t.Run(operation, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/api/export/exportPDF" || request.URL.Path == "/api/export/exportDocx" {
					calls++
				}
				writeStateResponse(writer, request.URL.Path)
			}))
			defer server.Close()
			t.Setenv("SIYUAN_BASE_URL", server.URL)
			t.Setenv("SIYUAN_TOKEN", "test-token")
			input := `{"version":"v1","operation":"` + operation + `","input":{"document_id":"d","output_path":"/tmp/x"},"mode":"preview"}`
			var previewOutput bytes.Buffer
			if code := Execute([]string{"export"}, strings.NewReader(input), &previewOutput); code != 0 {
				t.Fatalf("preview code = %d", code)
			}
			var preview Response
			if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
				t.Fatal(err)
			}
			data := preview.Data.(map[string]any)
			token := data["confirmation"].(map[string]any)["token"].(string)
			apply := `{"version":"v1","operation":"` + operation + `","input":{"document_id":"d","output_path":"/tmp/x"},"mode":"apply","confirmation_token":"` + token + `"}`
			var applyOutput bytes.Buffer
			if code := Execute([]string{"export"}, strings.NewReader(apply), &applyOutput); code != 0 {
				t.Fatalf("apply code = %d: %s", code, applyOutput.String())
			}
			if calls != 1 {
				t.Fatalf("apply sent %d export requests, want 1", calls)
			}
		})
	}
}

func TestStatefulMutationStaleConfirmation(t *testing.T) {
	state := "before"
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/getBlockKramdown" {
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"` + state + `"}}`))
			return
		}
		if request.URL.Path == "/api/block/updateBlock" {
			mutations++
		}
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"new"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	state = "after"
	applyRequest := `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"new"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatal("stale apply succeeded")
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if mutations != 0 {
		t.Fatalf("stale apply sent %d mutation requests", mutations)
	}
}

func TestRepeatedApplyBecomesStaleAfterMutation(t *testing.T) {
	state := "before"
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/block/getBlockKramdown":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"` + state + `"}}`))
		case "/api/block/updateBlock":
			mutations++
			state = "after"
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		default:
			writeStateResponse(writer, request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"new"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"update_block","input":{"block_id":"b","content":"new"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("first apply: %s", applyOutput.String())
	}
	applyOutput.Reset()
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("repeated apply succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if mutations != 1 {
		t.Fatalf("mutation requests = %d, want one successful apply", mutations)
	}
}

func TestCreateDocumentAbsenceChangeBecomesStale(t *testing.T) {
	ids := []string{}
	creates := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n","name":"Notes"}]}}`))
		case "/api/filetree/getIDsByHPath":
			encoded, _ := json.Marshal(ids)
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":` + string(encoded) + `}`))
		case "/api/filetree/createDocWithMd":
			creates++
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":"created"}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"create_document","input":{"notebook":"n","path":"/New","content":"x"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	ids = []string{"created"}
	applyRequest := `{"version":"v1","operation":"create_document","input":{"notebook":"n","path":"/New","content":"x"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("stale create succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if creates != 0 {
		t.Fatalf("create requests = %d, want zero", creates)
	}
}

func TestUnusedAssetSetChangeBecomesStale(t *testing.T) {
	assets := []map[string]any{{"path": "/assets/old.png", "size": 1, "updated": 1}}
	removals := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/asset/getUnusedAssets":
			encoded, _ := json.Marshal(assets)
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":` + string(encoded) + `}`))
		case "/api/asset/removeUnusedAssets":
			removals++
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"clean_unused_assets","input":{},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	assets = append(assets, map[string]any{"path": "/assets/new.png", "size": 2, "updated": 2})
	applyRequest := `{"version":"v1","operation":"clean_unused_assets","input":{},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("stale cleanup succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if removals != 0 {
		t.Fatalf("remove requests = %d, want zero", removals)
	}
}

func TestUploadSourceChangeBecomesStale(t *testing.T) {
	source := filepath.Join(t.TempDir(), "asset.bin")
	if err := os.WriteFile(source, []byte("before"), 0644); err != nil {
		t.Fatal(err)
	}
	uploads := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/file/readDir":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[]}`))
		case "/api/asset/upload":
			uploads++
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"succMap":{},"errFiles":[]}}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"upload_asset","input":{"path":"` + source + `"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	if err := os.WriteFile(source, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	applyRequest := `{"version":"v1","operation":"upload_asset","input":{"path":"` + source + `"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("stale upload succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if uploads != 0 {
		t.Fatalf("upload requests = %d, want zero", uploads)
	}
}

func TestDisappearedTargetReturnsStaleBeforeMutation(t *testing.T) {
	gone := false
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/getBlockKramdown" {
			if gone {
				writer.WriteHeader(http.StatusNotFound)
				_, _ = writer.Write([]byte(`{"code":-1,"msg":"not found","data":null}`))
				return
			}
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"before"}}`))
			return
		}
		if request.URL.Path == "/api/block/deleteBlock" {
			mutations++
		}
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"delete_block","input":{"block_id":"gone-block"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"organize"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	gone = true
	applyRequest := `{"version":"v1","operation":"delete_block","input":{"block_id":"gone-block"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"organize"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("disappeared target apply succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	if mutations != 0 {
		t.Fatalf("stale apply sent %d mutations", mutations)
	}
}

func TestChangedLocalExportDestinationReturnsStale(t *testing.T) {
	destination := t.TempDir() + "/export.md"
	if err := os.WriteFile(destination, []byte("before"), 0644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"export_markdown","input":{"document_id":"doc-1","output_path":"` + destination + `"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"export"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	if err := os.WriteFile(destination, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	applyRequest := `{"version":"v1","operation":"export_markdown","input":{"document_id":"doc-1","output_path":"` + destination + `"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"export"}, strings.NewReader(applyRequest), &applyOutput); code == 0 {
		t.Fatalf("changed destination apply succeeded: %s", applyOutput.String())
	}
	assertErrorJSON(t, applyOutput.Bytes(), ConfirmationStale)
	content, err := os.ReadFile(destination)
	if err != nil || string(content) != "changed" {
		t.Fatalf("destination changed after stale apply: %q, error=%v", content, err)
	}
}

func TestRenameFileAllowsAbsentDestination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/file/getFile":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["path"] == "/data/new.txt" {
				writer.WriteHeader(http.StatusNotFound)
				_, _ = writer.Write([]byte(`{"code":-1,"msg":"not found","data":null}`))
				return
			}
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":"old"}`))
		case "/api/file/readDir":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[{"name":"old.txt","isDir":false}]}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	var output bytes.Buffer
	request := `{"version":"v1","operation":"rename_file","input":{"old_path":"/data/old.txt","new_path":"/data/new.txt"},"mode":"preview"}`
	if code := Execute([]string{"maintain"}, strings.NewReader(request), &output); code != 0 {
		t.Fatalf("rename preview code = %d: %s", code, output.String())
	}
}

func TestUpdateDocumentReplacesContent(t *testing.T) {
	updates := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n","name":"Notes"}]}}`))
		case "/api/filetree/getIDsByHPath":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":["d"]}`))
		case "/api/block/getBlockKramdown":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"old"}}`))
		case "/api/block/updateBlock":
			updates++
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		default:
			writeStateResponse(writer, request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"update_document","input":{"notebook":"n","path":"/Doc","content":"new"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"update_document","input":{"notebook":"n","path":"/Doc","content":"new"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("apply: %s", applyOutput.String())
	}
	if updates != 1 {
		t.Fatalf("update requests = %d, want 1", updates)
	}
}

func TestCanonicalIALs(t *testing.T) {
	variants := []string{
		"# Test Doc\n{: id=\"20260802172657-tqmrp37\" updated=\"20260802172657\"}\n\nbody\n",
		"# Test Doc\n{: updated=\"20260802172657\" id=\"20260802172657-tqmrp37\"}\n\nbody\n",
	}
	var first string
	for i, variant := range variants {
		canonical := canonicalIALs(variant)
		if i == 0 {
			first = canonical
		} else if canonical != first {
			t.Fatalf("canonical form differs:\n%s\n%s", first, canonical)
		}
		if strings.Contains(canonical, "updated=\""+`" id=`+"\"") {
			t.Fatalf("attribute order not normalized: %q", canonical)
		}
	}
	if got := canonicalIALs("plain text without attributes"); got != "plain text without attributes" {
		t.Fatalf("plain text changed: %q", got)
	}
}

func TestNotebookNameSelectorAppliesByID(t *testing.T) {
	requested := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"nb-1","name":"Notes"}]}}`))
			return
		case "/api/notebook/renameNotebook", "/api/notebook/openNotebook", "/api/notebook/closeNotebook", "/api/notebook/removeNotebook":
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			requested[request.URL.Path] = body["notebook"]
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	operations := []struct{ operation, input string }{
		{"rename_notebook", `"notebook":"Notes","name":"Renamed"`},
		{"open_notebook", `"notebook":"Notes"`},
		{"close_notebook", `"notebook":"Notes"`},
		{"remove_notebook", `"notebook":"Notes"`},
	}
	for _, op := range operations {
		input := `{"version":"v1","operation":"` + op.operation + `","input":{` + op.input + `},"mode":"preview"}`
		var previewOutput bytes.Buffer
		if code := Execute([]string{"organize"}, strings.NewReader(input), &previewOutput); code != 0 {
			t.Fatalf("%s preview: %s", op.operation, previewOutput.String())
		}
		var preview Response
		if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
			t.Fatal(err)
		}
		token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
		apply := `{"version":"v1","operation":"` + op.operation + `","input":{` + op.input + `},"mode":"apply","confirmation_token":"` + token + `"}`
		var applyOutput bytes.Buffer
		if code := Execute([]string{"organize"}, strings.NewReader(apply), &applyOutput); code != 0 {
			t.Fatalf("%s apply: %s", op.operation, applyOutput.String())
		}
		var response Response
		if err := json.Unmarshal(applyOutput.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		data := response.Data.(map[string]any)
		if data["id"] != "nb-1" {
			t.Fatalf("%s response id = %v, want nb-1", op.operation, data["id"])
		}
	}
	wantPath := map[string]string{
		"rename_notebook": "/api/notebook/renameNotebook",
		"open_notebook":   "/api/notebook/openNotebook",
		"close_notebook":  "/api/notebook/closeNotebook",
		"remove_notebook": "/api/notebook/removeNotebook",
	}
	for operation, path := range wantPath {
		if requested[path] != "nb-1" {
			t.Fatalf("%s sent notebook %q, want resolved ID nb-1", operation, requested[path])
		}
	}
}

func TestCreateDocumentResolvesNotebookName(t *testing.T) {
	created := false
	requestedNotebook := ""
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"nb-1","name":"Notes"}]}}`))
		case "/api/filetree/getIDsByHPath":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[]}`))
		case "/api/filetree/createDocWithMd":
			created = true
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			requestedNotebook = body["notebook"]
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"id":"doc-1","box":"nb-1","path":"/New.sy","hPath":"/New"}}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"create_document","input":{"notebook":"Notes","path":"/New","content":"x"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"create_document","input":{"notebook":"Notes","path":"/New","content":"x"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"write"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("apply: %s", applyOutput.String())
	}
	if !created {
		t.Fatal("create request never sent")
	}
	if requestedNotebook != "nb-1" {
		t.Fatalf("create sent notebook %q, want resolved ID nb-1", requestedNotebook)
	}
	var response Response
	if err := json.Unmarshal(applyOutput.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data := response.Data.(map[string]any)
	if data["notebook"] != "nb-1" || data["id"] != "doc-1" {
		t.Fatalf("apply data = %#v", response.Data)
	}
}

func TestAmbiguousNotebookNameReturnsCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/notebook/lsNotebooks" {
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"nb-1","name":"Notes"},{"id":"nb-2","name":"Notes"}]}}`))
			return
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	request := `{"version":"v1","operation":"open_notebook","input":{"notebook":"Notes"},"mode":"preview"}`
	var output bytes.Buffer
	if code := Execute([]string{"organize"}, strings.NewReader(request), &output); code == 0 {
		t.Fatalf("ambiguous preview succeeded: %s", output.String())
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error == nil || response.Error.Code != AmbiguousSelector {
		t.Fatalf("error = %#v", response.Error)
	}
	if len(response.Error.Candidates) != 2 {
		t.Fatalf("candidates = %#v", response.Error.Candidates)
	}
}

func TestRenameTagApplyResolvesIdentity(t *testing.T) {
	renamed := false
	oldLabel := ""
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/tag/getTag":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":[{"label":"old","count":3}]}`))
		case "/api/tag/renameTag":
			renamed = true
			var body map[string]string
			_ = json.NewDecoder(request.Body).Decode(&body)
			oldLabel = body["oldLabel"]
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"rename_tag","input":{"tag":"old","name":"new"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"organize"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"rename_tag","input":{"tag":"old","name":"new"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"organize"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("apply: %s", applyOutput.String())
	}
	if !renamed || oldLabel != "old" {
		t.Fatalf("rename oldLabel = %q, renamed = %v", oldLabel, renamed)
	}
	var response Response
	if err := json.Unmarshal(applyOutput.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data := response.Data.(map[string]any)
	if data["tag"] != "old" || data["count"] != float64(3) {
		t.Fatalf("apply data = %#v", response.Data)
	}
}

func TestExportMarkdownApplyReturnsIdentityAndBytes(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "out.md")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeStateResponse(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"export_markdown","input":{"document_id":"doc-1","output_path":"` + destination + `"},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"export"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"export_markdown","input":{"document_id":"doc-1","output_path":"` + destination + `"},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"export"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("apply: %s", applyOutput.String())
	}
	var response Response
	if err := json.Unmarshal(applyOutput.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data := response.Data.(map[string]any)
	if data["document_id"] != "doc-1" {
		t.Fatalf("data = %#v", response.Data)
	}
	if bytes, ok := data["bytes"].(float64); !ok || bytes != float64(len("state")) {
		t.Fatalf("bytes = %#v", data["bytes"])
	}
	content, err := os.ReadFile(destination)
	if err != nil || string(content) != "state" {
		t.Fatalf("export content = %q, error = %v", content, err)
	}
}

func TestCleanUnusedAssetsApplyReturnsCounts(t *testing.T) {
	assets := []map[string]any{{"path": "/assets/a.png", "size": 1, "updated": 1}, {"path": "/assets/b.png", "size": 2, "updated": 2}}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/asset/getUnusedAssets":
			encoded, _ := json.Marshal(assets)
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":` + string(encoded) + `}`))
		case "/api/asset/removeUnusedAssets":
			assets = nil
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		default:
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	previewRequest := `{"version":"v1","operation":"clean_unused_assets","input":{},"mode":"preview"}`
	var previewOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(previewRequest), &previewOutput); code != 0 {
		t.Fatalf("preview: %s", previewOutput.String())
	}
	var preview Response
	if err := json.Unmarshal(previewOutput.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	token := preview.Data.(map[string]any)["confirmation"].(map[string]any)["token"].(string)
	applyRequest := `{"version":"v1","operation":"clean_unused_assets","input":{},"mode":"apply","confirmation_token":"` + token + `"}`
	var applyOutput bytes.Buffer
	if code := Execute([]string{"maintain"}, strings.NewReader(applyRequest), &applyOutput); code != 0 {
		t.Fatalf("apply: %s", applyOutput.String())
	}
	var response Response
	if err := json.Unmarshal(applyOutput.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	data := response.Data.(map[string]any)
	if data["removed"] != float64(2) || data["remaining"] != float64(0) {
		t.Fatalf("data = %#v", response.Data)
	}
}

func writeStateResponse(writer http.ResponseWriter, path string) {
	switch path {
	case "/api/block/getBlockKramdown":
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"state"}}`))
	case "/api/export/exportMdContent":
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"content":"state"}}`))
	default:
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"path":"/export/result"}}`))
	}
}
