// Package siyuan provides block-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// BlockInfo represents information about a block.
type BlockInfo struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	SubType  string `json:"subType"`
	Content  string `json:"content"`
	Markdown string `json:"markdown"`
}

// GetBlockKramdown retrieves a block's kramdown source.
func (c *Client) GetBlockKramdown(ctx context.Context, blockID string) (string, error) {
	req := map[string]string{"id": blockID}

	resp, err := c.Post(ctx, "/api/block/getBlockKramdown", req)
	if err != nil {
		return "", err
	}

	var result struct {
		Kramdown string `json:"kramdown"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal kramdown: %w", err)
	}
	return result.Kramdown, nil
}

// ChildBlock represents a child block.
type ChildBlock struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"subType,omitempty"`
}

// GetChildBlocks retrieves child blocks of a parent block.
func (c *Client) GetChildBlocks(ctx context.Context, blockID string) ([]ChildBlock, error) {
	req := map[string]string{"id": blockID}

	resp, err := c.Post(ctx, "/api/block/getChildBlocks", req)
	if err != nil {
		return nil, err
	}

	var result []ChildBlock
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal child blocks: %w", err)
	}
	return result, nil
}

// UpdateBlock updates a block's content.
func (c *Client) UpdateBlock(ctx context.Context, blockID, dataType, data string) error {
	req := map[string]string{
		"id":       blockID,
		"dataType": dataType,
		"data":     data,
	}
	_, err := c.Post(ctx, "/api/block/updateBlock", req)
	return err
}

// InsertBlock inserts a block at a specific position.
func (c *Client) InsertBlock(ctx context.Context, dataType, data, nextID, previousID, parentID string) (string, error) {
	req := map[string]string{
		"dataType":   dataType,
		"data":       data,
		"nextID":     nextID,
		"previousID": previousID,
		"parentID":   parentID,
	}

	resp, err := c.Post(ctx, "/api/block/insertBlock", req)
	if err != nil {
		return "", err
	}

	var result []struct {
		DoOperations []struct {
			ID string `json:"id"`
		} `json:"doOperations"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal insert result: %w", err)
	}

	if len(result) > 0 && len(result[0].DoOperations) > 0 {
		return result[0].DoOperations[0].ID, nil
	}
	return "", nil
}

// DeleteBlock deletes a block.
func (c *Client) DeleteBlock(ctx context.Context, blockID string) error {
	req := map[string]string{"id": blockID}
	_, err := c.Post(ctx, "/api/block/deleteBlock", req)
	return err
}

// MoveBlock moves a block to a new position.
func (c *Client) MoveBlock(ctx context.Context, blockID, previousID, parentID string) error {
	req := map[string]string{
		"id":         blockID,
		"previousID": previousID,
		"parentID":   parentID,
	}
	_, err := c.Post(ctx, "/api/block/moveBlock", req)
	return err
}
