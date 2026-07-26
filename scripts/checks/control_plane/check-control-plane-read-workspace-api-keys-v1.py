#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-workspace-api-keys-v1.json"
CONSUMER_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json"
FORMAL_UI_BOUNDARY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json"
FORMAL_UI_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-implementation-readiness-v1.json"
)
SHARED_SHELL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-shared-shell-v1.json"
ADMIN_TENANT_OVERVIEW_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-admin-tenant-overview-v1.json"
)
WORKSPACE_APPLICATIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-workspace-applications-v1.json"
)
APP_ROOT = REPO_ROOT / "apps/radishmind-web"
CONSOLE_PACKAGE_PATH = REPO_ROOT / "apps/radishmind-console/package.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_REQUIRED_STATES = {
    "ready",
    "empty",
    "denied",
    "stale",
    "partial_failure",
    "forbidden_projection",
}
EXPECTED_IMPLEMENTED_CAPABILITIES = {
    "workspace_api_keys_read_only_page",
    "api_key_summary_list_route_binding",
    "api_key_summary_collection_view_model",
    "workspace_api_keys_state_previews",
    "api_key_secret_value_hidden",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_write_or_execute_controls",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_admin_console_ready",
    "formal_user_workspace_complete",
    "real_api_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "repository_implementation_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "api_key_lifecycle_ready",
    "api_key_value_available",
    "api_key_hash_available",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_CONTRACT_SYMBOLS = {
    "CONTROL_PLANE_READ_ROUTES",
    "CONTROL_PLANE_READ_ROUTE_DEFINITIONS",
    "CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS",
    "APIKeySummary",
    "ControlPlaneReadCollectionViewModel",
    "ControlPlaneReadResponseByRoute",
    "controlPlaneReadResponseHasForbiddenOutput",
    "toControlPlaneReadCollectionViewModel",
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


def app_source_text() -> str:
    return "\n".join(path.read_text(encoding="utf-8") for path in sorted((APP_ROOT / "src").rglob("*")) if path.is_file())


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-consumer-contract-v1",
        "control-plane-read-formal-ui-boundary-v1",
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "control-plane-read-shared-shell-v1",
        "control-plane-read-admin-tenant-overview-v1",
        "control-plane-read-workspace-applications-v1",
    }
    require(expected.issubset(declared), "workspace api keys fixture must declare UI dependencies")

    dependency_statuses = {
        "control-plane-read-consumer-contract-v1": (
            CONSUMER_CONTRACT_FIXTURE_PATH,
            "consumer_contract_defined",
        ),
        "control-plane-read-formal-ui-boundary-v1": (
            FORMAL_UI_BOUNDARY_FIXTURE_PATH,
            "formal_ui_boundary_defined",
        ),
        "control-plane-read-formal-ui-implementation-readiness-v1": (
            FORMAL_UI_READINESS_FIXTURE_PATH,
            "formal_ui_implementation_readiness_defined",
        ),
        "control-plane-read-shared-shell-v1": (
            SHARED_SHELL_FIXTURE_PATH,
            "shared_read_shell_implemented",
        ),
        "control-plane-read-admin-tenant-overview-v1": (
            ADMIN_TENANT_OVERVIEW_FIXTURE_PATH,
            "admin_tenant_overview_implemented",
        ),
        "control-plane-read-workspace-applications-v1": (
            WORKSPACE_APPLICATIONS_FIXTURE_PATH,
            "workspace_applications_implemented",
        ),
    }
    for dependency_id, (path, expected_status) in dependency_statuses.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"dependency {dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"dependency {dependency_id} status drifted")


def assert_fixture(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_workspace_api_keys_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-workspace-api-keys-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(slice_info.get("status") == "workspace_api_keys_implemented", "workspace api keys status drifted")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")
    missing_capabilities = sorted(EXPECTED_IMPLEMENTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    require(set(fixture.get("required_states") or []) == EXPECTED_REQUIRED_STATES, "required UI states drifted")

    page = fixture.get("page") or {}
    require(page.get("id") == "workspace-api-keys", "page id drifted")
    require(page.get("route_id") == "api-key-summary-list-route", "page route id drifted")
    require(page.get("read_model") == "api-key-summary", "page read model drifted")
    require(page.get("required_scope") == "api_keys:read", "page scope drifted")
    require(page.get("status") == "offline_read_only_page_slice", "page status drifted")
    require(page.get("app_root") == "apps/radishmind-web/", "page app root drifted")
    require(page.get("local_ops_console_boundary") == "remain_local_ops_surface_only", "console boundary drifted")


def assert_app_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")

    package_json = load_json(APP_ROOT / "package.json")
    require(package_json.get("private") is True, "apps/radishmind-web package must remain private")
    scripts = package_json.get("scripts") or {}
    require(scripts.get("dev") == "vite --host 127.0.0.1 --port 4100", "web dev script drifted")
    require(scripts.get("build") == "tsc --noEmit && vite build", "web build script drifted")
    forbidden_scripts = sorted(set(fixture.get("forbidden_package_scripts") or []) & set(scripts))
    require(not forbidden_scripts, f"apps/radishmind-web must not add production scripts: {forbidden_scripts}")

    console_package = load_json(CONSOLE_PACKAGE_PATH)
    console_scripts = console_package.get("scripts") or {}
    forbidden_console_scripts = sorted(set(fixture.get("forbidden_package_scripts") or []) & set(console_scripts))
    require(not forbidden_console_scripts, f"apps/radishmind-console must remain local ops only: {forbidden_console_scripts}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    source = app_source_text()
    page_source = read_text("apps/radishmind-web/src/features/control-plane-read/workspaceApiKeys.ts")
    app_source = read_text("apps/radishmind-web/src/app/App.tsx")
    workspace_surface_source = read_text(
        "apps/radishmind-web/src/features/control-plane-read/applicationDevelopmentWorkspaceSurface.tsx"
    )
    lifecycle_panel_source = read_text(
        "apps/radishmind-web/src/features/control-plane-read/apiKeyLifecyclePanel.tsx"
    )
    for symbol in EXPECTED_CONTRACT_SYMBOLS:
        require(symbol in page_source, f"workspace api keys missing contract symbol: {symbol}")
    for state in EXPECTED_REQUIRED_STATES:
        require(state in page_source, f"workspace api keys missing state: {state}")
    for literal in (
        "buildWorkspaceApiKeysViewModel",
        "WorkspaceApiKeysViewModel",
        "canRequestLiveBackend:",
        "canMutate: false",
        "api-key-summary-list-route",
        "workspace-api-keys",
        "api_key_id",
        "owner_subject_ref",
        "expires_at",
        "last_used_at",
    ):
        require(literal in source, f"workspace api keys missing source literal: {literal}")
    for literal in (
        "ApplicationDevelopmentWorkspaceSurface",
        "offlineApiKeys={workspaceApiKeys}",
        "context={applicationDevelopmentWorkspaceContext}",
    ):
        require(
            literal in app_source,
            f"App.tsx missing application workspace ownership literal: {literal}",
        )
    for literal in (
        "APIKeyLifecyclePanel",
        'activeStage === "controlled_test"',
        "applicationActive={context.applicationActive}",
        "offlineView={offlineApiKeys}",
    ):
        require(
            literal in workspace_surface_source,
            f"application workspace surface missing API key lifecycle literal: {literal}",
        )
    for literal in (
        "OfflineAPIKeySummary",
        "WorkspaceApiKeysViewModel",
        "view.canRenderApiKeys",
        'id="workspace-api-keys"',
        "OfflineRow",
        "OfflineState",
        "OfflineMetric",
    ):
        require(
            literal in lifecycle_panel_source,
            f"API key lifecycle panel missing offline compatibility literal: {literal}",
        )
    for forbidden_literal in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden_literal) not in source, f"web source contains forbidden literal: {forbidden_literal}")
    for sensitive_literal in fixture.get("forbidden_sensitive_projection_literals") or []:
        require(str(sensitive_literal) not in page_source, f"workspace api keys source exposes sensitive key literal: {sensitive_literal}")
    for route_path in (
        "/v1/control-plane/tenants/{tenant_ref}/summary",
        "/v1/user-workspace/applications",
        "/v1/user-workspace/api-keys",
        "/v1/user-workspace/usage/quota-summary",
        "/v1/user-workspace/workflow-definitions",
        "/v1/user-workspace/runs",
        "/v1/control-plane/audit",
    ):
        exact_route_literal = re.compile(rf"(['\\\"])({re.escape(route_path)})\\1")
        require(
            exact_route_literal.search(source) is None,
            f"web source must not duplicate route path literal: {route_path}",
        )


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-workspace-api-keys-v1.py", [])'
        in check_repo,
        "check-repo.py must run workspace api keys checker",
    )
    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read_text(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_fixture(fixture)
    assert_app_files(fixture)
    assert_source_boundaries(fixture)
    assert_docs_and_fast_baseline(fixture)
    print("control plane read workspace api keys v1 checks passed.")


if __name__ == "__main__":
    main()
