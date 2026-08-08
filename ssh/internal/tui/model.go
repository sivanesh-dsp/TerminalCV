// Package tui implements the SSH portfolio as a full-screen Bubble Tea program.
// It is a self-contained application: it never shells out, reads the filesystem
// on behalf of the user, or exposes anything beyond the portfolio itself.
package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
)

type mode int

const (
	modeSplash mode = iota
	modeMenu
	modeSection
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
	innerW, bodyH int
	ready         bool
	quitting      bool

	mode     mode
	prevMode mode
	sections []section
	active   int
	menuIdx  int
	selExp   int
	selProj  int

	vp      viewport.Model
	search  textinput.Model
	results []resume.SearchHit
	resIdx  int
}

// New builds a Model for a session of the given terminal size.
func New(res *resume.Resume, cfg config.Config, store *session.Store, username, version string, width, height int) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "type to search…"
	ti.CharLimit = 64

	m := Model{
		res:      res,
		cfg:      cfg,
		store:    store,
		theme:    newTheme(),
		username: username,
		version:  version,
		now:      time.Now(),
		mode:     modeSplash,
		sections: buildSections(),
		search:   ti,
	}
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
	m.innerW = w - 4 // borders (2) + horizontal padding (2)
	if m.innerW < 10 {
		m.innerW = 10
	}
	// inner height minus header, footer and two rule lines.
	m.bodyH = h - 2 - 4
	if m.bodyH < 1 {
		m.bodyH = 1
	}
	m.vp = viewport.New(m.innerW, m.bodyH)
	m.refreshContent()
}

// refreshContent re-renders the active section into the viewport.
func (m *Model) refreshContent() {
	if m.mode != modeSection || len(m.sections) == 0 {
		return
	}
	sec := m.sections[m.active]
	m.vp.Width = m.innerW
	m.vp.Height = m.bodyH
	m.vp.SetContent(sec.render(m, m.innerW))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case readyMsg:
		if m.mode == modeSplash {
			m.mode = modeMenu
			m.ready = true
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	// Forward other messages (e.g. mouse) to the viewport in section mode.
	if m.mode == modeSection {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C always quits, in every mode.
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
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "?":
		if m.mode == modeHelp {
			m.mode = m.prevMode
		} else {
			m.prevMode = m.mode
			m.mode = modeHelp
		}
		return m, nil
	case "/":
		m.prevMode = m.mode
		m.mode = modeSearch
		m.search.SetValue("")
		m.search.Focus()
		m.results = nil
		m.resIdx = 0
		return m, textinput.Blink
	}

	switch m.mode {
	case modeMenu:
		return m.handleMenuKey(msg)
	case modeSection:
		return m.handleSectionKey(msg)
	case modeHelp:
		if msg.Type == tea.KeyEsc {
			m.mode = m.prevMode
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIdx > 0 {
			m.menuIdx--
		}
	case "down", "j":
		if m.menuIdx < len(m.sections)-1 {
			m.menuIdx++
		}
	case "home", "g":
		m.menuIdx = 0
	case "end", "G":
		m.menuIdx = len(m.sections) - 1
	case "enter", "right", "l", " ":
		m.openSection(m.menuIdx)
	}
	return m, nil
}

func (m Model) handleSectionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	sec := m.sections[m.active]
	switch msg.String() {
	case "esc", "backspace", "h":
		m.mode = modeMenu
		m.menuIdx = m.active
		return m, nil
	case "left", "shift+tab", "[":
		m.openSection((m.active - 1 + len(m.sections)) % len(m.sections))
		return m, nil
	case "right", "tab", "]":
		m.openSection((m.active + 1) % len(m.sections))
		return m, nil
	}
	if isPager(sec.id) {
		switch msg.String() {
		case "up", "k":
			m.pagerMove(-1)
			return m, nil
		case "down", "j":
			m.pagerMove(1)
			return m, nil
		}
	}
	// Prose sections: delegate scrolling to the viewport.
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *Model) pagerMove(delta int) {
	sec := m.sections[m.active]
	switch sec.id {
	case "experience":
		n := len(m.res.Experience)
		if n > 0 {
			m.selExp = (m.selExp + delta + n) % n
		}
	case "projects":
		n := len(m.res.Projects)
		if n > 0 {
			m.selProj = (m.selProj + delta + n) % n
		}
	}
	m.refreshContent()
	m.vp.GotoTop()
}

func (m *Model) openSection(i int) {
	m.active = i
	m.mode = modeSection
	m.refreshContent()
	m.vp.GotoTop()
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = m.prevMode
		if m.mode == modeSearch {
			m.mode = modeMenu
		}
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

func (m *Model) jumpToResult(hit resume.SearchHit) {
	for i, s := range m.sections {
		if strings.EqualFold(s.name, hit.Section) {
			m.openSection(i)
			return
		}
	}
	m.mode = modeMenu
}

// View is defined in view.go.

var _ tea.Model = Model{}
