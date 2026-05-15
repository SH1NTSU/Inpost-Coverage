#!/usr/bin/env bash
# Stops api + scraper + web started by scripts/start-all.sh.
# Idempotent: missing PID files / dead processes are silently ignored.

set -euo pipefail

PID_DIR="/tmp/inpost-pids"

for svc in api scraper web; do
  pidfile="$PID_DIR/$svc.pid"
  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile" 2>/dev/null || true)"
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      echo "→ stopping $svc (pid $pid)"
      kill "$pid" 2>/dev/null || true
      # `go run` spawns a child Go binary; kill the whole process group
      pkill -P "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
done

# Belt-and-suspenders: catch any straggler processes by name.
pkill -f "cmd/api"     2>/dev/null || true
pkill -f "cmd/scraper" 2>/dev/null || true
pkill -f "vite"        2>/dev/null || true

echo "✓ all stopped"
