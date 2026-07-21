package tools

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebFetchTool_Execute(t *testing.T) {
	tool := &WebFetchTool{}

	tests := []struct {
		name         string
		handler      func(w http.ResponseWriter, r *http.Request)
		args         map[string]interface{}
		wantSummary  string
		wantErr      bool
		checkContent func(t *testing.T, content string)
	}{
		{
			name: "Success - Plain Text",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w, "hello world")
			},
			args:        map[string]interface{}{"url": ""},
			wantSummary: "Fetched content from",
			wantErr:     false,
			checkContent: func(t *testing.T, content string) {
				if content != "hello world" {
					t.Errorf("expected 'hello world', got '%s'", content)
				}
			},
		},
		{
			name: "Success - JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"key": "value"}`)
			},
			args:        map[string]interface{}{"url": ""},
			wantSummary: "Fetched content from",
			wantErr:     false,
			checkContent: func(t *testing.T, content string) {
				if content != `{"key": "value"}` {
					t.Errorf("expected '{\"key\": \"value\"}', got '%s'", content)
				}
			},
		},
		{
			name: "Success - HTML",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, "<html><body><h1 >Hello</h1><p >World</p></body></html>")
			},
			args:        map[string]interface{}{"url": ""},
			wantSummary: "Fetched content from",
			wantErr:     false,
			checkContent: func(t *testing.T, content string) {
				// Due to simple regex replacement with spaces, we expect some whitespace
				// but the core words should be there.
				if !strings.Contains(content, "Hello") || !strings.Contains(content, "World") {
					t.Errorf("expected content to contain 'Hello' and 'World', got '%s'", content)
				}
			},
		},
		{
			name: "Error - Invalid URL",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			args:    map[string]interface{}{"url": " ://invalid"},
			wantErr: true,
		},
		{
			name: "Error - HTTP 404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			args:    map[string]interface{}{"url": ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			testArgs := tt.args
			if tt.args["url"] == "" {
				testArgs["url"] = server.URL
			}

			result, err := tool.Execute(testArgs)

			if (err != nil) != tt.wantErr {
				t.Errorf("WebFetchTool.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !strings.Contains(result.Summary, tt.wantSummary) {
					t.Errorf("expected summary to contain '%s', got '%s'", tt.wantSummary, result.Summary)
				}
				if tt.checkContent != nil {
					tt.checkContent(t, result.FullResult)
				}
			}
		})
	}
}
