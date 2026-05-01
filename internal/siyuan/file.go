// Package siyuan provides file operations API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// FileInfo represents a file or directory entry.
type FileInfo struct {
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
	Name      string `json:"name"`
	Updated   int64  `json:"updated"`
}

// ReadDir lists files in a directory.
func (c *Client) ReadDir(ctx context.Context, path string) ([]FileInfo, error) {
	req := map[string]string{
		"path": path,
	}

	resp, err := c.Post(ctx, "/api/file/readDir", req)
	if err != nil {
		return nil, err
	}

	var result []FileInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal files: %w", err)
	}
	return result, nil
}

// GetFile reads a file's content.
func (c *Client) GetFile(ctx context.Context, path string) (string, error) {
	req := map[string]string{
		"path": path,
	}

	resp, err := c.Post(ctx, "/api/file/getFile", req)
	if err != nil {
		return "", err
	}

	var content string
	if err := json.Unmarshal(resp.Data, &content); err != nil {
		// Return raw data if not valid JSON string
		return string(resp.Data), nil
	}
	return content, nil
}

// PutFile writes content to a file.
func (c *Client) PutFile(ctx context.Context, path string, content string, isDir bool) error {
	req := map[string]interface{}{
		"path":    path,
		"content": content,
		"isDir":   isDir,
	}

	_, err := c.Post(ctx, "/api/file/putFile", req)
	return err
}

// RemoveFile removes a file or directory.
func (c *Client) RemoveFile(ctx context.Context, path string) error {
	req := map[string]string{
		"path": path,
	}

	_, err := c.Post(ctx, "/api/file/removeFile", req)
	return err
}

// RenameFile renames or moves a file.
func (c *Client) RenameFile(ctx context.Context, oldPath, newPath string) error {
	req := map[string]string{
		"path":    oldPath,
		"newPath": newPath,
	}

	_, err := c.Post(ctx, "/api/file/renameFile", req)
	return err
}
