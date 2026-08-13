package human

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"siyuan/internal/tool"
)

func TestRoutingBareInvocation(t *testing.T) {
	var output bytes.Buffer
	if code := Execute(nil, strings.NewReader(""), &output); code == 0 {
		t.Fatal("bare invocation succeeded")
	}
	if !strings.Contains(output.String(), "Usage") {
		t.Fatalf("bare output = %q", output.String())
	}
}

func TestRoutingToolOnlyPipedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/system/version" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":"3.2.0"}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	request := `{"version":"v1","operation":"version","input":{}}`
	var agentOutput bytes.Buffer
	if code := tool.Execute([]string{"status"}, strings.NewReader(request), &agentOutput); code != 0 {
		t.Fatalf("agent path code = %d", code)
	}
	var humanOutput bytes.Buffer
	if code := Execute([]string{"status"}, strings.NewReader(request), &humanOutput); code != 0 {
		t.Fatalf("human path code = %d: %s", code, humanOutput.String())
	}
	if !bytes.Equal(agentOutput.Bytes(), humanOutput.Bytes()) {
		t.Fatalf("agent path and human path diverge:\nagent: %s\nhuman: %s", agentOutput.String(), humanOutput.String())
	}
}

func TestRoutingToolOnlyOnTerminalShowsUsage(t *testing.T) {
	interactiveStdin = func(io.Reader) bool { return true }
	var output bytes.Buffer
	if code := Execute([]string{"status"}, strings.NewReader(""), &output); code == 0 {
		t.Fatal("tool-only invocation on a terminal succeeded")
	}
	if !strings.Contains(output.String(), "Operations") || !strings.Contains(output.String(), "version") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRoutingToolsCatalog(t *testing.T) {
	var output bytes.Buffer
	if code := Execute([]string{"tools"}, strings.NewReader("ignored"), &output); code != 0 {
		t.Fatalf("code = %d", code)
	}
	var response tool.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || !response.OK || response.Tool != "tools" {
		t.Fatalf("output = %s, error = %v", output.String(), err)
	}
}

func TestRoutingUnknownTool(t *testing.T) {
	var output bytes.Buffer
	if code := Execute([]string{"missing"}, strings.NewReader("{}"), &output); code == 0 {
		t.Fatal("unknown tool succeeded")
	}
	var response tool.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || response.OK {
		t.Fatalf("output = %s, error = %v", output.String(), err)
	}
}

func TestRoutingHelp(t *testing.T) {
	// siyuan-cli -h prints the human usage text, not the agent catalog.
	for _, flag := range []string{"-h", "--help"} {
		var output bytes.Buffer
		if code := Execute([]string{flag}, strings.NewReader(""), &output); code != 0 {
			t.Fatalf("%s code = %d", flag, code)
		}
		if !strings.Contains(output.String(), "Usage") {
			t.Fatalf("%s output = %q", flag, output.String())
		}
	}

	output := bytes.Buffer{}
	if code := Execute([]string{"context", "-h"}, strings.NewReader(""), &output); code != 0 {
		t.Fatalf("context -h code = %d", code)
	}
	if !strings.Contains(output.String(), "Operations") || !strings.Contains(output.String(), "read_document") {
		t.Fatalf("context -h output = %q", output.String())
	}

	output.Reset()
	if code := Execute([]string{"context", "read_document", "-h"}, strings.NewReader(""), &output); code != 0 {
		t.Fatalf("operation -h code = %d", code)
	}
	for _, want := range []string{"--notebook", "--path", "--json"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("operation -h output = %q, missing %q", output.String(), want)
		}
	}
}

