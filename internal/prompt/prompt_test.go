package prompt

import (
	"strings"
	"testing"
)

func TestGetGemma4SystemPrompt_NotEmpty(t *testing.T) {
	p := GetGemma4SystemPrompt("")
	if p == "" {
		t.Fatal("GetGemma4SystemPrompt() returned empty string")
	}
}

func TestGetGemma4SystemPrompt_ContainsIdentity(t *testing.T) {
	p := GetGemma4SystemPrompt("")
	if !strings.Contains(p, "Zé") {
		t.Error("prompt should mention 'Zé' as identity")
	}
}

func TestGetGemma4SystemPrompt_ContainsToolCalling(t *testing.T) {
	p := GetGemma4SystemPrompt("")
	if !strings.Contains(p, "Tool Calling") {
		t.Error("prompt should mention 'Tool Calling' section")
	}
}

func TestGetGemma4SystemPrompt_ContainsErrorHandling(t *testing.T) {
	p := GetGemma4SystemPrompt("")
	if !strings.Contains(p, "error") && !strings.Contains(p, "Error") {
		t.Error("prompt should mention error handling guidelines")
	}
}

func TestGetGemma4SystemPrompt_WithModelName(t *testing.T) {
	p := GetGemma4SystemPrompt("gemma-4-ft")
	if !strings.Contains(p, "gemma-4-ft") {
		t.Error("prompt should include model name when provided")
	}
	if !strings.Contains(p, "You are running on model: gemma-4-ft") {
		t.Error("prompt should have model info header")
	}
}

func TestGetGemma4SystemPrompt_EmptyModelName(t *testing.T) {
	p := GetGemma4SystemPrompt("")
	if strings.Contains(p, "You are running on model:") {
		t.Error("prompt should not include model info when name is empty")
	}
}
