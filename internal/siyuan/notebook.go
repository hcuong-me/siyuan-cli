// Package siyuan provides notebook-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListNotebooks retrieves all notebooks.
func (c *Client) ListNotebooks(ctx context.Context) (*ListNotebooksResponse, error) {
	resp, err := c.Post(ctx, "/api/notebook/lsNotebooks", nil)
	if err != nil {
		return nil, err
	}

	var result ListNotebooksResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notebooks: %w", err)
	}
	return &result, nil
}

// CreateNotebook creates a new notebook.
func (c *Client) CreateNotebook(ctx context.Context, name string) (*Notebook, error) {
	req := CreateNotebookRequest{Name: name}
	resp, err := c.Post(ctx, "/api/notebook/createNotebook", req)
	if err != nil {
		return nil, err
	}

	var result CreateNotebookResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notebook: %w", err)
	}
	return &result.Notebook, nil
}

// RenameNotebook renames a notebook.
func (c *Client) RenameNotebook(ctx context.Context, notebookID, name string) error {
	req := RenameNotebookRequest{Notebook: notebookID, Name: name}
	_, err := c.Post(ctx, "/api/notebook/renameNotebook", req)
	return err
}

// RemoveNotebook removes a notebook.
func (c *Client) RemoveNotebook(ctx context.Context, notebookID string) error {
	req := NotebookIDRequest{Notebook: notebookID}
	_, err := c.Post(ctx, "/api/notebook/removeNotebook", req)
	return err
}

// OpenNotebook opens a closed notebook.
func (c *Client) OpenNotebook(ctx context.Context, notebookID string) error {
	req := NotebookIDRequest{Notebook: notebookID}
	_, err := c.Post(ctx, "/api/notebook/openNotebook", req)
	return err
}

// CloseNotebook closes an open notebook.
func (c *Client) CloseNotebook(ctx context.Context, notebookID string) error {
	req := NotebookIDRequest{Notebook: notebookID}
	_, err := c.Post(ctx, "/api/notebook/closeNotebook", req)
	return err
}
