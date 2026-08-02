package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sivanesh/portfolio-ssh/internal/ansi"
	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
	"github.com/sivanesh/portfolio-ssh/internal/version"
)

// Options configures a new interactive Shell.
type Options struct {
	RW         io.ReadWriter
	Width      int
	Height     int
	Color      bool
	Username   string
	Resume     *resume.Resume
	Store      *session.Store
	Config     config.Config
	Registry   *Registry
	PrevLogin  time.Time
	FirstVisit bool
}

// Shell is one interactive SSH session: a sandboxed command interpreter.
type Shell struct {
	in    *bufio.Reader
	out   io.Writer
	style *ansi.Style
	reg   *Registry
	res   *resume.Resume
	store *session.Store
	cfg   config.Config

	username string
	color    bool

	mu            sync.Mutex
	width, height int

	history   []string
	histIdx   int
	histDraft string

	prevLogin  time.Time
	firstVisit bool
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// New builds a Shell from Options.
func New(o Options) *Shell {
	w := o.Width
	if w <= 0 {
		w = 80
	}
	h := o.Height
	if h <= 0 {
		h = 24
	}
	return &Shell{
		in:         bufio.NewReader(o.RW),
		out:        &crlfWriter{w: o.RW},
		style:      ansi.New(o.Color),
		reg:        o.Registry,
		res:        o.Resume,
		store:      o.Store,
		cfg:        o.Config,
		username:   sanitizeUser(o.Username, o.Config.DefaultUser),
		color:      o.Color,
		width:      clamp(w, 20, 400),
		height:     clamp(h, 5, 200),
		histIdx:    0,
		prevLogin:  o.PrevLogin,
		firstVisit: o.FirstVisit,
	}
}

// Resize updates the terminal dimensions (called from the SSH window-change loop).
func (sh *Shell) Resize(w, h int) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if w > 0 {
		sh.width = clamp(w, 20, 400)
	}
	if h > 0 {
		sh.height = clamp(h, 5, 200)
	}
}

func (sh *Shell) w() int {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	return sh.width
}

// size returns the current terminal width and height (thread-safe).
func (sh *Shell) size() (int, int) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	return sh.width, sh.height
}

func (sh *Shell) writeln(s string) { io.WriteString(sh.out, s+"\n") }
func (sh *Shell) write(s string)   { io.WriteString(sh.out, s) }

func (sh *Shell) prompt() string {
	return sh.style.Accent(sh.username) +
		sh.style.Dim("@") +
		sh.style.Accent2(sh.res.Host) +
		sh.style.Dim(":") +
		sh.style.Link("~") +
		sh.style.Dim("$ ")
}

// Run drives the interactive read-eval loop until the session ends.
func (sh *Shell) Run() {
	sh.printWelcome()
	for {
		line, err := sh.readLine(sh.prompt())
		if err != nil {
			if errors.Is(err, io.EOF) {
				sh.writeln("")
				sh.writeln(sh.style.Dim("logout"))
			}
			return
		}
		args := Tokenize(line)
		if len(args) == 0 {
			continue
		}
		sh.pushHistory(line)
		if err := sh.dispatch(args, line); err != nil {
			if errors.Is(err, ErrExit) {
				return
			}
			sh.writeln(sh.style.Err("error: " + err.Error()))
		}
	}
}

// RunOnce executes a single command line non-interactively (SSH exec mode:
// `ssh host about`). It never grants an interactive shell.
func (sh *Shell) RunOnce(line string) {
	args := Tokenize(line)
	if len(args) == 0 {
		return
	}
	_ = sh.dispatch(args, line)
}

func (sh *Shell) dispatch(args []string, raw string) error {
	cmd := sh.reg.Get(args[0])
	if cmd == nil {
		sh.writeln(sh.style.Err("command not found: "+sanitize(args[0], 40)) +
			sh.style.Dim("  — type 'help' for the list"))
		return nil
	}
	ctx := &Context{
		Args:   args[1:],
		Raw:    raw,
		Shell:  sh,
		Out:    sh.out,
		Style:  sh.style,
		Resume: sh.res,
		Width:  sh.w(),
	}
	return cmd.Run(ctx)
}

