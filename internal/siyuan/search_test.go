package siyuan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFullTextSearchBlockRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/search/fullTextSearchBlock" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["query"] != "roadmap" || body["page"] != float64(2) || body["pageSize"] != float64(10) {
			t.Fatalf("request body = %#v", body)
		}
		_, _ = writer.Write([]byte(`{"code":0,"msg":"","data":{"blocks":[{"id":"b"}],"docMode":false,"matchedBlockCount":1,"matchedRootCount":1,"pageCount":1}}`))
	}))
	defer server.Close()
	t.Setenv("SIYUAN_BASE_URL", server.URL)
	t.Setenv("SIYUAN_TOKEN", "test-token")

	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.FullTextSearchBlock(context.Background(), "roadmap", 2, 10)
	if err != nil {
		t.Fatalf("FullTextSearchBlock() error = %v", err)
	}
	if result.MatchedBlockCount != 1 || len(result.Blocks) != 1 {
		t.Fatalf("result = %+v", result)
	}
}
