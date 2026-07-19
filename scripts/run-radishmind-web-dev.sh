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
gateway_request_postgres_dev_test=0
application_draft_dev=0
application_publish_dev=0
application_publish_postgres_dev_test=0
application_catalog_postgres_dev_test=0
api_key_local_product=0
workflow_http_tool_local_product=0
workflow_rag_dev=0
workflow_rag_promotion_local_product=0
workflow_rag_application_local_product=0
workflow_definition_local_product=0
workflow_definition_postgres_dev_test=0
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
  --gateway-request-postgres-dev-test
                           Enable durable dev/test Gateway Request History and the scoped Gateway Playground.
  --application-draft-dev Enable the explicit memory-dev Application Configuration Draft path.
  --application-publish-dev
                           Enable memory-dev Application Draft and Publish Candidate review.
  --application-publish-postgres-dev-test
                           Enable PostgreSQL dev/test Application Draft and Publish Candidate review.
  --application-catalog-postgres-dev-test
                           Enable the PostgreSQL dev/test Application Catalog lifecycle UI.
  --api-key-local-product  Enable the SQLite local-product Application/API key/Playground chain.
  --workflow-http-tool-local-product
                           Enable the SQLite local-product Workflow HTTP Tool chain.
  --workflow-rag-dev       Enable the Workflow RAG snapshot, exact draft, retrieval execution, and Run History chain.
  --workflow-rag-promotion-local-product
                           Enable the SQLite evaluation, promotion, draft binding, and publish review chain without retrieval execution.
  --workflow-rag-application-local-product
                           Enable the SQLite application RAG assignment, API key, invocation, Run History, and evaluation chain.
  --workflow-definition-local-product
                           Enable the SQLite Saved Draft → candidate → review → activation → v5 run chain.
  --workflow-definition-postgres-dev-test
                           Enable the same chain with PostgreSQL dev/test repositories.
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
    --gateway-request-postgres-dev-test)
      gateway_request_postgres_dev_test=1
      shift
      ;;
    --application-draft-dev)
      application_draft_dev=1
      shift
      ;;
    --application-publish-dev)
      application_publish_dev=1
      shift
      ;;
    --application-publish-postgres-dev-test)
      application_publish_postgres_dev_test=1
      shift
      ;;
    --application-catalog-postgres-dev-test)
      application_catalog_postgres_dev_test=1
      shift
      ;;
    --api-key-local-product)
      api_key_local_product=1
      shift
      ;;
    --workflow-http-tool-local-product)
      workflow_http_tool_local_product=1
      shift
      ;;
    --workflow-rag-dev)
      workflow_rag_dev=1
      shift
      ;;
    --workflow-rag-promotion-local-product)
      workflow_rag_promotion_local_product=1
      shift
      ;;
    --workflow-rag-application-local-product)
      workflow_rag_application_local_product=1
      shift
      ;;
    --workflow-definition-local-product)
      workflow_definition_local_product=1
      shift
      ;;
    --workflow-definition-postgres-dev-test)
      workflow_definition_postgres_dev_test=1
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

if [[ "${workflow_definition_postgres_dev_test}" -eq 1 ]]; then
  saved_draft_postgres_dev_test=1
  application_catalog_postgres_dev_test=1
fi
workflow_definition_enabled=0
if [[ "${workflow_definition_local_product}" -eq 1 || "${workflow_definition_postgres_dev_test}" -eq 1 ]]; then
  workflow_definition_enabled=1
