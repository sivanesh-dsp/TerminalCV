package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the whole screen: a rounded frame with a header bar, scrollable
// body, and a contextual footer of keyboard hints.
func (m Model) View() string {
	if m.quitting {
		return m.theme.dim.Render("\n  See you around — thanks for visiting.\n\n")
	}
	if m.mode == modeSplash || !m.ready {
		return m.splashView()
	}

	header := m.headerBar()
	footer := m.footerBar()
	rule := ruleLine(m.theme, m.innerW)

	body := lipgloss.NewStyle().
		MaxWidth(m.innerW).MaxHeight(m.bodyH).
		Render(m.bodyBlock())

	inner := strings.Join([]string{header, rule, body, rule, footer}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colBorder).
		Padding(0, 1).
		Render(inner)
}

func (m Model) splashView() string {
	t := m.theme
	block := lipgloss.JoinVertical(lipgloss.Center,
		t.headerName.Render(strings.ToUpper(m.res.Name)),
		t.dim.Render(m.res.Title),
		"",
		t.accent.Render("● ● ●"),
	)
	return lipgloss.Place(max(m.width, 1), max(m.height, 1),
		lipgloss.Center, lipgloss.Center, block)
}

// bodyBlock renders the mode-specific body clamped to exactly innerW × bodyH.
func (m Model) bodyBlock() string {
	styleBody := lipgloss.NewStyle().
		Width(m.innerW).Height(m.bodyH).
		MaxWidth(m.innerW).MaxHeight(m.bodyH)

	switch m.mode {
	case modeMenu:
		content := m.menuContent()
		colW := lipgloss.Width(strings.SplitN(content, "\n", 2)[0])
		pad := (m.innerW - colW) / 2
		if pad < 0 {
			pad = 0
		}
		return lipgloss.Place(m.innerW, m.bodyH, lipgloss.Left, lipgloss.Center, indentBlock(content, pad))
	case modeSection:
		return styleBody.Render(m.vp.View())
	case modeSearch:
		return styleBody.Render(m.searchContent())
	case modeHelp:
		return styleBody.Render(m.helpContent())
	}
	return styleBody.Render("")
}

func (m Model) menuContent() string {
	t := m.theme
	width := 0
	for _, s := range m.sections {
		if l := lipgloss.Width(s.name) + 2; l > width {
			width = l
		}
	}
	var lines []string
	for i, s := range m.sections {
		prefix := "  "
		style := t.menuItem
		if i == m.menuIdx {
			prefix = "▸ "
			style = t.selected
		}
		lines = append(lines, style.Render(padRight(prefix+s.name, width)))
	}
	return strings.Join(lines, "\n")
}

