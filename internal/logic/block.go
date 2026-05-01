// Package logic provides block business logic.
package logic

import (
	"context"
	"fmt"

	"siyuan/internal/siyuan"
)

// BlockLogic handles block business logic.
type BlockLogic struct {
	client *siyuan.Client
}

// NewBlockLogic creates a new BlockLogic.
func NewBlockLogic() (*BlockLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &BlockLogic{client: c}, nil
}

// GetKramdown retrieves a block's kramdown source.
func (l *BlockLogic) GetKramdown(ctx context.Context, blockID string) (string, error) {
	return l.client.GetBlockKramdown(ctx, blockID)
}

// GetChildren retrieves child blocks of a parent block.
func (l *BlockLogic) GetChildren(ctx context.Context, blockID string) ([]siyuan.ChildBlock, error) {
	return l.client.GetChildBlocks(ctx, blockID)
}

// Update updates a block's content.
func (l *BlockLogic) Update(ctx context.Context, blockID, content string) error {
	return l.client.UpdateBlock(ctx, blockID, "markdown", content)
}

// Append appends a block to a parent.
func (l *BlockLogic) Append(ctx context.Context, parentID, content string) (string, error) {
	return l.client.InsertBlock(ctx, "markdown", content, "", "", parentID)
}

// InsertAfter inserts a block after a specific block.
func (l *BlockLogic) InsertAfter(ctx context.Context, previousID, content string) (string, error) {
	return l.client.InsertBlock(ctx, "markdown", content, "", previousID, "")
}

// Delete deletes a block.
func (l *BlockLogic) Delete(ctx context.Context, blockID string) error {
	return l.client.DeleteBlock(ctx, blockID)
}

// Move moves a block to a new position.
func (l *BlockLogic) Move(ctx context.Context, blockID, previousID, parentID string) error {
	if previousID == "" && parentID == "" {
		return fmt.Errorf("either previousID or parentID must be specified")
	}
	return l.client.MoveBlock(ctx, blockID, previousID, parentID)
}
