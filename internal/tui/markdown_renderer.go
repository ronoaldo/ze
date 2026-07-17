package tui

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Style defines the ANSI sequences for different Markdown elements.
type Style struct {
	Bold      string
	Italic    string
	Underline string
	Reset     string
}

// RenderMarkdown takes a markdown string and returns a version with ANSI escapes for terminal.
func RenderMarkdown(input string, style Style) string {
	if input == "" {
		return ""
	}

	lines := strings.Split(input, "\n")
	var renderedLines []string
	var inTable bool
	var tableLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Detect table block
		if strings.HasPrefix(trimmed, "|") {
			inTable = true
			tableLines = append(tableLines, line)
			continue
		}

		// If we were in a table and the current line is NOT part of a table
		if inTable {
			renderedLines = append(renderedLines, renderTable(tableLines, style)...)
			tableLines = nil
			inTable = false
		}

		renderedLines = append(renderedLines, renderLine(line, style))
	}

	// Last table block
	if inTable {
		renderedLines = append(renderedLines, renderTable(tableLines, style)...)
	}

	return strings.Join(renderedLines, "\n")
}

func renderLine(line string, style Style) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return line
	}

	// 1. Headers
	if strings.HasPrefix(trimmed, "#") {
		content := strings.TrimLeft(trimmed, "# ")
		return style.Bold + style.Underline + content + style.Reset
	}

	// 2. Inline Styles
	// Bold: **text**
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	line = reBold.ReplaceAllString(line, style.Bold+"$1"+style.Reset)

	// Italic: *text*
	reItalic := regexp.MustCompile(`\*(.*?)\*`)
	line = reItalic.ReplaceAllString(line, style.Italic+"$1"+style.Reset)

	// Underline: __text__
	reUnderline := regexp.MustCompile(`__(.*?)__`)
	line = reUnderline.ReplaceAllString(line, style.Underline+"$1"+style.Reset)

	return line
}

func renderTable(lines []string, style Style) []string {
	if len(lines) == 0 {
		return nil
	}

	var rows [][]string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		
		content := trimmed[1:]
		if strings.HasSuffix(content, "|") {
			content = content[:len(content)-1]
		}
		
		parts := strings.Split(content, "|")
		row := []string{}
		for _, p := range parts {
			row = append(row, strings.TrimSpace(p))
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	numCols := 0
	for _, row := range rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}

	colWidths := make([]int, numCols)
	for _, row := range rows {
		for i, cell := range row {
			if i < numCols && utf8.RuneCountInString(cell) > colWidths[i] {
				colWidths[i] = utf8.RuneCountInString(cell)
			}
		}
	}

	var rendered []string
	for _, row := range rows {
		var sb strings.Builder
		sb.WriteString("|")
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			
			count := utf8.RuneCountInString(cell)
			padding := colWidths[i] - count
			if padding < 0 {
				padding = 0
			}
			
			sb.WriteString(" ")
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", padding))
			sb.WriteString(" |")
		}
		rendered = append(rendered, sb.String())
	}
	return rendered
}
