#!/usr/bin/env bash
# Background-runs api + scraper + web with logs in /tmp/inpost-logs/
# and PID files in /tmp/inpost-pids/ so `make stop` can clean them up.
# Usage: ./scripts/start-all.sh  (or `make start`)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="/tmp/inpost-logs"
PID_DIR="/tmp/inpost-pids"
mkdir -p "$LOG_DIR" "$PID_DIR"

# Kill any leftover processes from a previous run so ports don't collide.
for svc in api scraper web; do
  pidfile="$PID_DIR/$svc.pid"
  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile" 2>/dev/null || true)"
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      echo "  (stopping previous $svc pid=$pid)"
      kill "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
done

cd "$REPO_ROOT"

echo "→ Starting API on :8090..."
go run ./cmd/api > "$LOG_DIR/api.log" 2>&1 &
echo $! > "$PID_DIR/api.pid"

echo "→ Starting scraper..."
go run ./cmd/scraper > "$LOG_DIR/scraper.log" 2>&1 &
echo $! > "$PID_DIR/scraper.pid"

echo "→ Starting web dev server on :5173..."
(
  cd "$REPO_ROOT/web"
  if [ ! -d node_modules ]; then
    echo "  (first run: installing npm dependencies, this takes a minute)" >&2
    npm install --silent
  fi
  npm run dev
) > "$LOG_DIR/web.log" 2>&1 &
echo $! > "$PID_DIR/web.pid"

# Give services a moment to bind their ports before reporting.
sleep 2

cat <<EOF

╭───────────────────────────────────────────────────╮
│  InPost Network Control — running                 │
├───────────────────────────────────────────────────┤
│  Web UI    http://localhost:5173                  │
│  API       http://localhost:8090                  │
│  Scraper   background                             │
├───────────────────────────────────────────────────┤
│  Logs                                             │
│    tail -f $LOG_DIR/api.log
│    tail -f $LOG_DIR/scraper.log
│    tail -f $LOG_DIR/web.log
│                                                   │
│  Stop:  make stop                                 │
╰───────────────────────────────────────────────────╯
EOF
