//go:build windows

package tui

type platformTerminal struct{}

func (p *platformTerminal) getTerminalSize() (int, int) {
	return 80, 24
}

func (p *platformTerminal) enableRawMode() (any, error) {
	return nil, nil
}

func (p *platformTerminal) disableRawMode(original any) error {
	return nil
}
