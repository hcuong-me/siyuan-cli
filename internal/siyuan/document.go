// Package siyuan provides document-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// TreeNode represents a node in the document tree.
type TreeNode struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Icon     string     `json:"icon"`
	Children []TreeNode `json:"children,omitempty"`
}

// ListDocTreeResponse represents the response from /api/filetree/listDocTree.
type ListDocTreeResponse struct {
	Tree []TreeNode `json:"tree"`
}

// ListDocTree retrieves the document tree for a notebook.
func (c *Client) ListDocTree(ctx context.Context, notebookID string, maxListCount int) (*ListDocTreeResponse, error) {
	req := map[string]interface{}{
		"notebook": notebookID,
		"path":     "/",
	}
	if maxListCount > 0 {
		req["maxListCount"] = maxListCount
	}

	resp, err := c.Post(ctx, "/api/filetree/listDocTree", req)
	if err != nil {
		return nil, err
	}

	var result ListDocTreeResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc tree: %w", err)
	}
	return &result, nil
}

// GetIDsByHPath retrieves document IDs by human-readable path.
func (c *Client) GetIDsByHPath(ctx context.Context, notebookID, path string) ([]string, error) {
	req := map[string]string{
		"notebook": notebookID,
		"path":     path,
	}

	resp, err := c.Post(ctx, "/api/filetree/getIDsByHPath", req)
	if err != nil {
		return nil, err
	}

	var ids []string
	if err := json.Unmarshal(resp.Data, &ids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IDs: %w", err)
	}
	return ids, nil
}

// DocumentContent represents a document's content.
type DocumentContent struct {
	HPath   string `json:"hPath"`
	Content string `json:"content"`
}

// GetDocumentContent retrieves a document's Markdown content.
func (c *Client) GetDocumentContent(ctx context.Context, docID string) (*DocumentContent, error) {
	req := map[string]string{"id": docID}

	resp, err := c.Post(ctx, "/api/export/exportMdContent", req)
	if err != nil {
		return nil, err
	}

	var result DocumentContent
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document content: %w", err)
	}
	return &result, nil
}

// CreateDocumentWithMarkdown creates a new document with Markdown content.
func (c *Client) CreateDocumentWithMarkdown(ctx context.Context, notebookID, path, markdown string) (string, error) {
	req := map[string]string{
		"notebook": notebookID,
		"path":     path,
		"markdown": markdown,
	}

	resp, err := c.Post(ctx, "/api/filetree/createDocWithMd", req)
	if err != nil {
		return "", err
	}

	var docID string
	if err := json.Unmarshal(resp.Data, &docID); err != nil {
		return "", fmt.Errorf("failed to unmarshal document ID: %w", err)
	}
	return docID, nil
}

// RemoveDocument removes a document.
func (c *Client) RemoveDocument(ctx context.Context, notebookID, path string) error {
	req := map[string]string{
		"notebook": notebookID,
		"path":     path,
	}
	_, err := c.Post(ctx, "/api/filetree/removeDoc", req)
	return err
}

// AppendBlock appends a block to a document.
func (c *Client) AppendBlock(ctx context.Context, parentID, dataType, data string) error {
	req := map[string]string{
		"dataType": dataType,
		"data":     data,
		"parentID": parentID,
	}
	_, err := c.Post(ctx, "/api/block/appendBlock", req)
	return err
}

// GetHPathByID retrieves the human-readable path for a document by ID.
func (c *Client) GetHPathByID(ctx context.Context, docID string) (string, error) {
	req := map[string]string{"id": docID}

	resp, err := c.Post(ctx, "/api/filetree/getHPathByID", req)
	if err != nil {
		return "", err
	}

	var hpath string
	if err := json.Unmarshal(resp.Data, &hpath); err != nil {
		return "", fmt.Errorf("failed to unmarshal hpath: %w", err)
	}
	return hpath, nil
}
