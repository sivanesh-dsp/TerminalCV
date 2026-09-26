# Sivanesh B — Terminal Résumé (web + SSH + API)

One portfolio, **three synchronized surfaces**, inspired by
[terminal.shop](https://terminal.shop):

| Access | What you get |
| --- | --- |
| 🌐 <https://sivaneshbalaji.online> | Interactive **React** terminal in the browser |
| 💻 `ssh sivaneshbalaji.online` | A **real SSH** session — a full-screen, sandboxed TUI |
| 🔌 <https://sivaneshbalaji.online/api/v1/> | A read-only **REST API** (FastAPI) over the same résumé |
| 📚 <https://sivaneshbalaji.online/api/docs> | OpenAPI docs · [ReDoc](https://sivaneshbalaji.online/api/redoc) · [openapi.json](https://sivaneshbalaji.online/api/openapi.json) |

```bash
curl https://sivaneshbalaji.online/api/v1/profile
curl https://sivaneshbalaji.online/api/v1/skills
curl https://sivaneshbalaji.online/api/v1/projects
curl "https://sivaneshbalaji.online/api/v1/search?q=kubernetes"

curl -s https://sivaneshbalaji.online/api/v1/profile | jq
```

All three are driven by a **single source of truth** — [`shared/resume.json`](shared/resume.json).
Edit it once and the website, the SSH experience **and** the API update together.
There is no duplicated résumé data anywhere.

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
- A terminal.shop-style **master-detail browser**: minimal splash, a bordered
  tab bar, a grouped list with a solid accent selection bar, and a live detail
  pane — floating centered (not full-screen). Keyboard driven (↑↓ ←→ Enter Esc
  `a e p s r c` `/` `?` `q`), OSC 8 hyperlinks, responsive + resize, per-session
  colours from the client's `TERM`.
- Anonymous access (any username, no password), visitor stats, graceful
  shutdown, health checks, structured logging — and a hard security sandbox.

---

## 🗂️ Architecture

```text
                              shared/resume.json
                            SINGLE SOURCE OF TRUTH
                                      │
            ┌─────────────────────────┼─────────────────────────┐
            │ bundled at build time   │ mounted read-only       │ loaded at runtime
            ▼                         ▼                         ▼
    ┌───────────────┐         ┌───────────────┐         ┌───────────────┐
    │ Web portfolio │  GET    │  Résumé API   │         │ SSH terminal  │
    │ React + Vite  │────────►│    FastAPI    │         │      Go       │
    └───────────────┘/api/v1/*└───────┬───────┘         └───────────────┘
                                      │ REST (versioned, typed, read-only)
                                      ▼
                              ┌───────────────┐
                              │  Future MCP   │──► ChatGPT · Claude · agents
                              │    server     │
                              └───────────────┘
```

```text
terminal-resume/
├── shared/
│   └── resume.json         ← SINGLE SOURCE OF TRUTH (all three surfaces read this)
├── src/                    ← React website (bundles it; hydrates from /api/v1 at runtime)
│   └── services/resumeApi.ts ← REST client with bundled-data fallback
├── ssh/                    ← Go SSH TUI (loads shared/resume.json at runtime)
│   ├── cmd/portfolio-ssh/  ← entrypoint
│   └── internal/           ← config, resume, session, tui (Bubble Tea app)
├── deploy/                 ← Caddyfile (/, /api/*), nginx.conf, systemd unit
├── Dockerfile.web          ← build React → serve with Caddy (auto-HTTPS)
├── ssh/Dockerfile          ← build Go → minimal Alpine runtime
├── docker-compose.yml      ← website + SSH + API together
├── docs/DEPLOYMENT.md      ← full production guide
└── .github/workflows/      ← CI (lint/test/build) + Docker publish + Pages

../resume-api/              ← the REST API service (separate repository)
```

**No duplicated data:** the website bundles `shared/resume.json` at build time
*and* hydrates from the API at runtime; the SSH server reads the very same file
at runtime; the API mounts it read-only. Rendering code differs per surface
(TSX / Go / Pydantic) — the *content* lives in exactly one place.

**Why the SSH server does not call the API:** it already loads the same file in
the same deployment, so an HTTP hop would add a failure mode and buy nothing.
The browser is different — fetching `/api/v1/*` decouples the deployed bundle
from the data, so résumé edits appear without a site rebuild.

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

### REST API (dev)

The API lives in the sibling repository [`../resume-api`](../resume-api):

```bash
cd ../resume-api
make install
make dev           # http://127.0.0.1:8000/api/docs — reads ../terminal-resume/shared/resume.json
```

`npm run dev` proxies `/api` to `http://127.0.0.1:8000`, so the website hydrates
from the local API automatically. Without it the site silently falls back to the
bundled résumé — nothing breaks.

Type `api` in the terminal (browser) to see the endpoints, docs links and the
live API status.

### All three together (Docker)

```bash
cp .env.example .env       # set DOMAIN, SSH_PORT
docker compose up -d --build
# web: https://${DOMAIN}   ·   ssh sivanesh@${DOMAIN}   ·   API: https://${DOMAIN}/api/v1
```

Caddy routes `/api/*` to the API container; the API is never published to the
host. The compose file mounts `shared/resume.json` into it read-only.

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

- **Combined (recommended):** VPS + `docker compose` behind Caddy (auto-HTTPS),
  serving the site on `/`, the API on `/api/*` and SSH on `:22` —
  see [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md).
- **Website only:** Vercel (zero-config) or GitHub Pages
  (`.github/workflows/deploy.yml`).
- **SSH only, no Docker:** systemd unit at `deploy/portfolio-ssh.service`.

## 📄 License

MIT. Résumé content and PDF belong to Sivanesh B.
