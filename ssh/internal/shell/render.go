package shell

import (
	"strings"
	"unicode/utf8"

	"github.com/sivanesh/portfolio-ssh/internal/ansi"
)

// contentWidth caps line width for readability on very wide terminals while
// staying responsive on narrow ones.
func contentWidth(w int) int {
	if w > 100 {
		return 100
	}
	if w < 20 {
		return 20
	}
	return w
}

// heading renders an uppercase section title with an accent underline.
func heading(st *ansi.Style, title string) string {
	t := strings.ToUpper(title)
	return st.Bold(st.Accent(t)) + "\n" +
		st.Dim(strings.Repeat("─", utf8.RuneCountInString(t))) + "\n"
}

// wrapText word-wraps text to width, preserving long unbreakable words.
func wrapText(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	cur := ""
	for _, w := range words {
		switch {
		case cur == "":
			cur = w
		case utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) <= width:
			cur += " " + w
		default:
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// markerLines renders a pre-styled marker + wrapped, hanging-indented text.
// markerWidth is the marker's visible width (e.g. 2 for "▸ ").
func markerLines(marker string, markerWidth, width int, text string) string {
	lines := wrapText(text, width-markerWidth)
	var b strings.Builder
	for i, l := range lines {
		if i == 0 {
			b.WriteString(marker + l + "\n")
		} else {
			b.WriteString(strings.Repeat(" ", markerWidth) + l + "\n")
		}
	}
	return b.String()
}

// padRight pads s (by visible width) to n columns.
func padRight(s string, n int) string {
	if vw := ansi.VisibleWidth(s); vw < n {
		return s + strings.Repeat(" ", n-vw)
	}
	return s
}
