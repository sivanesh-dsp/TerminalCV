package shell

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/sivanesh/portfolio-ssh/internal/resume"
)

/* -------------------------------- neofetch -------------------------------- */

var neofetchLogo = []string{
	"    .-------.",
	"   /  o   o  \\",
	"  |    ^      |",
	"  |   '---'   |",
	"   \\_________/",
	"   sivanesh@os",
}

func skillCat(r *resume.Resume, name string) []string {
	for _, c := range r.Skills {
		if c.Name == name {
			return c.Skills
		}
	}
	return nil
}

func join(vals []string, max int) string {
	if max > 0 && len(vals) > max {
		vals = vals[:max]
	}
	return strings.Join(vals, " · ")
}

func renderNeofetch(ctx *Context) string {
	st := ctx.Style
	r := ctx.Resume
	info := [][2]string{
		{"OS", r.Title},
		{"Host", r.Name},
		{"Kernel", "kubernetes-1.x-cka"},
		{"Uptime", r.ExperienceLabel(time.Now()) + " in DevOps"},
		{"Shell", "zsh (bash-compatible)"},
		{"Cloud", join(skillCat(r, "Cloud Platforms"), 2)},
		{"Containers", join(skillCat(r, "Containerization & Orchestration"), 3)},
		{"CI/CD", join(skillCat(r, "CI/CD & GitOps"), 4)},
		{"IaC", join(skillCat(r, "Infrastructure as Code"), 3)},
		{"Observability", join(skillCat(r, "Observability & Monitoring"), 4)},
		{"Languages", join(skillCat(r, "Automation & Scripting"), 5)},
		{"Packages", fmt.Sprintf("%d technologies", len(r.AllTechnologies()))},
		{"Certs", "CKA · Terraform Associate"},
		{"Location", r.Contact.Location},
	}

	logoW := 0
	for _, l := range neofetchLogo {
		if len(l) > logoW {
			logoW = len(l)
		}
	}

	var b strings.Builder
	b.WriteString(st.Accent(r.Username) + "@" + st.Accent2(r.Host) + "\n")
	b.WriteString(st.Dim(strings.Repeat("─", len(r.Username)+len(r.Host)+1)) + "\n")

	rows := len(info)
	if len(neofetchLogo) > rows {
		rows = len(neofetchLogo)
	}
	for i := 0; i < rows; i++ {
		left := ""
		if i < len(neofetchLogo) {
			left = neofetchLogo[i]
		}
		b.WriteString(st.Accent(padRight(left, logoW+2)))
		if i < len(info) {
			b.WriteString(st.Accent(info[i][0]) + st.Dim(": ") + info[i][1])
		}
		b.WriteString("\n")
	}

	if st.Enabled {
		b.WriteString(strings.Repeat(" ", logoW+2))
		for _, c := range [][3]int{{94, 242, 176}, {122, 162, 247}, {255, 204, 102}, {255, 107, 107}, {125, 207, 255}, {230, 230, 230}} {
			b.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm   \x1b[0m", c[0], c[1], c[2]))
		}
		b.WriteString("\n")
	}
	return b.String()
}

/* --------------------------------- coffee --------------------------------- */

func coffee(ctx *Context) {
	st := ctx.Style
	cup := []string{
		"    ........",
		"    |      |]",
		"    \\      /",
		"     '----'",
	}
	steam := [][2]string{
		{"      ( (  ", "       ) ) "},
		{"       ) ) ", "      ( (  "},
	}
	const h = 6 // 2 steam + 4 cup lines
	for f := 0; f < 6; f++ {
		if f > 0 {
			out(ctx, fmt.Sprintf("\x1b[%dA", h))
		}
		var b strings.Builder
		s := steam[f%2]
		b.WriteString(st.Dim(s[0]) + "\x1b[K\n")
		b.WriteString(st.Dim(s[1]) + "\x1b[K\n")
		for _, l := range cup {
			b.WriteString(st.Accent(l) + "\x1b[K\n")
		}
		out(ctx, b.String())
		time.Sleep(220 * time.Millisecond)
	}
	out(ctx, "\n"+st.Bold("☕ Brewing…")+" a DevOps engineer’s true runtime.\n")
	out(ctx, st.Dim("Fun fact: uptime is powered by caffeine and YAML.")+"\n")
}

/* -------------------------------- fortune --------------------------------- */

var fortunes = []string{
	"It works on my cluster. — every engineer, before the incident.",
	"There are two hard problems in DevOps: cache invalidation, naming things, and off-by-one YAML indentation.",
	"The best time to write the runbook was before the outage. The second best time is now.",
	"Automate the boring, observe the rest, and sleep through the pager.",
	"Infrastructure as Code: because clicking in the console does not scale.",
	"A rollback a day keeps the incident review away.",
	"To err is human; to recover automatically is SRE.",
	"Immutable infrastructure: treat servers like cattle, not pets.",
	"“Have you tried kubectl describe?” — the answer to 80% of questions.",
	"Ship small, ship often, and let the pipeline carry the fear.",
}

func fortune(ctx *Context) {
	out(ctx, ctx.Style.Accent("✶ ")+fortunes[rand.Intn(len(fortunes))]+"\n")
}

/* --------------------------------- hire-me -------------------------------- */

func progressBar(ctx *Context, cells int) {
	st := ctx.Style
	for p := 0; p <= 100; p += 5 {
		filled := p * cells / 100
		bar := st.Accent(strings.Repeat("█", filled)) + st.Dim(strings.Repeat("░", cells-filled))
		out(ctx, fmt.Sprintf("\r%s %3d%%", bar, p))
		time.Sleep(35 * time.Millisecond)
	}
	out(ctx, "\n")
}

func hireMe(ctx *Context) {
	st := ctx.Style
	out(ctx, "Password: ")
	for i := 0; i < 8; i++ {
		out(ctx, st.Accent("*"))
		time.Sleep(90 * time.Millisecond)
	}
	out(ctx, "\n\n")
	time.Sleep(200 * time.Millisecond)
	out(ctx, st.Ok("Access Granted")+"\n\n")
	time.Sleep(250 * time.Millisecond)
	out(ctx, st.Bold("Congratulations!")+"\n\n")
	time.Sleep(200 * time.Millisecond)
	out(ctx, "You have successfully hired "+st.Accent(ctx.Resume.Name)+".\n\n")
	time.Sleep(200 * time.Millisecond)
	out(ctx, st.Dim("Initializing onboarding...")+"\n")
	time.Sleep(200 * time.Millisecond)
	progressBar(ctx, 22)
	out(ctx, "\n"+st.Accent2("→ Run ")+st.Accent("contact")+st.Accent2(" to make it official. 🚀")+"\n")
}

/* ---------------------------------- hack ---------------------------------- */

func hack(ctx *Context) {
	st := ctx.Style
	lines := []struct {
		text  string
		color func(string) string
	}{
		{"[+] Initializing exploit framework…", st.Ok},
		{"[+] Scanning 65 Kubernetes worker nodes… done", st.Ok},
		{"[+] Bypassing RBAC policies… token acquired", st.Warn},
		{"[+] Injecting into Terraform state… ok", st.Ok},
		{"[+] Escalating privileges via Argo Workflows…", st.Ok},
		{"[+] Exfiltrating production secrets… 100%", st.Ok},
		{"[!] JUST KIDDING 😄", st.Accent},
		{"[+] No systems were harmed. Type 'help' for real commands.", st.Dim},
	}
	for _, l := range lines {
		out(ctx, l.color(l.text)+"\n")
		time.Sleep(320 * time.Millisecond)
	}
}

/* ---------------------------------- sudo ---------------------------------- */

func sudoCmd(ctx *Context) error {
	st := ctx.Style
	target := strings.ToLower(strings.TrimSpace(strings.Join(ctx.Args, " ")))
	switch {
	case target == "hire-me" || target == "hireme":
		hireMe(ctx)
	case strings.HasPrefix(target, "rm -rf") || strings.Contains(target, "rm -rf /"):
		out(ctx, st.Err("Nice try. 🛡️  This filesystem is immutable (and version-controlled).")+"\n")
	case target == "":
		out(ctx, st.Warn("usage: ")+st.Accent("sudo hire-me")+"\n")
	default:
		out(ctx, st.Warn(ctx.Resume.Username+" is not in the sudoers file. This incident will be reported. 😏")+"\n")
	}
	return nil
}

/* --------------------------------- matrix --------------------------------- */

// playMatrix renders a full-screen "digital rain" until any key is pressed.
func (sh *Shell) playMatrix() {
	w, h := sh.size()
	if h < 3 {
		h = 24
	}
	io.WriteString(sh.out, "\x1b[?25l\x1b[2J")
	defer io.WriteString(sh.out, "\x1b[0m\x1b[2J\x1b[H\x1b[?25h")

	glyphs := []rune("アカサタナ0123456789ABCDEFｸﾂﾅ<>*/{}[]$#")
	drops := make([]int, w)
	for i := range drops {
		drops[i] = -rand.Intn(h)
	}

	stop := make(chan struct{})
	go func() {
		_, _, _ = sh.in.ReadRune() // any key (incl. Ctrl+C) exits
		close(stop)
	}()

	ticker := time.NewTicker(70 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			var b strings.Builder
			for x := 0; x < w; x++ {
				y := drops[x]
				if y >= 1 && y <= h {
					fmt.Fprintf(&b, "\x1b[%d;%dH\x1b[38;2;190;255;190m%c", y, x+1, glyphs[rand.Intn(len(glyphs))])
				}
				if y-1 >= 1 && y-1 <= h {
					fmt.Fprintf(&b, "\x1b[%d;%dH\x1b[38;2;40;140;60m%c", y-1, x+1, glyphs[rand.Intn(len(glyphs))])
				}
				if tail := y - 7; tail >= 1 && tail <= h {
					fmt.Fprintf(&b, "\x1b[%d;%dH ", tail, x+1)
				}
				drops[x]++
				if drops[x]-7 > h && rand.Float64() > 0.95 {
					drops[x] = 0
				}
			}
			io.WriteString(sh.out, b.String())
		}
	}
}

