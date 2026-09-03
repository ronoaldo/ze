package tui

// Palette defines the ANSI colors used in the TUI.
type Palette struct {
	Reset     string
	Bold      string
	Dim       string
	Cyan      string
	Green     string
	Red       string
	Yellow    string
	Italic    string
	Underline string
}

// DefaultPalette returns the default ANSI color palette.
func DefaultPalette() Palette {
	return Palette{
		Reset:     "\x1b[0m",
		Bold:      "\x1b[1m",
		Dim:       "\x1b[2m",
		Cyan:      "\x1b[36m",
		Green:     "\x1b[32m",
		Red:       "\x1b[31m",
		Yellow:    "\x1b[33m",
		Italic:    "\x1b[3m",
		Underline: "\x1b[4m",
	}
}

// NoColorPalette returns a palette with no colors.
func NoColorPalette() Palette {
	return Palette{
		Reset:     "",
		Bold:      "",
		Dim:       "",
		Cyan:      "",
		Green:     "",
		Red:       "",
		Yellow:    "",
		Italic:    "",
		Underline: "",
	}
}
