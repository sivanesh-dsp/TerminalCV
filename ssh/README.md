# Portfolio SSH TUI

A résumé served over **real SSH** as a full-screen **terminal UI** — inspired by
[terminal.shop](https://terminal.shop). This is **not** a shell. On connect the
visitor is dropped straight into an interactive portfolio application: no prompt,
no command parser, no path to the host OS.

```
ssh mydomain.dev          # no username required — any username is accepted
```

Built with the [Charm](https://charm.sh) stack:
[`wish`](https://github.com/charmbracelet/wish) (SSH),
[`bubbletea`](https://github.com/charmbracelet/bubbletea) (TUI runtime),
[`bubbles`](https://github.com/charmbracelet/bubbles) and
[`lipgloss`](https://github.com/charmbracelet/lipgloss) (styling).

## Shared data

The server reads the **same** `shared/resume.json` that the React website
consumes — there is no duplicated résumé content. Edit that one file and both
frontends update. It is loaded at runtime (`RESUME_PATH`, or auto-discovered).

```
                 shared/resume.json  (single source of truth)
                         │
          ┌──────────────┴──────────────┐
          ▼                             ▼
    React web UI                  Bubble Tea SSH TUI
    (browser)                     (terminal)
```

## Experience

On connect: a minimal splash, then a centered **master-detail browser** styled
after [terminal.shop](https://terminal.shop) — a tab bar, a grouped list with a
solid accent selection bar, and a live detail pane. It floats in the middle of
the screen (no full-screen frame). Everything is keyboard driven.

| Key            | Action                                               |
| -------------- | ---------------------------------------------------- |
| `↑ / ↓` `k/j`  | browse the list · scroll the detail pane             |
| `← / →` `Tab`  | switch tab (category)                                |
| `a e p s r c`  | jump straight to a tab                               |
| `Enter`        | focus the detail pane (then `↑/↓` scroll)            |
| `Esc`          | leave the detail pane                                |
| `/`            | search (matches every section)                       |
| `?`            | keyboard-shortcuts overlay                           |
| `q` `Ctrl+C`   | quit / disconnect                                    |

**Tabs:** about · experience · projects · skills · resume · contact. Experience
and projects list one item per role/project; selecting one shows its detail live
on the right. CONTACT renders **OSC 8 hyperlinks**, clickable in supporting
terminals. The layout is responsive: bordered tabs collapse to a compact line on
narrow terminals, and it reflows on resize (40×15 → large windows).

Colours come from a **per-session lipgloss renderer** whose profile is derived
from the client's `TERM`, so each visitor is styled for their own terminal.

## Develop

```bash
cd ssh

# run it locally on :2222 (host key + state land in ./data)
RESUME_PATH=../shared/resume.json go run ./cmd/portfolio-ssh

# from another terminal:
ssh -p 2222 localhost
```

> Tip: pass `-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no` while
> the host key changes during development.

## Test / vet / build

```bash
cd ssh
go vet ./...
go test -race ./...   # résumé loader, cross-section search, TUI layout/render
go build -o portfolio-ssh ./cmd/portfolio-ssh
```

## Configuration (environment variables)

| Variable        | Default              | Description                                   |
| --------------- | -------------------- | --------------------------------------------- |
| `SSH_ADDR`      | `:2222`              | SSH listen address                            |
| `HEALTH_ADDR`   | `:8081`              | HTTP health endpoint (`/healthz`, `/version`) |
| `HOST_KEY_PATH` | `data/host_ed25519`  | Persisted ed25519 host key (auto-generated)   |
| `STATE_PATH`    | `data/state.json`    | Visitor stats + last-login store              |
| `RESUME_PATH`   | *(auto-discovered)*  | Path to `shared/resume.json`                  |
| `IDLE_TIMEOUT`  | `5m`                 | Disconnect idle sessions                      |
| `MAX_TIMEOUT`   | `60m`                | Hard session length cap                       |

## Security model

- **Anonymous auth** via SSH `none` (no password, no key). Any username is
  accepted and used only as session metadata — it never grants authorization.
- **Not a shell:** the only interaction is the Bubble Tea program. The server
  never calls `os/exec`, spawns a subprocess, or evaluates user input as a
  command.
- **No forwarding / no SFTP:** local & reverse port-forwarding, agent forwarding
  and subsystems are all denied (no handlers registered).
- **PTY required:** non-interactive sessions (`ssh host somecommand`) are
  rejected by the `activeterm` middleware — there is no exec path.
- **No filesystem/network from a session:** the TUI only formats in-memory
  résumé data.
- **Timeouts** (idle + max) and **panic isolation** per session; the systemd
  unit adds seccomp + namespace hardening.

## Architecture

```
cmd/portfolio-ssh   entrypoint: wish server, host key, health, graceful shutdown
internal/
  config            env-driven configuration
  resume            data model + runtime loader + cross-section Search()
  session           persisted visitor stats + active-session gauge
  version           build metadata (ldflags)
  tui               the Bubble Tea application:
    model.go        root model, Init/Update, input handling, tabs/focus
    view.go         layout: splash · masthead · tab bar · list · detail · footer
    content.go      tab/item model + per-item detail renderers (from resume.json)
    theme.go        lipgloss palette built from the per-session renderer
    layout.go       wrapping, OSC 8 links, width/truncate helpers
```
