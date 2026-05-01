package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultBaseURL(t *testing.T) {
	// Clear env vars
	_ = os.Unsetenv("SIYUAN_BASE_URL")
	_ = os.Setenv("SIYUAN_TOKEN", "test-token")
	defer func() { _ = os.Unsetenv("SIYUAN_TOKEN") }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://127.0.0.1:6806" {
		t.Errorf("BaseURL = %q, want default", cfg.BaseURL)
	}
}

func TestLoad_CustomBaseURL(t *testing.T) {
	_ = os.Setenv("SIYUAN_BASE_URL", "http://custom:8080")
	_ = os.Setenv("SIYUAN_TOKEN", "test-token")
	defer func() { _ = os.Unsetenv("SIYUAN_BASE_URL") }()
	defer func() { _ = os.Unsetenv("SIYUAN_TOKEN") }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://custom:8080" {
		t.Errorf("BaseURL = %q, want custom", cfg.BaseURL)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	_ = os.Unsetenv("SIYUAN_TOKEN")
	_ = os.Setenv("SIYUAN_BASE_URL", "http://127.0.0.1:6806")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for missing token")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg:  Config{BaseURL: "http://localhost:6806", Token: "token"},
		},
		{
			name:    "missing token",
			cfg:     Config{BaseURL: "http://localhost:6806", Token: ""},
			wantErr: true,
		},
		{
			name:    "missing baseurl",
			cfg:     Config{BaseURL: "", Token: "token"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
