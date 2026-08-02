// Package ansi provides terminal styling: truecolor text, bold/underline,
// OSC 8 hyperlinks and helpers to measure visible width (ignoring escapes).
// All styling is a no-op when color is disabled (dumb terminals / NO_COLOR).
package ansi

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	esc   = "\x1b"
	bel   = "\x07"
	Reset = esc + "[0m"
)

// Palette mirrors the website's "dark" theme accents.
var (
	accent  = [3]int{94, 242, 176}  // mint
	accent2 = [3]int{122, 162, 247} // blue
	dim     = [3]int{122, 133, 144} // gray
	warn    = [3]int{255, 204, 102} // amber
	errc    = [3]int{255, 107, 107} // red
	link    = [3]int{125, 207, 255} // cyan
)

// Style renders text with ANSI codes when Enabled, otherwise returns it as-is.
type Style struct {
	Enabled bool
}

func New(enabled bool) *Style { return &Style{Enabled: enabled} }

func (s *Style) fg(c [3]int, text string) string {
	if !s.Enabled || text == "" {
		return text
	}
	return fmt.Sprintf("%s[38;2;%d;%d;%dm%s%s", esc, c[0], c[1], c[2], text, Reset)
}

func (s *Style) Accent(t string) string  { return s.fg(accent, t) }
func (s *Style) Accent2(t string) string { return s.fg(accent2, t) }
func (s *Style) Dim(t string) string     { return s.fg(dim, t) }
func (s *Style) Warn(t string) string    { return s.fg(warn, t) }
func (s *Style) Err(t string) string     { return s.fg(errc, t) }
func (s *Style) Ok(t string) string      { return s.fg(accent, t) }
func (s *Style) Link(t string) string    { return s.fg(link, t) }

func (s *Style) Bold(t string) string {
	if !s.Enabled || t == "" {
		return t
	}
	return esc + "[1m" + t + Reset
}

func (s *Style) Underline(t string) string {
	if !s.Enabled || t == "" {
		return t
	}
	return esc + "[4m" + t + Reset
}

// Hyperlink wraps text in an OSC 8 escape. Terminals that don't support it
// simply render the text. Disabled when styling is off.
func (s *Style) Hyperlink(uri, text string) string {
	if !s.Enabled {
		return text
	}
	return esc + "]8;;" + uri + bel + text + esc + "]8;;" + bel
}

// Strip removes ANSI CSI/OSC escape sequences from a string.
func Strip(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) {
			switch s[i+1] {
			case '[': // CSI: ESC [ ... final byte in @-~
				j := i + 2
				for j < len(s) && (s[j] < '@' || s[j] > '~') {
					j++
				}
				i = j + 1
				continue
			case ']': // OSC: ESC ] ... terminated by BEL or ST (ESC \)
				j := i + 2
				for j < len(s) && s[j] != 0x07 {
					if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
						j++
						break
					}
					j++
				}
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// VisibleWidth returns the number of visible runes (escapes excluded).
func VisibleWidth(s string) int {
	return utf8.RuneCountInString(Strip(s))
}
