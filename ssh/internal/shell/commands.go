package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
)

// vfiles is the virtual filesystem exposed by ls / cat / tree / completion.
var vfiles = []string{
	"about.txt",
	"experience.txt",
	"skills.txt",
	"projects.txt",
	"certifications.txt",
	"education.txt",
	"achievements.txt",
	"timeline.txt",
	"contact.txt",
	"skills.json",
	"resume.pdf",
}

func out(ctx *Context, s string) { io.WriteString(ctx.Out, s) }

func kv(st stStyle, label, value string) string {
	return st.Dim(padRight(label, 11)) + value + "\n"
}

// stStyle is a tiny alias to keep signatures short.
type stStyle = interface {
	Accent(string) string
	Accent2(string) string
	Dim(string) string
	Warn(string) string
	Err(string) string
	Ok(string) string
	Link(string) string
	Bold(string) string
	Underline(string) string
	Hyperlink(string, string) string
}

/* ------------------------------ section renderers ------------------------------ */

func renderAbout(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	r := ctx.Resume
	var b strings.Builder
	b.WriteString(heading(st, "about"))
	for _, l := range wrapText(r.Summary, w) {
		b.WriteString(l + "\n")
	}
	b.WriteString("\n")
	b.WriteString(kv(st, "name", st.Bold(r.Name)))
	b.WriteString(kv(st, "role", r.Title))
	b.WriteString(kv(st, "location", r.Contact.Location))
	b.WriteString(kv(st, "focus", "Platform Engineering · DevOps · Kubernetes · IaC · CI/CD"))
	return b.String()
}

func renderExperience(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	var b strings.Builder
	b.WriteString(heading(st, "experience"))
	for i, e := range ctx.Resume.Experience {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(st.Bold(e.Role) + st.Dim("  ("+e.Start+" – "+e.End+")") + "\n")
		b.WriteString(st.Accent2(e.Company) + st.Dim(" · "+e.Location) + "\n")
		for _, h := range e.Highlights {
			b.WriteString(markerLines(st.Accent("▸ "), 2, w, h))
		}
	}
	return b.String()
}

func renderProjects(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	var b strings.Builder
	b.WriteString(heading(st, "projects"))
	b.WriteString(st.Dim("Key initiatives drawn from professional experience.") + "\n\n")
	for i, p := range ctx.Resume.Projects {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(st.Accent("◈ ") + st.Bold(p.Name) + "\n")
		for _, l := range wrapText(p.Description, w-2) {
			b.WriteString("  " + l + "\n")
		}
		b.WriteString("  " + st.Dim("["+strings.Join(p.Tech, "] [")+"]") + "\n")
	}
	return b.String()
}

