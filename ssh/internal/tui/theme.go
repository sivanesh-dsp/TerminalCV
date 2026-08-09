package tui

import "github.com/charmbracelet/lipgloss"

// Palette mirrors the web portfolio's default theme so both frontends share
// branding: mint accent, soft foreground, muted dim, amber warnings.
var (
	colFg      = lipgloss.Color("#E6E6E6")
	colDim     = lipgloss.Color("#7A8590")
	colAccent  = lipgloss.Color("#5EF2B0")
	colAccent2 = lipgloss.Color("#7AA2F7")
	colWarn    = lipgloss.Color("#FFCC66")
	colBorder  = lipgloss.Color("#2A3340")
	colInk     = lipgloss.Color("#0A0E14") // near-black, for text on accent bars
)

// theme groups every reusable lipgloss style. Minimal and monochrome with a
// single mint accent — a premium terminal look a la terminal.shop.
type theme struct {
	base      lipgloss.Style
	brand     lipgloss.Style
	mark      lipgloss.Style
	key       lipgloss.Style
	keyDesc   lipgloss.Style
	title     lipgloss.Style
	accent    lipgloss.Style
	accent2   lipgloss.Style
	dim       lipgloss.Style
	warn      lipgloss.Style
	bullet    lipgloss.Style
	group     lipgloss.Style
	item      lipgloss.Style
	selBar    lipgloss.Style
	tabActive lipgloss.Style
	tabIdle   lipgloss.Style
	rule      lipgloss.Style
	link      lipgloss.Style
}

// newTheme builds all styles from a per-session renderer so colours are chosen
// from the CONNECTING client's terminal, not the server's stdout. A nil
// renderer falls back to the global default (used in tests).
func newTheme(r *lipgloss.Renderer) theme {
	if r == nil {
		r = lipgloss.DefaultRenderer()
	}
	return theme{
		base:      r.NewStyle().Foreground(colFg),
		brand:     r.NewStyle().Foreground(colAccent).Bold(true),
		mark:      r.NewStyle().Foreground(colAccent).Bold(true),
		key:       r.NewStyle().Foreground(colAccent).Bold(true),
		keyDesc:   r.NewStyle().Foreground(colDim),
		title:     r.NewStyle().Foreground(colFg).Bold(true),
		accent:    r.NewStyle().Foreground(colAccent),
		accent2:   r.NewStyle().Foreground(colAccent2),
		dim:       r.NewStyle().Foreground(colDim),
		warn:      r.NewStyle().Foreground(colWarn),
		bullet:    r.NewStyle().Foreground(colAccent),
		group:     r.NewStyle().Foreground(colDim),
		item:      r.NewStyle().Foreground(colFg),
		selBar:    r.NewStyle().Foreground(colInk).Background(colAccent).Bold(true),
		tabActive: r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colAccent).Padding(0, 1),
		tabIdle:   r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colBorder).Padding(0, 1),
		rule:      r.NewStyle().Foreground(colBorder),
		link:      r.NewStyle().Foreground(colAccent2).Underline(true),
	}
}