func (m Model) searchContent() string {
	t := m.theme
	var b strings.Builder
	b.WriteString(t.title.Render("SEARCH") + "\n")
	b.WriteString(ruleLine(t, min(m.innerW, 48)) + "\n\n")
	b.WriteString(m.search.View() + "\n\n")

	q := strings.TrimSpace(m.search.Value())
	if q == "" {
		b.WriteString(t.dim.Render("Search across every résumé section."))
		return b.String()
	}
	if len(m.results) == 0 {
		b.WriteString(t.dim.Render("No results for “" + q + "”."))
		return b.String()
	}
	b.WriteString(t.dim.Render(fmt.Sprintf("%d result(s)", len(m.results))) + "\n\n")
	maxShow := m.bodyH - 8
	if maxShow < 1 {
		maxShow = 1
	}
	for i, hit := range m.results {
		if i >= maxShow {
			b.WriteString(t.dim.Render(fmt.Sprintf("  … %d more", len(m.results)-maxShow)))
			break
		}
		marker := "  "
		nameStyle := t.menuItem
		if i == m.resIdx {
			marker = t.accent.Render("▸ ")
			nameStyle = t.selected
		}
		line := marker + t.accent2.Render(padRight(hit.Section, 15)) + nameStyle.Render(truncate(hit.Title, m.innerW-20))
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) helpContent() string {
	t := m.theme
	rows := [][2]string{
		{"↑ / ↓  k / j", "navigate menu · scroll · browse items"},
		{"← / →  Tab", "previous / next section"},
		{"Enter", "open selected menu item / search result"},
		{"Esc", "back to menu"},
		{"/", "open search"},
		{"?", "toggle this help"},
		{"q  Ctrl+C", "quit / disconnect"},
	}
	var b strings.Builder
	b.WriteString(t.title.Render("KEYBOARD SHORTCUTS") + "\n")
	b.WriteString(ruleLine(t, min(m.innerW, 48)) + "\n\n")
	for _, r := range rows {
		b.WriteString(t.key.Render(padRight(r[0], 16)) + t.keyDesc.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + t.dim.Render("Press Esc or ? to close."))
	return b.String()
}

func (m Model) headerBar() string {
	t := m.theme
	left := t.headerName.Render(strings.ToUpper(m.res.Name))
	right := t.headerTag.Render(m.headerRight())
	return joinEnds(left, right, m.innerW)
}

func (m Model) headerRight() string {
	switch m.mode {
	case modeSection:
		return m.sections[m.active].name
	case modeSearch:
		return "SEARCH"
	case modeHelp:
		return "HELP"
	default:
		return "PORTFOLIO"
	}
}

func (m Model) footerBar() string {
	t := m.theme
	hints := m.footerHints()
	left := renderHints(t, hints)
	right := ""
	if m.mode == modeSection {
		sec := m.sections[m.active]
		if sec.id == "experience" && len(m.res.Experience) > 0 {
			right = t.dim.Render(fmt.Sprintf("%02d / %02d", m.selExp+1, len(m.res.Experience)))
		} else if sec.id == "projects" && len(m.res.Projects) > 0 {
			right = t.dim.Render(fmt.Sprintf("%02d / %02d", m.selProj+1, len(m.res.Projects)))
		}
	}
	return joinEnds(left, right, m.innerW)
}

// footerHints returns [key, description] pairs for the current mode. On narrow
// terminals a shorter set is used.
func (m Model) footerHints() [][2]string {
	narrow := m.innerW < 64
	switch m.mode {
	case modeMenu:
		if narrow {
			return [][2]string{{"↑↓", "nav"}, {"↵", "open"}, {"/", "search"}, {"q", "quit"}}
		}
		return [][2]string{{"↑↓", "navigate"}, {"↵", "open"}, {"/", "search"}, {"?", "help"}, {"q", "quit"}}
	case modeSection:
		move := "scroll"
		if isPager(m.sections[m.active].id) {
			move = "browse"
		}
		if narrow {
			return [][2]string{{"↑↓", move}, {"←→", "section"}, {"esc", "menu"}, {"q", "quit"}}
		}
		return [][2]string{{"↑↓", move}, {"←→", "section"}, {"/", "search"}, {"esc", "menu"}, {"q", "quit"}}
	case modeSearch:
		return [][2]string{{"↑↓", "move"}, {"↵", "open"}, {"esc", "close"}}
	case modeHelp:
		return [][2]string{{"esc", "close"}}
	}
	return nil
}

func renderHints(t theme, pairs [][2]string) string {
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, t.key.Render(p[0])+" "+t.keyDesc.Render(p[1]))
	}
	return strings.Join(parts, t.dim.Render("   "))
}

// joinEnds places left and right on one line of exactly width columns.
func joinEnds(left, right string, width int) string {
	lw, rw := lipgloss.Width(left), lipgloss.Width(right)
	if lw+rw+1 > width {
		// Not enough room: keep the left side, clamp to width.
		return lipgloss.NewStyle().MaxWidth(width).Render(left)
	}
	gap := width - lw - rw
	return left + strings.Repeat(" ", gap) + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// indentBlock prepends n spaces to every line so a block can be positioned as a
// left-aligned unit (avoiding lipgloss per-line centering).
func indentBlock(s string, n int) string {
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}
