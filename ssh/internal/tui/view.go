package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View centers a floating content block on an otherwise black screen — no
// full-screen frame (the terminal.shop aesthetic).
func (m Model) View() string {
	if m.quitting {
		return m.theme.dim.Render("\n  thanks for visiting — bye.\n\n")
	}
	if m.mode == modeSplash || !m.ready {
		return m.splashView()
	}

	var block string
	switch m.mode {
	case modeSearch:
		block = m.searchView()
	case modeHelp:
		block = m.helpView()
	default:
		block = m.browseView()
	}
	return lipgloss.Place(max(m.width, 1), max(m.height, 1),
		lipgloss.Center, lipgloss.Center, block)
}

// splashView shows a single minimal word, like terminal.shop's "terminal".
func (m Model) splashView() string {
	block := m.theme.brand.Render(m.res.Name)
	return lipgloss.Place(max(m.width, 1), max(m.height, 1),
		lipgloss.Center, lipgloss.Center, block)
}

// browseView assembles masthead · tab bar · two-column body · footer.
func (m Model) browseView() string {
	t := m.theme
	masthead := joinEnds(
		t.brand.Render(m.res.Name),
		t.dim.Render(m.res.Title),
		m.contentW,
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.leftPane(),
		spacerCol(3, m.bodyH),
		m.rightPane(),
	)

	footer := m.footer()

	return lipgloss.JoinVertical(lipgloss.Left,
		masthead,
		"",
		m.tabBar(),
		"",
		body,
		"",
		footer,
	)
}

func (m Model) tabBar() string {
	t := m.theme
	// Measure the bordered version; fall back to a compact single line if it
	// would overflow the content width (narrow terminals).
	var cells []string
	for i, tb := range m.tabs {
		inner := m.tabInner(tb, i == m.tab)
		if i == m.tab {
			cells = append(cells, t.tabActive.Render(inner))
		} else {
			cells = append(cells, t.tabIdle.Render(inner))
		}
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	if lipgloss.Width(bar) <= m.contentW {
		return centerBlock(bar, m.contentW)
	}
	return m.compactTabBar()
}

func (m Model) tabInner(tb tabDef, active bool) string {
	t := m.theme
	key := t.key.Render(string(tb.key))
	if active {
		return key + " " + t.brand.Render(tb.label)
	}
	return key + " " + t.dim.Render(tb.label)
}

func (m Model) compactTabBar() string {
	t := m.theme
	var parts []string
	for i, tb := range m.tabs {
		lbl := tb.label
		if i == m.tab {
			parts = append(parts, t.key.Render(string(tb.key))+" "+t.brand.Render(lbl))
		} else {
			parts = append(parts, t.key.Render(string(tb.key))+" "+t.dim.Render(lbl))
		}
	}
	line := strings.Join(parts, t.dim.Render("  "))
	return centerBlock(truncateANSI(line, m.contentW), m.contentW)
}

// leftPane renders the grouped item list with a solid accent selection bar.
func (m Model) leftPane() string {
	t := m.theme
	items := m.tabs[m.tab].items
	var lines []string
	lastGroup := ""
	for i, it := range items {
		if it.group != lastGroup {
			if it.group != "" {
				if len(lines) > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, t.group.Render(truncate("~ "+it.group+" ~", m.leftW)))
			}
			lastGroup = it.group
		}
		label := truncate(it.label, m.leftW-1)
		if i == m.sel {
			lines = append(lines, t.selBar.Render(padRight(" "+label, m.leftW)))
		} else {
			lines = append(lines, t.item.Render(" "+label))
		}
	}
	return lipgloss.NewStyle().Width(m.leftW).Height(m.bodyH).MaxHeight(m.bodyH).
		Render(strings.Join(lines, "\n"))
}

// rightPane renders the live detail viewport, with a focus indicator.
func (m Model) rightPane() string {
	t := m.theme
	content := m.vp.View()
	block := lipgloss.NewStyle().Width(m.rightW).Height(m.bodyH).MaxHeight(m.bodyH).Render(content)
	// Scroll affordance when the detail overflows.
	if m.vp.TotalLineCount() > m.bodyH {
		hint := "▾ more"
		if m.detailFocus {
			hint = fmt.Sprintf("%d%%", int(m.vp.ScrollPercent()*100))
		}
		block = overlayBottomRight(block, t.dim.Render(hint), m.rightW)
	}
	return block
}

