package agent

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ronoaldo/ze/internal/llm"
)

// SessionManager manages the persistence of chat history.
type SessionManager struct {
	SessionDir string
}

// NewSessionManager creates a new SessionManager with the default directory.
func NewSessionManager() (*SessionManager, error) {
	baseDir := os.Getenv("ZE_HOME")
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get user home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".config", "ze")
	}

	sessionDir := filepath.Join(baseDir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create session directory: %w", err)
	}

	return &SessionManager{SessionDir: sessionDir}, nil
}

// NewSessionManagerWithDir creates a new SessionManager with a custom directory.
func NewSessionManagerWithDir(sessionDir string) *SessionManager {
	return &SessionManager{SessionDir: sessionDir}
}

// GenerateSessionID generates a new unique session ID.
func (s *SessionManager) GenerateSessionID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func (s *SessionManager) getSessionPath(sessionID string) string {
	return filepath.Join(s.SessionDir, sessionID+".json")
}

// SaveSession saves the chat history to a JSON file.
func (s *SessionManager) SaveSession(sessionID string, history []llm.ChatMessage) error {
	path := s.getSessionPath(sessionID)
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	return nil
}

// LoadSession loads the chat history from a JSON file.
func (s *SessionManager) LoadSession(sessionID string) ([]llm.ChatMessage, error) {
	path := s.getSessionPath(sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session found
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var history []llm.ChatMessage
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}
