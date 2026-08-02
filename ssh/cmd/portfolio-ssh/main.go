// Command portfolio-ssh serves the interactive résumé over SSH. Visitors get a
// fully sandboxed shell emulator — it never executes host OS commands, never
// grants filesystem/network access, and only renders the portfolio.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	"github.com/sivanesh/portfolio-ssh/internal/config"
	"github.com/sivanesh/portfolio-ssh/internal/resume"
	"github.com/sivanesh/portfolio-ssh/internal/session"
	"github.com/sivanesh/portfolio-ssh/internal/shell"
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

	signer, err := loadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		logger.Error("failed to prepare host key", "err", err)
		os.Exit(1)
	}

	store := session.NewStore(cfg.StatePath)
	registry := shell.BuildRegistry(cfg)

	var wg sync.WaitGroup
	srv := &gliderssh.Server{
		Addr:        cfg.SSHAddr,
		Version:     "portfolio-" + version.Version,
		IdleTimeout: cfg.IdleTimeout,
		MaxTimeout:  cfg.MaxTimeout,
		Handler:     makeHandler(&wg, res, store, cfg, registry, logger),
		// No PasswordHandler/PublicKeyHandler => NoClientAuth (anonymous access).
		// No forwarding or subsystem handlers => port-forwarding, agent
		// forwarding and SFTP are all denied by default. Fully sandboxed.
	}
	srv.AddHostKey(signer)

	healthSrv := startHealthServer(cfg.HealthAddr, store, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("ssh server listening",
			"addr", cfg.SSHAddr, "version", version.Version, "user_hint", cfg.DefaultUser)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, gliderssh.ErrServerClosed) {
			logger.Error("ssh server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down…")

	shutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutCtx) // stop accepting new connections
	_ = healthSrv.Shutdown(shutCtx)

	// Give active sessions a chance to finish, then force-close.
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		logger.Info("all sessions closed cleanly")
	case <-shutCtx.Done():
		logger.Warn("shutdown timeout — closing remaining sessions")
	}
	_ = srv.Close()
	logger.Info("bye")
}

// makeHandler builds the per-session handler. Each session is isolated and can
// only interact with the sandboxed shell interpreter.
func makeHandler(
	wg *sync.WaitGroup,
	res *resume.Resume,
	store *session.Store,
	cfg config.Config,
	registry *shell.Registry,
	logger *slog.Logger,
) gliderssh.Handler {
	return func(s gliderssh.Session) {
		wg.Add(1)
		defer wg.Done()

		// A panic in one session must never take down the server.
		defer func() {
			if r := recover(); r != nil {
				logger.Error("session panic recovered", "err", r, "user", s.User())
			}
		}()

		user := s.User()
		remote := s.RemoteAddr().String()
		store.IncActive()
		defer store.DecActive()
		prev, first := store.Visit(user)

		logger.Info("session started", "user", user, "remote", remote, "active", store.Active())
		start := time.Now()
		defer func() {
			logger.Info("session ended", "user", user, "remote", remote, "dur", time.Since(start).String())
		}()

		ptyReq, winCh, isPty := s.Pty()

		opts := shell.Options{
			RW:         s,
			Username:   user,
			Resume:     res,
			Store:      store,
			Config:     cfg,
			Registry:   registry,
			PrevLogin:  prev,
			FirstVisit: first,
		}

		// Non-interactive exec mode: `ssh host about` runs one command.
		if raw := strings.TrimSpace(s.RawCommand()); raw != "" {
			opts.Width, opts.Height, opts.Color = 100, 24, false
			shell.New(opts).RunOnce(raw)
			return
		}

		if !isPty {
			io.WriteString(s, "This is an interactive portfolio.\r\n"+
				"Connect with a terminal:  ssh "+sanitizeForNotice(user)+"@<host>\r\n"+
				"Or run one command:       ssh "+sanitizeForNotice(user)+"@<host> about\r\n")
			return
		}

		opts.Width, opts.Height = ptyReq.Window.Width, ptyReq.Window.Height
		opts.Color = ptyReq.Term != "" && ptyReq.Term != "dumb"

		sh := shell.New(opts)
		go func() {
			for win := range winCh {
				sh.Resize(win.Width, win.Height)
			}
		}()
		sh.Run()
	}
}

func sanitizeForNotice(u string) string {
	if u == "" {
		return "guest"
	}
	var b strings.Builder
	for _, r := range u {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
		if b.Len() >= 32 {
			break
		}
	}
	if b.Len() == 0 {
		return "guest"
	}
	return b.String()
}

// loadOrCreateHostKey loads a persisted ed25519 host key or generates one.
func loadOrCreateHostKey(path string) (gossh.Signer, error) {
	if b, err := os.ReadFile(path); err == nil {
		return gossh.ParsePrivateKey(b)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	block, err := gossh.MarshalPrivateKey(priv, "portfolio-ssh")
	if err != nil {
		return nil, err
	}
	if dir := filepath.Dir(path); dir != "" {
		_ = os.MkdirAll(dir, 0o700)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		// Non-fatal: fall back to an ephemeral in-memory key.
		slog.Warn("could not persist host key; using ephemeral key", "path", path, "err", err)
	}
	return gossh.NewSignerFromKey(priv)
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
