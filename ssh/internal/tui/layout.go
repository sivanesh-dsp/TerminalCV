package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// osc8 wraps label in an OSC 8 terminal hyperlink. Terminals that don't support
// it simply render the label, so this is always safe.
func osc8(url, label string) string {
	return "\x1b]8;;" + url + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}

// ruleLine returns a horizontal rule of the given width using a light box char.
func ruleLine(t theme, width int) string {
	if width < 1 {
		width = 1
	}
	return t.rule.Render(strings.Repeat("─", width))
}

// wrap hard-wraps text to width, preserving word boundaries.
func wrap(text string, width int) string {
	if width < 8 {
		width = 8
	}
	var out strings.Builder
	for i, para := range strings.Split(text, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(wrapParagraph(para, width))
	}
	return out.String()
}

func wrapParagraph(para string, width int) string {
	words := strings.Fields(para)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineLen := 0
	for i, w := range words {
		wl := lipgloss.Width(w)
		if i > 0 {
			if lineLen+1+wl > width {
				b.WriteByte('\n')
				lineLen = 0
			} else {
				b.WriteByte(' ')
				lineLen++
			}
		}
		b.WriteString(w)
		lineLen += wl
	}
	return b.String()
}

// padRight pads s (by display width) to at least n columns.
func padRight(s string, n int) string {
	if d := lipgloss.Width(s); d < n {
		return s + strings.Repeat(" ", n-d)
	}
	return s
}

// truncate shortens s to width display columns, adding an ellipsis if cut.
func truncate(s string, width int) string {
	if width < 1 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > width {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
