package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// section is one navigable portfolio screen. render returns the body content
// for the given inner width; the model wraps it in the frame/header/footer.
type section struct {
	id     string
	name   string
	render func(m *Model, w int) string
}

// sections defines the ordered portfolio structure. The same order drives the
// menu, ←/→ paging and search jumps.
func buildSections() []section {
	return []section{
		{"about", "ABOUT", renderAbout},
		{"experience", "EXPERIENCE", renderExperience},
		{"projects", "PROJECTS", renderProjects},
		{"skills", "SKILLS", renderSkills},
		{"techstack", "TECH STACK", renderTechStack},
		{"certifications", "CERTIFICATIONS", renderCerts},
		{"education", "EDUCATION", renderEducation},
		{"achievements", "ACHIEVEMENTS", renderAchievements},
		{"timeline", "TIMELINE", renderTimeline},
		{"contact", "CONTACT", renderContact},
	}
}

// isPager reports whether a section shows one item at a time (↑/↓ browse).
func isPager(id string) bool { return id == "experience" || id == "projects" }

func renderAbout(m *Model, w int) string {
	t, r := m.theme, m.res
	var b strings.Builder
	b.WriteString(t.headerName.Render(r.Name) + "\n")
	b.WriteString(t.dim.Render(r.Title) + "\n\n")
	b.WriteString(t.base.Render(wrap(r.Summary, w)) + "\n\n")
	b.WriteString(ruleLine(t, min(w, 48)) + "\n")
	facts := [][2]string{
		{"Experience", r.ExperienceLabel(m.now)},
		{"Employers", fmt.Sprintf("%d", r.Employers())},
		{"Technologies", fmt.Sprintf("%d", len(r.AllTechnologies()))},
		{"Certifications", fmt.Sprintf("%d", len(r.Certifications))},
		{"Location", r.Contact.Location},
	}
	for _, f := range facts {
		b.WriteString(t.dim.Render(padRight(f[0], 15)) + t.base.Render(f[1]) + "\n")
	}
	return b.String()
}

func renderExperience(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Experience) == 0 {
		return t.dim.Render("No experience listed.")
	}
	e := r.Experience[m.selExp]
	var b strings.Builder
	b.WriteString(t.title.Render(e.Role) + "\n")
	b.WriteString(t.accent2.Render(e.Company))
	if e.Location != "" {
		b.WriteString(t.dim.Render("  ·  " + e.Location))
	}
	b.WriteString("\n")
	b.WriteString(t.dim.Render(e.Start+" – "+e.End) + "\n\n")
	for _, h := range e.Highlights {
		bullet := t.bullet.Render("• ")
		lines := strings.Split(wrap(h, w-2), "\n")
		for i, ln := range lines {
			if i == 0 {
				b.WriteString(bullet + t.base.Render(ln) + "\n")
			} else {
				b.WriteString("  " + t.base.Render(ln) + "\n")
			}
		}
	}
	return b.String()
}

func renderProjects(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Projects) == 0 {
		return t.dim.Render("No projects listed.")
	}
	p := r.Projects[m.selProj]
	var b strings.Builder
	b.WriteString(t.title.Render(p.Name) + "\n\n")
	b.WriteString(t.base.Render(wrap(p.Description, w)) + "\n")
	if len(p.Tech) > 0 {
		b.WriteString("\n" + t.dim.Render("Tech: ") + t.accent.Render(strings.Join(p.Tech, "  ·  ")) + "\n")
	}
	return b.String()
}

func renderSkills(m *Model, w int) string {
	t, r := m.theme, m.res
	var b strings.Builder
	for i, c := range r.Skills {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(t.accent.Render(c.Name) + "\n")
		b.WriteString(ruleLine(t, min(w, 44)) + "\n")
		b.WriteString(t.base.Render(wrap(strings.Join(c.Skills, "   "), w)) + "\n")
	}
	return b.String()
}

