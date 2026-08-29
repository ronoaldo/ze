package tui

// InputBuffer manages the text content and cursor position for a single line of input.
type InputBuffer struct {
	content   []rune
	cursorPos int
}

// NewInputBuffer creates a new, empty InputBuffer.
func NewInputBuffer() *InputBuffer {
	return &InputBuffer{
		content:   nil,
		cursorPos: 0,
	}
}

// Insert adds a rune at the current cursor position.
func (b *InputBuffer) Insert(r rune) {
	if b.content == nil {
		b.content = []rune{r}
		b.cursorPos = 1
		return
	}

	newContent := make([]rune, len(b.content)+1)
	copy(newContent[:b.cursorPos], b.content[:b.cursorPos])
	newContent[b.cursorPos] = r
	copy(newContent[b.cursorPos+1:], b.content[b.cursorPos:])
	b.content = newContent
	b.cursorPos++
}

// Delete removes the character at the current cursor position.
func (b *InputBuffer) Delete() {
	if b.content == nil || len(b.content) == 0 {
		return
	}
	if b.cursorPos < len(b.content) {
		newContent := make([]rune, len(b.content)-1)
		copy(newContent[:b.cursorPos], b.content[:b.cursorPos])
		copy(newContent[b.cursorPos:], b.content[b.cursorPos+1:])
		b.content = newContent
	}
}

// Backspace removes the character before the current cursor position.
func (b *InputBuffer) Backspace() {
	if b.content == nil || len(b.content) == 0 {
		return
	}
	if b.cursorPos > 0 {
		newContent := make([]rune, len(b.content)-1)
		copy(newContent[:b.cursorPos-1], b.content[:b.cursorPos-1])
		copy(newContent[b.cursorPos-1:], b.content[b.cursorPos:])
		b.content = newContent
		b.cursorPos--
	}
}

// MoveLeft moves the cursor one position to the left.
func (b *InputBuffer) MoveLeft() {
	if b.cursorPos > 0 {
		b.cursorPos--
	}
}

// MoveRight moves the cursor one position to the right.
func (b *InputBuffer) MoveRight() {
	if b.cursorPos < len(b.content) {
		b.cursorPos++
	}
}

// MoveHome moves the cursor to the beginning of the buffer.
func (b *InputBuffer) MoveHome() {
	b.cursorPos = 0
}

// MoveEnd moves the cursor to the end of the buffer.
func (b *InputBuffer) MoveEnd() {
	b.cursorPos = len(b.content)
}

// String returns the content of the buffer as a string.
func (b *InputBuffer) String() string {
	if b.content == nil {
		return ""
	}
	return string(b.content)
}

// Len returns the number of characters in the buffer.
func (b *InputBuffer) Len() int {
	if b.content == nil {
		return 0
	}
	return len(b.content)
}

// CursorPosition returns the current cursor position.
func (b *InputBuffer) CursorPosition() int {
	return b.cursorPos
}

// Clear resets the buffer and cursor position.
func (b *InputBuffer) Clear() {
	b.content = nil
	b.cursorPos = 0
}

// SetContent sets the content of the buffer.
func (b *InputBuffer) SetContent(s string) {
	b.content = []rune(s)
	b.cursorPos = len(b.content)
}
