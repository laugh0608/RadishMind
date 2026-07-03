#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DEV_SCRIPT="${ROOT_DIR}/scripts/run-radishmind-web-dev.sh"
CONSOLE_DEV_SCRIPT="${ROOT_DIR}/scripts/run-radishmind-console-dev.sh"
PLATFORM_SCRIPT="${ROOT_DIR}/scripts/run-platform-service.sh"

if [[ ! -f "${WEB_DEV_SCRIPT}" ]]; then
  echo "Cannot find ${WEB_DEV_SCRIPT}. Run this script from the RadishMind repository root." >&2
  exit 1
fi

usage() {
  cat <<'EOF'
Usage: ./start.sh [command] [args]

Commands:
  web-live       Start product web UI + dev-only fake read backend. Default for frontend/backend integration.
  web-offline    Start product web UI with offline fixtures only.
  console        Start local ops console + platform backend.
  platform       Start platform backend with dev-only read auth enabled.
  web-build      Run apps/radishmind-web build.
  diagnostics    Run platform diagnostics.
  check-fast     Run ./scripts/check-repo.sh --fast.
  help           Show this help.

Without a command, this script opens an interactive menu.
Extra args are forwarded to the selected dev launcher.
EOF
}

ensure_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Command '$1' not found. Install it before running this script." >&2
    exit 1
  fi
}

print_banner() {
  cat <<'EOF'
========================================
   ____           _ _     _     __  __ _           _
  |  _ \ __ _  __| (_)___| |__ |  \/  (_)_ __   __| |
  | |_) / _` |/ _` | / __| '_ \| |\/| | | '_ \ / _` |
  |  _ < (_| | (_| | \__ \ | | | |  | | | | | | (_| |
  |_| \_\__,_|\__,_|_|___/_| |_|_|  |_|_|_| |_|\__,_|
        RadishMind local dev launcher
========================================
EOF
}

print_menu() {
  echo
  echo "RadishMind local start menu"
  echo "---------------------------"
  echo "  0. Exit"
  echo
  echo "[Product UI]"
  echo "  1. Product Web dev-live  (platform @ http://127.0.0.1:7000 + web @ http://127.0.0.1:4100)"
  echo "  2. Product Web offline   (web @ http://127.0.0.1:4100)"
  echo
  echo "[Ops / Backend]"
  echo "  3. Local Ops Console     (platform @ http://127.0.0.1:7000 + console @ http://127.0.0.1:4000)"
  echo "  4. Platform backend      (dev-only read auth enabled @ http://127.0.0.1:7000)"
  echo "  5. Platform diagnostics"
  echo
  echo "[Verification]"
  echo "  6. Web build             (apps/radishmind-web)"
  echo "  7. Fast repo check       (./scripts/check-repo.sh --fast)"
  echo
  echo "Tip: Product Web dev-live is the default choice for frontend/backend integration."
  echo "Tip: If macOS Control Center / AirPlay occupies backend 7000, run ./start.sh web-live --backend-url http://127.0.0.1:7100."
  echo
}

run_web_live() {
  exec "${WEB_DEV_SCRIPT}" --mode dev-live "$@"
}

run_web_offline() {
  exec "${WEB_DEV_SCRIPT}" --mode offline "$@"
}

run_console() {
  exec "${CONSOLE_DEV_SCRIPT}" "$@"
}

run_platform() {
  export RADISHMIND_PLATFORM_LISTEN_ADDR="${RADISHMIND_PLATFORM_LISTEN_ADDR:-127.0.0.1:7000}"
  export RADISHMIND_PLATFORM_PROVIDER="${RADISHMIND_PLATFORM_PROVIDER:-mock}"
  export RADISHMIND_PLATFORM_MODEL="${RADISHMIND_PLATFORM_MODEL:-radishmind-local-dev}"
  export RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH="1"
  exec "${PLATFORM_SCRIPT}" serve "$@"
}

run_web_build() {
  ensure_command npm
  cd "${ROOT_DIR}/apps/radishmind-web"
  exec npm run build
}

run_diagnostics() {
  exec "${PLATFORM_SCRIPT}" diagnostics "$@"
}

run_check_fast() {
  exec "${ROOT_DIR}/scripts/check-repo.sh" --fast "$@"
}

run_command() {
  local command="$1"
  shift || true
  case "${command}" in
    web-live | web | dev-live) run_web_live "$@" ;;
    web-offline | offline) run_web_offline "$@" ;;
    console) run_console "$@" ;;
    platform | backend) run_platform "$@" ;;
    web-build | build) run_web_build "$@" ;;
    diagnostics | diag) run_diagnostics "$@" ;;
    check-fast | check) run_check_fast "$@" ;;
    help | -h | --help) usage ;;
    *) echo "Unknown command: ${command}" >&2; usage >&2; exit 2 ;;
  esac
}

if [[ $# -gt 0 ]]; then
  run_command "$@"
  exit 0
fi

print_banner
print_menu
read -rp "Enter choice number: " choice

case "${choice}" in
  0) echo "Bye."; exit 0 ;;
  1) run_web_live ;;
  2) run_web_offline ;;
  3) run_console ;;
  4) run_platform ;;
  5) run_diagnostics ;;
  6) run_web_build ;;
  7) run_check_fast ;;
  *) echo "Unknown choice: ${choice}" >&2; exit 2 ;;
esac
