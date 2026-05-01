// Package config handles configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	BaseURL string // SiYuan API base URL
	Token   string // SiYuan API token
}

// Load reads configuration from environment variables.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	baseURL := os.Getenv("SIYUAN_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:6806" // Default value
	}

	token := os.Getenv("SIYUAN_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("SIYUAN_TOKEN environment variable is required. Get your token from Settings > About in SiYuan")
	}

	return &Config{
		BaseURL: baseURL,
		Token:   token,
	}, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}
