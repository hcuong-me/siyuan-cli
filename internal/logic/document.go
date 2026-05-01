// Package logic provides document business logic.
package logic

import (
	"context"
	"fmt"

	"siyuan/internal/siyuan"
)

// DocumentInfo represents a document with full information.
type DocumentInfo struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Children []DocumentInfo `json:"children,omitempty"`
}

// DocumentLogic handles document business logic.
type DocumentLogic struct {
	client *siyuan.Client
}

// NewDocumentLogic creates a new DocumentLogic.
func NewDocumentLogic() (*DocumentLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &DocumentLogic{client: c}, nil
}

// ListTree returns the document tree for a notebook with names populated.
func (l *DocumentLogic) ListTree(ctx context.Context, notebookID string, maxListCount int) ([]DocumentInfo, error) {
	treeResp, err := l.client.ListDocTree(ctx, notebookID, maxListCount)
	if err != nil {
		return nil, err
	}

	// Convert tree nodes to document info with names
	return l.enrichTreeNodes(ctx, treeResp.Tree)
}

// enrichTreeNodes converts TreeNodes to DocumentInfo by fetching names.
func (l *DocumentLogic) enrichTreeNodes(ctx context.Context, nodes []siyuan.TreeNode) ([]DocumentInfo, error) {
	result := make([]DocumentInfo, len(nodes))

	for i, node := range nodes {
		// Get the human-readable path (name) for this document
		hpath, err := l.client.GetHPathByID(ctx, node.ID)
		if err != nil {
			// If we can't get the name, use the ID as fallback
			hpath = node.ID
		}

		result[i] = DocumentInfo{
			ID:   node.ID,
			Name: hpath,
			Path: hpath,
		}

		// Recursively process children
		if len(node.Children) > 0 {
			children, err := l.enrichTreeNodes(ctx, node.Children)
			if err != nil {
				return nil, err
			}
			result[i].Children = children
		}
	}

	return result, nil
}

// Get retrieves a document's content by path.
func (l *DocumentLogic) Get(ctx context.Context, notebookID, path string) (*siyuan.DocumentContent, error) {
	// Get document ID from path
	ids, err := l.client.GetIDsByHPath(ctx, notebookID, path)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("document not found: %s", path)
	}

	// Get content
	return l.client.GetDocumentContent(ctx, ids[0])
}

// Create creates a new document with Markdown content.
func (l *DocumentLogic) Create(ctx context.Context, notebookID, path, markdown string) (string, error) {
	return l.client.CreateDocumentWithMarkdown(ctx, notebookID, path, markdown)
}

// Update updates a document's content (appends new content).
func (l *DocumentLogic) Update(ctx context.Context, notebookID, path, markdown string) error {
	// Get document ID from path
	ids, err := l.client.GetIDsByHPath(ctx, notebookID, path)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("document not found: %s", path)
	}

	// Append block
	return l.client.AppendBlock(ctx, ids[0], "markdown", markdown)
}

// Remove removes a document.
func (l *DocumentLogic) Remove(ctx context.Context, notebookID, path string) error {
	return l.client.RemoveDocument(ctx, notebookID, path)
}
