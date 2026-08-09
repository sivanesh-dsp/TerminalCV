package tui

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
)

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]")

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

func testResume(t *testing.T) *resume.Resume {
	t.Helper()
	r, err := resume.Load("../../../shared/resume.json")
	if err != nil {
		t.Fatalf("load resume: %v", err)
	}
	return r
}

func newReady(t *testing.T, w, h int) Model {
	m := New(testResume(t), config.Config{}, nil, nil, &bytes.Buffer{}, "guest", "test", w, h)
	m.ready = true
	m.mode = modeBrowse
	return m
}

// TestViewNoPanic renders at several sizes and checks the frame never exceeds
// the terminal height and never panics.
func TestViewNoPanic(t *testing.T) {
	for _, s := range [][2]int{{80, 24}, {100, 30}, {120, 40}, {60, 20}, {40, 15}} {
		m := newReady(t, s[0], s[1])
		out := m.View()
		lines := strings.Split(out, "\n")
		if len(lines) > s[1] {
			t.Errorf("%dx%d: %d lines exceeds height %d", s[0], s[1], len(lines), s[1])
		}
	}
}

func TestTabsAndSelectionRender(t *testing.T) {
	m := newReady(t, 100, 30)
	out := stripANSI(m.View())
	for _, tab := range []string{"About", "Experience", "Projects", "Skills", "Resume", "Contact"} {
		if !strings.Contains(out, tab) {
			t.Errorf("tab %q missing from tab bar", tab)
		}
	}
	// The about detail should be visible by default.
	if !strings.Contains(out, m.res.Name) {
		t.Errorf("about detail (name) not rendered")
	}
}

func TestTabAcceleratorAndDetail(t *testing.T) {
	m := newReady(t, 100, 30)
	// Jump to experience via 'e'.
	nm, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m2 := nm.(Model)
	if m2.tabs[m2.tab].id != "experience" {
		t.Fatalf("expected experience tab, got %q", m2.tabs[m2.tab].id)
	}
	out := stripANSI(m2.View())
	if !strings.Contains(out, m2.res.Experience[0].Company) {
		t.Errorf("experience detail not rendered")
	}
	// Left list should show the company as an item.
	if !strings.Contains(out, "~ Experience ~") {
		t.Errorf("group header missing")
	}
}

func TestSearchJump(t *testing.T) {
	m := newReady(t, 100, 30)
	m.results = m.res.Search("kubernetes")
	if len(m.results) == 0 {
		t.Fatal("expected kubernetes results")
	}
	m.jumpToResult(m.results[0])
	if m.mode != modeBrowse {
		t.Errorf("jump should return to browse mode")
	}
}

func TestContactCopyEmitsOSC52(t *testing.T) {
	m := newReady(t, 100, 30)
	// Jump to the contact tab.
	nm, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = nm.(Model)
	if m.tabs[m.tab].id != "contact" {
		t.Fatalf("expected contact tab, got %q", m.tabs[m.tab].id)
	}
	it := m.curItem()
	if it == nil || it.copy == "" {
		t.Fatalf("contact item should be copyable")
	}
	// Press 'y' to copy.
	nm, _ = m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = nm.(Model)
	// The OSC 52 sequence for the value must have been written to the session.
	buf := m.out.(*bytes.Buffer)
	if !strings.Contains(buf.String(), osc52(it.copy)) {
		t.Errorf("expected OSC 52 clipboard write for the copied value")
	}
	// The detail pane shows the confirmation.
	if !strings.Contains(stripANSI(m.View()), "copied to clipboard") {
		t.Errorf("expected copied confirmation in detail")
	}
}

func TestResumeTabIsFlat(t *testing.T) {
	m := newReady(t, 100, 30)
	nm, _ := m.handleBrowseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = nm.(Model)
	out := stripANSI(m.View())
	// The old credentials/highlights sub-headers must be gone.
	if strings.Contains(out, "credentials") || strings.Contains(out, "highlights") {
		t.Errorf("resume tab should not show sub-group headers")
	}
	for _, item := range []string{"Certifications", "Education", "Achievements", "Timeline"} {
		if !strings.Contains(out, item) {
			t.Errorf("resume item %q missing", item)
		}
	}
}
