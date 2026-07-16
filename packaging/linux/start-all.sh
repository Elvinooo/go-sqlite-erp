#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

mkdir -p logs run data/uploads

chmod +x "$BASE_DIR/erp-server" "$BASE_DIR/erp-frontend" "$BASE_DIR"/*.sh 2>/dev/null || true

is_running() {
  local pid_file="$1"
  [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null
}

if is_running run/backend.pid; then
  echo "backend already running, pid=$(cat run/backend.pid)"
else
  nohup "$BASE_DIR/erp-server" >> "$BASE_DIR/logs/backend.out.log" 2>&1 &
  echo $! > run/backend.pid
  echo "backend started, pid=$(cat run/backend.pid)"
fi

if is_running run/frontend.pid; then
  echo "frontend already running, pid=$(cat run/frontend.pid)"
else
  nohup "$BASE_DIR/erp-frontend" -listen "0.0.0.0:5173" -web "$BASE_DIR/web" -backend "http://127.0.0.1:18080" >> "$BASE_DIR/logs/frontend.out.log" 2>&1 &
  echo $! > run/frontend.pid
  echo "frontend started, pid=$(cat run/frontend.pid)"
fi

wait_for_http() {
  local url="$1"
  local seconds="${2:-15}"
  local i
  for ((i = 1; i <= seconds; i++)); do
    if command -v curl >/dev/null 2>&1; then
      if curl -fsS --max-time 1 "$url" >/dev/null 2>&1; then
        return 0
      fi
    elif command -v wget >/dev/null 2>&1; then
      if wget -qO- --timeout=1 "$url" >/dev/null 2>&1; then
        return 0
      fi
    else
      return 0
    fi
    sleep 1
  done
  return 1
}

if wait_for_http "http://127.0.0.1:18080/healthz" 20 && wait_for_http "http://127.0.0.1:5173/login" 20; then
  echo "services ready"
  echo "open http://127.0.0.1:5173/login"
else
  echo "startup timeout, see logs/backend.out.log and logs/frontend.out.log"
  echo "--- backend log tail ---"
  tail -n 40 "$BASE_DIR/logs/backend.out.log" 2>/dev/null || true
  echo "--- frontend log tail ---"
  tail -n 40 "$BASE_DIR/logs/frontend.out.log" 2>/dev/null || true
  exit 1
fi
