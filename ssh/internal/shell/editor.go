package shell

import (
	"fmt"
	"io"
	"strings"

	"github.com/sivanesh/portfolio-ssh/internal/ansi"
)

// readLine reads and edits one line in raw mode, returning it on Enter.
// It implements history (↑/↓), cursor movement, the common Ctrl shortcuts and
// Tab completion. Returns io.EOF on Ctrl+D at an empty prompt.
func (sh *Shell) readLine(prompt string) (string, error) {
	promptWidth := ansi.VisibleWidth(prompt)
	var buf []rune
	pos := 0
	sh.histIdx = len(sh.history)
	sh.histDraft = ""

	sh.write(prompt)

	redraw := func() {
		var b strings.Builder
		b.WriteString("\r\x1b[K")
		b.WriteString(prompt)
		b.WriteString(string(buf))
		if back := len(buf) - pos; back > 0 {
			fmt.Fprintf(&b, "\x1b[%dD", back)
		}
		sh.write(b.String())
	}

	for {
		r, _, err := sh.in.ReadRune()
		if err != nil {
			return "", io.EOF
		}

		switch r {
		case '\r', '\n':
			if r == '\r' && sh.in.Buffered() > 0 {
				if b, _ := sh.in.Peek(1); len(b) == 1 && b[0] == '\n' {
					_, _, _ = sh.in.ReadRune()
				}
			}
			sh.write("\r\n")
			return string(buf), nil

		case 0x03: // Ctrl+C — abort current line
			sh.write("^C\r\n")
			buf, pos = nil, 0
			sh.histIdx = len(sh.history)
			sh.write(prompt)

		case 0x04: // Ctrl+D — EOF if empty, else delete-forward
			if len(buf) == 0 {
				return "", io.EOF
			}
			if pos < len(buf) {
				buf = append(buf[:pos], buf[pos+1:]...)
				redraw()
			}

		case 0x0c: // Ctrl+L — clear screen
			sh.write("\x1b[2J\x1b[3J\x1b[H")
			sh.write(prompt + string(buf))
			if back := len(buf) - pos; back > 0 {
				sh.write(fmt.Sprintf("\x1b[%dD", back))
			}

		case 0x01: // Ctrl+A — home
			pos = 0
			redraw()
		case 0x05: // Ctrl+E — end
			pos = len(buf)
			redraw()
		case 0x15: // Ctrl+U — kill to start
			buf = append([]rune{}, buf[pos:]...)
			pos = 0
			redraw()
		case 0x0b: // Ctrl+K — kill to end
			buf = buf[:pos]
			redraw()
		case 0x17: // Ctrl+W — delete previous word
			start := pos
			for start > 0 && buf[start-1] == ' ' {
				start--
			}
			for start > 0 && buf[start-1] != ' ' {
				start--
			}
			buf = append(buf[:start], buf[pos:]...)
			pos = start
			redraw()

		case 0x08, 0x7f: // Backspace
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
				redraw()
			}

		case '\t': // Tab completion
			newLine, cands := sh.autocomplete(string(buf))
			if newLine != string(buf) {
				buf = []rune(newLine)
				pos = len(buf)
			}
			if len(cands) > 1 {
				sh.write("\r\n")
				sh.writeln(sh.formatCandidates(cands, promptWidth))
				sh.write(prompt + string(buf))
				if back := len(buf) - pos; back > 0 {
					sh.write(fmt.Sprintf("\x1b[%dD", back))
				}
			} else {
				redraw()
			}

		case 0x1b: // escape sequence (arrows / home / end / delete)
			switch sh.readEscape() {
			case "UP":
				sh.historyPrev(&buf, &pos)
				redraw()
			case "DOWN":
				sh.historyNext(&buf, &pos)
				redraw()
			case "LEFT":
				if pos > 0 {
					pos--
					redraw()
				}
			case "RIGHT":
				if pos < len(buf) {
					pos++
					redraw()
				}
			case "WORD_LEFT":
				pos = wordLeft(buf, pos)
				redraw()
			case "WORD_RIGHT":
				pos = wordRight(buf, pos)
				redraw()
			case "HOME":
				pos = 0
				redraw()
			case "END":
				pos = len(buf)
				redraw()
			case "DELETE":
				if pos < len(buf) {
					buf = append(buf[:pos], buf[pos+1:]...)
					redraw()
				}
			}

		default:
			if r >= 0x20 && r != 0x7f {
				buf = append(buf, 0)
				copy(buf[pos+1:], buf[pos:])
				buf[pos] = r
				pos++
				redraw()
			}
		}
	}
}

