package tui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderMarkdown(t *testing.T) {
	style := Style{
		Bold:      "\033[1m",
		Italic:    "\033[3m",
		Underline: "\033[4m",
		Reset:     "\033[0m",
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Bold",
			input:    "**bold**",
			expected: []string{"\033[1mbold\033[0m"},
		},
		{
			name:     "Italic",
			input:    "*italic*",
			expected: []string{"\033[3mitalic\033[0m"},
		},
		{
			name:     "Underline",
			input:    "__underline__",
			expected: []string{"\033[4munderline\033[0m"},
		},
		{
			name:     "Header",
			input:    "# Header",
			expected: []string{"\033[1m\033[4mHeader\033[0m"},
		},
		{
			name:  "Table",
			input: "| col1 | col2 |\n| val1 | val2 |",
			expected: []string{
				"| col1 | col2 |",
				"| val1 | val2 |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderMarkdown(tt.input, style)
			gotLines := strings.Split(got, "\n")

			if len(gotLines) != len(tt.expected) {
				t.Fatalf("RenderMarkdown() got %d lines, want %d. \nGot: %q\nWant: %q", len(gotLines), len(tt.expected), gotLines, tt.expected)
			}

			for i := range gotLines {
				if gotLines[i] != tt.expected[i] {
					t.Errorf("Line %d: got %q, want %q", i, gotLines[i], tt.expected[i])
				}
			}
		})
	}
}

func TestTableAlignment(t *testing.T) {
	style := Style{
		Bold:      "\033[1m",
		Italic:    "\033[3m",
		Underline: "\033[4m",
		Reset:     "\033[0m",
	}
	input := "| Estado | População |\n| Amapá | 1.3M |\n| São Paulo | 12M |"
	rendered := RenderMarkdown(input, style)
	lines := strings.Split(rendered, "\n")

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// Check if all rows have the same length (using RuneCount for visual alignment)
	firstLen := utf8.RuneCountInString(lines[0])
	for i, line := range lines {
		currentLen := utf8.RuneCountInString(line)
		if currentLen != firstLen {
			t.Errorf("Line %d has %d runes, want %d. Line: %q", i, currentLen, firstLen, line)
		}
	}
}