/* ------------------------------- registration ------------------------------ */

func addFunCommands(reg *Registry) {
	reg.Add(&Command{
		Name: "neofetch", Category: CatFun, Summary: "System info, résumé-style",
		Run: func(ctx *Context) error { out(ctx, renderNeofetch(ctx)); return nil },
	})
	reg.Add(&Command{
		Name: "coffee", Category: CatFun, Summary: "Brew a virtual coffee ☕",
		Run: func(ctx *Context) error { coffee(ctx); return nil },
	})
	reg.Add(&Command{
		Name: "fortune", Category: CatFun, Summary: "A random DevOps aphorism",
		Run: func(ctx *Context) error { fortune(ctx); return nil },
	})
	reg.Add(&Command{
		Name: "matrix", Category: CatFun, Summary: "Enter the matrix (any key exits)",
		Run: func(ctx *Context) error { ctx.Shell.playMatrix(); return nil },
	})
	reg.Add(&Command{
		Name: "sudo", Category: CatFun, Hidden: true, Summary: "Run with elevated privileges 😉",
		Usage: "sudo hire-me", Run: sudoCmd,
	})
	reg.Add(&Command{
		Name: "hack", Category: CatFun, Hidden: true, Summary: "Totally-legit hacking sequence",
		Run: func(ctx *Context) error { hack(ctx); return nil },
	})
}