func renderTechStack(m *Model, w int) string {
	t, r := m.theme, m.res
	label := 0
	for _, c := range r.Skills {
		if l := lipgloss.Width(c.Name); l > label {
			label = l
		}
	}
	if label > 22 {
		label = 22
	}
	var b strings.Builder
	for _, c := range r.Skills {
		head := t.accent.Render(padRight(truncate(c.Name, label), label))
		vals := wrap(strings.Join(c.Skills, ", "), w-label-2)
		lines := strings.Split(vals, "\n")
		for i, ln := range lines {
			if i == 0 {
				b.WriteString(head + "  " + t.base.Render(ln) + "\n")
			} else {
				b.WriteString(strings.Repeat(" ", label+2) + t.base.Render(ln) + "\n")
			}
		}
	}
	b.WriteString("\n" + t.dim.Render(fmt.Sprintf("%d technologies across %d categories.",
		len(r.AllTechnologies()), len(r.Skills))))
	return b.String()
}

func renderCerts(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Certifications) == 0 {
		return t.dim.Render("No certifications listed.")
	}
	var b strings.Builder
	for _, c := range r.Certifications {
		b.WriteString(t.bullet.Render("• ") + t.base.Render(c.Name) + "\n")
		if c.Issuer != "" {
			b.WriteString("  " + t.dim.Render(c.Issuer) + "\n")
		}
	}
	return b.String()
}

func renderEducation(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Education) == 0 {
		return t.dim.Render("No education listed.")
	}
	var b strings.Builder
	for i, ed := range r.Education {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(t.title.Render(ed.Degree) + "\n")
		b.WriteString(t.accent2.Render(ed.Institution) + "\n")
		meta := ed.Start + " – " + ed.End
		if ed.Location != "" {
			meta += "  ·  " + ed.Location
		}
		b.WriteString(t.dim.Render(meta) + "\n")
	}
	return b.String()
}

func renderAchievements(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Achievements) == 0 {
		return t.dim.Render("No achievements listed.")
	}
	var b strings.Builder
	for _, a := range r.Achievements {
		lines := strings.Split(wrap(a, w-2), "\n")
		for i, ln := range lines {
			if i == 0 {
				b.WriteString(t.bullet.Render("▸ ") + t.base.Render(ln) + "\n")
			} else {
				b.WriteString("  " + t.base.Render(ln) + "\n")
			}
		}
	}
	return b.String()
}

func renderTimeline(m *Model, w int) string {
	t, r := m.theme, m.res
	if len(r.Timeline) == 0 {
		return t.dim.Render("No timeline available.")
	}
	dateW := 0
	for _, ev := range r.Timeline {
		if l := lipgloss.Width(ev.Date); l > dateW {
			dateW = l
		}
	}
	var b strings.Builder
	for i, ev := range r.Timeline {
		date := t.accent.Render(padRight(ev.Date, dateW))
		b.WriteString(date + t.rule.Render(" ──── ") + t.base.Render(ev.Title) + "\n")
		if ev.Subtitle != "" {
			b.WriteString(strings.Repeat(" ", dateW) + t.rule.Render("   │  ") +
				t.dim.Render(truncate(ev.Subtitle, w-dateW-6)) + "\n")
		} else if i < len(r.Timeline)-1 {
			b.WriteString(strings.Repeat(" ", dateW) + t.rule.Render("   │") + "\n")
		}
	}
	return b.String()
}

func renderContact(m *Model, w int) string {
	t, r := m.theme, m.res
	c := r.Contact
	var b strings.Builder
	row := func(label, value, url string) {
		disp := value
		if url != "" {
			disp = osc8(url, value)
		}
		b.WriteString(t.dim.Render(padRight(label, 12)) + t.link.Render(disp) + "\n")
	}
	if c.Email != "" {
		row("Email", c.Email, "mailto:"+c.Email)
	}
	if c.Phone != "" {
		b.WriteString(t.dim.Render(padRight("Phone", 12)) + t.base.Render(c.Phone) + "\n")
	}
	if c.GitHub != nil {
		row("GitHub", c.GitHub.URL, c.GitHub.URL)
	}
	if c.LinkedIn != nil {
		row("LinkedIn", c.LinkedIn.URL, c.LinkedIn.URL)
	}
	if c.Portfolio != nil {
		row("Portfolio", c.Portfolio.URL, c.Portfolio.URL)
	}
	if c.Location != "" {
		b.WriteString(t.dim.Render(padRight("Location", 12)) + t.base.Render(c.Location) + "\n")
	}
	b.WriteString("\n" + ruleLine(t, min(w, 44)) + "\n")
	b.WriteString(t.dim.Render("Links are clickable in terminals that support OSC 8 hyperlinks."))
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