// readEscape consumes a CSI/SS3 sequence (ESC already read) and names the key.
func (sh *Shell) readEscape() string {
	if sh.in.Buffered() == 0 {
		return "" // lone ESC
	}
	b1, _, err := sh.in.ReadRune()
	if err != nil || (b1 != '[' && b1 != 'O') {
		return ""
	}
	var params []rune
	var final rune
	for {
		r, _, err := sh.in.ReadRune()
		if err != nil {
			return ""
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '~' {
			final = r
			break
		}
		params = append(params, r)
	}
	p := string(params)
	mod := strings.Contains(p, ";5") || strings.Contains(p, ";2")
	switch final {
	case 'A':
		return "UP"
	case 'B':
		return "DOWN"
	case 'C':
		if mod {
			return "WORD_RIGHT"
		}
		return "RIGHT"
	case 'D':
		if mod {
			return "WORD_LEFT"
		}
		return "LEFT"
	case 'H':
		return "HOME"
	case 'F':
		return "END"
	case '~':
		switch p {
		case "1", "7":
			return "HOME"
		case "4", "8":
			return "END"
		case "3":
			return "DELETE"
		}
	}
	return ""
}

func (sh *Shell) historyPrev(buf *[]rune, pos *int) {
	if len(sh.history) == 0 {
		return
	}
	if sh.histIdx == len(sh.history) {
		sh.histDraft = string(*buf)
	}
	if sh.histIdx > 0 {
		sh.histIdx--
	}
	*buf = []rune(sh.history[sh.histIdx])
	*pos = len(*buf)
}

func (sh *Shell) historyNext(buf *[]rune, pos *int) {
	if sh.histIdx >= len(sh.history) {
		return
	}
	sh.histIdx++
	if sh.histIdx == len(sh.history) {
		*buf = []rune(sh.histDraft)
	} else {
		*buf = []rune(sh.history[sh.histIdx])
	}
	*pos = len(*buf)
}

func wordLeft(buf []rune, pos int) int {
	for pos > 0 && buf[pos-1] == ' ' {
		pos--
	}
	for pos > 0 && buf[pos-1] != ' ' {
		pos--
	}
	return pos
}

func wordRight(buf []rune, pos int) int {
	n := len(buf)
	for pos < n && buf[pos] == ' ' {
		pos++
	}
	for pos < n && buf[pos] != ' ' {
		pos++
	}
	return pos
}

// autocomplete returns the completed line and, when ambiguous, the candidates.
func (sh *Shell) autocomplete(line string) (string, []string) {
	trailing := strings.HasSuffix(line, " ")
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return line, nil
	}

	complete := func(prefix string, options []string, base string) (string, []string) {
		matches := prefixMatches(options, prefix)
		switch len(matches) {
		case 0:
			return line, nil
		case 1:
			return base + matches[0] + " ", nil
		default:
			if cp := commonPrefix(matches); len(cp) > len(prefix) {
				return base + cp, matches
			}
			return line, matches
		}
	}

	// Command-name completion (still typing the first word).
	if len(fields) == 1 && !trailing {
		return complete(fields[0], sh.reg.Names(), "")
	}
	// `cat <file>` filename completion.
	if fields[0] == "cat" {
		prefix := ""
		if !trailing {
			prefix = fields[len(fields)-1]
		}
		return complete(prefix, vfiles, "cat ")
	}
	return line, nil
}

func (sh *Shell) formatCandidates(cands []string, _ int) string {
	return sh.style.Dim(strings.Join(cands, "  "))
}

func prefixMatches(options []string, prefix string) []string {
	var out []string
	for _, o := range options {
		if strings.HasPrefix(o, prefix) {
			out = append(out, o)
		}
	}
	return out
}

func commonPrefix(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	p := ss[0]
	for _, s := range ss[1:] {
		for !strings.HasPrefix(s, p) {
			p = p[:len(p)-1]
			if p == "" {
				return ""
			}
		}
	}
	return p
}
