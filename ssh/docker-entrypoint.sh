#!/bin/sh
# Entrypoint for the portfolio SSH server.
#
# Runs briefly as root only to make the data + log directories writable by the
# unprivileged `app` user — this matters when they are bind-mounted from the
# host (root-owned) rather than Docker-managed volumes — then drops privileges
# and hands PID 1 to the server (via exec) so graceful shutdown still works.
set -e

log_dir=$(dirname "${LOG_FILE:-/data/ssh.log}")
for d in /data "$log_dir"; do
	mkdir -p "$d" 2>/dev/null || true
	chown -R app:app "$d" 2>/dev/null || true
done

exec su-exec app:app /app/portfolio-ssh "$@"