func renderSkills(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	var b strings.Builder
	b.WriteString(heading(st, "skills"))
	for _, c := range ctx.Resume.Skills {
		b.WriteString(st.Accent2(c.Name) + "\n")
		b.WriteString(st.Dim(strings.Repeat("─", utf8.RuneCountInString(c.Name))) + "\n")
		for _, l := range wrapText(strings.Join(c.Skills, " · "), w) {
			b.WriteString(l + "\n")
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderTechstack(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	r := ctx.Resume
	var b strings.Builder
	b.WriteString(heading(st, "techstack"))
	b.WriteString(st.Dim(fmt.Sprintf("%d technologies across the stack.", len(r.AllTechnologies()))) + "\n\n")
	for i, c := range r.Skills {
		branch, pipe := "├─ ", "│  "
		if i == len(r.Skills)-1 {
			branch, pipe = "└─ ", "   "
		}
		b.WriteString(st.Dim(branch) + st.Accent2(c.Name) + "\n")
		for _, l := range wrapText(strings.Join(c.Skills, " · "), w-3) {
			b.WriteString(st.Dim(pipe) + l + "\n")
		}
	}
	return b.String()
}

func renderCerts(ctx *Context) string {
	st := ctx.Style
	var b strings.Builder
	b.WriteString(heading(st, "certifications"))
	for _, c := range ctx.Resume.Certifications {
		b.WriteString(st.Ok("✔ ") + st.Bold(c.Name) + "\n")
		if c.Issuer != "" {
			b.WriteString("  " + st.Dim(c.Issuer) + "\n")
		}
	}
	return b.String()
}

func renderEducation(ctx *Context) string {
	st := ctx.Style
	var b strings.Builder
	b.WriteString(heading(st, "education"))
	for _, e := range ctx.Resume.Education {
		b.WriteString(st.Bold(e.Degree) + st.Dim("  ("+e.Start+" – "+e.End+")") + "\n")
		line := st.Accent2(e.Institution)
		if e.Location != "" {
			line += st.Dim(" · " + e.Location)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func renderAchievements(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	var b strings.Builder
	b.WriteString(heading(st, "achievements"))
	for _, a := range ctx.Resume.Achievements {
		b.WriteString(markerLines(st.Warn("★ "), 2, w, a))
	}
	return b.String()
}

func renderTimeline(ctx *Context) string {
	st := ctx.Style
	var b strings.Builder
	b.WriteString(heading(st, "timeline"))
	for i, t := range ctx.Resume.Timeline {
		last := i == len(ctx.Resume.Timeline)-1
		b.WriteString(st.Accent2(padRight(t.Date, 10)) + st.Accent("● ") + st.Bold(t.Title) + "\n")
		connector := "│"
		if last {
			connector = " "
		}
		if t.Subtitle != "" {
			b.WriteString(strings.Repeat(" ", 10) + st.Dim(connector+"   "+t.Subtitle) + "\n")
		} else if !last {
			b.WriteString(strings.Repeat(" ", 10) + st.Dim(connector) + "\n")
		}
	}
	return b.String()
}

func renderStats(ctx *Context) string {
	st := ctx.Style
	r := ctx.Resume
	now := time.Now()
	rows := [][2]string{
		{"Experience", r.ExperienceLabel(now) + " (since Aug 2024)"},
		{"Technologies", fmt.Sprintf("%d", len(r.AllTechnologies()))},
		{"Certifications", fmt.Sprintf("%d", len(r.Certifications))},
		{"Projects", fmt.Sprintf("%d", len(r.Projects))},
		{"Employers", fmt.Sprintf("%d", r.Employers())},
	}
	if g := r.Contact.GitHub; g != nil {
		rows = append(rows, [2]string{"GitHub", st.Hyperlink(g.URL, st.Link(g.URL))})
	} else {
		rows = append(rows, [2]string{"GitHub repos", st.Warn("n/a — not listed on résumé")})
	}
	var b strings.Builder
	b.WriteString(heading(st, "stats"))
	for _, row := range rows {
		b.WriteString(st.Dim(padRight(row[0], 16)) + row[1] + "\n")
	}
	return b.String()
}

func renderContact(ctx *Context) string {
	st := ctx.Style
	c := ctx.Resume.Contact
	var b strings.Builder
	b.WriteString(heading(st, "contact"))
	b.WriteString(st.Dim(padRight("email", 11)) + st.Hyperlink("mailto:"+c.Email, st.Link(c.Email)) + "\n")
	if c.Phone != "" {
		b.WriteString(st.Dim(padRight("phone", 11)) + st.Hyperlink("tel:"+c.Phone, st.Link(c.Phone)) + "\n")
	}
	b.WriteString(st.Dim(padRight("location", 11)) + c.Location + "\n")
	if c.LinkedIn != nil {
		b.WriteString(st.Dim(padRight("linkedin", 11)) +
			st.Hyperlink(c.LinkedIn.URL, st.Link(c.LinkedIn.URL)) +
			st.Dim(" (@"+c.LinkedIn.Handle+")") + "\n")
	}
	if c.GitHub != nil {
		b.WriteString(st.Dim(padRight("github", 11)) + st.Hyperlink(c.GitHub.URL, st.Link(c.GitHub.URL)) + "\n")
	} else {
		b.WriteString(st.Dim(padRight("github", 11)) + st.Warn("not listed on résumé") + "\n")
	}
	return b.String()
}

func renderSkillsJSON(ctx *Context) string {
	b, err := json.MarshalIndent(ctx.Resume.Skills, "", "  ")
	if err != nil {
		return ctx.Style.Err("failed to encode skills\n")
	}
	return string(b) + "\n"
}

/* --------------------------------- filesystem --------------------------------- */

func catRenderers() map[string]func(*Context) string {
	return map[string]func(*Context) string{
		"about.txt":          renderAbout,
		"experience.txt":     renderExperience,
		"projects.txt":       renderProjects,
		"skills.txt":         renderSkills,
		"certifications.txt": renderCerts,
		"education.txt":      renderEducation,
		"achievements.txt":   renderAchievements,
		"timeline.txt":       renderTimeline,
		"contact.txt":        renderContact,
		"skills.json":        renderSkillsJSON,
	}
}

func renderLS(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	maxw := 0
	for _, f := range vfiles {
		if len(f) > maxw {
			maxw = len(f)
		}
	}
	col := maxw + 2
	cols := w / col
	if cols < 1 {
		cols = 1
	}
	var b strings.Builder
	for i, f := range vfiles {
		name := padRight(f, col)
		if strings.HasSuffix(f, ".pdf") {
			name = st.Err(padRight(f, col))
		} else {
			name = st.Accent(padRight(f, col))
		}
		b.WriteString(name)
		if (i+1)%cols == 0 {
			b.WriteString("\n")
		}
	}
	if len(vfiles)%cols != 0 {
		b.WriteString("\n")
	}
	return b.String()
}

func renderTree(ctx *Context) string {
	st := ctx.Style
	var b strings.Builder
	b.WriteString(st.Accent2("/home/"+ctx.Resume.Username) + "\n")
	for i, f := range vfiles {
		branch := "├── "
		if i == len(vfiles)-1 {
			branch = "└── "
		}
		name := st.Accent(f)
		if strings.HasSuffix(f, ".pdf") {
			name = st.Err(f)
		}
		b.WriteString(st.Dim(branch) + name + "\n")
	}
	b.WriteString(st.Dim(fmt.Sprintf("\n%d files", len(vfiles))) + "\n")
	return b.String()
}

/* ----------------------------------- search ----------------------------------- */

type searchHit struct{ section, text string }

func searchCorpus(r *resume.Resume) []searchHit {
	hits := []searchHit{{"summary", r.Summary}}
	for _, e := range r.Experience {
		hits = append(hits, searchHit{"experience · " + e.Company,
			fmt.Sprintf("%s at %s, %s (%s–%s)", e.Role, e.Company, e.Location, e.Start, e.End)})
		for _, h := range e.Highlights {
			hits = append(hits, searchHit{"experience · " + e.Company, h})
		}
	}
	for _, c := range r.Skills {
		for _, s := range c.Skills {
			hits = append(hits, searchHit{"skills · " + c.Name, s})
		}
	}
	for _, p := range r.Projects {
		hits = append(hits, searchHit{"project · " + p.Name,
			p.Name + " — " + p.Description + " [" + strings.Join(p.Tech, ", ") + "]"})
	}
	for _, c := range r.Certifications {
		t := c.Name
		if c.Issuer != "" {
			t += " — " + c.Issuer
		}
		hits = append(hits, searchHit{"certifications", t})
	}
	for _, e := range r.Education {
		hits = append(hits, searchHit{"education", e.Degree + " — " + e.Institution})
	}
	for _, a := range r.Achievements {
		hits = append(hits, searchHit{"achievements", a})
	}
	return hits
}

func highlightLine(st stStyle, line, query string) string {
	if query == "" {
		return line
	}
	lower, needle := strings.ToLower(line), strings.ToLower(query)
	var b strings.Builder
	for i := 0; i < len(line); {
		idx := strings.Index(lower[i:], needle)
		if idx < 0 {
			b.WriteString(line[i:])
			break
		}
		b.WriteString(line[i : i+idx])
		b.WriteString(st.Bold(st.Warn(line[i+idx : i+idx+len(needle)])))
		i += idx + len(needle)
	}
	return b.String()
}

func renderSearch(ctx *Context) string {
	st, w := ctx.Style, contentWidth(ctx.Width)
	query := strings.TrimSpace(strings.Join(ctx.Args, " "))
	if query == "" {
		return st.Warn("usage: ") + st.Accent("search <term>") + st.Dim("  — e.g. search kubernetes") + "\n"
	}
	safe := sanitize(query, 60)
	var matched []searchHit
	for _, h := range searchCorpus(ctx.Resume) {
		if strings.Contains(strings.ToLower(h.text), strings.ToLower(query)) {
			matched = append(matched, h)
		}
	}
	var b strings.Builder
	if len(matched) == 0 {
		return st.Warn("No matches for “"+safe+"”. Try: kubernetes, terraform, jenkins.") + "\n"
	}
	b.WriteString(st.Dim(fmt.Sprintf("%d match(es) for ", len(matched))) + st.Accent("“"+safe+"”") + "\n\n")
	for _, h := range matched {
		b.WriteString(st.Accent2(h.section) + "\n")
		for _, l := range wrapText(h.text, w) {
			b.WriteString("  " + highlightLine(st, l, query) + "\n")
		}
	}
	return b.String()
}

/* ---------------------------------- help ------------------------------------- */

func renderHelp(ctx *Context) string {
	st := ctx.Style
	reg := ctx.Shell.reg
	var b strings.Builder
	b.WriteString("Available commands — " + st.Accent("Tab") + " completes, " +
		st.Accent("↑/↓") + " recall history.\n\n")
	for _, cat := range CategoryOrder {
		cmds := reg.ByCategory(cat)
		if len(cmds) == 0 {
			continue
		}
		b.WriteString(st.Accent2(string(cat)) + "\n")
		for _, c := range cmds {
			b.WriteString("  " + st.Accent(padRight(c.Name, 15)) + st.Dim(c.Summary) + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(st.Dim("Hidden goodies exist. Try ") + st.Accent("sudo hire-me") +
		st.Dim(" or ") + st.Accent("matrix") + st.Dim(". 🥚") + "\n")
	return b.String()
}

/* ------------------------------ registry builder ------------------------------ */

// BuildRegistry constructs the full command set. cfg is captured by commands
// that need public URLs (resume/download/website).
func BuildRegistry(cfg config.Config) *Registry {
	reg := NewRegistry()
	simple := func(name string, aliases []string, cat Category, summary string, render func(*Context) string) {
		reg.Add(&Command{
			Name: name, Aliases: aliases, Category: cat, Summary: summary,
			Run: func(ctx *Context) error { out(ctx, render(ctx)); return nil },
		})
	}

	// INFO
	simple("about", nil, CatInfo, "Who I am — the professional summary", renderAbout)
	reg.Add(&Command{
		Name: "whoami", Category: CatInfo, Summary: "Print the current user",
		Run: func(ctx *Context) error {
			st := ctx.Style
			out(ctx, st.Accent(ctx.Resume.Username)+" — "+ctx.Resume.Name+", "+ctx.Resume.Title+".\n"+
				st.Dim("Type ")+st.Accent("about")+st.Dim(" for the full story.")+"\n")
			return nil
		},
	})

	// RÉSUMÉ
	simple("experience", []string{"exp", "work"}, CatResume, "Professional work history", renderExperience)
	simple("projects", nil, CatResume, "Key engineering initiatives", renderProjects)
	simple("skills", nil, CatResume, "Technical skills by category", renderSkills)
	simple("techstack", []string{"stack", "tech"}, CatResume, "Technologies grouped by category", renderTechstack)
	simple("certifications", []string{"certs", "cert"}, CatResume, "Professional certifications", renderCerts)
	simple("education", []string{"edu"}, CatResume, "Academic background", renderEducation)
	simple("achievements", []string{"awards"}, CatResume, "Quantified wins & highlights", renderAchievements)
	simple("timeline", nil, CatResume, "Career journey as an ASCII timeline", renderTimeline)
	simple("stats", nil, CatResume, "Snapshot dashboard of the numbers", renderStats)
	reg.Add(&Command{
		Name: "resume", Aliases: []string{"cv", "download"}, Category: CatResume,
		Summary: "Full résumé + PDF download link",
		Run: func(ctx *Context) error {
			st := ctx.Style
			out(ctx, renderAbout(ctx)+"\n"+renderExperience(ctx)+"\n"+renderSkills(ctx)+"\n"+
				renderCerts(ctx)+"\n"+renderEducation(ctx)+"\n")
			out(ctx, st.Dim(strings.Repeat("─", 50))+"\n")
			out(ctx, st.Accent("Download PDF : ")+st.Hyperlink(cfg.ResumeURL, st.Link(cfg.ResumeURL))+"\n")
			out(ctx, st.Accent("View online  : ")+st.Hyperlink(cfg.WebURL, st.Link(cfg.WebURL))+"\n")
			return nil
		},
	})

	// CONNECT
	simple("contact", nil, CatConnect, "All the ways to reach me", renderContact)
	reg.Add(&Command{
		Name: "email", Category: CatConnect, Summary: "Email address (mailto)",
		Run: func(ctx *Context) error {
			st := ctx.Style
			out(ctx, st.Hyperlink("mailto:"+ctx.Resume.Contact.Email, st.Link(ctx.Resume.Contact.Email))+"\n")
			return nil
		},
	})
	reg.Add(&Command{
		Name: "linkedin", Category: CatConnect, Summary: "LinkedIn profile",
		Run: func(ctx *Context) error {
			st := ctx.Style
			if l := ctx.Resume.Contact.LinkedIn; l != nil {
				out(ctx, st.Hyperlink(l.URL, st.Link(l.URL))+st.Dim(" (@"+l.Handle+")")+"\n")
			} else {
				out(ctx, st.Warn("LinkedIn is not listed on the résumé.")+"\n")
			}
			return nil
		},
	})
	reg.Add(&Command{
		Name: "github", Category: CatConnect, Summary: "GitHub profile",
		Run: func(ctx *Context) error {
			st := ctx.Style
			if g := ctx.Resume.Contact.GitHub; g != nil {
				out(ctx, st.Hyperlink(g.URL, st.Link(g.URL))+"\n")
			} else {
				out(ctx, st.Warn("GitHub is not listed on the résumé — nothing to link here yet.")+"\n")
			}
			return nil
		},
	})
	reg.Add(&Command{
		Name: "blog", Category: CatConnect, Summary: "Writing / blog",
		Run: func(ctx *Context) error {
			out(ctx, ctx.Style.Warn("No blog on the résumé (yet). Watch this space.")+"\n")
			return nil
		},
	})

	// SYSTEM
	reg.Add(&Command{
		Name: "help", Aliases: []string{"?", "commands"}, Category: CatSystem,
		Summary: "List everything you can do",
		Run:     func(ctx *Context) error { out(ctx, renderHelp(ctx)); return nil },
	})
	reg.Add(&Command{
		Name: "search", Aliases: []string{"grep", "find"}, Category: CatSystem,
		Summary: "Search the résumé, e.g. search kubernetes", Usage: "search <term>",
		Run: func(ctx *Context) error { out(ctx, renderSearch(ctx)); return nil },
	})
	reg.Add(&Command{
		Name: "ls", Category: CatSystem, Summary: "List résumé \"files\"",
		Run: func(ctx *Context) error { out(ctx, renderLS(ctx)); return nil },
	})
	reg.Add(&Command{
		Name: "tree", Category: CatSystem, Summary: "Show the résumé filesystem tree",
		Run: func(ctx *Context) error { out(ctx, renderTree(ctx)); return nil },
	})
	reg.Add(&Command{
		Name: "cat", Category: CatSystem, Summary: "Print a file, e.g. cat about.txt", Usage: "cat <file>",
		Run: func(ctx *Context) error {
			st := ctx.Style
			if len(ctx.Args) == 0 {
				out(ctx, st.Warn("usage: cat <file>")+st.Dim("  — try 'ls' to see files")+"\n")
				return nil
			}
			file := strings.ToLower(ctx.Args[0])
			if file == "resume.pdf" {
				out(ctx, st.Warn("resume.pdf: binary file. Run ")+st.Accent("resume")+st.Warn(" for the download link.")+"\n")
				return nil
			}
			if r, ok := catRenderers()[file]; ok {
				out(ctx, r(ctx))
				return nil
			}
			out(ctx, st.Err("cat: "+sanitize(file, 40)+": No such file")+st.Dim("  (try 'ls')")+"\n")
			return nil
		},
	})
	reg.Add(&Command{
		Name: "pwd", Category: CatSystem, Summary: "Print working directory",
		Run: func(ctx *Context) error { out(ctx, "/home/"+ctx.Resume.Username+"\n"); return nil },
	})
	reg.Add(&Command{
		Name: "echo", Category: CatSystem, Summary: "Echo text back",
		Run: func(ctx *Context) error { out(ctx, sanitize(strings.Join(ctx.Args, " "), 200)+"\n"); return nil },
	})
	reg.Add(&Command{
		Name: "date", Category: CatSystem, Summary: "Show the current date & time",
		Run: func(ctx *Context) error { out(ctx, time.Now().Format("Mon 02 Jan 2006 15:04:05 MST")+"\n"); return nil },
	})
	reg.Add(&Command{
		Name: "history", Category: CatSystem, Summary: "Show command history (history -c to clear)",
		Run: func(ctx *Context) error {
			st := ctx.Style
			if len(ctx.Args) > 0 && ctx.Args[0] == "-c" {
				ctx.Shell.history = nil
				out(ctx, st.Dim("History cleared.")+"\n")
				return nil
			}
			h := ctx.Shell.History()
			if len(h) == 0 {
				out(ctx, st.Dim("No history yet.")+"\n")
				return nil
			}
			var b strings.Builder
			for i, cmd := range h {
				b.WriteString(st.Dim(fmt.Sprintf("%4d  ", i+1)) + cmd + "\n")
			}
			out(ctx, b.String())
			return nil
		},
	})
	reg.Add(&Command{
		Name: "clear", Aliases: []string{"cls"}, Category: CatSystem, Summary: "Clear the screen",
		Run: func(ctx *Context) error { out(ctx, "\x1b[2J\x1b[3J\x1b[H"); return nil },
	})
	reg.Add(&Command{
		Name: "exit", Aliases: []string{"quit", "logout"}, Category: CatSystem, Summary: "End the session",
		Run: func(ctx *Context) error {
			st := ctx.Style
			out(ctx, st.Dim("Thanks for stopping by — ")+st.Accent(ctx.Resume.Name)+st.Dim(" will be in touch. 👋")+"\n")
			return ErrExit
		},
	})

	// FUN (defined in fun.go)
	addFunCommands(reg)
	return reg
}
