# Deployment guide

One portfolio, two synchronized frontends:

- **Website** — `https://mydomain.dev` (static React site, served by Caddy with automatic HTTPS)
- **SSH** — `ssh sivanesh@mydomain.dev` (Go server, fully sandboxed)

Both read the single `shared/resume.json`, so updating the résumé updates both.

---

## 1. Prerequisites

- A Linux VPS with a public IP (any provider).
- A domain with DNS control.
- Docker + Docker Compose **or** a Go toolchain (for the systemd path).

## 2. DNS

Point your domain at the server:

| Type | Name           | Value            |
| ---- | -------------- | ---------------- |
| A    | `mydomain.dev` | `<server-ip>`    |
| AAAA | `mydomain.dev` | `<server-ipv6>`  |

SSH uses the **same hostname** — no extra record needed.

## 3. Free up port 22 for the portfolio (recommended)

Visitors expect `ssh sivanesh@mydomain.dev` (port 22). Move the host's own
administrative SSH daemon to another port first, so you don't lock yourself out:

```bash
sudo sed -i 's/^#\?Port 22/Port 2200/' /etc/ssh/sshd_config
sudo systemctl restart ssh          # now administer via: ssh -p 2200 you@server
```

Open the firewall:

```bash
sudo ufw allow 80,443/tcp           # website
sudo ufw allow 22/tcp               # portfolio SSH
sudo ufw allow 2200/tcp             # your admin SSH
```

---

## 4. Deploy with Docker Compose (recommended)

```bash
git clone <this-repo> portfolio && cd portfolio
cp .env.example .env
```

Edit `.env`:

```dotenv
DOMAIN=mydomain.dev
SSH_PORT=22
VERSION=1.0.0
```

Bring it up:

```bash
docker compose up -d --build
```

That's it. Caddy fetches a Let's Encrypt certificate automatically.

- Website: `https://mydomain.dev`
- SSH: `ssh sivanesh@mydomain.dev`

### What runs

| Service | Image               | Ports         | Volume                          |
| ------- | ------------------- | ------------- | ------------------------------- |
| `web`   | `portfolio-web`     | 80, 443       | `caddy_data` (TLS certificates) |
| `ssh`   | `portfolio-ssh`     | `${SSH_PORT}` | `ssh_data` (host key + stats)   |

The SSH **host key** and **visitor stats** persist in the `ssh_data` volume, so
clients don't get host-key-changed warnings across restarts.

### Operate

```bash
docker compose ps
docker compose logs -f ssh
docker compose logs -f web
docker compose pull && docker compose up -d   # update to newer images
docker compose down                            # stop
```

Health checks are built in (`/healthz` on the SSH container; HTTP 200 on web).

---

## 5. Update the résumé

Edit **one** file and redeploy:

```bash
$EDITOR shared/resume.json
docker compose up -d --build       # rebuilds web bundle; SSH picks up the file
```

> To update the SSH content **without** rebuilding, mount the file live by adding
> to the `ssh` service in `docker-compose.yml`:
> `volumes: ["./shared/resume.json:/app/shared/resume.json:ro"]`.

---

## 6. Alternative: systemd (no Docker) for the SSH server

```bash
cd ssh
go build -ldflags "-X github.com/sivanesh/portfolio-ssh/internal/version.Version=1.0.0" \
  -o portfolio-ssh ./cmd/portfolio-ssh

sudo useradd --system --home /var/lib/portfolio-ssh --shell /usr/sbin/nologin portfolio
sudo mkdir -p /opt/portfolio-ssh
sudo cp portfolio-ssh /opt/portfolio-ssh/
sudo cp ../shared/resume.json /opt/portfolio-ssh/resume.json
sudo cp ../deploy/portfolio-ssh.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now portfolio-ssh
sudo systemctl status portfolio-ssh
```

The unit binds `:22` via `CAP_NET_BIND_SERVICE` and applies seccomp/namespace
hardening. Serve the website separately with Caddy or Nginx (`deploy/Caddyfile`,
`deploy/nginx.conf`).

---

## 7. Website-only options

- **Vercel** — import the repo; preset **Vite** is auto-detected (Build
  `npm run build`, Output `dist`). No env vars needed.
- **GitHub Pages** — the included `.github/workflows/deploy.yml` builds with
  `VITE_BASE=/<repo>/` and publishes `dist`. (Pages cannot host SSH.)

## 8. CI/CD

- `.github/workflows/ci.yml` — lints/builds the web app and vets/tests/builds
  the Go server on every push & PR.
- `.github/workflows/docker-publish.yml` — builds and pushes
  `ghcr.io/<owner>/portfolio-web` and `…/portfolio-ssh` on pushes to `main`
  and version tags. On the server, `docker compose pull && up -d` to release.

---

## 9. Troubleshooting

| Symptom                                   | Fix                                                                 |
| ----------------------------------------- | ------------------------------------------------------------------- |
| `ssh` prompts for a password              | Use `-o PreferredAuthentications=none` or ensure no forced auth.    |
| Host key changed warning                  | Expected if the `ssh_data` volume was reset; remove the stale entry.|
| Certificate not issued                    | Check DNS resolves and ports 80/443 are open to the internet.       |
| Website 404s a deep link                  | Handled — Caddy/Nginx fall back to `index.html`.                    |
| Port 22 already in use                    | Move the host sshd (step 3) or set `SSH_PORT` to another port.      |
