// Package siyuan provides tag-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tag represents a tag in SiYuan.
type Tag struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetTags retrieves all tags.
func (c *Client) GetTags(ctx context.Context) ([]Tag, error) {
	// API requires empty body, not nil
	resp, err := c.Post(ctx, "/api/tag/getTag", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result []Tag
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}
	return result, nil
}

// TagSearchResult represents the search result for tags.
type TagSearchResult struct {
	K       string   `json:"k"`
	Tags    []string `json:"tags"`
	Blocks  []BlockResult `json:"blocks,omitempty"`
}

// BlockResult represents a block in search results.
type BlockResult struct {
	BlockID     string `json:"blockID"`
	Content     string `json:"content"`
	RootID      string `json:"rootID"`
	HPath       string `json:"hPath"`
	Updated     string `json:"updated"`
	Created     string `json:"created"`
}

// SearchTags searches for blocks with specific tags.
func (c *Client) SearchTags(ctx context.Context, keyword string) (*TagSearchResult, error) {
	req := map[string]string{"k": keyword}

	resp, err := c.Post(ctx, "/api/search/searchTag", req)
	if err != nil {
		return nil, err
	}

	var result TagSearchResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tag search results: %w", err)
	}
	return &result, nil
}

// RenameTagRequest represents the request to rename a tag.
type RenameTagRequest struct {
	OldLabel string `json:"oldLabel"`
	NewLabel string `json:"newLabel"`
}

// RenameTag renames a tag across all documents.
func (c *Client) RenameTag(ctx context.Context, oldLabel, newLabel string) error {
	req := RenameTagRequest{
		OldLabel: oldLabel,
		NewLabel: newLabel,
	}
	_, err := c.Post(ctx, "/api/tag/renameTag", req)
	return err
}

// RemoveTag removes a tag from all documents.
func (c *Client) RemoveTag(ctx context.Context, label string) error {
	req := map[string]string{"label": label}
	_, err := c.Post(ctx, "/api/tag/removeTag", req)
	return err
}

// GetDocsByTag retrieves documents that have a specific tag.
func (c *Client) GetDocsByTag(ctx context.Context, label string) ([]string, error) {
	req := map[string]string{"label": label}

	resp, err := c.Post(ctx, "/api/tag/getDocsByTag", req)
	if err != nil {
		return nil, err
	}

	var result []string
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docs by tag: %w", err)
	}
	return result, nil
}
