// Package logic provides attribute business logic.
package logic

import (
	"context"

	"siyuan/internal/siyuan"
)

// AttrLogic handles attribute business logic.
type AttrLogic struct {
	client *siyuan.Client
}

// NewAttrLogic creates a new AttrLogic.
func NewAttrLogic() (*AttrLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &AttrLogic{client: c}, nil
}

// Get retrieves all attributes of a block.
func (l *AttrLogic) Get(ctx context.Context, blockID string) (siyuan.BlockAttrs, error) {
	return l.client.GetBlockAttrs(ctx, blockID)
}

// Set sets attributes on a block.
func (l *AttrLogic) Set(ctx context.Context, blockID string, attrs map[string]string) error {
	return l.client.SetBlockAttrs(ctx, blockID, attrs)
}

// SetSingle sets a single attribute on a block.
func (l *AttrLogic) SetSingle(ctx context.Context, blockID, key, value string) error {
	return l.client.SetBlockAttrs(ctx, blockID, map[string]string{key: value})
}

// Reset removes a specific attribute from a block.
func (l *AttrLogic) Reset(ctx context.Context, blockID, key string) error {
	return l.client.ResetBlockAttr(ctx, blockID, key)
}

// GetBookmarkLabels retrieves all bookmark labels.
func (l *AttrLogic) GetBookmarkLabels(ctx context.Context) ([]siyuan.BookmarkLabel, error) {
	return l.client.GetBookmarkLabels(ctx)
}
