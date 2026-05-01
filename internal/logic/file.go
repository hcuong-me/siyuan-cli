// Package logic provides file business logic.
package logic

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"siyuan/internal/siyuan"
)

// FileLogic handles file operations business logic.
type FileLogic struct {
	client *siyuan.Client
}

// FileInfo represents a file or directory entry.
type FileInfo struct {
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Updated   int64  `json:"updated"`
}

// NewFileLogic creates a new FileLogic.
func NewFileLogic() (*FileLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &FileLogic{client: c}, nil
}

// Tree lists files in a directory (non-recursive).
func (l *FileLogic) Tree(ctx context.Context, path string) ([]FileInfo, error) {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/data/" + path
	}

	infos, err := l.client.ReadDir(ctx, path)
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, len(infos))
	for i, info := range infos {
		files[i] = FileInfo{
			IsDir:     info.IsDir,
			IsSymlink: info.IsSymlink,
			Name:      info.Name,
			Path:      filepath.Join(path, info.Name),
			Updated:   info.Updated,
		}
	}

	return files, nil
}

// Read reads a file's content.
func (l *FileLogic) Read(ctx context.Context, path string) (string, error) {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/data/" + path
	}

	return l.client.GetFile(ctx, path)
}

// Write writes content to a file.
func (l *FileLogic) Write(ctx context.Context, path string, content string, isDir bool) error {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/data/" + path
	}

	return l.client.PutFile(ctx, path, content, isDir)
}

// Remove removes a file or directory.
func (l *FileLogic) Remove(ctx context.Context, path string) error {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/data/" + path
	}

	return l.client.RemoveFile(ctx, path)
}

// Rename renames or moves a file.
func (l *FileLogic) Rename(ctx context.Context, oldPath, newPath string) error {
	// Ensure paths start with /
	if !strings.HasPrefix(oldPath, "/") {
		oldPath = "/data/" + oldPath
	}
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/data/" + newPath
	}

	return l.client.RenameFile(ctx, oldPath, newPath)
}

// ValidateFilePath validates a file path.
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain ..")
	}
	return nil
}
