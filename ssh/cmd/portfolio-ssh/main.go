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
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	bm "github.com/charmbracelet/wish/bubbletea"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
	"github.com/sivanesh/portfolio-ssh/internal/tui"
	"github.com/sivanesh/portfolio-ssh/internal/version"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	res, err := resume.Load(cfg.ResumePath)
	if err != nil {
		logger.Error("failed to load résumé data", "err", err)
		os.Exit(1)
	}

	store := session.NewStore(cfg.StatePath)

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
			sessionMiddleware(store, logger),
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
		m := tui.New(res, cfg, store, s.User(), version.Version, w, h)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

// sessionMiddleware records visitor stats and structured logs around each
// session without ever inspecting or executing session input.
func sessionMiddleware(store *session.Store, logger *slog.Logger) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("session panic recovered", "err", r, "user", s.User())
				}
			}()
			store.IncActive()
			defer store.DecActive()
			store.Visit(s.User())

			remote := s.RemoteAddr().String()
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
