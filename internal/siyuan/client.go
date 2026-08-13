// Package siyuan provides a client for the SiYuan API.
package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

// Post makes a POST request to the API and decodes the response envelope.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	respBody, err := c.post(ctx, path, body)
	if err != nil {
		return nil, err
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

// PostRaw makes a POST request and returns the raw response body. Some
// endpoints, such as /api/file/getFile, return file content directly instead
// of the JSON envelope. Error responses still use the envelope.
func (c *Client) PostRaw(ctx context.Context, path string, body interface{}) ([]byte, error) {
	respBody, err := c.post(ctx, path, body)
	if err != nil {
		return nil, err
	}
	var apiResp Response
	if json.Unmarshal(respBody, &apiResp) == nil && apiResp.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", apiResp.Code, apiResp.Msg)
	}
	return respBody, nil
}

func (c *Client) post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	return c.doPost(ctx, path, bodyReader, "application/json")
}

// postMultipart makes an authenticated multipart/form-data request and
// decodes the standard SiYuan response envelope. Form values are kept as
// strings because the putFile endpoint expects path, isDir, and content as
// multipart fields rather than a JSON object.
func (c *Client) postMultipart(ctx context.Context, path string, fields map[string]string) (*Response, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write multipart field %q: %w", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart body: %w", err)
	}

	respBody, err := c.doPost(ctx, path, &body, writer.FormDataContentType())
	if err != nil {
		return nil, err
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

func (c *Client) doPost(ctx context.Context, path string, bodyReader io.Reader, contentType string) ([]byte, error) {
	url := c.config.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Token "+c.config.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return respBody, nil
}

// Get makes a GET request to the API (note: SiYuan uses POST mostly).
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Post(ctx, path, nil)
}
