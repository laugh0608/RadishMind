#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-dev-live-consumer-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-fake-store-handler-implementation-v1": (
        "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json",
        "fake_store_handler_implemented",
    ),
    "control-plane-read-consumer-contract-v1": (
        "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json",
        "consumer_contract_defined",
    ),
    "control-plane-read-formal-ui-readiness-close-v1": (
        "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json",
        "formal_ui_readiness_closed",
    ),
}
EXPECTED_ROUTES = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_HEADERS = {
    "X-RadishMind-Dev-Read-Identity",
    "X-RadishMind-Dev-Read-Tenant",
    "X-RadishMind-Dev-Read-Subject",
    "X-RadishMind-Dev-Read-Scopes",
    "X-RadishMind-Dev-Read-Audit",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_admin_console_ready",
    "formal_user_workspace_complete",
    "production_api_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "repository_migration_ready",
    "repository_implementation_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "billing_ready",
    "cost_ledger_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read_text(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def assert_fixture(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_dev_live_consumer_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-dev-live-consumer-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(slice_info.get("status") == "dev_live_consumer_implemented", "unexpected slice status")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    boundary = fixture.get("consumer_boundary") or {}
    require(boundary.get("app_root") == "apps/radishmind-web/", "app root drifted")
    require(boundary.get("default_source") == "offline_fixture", "default source must stay offline fixture")
    require(boundary.get("optional_source") == "dev_live_http", "optional source drifted")
    require(boundary.get("backend_store") == "fixture-backed fake store", "backend store boundary drifted")
    require(boundary.get("production_api_consumer") is False, "must not claim production API consumer")
    require(boundary.get("database_backed") is False, "must not claim database-backed read")
    require(boundary.get("radish_oidc_backed") is False, "must not claim Radish OIDC-backed read")

    require(set(fixture.get("required_routes") or []) == EXPECTED_ROUTES, "required routes drifted")
    require(set(fixture.get("required_dev_headers") or []) == EXPECTED_HEADERS, "required dev headers drifted")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require(set(EXPECTED_DEPENDENCIES).issubset(declared), "fixture must declare all dependencies")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"dependency {dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"dependency {dependency_id} status drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    source_files = [
        "apps/radishmind-web/src/features/control-plane-read/devLiveReadConsumer.ts",
        "apps/radishmind-web/src/app/App.tsx",
        "contracts/typescript/control-plane-read-api.ts",
        "services/platform/internal/config/config.go",
        "services/platform/internal/httpapi/control_plane_read.go",
        "services/platform/internal/httpapi/control_plane_read_auth.go",
        "services/platform/internal/httpapi/server.go",
    ]
    source = "\n".join(read_text(path) for path in source_files)
    for literal in fixture.get("required_source_literals") or []:
        require(str(literal) in source, f"source missing required literal: {literal}")

    dev_consumer = read_text("apps/radishmind-web/src/features/control-plane-read/devLiveReadConsumer.ts")
    for literal in fixture.get("forbidden_source_literals") or []:
        require(
            str(literal) not in dev_consumer,
            f"devLiveReadConsumer.ts must not contain forbidden literal: {literal}",
        )

    require("mode: source === DEV_LIVE_SOURCE ? \"dev_live_http\" : \"offline_fixture\"" in dev_consumer, "live mode must be explicit opt-in")
    require("return {};" in dev_consumer, "offline mode must not fetch live routes")
    for header in EXPECTED_HEADERS:
        require(header in dev_consumer, f"dev consumer missing header {header}")

    app_source = read_text("apps/radishmind-web/src/app/App.tsx")
    require("initialControlPlaneReadDevLiveLoadState" in app_source, "App must surface data source state")
    require("buildAdminTenantOverviewViewModel(liveCollections" in app_source, "tenant page must accept live collection override")
    require("buildWorkspaceRunHistoryViewModel(liveCollections" in app_source, "run history page must accept live collection override")

    server_source = read_text("services/platform/internal/httpapi/server.go")
    require("http://127.0.0.1:4100" in server_source, "web dev origin must be allowed for dev live read")
    for header in EXPECTED_HEADERS:
        require(header in source, f"backend source missing dev header {header}")


def assert_tests(fixture: dict[str, Any]) -> None:
    test_source = "\n".join(
        read_text(path)
        for path in [
            "services/platform/internal/httpapi/control_plane_read_test.go",
            "services/platform/internal/httpapi/server_test.go",
        ]
    )
    for literal in fixture.get("required_test_literals") or []:
        require(str(literal) in test_source, f"tests missing literal: {literal}")


def assert_baseline_and_docs(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py", [])'
        in check_repo,
        "dev live consumer checker must be in check-repo.py",
    )

    doc_refs = fixture.get("required_doc_references") or {}
    require(isinstance(doc_refs, dict), "required_doc_references must be an object")
    for relative_path, required_literals in doc_refs.items():
        document = read_text(str(relative_path))
        for literal in required_literals:
            require(str(literal) in document, f"{relative_path} missing doc reference: {literal}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    commands = {str(item.get("command")) for item in fixture.get("testing_strategy") or [] if isinstance(item, dict)}
    expected_commands = {
        "./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-dev-live-consumer-v1.py",
        "npm run build",
        "./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check",
        "./scripts/check-repo.sh --fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture(fixture)
    assert_dependencies(fixture)
    assert_required_files(fixture)
    assert_source_boundaries(fixture)
    assert_tests(fixture)
    assert_testing_strategy(fixture)
    assert_baseline_and_docs(fixture)
    print("control-plane-read-dev-live-consumer-v1 check passed")


if __name__ == "__main__":
    main()
