# Sivanesh B — Terminal Résumé (web + SSH)

One portfolio, **two synchronized frontends**, inspired by
[terminal.shop](https://terminal.shop):

| Access                              | What you get                                            |
| ----------------------------------- | ------------------------------------------------------- |
| 🌐 `https://mydomain.dev`           | Interactive **React** terminal in the browser           |
| 💻 `ssh mydomain.dev`               | A **real SSH** session — a full-screen, sandboxed TUI   |

Both are driven by a **single source of truth** — [`shared/resume.json`](shared/resume.json).
Edit it once and the website **and** the SSH experience update together. There is
no duplicated résumé data anywhere.

> The SSH experience is **not** a browser fake — it is an actual SSH server you
> connect to with any client (Terminal, iTerm2, Ghostty, Windows Terminal, …).

---

## ✨ Highlights

**Two renderers, one portfolio**
- The **browser** is a command-driven terminal; the **SSH** side is a
  full-screen keyboard-driven TUI. Same content, different medium.
- Sections everywhere: `about`, `experience`, `projects`, `skills`,
  `techstack`, `certifications`, `education`, `achievements`, `timeline`,
  `contact`, plus cross-section `search`. Honest about missing data
  (no invented blog).

**Website** (`src/`)
- React 18 + TypeScript + Vite + Tailwind + Framer Motion.
- Blinking cursor, ↑/↓ history, live suggestions, ⌘K palette, three themes +
  high-contrast, copy/download/print, animated welcome, fully accessible.
- Command history, Tab autocomplete, `sudo hire-me`, `neofetch`, `matrix`,
  `coffee`, `fortune`.

**SSH TUI** (`ssh/`)
- Go + the [Charm](https://charm.sh) stack: `wish` (SSH), `bubbletea`,
  `bubbles`, `lipgloss`. Launches straight into the portfolio — **no shell**,
  no prompt, no command parser.
- Keyboard navigation (↑↓ ←→ Enter Esc `/` `?` `q`), pager sections, splash
  animation, OSC 8 hyperlinks, responsive layout + resize, ANSI colours.
- Anonymous access (any username, no password), visitor stats, graceful
  shutdown, health checks, structured logging — and a hard security sandbox.

---

## 🗂️ Architecture

```
terminal-resume/
├── shared/
│   └── resume.json         ← SINGLE SOURCE OF TRUTH (both frontends read this)
├── src/                    ← React website (imports shared/resume.json)
├── ssh/                    ← Go SSH TUI (loads shared/resume.json at runtime)
│   ├── cmd/portfolio-ssh/  ← entrypoint
│   └── internal/           ← config, resume, session, tui (Bubble Tea app)
├── deploy/                 ← Caddyfile, nginx.conf, systemd unit
├── Dockerfile.web          ← build React → serve with Caddy (auto-HTTPS)
├── ssh/Dockerfile          ← build Go → minimal Alpine runtime
├── docker-compose.yml      ← website + SSH together
├── docs/DEPLOYMENT.md      ← full production guide
└── .github/workflows/      ← CI (lint/test/build) + Docker publish + Pages
```

**No duplicated data:** the website bundles `shared/resume.json` at build time;
the SSH server reads the very same file at runtime. Rendering code differs per
frontend (TSX vs Go) — the *content* lives in exactly one place.

---

## 🚀 Quick start

### Website (dev)

```bash
npm install
npm run dev        # http://localhost:5173
npm run build      # type-check + production build
npm run lint
```

### SSH TUI (dev)

```bash
cd ssh
RESUME_PATH=../shared/resume.json go run ./cmd/portfolio-ssh   # listens on :2222

# from another terminal — no username required
ssh -p 2222 localhost
```

```bash
cd ssh && go test -race ./...    # résumé loader, search, TUI layout/render
```

See [`ssh/README.md`](ssh/README.md) for full SSH docs.

### Both together (Docker)

```bash
cp .env.example .env       # set DOMAIN, SSH_PORT
docker compose up -d --build
# web: https://${DOMAIN}   ·   ssh sivanesh@${DOMAIN}
```

---

## 🔒 Security (SSH)

Visitors **never** get a real shell. The server:

- authenticates anonymously (SSH `none`) — the username only personalises the prompt;
- runs **only** its built-in interpreter — it never calls `os/exec` or a shell;
- denies port-forwarding, agent-forwarding and SFTP;
- exposes no filesystem or network from a session;
- sanitises all echoed input, applies idle/max timeouts, and isolates panics;
- (systemd) adds seccomp + namespace hardening.

---

## 🛠️ Customize

Edit [`shared/resume.json`](shared/resume.json) — name, title, summary, skills,
experience, projects, certs, education, timeline, contact. `username`/`host`
control the shell prompt (`<username>@<host>:~$`). Optional `contact.github` /
`contact.portfolio` light up the relevant commands automatically.

Replace `public/…Resume.pdf` (web) and `shared/resume.json`'s `resumeFile`, and
set `RESUME_URL`/`WEB_URL` for the SSH `resume`/`contact` commands.

---

## ☁️ Deploy

- **Combined (recommended):** VPS + `docker compose` behind Caddy (auto-HTTPS) —
  see [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md).
- **Website only:** Vercel (zero-config) or GitHub Pages
  (`.github/workflows/deploy.yml`).
- **SSH only, no Docker:** systemd unit at `deploy/portfolio-ssh.service`.

## 📄 License

MIT. Résumé content and PDF belong to Sivanesh B.
