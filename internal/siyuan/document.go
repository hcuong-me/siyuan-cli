// Package siyuan provides document-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// TreeNode represents a node in the document tree.
type TreeNode struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Icon     string     `json:"icon"`
	Children []TreeNode `json:"children,omitempty"`
}

// ListDocTreeResponse represents the response from /api/filetree/listDocsByPath.
type ListDocTreeResponse struct {
	Tree []TreeNode `json:"tree"`
}

// DocFileEntry is one document file returned by /api/filetree/listDocsByPath.
type DocFileEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Icon         string `json:"icon"`
	SubFileCount int    `json:"subFileCount"`
}

type listDocsByPathResponse struct {
	Box   string         `json:"box"`
	Path  string         `json:"path"`
	Files []DocFileEntry `json:"files"`
}

// ListDocTree retrieves the document tree for a notebook. maxDepth is a local
// traversal bound: zero means unlimited, one returns only root documents.
// listDocsByPath itself has no depth parameter, so it is never sent as a
// server-side list-count hint.
func (c *Client) ListDocTree(ctx context.Context, notebookID string, maxDepth int) (*ListDocTreeResponse, error) {
	tree, err := c.listDocTree(ctx, notebookID, "/", 1, maxDepth)
	if err != nil {
		return nil, err
	}
	return &ListDocTreeResponse{Tree: tree}, nil
}

func (c *Client) listDocTree(ctx context.Context, notebookID, dirPath string, depth, maxDepth int) ([]TreeNode, error) {
	req := map[string]any{
		"notebook": notebookID,
		"path":     dirPath,
	}
	resp, err := c.Post(ctx, "/api/filetree/listDocsByPath", req)
	if err != nil {
		return nil, err
	}

	var result listDocsByPathResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc tree: %w", err)
	}

	nodes := make([]TreeNode, 0, len(result.Files))
	for _, file := range result.Files {
		node := TreeNode{ID: file.ID, Name: file.Name, Path: file.Path, Icon: file.Icon}
		if file.SubFileCount > 0 && (maxDepth <= 0 || depth < maxDepth) {
			children, err := c.listDocTree(ctx, notebookID, strings.TrimSuffix(file.Path, ".sy"), depth+1, maxDepth)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
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

// CreatedDocument is the normalized result returned by createDocWithMd. Older
// servers return only ID; newer servers also return box, storage path, and
// human-readable path.
type CreatedDocument struct {
	ID    string `json:"id"`
	Box   string `json:"box,omitempty"`
	Path  string `json:"path,omitempty"`
	HPath string `json:"hPath,omitempty"`
}

// CreateDocumentWithMarkdown creates a new document with Markdown content and
// preserves the legacy string-ID API.
func (c *Client) CreateDocumentWithMarkdown(ctx context.Context, notebookID, path, markdown string) (string, error) {
	result, err := c.CreateDocumentWithMarkdownResult(ctx, notebookID, path, markdown)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// CreateDocumentWithMarkdownResult accepts both documented response shapes.
func (c *Client) CreateDocumentWithMarkdownResult(ctx context.Context, notebookID, path, markdown string) (CreatedDocument, error) {
	req := map[string]string{
		"notebook": notebookID,
		"path":     path,
		"markdown": markdown,
	}

	resp, err := c.Post(ctx, "/api/filetree/createDocWithMd", req)
	if err != nil {
		return CreatedDocument{}, err
	}

	var docID string
	if err := json.Unmarshal(resp.Data, &docID); err == nil {
		return CreatedDocument{ID: docID}, nil
	}

	// Newer SiYuan versions return the complete document descriptor while
	// older versions return only the ID string. Keep the existing string
	// return type and normalize both response shapes to the document ID.
	var result struct {
		ID    string `json:"id"`
		Box   string `json:"box"`
		Path  string `json:"path"`
		HPath string `json:"hPath"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return CreatedDocument{}, fmt.Errorf("failed to unmarshal document ID: %w", err)
	}
	if result.ID == "" {
		return CreatedDocument{}, fmt.Errorf("failed to unmarshal document ID: response object has no id")
	}
	return CreatedDocument{ID: result.ID, Box: result.Box, Path: result.Path, HPath: result.HPath}, nil
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

// RemoveDocumentByID removes a document using its stable block ID.
func (c *Client) RemoveDocumentByID(ctx context.Context, docID string) error {
	_, err := c.Post(ctx, "/api/filetree/removeDocByID", map[string]string{"id": docID})
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
