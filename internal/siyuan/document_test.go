package siyuan

import (
	"context"
	"encoding/json"
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

func TestListDocTreeIntegration_WithMaxCount(t *testing.T) {
	// Skip if no token configured
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
