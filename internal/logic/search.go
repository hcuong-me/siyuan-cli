// Package logic provides search business logic.
package logic

import (
	"context"

	"siyuan/internal/siyuan"
)

// SearchLogic handles search business logic.
type SearchLogic struct {
	client *siyuan.Client
}

// NewSearchLogic creates a new SearchLogic.
func NewSearchLogic() (*SearchLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &SearchLogic{client: c}, nil
}

// SearchBlocks performs full-text search on blocks.
func (l *SearchLogic) SearchBlocks(ctx context.Context, keyword string, page, size int) (*siyuan.SearchBlockResult, error) {
	return l.client.FullTextSearchBlock(ctx, keyword, page, size)
}

// SearchDocs searches for documents.
func (l *SearchLogic) SearchDocs(ctx context.Context, keyword, notebook, path string) ([]siyuan.SearchDocResult, error) {
	return l.client.SearchDocs(ctx, keyword, notebook, path)
}