func (m Model) footer() string {
	t := m.theme
	rule := t.rule.Render(strings.Repeat("─", m.contentW))
	var pairs [][2]string
	if m.detailFocus {
		pairs = [][2]string{{"↑/↓", "scroll"}, {"esc", "back"}, {"←/→", "tab"}, {"/", "search"}, {"q", "quit"}}
	} else {
		pairs = [][2]string{{"↑/↓", "browse"}, {"←/→", "tab"}}
		if it := m.curItem(); it != nil && it.copy != "" {
			pairs = append(pairs, [2]string{"y", "copy"})
		} else {
			pairs = append(pairs, [2]string{"↵", "open"})
		}
		pairs = append(pairs, [2]string{"/", "search"}, [2]string{"?", "help"}, [2]string{"q", "quit"})
	}
	hints := renderHints(t, pairs)
	return rule + "\n" + centerBlock(truncateANSI(hints, m.contentW), m.contentW)
}

func (m Model) searchView() string {
	t := m.theme
	w := m.contentW
	var b strings.Builder
	b.WriteString(t.brand.Render("search") + "\n")
	b.WriteString(t.rule.Render(strings.Repeat("─", min(w, 48))) + "\n\n")
	b.WriteString(m.search.View() + "\n\n")

	q := strings.TrimSpace(m.search.Value())
	switch {
	case q == "":
		b.WriteString(t.dim.Render("Search across every résumé section."))
	case len(m.results) == 0:
		b.WriteString(t.dim.Render("No results for “" + q + "”."))
	default:
		b.WriteString(t.dim.Render(fmt.Sprintf("%d result(s)", len(m.results))) + "\n\n")
		maxShow := m.bodyH
		for i, hit := range m.results {
			if i >= maxShow {
				b.WriteString(t.dim.Render(fmt.Sprintf("  … %d more", len(m.results)-maxShow)))
				break
			}
			marker := "  "
			nameStyle := t.item
			if i == m.resIdx {
				marker = t.accent.Render("▸ ")
				nameStyle = t.brand
			}
			b.WriteString(marker + t.accent2.Render(padRight(strings.ToLower(hit.Section), 15)) +
				nameStyle.Render(truncate(hit.Title, w-20)) + "\n")
		}
	}
	b.WriteString("\n\n" + t.rule.Render(strings.Repeat("─", w)) + "\n")
	b.WriteString(centerBlock(renderHints(t, [][2]string{{"↑/↓", "move"}, {"↵", "open"}, {"esc", "close"}}), w))
	return lipgloss.NewStyle().Width(w).Render(b.String())
}

func (m Model) helpView() string {
	t := m.theme
	w := m.contentW
	rows := [][2]string{
		{"↑ / ↓  k / j", "browse the list · scroll the detail pane"},
		{"← / →  Tab", "switch tab (category)"},
		{"a e p s r c", "jump straight to a tab"},
		{"Enter", "focus the detail pane"},
		{"Esc", "leave the detail pane"},
		{"/", "search every section"},
		{"?", "toggle this help"},
		{"q  Ctrl+C", "quit / disconnect"},
	}
	var b strings.Builder
	b.WriteString(t.brand.Render("keyboard") + "\n")
	b.WriteString(t.rule.Render(strings.Repeat("─", min(w, 48))) + "\n\n")
	for _, r := range rows {
		b.WriteString(t.key.Render(padRight(r[0], 16)) + t.keyDesc.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + t.dim.Render("Press Esc or ? to close."))
	return lipgloss.NewStyle().Width(w).Render(b.String())
}

/* -------------------------------- helpers --------------------------------- */

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
		return truncateANSI(left, width)
	}
	return left + strings.Repeat(" ", width-lw-rw) + right
}

// centerBlock horizontally centers a (possibly multi-line) block within width,
// indenting every line by the same amount so borders stay aligned.
func centerBlock(s string, width int) string {
	lines := strings.Split(s, "\n")
	max := 0
	for _, ln := range lines {
		if w := lipgloss.Width(ln); w > max {
			max = w
		}
	}
	if max >= width {
		return s
	}
	pad := strings.Repeat(" ", (width-max)/2)
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

// spacerCol returns a blank column of the given width and height.
func spacerCol(width, height int) string {
	line := strings.Repeat(" ", width)
	rows := make([]string, height)
	for i := range rows {
		rows[i] = line
	}
	return strings.Join(rows, "\n")
}

// overlayBottomRight writes label onto the last line of block, right-aligned.
func overlayBottomRight(block, label string, width int) string {
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return block
	}
	last := len(lines) - 1
	lw := lipgloss.Width(label)
	if lw < width {
		lines[last] = strings.Repeat(" ", width-lw) + label
	}
	return strings.Join(lines, "\n")
}

func truncateANSI(s string, width int) string {
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
