// Package siyuan provides asset API methods.
package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// UploadAssetResult represents the result of uploading assets.
type UploadAssetResult struct {
	SuccMap  map[string]string `json:"succMap"`
	ErrFiles []string          `json:"errFiles"`
}

// AssetInfo represents an asset file.
type AssetInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Updated int64  `json:"updated"`
}

// UploadAsset uploads an asset file to SiYuan.
func (c *Client) UploadAsset(ctx context.Context, filePath string) (*UploadAssetResult, error) {
	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add assetsDirPath field
	if err := writer.WriteField("assetsDirPath", "/assets/"); err != nil {
		return nil, fmt.Errorf("failed to write field: %w", err)
	}

	// Add file
	part, err := writer.CreateFormFile("file[]", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Make request
	url := c.config.BaseURL + "/api/asset/upload"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Token "+c.config.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", apiResp.Code, apiResp.Msg)
	}

	var result UploadAssetResult
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload result: %w", err)
	}

	return &result, nil
}

// GetUnusedAssets returns a list of unused asset files.
func (c *Client) GetUnusedAssets(ctx context.Context) ([]AssetInfo, error) {
	resp, err := c.Post(ctx, "/api/asset/getUnusedAssets", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var result []AssetInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal unused assets: %w", err)
	}
	return result, nil
}

// RemoveUnusedAssets removes all unused asset files.
func (c *Client) RemoveUnusedAssets(ctx context.Context) error {
	_, err := c.Post(ctx, "/api/asset/removeUnusedAssets", map[string]interface{}{})
	return err
}

// RemoveAsset removes a specific asset file.
func (c *Client) RemoveAsset(ctx context.Context, path string) error {
	req := map[string]string{
		"path": path,
	}
	_, err := c.Post(ctx, "/api/file/removeFile", req)
	return err
}
