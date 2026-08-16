// Command portfolio-ssh serves the résumé as a full-screen TUI over SSH.
//
// It is NOT a shell. On connect, the visitor is dropped straight into a
// Bubble Tea portfolio application: there is no prompt, no command parser and
// no path to the host OS. The server never executes user input, never touches
// the filesystem on the user's behalf, and denies port/agent forwarding and
// SFTP. Any username is accepted anonymously as session metadata only.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
	"github.com/sivanesh/portfolio-ssh/internal/tui"
	"github.com/sivanesh/portfolio-ssh/internal/version"
)

func main() {
	// Logs go to stdout (so `docker compose logs ssh` works) and, when LOG_FILE
	// is set, are also appended to that file for durable on-disk persistence.
	var logWriter io.Writer = os.Stdout
	var logFile *os.File
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		f, err := os.OpenFile(lf, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			slog.New(slog.NewTextHandler(os.Stdout, nil)).
				Warn("could not open LOG_FILE; logging to stdout only", "path", lf, "err", err)
		} else {
			logFile = f
			logWriter = io.MultiWriter(os.Stdout, f)
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}
	logger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	res, err := resume.Load(cfg.ResumePath)
	if err != nil {
		logger.Error("failed to load résumé data", "err", err)
		os.Exit(1)
	}

	store := session.NewStore(cfg.StatePath)

	// Optional probe log: non-interactive scanner/bot connections (no PTY) are
	// recorded here — one line per source IP — for fail2ban to act on. Kept out
	// of the main log so it stays clean and stats aren't inflated by bots.
	var probeLog io.Writer
	if pl := os.Getenv("PROBE_LOG"); pl != "" {
		if f, err := os.OpenFile(pl, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
			logger.Warn("could not open PROBE_LOG; probes will not be recorded", "path", pl, "err", err)
		} else {
			probeLog = f
			defer f.Close()
		}
	}

	srv, err := wish.NewServer(
		wish.WithAddress(cfg.SSHAddr),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithIdleTimeout(cfg.IdleTimeout),
		wish.WithMaxTimeout(cfg.MaxTimeout),
		wish.WithVersion("portfolio-"+version.Version),
		// No auth handlers => anonymous "none" auth; any username is accepted.
		// No forwarding/SFTP handlers => those requests are denied. Sandboxed.
		wish.WithMiddleware(
			bm.Middleware(teaHandler(res, cfg, store)),
			activeterm.Middleware(), // require an interactive PTY
			sessionMiddleware(store, logger, probeLog),
		),
	)
	if err != nil {
		logger.Error("failed to create ssh server", "err", err)
		os.Exit(1)
	}

	healthSrv := startHealthServer(cfg.HealthAddr, store, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("ssh portfolio TUI listening",
			"addr", cfg.SSHAddr, "version", version.Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Error("ssh server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down…")

	shutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		logger.Warn("ssh shutdown", "err", err)
	}
	_ = healthSrv.Shutdown(shutCtx)
	logger.Info("bye")
}

// teaHandler builds the per-session Bubble Tea program. Each connection gets an
// isolated Model; the only possible interaction is the portfolio TUI.
func teaHandler(res *resume.Resume, cfg config.Config, store *session.Store) bm.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, _ := s.Pty()
		w, h := pty.Window.Width, pty.Window.Height
		// Build the renderer with a FORCED colour profile derived from the
		// client's TERM. This binds styling to the session output while
		// skipping termenv's interactive terminal queries (which would block
		// on clients that don't answer) and the Ascii downgrade MakeRenderer
		// applies when no colour-profile middleware is present.
		renderer := lipgloss.NewRenderer(s)
		renderer.SetColorProfile(colorProfile(pty.Term))
		renderer.SetHasDarkBackground(true)
		m := tui.New(res, cfg, store, renderer, s, pty.Term, s.User(), version.Version, w, h)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

// colorProfile picks a terminal colour profile from the client's TERM string.
// PTY is guaranteed (activeterm), so we default to 256 colours and upgrade to
// truecolor when advertised, downgrading only for explicitly dumb/mono terms.
func colorProfile(term string) termenv.Profile {
	t := strings.ToLower(term)
	switch {
	case t == "" || t == "dumb":
		return termenv.Ascii
	case strings.Contains(t, "truecolor") || strings.Contains(t, "24bit"):
		return termenv.TrueColor
	case strings.Contains(t, "256"):
		return termenv.ANSI256
	case strings.Contains(t, "color"):
		return termenv.ANSI256
	default:
		return termenv.ANSI256
	}
}

// sessionMiddleware records visitor stats and structured logs around each
// interactive session. Non-interactive connections (SSH scanners/bots that
// never request a PTY) are treated as noise: they are NOT counted as visitors
// and are kept out of the main log; when a probe log is configured, their
// source IP is recorded there for fail2ban. It never inspects or executes
// session input.
func sessionMiddleware(store *session.Store, logger *slog.Logger, probeLog io.Writer) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("session panic recovered", "err", r, "user", s.User())
				}
			}()

			remote := s.RemoteAddr().String()
			_, _, isPty := s.Pty()

			// No PTY => an automated probe, not a real visitor. Record its IP
			// for fail2ban (if enabled), then let activeterm reject it. Do not
			// count it or write it to the main log.
			if !isPty {
				if probeLog != nil {
					host, _, err := net.SplitHostPort(remote)
					if err != nil {
						host = remote
					}
					fmt.Fprintf(probeLog, "%s probe from %s\n",
						time.Now().UTC().Format(time.RFC3339), host)
				}
				next(s)
				return
			}

			store.IncActive()
			defer store.DecActive()
			store.Visit(s.User())

			start := time.Now()
			logger.Info("session started", "user", s.User(), "remote", remote, "active", store.Active())
			next(s)
			logger.Info("session ended", "user", s.User(), "remote", remote, "dur", time.Since(start).String())
		}
	}
}

// startHealthServer exposes liveness/readiness endpoints for Docker/K8s probes.
func startHealthServer(addr string, store *session.Store, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}
	mux.HandleFunc("/healthz", ok)
	mux.HandleFunc("/readyz", ok)
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"version":"`+version.Version+`","active":`+itoa(store.Active())+`}`)
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("health server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server error", "err", err)
		}
	}()
	return srv
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
