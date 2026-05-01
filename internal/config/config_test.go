package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultBaseURL(t *testing.T) {
	// Clear env vars
	os.Unsetenv("SIYUAN_BASE_URL")
	os.Setenv("SIYUAN_TOKEN", "test-token")
	defer os.Unsetenv("SIYUAN_TOKEN")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://127.0.0.1:6806" {
		t.Errorf("BaseURL = %q, want default", cfg.BaseURL)
	}
}

func TestLoad_CustomBaseURL(t *testing.T) {
	os.Setenv("SIYUAN_BASE_URL", "http://custom:8080")
	os.Setenv("SIYUAN_TOKEN", "test-token")
	defer os.Unsetenv("SIYUAN_BASE_URL")
	defer os.Unsetenv("SIYUAN_TOKEN")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://custom:8080" {
		t.Errorf("BaseURL = %q, want custom", cfg.BaseURL)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	os.Unsetenv("SIYUAN_TOKEN")
	os.Setenv("SIYUAN_BASE_URL", "http://127.0.0.1:6806")

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
