#!/usr/bin/env bash
set -euo pipefail

backend_url="http://127.0.0.1:7000"
frontend_url="http://127.0.0.1:4000"
timeout_seconds=60
verify_only=0
exit_after_probe=0

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
console_dir="${repo_root}/apps/radishmind-console"
platform_wrapper="${repo_root}/scripts/run-platform-service.sh"
log_dir="${repo_root}/tmp/radishmind-console-dev"
spawned_pids=()

usage() {
  cat <<'EOF'
Usage: scripts/run-radishmind-console-dev.sh [options]

Options:
  --backend-url URL        Platform base URL. Default: http://127.0.0.1:7000
  --frontend-url URL       Console URL. Default: http://127.0.0.1:4000
  --timeout-seconds N      Probe timeout. Default: 60
  --verify-only            Probe existing backend/frontend processes only.
  --exit-after-probe       Start missing local processes, probe, then stop spawned processes.
  -h, --help               Show this help.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --backend-url)
      backend_url="${2:?missing value for --backend-url}"
      shift 2
      ;;
    --frontend-url)
      frontend_url="${2:?missing value for --frontend-url}"
      shift 2
      ;;
    --timeout-seconds)
      timeout_seconds="${2:?missing value for --timeout-seconds}"
      shift 2
      ;;
    --verify-only)
      verify_only=1
      shift
      ;;
    --exit-after-probe)
      exit_after_probe=1
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "unsupported argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

step() {
  echo "[radishmind-console-dev] $*"
}

require_command() {
  local name="$1"
  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "Missing required command: ${name}" >&2
    exit 1
  fi
}

find_python() {
  if command -v python3 >/dev/null 2>&1; then
    command -v python3
  elif command -v python >/dev/null 2>&1; then
    command -v python
  else
    echo "Missing required command: python3 or python" >&2
    exit 1
  fi
}

python_bin="$(find_python)"

url_part() {
  local url="$1"
  local field="$2"
  "${python_bin}" - "$url" "$field" <<'PY'
import sys
from urllib.parse import urlparse

url = sys.argv[1]
field = sys.argv[2]
parsed = urlparse(url)
if parsed.scheme not in {"http", "https"}:
    raise SystemExit(f"URL must use http or https: {url}")
if parsed.port is None:
    raise SystemExit(f"URL must include an explicit local development port: {url}")
if field == "host":
    print(parsed.hostname or "")
elif field == "port":
    print(parsed.port)
elif field == "scheme":
    print(parsed.scheme)
else:
    raise SystemExit(f"unsupported field: {field}")
PY
}

assert_browser_safe_port() {
  local url="$1"
  local port="$2"
  "${python_bin}" - "$url" "$port" <<'PY'
import sys

url = sys.argv[1]
port = int(sys.argv[2])
blocked = {
    1, 7, 9, 11, 13, 15, 17, 19, 20, 21, 22, 23, 25, 37, 42, 43, 53, 69,
    77, 79, 87, 95, 101, 102, 103, 104, 109, 110, 111, 113, 115, 117, 119,
    123, 135, 137, 139, 143, 161, 179, 389, 427, 465, 512, 513, 514, 515,
    526, 530, 531, 532, 540, 548, 554, 556, 563, 587, 601, 636, 989, 990,
    993, 995, 1719, 1720, 1723, 2049, 3659, 4045, 4190, 5060, 5061, 6000,
    6566, 6697, 10080,
}
if port in blocked or 6665 <= port <= 6669:
    print(f"Browser unsafe port: {port}", file=sys.stderr)
    print("Browsers can fail with ERR_UNSAFE_PORT before reaching RadishMind.", file=sys.stderr)
    print("Use the local defaults instead: backend http://127.0.0.1:7000 and frontend http://127.0.0.1:4000.", file=sys.stderr)
    raise SystemExit(1)
PY
}

port_is_open() {
  local host="$1"
  local port="$2"
  "${python_bin}" - "$host" "$port" <<'PY'
import socket
import sys

host = sys.argv[1]
port = int(sys.argv[2])
try:
    with socket.create_connection((host, port), timeout=0.5):
        raise SystemExit(0)
except OSError:
    raise SystemExit(1)
PY
}

