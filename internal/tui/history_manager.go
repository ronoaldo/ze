package tui

// HistoryManager manages the command history for the TUI.
type HistoryManager struct {
	history []string
	index   int // current index in history. -1 means "new line"
	maxSize int
}

// NewHistoryManager creates a new HistoryManager with a maximum size.
func NewHistoryManager(maxSize int) *HistoryManager {
	return &HistoryManager{
		history: make([]string, 0),
		index:   -1,
		maxSize: maxSize,
	}
}

// Add adds a new command to the history.
func (h *HistoryManager) Add(line string) {
	if line == "" {
		return
	}
	// Don't add duplicate consecutive lines
	if len(h.history) > 0 && h.history[len(h.history)-1] == line {
		h.index = -1
		return
	}
	h.history = append(h.history, line)
	if len(h.history) > h.maxSize {
		h.history = h.history[1:]
	}
	h.index = -1
}

// Prev moves to the previous command in history (older).
func (h *HistoryManager) Prev() (string, bool) {
	if h.index == -1 {
		if len(h.history) > 0 {
			h.index = len(h.history) - 1
			return h.history[h.index], true
		}
		return "", false
	}

	if h.index <= 0 {
		return "", false
	}

	h.index--
	return h.history[h.index], true
}

// Next moves to the next command in history (newer).
func (h *HistoryManager) Next() (string, bool) {
	if h.index == -1 {
		return "", false
	}

	if h.index >= len(h.history)-1 {
		h.index = -1
		return "", false
	}

	h.index++
	return h.history[h.index], true
}

// Clear clears the history.
func (h *HistoryManager) Clear() {
	h.history = make([]string, 0)
	h.index = -1
}

// GetHistory returns the current history slice.
func (h *HistoryManager) GetHistory() []string {
	return h.history
}
