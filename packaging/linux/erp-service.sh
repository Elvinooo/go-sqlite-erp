#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

mkdir -p logs run data/uploads

BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  if [[ -n "$FRONTEND_PID" ]] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
  if [[ -n "$BACKEND_PID" ]] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

"$BASE_DIR/erp-server" &
BACKEND_PID=$!
echo "$BACKEND_PID" > run/backend.pid

"$BASE_DIR/erp-frontend" -listen "0.0.0.0:5173" -web "$BASE_DIR/web" -backend "http://127.0.0.1:18080" &
FRONTEND_PID=$!
echo "$FRONTEND_PID" > run/frontend.pid

wait -n "$BACKEND_PID" "$FRONTEND_PID"
