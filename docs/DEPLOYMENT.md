# Deployment guide

One portfolio, two synchronized frontends:

- **Website** — `https://mydomain.dev` (static React site, served by Caddy with automatic HTTPS)
- **SSH** — `ssh mydomain.dev` (Go + Charm/wish TUI, fully sandboxed — no shell)

Both read the single `shared/resume.json`, so updating the résumé updates both.

---

## Web on Vercel + SSH on its own host (the split you have)

**Vercel can't run the SSH server** — it only serves HTTP, so it can't listen on
raw TCP port 22. Keep the website on Vercel and run the tiny Go SSH server on any
box with a public IP. Two easy options:

### A) Cheap VPS (~$4–6/mo: Hetzner, DigitalOcean, Vultr, or Oracle free tier)

On a fresh Ubuntu 22.04+ server:

```bash
# 1. Install Docker
curl -fsSL https://get.docker.com | sh

# 2. Get the code
git clone <this-repo> && cd terminal-resume

# 3. Free port 22 for the portfolio — move your ADMIN sshd first or you'll lock out!
sudo sed -i 's/^#\?Port 22/Port 2200/' /etc/ssh/sshd_config && sudo systemctl restart ssh
#    from now on administer the box via:  ssh -p 2200 you@server

# 4. Run ONLY the SSH server (it points back at your Vercel website)
docker compose -f docker-compose.ssh.yml up -d --build

# 5. Firewall
sudo ufw allow 22/tcp && sudo ufw allow 2200/tcp && sudo ufw --force enable
```

Test from your laptop:

```bash
ssh sivanesh@<server-ip>
```

Don't want to move the host's sshd? Run the portfolio on another port instead —
set `SSH_PORT=2222` before step 4, and connect with `ssh -p 2222 sivanesh@<ip>`.

### B) Fly.io (no server to manage) — config in `deploy/fly.toml`

Run from the repo root:

```bash
curl -L https://fly.io/install.sh | sh
fly auth login
fly launch --no-deploy --copy-config --config deploy/fly.toml --name <your-app>
fly volume create ssh_data --size 1 --region bom --config deploy/fly.toml
fly ips allocate-v4 --config deploy/fly.toml      # dedicated IPv4 (~$2/mo): required for raw TCP
fly deploy --config deploy/fly.toml --dockerfile ssh/Dockerfile
ssh sivanesh@<your-app>.fly.dev
```

### Give it a hostname (optional, terminal.shop-style)

`*.vercel.app` **cannot** be used for SSH. To get `ssh sivanesh@cv.yourdomain.dev`:

1. Own a domain (Cloudflare, Namecheap, Porkbun…).
2. Add a DNS **A record** `cv` (or `ssh`) → your server's IP (VPS), or a
   **CNAME** to `<your-app>.fly.dev` (Fly).
3. Connect: `ssh sivanesh@cv.yourdomain.dev`.

You can even add that domain to Vercel for the website and use a *different*
subdomain for SSH.

### Notes

- Any username works (`sivanesh@`, `guest@`) — it only personalises the prompt.
- The host key + visitor stats persist in the `ssh_data` volume/mount, so users
  don't get "host key changed" warnings across restarts and redeploys.
- Advertise it on the site — e.g. the `contact` output could read
  `ssh sivanesh@cv.yourdomain.dev`.

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

Visitors expect `ssh mydomain.dev` (port 22). Move the host's own
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
- SSH: `ssh mydomain.dev`

### Faster: pull prebuilt images instead of building on the VM

Building the React + Go images on a small/free-tier box is slow and RAM-hungry.
CI publishes **multi-arch** images (amd64 + arm64) to GHCR on every push to
`main` (`.github/workflows/docker-publish.yml`), so on the server you can just
**pull** — no compiler, no source build:

```bash
export DOMAIN=mydomain.dev SSH_PORT=22
docker compose -f docker-compose.prod.yml pull     # fetch latest images
docker compose -f docker-compose.prod.yml up -d    # (re)start, never builds
```

`docker-compose.prod.yml` references `ghcr.io/<owner>/portfolio-web` and
`portfolio-ssh` directly (override with `GHCR_OWNER` / `IMAGE_TAG`). It keeps the
same ports, volumes, health checks and log persistence as the build file. To
redeploy after a new push: `pull` again, then `up -d`.

> If the packages are private, `docker login ghcr.io -u <user>` first (PAT with
> `read:packages`); if public, no login is needed.

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

## 9. SSH bot noise & fail2ban

Any public SSH on port 22 is constantly probed by internet scanners/bots. This
is harmless — the server is sandboxed (no shell, no host access) — but it spams
logs and would inflate visitor stats. Two defenses are built in:

1. **App filter (automatic).** Non-interactive connections (no PTY — i.e. bots)
   are never counted as visitors and are kept out of the main log. When
   `PROBE_LOG` is set (the compose files set it to
   `/var/log/portfolio/probes.log` → `./logs/ssh/probes.log` on the host), their
   source IP is recorded there, one line per probe, for fail2ban.

2. **fail2ban (opt-in).** Ban repeat-offender IPs at the host firewall:

   ```bash
   sudo apt-get install -y fail2ban
   sudo cp deploy/fail2ban/filter.d/portfolio-ssh.conf /etc/fail2ban/filter.d/
   sudo cp deploy/fail2ban/jail.d/portfolio-ssh.local  /etc/fail2ban/jail.d/
   # Make sure logpath in the jail matches your repo path (default /opt/terminal-cv).
   sudo systemctl restart fail2ban

   # verify
   sudo fail2ban-regex logs/ssh/probes.log deploy/fail2ban/filter.d/portfolio-ssh.conf
   sudo fail2ban-client status portfolio-ssh
   ```

   The jail bans an IP that probes 10+ times in 10 minutes for 24 h. Because the
   SSH port is **published by Docker** (traffic is FORWARDed to the container and
   bypasses the host `INPUT` chain), the ban is applied in the **`DOCKER-USER`**
   chain, which every published-port packet traverses. Once banned, the bot's
   packets are dropped at the firewall, so the probe log goes quiet too.

---

## 10. Troubleshooting

| Symptom                                   | Fix                                                                 |
| ----------------------------------------- | ------------------------------------------------------------------- |
| `ssh` prompts for a password              | Use `-o PreferredAuthentications=none` or ensure no forced auth.    |
| Host key changed warning                  | Expected if the `ssh_data` volume was reset; remove the stale entry.|
| Certificate not issued                    | Check DNS resolves and ports 80/443 are open to the internet.       |
| Website 404s a deep link                  | Handled — Caddy/Nginx fall back to `index.html`.                    |
| Port 22 already in use                    | Move the host sshd (step 3) or set `SSH_PORT` to another port.      |
