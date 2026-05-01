// Package siyuan provides template API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// TemplateInfo represents a file in the templates directory.
type TemplateInfo struct {
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
	Name      string `json:"name"`
	Updated   int64  `json:"updated"`
}

// RenderTemplateResult represents the result of rendering a template.
type RenderTemplateResult struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

// ListTemplates lists all template files in the templates directory.
func (c *Client) ListTemplates(ctx context.Context) ([]TemplateInfo, error) {
	req := map[string]string{
		"path": "/data/templates",
	}

	resp, err := c.Post(ctx, "/api/file/readDir", req)
	if err != nil {
		return nil, err
	}

	var result []TemplateInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal templates: %w", err)
	}
	return result, nil
}

// GetTemplate reads a template file content.
func (c *Client) GetTemplate(ctx context.Context, path string) (string, error) {
	req := map[string]string{
		"path": path,
	}

	resp, err := c.Post(ctx, "/api/file/getFile", req)
	if err != nil {
		return "", err
	}

	// getFile returns the content directly as a string in data
	var content string
	if err := json.Unmarshal(resp.Data, &content); err != nil {
		// Try to return as raw string if unmarshal fails
		return string(resp.Data), nil
	}
	return content, nil
}

// RenderTemplate renders a template file and returns the result.
func (c *Client) RenderTemplate(ctx context.Context, docID, templatePath string) (*RenderTemplateResult, error) {
	req := map[string]string{
		"id":   docID,
		"path": templatePath,
	}

	resp, err := c.Post(ctx, "/api/template/render", req)
	if err != nil {
		return nil, err
	}

	var result RenderTemplateResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal render result: %w", err)
	}
	return &result, nil
}

// RemoveTemplate removes a template file.
func (c *Client) RemoveTemplate(ctx context.Context, path string) error {
	req := map[string]string{
		"path": path,
	}

	_, err := c.Post(ctx, "/api/file/removeFile", req)
	return err
}
