package tui

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseEscapeSequence(t *testing.T) {
	tests := []struct {
		name    string
		input   string // Should NOT include the initial \x1b if we pass it after \x1b
		wantKey Key
		wantErr bool
	}{
		{"Up", "[A", KeyUp, false},
		{"Down", "[B", KeyDown, false},
		{"Right", "[C", KeyRight, false},
		{"Left", "[D", KeyLeft, false},
		{"Home (H)", "[H", KeyHome, false},
		{"Home (1~)", "[1~", KeyHome, false},
		{"Home (F)", "[F", KeyHome, false},
		{"End (G)", "[G", KeyEnd, false},
		{"End (4~)", "[4~", KeyEnd, false},
		{"End (L)", "[L", KeyEnd, false},
		{"Delete", "[3~", KeyDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			got, err := ParseEscapeSequence(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEscapeSequence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantKey {
				t.Errorf("ParseEscapeSequence() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}

func TestReadKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey Key
		wantErr bool
	}{
		{"Enter", "\r", KeyEnter, false},
		{"Backspace", "\x7f", KeyBackspace, false},
		{"Up", "\x1b[A", KeyUp, false},
		{"Left", "\x1b[D", KeyLeft, false},
		{"Home", "\x1b[H", KeyHome, false},
		{"Delete", "\x1b[3~", KeyDelete, false},
		{"Char 'a'", "a", Key('a'), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			got, err := ReadKey(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantKey {
				t.Errorf("ReadKey() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}
