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

type listSnapshotsResponse struct {
	Snapshots  []SnapshotInfo `json:"snapshots"`
	PageCount  int            `json:"pageCount"`
	TotalCount int            `json:"totalCount"`
}

// ListSnapshots returns all repository snapshots.
func (c *Client) ListSnapshots(ctx context.Context) ([]SnapshotInfo, error) {
	resp, err := c.Post(ctx, "/api/repo/getRepoSnapshots", map[string]any{"page": 1})
	if err != nil {
		return nil, err
	}

	var result listSnapshotsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshots: %w", err)
	}
	all := append([]SnapshotInfo(nil), result.Snapshots...)
	for page := 2; page <= result.PageCount; page++ {
		pageResp, pageErr := c.Post(ctx, "/api/repo/getRepoSnapshots", map[string]any{"page": page})
		if pageErr != nil {
			return nil, pageErr
		}
		var pageResult listSnapshotsResponse
		if unmarshalErr := json.Unmarshal(pageResp.Data, &pageResult); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshots page %d: %w", page, unmarshalErr)
		}
		all = append(all, pageResult.Snapshots...)
	}
	return all, nil
}

// CreateSnapshot creates a new repository snapshot.
func (c *Client) CreateSnapshot(ctx context.Context, memo string) error {
	req := map[string]string{
		"memo": memo,
	}
	_, err := c.Post(ctx, "/api/repo/createSnapshot", req)
	return err
}

// RestoreSnapshot restores the repository to a specific snapshot.
func (c *Client) RestoreSnapshot(ctx context.Context, id string) error {
	req := map[string]string{
		"id": id,
	}
	_, err := c.Post(ctx, "/api/repo/checkoutRepo", req)
	return err
}
