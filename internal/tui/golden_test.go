package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderMarkdown_GoldenFile(t *testing.T) {
	inputPath := filepath.Join("..", "..", "testdata", "golden", "input.md")
	outputPath := filepath.Join("..", "..", "testdata", "golden", "markdown_output.txt")

	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read input file: %v", err)
	}

	palette := DefaultPalette()
	style := Style{
		Bold:      palette.Bold,
		Italic:    palette.Italic,
		Underline: palette.Underline,
		Reset:     palette.Reset,
	}

	got := RenderMarkdown(string(input), style)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		err := os.WriteFile(outputPath, []byte(got), 0644)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", outputPath)
		return
	}

	expected, err := os.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file does not exist. Run with UPDATE_GOLDEN=1 to create it: %s", outputPath)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	if got != string(expected) {
		t.Errorf("RenderMarkdown() output mismatch.\nGOT:\n%q\nWANT:\n%q", got, string(expected))
	}
}
