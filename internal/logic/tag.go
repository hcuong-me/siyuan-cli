// Package logic provides tag business logic.
package logic

import (
	"context"

	"siyuan/internal/siyuan"
)

// TagLogic handles tag business logic.
type TagLogic struct {
	client *siyuan.Client
}

// NewTagLogic creates a new TagLogic.
func NewTagLogic() (*TagLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &TagLogic{client: c}, nil
}

// List returns all tags.
func (l *TagLogic) List(ctx context.Context) ([]siyuan.Tag, error) {
	return l.client.GetTags(ctx)
}

// Search searches for tags.
func (l *TagLogic) Search(ctx context.Context, keyword string) (*siyuan.TagSearchResult, error) {
	return l.client.SearchTags(ctx, keyword)
}

// Rename renames a tag.
func (l *TagLogic) Rename(ctx context.Context, oldLabel, newLabel string) error {
	return l.client.RenameTag(ctx, oldLabel, newLabel)
}

// Remove removes a tag.
func (l *TagLogic) Remove(ctx context.Context, label string) error {
	return l.client.RemoveTag(ctx, label)
}

// GetDocsByTag returns documents with a specific tag.
func (l *TagLogic) GetDocsByTag(ctx context.Context, label string) ([]string, error) {
	return l.client.GetDocsByTag(ctx, label)
}
