package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// WebFetchArgs defines the arguments for WebFetchTool.
type WebFetchArgs struct {
	URL string `json:"url"`
}

// WebFetchTool implements downloading content from a web URL.
type WebFetchTool struct{}

// Name returns the name of the tool.
func (t *WebFetchTool) Name() string { return "web_fetch" }

// Execute downloads the content from the provided URL and returns it as text.
func (t *WebFetchTool) Execute(args map[string]interface{}) (ToolResult, error) {
	var a WebFetchArgs
	if err := mapToStruct(args, &a); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}

	if a.URL == "" {
		return ToolResult{}, fmt.Errorf("missing 'url' argument")
	}

	parsedURL, err := url.ParseRequestURI(a.URL)
	if err != nil {
		return ToolResult{}, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ToolResult{}, fmt.Errorf("unsupported protocol: %s. Only http and https are allowed", parsedURL.Scheme)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(a.URL)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ToolResult{}, fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	// Limit the size of the body to 2MB
	limitReader := io.LimitReader(resp.Body, 2*1024*1024)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	content := string(bodyBytes)

	var cleanedContent string
	switch {
	case strings.Contains(contentType, "text/html"):
		cleanedContent = cleanHTML(content)
	case strings.Contains(contentType, "application/json"),
		strings.Contains(contentType, "text/plain"),
		strings.Contains(contentType, "text/markdown"),
		strings.Contains(contentType, "text/csv"):
		cleanedContent = content
	default:
		// If we don't know the content type, we try to see if it's something we can read as text
		// This is a fallback.
		cleanedContent = content
	}

	return ToolResult{
		FullResult: strings.TrimSpace(cleanedContent),
		Summary:    fmt.Sprintf("Fetched content from %s (%s)", a.URL, contentType),
	}, nil
}

// JSONSchema returns the JSON schema for the tool's arguments.
func (t *WebFetchTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":        "web_fetch",
		"description": "Downloads content from a web URL and returns it as text. Supports HTML (cleaned), JSON, Markdown, TXT, and CSV.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{"type": "string"},
			},
			"required": []string{"url"},
		},
	}
}

var reTags = regexp.MustCompile("<[^>]*>")

func cleanHTML(input string) string {
	// Replace HTML tags with a space to avoid joining words together
	s := reTags.ReplaceAllString(input, " ")
	// Replace multiple spaces with a single space
	reSpaces := regexp.MustCompile(`\s+`)
	s = reSpaces.ReplaceAllString(s, " ")
	return s
}
