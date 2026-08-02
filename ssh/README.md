# Portfolio SSH Server

A fully sandboxed, interactive résumé served over **real SSH**. Visitors connect
with any SSH client and get a Linux-like shell that renders the portfolio — it
**never** executes host OS commands, touches the filesystem, or opens the
network from a session.

```
ssh sivanesh@mydomain.dev      # any username works — it just personalises the prompt
```

Built with [`gliderlabs/ssh`](https://github.com/gliderlabs/ssh) and a custom
raw-mode line editor (history, arrows, Tab completion, Ctrl-shortcuts, resize).

## Shared data

The server reads the **same** `shared/resume.json` that the React website
consumes — there is no duplicated résumé content. Edit that one file and both
frontends update. The file is loaded at runtime (`RESUME_PATH`, or auto-discovered).

## Develop

```bash
cd ssh

# run it locally on :2222 (host key + state land in ./data)
RESUME_PATH=../shared/resume.json go run ./cmd/portfolio-ssh

# from another terminal:
ssh -p 2222 sivanesh@localhost            # interactive
ssh -p 2222 sivanesh@localhost about      # one-shot (exec mode)
```

> Tip: pass `-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no` while
> the host key changes during development.

## Test / vet / build

```bash
cd ssh
go vet ./...
go test ./...        # parser, command rendering, résumé loading, autocomplete
go build -o portfolio-ssh ./cmd/portfolio-ssh
```

## Configuration (environment variables)

| Variable        | Default                       | Description                                   |
| --------------- | ----------------------------- | --------------------------------------------- |
| `SSH_ADDR`      | `:2222`                       | SSH listen address                            |
| `HEALTH_ADDR`   | `:8081`                       | HTTP health endpoint (`/healthz`, `/version`) |
| `HOST_KEY_PATH` | `data/host_ed25519`           | Persisted ed25519 host key (auto-generated)   |
| `STATE_PATH`    | `data/state.json`             | Visitor stats + last-login store              |
| `RESUME_PATH`   | *(auto-discovered)*           | Path to `shared/resume.json`                  |
| `WEB_URL`       | `https://mydomain.dev`        | Website URL shown in commands                 |
| `RESUME_URL`    | `https://mydomain.dev/...pdf` | PDF link shown by `resume` / `download`       |
| `IDLE_TIMEOUT`  | `5m`                          | Disconnect idle sessions                      |
| `MAX_TIMEOUT`   | `60m`                         | Hard session length cap                       |
| `DEFAULT_USER`  | `sivanesh`                    | Prompt username when none is supplied         |

## Commands

`help about whoami experience projects skills education certifications
achievements techstack timeline contact resume github linkedin blog search
stats history clear exit pwd ls cat tree echo date` — plus hidden fun:
`neofetch matrix coffee fortune sudo hire-me hack`.

## Security model

- **Anonymous auth** via SSH `none` (no password, no key) — identifies the
  session only. Configurable username, any value accepted.
- **No shell escape:** only the built-in command interpreter runs; the process
  never calls `os/exec` or a real shell.
- **No forwarding / no SFTP:** local & reverse port-forwarding, agent
  forwarding and subsystems are all denied (no handlers registered).
- **No filesystem/network from a session:** commands only format in-memory data.
- **Input sanitisation:** any echoed user input (search, echo, unknown command,
  username) is stripped of control/escape bytes.
- **Timeouts** (idle + max) and **panic isolation** per session; systemd unit
  adds seccomp + namespace hardening.

## Architecture

```
cmd/portfolio-ssh   entrypoint: config, host key, server, health, graceful shutdown
internal/
  config            env-driven configuration
  resume            data model + runtime loader for shared/resume.json
  ansi              truecolor styling, OSC 8 hyperlinks, width helpers
  session           persisted visitor stats + active-session gauge
  shell             the sandboxed interpreter:
    shell.go        session loop, welcome/MOTD, prompt, dispatch
    editor.go       raw-mode line editor (history, arrows, Tab, Ctrl-keys)
    parser.go       shell-like tokenizer (quotes + escapes)
    registry.go     command registry
    commands.go     all résumé/system commands + renderers
    fun.go          neofetch, matrix, coffee, hire-me, fortune, hack
```
