// Package tui implements the SSH portfolio as a Bubble Tea program styled after
// terminal.shop: a minimal splash, a centered master-detail browser (tab bar +
// grouped list + live detail pane), and a keyed footer. It is fully sandboxed —
// it never shells out or exposes anything beyond the portfolio.
package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
)

type mode int

const (
	modeSplash mode = iota
	modeBrowse
	modeSearch
	modeHelp
)

type readyMsg struct{}

// Model is the root Bubble Tea state for one SSH session.
type Model struct {
	res      *resume.Resume
	cfg      config.Config
	store    *session.Store
	theme    theme
	username string
	version  string
	now      time.Time

	width, height int
	ready         bool
	quitting      bool

	mode     mode
	prevMode mode

	tabs        []tabDef
	tab         int  // active tab index
	sel         int  // selected item index within the active tab
	detailFocus bool // when true, ↑/↓ scroll the detail pane

	// layout (recomputed on resize)
	contentW, leftW, rightW, bodyH int

	vp      viewport.Model
	search  textinput.Model
	results []resume.SearchHit
	resIdx  int
}

// New builds a Model for a session of the given terminal size. renderer is the
// per-session lipgloss renderer (nil → global default, used in tests).
func New(res *resume.Resume, cfg config.Config, store *session.Store, renderer *lipgloss.Renderer, username, version string, width, height int) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "type to search…"
	ti.CharLimit = 64

	m := Model{
		res:      res,
		cfg:      cfg,
		store:    store,
		theme:    newTheme(renderer),
		username: username,
		version:  version,
		now:      time.Now(),
		mode:     modeSplash,
		search:   ti,
	}
	m.tabs = m.buildTabs()
	m.setSize(width, height)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(700*time.Millisecond, func(time.Time) tea.Msg { return readyMsg{} })
}

func (m *Model) setSize(w, h int) {
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}
	m.width, m.height = w, h

	// Centered content column (terminal.shop keeps generous margins). On narrow
	// terminals it uses most of the width (simplified layout).
	m.contentW = w - 8
	if m.contentW > 90 {
		m.contentW = 90
	}
	if w < 72 {
		m.contentW = w - 2
	}
	if m.contentW < 30 {
		m.contentW = 30
	}

	m.leftW = 24
	if m.contentW < 60 {
		m.leftW = m.contentW/3 + 2
	}
	m.rightW = m.contentW - m.leftW - 3 // 3 = gutter
	if m.rightW < 16 {
		m.rightW = 16
	}

	// Body height. Keep it capped so the whole block floats in the middle of
	// tall terminals with black margins (the terminal.shop look).
	m.bodyH = h - 12
	if m.bodyH < 6 {
		m.bodyH = 6
	}
	if m.bodyH > 22 {
		m.bodyH = 22
	}

	m.vp = viewport.New(m.rightW, m.bodyH)
	m.refreshDetail()
}

// curItem returns the currently selected item, or nil if none.
func (m *Model) curItem() *itemDef {
	if m.tab < 0 || m.tab >= len(m.tabs) {
		return nil
	}
	items := m.tabs[m.tab].items
	if m.sel < 0 || m.sel >= len(items) {
		return nil
	}
	return &items[m.sel]
}

