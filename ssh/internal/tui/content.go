package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// itemDef is one navigable entry in the left list. group is the "~ Header ~" it
// sits under ("" = no header); render produces the right-pane detail. copy, when
// non-empty, is the value yanked to the clipboard when the item is activated.
type itemDef struct {
	group  string
	label  string
	copy   string
	render func(m *Model, w int) string
}

// tabDef is a top-bar category with a single-letter accelerator.
type tabDef struct {
	id    string
	label string
	key   rune
	items []itemDef
}

// buildTabs assembles the whole portfolio as tabs → items, all sourced from
// resume.json.
func (m *Model) buildTabs() []tabDef {
	r := m.res
	tabs := []tabDef{
		{id: "about", label: "About", key: 'a', items: []itemDef{
			{group: "Profile", label: "About", render: renderAbout},
		}},
	}

	// EXPERIENCE — one item per role.
	var exp []itemDef
	for i := range r.Experience {
		i := i
		exp = append(exp, itemDef{
			group:  "Experience",
			label:  r.Experience[i].Company,
			render: func(m *Model, w int) string { return renderExperienceItem(m, i, w) },
		})
	}
	if len(exp) > 0 {
		tabs = append(tabs, tabDef{id: "experience", label: "Experience", key: 'e', items: exp})
	}

	// PROJECTS — one item per project.
	var proj []itemDef
	for i := range r.Projects {
		i := i
		proj = append(proj, itemDef{
			group:  "Projects",
			label:  shortLabel(r.Projects[i].Name, 24),
			render: func(m *Model, w int) string { return renderProjectItem(m, i, w) },
		})
	}
	if len(proj) > 0 {
		tabs = append(tabs, tabDef{id: "projects", label: "Projects", key: 'p', items: proj})
	}

	// SKILLS — skills + tech stack.
	tabs = append(tabs, tabDef{id: "skills", label: "Skills", key: 's', items: []itemDef{
		{group: "Toolkit", label: "Skills", render: renderSkills},
		{group: "Toolkit", label: "Tech Stack", render: renderTechStack},
	}})

	// RESUME — flat list, no sub-grouping (credentials/highlights merged).
	tabs = append(tabs, tabDef{id: "resume", label: "Resume", key: 'r', items: []itemDef{
		{label: "Certifications", render: renderCerts},
		{label: "Education", render: renderEducation},
		{label: "Achievements", render: renderAchievements},
		{label: "Timeline", render: renderTimeline},
	}})

	// CONTACT — one item per field, each copyable to the clipboard.
	var contact []itemDef
	addField := func(label, value string) {
		if value == "" {
			return
		}
		contact = append(contact, itemDef{
			group:  "Connect",
			label:  label,
			copy:   value,
			render: func(m *Model, w int) string { return renderContactField(m, label, value, w) },
		})
	}
	c := r.Contact
	addField("Email", c.Email)
	if c.GitHub != nil {
		addField("GitHub", c.GitHub.URL)
	}
	if c.LinkedIn != nil {
		addField("LinkedIn", c.LinkedIn.URL)
	}
	if c.Portfolio != nil {
		addField("Portfolio", c.Portfolio.URL)
	}
	addField("Phone", c.Phone)
	addField("Location", c.Location)
	tabs = append(tabs, tabDef{id: "contact", label: "Contact", key: 'c', items: contact})

	return tabs
}

func shortLabel(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	return truncate(s, n)
}

/* ----------------------------- detail renderers ---------------------------- */

func renderAbout(m *Model, w int) string {
	t, r := m.theme, m.res
	var b strings.Builder
	b.WriteString(t.title.Render(r.Name) + "\n")
	b.WriteString(t.dim.Render(r.Title) + "\n\n")
	b.WriteString(t.base.Render(wrap(r.Summary, w)) + "\n\n")
	facts := [][2]string{
		{"Experience", r.ExperienceLabel(m.now)},
		{"Employers", fmt.Sprintf("%d", r.Employers())},
		{"Technologies", fmt.Sprintf("%d", len(r.AllTechnologies()))},
		{"Certifications", fmt.Sprintf("%d", len(r.Certifications))},
		{"Location", r.Contact.Location},
	}
	for _, f := range facts {
		b.WriteString(t.dim.Render(padRight(f[0], 15)) + t.accent.Render(f[1]) + "\n")
	}
	return b.String()
}

func renderExperienceItem(m *Model, i, w int) string {
	t, r := m.theme, m.res
	if i < 0 || i >= len(r.Experience) {
		return ""
	}
	e := r.Experience[i]
	var b strings.Builder
	b.WriteString(t.title.Render(e.Role) + "\n")
	b.WriteString(t.accent2.Render(e.Company))
	if e.Location != "" {
		b.WriteString(t.dim.Render("  ·  " + e.Location))
	}
	b.WriteString("\n")
	b.WriteString(t.accent.Render(e.Start+" – "+e.End) + "\n\n")
	for _, h := range e.Highlights {
		lines := strings.Split(wrap(h, w-2), "\n")
		for j, ln := range lines {
			if j == 0 {
				b.WriteString(t.bullet.Render("• ") + t.base.Render(ln) + "\n")
			} else {
				b.WriteString("  " + t.base.Render(ln) + "\n")
			}
		}
	}
	return b.String()
}

func renderProjectItem(m *Model, i, w int) string {
	t, r := m.theme, m.res
	if i < 0 || i >= len(r.Projects) {
		return ""
	}
	p := r.Projects[i]
	var b strings.Builder
	b.WriteString(t.title.Render(p.Name) + "\n\n")
	b.WriteString(t.base.Render(wrap(p.Description, w)) + "\n")
	if len(p.Tech) > 0 {
		b.WriteString("\n" + t.dim.Render("tech  ") + t.accent.Render(strings.Join(p.Tech, "  ·  ")) + "\n")
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
	b.WriteString("\n" + t.dim.Render(fmt.Sprintf("%d technologies · %d categories",
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

// renderContactField shows one contact value prominently with a copy hint.
// Values are plain (not OSC 8) so they render cleanly and are mouse-selectable
// even in terminals without hyperlink support.
func renderContactField(m *Model, label, value string, w int) string {
	t := m.theme
	var b strings.Builder
	b.WriteString(t.dim.Render(label) + "\n\n")
	b.WriteString(t.accent2.Render(wrap(value, w)) + "\n\n")
	if m.copiedLabel == label {
		b.WriteString(t.accent.Render("✓ copied to clipboard"))
	} else {
		b.WriteString(t.dim.Render("press ") + t.key.Render("y") + t.dim.Render(" or ") +
			t.key.Render("↵") + t.dim.Render(" to copy"))
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
