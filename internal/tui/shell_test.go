package tui

import (
	"bytes"
	"errors"
	"testing"
	"bufio"
)

func TestRun_ShellBehavior_CookedMode(t *testing.T) {
	// Simulating: User presses ENTER (empty line), then types "hi" and ENTER.
	// The terminal (Cooked Mode) handles the empty line and only passes the command.
	input := []byte("\nhi\n")
	
	r := bytes.NewReader(input)
	w := &bytes.Buffer{}
	
	tui := &TUI{
		w:      w,
		r:      r,
		reader: bufio.NewReader(r),
	}

	errStop := errors.New("stop loop")
	handlerCalled := false
	handler := func(msg string) (string, error) {
		handlerCalled = true
		if msg != "hi" {
			t.Errorf("Expected 'hi', got: %q", msg)
		}
		return "", errStop
	}

	_ = tui.Run(handler)
	
	if !handlerCalled {
		t.Fatal("Handler was never called")
	}
}

func TestRun_ShellBehavior_StandardInput(t *testing.T) {
	input := []byte("olá\n")
	r := bytes.NewReader(input)
	w := &bytes.Buffer{}
	tui := &TUI{
		w:      w,
		r:      r,
		reader: bufio.NewReader(r),
	}

	errStop := errors.New("stop loop")
	handlerCalled := false
	handler := func(msg string) (string, error) {
		handlerCalled = true
		if msg != "olá" {
			t.Errorf("Expected 'olá', got: %q", msg)
		}
		return "", errStop
	}

	_ = tui.Run(handler)
	
	if !handlerCalled {
		t.Fatal("Handler was never called")
	}
}