probe_json() {
  local url="$1"
  local expected_kind="$2"
  "${python_bin}" - "$url" "$expected_kind" <<'PY'
import json
import sys
from urllib.request import Request, urlopen

url = sys.argv[1]
expected_kind = sys.argv[2]
request = Request(url, headers={"Accept": "application/json"}, method="GET")
with urlopen(request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
    document = json.loads(response.read().decode("utf-8"))
if expected_kind and document.get("kind") != expected_kind:
    raise SystemExit(f"Unexpected kind from {url}. expected={expected_kind} actual={document.get('kind')}")
PY
}

probe_page() {
  local url="$1"
  "${python_bin}" - "$url" <<'PY'
import sys
from urllib.request import Request, urlopen

url = sys.argv[1]
request = Request(url, method="GET")
with urlopen(request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
    body = response.read().decode("utf-8", errors="replace")
if '<div id="root">' not in body:
    raise SystemExit("Frontend responded, but it does not look like the RadishMind console shell.")
PY
}

probe_cors() {
  local overview_url="$1"
  local origin="$2"
  "${python_bin}" - "$overview_url" "$origin" <<'PY'
import sys
from urllib.request import Request, urlopen

url = sys.argv[1]
origin = sys.argv[2]
request = Request(
    url,
    headers={
        "Origin": origin,
        "Access-Control-Request-Method": "GET",
    },
    method="OPTIONS",
)
with urlopen(request, timeout=5) as response:
    allow_origin = response.headers.get("Access-Control-Allow-Origin")
if allow_origin != origin:
    raise SystemExit(f"CORS preflight did not allow {origin}. actual Access-Control-Allow-Origin={allow_origin}")
PY
}

wait_until() {
  local name="$1"
  shift
  local deadline=$((SECONDS + timeout_seconds))
  local last_error=""
  while (( SECONDS < deadline )); do
    if last_error="$("$@" 2>&1)"; then
      step "${name} probe passed."
      return 0
    fi
    sleep 0.75
  done
  echo "${name} probe failed after ${timeout_seconds}s. Last error: ${last_error}" >&2
  return 1
}

show_failure_help() {
  local message="$1"
  {
    echo ""
    echo "RadishMind console dev entry failed:"
    echo "${message}"
    echo ""
    echo "Common local failures:"
    echo "- Port conflict: backend must answer on http://127.0.0.1:7000 and frontend on http://127.0.0.1:4000. Stop the other process or change the existing service back to the defaults."
    echo "- CORS/preflight: the platform service only allows http://127.0.0.1:4000 and http://localhost:4000 as local console origins."
    echo "- Browser unsafe port: ERR_UNSAFE_PORT means the browser blocked the port before reaching the app. Keep backend 7000 and frontend 4000."
    echo "- Missing dependencies: run npm install in apps/radishmind-console if npm cannot start Vite."
    echo "- Backend diagnostics: run ./scripts/run-platform-service.sh diagnostics from the repo root."
    echo "- Logs: ${log_dir}"
  } >&2
}

cleanup() {
  if [[ "${verify_only}" -eq 1 ]]; then
    return
  fi
  for pid in "${spawned_pids[@]:-}"; do
    if kill -0 "${pid}" >/dev/null 2>&1; then
      step "Stopping process ${pid}."
      kill "${pid}" >/dev/null 2>&1 || true
    fi
  done
}

trap cleanup EXIT INT TERM

backend_host="$(url_part "${backend_url}" host)"
backend_port="$(url_part "${backend_url}" port)"
frontend_host="$(url_part "${frontend_url}" host)"
frontend_port="$(url_part "${frontend_url}" port)"
assert_browser_safe_port "${backend_url}" "${backend_port}"
assert_browser_safe_port "${frontend_url}" "${frontend_port}"

healthz_url="${backend_url%/}/healthz"
overview_url="${backend_url%/}/v1/platform/overview"
frontend_origin="${frontend_url%/}"

if [[ ! -f "${platform_wrapper}" ]]; then
  show_failure_help "Missing platform wrapper: ${platform_wrapper}"
  exit 1
fi
if [[ ! -f "${console_dir}/package.json" ]]; then
  show_failure_help "Missing RadishMind console package.json: ${console_dir}"
  exit 1
fi

if [[ "${verify_only}" -eq 0 ]]; then
  require_command npm
  mkdir -p "${log_dir}"

  if port_is_open "${backend_host}" "${backend_port}"; then
    step "Backend port ${backend_port} is already open; reusing it if probes pass."
  else
    step "Starting platform. Logs: ${log_dir}/platform.out.log ; ${log_dir}/platform.err.log"
    (
      export RADISHMIND_PLATFORM_LISTEN_ADDR="${RADISHMIND_PLATFORM_LISTEN_ADDR:-${backend_host}:${backend_port}}"
      export RADISHMIND_PLATFORM_PROVIDER="${RADISHMIND_PLATFORM_PROVIDER:-mock}"
      export RADISHMIND_PLATFORM_MODEL="${RADISHMIND_PLATFORM_MODEL:-radishmind-local-dev}"
      exec "${platform_wrapper}" serve
    ) >"${log_dir}/platform.out.log" 2>"${log_dir}/platform.err.log" &
    spawned_pids+=("$!")
  fi

  if port_is_open "${frontend_host}" "${frontend_port}"; then
    step "Frontend port ${frontend_port} is already open; reusing it if probes pass."
  else
    step "Starting console. Logs: ${log_dir}/console.out.log ; ${log_dir}/console.err.log"
    (
      cd "${console_dir}"
      export VITE_RADISHMIND_PLATFORM_BASE_URL="${VITE_RADISHMIND_PLATFORM_BASE_URL:-${backend_url%/}}"
      exec npm run dev
    ) >"${log_dir}/console.out.log" 2>"${log_dir}/console.err.log" &
    spawned_pids+=("$!")
  fi
fi

if ! wait_until "backend healthz" probe_json "${healthz_url}" ""; then
  show_failure_help "backend healthz probe failed"
  exit 1
fi
if ! wait_until "platform overview" probe_json "${overview_url}" "platform_overview"; then
  show_failure_help "platform overview probe failed"
  exit 1
fi
if ! wait_until "local console CORS" probe_cors "${overview_url}" "${frontend_origin}"; then
  show_failure_help "local console CORS probe failed"
  exit 1
fi
if ! wait_until "frontend console" probe_page "${frontend_url}"; then
  show_failure_help "frontend console probe failed"
  exit 1
fi

step "Local console is ready: ${frontend_url}"
step "Backend probes passed: ${healthz_url} ; ${overview_url}"
step "This is a dev-only launcher, not a production supervisor. It does not implement executor, durable store, confirmation, writeback, or replay."

if [[ "${verify_only}" -eq 0 && "${exit_after_probe}" -eq 0 && "${#spawned_pids[@]}" -gt 0 ]]; then
  step "Press Ctrl+C to stop spawned local dev processes."
  while true; do
    for pid in "${spawned_pids[@]}"; do
      if ! kill -0 "${pid}" >/dev/null 2>&1; then
        show_failure_help "Spawned process exited early: pid=${pid}. Check logs in ${log_dir}"
        exit 1
      fi
    done
    sleep 2
  done
fi
