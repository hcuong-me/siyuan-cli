// Package siyuan provides a client for the SiYuan API.
package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"siyuan/internal/config"
)

// Client is an HTTP client for the SiYuan API.
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// Response is the standard response format from SiYuan API.
type Response struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// New creates a new SiYuan API client.
func New() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// NewWithConfig creates a client with the given configuration.
func NewWithConfig(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// SetHTTPClient allows replacing the default HTTP client (useful for testing).
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// Post makes a POST request to the API.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+c.config.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", apiResp.Code, apiResp.Msg)
	}

	return &apiResp, nil
}

// Get makes a GET request to the API (note: SiYuan uses POST mostly).
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Post(ctx, path, nil)
}
