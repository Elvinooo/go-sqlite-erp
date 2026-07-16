#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

check_pid() {
  local name="$1"
  local pid_file="$2"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "$name running, pid=$(cat "$pid_file")"
  else
    echo "$name stopped"
  fi
}

check_pid "backend" run/backend.pid
check_pid "frontend" run/frontend.pid

if command -v curl >/dev/null 2>&1; then
  echo "backend health:"
  curl -fsS "http://127.0.0.1:18080/healthz" || true
  echo
fi
