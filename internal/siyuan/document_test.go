package siyuan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListDocTree_Unmarshal(t *testing.T) {
	jsonData := `{"tree":[{"id":"20260501124630-gotdnlt"}]}`

	var resp ListDocTreeResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Tree == nil {
		t.Error("expected Tree to not be nil")
	}

	if len(resp.Tree) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Tree))
	}

	if len(resp.Tree) > 0 && resp.Tree[0].ID != "20260501124630-gotdnlt" {
		t.Errorf("expected ID 20260501124630-gotdnlt, got %s", resp.Tree[0].ID)
	}
}

func TestListDocTree_WalksSubdirectories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/filetree/listDocsByPath" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		switch body["path"] {
		case "/":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"box":"b","path":"/","files":[
				{"id":"parent","name":"Parent","path":"/parent.sy","subFileCount":1},
				{"id":"leaf","name":"Leaf","path":"/leaf.sy","subFileCount":0}
			]}}`))
		case "/parent":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"box":"b","path":"/parent","files":[
				{"id":"child","name":"Child","path":"/parent/child.sy","subFileCount":0}
			]}}`))
		default:
			t.Fatalf("unexpected path %q", body["path"])
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ListDocTree(context.Background(), "b", 0)
	if err != nil {
		t.Fatalf("ListDocTree() error = %v", err)
	}
	if len(result.Tree) != 2 {
		t.Fatalf("root has %d nodes, want 2", len(result.Tree))
	}
	parent := result.Tree[0]
	if parent.Name != "Parent" || parent.Path != "/parent.sy" {
		t.Fatalf("parent = %+v", parent)
	}
	if len(parent.Children) != 1 || parent.Children[0].Name != "Child" {
		t.Fatalf("parent children = %+v", parent.Children)
	}
}

func TestListDocTree_NullTree(t *testing.T) {
	jsonData := `{"tree":null}`

	var resp ListDocTreeResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Tree != nil {
		t.Error("expected Tree to be nil for null input")
	}
}

func TestListDocTree_MaxDepthIsLocalTraversalBound(t *testing.T) {
	calls := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		requestPath, _ := body["path"].(string)
		calls = append(calls, requestPath)
		switch requestPath {
		case "/":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"files":[{"id":"parent","name":"Parent","path":"/parent.sy","subFileCount":1}]}}`))
		case "/parent":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"files":[{"id":"child","name":"Child","path":"/parent/child.sy","subFileCount":1}]}}`))
		case "/parent/child":
			_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"files":[]}}`))
		default:
			t.Fatalf("unexpected path %q", requestPath)
		}
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDocTree(context.Background(), "box", 1); err != nil {
		t.Fatalf("ListDocTree(maxDepth=1) error = %v", err)
	}
	if len(calls) != 1 || calls[0] != "/" {
		t.Fatalf("maxDepth=1 calls = %v, want root only", calls)
	}
	calls = nil
	if _, err := client.ListDocTree(context.Background(), "box", 2); err != nil {
		t.Fatalf("ListDocTree(maxDepth=2) error = %v", err)
	}
	if len(calls) != 2 || calls[0] != "/" || calls[1] != "/parent" {
		t.Fatalf("maxDepth=2 calls = %v, want root and parent", calls)
	}
}

func TestListDocTreeIntegration_WithMaxDepth(t *testing.T) {
	if os.Getenv("SIYUAN_INTEGRATION_TEST") != "1" {
		t.Skip("set SIYUAN_INTEGRATION_TEST=1 to run live SiYuan integration tests")
	}

	// Skip if no token configured.
	c, err := New()
	if err != nil {
		t.Skip("Skipping integration test - no config:", err)
	}

	resp, err := c.ListDocTree(context.Background(), "20260501124624-2qu5nyw", 10)
	if err != nil {
		t.Fatalf("ListDocTree failed: %v", err)
	}

	t.Logf("Response: %+v", resp)
	t.Logf("Tree: %+v", resp.Tree)
	t.Logf("Tree length: %d", len(resp.Tree))
}
