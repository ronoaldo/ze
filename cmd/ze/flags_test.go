package main

import (
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		env         map[string]string
		wantURL     string
		wantTimeout time.Duration
		wantVersion  bool
		wantErr     bool
	}{
		{
			name:        "Default values",
			args:        []string{},
			env:         map[string]string{},
			wantURL:     "http://localhost:1234",
			wantTimeout: 5 * time.Minute,
			wantVersion:  false,
			wantErr:     false,
		},
		{
			name:        "Environment variables",
			args:        []string{},
			env:         map[string]string{"LLAMA_URL": "http://env-url:8080", "LLAMA_TIMEOUT": "30s"},
			wantURL:     "http://env-url:8080",
			wantTimeout: 30 * time.Second,
			wantVersion:  false,
			wantErr:     false,
		},
		{
			name:        "Flags override environment",
			args:        []string{"-url", "http://flag-url:9000", "-timeout", "10s"},
			env:         map[string]string{"LLAMA_URL": "http://env-url:8080", "LLAMA_TIMEOUT": "30s"},
			wantURL:     "http://flag-url:9000",
			wantTimeout: 10 * time.Second,
			wantVersion:  false,
			wantErr:     false,
		},
		{
			name:        "Flag with equals sign",
			args:        []string{"-url=http://equals-url:7000", "-timeout=5m"},
			env:         map[string]string{},
			wantURL:     "http://equals-url:7000",
			wantTimeout: 5 * time.Minute,
			wantVersion:  false,
			wantErr:     false,
		},
		{
			name:        "Version flag",
			args:        []string{"-version"},
			env:         map[string]string{},
			wantURL:     "http://localhost:1234",
			wantTimeout: 5 * time.Minute,
			wantVersion:  true,
			wantErr:     false,
		},
		{
			name:        "Short version flag",
			args:        []string{"-v"},
			env:         map[string]string{},
			wantURL:     "http://localhost:1234",
			wantTimeout: 5 * time.Minute,
			wantVersion:  true,
			wantErr:     false,
		},
		{
			name:        "Invalid timeout",
			args:        []string{"-timeout", "invalid"},
			env:         map[string]string{},
			wantURL:     "http://localhost:1234",
			wantTimeout: 0,
			wantVersion:  false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseConfig(tt.args, tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cfg.URL != tt.wantURL {
					t.Errorf("ParseConfig() URL = %v, want %v", cfg.URL, tt.wantURL)
				}
				if cfg.Timeout != tt.wantTimeout {
					t.Errorf("ParseConfig() Timeout = %v, want %v", cfg.Timeout, tt.wantTimeout)
				}
				if cfg.Version != tt.wantVersion {
					t.Errorf("ParseConfig() Version = %v, want %v", cfg.Version, tt.wantVersion)
				}
			}
		})
	}
}