func (sh *Shell) pushHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if n := len(sh.history); n == 0 || sh.history[n-1] != line {
		sh.history = append(sh.history, line)
		if len(sh.history) > 500 {
			sh.history = sh.history[len(sh.history)-500:]
		}
	}
	sh.histIdx = len(sh.history)
	sh.histDraft = ""
}

// History returns a copy of the session command history.
func (sh *Shell) History() []string {
	out := make([]string, len(sh.history))
	copy(out, sh.history)
	return out
}

func (sh *Shell) printWelcome() {
	width := sh.w()
	sh.writeln("")
	for _, line := range strings.Split(bannerFor(width), "\n") {
		sh.writeln(sh.style.Accent(line))
	}
	sh.writeln("")
	sh.writeln(sh.style.Bold("Welcome to " + sh.res.Name + "'s Interactive Resume"))
	sh.writeln(sh.style.Dim(sh.res.Title))
	sh.writeln("Type " + sh.style.Accent("help") + " to begin.")
	sh.writeln("")

	// Session info line(s).
	deployed := "dev build"
	if version.BuildTime != "" {
		deployed = version.BuildTime
	}
	sh.writeln(sh.style.Dim("  portfolio ") + sh.style.Accent2("v"+version.Version) +
		sh.style.Dim(" · deployed "+deployed))
	sh.writeln(sh.style.Dim(fmt.Sprintf("  active sessions: %d · total visits: %d",
		sh.store.Active(), sh.store.Total())))

	last := "Never"
	if !sh.firstVisit && !sh.prevLogin.IsZero() {
		last = sh.prevLogin.Format("Mon 02 Jan 2006 15:04 MST")
	}
	sh.writeln(sh.style.Dim("  last login: ") + sh.style.Link(last))
	sh.writeln("")
	sh.writeln(sh.style.Accent("  tip: ") + sh.style.Dim(sh.randomTip()))
	sh.writeln("")
}

var tips = []string{
	"Try 'neofetch' for a developer-profile system readout.",
	"Run 'search kubernetes' to grep every résumé section.",
	"'timeline' draws my career as an ASCII timeline.",
	"Psst… try 'sudo hire-me'. 🥚",
	"'theme' isn't here, but 'matrix' is. Give it a go.",
	"Use ↑/↓ for history and Tab to autocomplete commands.",
	"'contact' prints clickable links in supporting terminals.",
	"Ctrl+L clears the screen, Ctrl+C cancels a line.",
	"'stats' shows the numbers behind the résumé.",
	"Also on the web: open the same portfolio in a browser.",
}

func (sh *Shell) randomTip() string {
	return tips[rand.Intn(len(tips))]
}

// sanitize strips control/escape runes from untrusted text and caps length,
// preventing echoed input from injecting escape sequences into the client.
func sanitize(s string, max int) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\t' {
			b.WriteRune(' ')
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
		if b.Len() >= max {
			break
		}
	}
	return b.String()
}

func sanitizeUser(u, fallback string) string {
	u = sanitize(strings.TrimSpace(u), 32)
	// Keep it shell-like: letters, digits, dash, underscore, dot.
	u = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		default:
			return -1
		}
	}, u)
	if u == "" {
		return fallback
	}
	return u
}

// crlfWriter converts lone "\n" into "\r\n" so output is correct on a raw PTY,
// without doubling carriage returns for existing "\r\n" pairs.
type crlfWriter struct {
	w    io.Writer
	last byte
}

func (c *crlfWriter) Write(p []byte) (int, error) {
	var buf []byte
	for _, b := range p {
		if b == '\n' && c.last != '\r' {
			buf = append(buf, '\r')
		}
		buf = append(buf, b)
		c.last = b
	}
	if _, err := c.w.Write(buf); err != nil {
		return 0, err
	}
	return len(p), nil
}
