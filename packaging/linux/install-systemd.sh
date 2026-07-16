#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "please run as root: sudo ./install-systemd.sh"
  exit 1
fi

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_FILE="/etc/systemd/system/erp-runtime.service"

chmod +x "$BASE_DIR/erp-server" "$BASE_DIR/erp-frontend" "$BASE_DIR"/*.sh

cat > "$SERVICE_FILE" <<SERVICE
[Unit]
Description=ERP Runtime
After=network.target

[Service]
Type=simple
WorkingDirectory=$BASE_DIR
ExecStart=$BASE_DIR/erp-service.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SERVICE

systemctl daemon-reload
systemctl enable erp-runtime.service
systemctl restart erp-runtime.service
systemctl status erp-runtime.service --no-pager
