package prompt

import (
	"strings"
	"testing"
)

func TestGetGemma4SystemPrompt_NotEmpty(t *testing.T) {
	p := GetGemma4SystemPrompt()
	if p == "" {
		t.Fatal("GetGemma4SystemPrompt() returned empty string")
	}
}

func TestGetGemma4SystemPrompt_ContainsIdentity(t *testing.T) {
	p := GetGemma4SystemPrompt()
	if !strings.Contains(p, "Zé") {
		t.Error("prompt should mention 'Zé' as identity")
	}
}

func TestGetGemma4SystemPrompt_ContainsToolNames(t *testing.T) {
	p := GetGemma4SystemPrompt()
	tools := []string{"read_file", "write_file", "list_files", "go_doc"}
	for _, tool := range tools {
		if !strings.Contains(p, tool) {
			t.Errorf("prompt should mention tool '%s'", tool)
		}
	}
}

func TestGetGemma4SystemPrompt_ContainsErrorHandling(t *testing.T) {
	p := GetGemma4SystemPrompt()
	if !strings.Contains(p, "error") && !strings.Contains(p, "Error") {
		t.Error("prompt should mention error handling guidelines")
	}
}
