#!/usr/bin/env bash
set -euo pipefail

mode="dev-live"
backend_url="http://127.0.0.1:7000"
frontend_url="http://127.0.0.1:4100"
tenant_ref="tenant_demo"
subject_ref="subject_demo_user"
timeout_seconds=60
verify_only=0
exit_after_probe=0
saved_draft_dev=0
saved_draft_postgres_dev_test=0
workflow_diagnostics_dev=0
saved_draft_workspace_id="workspace_demo"
saved_draft_application_id="app_flow_copilot"

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
web_dir="${repo_root}/apps/radishmind-web"
platform_wrapper="${repo_root}/scripts/run-platform-service.sh"
platform_dir="${repo_root}/services/platform"
log_dir="${repo_root}/tmp/radishmind-web-dev"
spawned_pids=()

usage() {
  cat <<'EOF'
Usage: scripts/run-radishmind-web-dev.sh [options]

Options:
  --mode MODE             offline or dev-live. Default: dev-live
  --offline               Shortcut for --mode offline.
  --dev-live              Shortcut for --mode dev-live.
  --backend-url URL       Platform base URL. Default: http://127.0.0.1:7000
  --frontend-url URL      Web UI URL. Default: http://127.0.0.1:4100
  --timeout-seconds N     Probe timeout. Default: 60
  --saved-draft-dev       Enable the explicit memory-dev Saved Draft read/write path.
  --saved-draft-postgres-dev-test
                           Enable the explicit PostgreSQL dev/test Saved Draft path.
  --workflow-diagnostics-dev
                           Enable fixed mock Workflow failure scenarios; requires a Saved Draft dev mode.
  --verify-only           Probe existing backend/frontend processes only.
  --exit-after-probe      Start missing local processes, probe, then stop spawned processes.
  -h, --help              Show this help.

Modes:
  offline                 Start/probe only apps/radishmind-web with offline fixtures.
  dev-live                Start/probe platform + web, with dev-only fake read auth enabled.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      mode="${2:?missing value for --mode}"
      shift 2
      ;;
    --offline)
      mode="offline"
      shift
      ;;
    --dev-live)
      mode="dev-live"
      shift
      ;;
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
    --saved-draft-dev)
      saved_draft_dev=1
      shift
      ;;
    --saved-draft-postgres-dev-test)
      saved_draft_postgres_dev_test=1
      shift
      ;;
    --workflow-diagnostics-dev)
      workflow_diagnostics_dev=1
      shift
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

case "${mode}" in
  offline | dev-live) ;;
  *)
    echo "unsupported mode: ${mode}" >&2
    usage >&2
    exit 2
    ;;
esac

