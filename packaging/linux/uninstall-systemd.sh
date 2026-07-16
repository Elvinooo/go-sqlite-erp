#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "please run as root: sudo ./uninstall-systemd.sh"
  exit 1
fi

systemctl stop erp-runtime.service 2>/dev/null || true
systemctl disable erp-runtime.service 2>/dev/null || true
rm -f /etc/systemd/system/erp-runtime.service
systemctl daemon-reload
echo "erp-runtime.service removed"
