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
)

// theme groups every reusable lipgloss style. Kept minimal and monochrome with
// a single mint accent — closer to a premium terminal app than a neon hacker UI.
type theme struct {
	base       lipgloss.Style
	headerName lipgloss.Style
	headerTag  lipgloss.Style
	key        lipgloss.Style
	keyDesc    lipgloss.Style
	title      lipgloss.Style
	accent     lipgloss.Style
	accent2    lipgloss.Style
	dim        lipgloss.Style
	warn       lipgloss.Style
	bullet     lipgloss.Style
	selected   lipgloss.Style
	menuItem   lipgloss.Style
	menuActive lipgloss.Style
	rule       lipgloss.Style
	link       lipgloss.Style
}

func newTheme() theme {
	return theme{
		base:       lipgloss.NewStyle().Foreground(colFg),
		headerName: lipgloss.NewStyle().Foreground(colAccent).Bold(true),
		headerTag:  lipgloss.NewStyle().Foreground(colDim),
		key:        lipgloss.NewStyle().Foreground(colAccent).Bold(true),
		keyDesc:    lipgloss.NewStyle().Foreground(colDim),
		title:      lipgloss.NewStyle().Foreground(colAccent).Bold(true),
		accent:     lipgloss.NewStyle().Foreground(colAccent),
		accent2:    lipgloss.NewStyle().Foreground(colAccent2),
		dim:        lipgloss.NewStyle().Foreground(colDim),
		warn:       lipgloss.NewStyle().Foreground(colWarn),
		bullet:     lipgloss.NewStyle().Foreground(colAccent),
		selected:   lipgloss.NewStyle().Foreground(colAccent).Bold(true),
		menuItem:   lipgloss.NewStyle().Foreground(colFg),
		menuActive: lipgloss.NewStyle().Foreground(colAccent).Bold(true),
		rule:       lipgloss.NewStyle().Foreground(colBorder),
		link:       lipgloss.NewStyle().Foreground(colAccent2).Underline(true),
	}
}
