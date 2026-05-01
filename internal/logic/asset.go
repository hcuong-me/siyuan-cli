// Package logic provides asset business logic.
package logic

import (
	"context"
	"fmt"

	"siyuan/internal/siyuan"
)

// AssetLogic handles asset business logic.
type AssetLogic struct {
	client *siyuan.Client
}

// Asset represents an asset file.
type Asset struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Updated int64  `json:"updated"`
}

// NewAssetLogic creates a new AssetLogic.
func NewAssetLogic() (*AssetLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &AssetLogic{client: c}, nil
}

// Upload uploads an asset file.
func (l *AssetLogic) Upload(ctx context.Context, filePath string) (*siyuan.UploadAssetResult, error) {
	return l.client.UploadAsset(ctx, filePath)
}

// Unused returns all unused asset files.
func (l *AssetLogic) Unused(ctx context.Context) ([]Asset, error) {
	infos, err := l.client.GetUnusedAssets(ctx)
	if err != nil {
		return nil, err
	}

	assets := make([]Asset, len(infos))
	for i, info := range infos {
		assets[i] = Asset{
			Path:    info.Path,
			Size:    info.Size,
			Updated: info.Updated,
		}
	}
	return assets, nil
}

// Clean removes all unused assets.
func (l *AssetLogic) Clean(ctx context.Context) error {
	return l.client.RemoveUnusedAssets(ctx)
}

// Remove removes a specific asset file.
func (l *AssetLogic) Remove(ctx context.Context, path string) error {
	return l.client.RemoveAsset(ctx, path)
}

// ValidateAssetPath checks if the path is a valid asset path.
func ValidateAssetPath(path string) error {
	if path == "" {
		return fmt.Errorf("asset path is required")
	}
	return nil
}
