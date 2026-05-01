// Package siyuan provides search-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// SearchBlockResult represents a block search result.
type SearchBlockResult struct {
	Blocks            []SearchBlock `json:"blocks"`
	DocMode           bool          `json:"docMode"`
	MatchedBlockCount int           `json:"matchedBlockCount"`
	MatchedRootCount  int           `json:"matchedRootCount"`
	PageCount         int           `json:"pageCount"`
}

// SearchBlock represents a single block in search results.
type SearchBlock struct {
	ID      string `json:"id"`
	RootID  string `json:"rootID"`
	Box     string `json:"box"`
	Path    string `json:"path"`
	HPath   string `json:"hPath"`
	Content string `json:"content"`
	Type    string `json:"type"`
	SubType string `json:"subType"`
	Updated string `json:"updated"`
}

// FullTextSearchBlock performs full-text search on blocks.
func (c *Client) FullTextSearchBlock(ctx context.Context, keyword string, page, size int) (*SearchBlockResult, error) {
	req := map[string]interface{}{
		"k":    keyword,
		"page": page,
		"size": size,
	}

	resp, err := c.Post(ctx, "/api/search/fullTextSearchBlock", req)
	if err != nil {
		return nil, err
	}

	var result SearchBlockResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}
	return &result, nil
}

// SearchDocResult represents a document search result.
type SearchDocResult struct {
	Box     string `json:"box"`
	BoxIcon string `json:"boxIcon"`
	HPath   string `json:"hPath"`
	Path    string `json:"path"`
}

// SearchDocs searches for documents by keyword.
func (c *Client) SearchDocs(ctx context.Context, keyword, notebook, path string) ([]SearchDocResult, error) {
	req := map[string]string{
		"k":        keyword,
		"notebook": notebook,
		"path":     path,
	}

	resp, err := c.Post(ctx, "/api/filetree/searchDocs", req)
	if err != nil {
		return nil, err
	}

	var result []SearchDocResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc search results: %w", err)
	}
	return result, nil
}
