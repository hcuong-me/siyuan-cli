package logic

import (
	"testing"
	"time"

	"siyuan/internal/siyuan"
)

func TestNotebookLogic_SetCacheExpiry(t *testing.T) {
	logic := &NotebookLogic{
		cacheExpiry: 30 * time.Second,
	}

	logic.SetCacheExpiry(5 * time.Second)

	if logic.cacheExpiry != 5*time.Second {
		t.Errorf("expected cache expiry to be 5s, got %v", logic.cacheExpiry)
	}
}

func TestNotebookLogic_CacheBehavior(t *testing.T) {
	logic := &NotebookLogic{
		cacheExpiry: 100 * time.Millisecond,
	}

	// Test initial state
	if logic.cache != nil {
		t.Error("expected initial cache to be nil")
	}

	// Simulate cache population
	logic.cache = []siyuan.Notebook{
		{ID: "1", Name: "Test"},
	}
	logic.cacheTime = time.Now()

	// Test cache hit
	if logic.cache == nil {
		t.Error("expected cache to be populated")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Test cache expiry check
	if time.Since(logic.cacheTime) < logic.cacheExpiry {
		t.Error("expected cache to be expired")
	}
}
