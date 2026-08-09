package tui

import (
	"encoding/base64"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// osc52 returns the OSC 52 clipboard-set sequence for s. Supported by iTerm2,
// Ghostty, kitty, WezTerm, tmux, Windows Terminal and others; ignored elsewhere.
func osc52(s string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(s)) + "\x07"
}

// osc52For wraps the OSC 52 sequence in a tmux/screen DCS passthrough when the
// client's TERM indicates a multiplexer, so the clipboard write reaches the
// outer terminal instead of being swallowed.
func osc52For(term, s string) string {
	seq := osc52(s)
	lt := strings.ToLower(term)
	if strings.HasPrefix(lt, "tmux") || strings.HasPrefix(lt, "screen") {
		return "\x1bPtmux;\x1b" + seq + "\x1b\\"
	}
	return seq
}

// markFrag is a run of text with a single style (bold or not).
type markFrag struct {
	text string
	bold bool
}

// markWord is a whitespace-delimited word made of one or more fragments, so a
// bold span that ends mid-word (e.g. "Windows**,") stays glued to its
// punctuation instead of being split by a space.
type markWord struct {
	frags []markFrag
	width int
}

// parseMarkup converts a single line into words, stripping `**` markers and
// tagging the bold runs. Newlines must be handled by the caller.
func parseMarkup(text string) []markWord {
	var words []markWord
	var cur markWord
	var frag strings.Builder
	var fragBold, haveFrag bool

	pushFrag := func() {
		if haveFrag {
			s := frag.String()
			cur.frags = append(cur.frags, markFrag{text: s, bold: fragBold})
			cur.width += lipgloss.Width(s)
			frag.Reset()
			haveFrag = false
		}
	}
	pushWord := func() {
		pushFrag()
		if len(cur.frags) > 0 {
			words = append(words, cur)
			cur = markWord{}
		}
	}

	bold := false
	rs := []rune(text)
	for i := 0; i < len(rs); i++ {
		if rs[i] == '*' && i+1 < len(rs) && rs[i+1] == '*' {
			bold = !bold
			i++
			continue
		}
		if rs[i] == ' ' || rs[i] == '\t' {
			pushWord()
			continue
		}
		if haveFrag && bold != fragBold {
			pushFrag()
		}
		if !haveFrag {
			fragBold = bold
			haveFrag = true
		}
		frag.WriteRune(rs[i])
	}
	pushWord()
	return words
}

// markupWrap word-wraps text to width and returns a fully styled string, with
// `**…**` spans rendered in the accent (highlight) style. Wrapping width is
// computed from the visible words only (markers are stripped).
func markupWrap(t theme, text string, width int) string {
	if width < 8 {
		width = 8
	}
	var out strings.Builder
	for pi, para := range strings.Split(text, "\n") {
		if pi > 0 {
			out.WriteByte('\n')
		}
		lineLen := 0
		first := true
		for _, wd := range parseMarkup(para) {
			if !first && lineLen+1+wd.width > width {
				out.WriteByte('\n')
				lineLen = 0
				first = true
			}
			if !first {
				out.WriteByte(' ')
				lineLen++
			}
			for _, fr := range wd.frags {
				if fr.bold {
					out.WriteString(t.mark.Render(fr.text))
				} else {
					out.WriteString(t.base.Render(fr.text))
				}
			}
			lineLen += wd.width
			first = false
		}
	}
	return out.String()
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