fi

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
if [[ "${gateway_request_postgres_dev_test}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--gateway-request-postgres-dev-test requires --mode dev-live" >&2
  exit 2
fi
if [[ "${application_draft_dev}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--application-draft-dev requires --mode dev-live" >&2
  exit 2
fi
if [[ "${application_publish_dev}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--application-publish-dev requires --mode dev-live" >&2
  exit 2
fi
if [[ "${application_publish_postgres_dev_test}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--application-publish-postgres-dev-test requires --mode dev-live" >&2
  exit 2
fi
if [[ "${application_catalog_postgres_dev_test}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--application-catalog-postgres-dev-test requires --mode dev-live" >&2
  exit 2
fi
if [[ "${api_key_local_product}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--api-key-local-product requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_http_tool_local_product}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--workflow-http-tool-local-product requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_rag_dev}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--workflow-rag-dev requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_rag_promotion_local_product}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--workflow-rag-promotion-local-product requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_rag_application_local_product}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "--workflow-rag-application-local-product requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_definition_enabled}" -eq 1 && "${mode}" != "dev-live" ]]; then
  echo "Workflow Definition local product requires --mode dev-live" >&2
  exit 2
fi
if [[ "${workflow_definition_local_product}" -eq 1 && "${workflow_definition_postgres_dev_test}" -eq 1 ]]; then
  echo "Choose either --workflow-definition-local-product or --workflow-definition-postgres-dev-test" >&2
  exit 2
fi
if [[ "${application_publish_dev}" -eq 1 && "${application_publish_postgres_dev_test}" -eq 1 ]]; then
  echo "Choose either --application-publish-dev or --application-publish-postgres-dev-test" >&2
  exit 2
fi
if [[ "${application_draft_dev}" -eq 1 && "${application_publish_postgres_dev_test}" -eq 1 ]]; then
  echo "--application-draft-dev cannot be combined with --application-publish-postgres-dev-test" >&2
  exit 2
fi

explicit_component_mode=0
if [[ "${saved_draft_dev}" -eq 1 || "${saved_draft_postgres_dev_test}" -eq 1 ||
  "${gateway_request_postgres_dev_test}" -eq 1 || "${application_draft_dev}" -eq 1 ||
  "${application_publish_dev}" -eq 1 || "${application_publish_postgres_dev_test}" -eq 1 ||
  "${application_catalog_postgres_dev_test}" -eq 1 ]]; then
  explicit_component_mode=1
fi
if [[ "${api_key_local_product}" -eq 1 && "${explicit_component_mode}" -eq 1 ]]; then
  echo "--api-key-local-product cannot be combined with explicit memory/PostgreSQL component modes" >&2
  exit 2
fi
if [[ "${workflow_http_tool_local_product}" -eq 1 && "${explicit_component_mode}" -eq 1 ]]; then
  echo "--workflow-http-tool-local-product cannot be combined with explicit memory/PostgreSQL component modes" >&2
  exit 2
fi
if [[ "${workflow_rag_promotion_local_product}" -eq 1 && "${explicit_component_mode}" -eq 1 ]]; then
  echo "--workflow-rag-promotion-local-product cannot be combined with explicit memory/PostgreSQL component modes" >&2
  exit 2
fi
if [[ "${workflow_rag_application_local_product}" -eq 1 && "${explicit_component_mode}" -eq 1 ]]; then
  echo "--workflow-rag-application-local-product cannot be combined with explicit memory/PostgreSQL component modes" >&2
  exit 2
fi
if [[ "${workflow_definition_local_product}" -eq 1 && "${explicit_component_mode}" -eq 1 ]]; then
  echo "--workflow-definition-local-product cannot be combined with explicit memory/PostgreSQL component modes" >&2
  exit 2
fi
platform_profile="local-product"
if [[ "${explicit_component_mode}" -eq 1 ]]; then
  platform_profile="configured"
fi

saved_draft_enabled=0
if [[ "${saved_draft_dev}" -eq 1 || "${saved_draft_postgres_dev_test}" -eq 1 ||
  "${workflow_http_tool_local_product}" -eq 1 || "${workflow_rag_dev}" -eq 1 ||
  "${workflow_rag_application_local_product}" -eq 1 || "${workflow_definition_enabled}" -eq 1 ]]; then
  saved_draft_enabled=1
fi

workflow_rag_snapshot_enabled=0
if [[ "${platform_profile}" == "local-product" || "${saved_draft_enabled}" -eq 1 ]]; then
  workflow_rag_snapshot_enabled=1
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
		export RADISHMIND_APPLICATION_DRAFT_DEV_HTTP="1"
		export RADISHMIND_APPLICATION_DRAFT_DEV_WRITE="1"
		export RADISHMIND_APPLICATION_DRAFT_STORE="postgres_dev_test"
		export RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL="${database_url}"
		export RADISHMIND_WORKFLOW_RUN_STORE="postgres_dev_test"
		export RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL="${database_url}"
		export RADISHMIND_GATEWAY_REQUEST_STORE="postgres_dev_test"
		export RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL="${database_url}"
		export RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV="1"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP="1"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE="1"
		export RADISHMIND_APPLICATION_PUBLISH_STORE="postgres_dev_test"
		export RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL="${database_url}"
		export RADISHMIND_APPLICATION_CATALOG_DEV_HTTP="1"
		export RADISHMIND_APPLICATION_CATALOG_DEV_WRITE="1"
		export RADISHMIND_APPLICATION_CATALOG_STORE="postgres_dev_test"
		export RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL="${database_url}"
    cd "${platform_dir}"
		go run ./cmd/radishmind-workflow-draft-migrate status >/dev/null || return
		go run ./cmd/radishmind-application-draft-migrate status >/dev/null || return
		go run ./cmd/radishmind-workflow-run-migrate status >/dev/null || return
		go run ./cmd/radishmind-gateway-request-migrate status >/dev/null || return
		if [[ "${application_publish_postgres_dev_test}" -eq 1 ]]; then
			go run ./cmd/radishmind-application-publish-migrate status >/dev/null || return
		fi
		if [[ "${application_catalog_postgres_dev_test}" -eq 1 ]]; then
			go run ./cmd/radishmind-application-catalog-migrate status >/dev/null || return
		fi
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
  "${python_bin}" - "$base_url" "$tenant" "$subject" "${application_catalog_postgres_dev_test}" "${saved_draft_workspace_id}" "${platform_profile}" <<'PY'
import json
import sys
from urllib.parse import quote
from urllib.request import Request, urlopen

base_url = sys.argv[1].rstrip("/")
tenant = sys.argv[2]
subject = sys.argv[3]
catalog_enabled = sys.argv[4] == "1"
workspace_id = sys.argv[5]
local_product = sys.argv[6] == "local-product"
application_route = "/v1/user-workspace/applications"
if catalog_enabled or local_product:
    application_route += f"?workspace_id={quote(workspace_id)}&lifecycle_state=active&limit=1"
api_key_route = "/v1/user-workspace/api-keys"
if local_product:
    api_key_route += f"?workspace_id={quote(workspace_id)}&limit=1"
routes = [
    f"/v1/control-plane/tenants/{quote(tenant)}/summary",
    application_route,
    api_key_route,
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

probe_gateway_api_key_mode() {
  local base_url="$1"
  "${python_bin}" - "$base_url" <<'PY'
import json
import sys
from urllib.error import HTTPError
from urllib.request import Request, urlopen

url = sys.argv[1].rstrip("/") + "/v1/models"
request = Request(url, headers={"Accept": "application/json", "X-Request-Id": "api-key-mode-probe"}, method="GET")
try:
    urlopen(request, timeout=5)
except HTTPError as error:
    document = json.loads(error.read().decode("utf-8"))
    if error.code != 401 or document.get("error", {}).get("code") != "api_key_missing":
        raise SystemExit(f"Gateway API key mode probe returned HTTP {error.code}: {document}")
else:
    raise SystemExit("Gateway accepted an unauthenticated model request; api_key_dev_test mode is not active")
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

probe_workflow_definition_route() {
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

base_url, tenant, subject, workspace_id, application_id = sys.argv[1:]
query = urlencode({"workspace_id": workspace_id, "application_id": application_id})
url = f"{base_url.rstrip('/')}/v1/user-workspace/workflow-definition-candidates?{query}"
request = Request(url, headers={
    "Accept": "application/json",
    "X-Request-Id": "dev-live-workflow-definition-probe",
    "X-RadishMind-Dev-Read-Identity": "dev-live-workflow-definition-probe",
    "X-RadishMind-Dev-Read-Tenant": tenant,
    "X-RadishMind-Dev-Read-Subject": subject,
    "X-RadishMind-Dev-Read-Scopes": "workflow_definitions:read",
    "X-RadishMind-Dev-Read-Audit": "audit_dev_live_workflow_definition_probe",
    "X-RadishMind-Dev-Workflow-Workspace": workspace_id,
    "X-RadishMind-Dev-Workflow-Application": application_id,
}, method="GET")
with urlopen(request, timeout=5) as response:
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
    document = json.loads(response.read().decode("utf-8"))
if document.get("failure_code") is not None or not isinstance(document.get("candidates"), list):
    raise SystemExit("Workflow Definition route did not return a successful candidates[] envelope")
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

probe_workflow_rag_execution_route() {
  local base_url="$1"
  local tenant="$2"
  local subject="$3"
  local workspace_id="$4"
  local application_id="$5"
  "${python_bin}" - "$base_url" "$tenant" "$subject" "$workspace_id" "$application_id" <<'PY'
import json
import sys
from urllib.request import Request, urlopen

base_url, tenant, subject, workspace_id, application_id = sys.argv[1:]
url = f"{base_url.rstrip('/')}/v1/user-workspace/workflow-drafts/draft_rag_probe_missing/retrieval-executions"
request = Request(
    url,
    data=json.dumps({
        "workspace_id": workspace_id,
        "application_id": application_id,
        "draft_version": 1,
        "input_text": "workflow RAG execution gate probe",
        "model": "",
        "temperature": None,
    }).encode("utf-8"),
    headers={
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-Request-Id": "dev-live-workflow-rag-execution-probe",
        "X-RadishMind-Dev-Read-Identity": "dev-live-workflow-rag-execution-probe",
        "X-RadishMind-Dev-Read-Tenant": tenant,
        "X-RadishMind-Dev-Read-Subject": subject,
        "X-RadishMind-Dev-Read-Scopes": "workflow_rag:execute,workflow_runs:execute,workflow_drafts:read,workflow_rag_snapshots:read",
        "X-RadishMind-Dev-Read-Audit": "audit_dev_live_workflow_rag_execution_probe",
        "X-RadishMind-Dev-Workflow-Workspace": workspace_id,
        "X-RadishMind-Dev-Workflow-Application": application_id,
    },
    method="POST",
)
with urlopen(request, timeout=5) as response:
    document = json.loads(response.read().decode("utf-8"))
    if response.status < 200 or response.status >= 300:
        raise SystemExit(f"Unexpected HTTP status {response.status} from {url}")
if document.get("failure_code") != "workflow_rag_draft_ineligible" or document.get("run") is not None:
    raise SystemExit(f"Workflow RAG execution probe returned incompatible evidence: {document.get('failure_code')}")
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
    echo "- Workflow mode: choose --saved-draft-dev, --saved-draft-postgres-dev-test, or --workflow-http-tool-local-product so backend and web opt in together."
    echo "- Gateway Request History: use --gateway-request-postgres-dev-test after running the PostgreSQL check entry."
    echo "- API key local product: use --api-key-local-product by itself so SQLite lifecycle and api_key_dev_test auth stay aligned."
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
if [[ "${saved_draft_postgres_dev_test}" -eq 1 || "${gateway_request_postgres_dev_test}" -eq 1 || "${application_publish_postgres_dev_test}" -eq 1 || "${application_catalog_postgres_dev_test}" -eq 1 ]]; then
  require_command go
  saved_draft_database_url="$(build_saved_draft_database_url)"
  if ! probe_saved_draft_postgres_migration "${saved_draft_database_url}"; then
    show_failure_help "PostgreSQL dev/test migration preflight failed"
    exit 1
  fi
  step "PostgreSQL dev/test migration preflight passed."
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
        if [[ "${workflow_rag_snapshot_enabled}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV="1"
        fi
        if [[ "${workflow_rag_dev}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_RAG_EXECUTION_DEV="1"
        fi
        if [[ "${workflow_rag_promotion_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_RAG_EVALUATION_DEV="1"
          export RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV="1"
        fi
        if [[ "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_RAG_APPLICATION_INVOCATION_DEV="1"
        fi
        if [[ "${workflow_definition_enabled}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_DEFINITION_RELEASE_DEV="1"
          export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
        fi
        if [[ "${api_key_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export RADISHMIND_GATEWAY_AUTH_MODE="api_key_dev_test"
        fi
        if [[ "${application_publish_postgres_dev_test}" -eq 1 ]]; then
          export RADISHMIND_APPLICATION_DRAFT_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_DRAFT_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_DRAFT_STORE="postgres_dev_test"
          export RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
          export RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_PUBLISH_STORE="postgres_dev_test"
          export RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
        elif [[ "${application_publish_dev}" -eq 1 ]]; then
          export RADISHMIND_APPLICATION_DRAFT_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_DRAFT_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_DRAFT_STORE="memory_dev"
          export RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_PUBLISH_STORE="memory_dev"
        elif [[ "${application_draft_dev}" -eq 1 ]]; then
          export RADISHMIND_APPLICATION_DRAFT_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_DRAFT_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_DRAFT_STORE="memory_dev"
        fi
        if [[ "${saved_draft_dev}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
          export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
          export RADISHMIND_WORKFLOW_TOOL_ACTION_DEV="1"
          export RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="memory_dev"
        elif [[ "${saved_draft_postgres_dev_test}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE="1"
          export RADISHMIND_WORKFLOW_EXECUTOR_DEV="1"
          export RADISHMIND_WORKFLOW_TOOL_ACTION_DEV="1"
          export RADISHMIND_WORKFLOW_HTTP_TOOL_EXECUTION_DEV="1"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE="postgres_dev_test"
          export RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
			export RADISHMIND_WORKFLOW_RUN_STORE="postgres_dev_test"
			export RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
        fi
        if [[ "${workflow_diagnostics_dev}" -eq 1 ]]; then
          export RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV="1"
        fi
        if [[ "${gateway_request_postgres_dev_test}" -eq 1 ]]; then
          export RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV="1"
          export RADISHMIND_GATEWAY_REQUEST_STORE="postgres_dev_test"
          export RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
        fi
        if [[ "${application_catalog_postgres_dev_test}" -eq 1 ]]; then
          export RADISHMIND_APPLICATION_CATALOG_DEV_HTTP="1"
          export RADISHMIND_APPLICATION_CATALOG_DEV_WRITE="1"
          export RADISHMIND_APPLICATION_CATALOG_STORE="postgres_dev_test"
          export RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL="${saved_draft_database_url}"
        fi
        exec "${platform_wrapper}" --profile "${platform_profile}" serve
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
        if [[ "${application_draft_dev}" -eq 1 || "${application_publish_dev}" -eq 1 || "${application_publish_postgres_dev_test}" -eq 1 || "${api_key_local_product}" -eq 1 || "${workflow_rag_promotion_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE="dev-application-draft-http"
          export VITE_RADISHMIND_APPLICATION_DRAFT_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${application_publish_dev}" -eq 1 || "${application_publish_postgres_dev_test}" -eq 1 || "${api_key_local_product}" -eq 1 || "${workflow_rag_promotion_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE="dev-application-publish-http"
          export VITE_RADISHMIND_APPLICATION_PUBLISH_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${application_catalog_postgres_dev_test}" -eq 1 || "${api_key_local_product}" -eq 1 || "${workflow_http_tool_local_product}" -eq 1 || "${workflow_rag_dev}" -eq 1 || "${workflow_rag_promotion_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 || "${workflow_definition_enabled}" -eq 1 ]]; then
          export VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE="dev-application-catalog-http"
          export VITE_RADISHMIND_APPLICATION_CATALOG_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${api_key_local_product}" -eq 1 || "${workflow_http_tool_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE="dev-api-key-lifecycle-http"
          export VITE_RADISHMIND_API_KEY_LIFECYCLE_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${api_key_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_GATEWAY_AUTH_MODE="api_key_dev_test"
        fi
        if [[ "${saved_draft_enabled}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE="dev-saved-draft-http"
          export VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE="dev-workflow-executor-http"
          export VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE="dev-workflow-http-tool-http"
          export VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SCOPE_GRANTS="workflow_drafts:read,workflow_tool_actions:plan,workflow_tool_actions:read,workflow_tool_actions:confirm,workflow_tool_actions:execute,workflow_runs:execute"
        fi
        if [[ "${workflow_definition_enabled}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_SOURCE="dev-workflow-definition-promotion-http"
          export VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${workflow_rag_snapshot_enabled}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_RAG_SOURCE="dev-workflow-rag-http"
          export VITE_RADISHMIND_WORKFLOW_RAG_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_WORKFLOW_RAG_WORKSPACE_ID="${saved_draft_workspace_id}"
          if [[ "${workflow_rag_dev}" -eq 1 ]]; then
            export VITE_RADISHMIND_WORKFLOW_RAG_SCOPES="workflow_rag_snapshots:read,workflow_rag_snapshots:write,workflow_rag_snapshots:archive,workflow_rag:execute,workflow_runs:execute,workflow_drafts:read"
          else
            export VITE_RADISHMIND_WORKFLOW_RAG_SCOPES="workflow_rag_snapshots:read,workflow_rag_snapshots:write,workflow_rag_snapshots:archive"
          fi
        fi
        if [[ "${workflow_rag_promotion_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SOURCE="dev-workflow-rag-evaluation-http"
          export VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SCOPES="workflow_rag_evaluation_datasets:read,workflow_rag_evaluation_datasets:read_content,workflow_rag_evaluation_datasets:write,workflow_rag_evaluation_datasets:review,workflow_rag_evaluation_datasets:archive,workflow_rag_snapshots:read"
          export VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SOURCE="dev-workflow-rag-promotion-http"
          export VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SCOPES="workflow_rag_promotions:read,workflow_rag_promotions:write,workflow_rag_promotions:review,workflow_rag_evaluation_datasets:read,workflow_rag_snapshots:read,application_drafts:read"
        fi
        if [[ "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_SOURCE="dev-workflow-rag-application-runtime-http"
          export VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_WORKSPACE_ID="${saved_draft_workspace_id}"
        fi
        if [[ "${workflow_diagnostics_dev}" -eq 1 ]]; then
          export VITE_RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV="true"
        fi
        if [[ "${gateway_request_postgres_dev_test}" -eq 1 || "${api_key_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]]; then
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE="dev-gateway-request-history-http"
          export VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE="dev-gateway-playground-http"
          export VITE_RADISHMIND_GATEWAY_PLAYGROUND_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL="${RADISHMIND_PLATFORM_MODEL:-radishmind-local-dev}"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL="${backend_url%/}"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF="${tenant_ref}"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID="${saved_draft_workspace_id}"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF="consumer_web_dev"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID="${saved_draft_application_id}"
          export VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF="${subject_ref}"
        fi
      else
        unset VITE_RADISHMIND_READ_SOURCE
        unset VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_HTTP_TOOL_SCOPE_GRANTS
        unset VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_WORKSPACE_ID
        unset VITE_RADISHMIND_WORKFLOW_RAG_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_RAG_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_RAG_WORKSPACE_ID
        unset VITE_RADISHMIND_WORKFLOW_RAG_SCOPES
        unset VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_RAG_EVALUATION_SCOPES
        unset VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SCOPES
        unset VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_SOURCE
        unset VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_BASE_URL
        unset VITE_RADISHMIND_WORKFLOW_RAG_APPLICATION_RUNTIME_WORKSPACE_ID
        unset VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE
        unset VITE_RADISHMIND_APPLICATION_DRAFT_BASE_URL
        unset VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID
        unset VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE
        unset VITE_RADISHMIND_APPLICATION_PUBLISH_BASE_URL
        unset VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID
        unset VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE
        unset VITE_RADISHMIND_APPLICATION_CATALOG_BASE_URL
        unset VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID
        unset VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE
        unset VITE_RADISHMIND_API_KEY_LIFECYCLE_BASE_URL
        unset VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID
        unset VITE_RADISHMIND_GATEWAY_AUTH_MODE
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SOURCE
        unset VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE
        unset VITE_RADISHMIND_GATEWAY_PLAYGROUND_BASE_URL
        unset VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID
        unset VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF
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
  if [[ "${api_key_local_product}" -eq 1 || "${workflow_rag_application_local_product}" -eq 1 ]] && ! wait_until "Gateway API key auth mode" probe_gateway_api_key_mode "${backend_url}"; then
    show_failure_help "Gateway api_key_dev_test mode probe failed"
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
  if [[ "${workflow_rag_dev}" -eq 1 ]] && ! wait_until \
    "Workflow RAG execution dev route" \
    probe_workflow_rag_execution_route \
    "${backend_url}" \
    "${tenant_ref}" \
    "${subject_ref}" \
    "${saved_draft_workspace_id}" \
    "${saved_draft_application_id}"; then
    show_failure_help "Workflow RAG execution dev route probe failed"
    exit 1
  fi
  if [[ "${workflow_definition_enabled}" -eq 1 ]] && ! wait_until \
    "Workflow Definition promotion dev route" \
    probe_workflow_definition_route \
    "${backend_url}" \
    "${tenant_ref}" \
    "${subject_ref}" \
    "${saved_draft_workspace_id}" \
    "${saved_draft_application_id}"; then
    show_failure_help "Workflow Definition promotion dev route probe failed"
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
  if [[ "${gateway_request_postgres_dev_test}" -eq 1 ]]; then
    step "Gateway Playground and Request History PostgreSQL dev/test mode enabled for ${saved_draft_workspace_id}/consumer_web_dev/${saved_draft_application_id}."
  fi
  if [[ "${api_key_local_product}" -eq 1 ]]; then
    step "API key SQLite local-product Web chain enabled for ${saved_draft_workspace_id}; raw credentials remain browser-memory only."
  fi
  if [[ "${workflow_http_tool_local_product}" -eq 1 ]]; then
    step "Workflow HTTP Tool SQLite local-product Web chain enabled for ${saved_draft_workspace_id}/${saved_draft_application_id}; approve and execute remain separate actions."
  fi
  if [[ "${workflow_rag_dev}" -eq 1 ]]; then
    step "Workflow RAG Web chain enabled for ${saved_draft_workspace_id}/${saved_draft_application_id}; execution is synchronous, metadata-only, and dev/test only."
  fi
  if [[ "${workflow_rag_promotion_local_product}" -eq 1 ]]; then
    step "Workflow RAG promotion SQLite local-product Web chain enabled for ${saved_draft_workspace_id}/${saved_draft_application_id}; approve, attach, and publish review remain separate actions."
  fi
  if [[ "${workflow_rag_application_local_product}" -eq 1 ]]; then
    step "Workflow RAG application SQLite local-product Web chain enabled for ${saved_draft_workspace_id}/${saved_draft_application_id}; assignment, API key handoff, invocation, history, and evaluation remain explicit dev/test actions."
  fi
  if [[ "${workflow_definition_enabled}" -eq 1 ]]; then
    definition_store="SQLite"
    if [[ "${workflow_definition_postgres_dev_test}" -eq 1 ]]; then
      definition_store="PostgreSQL dev/test"
    fi
    step "Workflow Definition ${definition_store} product chain enabled for ${saved_draft_workspace_id}/${saved_draft_application_id}; review, activation, execution, comparison, and evaluation remain explicit actions."
  fi
fi
step "This is a dev-only launcher, not a production supervisor. Controlled execution is dev-only; production auth, secret resolution, unrestricted tools, automatic confirmation, writeback and replay remain disabled."

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