func TestRoutingFlagPathPassesOneArgument(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hits++
		switch request.URL.Path {
		case "/api/notebook/lsNotebooks":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n","name":"Notes"}]}}`))
		case "/api/filetree/getIDsByHPath":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":["d"]}`))
		case "/api/export/exportMdContent":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"hPath":"/Projects/Roadmap","content":"content"}}`))
		default:
			t.Fatalf("unexpected path = %s", request.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	code := Execute([]string{"context", "read_document", "--notebook", "n", "--path", "/Projects/Roadmap"}, strings.NewReader(""), &output)
	if code != 0 {
		t.Fatalf("code = %d: %s", code, output.String())
	}
	// Only a single-argument dispatch to tool.Execute reaches the server:
	// lsNotebooks resolution, getIDsByHPath, then exportMdContent.
	if hits != 3 {
		t.Fatalf("hits = %d, want 3", hits)
	}
	if !strings.Contains(output.String(), "Document") || !strings.Contains(output.String(), "content") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestJSONOutputByteIdentical(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"notebooks":[{"id":"n1","name":"Notes"}]}}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	request := `{"version":"v1","operation":"list_notebooks","input":{}}`
	var agentOutput bytes.Buffer
	if code := tool.Execute([]string{"context"}, strings.NewReader(request), &agentOutput); code != 0 {
		t.Fatalf("agent path code = %d", code)
	}
	var humanOutput bytes.Buffer
	if code := Execute([]string{"context", "list_notebooks", "--json"}, strings.NewReader(""), &humanOutput); code != 0 {
		t.Fatalf("human path code = %d: %s", code, humanOutput.String())
	}
	if !bytes.Equal(agentOutput.Bytes(), humanOutput.Bytes()) {
		t.Fatalf("--json diverges from agent path:\nagent: %s\nhuman: %s", agentOutput.String(), humanOutput.String())
	}
}

func TestMutationPreviewNeverAppliesWithoutConfirmation(t *testing.T) {
	interactiveStdin = func(io.Reader) bool { return false }
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/updateBlock" {
			mutations++
		}
		writeHumanState(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	code := Execute([]string{"write", "update_block", "--block_id", "b", "--content", "new"}, strings.NewReader(""), &output)
	if code == 0 {
		t.Fatal("non-TTY apply without --yes succeeded")
	}
	if mutations != 0 {
		t.Fatalf("mutations = %d, want 0", mutations)
	}
	if !strings.Contains(output.String(), "Preview") {
		t.Fatalf("preview missing from output = %q", output.String())
	}
	if !strings.Contains(output.String(), "Refusing to apply") || !strings.Contains(output.String(), "--yes") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestMutationFlowYesApplies(t *testing.T) {
	interactiveStdin = func(io.Reader) bool { return false }
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/updateBlock" {
			mutations++
		}
		writeHumanState(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	code := Execute([]string{"write", "update_block", "--block_id", "b", "--content", "new", "--yes"}, strings.NewReader(""), &output)
	if code != 0 {
		t.Fatalf("code = %d: %s", code, output.String())
	}
	if mutations != 1 {
		t.Fatalf("mutations = %d, want 1", mutations)
	}
}

func TestMutationFlowInteractivePrompt(t *testing.T) {
	interactiveStdin = func(io.Reader) bool { return true }
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/updateBlock" {
			mutations++
		}
		writeHumanState(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	code := Execute([]string{"write", "update_block", "--block_id", "b", "--content", "new"}, strings.NewReader("y\n"), &output)
	if code != 0 {
		t.Fatalf("code = %d: %s", code, output.String())
	}
	if mutations != 1 {
		t.Fatalf("mutations = %d, want 1", mutations)
	}
	if !strings.Contains(output.String(), "Apply? [y/N]") {
		t.Fatalf("prompt missing from output = %q", output.String())
	}

	mutations = 0
	output.Reset()
	code = Execute([]string{"write", "update_block", "--block_id", "b", "--content", "new"}, strings.NewReader("n\n"), &output)
	if code != 0 {
		t.Fatalf("decline code = %d: %s", code, output.String())
	}
	if mutations != 0 {
		t.Fatalf("declined apply sent %d mutations", mutations)
	}
	if !strings.Contains(output.String(), "Cancelled.") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestMutationApplyWithoutTokenRejected(t *testing.T) {
	interactiveStdin = func(io.Reader) bool { return false }
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		t.Fatalf("unexpected request to %s", request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	code := Execute([]string{"write", "update_block", "--block_id", "b", "--content", "x", "--mode", "apply"}, strings.NewReader(""), &output)
	if code == 0 {
		t.Fatal("apply without token succeeded")
	}
	if !strings.Contains(output.String(), "confirmation token") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestJSONSideEffectIsPreviewOnly(t *testing.T) {
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/block/updateBlock" {
			mutations++
		}
		writeHumanState(writer, request.URL.Path)
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	var output bytes.Buffer
	if code := Execute([]string{"write", "update_block", "--block_id", "b", "--content", "new", "--json"}, strings.NewReader(""), &output); code != 0 {
		t.Fatalf("JSON preview code = %d: %s", code, output.String())
	}
	var response tool.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("JSON preview is not one envelope: %v; %s", err, output.String())
	}
	if !response.OK || mutations != 0 {
		t.Fatalf("response = %#v, mutations = %d", response, mutations)
	}
	data, ok := response.Data.(map[string]any)
	if !ok || data["confirmation"] == nil {
		t.Fatalf("preview data = %#v", response.Data)
	}

	for _, args := range [][]string{
		{"write", "update_block", "--block_id", "b", "--content", "new", "--json", "--yes"},
		{"write", "update_block", "--block_id", "b", "--content", "new", "--json", "--mode", "apply"},
	} {
		output.Reset()
		if code := Execute(args, strings.NewReader(""), &output); code == 0 {
			t.Fatalf("args %v unexpectedly succeeded", args)
		}
		var rejected tool.Response
		if err := json.Unmarshal(output.Bytes(), &rejected); err != nil || rejected.Error == nil || rejected.Error.Code != tool.InvalidRequest {
			t.Fatalf("args %v response = %s, error = %v", args, output.String(), err)
		}
	}
	if mutations != 0 {
		t.Fatalf("JSON rejection sent %d mutations", mutations)
	}
}

func writeHumanState(writer http.ResponseWriter, path string) {
	switch path {
	case "/api/block/getBlockKramdown":
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"kramdown":"state"}}`))
	default:
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":null}`))
	}
}
