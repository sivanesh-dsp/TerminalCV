package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// itemDef is one navigable entry in the left list. group is the "~ header ~" it
// sits under; render produces the right-pane detail for the given inner width.
type itemDef struct {
	group  string
	label  string
	render func(m *Model, w int) string
}

// tabDef is a top-bar category. key is a single-letter accelerator (like
// terminal.shop's "s shop"). It owns an ordered list of items.
type tabDef struct {
	id    string
	label string
	key   rune
	items []itemDef
}

// buildTabs assembles the whole portfolio as tabs → items, all sourced from
// resume.json. Experience and projects become one item each so the left list
// works exactly like a product list with a live detail pane.
func (m *Model) buildTabs() []tabDef {
	r := m.res
	tabs := []tabDef{
		{id: "about", label: "about", key: 'a', items: []itemDef{
			{group: "profile", label: "about", render: renderAbout},
		}},
	}

	// EXPERIENCE — one item per role.
	var exp []itemDef
	for i := range r.Experience {
		i := i
		exp = append(exp, itemDef{
			group:  "experience",
			label:  r.Experience[i].Company,
			render: func(m *Model, w int) string { return renderExperienceItem(m, i, w) },
		})
	}
	if len(exp) > 0 {
		tabs = append(tabs, tabDef{id: "experience", label: "experience", key: 'e', items: exp})
	}

	// PROJECTS — one item per project.
	var proj []itemDef
	for i := range r.Projects {
		i := i
		proj = append(proj, itemDef{
			group:  "projects",
			label:  shortLabel(r.Projects[i].Name, 24),
			render: func(m *Model, w int) string { return renderProjectItem(m, i, w) },
		})
	}
	if len(proj) > 0 {
		tabs = append(tabs, tabDef{id: "projects", label: "projects", key: 'p', items: proj})
	}

	// SKILLS — skills + tech stack.
	tabs = append(tabs, tabDef{id: "skills", label: "skills", key: 's', items: []itemDef{
		{group: "toolkit", label: "skills", render: renderSkills},
		{group: "toolkit", label: "tech stack", render: renderTechStack},
	}})

	// RESUME — credentials + highlights.
	tabs = append(tabs, tabDef{id: "resume", label: "resume", key: 'r', items: []itemDef{
		{group: "credentials", label: "certifications", render: renderCerts},
		{group: "credentials", label: "education", render: renderEducation},
		{group: "highlights", label: "achievements", render: renderAchievements},
		{group: "highlights", label: "timeline", render: renderTimeline},
	}})

	// CONTACT.
	tabs = append(tabs, tabDef{id: "contact", label: "contact", key: 'c', items: []itemDef{
		{group: "connect", label: "contact", render: renderContact},
	}})

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
		{"experience", r.ExperienceLabel(m.now)},
		{"employers", fmt.Sprintf("%d", r.Employers())},
		{"technologies", fmt.Sprintf("%d", len(r.AllTechnologies()))},
		{"certifications", fmt.Sprintf("%d", len(r.Certifications))},
		{"location", r.Contact.Location},
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

func renderContact(m *Model, w int) string {
	t, r := m.theme, m.res
	c := r.Contact
	var b strings.Builder
	row := func(label, value, url string) {
		disp := value
		if url != "" {
			disp = osc8(url, value)
		}
		b.WriteString(t.dim.Render(padRight(label, 11)) + t.link.Render(disp) + "\n")
	}
	if c.Email != "" {
		row("email", c.Email, "mailto:"+c.Email)
	}
	if c.Phone != "" {
		b.WriteString(t.dim.Render(padRight("phone", 11)) + t.base.Render(c.Phone) + "\n")
	}
	if c.GitHub != nil {
		row("github", c.GitHub.URL, c.GitHub.URL)
	}
	if c.LinkedIn != nil {
		row("linkedin", c.LinkedIn.URL, c.LinkedIn.URL)
	}
	if c.Portfolio != nil {
		row("portfolio", c.Portfolio.URL, c.Portfolio.URL)
	}
	if c.Location != "" {
		b.WriteString(t.dim.Render(padRight("location", 11)) + t.base.Render(c.Location) + "\n")
	}
	b.WriteString("\n" + t.dim.Render("Links are clickable where OSC 8 hyperlinks are supported."))
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
