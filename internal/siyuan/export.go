// Package siyuan provides export-related API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// ExportHTMLResponse represents the response from exporting to HTML.
type ExportHTMLResponse struct {
	Content string `json:"content"`
	Folder  string `json:"folder"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

// ExportHTML exports a document as HTML.
func (c *Client) ExportHTML(ctx context.Context, docID string) (*ExportHTMLResponse, error) {
	req := map[string]interface{}{
		"id":  docID,
		"pdf": false,
	}

	resp, err := c.Post(ctx, "/api/export/exportHTML", req)
	if err != nil {
		return nil, err
	}

	var result ExportHTMLResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal HTML export: %w", err)
	}
	return &result, nil
}

// ExportPDFResponse represents the response from exporting to PDF.
type ExportPDFResponse struct {
	Path string `json:"path"`
}

// ExportPDF exports a document as PDF.
func (c *Client) ExportPDF(ctx context.Context, docID, savePath string) (*ExportPDFResponse, error) {
	req := map[string]interface{}{
		"id":           docID,
		"savePath":     savePath,
		"removeAssets": false,
	}

	resp, err := c.Post(ctx, "/api/export/exportPDF", req)
	if err != nil {
		return nil, err
	}

	// Check if response contains data
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return &ExportPDFResponse{Path: "PDF exported (no path returned)"}, nil
	}

	var result ExportPDFResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PDF export: %w", err)
	}
	return &result, nil
}

// ExportDocxResponse represents the response from exporting to DOCX.
type ExportDocxResponse struct {
	Path string `json:"path"`
}

// ExportDocx exports a document as DOCX.
func (c *Client) ExportDocx(ctx context.Context, docID, savePath string) (*ExportDocxResponse, error) {
	req := map[string]interface{}{
		"id":           docID,
		"savePath":     savePath,
		"removeAssets": false,
	}

	resp, err := c.Post(ctx, "/api/export/exportDocx", req)
	if err != nil {
		return nil, err
	}

	var result ExportDocxResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DOCX export: %w", err)
	}
	return &result, nil
}

// ExportResourcesResponse represents the response from exporting resources.
type ExportResourcesResponse struct {
	Path string `json:"path"`
}

// ExportResources exports files and folders as a ZIP.
func (c *Client) ExportResources(ctx context.Context, paths []string, name string) (*ExportResourcesResponse, error) {
	req := map[string]interface{}{
		"paths": paths,
		"name":  name,
	}

	resp, err := c.Post(ctx, "/api/export/exportResources", req)
	if err != nil {
		return nil, err
	}

	var result ExportResourcesResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resources export: %w", err)
	}
	return &result, nil
}
