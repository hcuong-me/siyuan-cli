// Package siyuan provides repository snapshot API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// SnapshotInfo represents a repository snapshot.
type SnapshotInfo struct {
	ID      string `json:"id"`
	Memo    string `json:"memo"`
	Created int64  `json:"created"`
	Count   int    `json:"count"`
	Size    int64  `json:"size"`
}

// ListSnapshots returns all repository snapshots.
func (c *Client) ListSnapshots(ctx context.Context) ([]SnapshotInfo, error) {
	resp, err := c.Post(ctx, "/api/repo/getRepoSnapshots", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result []SnapshotInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshots: %w", err)
	}
	return result, nil
}

// GetCurrentSnapshot returns the current snapshot information.
func (c *Client) GetCurrentSnapshot(ctx context.Context) (*SnapshotInfo, error) {
	resp, err := c.Post(ctx, "/api/repo/getRepoSnapshot", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result SnapshotInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return &result, nil
}

// CreateSnapshot creates a new repository snapshot.
func (c *Client) CreateSnapshot(ctx context.Context, memo string) (*SnapshotInfo, error) {
	req := map[string]string{
		"memo": memo,
	}
	resp, err := c.Post(ctx, "/api/repo/createRepoSnapshot", req)
	if err != nil {
		return nil, err
	}

	var result SnapshotInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return &result, nil
}

// RestoreSnapshot restores the repository to a specific snapshot.
func (c *Client) RestoreSnapshot(ctx context.Context, id string) error {
	req := map[string]string{
		"id": id,
	}
	_, err := c.Post(ctx, "/api/repo/checkoutRepo", req)
	return err
}

// RemoveSnapshot removes a repository snapshot.
func (c *Client) RemoveSnapshot(ctx context.Context, id string) error {
	req := map[string]string{
		"id": id,
	}
	_, err := c.Post(ctx, "/api/repo/removeRepoSnapshot", req)
	return err
}
