// Package logic provides export business logic.
package logic

import (
	"context"
	"fmt"
	"os"

	"siyuan/internal/siyuan"
)

// ExportLogic handles export business logic.
type ExportLogic struct {
	client *siyuan.Client
}

// NewExportLogic creates a new ExportLogic.
func NewExportLogic() (*ExportLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &ExportLogic{client: c}, nil
}

// ExportMarkdown exports a document as Markdown.
// This is the same as GetDocumentContent but with file output.
func (l *ExportLogic) ExportMarkdown(ctx context.Context, docID, outputPath string) (*siyuan.DocumentContent, error) {
	content, err := l.client.GetDocumentContent(ctx, docID)
	if err != nil {
		return nil, err
	}

	// Write to file if outputPath is specified
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(content.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write markdown file: %w", err)
		}
	}

	return content, nil
}

// ExportHTML exports a document as HTML.
func (l *ExportLogic) ExportHTML(ctx context.Context, docID, outputPath string) (*siyuan.ExportHTMLResponse, error) {
	resp, err := l.client.ExportHTML(ctx, docID)
	if err != nil {
		return nil, err
	}

	// Write to file if outputPath is specified
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(resp.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write HTML file: %w", err)
		}
	}

	return resp, nil
}

// ExportPDF exports a document as PDF.
func (l *ExportLogic) ExportPDF(ctx context.Context, docID, outputPath string) (*siyuan.ExportPDFResponse, error) {
	// PDF is saved on the SiYuan server. Preserve the caller's server-owned
	// save path instead of silently replacing it with a local temp directory.
	if outputPath == "" {
		outputPath = "/tmp"
	}
	resp, err := l.client.ExportPDF(ctx, docID, outputPath)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExportDocx exports a document as DOCX.
func (l *ExportLogic) ExportDocx(ctx context.Context, docID, outputPath string) (*siyuan.ExportDocxResponse, error) {
	// DOCX is saved on the SiYuan server. Preserve the caller's server-owned
	// save path instead of silently replacing it with a local temp directory.
	if outputPath == "" {
		outputPath = "/tmp"
	}
	resp, err := l.client.ExportDocx(ctx, docID, outputPath)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExportNotebook exports a notebook as Markdown.
func (l *ExportLogic) ExportNotebook(_ context.Context, _, _ string) error {
	// This would require iterating through all documents in the notebook
	// and exporting them. For now, this is a placeholder.
	return fmt.Errorf("notebook export not yet implemented")
}
