package shell

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
)

const testResumePath = "../../../shared/resume.json"

// rwBuf is an io.ReadWriter whose reads return EOF (no interactive input) and
// whose writes accumulate into a buffer we can assert against.
type rwBuf struct{ b *bytes.Buffer }

func (r *rwBuf) Read([]byte) (int, error) { return 0, io.EOF }
func (r *rwBuf) Write(p []byte) (int, error) {
	return r.b.Write(p)
}

func newTestShell(t *testing.T) (*Shell, *bytes.Buffer) {
	t.Helper()
	res, err := resume.Load(testResumePath)
	if err != nil {
		t.Fatalf("load resume: %v", err)
	}
	cfg := config.Config{
		WebURL:      "https://example.dev",
		ResumeURL:   "https://example.dev/resume.pdf",
		DefaultUser: "sivanesh",
	}
	buf := &bytes.Buffer{}
	sh := New(Options{
		RW:       &rwBuf{b: buf},
		Width:    100,
		Height:   24,
		Color:    false,
		Username: "guest",
		Resume:   res,
		Store:    session.NewStore(""),
		Config:   cfg,
		Registry: BuildRegistry(cfg),
	})
	return sh, buf
}

// run executes a command line and returns everything it printed.
func run(t *testing.T, sh *Shell, buf *bytes.Buffer, line string) string {
	t.Helper()
	buf.Reset()
	sh.RunOnce(line)
	return buf.String()
}

func TestContentCommandsRender(t *testing.T) {
	sh, buf := newTestShell(t)
	cases := map[string]string{
		"about":          "ABOUT",
		"whoami":         "sivanesh",
		"experience":     "Cprime",
		"projects":       "Jenkins",
		"skills":         "Kubernetes",
		"techstack":      "technologies",
		"certifications": "Kubernetes Administrator",
		"education":      "MCA",
		"achievements":   "Jenkins",
		"timeline":       "MCA",
		"stats":          "Technologies",
		"contact":        "sivaneshsiva240@gmail.com",
		"help":           "SYSTEM",
		"resume":         "Download",
		"ls":             "about.txt",
		"tree":           "resume.pdf",
		"pwd":            "/home/sivanesh",
		"neofetch":       "Cloud",
	}
	for cmd, want := range cases {
		out := run(t, sh, buf, cmd)
		if strings.TrimSpace(out) == "" {
			t.Errorf("%q produced no output", cmd)
		}
		if !strings.Contains(out, want) {
			t.Errorf("%q output missing %q; got:\n%s", cmd, want, out)
		}
	}
}

func TestCatFiles(t *testing.T) {
	sh, buf := newTestShell(t)
	for _, f := range vfiles {
		out := run(t, sh, buf, "cat "+f)
		if strings.TrimSpace(out) == "" {
			t.Errorf("cat %s produced no output", f)
		}
	}
	// skills.json should be valid-looking JSON.
	if out := run(t, sh, buf, "cat skills.json"); !strings.Contains(out, "\"skills\"") {
		t.Errorf("cat skills.json not JSON-like: %s", out)
	}
	// unknown file errors.
	if out := run(t, sh, buf, "cat nope.txt"); !strings.Contains(out, "No such file") {
		t.Errorf("cat nope.txt should error, got: %s", out)
	}
}

func TestSearch(t *testing.T) {
	sh, buf := newTestShell(t)
	if out := run(t, sh, buf, "search kubernetes"); !strings.Contains(strings.ToLower(out), "match") {
		t.Errorf("search kubernetes should report matches: %s", out)
	}
	if out := run(t, sh, buf, "search zzzznotfound"); !strings.Contains(out, "No matches") {
		t.Errorf("search miss should say no matches: %s", out)
	}
	if out := run(t, sh, buf, "search"); !strings.Contains(out, "usage") {
		t.Errorf("bare search should show usage: %s", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	sh, buf := newTestShell(t)
	if out := run(t, sh, buf, "definitelynotacommand"); !strings.Contains(out, "command not found") {
		t.Errorf("expected command-not-found, got: %s", out)
	}
}

func TestExitReturnsErrExit(t *testing.T) {
	sh, _ := newTestShell(t)
	if err := sh.dispatch([]string{"exit"}, "exit"); !errors.Is(err, ErrExit) {
		t.Errorf("exit should return ErrExit, got %v", err)
	}
}

func TestSudoHireMeMentionsName(t *testing.T) {
	sh, buf := newTestShell(t)
	// hire-me animation is short; assert it completes and greets by name.
	out := run(t, sh, buf, "sudo hire-me")
	if !strings.Contains(out, "Access Granted") || !strings.Contains(out, "Sivanesh") {
		t.Errorf("sudo hire-me output unexpected: %s", out)
	}
}

func TestAutocomplete(t *testing.T) {
	sh, _ := newTestShell(t)
	if line, _ := sh.autocomplete("ski"); strings.TrimSpace(line) != "skills" {
		t.Errorf("autocomplete(ski) = %q, want skills", line)
	}
	if line, _ := sh.autocomplete("cat ab"); strings.TrimSpace(line) != "cat about.txt" {
		t.Errorf("autocomplete(cat ab) = %q, want cat about.txt", line)
	}
	if _, cands := sh.autocomplete("e"); len(cands) < 2 {
		t.Errorf("autocomplete(e) should be ambiguous, got %v", cands)
	}
}