// refreshDetail re-renders the selected item into the detail viewport.
func (m *Model) refreshDetail() {
	m.vp.Width = m.rightW
	m.vp.Height = m.bodyH
	if it := m.curItem(); it != nil {
		m.vp.SetContent(it.render(m, m.rightW))
	} else {
		m.vp.SetContent("")
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case readyMsg:
		if m.mode == modeSplash {
			m.mode = modeBrowse
			m.ready = true
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	if m.mode == modeBrowse && m.detailFocus {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}
	if m.mode == modeSplash {
		return m, nil
	}
	if m.mode == modeSearch {
		return m.handleSearchKey(msg)
	}

	switch msg.String() {
	case "?":
		if m.mode == modeHelp {
			m.mode = m.prevMode
		} else {
			m.prevMode = m.mode
			m.mode = modeHelp
		}
		return m, nil
	case "/":
		m.prevMode = modeBrowse
		m.mode = modeSearch
		m.search.SetValue("")
		m.search.Focus()
		m.results = nil
		m.resIdx = 0
		return m, textinput.Blink
	}

	if m.mode == modeHelp {
		if msg.Type == tea.KeyEsc || msg.String() == "q" {
			m.mode = m.prevMode
		}
		return m, nil
	}
	return m.handleBrowseKey(msg)
}

func (m Model) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Single-letter tab accelerators (like terminal.shop's "s shop").
	if len(msg.Runes) == 1 && !m.detailFocus {
		for i, tb := range m.tabs {
			if msg.Runes[0] == tb.key {
				m.selectTab(i)
				return m, nil
			}
		}
	}

	switch msg.String() {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "left", "shift+tab", "[", "h":
		m.selectTab((m.tab - 1 + len(m.tabs)) % len(m.tabs))
		return m, nil
	case "right", "tab", "]", "l":
		m.selectTab((m.tab + 1) % len(m.tabs))
		return m, nil
	case "enter":
		m.detailFocus = true
		return m, nil
	case "esc":
		m.detailFocus = false
		return m, nil
	case "up", "k":
		if m.detailFocus {
			m.vp.LineUp(1)
		} else {
			m.moveSel(-1)
		}
		return m, nil
	case "down", "j":
		if m.detailFocus {
			m.vp.LineDown(1)
		} else {
			m.moveSel(1)
		}
		return m, nil
	case "pgup":
		m.vp.HalfViewUp()
		return m, nil
	case "pgdown", " ":
		m.vp.HalfViewDown()
		return m, nil
	case "home", "g":
		if m.detailFocus {
			m.vp.GotoTop()
		} else {
			m.sel = 0
			m.detailFocus = false
			m.refreshDetail()
		}
		return m, nil
	case "end", "G":
		if m.detailFocus {
			m.vp.GotoBottom()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) selectTab(i int) {
	m.tab = i
	m.sel = 0
	m.detailFocus = false
	m.refreshDetail()
	m.vp.GotoTop()
}

func (m *Model) moveSel(delta int) {
	n := len(m.tabs[m.tab].items)
	if n == 0 {
		return
	}
	m.sel = (m.sel + delta + n) % n
	m.refreshDetail()
	m.vp.GotoTop()
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeBrowse
		return m, nil
	case tea.KeyUp:
		if m.resIdx > 0 {
			m.resIdx--
		}
		return m, nil
	case tea.KeyDown:
		if m.resIdx < len(m.results)-1 {
			m.resIdx++
		}
		return m, nil
	case tea.KeyEnter:
		if len(m.results) > 0 {
			m.jumpToResult(m.results[m.resIdx])
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.results = m.res.Search(m.search.Value())
	m.resIdx = 0
	return m, cmd
}

// jumpToResult opens the tab/item whose section matches the hit.
func (m *Model) jumpToResult(hit resume.SearchHit) {
	m.mode = modeBrowse
	target := sectionToTab(hit.Section)
	for i, tb := range m.tabs {
		if tb.id == target {
			m.selectTab(i)
			return
		}
	}
}

// sectionToTab maps a Search hit's section label to a tab id.
func sectionToTab(section string) string {
	switch section {
	case "ABOUT":
		return "about"
	case "EXPERIENCE":
		return "experience"
	case "PROJECTS":
		return "projects"
	case "SKILLS", "TECH STACK":
		return "skills"
	case "CERTIFICATIONS", "EDUCATION", "ACHIEVEMENTS", "TIMELINE":
		return "resume"
	case "CONTACT":
		return "contact"
	}
	return "about"
}

var _ tea.Model = Model{}
