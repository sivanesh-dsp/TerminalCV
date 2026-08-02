package shell

import (
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/sivanesh/portfolio-ssh/internal/ansi"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
)

// ErrExit is returned by a command to end the interactive session.
var ErrExit = errors.New("exit")

type Category string

const (
	CatInfo    Category = "INFO"
	CatResume  Category = "RÉSUMÉ"
	CatConnect Category = "CONNECT"
	CatSystem  Category = "SYSTEM"
	CatFun     Category = "FUN"
)

// CategoryOrder controls how `help` groups commands.
var CategoryOrder = []Category{CatInfo, CatResume, CatConnect, CatSystem, CatFun}

// Context carries everything a command needs to render its output.
type Context struct {
	Args   []string
	Raw    string
	Shell  *Shell
	Out    io.Writer
	Style  *ansi.Style
	Resume *resume.Resume
	Width  int
}

// Command is a single shell command. Run writes to ctx.Out and may return
// ErrExit to terminate the session.
type Command struct {
	Name     string
	Aliases  []string
	Summary  string
	Usage    string
	Category Category
	Hidden   bool
	Run      func(ctx *Context) error
}

// Registry holds all commands and resolves names + aliases.
type Registry struct {
	order  []*Command
	byName map[string]*Command
}

func NewRegistry() *Registry {
	return &Registry{byName: map[string]*Command{}}
}

func (r *Registry) Add(c *Command) {
	r.order = append(r.order, c)
	r.byName[strings.ToLower(c.Name)] = c
	for _, a := range c.Aliases {
		r.byName[strings.ToLower(a)] = c
	}
}

func (r *Registry) Get(name string) *Command {
	return r.byName[strings.ToLower(strings.TrimSpace(name))]
}

func (r *Registry) All() []*Command { return r.order }

// Names returns primary command names, sorted (used for Tab completion).
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.order))
	for _, c := range r.order {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	return names
}

// ByCategory returns non-hidden commands in a category, sorted by name.
func (r *Registry) ByCategory(cat Category) []*Command {
	var out []*Command
	for _, c := range r.order {
		if c.Category == cat && !c.Hidden {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
