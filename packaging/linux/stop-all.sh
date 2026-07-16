#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

stop_pid_file() {
  local name="$1"
  local pid_file="$2"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file")"
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      sleep 1
      if kill -0 "$pid" 2>/dev/null; then
        kill -9 "$pid" 2>/dev/null || true
      fi
      echo "$name stopped"
    else
      echo "$name not running"
    fi
    rm -f "$pid_file"
  else
    echo "$name pid file not found"
  fi
}

stop_pid_file "frontend" run/frontend.pid
stop_pid_file "backend" run/backend.pid
