// Package siyuan provides attribute-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// BlockAttrs represents the attributes of a block.
type BlockAttrs map[string]string

// GetBlockAttrs retrieves all attributes of a block.
func (c *Client) GetBlockAttrs(ctx context.Context, blockID string) (BlockAttrs, error) {
	req := map[string]string{"id": blockID}

	resp, err := c.Post(ctx, "/api/attr/getBlockAttrs", req)
	if err != nil {
		return nil, err
	}

	var result BlockAttrs
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block attrs: %w", err)
	}
	return result, nil
}

// SetBlockAttrsRequest represents the request to set block attributes.
type SetBlockAttrsRequest struct {
	ID    string            `json:"id"`
	Attrs map[string]string `json:"attrs"`
}

// SetBlockAttrs sets attributes on a block.
func (c *Client) SetBlockAttrs(ctx context.Context, blockID string, attrs map[string]string) error {
	req := SetBlockAttrsRequest{
		ID:    blockID,
		Attrs: attrs,
	}
	_, err := c.Post(ctx, "/api/attr/setBlockAttrs", req)
	return err
}

// ResetBlockAttr removes a specific attribute from a block.
func (c *Client) ResetBlockAttr(ctx context.Context, blockID, key string) error {
	// Setting an empty value effectively removes the attribute
	req := SetBlockAttrsRequest{
		ID: blockID,
		Attrs: map[string]string{
			key: "",
		},
	}
	_, err := c.Post(ctx, "/api/attr/setBlockAttrs", req)
	return err
}

// BookmarkLabel represents a bookmark label.
type BookmarkLabel struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetBookmarkLabels retrieves all bookmark labels.
func (c *Client) GetBookmarkLabels(ctx context.Context) ([]BookmarkLabel, error) {
	resp, err := c.Post(ctx, "/api/attr/getBookmarkLabels", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result []BookmarkLabel
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookmark labels: %w", err)
	}
	return result, nil
}