if [[ "${saved_draft_dev}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--saved-draft-dev requires --mode dev-live" >&2
  exit 2
fi
if [[ "${saved_draft_postgres_dev_test}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--saved-draft-postgres-dev-test requires --mode dev-live" >&2
  exit 2
fi
if [[ "${saved_draft_dev}" -eq 1 && "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
  echo "Choose either --saved-draft-dev or --saved-draft-postgres-dev-test" >&2
  exit 2
fi
if [[ "${workflow_diagnostics_dev}" -eq 1 && "${saved_draft_dev}" -eq 0 && "${saved_draft_postgres_dev_test}" -eq 0 ]]; then
  echo "--workflow-diagnostics-dev requires --saved-draft-dev or --saved-draft-postgres-dev-test" >&2
  exit 2
fi

saved_draft_enabled=0
if [[ "${saved_draft_dev}" -eq 1 || "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
  saved_draft_enabled=1
fi

step() {
  echo "[radishmind-web-dev] $*"
}

require_command() {
  local name="$1"
  if ! command -v "${name}" >/dev/null 2>&1; then
    echo "Missing required command: ${name}" >&2
    exit 1
  fi
}

find_python() {
  local python_bin="${repo_root}/.venv/bin/python"
  if [[ -x "${python_bin}" ]]; then
    echo "${python_bin}"
  else
    echo "Missing repository virtual environment: ${python_bin}" >&2
    echo "Run ./scripts/bootstrap-dev.sh before running scripts/run-radishmind-web-dev.sh" >&2
    exit 1
  fi
}

python_bin="$(find_python)"

build_saved_draft_database_url() {
  if [[ -n "${RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL:-}" ]]; then
    printf '%s\n' "${RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL}"
    return
  fi
  "${python_bin}" - <<'PY'
import os
from urllib.parse import quote, urlencode

host = os.environ.get("PGHOST", "127.0.0.1")
port = os.environ.get("PGPORT", "55432")
user = quote(os.environ.get("PGUSER", "radishmind_runtime"), safe="")
password = os.environ.get("PGPASSWORD", "")
userinfo = user if not password else f"{user}:{quote(password, safe='')}"
database = quote(os.environ.get("PGDATABASE", "radishmind_saved_draft_test"), safe="")
query = urlencode({"sslmode": os.environ.get("PGSSLMODE", "disable")})
if ":" in host and not host.startswith("["):
    host = f"[{host}]"
print(f"postgresql://{userinfo}@{host}:{port}/{database}?{query}")
PY
}

probe_saved_draft_postgres_migration() {
  local database_url="$1"
  (
    export RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
		export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="postgres_dev_test"
    export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL="${database_url}"
		export RADISHMIND_WORKFLOW_RUN_STORE="postgres_dev_test"
		export RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL="${database_url}"
    cd "${platform_dir}"
    go run ./cmd/radishmind-workflow-draft-migrate status >/dev/null
		go run ./cmd/radishmind-workflow-run-migrate status >/dev/null
  )
}

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
    print(f"Browsers can fail with ERR_UNSAFE_PORT before reaching {url}.", file=sys.stderr)
    print("Use the local defaults instead: backend http://127.0.0.1:7000 and web http://127.0.0.1:4100.", file=sys.stderr)
    raise SystemExit(1)
PY
}

port_is_open() {
  local host="$1"
  local port="$2"
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi
  "${python_bin}" - "$host" "$port" <<'PY'
import socket
import sys

host = sys.argv[1]
port = int(sys.argv[2])
try:
    with socket.create_connection((host, port), timeout=0.5):
        raise SystemExit(0)
except PermissionError:
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

probe_control_plane_read_routes() {
  local base_url="$1"
  local tenant="$2"
  local subject="$3"
  "${python_bin}" - "$base_url" "$tenant" "$subject" <<'PY'
import json
import sys
from urllib.parse import quote
from urllib.request import Request, urlopen

base_url = sys.argv[1].rstrip("/")
tenant = sys.argv[2]
subject = sys.argv[3]
routes = [
    f"/v1/control-plane/tenants/{quote(tenant)}/summary",
    "/v1/user-workspace/applications",
    "/v1/user-workspace/api-keys",
    "/v1/user-workspace/usage/quota-summary",
    "/v1/user-workspace/workflow-definitions",
    "/v1/user-workspace/runs",
    "/v1/control-plane/audit",
]
headers = {
    "Accept": "application/json",
    "X-Request-Id": "dev-live-script-probe",
    "X-RadishMind-Dev-Read-Identity": "dev-live-read-consumer",
    "X-RadishMind-Dev-Read-Tenant": tenant,
    "X-RadishMind-Dev-Read-Subject": subject,
    "X-RadishMind-Dev-Read-Scopes": "tenant:read,applications:read,api_keys:read,usage:read,runs:read,audit:read",
    "X-RadishMind-Dev-Read-Audit": "audit_dev_live_script_probe",
}
for path in routes:
    url = f"{base_url}{path}"
    request = Request(url, headers=headers, method="GET")
    with urlopen(request, timeout=5) as response:
        if response.status < 200 or response.status >= 300:
            raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
        document = json.loads(response.read().decode("utf-8"))
    if not isinstance(document, dict):
        raise SystemExit(f"Route returned non-object JSON: {url}")
    if document.get("failure_code") is not None:
        raise SystemExit(f"Route returned failure_code={document.get('failure_code')} from {url}")
    if not isinstance(document.get("items"), list):
        raise SystemExit(f"Route did not return items[] from {url}")
PY
}

probe_saved_workflow_draft_route() {
  local base_url="$1"
  local tenant="$2"
  local subject="$3"
  local workspace_id="$4"
  local application_id="$5"
  "${python_bin}" - "$base_url" "$tenant" "$subject" "$workspace_id" "$application_id" <<'PY'
import json
import sys
from urllib.parse import urlencode
from urllib.request import Request, urlopen

base_url = sys.argv[1].rstrip("/")
tenant = sys.argv[2]
subject = sys.argv[3]
workspace_id = sys.argv[4]
application_id = sys.argv[5]
query = urlencode({"workspace_id": workspace_id, "application_id": application_id})
url = f"{base_url}/v1/user-workspace/workflow-drafts?{query}"
request = Request(
    url,
    headers={
        "Accept": "application/json",
        "X-Request-Id": "dev-live-saved-draft-probe",
        "X-RadishMind-Dev-Read-Identity": "dev-live-saved-draft-probe",
        "X-RadishMind-Dev-Read-Tenant": tenant,
        "X-RadishMind-Dev-Read-Subject": subject,
        "X-RadishMind-Dev-Read-Scopes": "workflow_drafts:read,workflow_drafts:write",
        "X-RadishMind-Dev-Read-Audit": "audit_dev_live_saved_draft_probe",
        "X-RadishMind-Dev-Workflow-Workspace": workspace_id,
        "X-RadishMind-Dev-Workflow-Application": application_id,
    },
    method="GET",
)
with urlopen(request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
    document = json.loads(response.read().decode("utf-8"))
if document.get("failure_code") is not None:
    raise SystemExit(f"Saved Draft route returned failure_code={document.get('failure_code')} from {url}")
if not isinstance(document.get("draft_summaries"), list):
    raise SystemExit(f"Saved Draft route did not return draft_summaries[] from {url}")
PY
}

probe_workflow_executor_route() {
  local base_url="$1"
  local tenant="$2"
  local subject="$3"
  local workspace_id="$4"
  local application_id="$5"
  "${python_bin}" - "$base_url" "$tenant" "$subject" "$workspace_id" "$application_id" <<'PY'
import json
import sys
from urllib.request import Request, urlopen

base_url = sys.argv[1].rstrip("/")
tenant = sys.argv[2]
subject = sys.argv[3]
workspace_id = sys.argv[4]
application_id = sys.argv[5]
url = f"{base_url}/v1/user-workspace/workflow-drafts/draft_executor_probe_missing/runs"
body = json.dumps({
    "workspace_id": workspace_id,
    "application_id": application_id,
    "input_text": "executor dev gate probe",
    "condition_values": {},
}).encode("utf-8")
request = Request(
    url,
    data=body,
    headers={
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-Request-Id": "dev-live-workflow-executor-probe",
        "X-RadishMind-Dev-Read-Identity": "dev-live-workflow-executor-probe",
        "X-RadishMind-Dev-Read-Tenant": tenant,
        "X-RadishMind-Dev-Read-Subject": subject,
        "X-RadishMind-Dev-Read-Scopes": "workflow_drafts:read,workflow_runs:execute,workflow_runs:read",
        "X-RadishMind-Dev-Read-Audit": "audit_dev_live_workflow_executor_probe",
        "X-RadishMind-Dev-Workflow-Workspace": workspace_id,
        "X-RadishMind-Dev-Workflow-Application": application_id,
    },
    method="POST",
)
with urlopen(request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
    document = json.loads(response.read().decode("utf-8"))
if document.get("failure_code") != "workflow_run_draft_not_found":
    raise SystemExit(f"Workflow Executor dev gate probe returned unexpected failure_code={document.get('failure_code')}")
if document.get("run") is not None:
    raise SystemExit("Workflow Executor dev gate probe must not create a run for a missing draft")
list_url = f"{base_url}/v1/user-workspace/workflow-runs?workspace_id={workspace_id}&application_id={application_id}&limit=1"
list_request = Request(list_url, headers={key: value for key, value in request.headers.items() if key.lower() != "content-type"}, method="GET")
with urlopen(list_request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {list_url}")
    list_document = json.loads(response.read().decode("utf-8"))
if list_document.get("failure_code") is not None or not isinstance(list_document.get("runs"), list):
    raise SystemExit("Workflow run history list probe did not return a successful runs[] envelope")
PY
}

probe_cors() {
  local read_url="$1"
  local origin="$2"
  "${python_bin}" - "$read_url" "$origin" <<'PY'
import sys
from urllib.request import Request, urlopen

url = sys.argv[1]
origin = sys.argv[2]
request = Request(
    url,
    headers={
        "Origin": origin,
        "Access-Control-Request-Method": "GET",
        "Access-Control-Request-Headers": "X-RadishMind-Dev-Read-Identity",
    },
    method="OPTIONS",
)
with urlopen(request, timeout=5) as response:
    allow_origin = response.headers.get("Access-Control-Allow-Origin")
    allow_headers = response.headers.get("Access-Control-Allow-Headers", "")
if allow_origin != origin:
    raise SystemExit(f"CORS preflight did not allow {origin}. actual Access-Control-Allow-Origin={allow_origin}")
if "X-RadishMind-Dev-Read-Identity" not in allow_headers:
    raise SystemExit("CORS preflight did not allow dev read identity header")
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
    raise SystemExit("Frontend responded, but it does not look like the RadishMind web shell.")
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

verify_existing_backend() {
  local last_error=""
  if last_error="$(probe_json "${healthz_url}" "" 2>&1)"; then
    step "Existing backend healthz probe passed."
    return 0
  fi
  {
    echo "Backend port ${backend_port} is open, but ${healthz_url} did not answer like RadishMind platform."
    echo "Last error: ${last_error}"
    if command -v lsof >/dev/null 2>&1; then
      echo ""
      echo "Current listener on backend port ${backend_port}:"
      lsof -nP -iTCP:"${backend_port}" -sTCP:LISTEN 2>/dev/null || true
    fi
  } >&2
  show_failure_help "backend port ${backend_port} is occupied by a non-RadishMind service. Stop that process or pass --backend-url with a free local port."
  exit 1
}

show_failure_help() {
  local message="$1"
  {
    echo ""
    echo "RadishMind web dev entry failed:"
    echo "${message}"
    echo ""
    echo "Common local failures:"
    echo "- Port conflict: backend should answer on http://127.0.0.1:7000 and web on http://127.0.0.1:4100."
    echo "- macOS port 7000 conflict: AirPlay / Control Center may occupy it; retry with --backend-url http://127.0.0.1:7100."
    echo "- Dev-live auth: backend must be started with RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1 for fake-store-backed read routes."
    echo "- Saved Draft mode: choose --saved-draft-dev or --saved-draft-postgres-dev-test so backend and web opt in together."
    echo "- PostgreSQL dev/test mode: start and migrate it with ./scripts/run-workflow-saved-draft-postgres-dev-test.sh check."
    echo "- CORS/preflight: platform should allow http://127.0.0.1:4100 and dev read headers in local development."
    echo "- Missing dependencies: run npm install in apps/radishmind-web if npm cannot start Vite."
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
  spawned_pids=()
}

handle_interrupt() {
  cleanup
  exit 130
}

trap cleanup EXIT
trap handle_interrupt INT TERM

backend_host="$(url_part "${backend_url}" host)"
backend_port="$(url_part "${backend_url}" port)"
frontend_host="$(url_part "${frontend_url}" host)"
frontend_port="$(url_part "${frontend_url}" port)"
assert_browser_safe_port "${backend_url}" "${backend_port}"
assert_browser_safe_port "${frontend_url}" "${frontend_port}"

healthz_url="${backend_url%/}/healthz"
tenant_summary_url="${backend_url%/}/v1/control-plane/tenants/${tenant_ref}/summary"
frontend_origin="${frontend_url%/}"

if [[ ! -f "${web_dir}/package.json" ]]; then
  show_failure_help "Missing RadishMind web package.json: ${web_dir}"
  exit 1
fi
if [[ "${mode}" == "dev-live" && ! -f "${platform_wrapper}" ]]; then
  show_failure_help "Missing platform wrapper: ${platform_wrapper}"
  exit 1
fi

saved_draft_database_url=""
if [[ "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
  require_command go
  saved_draft_database_url="$(build_saved_draft_database_url)"
  if ! probe_saved_draft_postgres_migration "${saved_draft_database_url}"; then
    show_failure_help "Saved Draft PostgreSQL migration preflight failed"
    exit 1
  fi
  step "Saved Draft PostgreSQL migration preflight passed."
fi

if [[ "${verify_only}" -eq 0 ]]; then
  require_command npm
  mkdir -p "${log_dir}"

  if [[ "${mode}" == "dev-live" ]]; then
    if port_is_open "${backend_host}" "${backend_port}"; then
      step "Backend port ${backend_port} is already open; reusing it if probes pass."
      verify_existing_backend
    else
      step "Starting platform with dev-only read auth. Logs: ${log_dir}/platform.out.log ; ${log_dir}/platform.err.log"
      (
        export RADISHMIND_PLATFORM_LISTEN_ADDR="${RADISHMIND_PLATFORM_LISTEN_ADDR:-${backend_host}:${backend_port}}"
        export RADISHMIND_PLATFORM_PROVIDER="${RADISHMIND_PLATFORM_PROVIDER:-mock}"
        export RADISHMIND_PLATFORM_MODEL="${RADISHMIND_PLATFORM_MODEL:-radishmind-local-dev}"
        export RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH="1"
        if [[ "${saved_draft_dev}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
          export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="memory_dev"
        elif [[ "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
          export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="postgres_dev_test"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
			export RADISHMIND_WORKFLOW_RUN_STORE="postgres_dev_test"
			export RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
        fi
        if [[ "${workflow_diagnostics_dev}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV="1"
        fi
        exec "${platform_wrapper}" serve
      ) >"${log_dir}/platform.out.log" 2>"${log_dir}/platform.err.log" &
      spawned_pids+=("$!")
    fi
  fi

  if port_is_open "${frontend_host}" "${frontend_port}"; then
    step "Web port ${frontend_port} is already open; reusing it if probes pass."
  else
    step "Starting web (${mode}). Logs: ${log_dir}/web.out.log ; ${log_dir}/web.err.log"
    (
      cd "${web_dir}"
      if [[ "${mode}" == "dev-live" ]]; then
        export VITE_RADISHMIND_READ_SOURCE="dev-live-http"
        export VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL="${backend_url%/}"
        export VITE_RADISHMIND_DEV_READ_TENANT_REF="${tenant_ref}"
        export VITE_RADISHMIND_DEV_READ_SUBJECT_REF="${subject_ref}"
        if [[ "${saved_draft_enabled}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE="dev-saved-draft-http"
          export VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE="dev-workflow-executor-http"
        fi
        if [[ "${workflow_diagnostics_dev}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV="true"
        fi
      else
        unset VITE_RADISHMIND_READ_SOURCE
        unset VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE
      fi
      exec npm run dev
    ) >"${log_dir}/web.out.log" 2>"${log_dir}/web.err.log" &
    spawned_pids+=("$!")
  fi
fi

if [[ "${mode}" == "dev-live" ]]; then
  if ! wait_until "backend healthz" probe_json "${healthz_url}" ""; then
    show_failure_help "backend healthz probe failed"
    exit 1
  fi
  if ! wait_until "dev-live read route CORS" probe_cors "${tenant_summary_url}" "${frontend_origin}"; then
    show_failure_help "dev-live read route CORS probe failed"
    exit 1
  fi
  if ! wait_until "dev-live read routes" probe_control_plane_read_routes "${backend_url}" "${tenant_ref}" "${subject_ref}"; then
    show_failure_help "dev-live read route probe failed"
    exit 1
  fi
  if [[ "${saved_draft_enabled}" -eq 1 ]] && ! wait_until \
    "Saved Draft dev route" \
    probe_saved_workflow_draft_route \
    "${backend_url}" \
    "${tenant_ref}" \
    "${subject_ref}" \
    "${saved_draft_workspace_id}" \
    "${saved_draft_application_id}"; then
    show_failure_help "Saved Draft dev route probe failed"
    exit 1
  fi
  if [[ "${saved_draft_enabled}" -eq 1 ]] && ! wait_until \
    "Workflow Executor dev route" \
    probe_workflow_executor_route \
    "${backend_url}" \
    "${tenant_ref}" \
    "${subject_ref}" \
    "${saved_draft_workspace_id}" \
    "${saved_draft_application_id}"; then
    show_failure_help "Workflow Executor dev route probe failed"
    exit 1
  fi
fi

if ! wait_until "frontend web" probe_page "${frontend_url}"; then
  show_failure_help "frontend web probe failed"
  exit 1
fi

step "RadishMind web is ready: ${frontend_url}"
if [[ "${mode}" == "dev-live" ]]; then
  step "Dev-live read backend passed: ${healthz_url} ; ${tenant_summary_url}"
  if [[ "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
    step "Saved Draft PostgreSQL dev/test read/write mode passed for ${saved_draft_workspace_id}/${saved_draft_application_id}."
  elif [[ "${saved_draft_dev}" -eq 1 ]]; then
    step "Saved Draft memory-dev read/write mode passed for ${saved_draft_workspace_id}/${saved_draft_application_id}."
  fi
fi
step "This is a dev-only launcher, not a production supervisor. Controlled executor v0 is dev-only; production auth, secret resolution, unrestricted tools, confirmation commit, writeback and replay remain disabled."

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
