// Package logic provides template business logic.
package logic

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"siyuan/internal/siyuan"
)

// TemplateLogic handles template business logic.
type TemplateLogic struct {
	client *siyuan.Client
}

// Template represents a template file.
type Template struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Updated int64  `json:"updated"`
}

// NewTemplateLogic creates a new TemplateLogic.
func NewTemplateLogic() (*TemplateLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &TemplateLogic{client: c}, nil
}

// List returns all templates.
func (l *TemplateLogic) List(ctx context.Context) ([]Template, error) {
	infos, err := l.client.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	templates := make([]Template, 0, len(infos))
	for _, info := range infos {
		// Skip directories and non-markdown files
		if info.IsDir {
			continue
		}
		// Only include .md files
		if !strings.HasSuffix(info.Name, ".md") {
			continue
		}
		templates = append(templates, Template{
			Name:    info.Name,
			Path:    filepath.Join("/data/templates", info.Name),
			Updated: info.Updated,
		})
	}

	return templates, nil
}

// Get returns the content of a template.
func (l *TemplateLogic) Get(ctx context.Context, path string) (string, error) {
	// Ensure path is absolute or starts with /data/templates
	if !strings.HasPrefix(path, "/data/templates") {
		path = filepath.Join("/data/templates", path)
	}

	return l.client.GetTemplate(ctx, path)
}

// Render renders a template and returns the rendered content.
func (l *TemplateLogic) Render(ctx context.Context, docID, templatePath string) (*siyuan.RenderTemplateResult, error) {
	// Ensure template path is absolute
	if !strings.HasPrefix(templatePath, "/data/templates") {
		templatePath = filepath.Join("/data/templates", templatePath)
	}

	return l.client.RenderTemplate(ctx, docID, templatePath)
}

// Remove removes a template file.
func (l *TemplateLogic) Remove(ctx context.Context, path string) error {
	// Ensure path is absolute
	if !strings.HasPrefix(path, "/data/templates") {
		path = filepath.Join("/data/templates", path)
	}

	return l.client.RemoveTemplate(ctx, path)
}

// ValidateTemplatePath checks if the path is a valid template path.
func ValidateTemplatePath(path string) error {
	if path == "" {
		return fmt.Errorf("template path is required")
	}
	if !strings.HasSuffix(path, ".md") {
		return fmt.Errorf("template path must end with .md")
	}
	return nil
}
