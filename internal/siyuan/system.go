// Package siyuan provides system-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetVersion retrieves the SiYuan system version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	resp, err := c.Post(ctx, "/api/system/version", nil)
	if err != nil {
		return "", err
	}

	var version string
	if err := json.Unmarshal(resp.Data, &version); err != nil {
		return "", fmt.Errorf("failed to unmarshal version: %w", err)
	}
	return version, nil
}

// GetCurrentTime retrieves the SiYuan system current time.
func (c *Client) GetCurrentTime(ctx context.Context) (int64, error) {
	resp, err := c.Post(ctx, "/api/system/currentTime", nil)
	if err != nil {
		return 0, err
	}

	var timestamp int64
	if err := json.Unmarshal(resp.Data, &timestamp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal timestamp: %w", err)
	}
	return timestamp, nil
}

// BootProgress represents the boot progress information.
type BootProgress struct {
	Details  string `json:"details"`
	Progress int    `json:"progress"`
}

// GetBootProgress retrieves the SiYuan boot progress.
func (c *Client) GetBootProgress(ctx context.Context) (*BootProgress, error) {
	resp, err := c.Post(ctx, "/api/system/bootProgress", nil)
	if err != nil {
		return nil, err
	}

	var result BootProgress
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal boot progress: %w", err)
	}
	return &result, nil
}
