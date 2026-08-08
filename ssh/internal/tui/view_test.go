package tui

import (
	"regexp"
	"strings"
	"testing"

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
	m := New(testResume(t), config.Config{}, nil, "guest", "test", w, h)
	m.ready = true
	m.mode = modeMenu
	return m
}

// TestViewDimensions checks the frame is exactly the terminal size at a range
// of sizes and never panics.
func TestViewDimensions(t *testing.T) {
	sizes := [][2]int{{80, 24}, {100, 30}, {120, 40}, {60, 20}, {40, 15}}
	for _, s := range sizes {
		m := newReady(t, s[0], s[1])
		out := m.View()
		lines := strings.Split(out, "\n")
		if len(lines) != s[1] {
			t.Errorf("%dx%d: got %d lines, want %d", s[0], s[1], len(lines), s[1])
		}
	}
}

func TestMenuLeftAligned(t *testing.T) {
	m := newReady(t, 80, 24)
	out := stripANSI(m.View())
	// Every menu item name should begin at the same column (rune-based).
	var cols []int
	for _, name := range []string{"ABOUT", "EXPERIENCE", "PROJECTS", "SKILLS", "CONTACT"} {
		for _, line := range strings.Split(out, "\n") {
			if i := strings.Index(line, name); i >= 0 {
				cols = append(cols, len([]rune(line[:i])))
				break
			}
		}
	}
	for i := 1; i < len(cols); i++ {
		if cols[i] != cols[0] {
			t.Errorf("menu items not left-aligned: columns %v", cols)
			break
		}
	}
}

func TestSectionRendersContent(t *testing.T) {
	m := newReady(t, 100, 30)
	m.openSection(1) // EXPERIENCE
	out := m.View()
	if !strings.Contains(out, "EXPERIENCE") {
		t.Errorf("experience header missing")
	}
	if !strings.Contains(out, m.res.Experience[0].Company) {
		t.Errorf("experience company missing")
	}
	if !strings.Contains(out, "01 / 02") {
		t.Errorf("pager counter missing")
	}
}

func TestSearchModeRenders(t *testing.T) {
	m := newReady(t, 100, 30)
	m.mode = modeSearch
	m.search.SetValue("kubernetes")
	m.results = m.res.Search("kubernetes")
	out := m.View()
	if !strings.Contains(out, "SEARCH") || !strings.Contains(out, "result") {
		t.Errorf("search view missing results header")
	}
}
