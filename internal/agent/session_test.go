package agent

import (
	"os"
	"testing"

	"github.com/ronoaldo/ze/internal/llm"
)

func TestSessionManager(t *testing.T) {
	// Create a temporary directory for testing session storage
	tmpDir, err := os.MkdirTemp("", "ze_sessions_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sm := NewSessionManagerWithDir(tmpDir)

	// Test Session ID Generation
	id1, err := sm.GenerateSessionID()
	if err != nil {
		t.Fatalf("Failed to generate session ID: %v", err)
	}
	if id1 == "" {
		t.Fatal("Generated session ID is empty")
	}

	id2, err := sm.GenerateSessionID()
	if err != nil {
		t.Fatalf("Failed to generate session ID: %v", err)
	}
	if id1 == id2 {
		t.Error("Generated two identical session IDs")
	}

	// Test Saving and Loading Session
	history := []llm.ChatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	err = sm.SaveSession(id1, history)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	loadedHistory, err := sm.LoadSession(id1)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if len(loadedHistory) != len(history) {
		t.Errorf("Expected history length %d, got %d", len(history), len(loadedHistory))
	}

	for i := range history {
		if loadedHistory[i].Role != history[i].Role || loadedHistory[i].Content != history[i].Content {
			t.Errorf("Mismatch at index %d: expected %v, got %v", i, history[i], loadedHistory[i])
		}
	}

	// Test Loading non-existent session
	nonExistentID := "non-existent-id"
	loadedHistory, err = sm.LoadSession(nonExistentID)
	if err != nil {
		t.Errorf("Expected no error when loading non-existent session, got %v", err)
	}
	if loadedHistory != nil {
		t.Errorf("Expected nil history for non-existent session, got %v", loadedHistory)
	}
}
