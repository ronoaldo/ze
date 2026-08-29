package tui

// terminal defines the interface for platform-specific terminal operations.
type terminal interface {
	getTerminalSize() (int, int)
	enableRawMode() (any, error)
	disableRawMode(original any) error
}

// newTerminal returns a platform-specific terminal implementation.
func newTerminal() terminal {
	return &platformTerminal{}
}
