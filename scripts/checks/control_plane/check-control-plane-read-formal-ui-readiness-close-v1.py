#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json"
APP_ROOT = REPO_ROOT / "apps/radishmind-web"
CONSOLE_PACKAGE_PATH = REPO_ROOT / "apps/radishmind-console/package.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-route-contract-v1": (
        "scripts/checks/fixtures/control-plane-read-route-contract-v1.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-response-fixtures-v1": (
        "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-negative-contract-v1": (
        "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-fake-store-handler-implementation-v1": (
        "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json",
        "fake_store_handler_implemented",
    ),
    "control-plane-read-auth-db-preconditions-v1": (
        "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json",
        "auth_db_preconditions_defined",
    ),
    "control-plane-read-consumer-contract-v1": (
        "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json",
        "consumer_contract_defined",
    ),
    "control-plane-read-formal-ui-boundary-v1": (
        "scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json",
        "formal_ui_boundary_defined",
    ),
    "control-plane-read-formal-ui-implementation-readiness-v1": (
        "scripts/checks/fixtures/control-plane-read-formal-ui-implementation-readiness-v1.json",
        "formal_ui_implementation_readiness_defined",
    ),
    "control-plane-read-shared-shell-v1": (
        "scripts/checks/fixtures/control-plane-read-shared-shell-v1.json",
        "shared_read_shell_implemented",
    ),
    "control-plane-read-admin-tenant-overview-v1": (
        "scripts/checks/fixtures/control-plane-read-admin-tenant-overview-v1.json",
        "admin_tenant_overview_implemented",
    ),
    "control-plane-read-workspace-applications-v1": (
        "scripts/checks/fixtures/control-plane-read-workspace-applications-v1.json",
        "workspace_applications_implemented",
    ),
    "control-plane-read-workspace-api-keys-v1": (
        "scripts/checks/fixtures/control-plane-read-workspace-api-keys-v1.json",
        "workspace_api_keys_implemented",
    ),
    "control-plane-read-workspace-usage-quota-v1": (
        "scripts/checks/fixtures/control-plane-read-workspace-usage-quota-v1.json",
        "workspace_usage_quota_implemented",
    ),
    "control-plane-read-workspace-workflow-definitions-v1": (
        "scripts/checks/fixtures/control-plane-read-workspace-workflow-definitions-v1.json",
        "workspace_workflow_definitions_implemented",
    ),
    "control-plane-read-workspace-run-history-v1": (
        "scripts/checks/fixtures/control-plane-read-workspace-run-history-v1.json",
        "workspace_run_history_implemented",
    ),
    "control-plane-read-admin-audit-log-v1": (
        "scripts/checks/fixtures/control-plane-read-admin-audit-log-v1.json",
        "admin_audit_log_implemented",
    ),
}

EXPECTED_SURFACES = {
    "admin-tenant-overview": {
        "surface": "Admin Control Plane",
        "route_id": "tenant-summary-route",
        "route_key": "tenantSummary",
        "read_model": "tenant-summary",
        "required_scope": "tenant:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/adminTenantOverview.ts",
        "builder": "buildAdminTenantOverviewViewModel",
        "summary_binding": "tenant",
    },
    "admin-audit-log": {
        "surface": "Admin Control Plane",
        "route_id": "audit-summary-list-route",
        "route_key": "audit",
        "read_model": "audit-summary",
        "required_scope": "audit:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/adminAuditLog.ts",
        "builder": "buildAdminAuditLogViewModel",
        "summary_binding": "auditEvents",
    },
    "workspace-applications": {
        "surface": "User Workspace",
        "route_id": "application-summary-list-route",
        "route_key": "applications",
        "read_model": "application-summary",
        "required_scope": "applications:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workspaceApplications.ts",
        "builder": "buildWorkspaceApplicationsViewModel",
        "summary_binding": "applications",
    },
    "workspace-api-keys": {
        "surface": "User Workspace",
        "route_id": "api-key-summary-list-route",
        "route_key": "apiKeys",
        "read_model": "api-key-summary",
        "required_scope": "api_keys:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workspaceApiKeys.ts",
        "builder": "buildWorkspaceApiKeysViewModel",
        "summary_binding": "apiKeys",
    },
    "workspace-usage-quota": {
        "surface": "User Workspace",
        "route_id": "quota-summary-route",
        "route_key": "quotaSummary",
        "read_model": "quota-summary",
        "required_scope": "usage:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workspaceUsageQuota.ts",
        "builder": "buildWorkspaceUsageQuotaViewModel",
        "summary_binding": "quota",
    },
    "workspace-workflow-definitions": {
        "surface": "User Workspace",
        "route_id": "workflow-definition-summary-list-route",
        "route_key": "workflowDefinitions",
        "read_model": "workflow-definition-summary",
        "required_scope": "applications:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workspaceWorkflowDefinitions.ts",
        "builder": "buildWorkspaceWorkflowDefinitionsViewModel",
        "summary_binding": "workflowDefinitions",
    },
    "workspace-run-history": {
        "surface": "User Workspace",
        "route_id": "run-record-summary-list-route",
        "route_key": "runs",
        "read_model": "run-record-summary",
        "required_scope": "runs:read",
        "source_file": "apps/radishmind-web/src/features/control-plane-read/workspaceRunHistory.ts",
        "builder": "buildWorkspaceRunHistoryViewModel",
        "summary_binding": "runs",
    },
}

EXPECTED_REQUIRED_STATES = {
    "ready",
    "empty",
    "denied",
    "stale",
    "partial_failure",
    "forbidden_projection",
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
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_SHARED_SYMBOLS = {
    "CONTROL_PLANE_READ_ROUTES",
    "CONTROL_PLANE_READ_ROUTE_DEFINITIONS",
    "CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS",
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
    require(set(EXPECTED_DEPENDENCIES).issubset(declared), "readiness close fixture must declare all UI dependencies")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(REPO_ROOT / relative_path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"dependency {dependency_id} id drifted")
        require(slice_info.get("status") == expected_status, f"dependency {dependency_id} status drifted")


def assert_fixture(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_formal_ui_readiness_close_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-formal-ui-readiness-close-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(slice_info.get("status") == "formal_ui_readiness_closed", "formal UI readiness close status drifted")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")
    require(set(fixture.get("required_states") or []) == EXPECTED_REQUIRED_STATES, "required UI states drifted")
    require(set(fixture.get("required_shared_symbols") or []) == EXPECTED_SHARED_SYMBOLS, "shared symbols drifted")

    close_scope = fixture.get("close_scope") or {}
    require(close_scope.get("status") == "read_side_formal_ui_surface_matrix_closed", "close scope status drifted")
    require(close_scope.get("app_root") == "apps/radishmind-web/", "app root drifted")
    require(close_scope.get("local_ops_console_boundary") == "remain_local_ops_surface_only", "console boundary drifted")

    surfaces = fixture.get("surface_matrix") or []
    require(isinstance(surfaces, list), "surface_matrix must be a list")
    indexed = {str(surface.get("page_id")): surface for surface in surfaces if isinstance(surface, dict)}
    require(set(indexed) == set(EXPECTED_SURFACES), "surface matrix page ids drifted")
    for page_id, expected in EXPECTED_SURFACES.items():
        surface = indexed[page_id]
        for key, expected_value in expected.items():
            require(surface.get(key) == expected_value, f"{page_id} {key} drifted")
        require(surface.get("render_anchor") == page_id, f"{page_id} render anchor drifted")
        require(surface.get("request_ref") == "requestId", f"{page_id} request ref drifted")
        require(surface.get("audit_ref") == "auditRef", f"{page_id} audit ref drifted")


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
    app_source = read_text("apps/radishmind-web/src/app/App.tsx")
    shell_source = read_text("apps/radishmind-web/src/features/control-plane-read/readShell.ts")
    for literal in fixture.get("forbidden_source_literals") or []:
        require(str(literal) not in source, f"apps/radishmind-web must not contain source literal: {literal}")
    for literal in fixture.get("forbidden_control_literals") or []:
        require(str(literal) not in source, f"apps/radishmind-web must not contain control literal: {literal}")

    require("buildControlPlaneReadShellViewModel" in app_source, "App must render shared read shell")
    require("id=\"routes\"" in app_source, "App must keep route catalog section")
    require("id=\"states\"" in app_source, "App must keep shared states section")
    require("id=\"guard\"" in app_source, "App must keep forbidden output guard section")

    for page_id, expected in EXPECTED_SURFACES.items():
        page_source = read_text(expected["source_file"])
        for symbol in EXPECTED_SHARED_SYMBOLS:
            require(symbol in page_source, f"{page_id} missing shared symbol: {symbol}")
        for state in EXPECTED_REQUIRED_STATES:
            require(state in page_source or state in shell_source, f"{page_id} missing shared state coverage: {state}")
        for literal in (
            expected["builder"],
            expected["route_id"],
            expected["read_model"],
            expected["required_scope"],
            f"CONTROL_PLANE_READ_ROUTES.{expected['route_key']}",
            "canRequestLiveBackend: false",
            "canMutate: false",
            "request_id",
            "audit_ref",
            "failure_code",
            "controlPlaneReadResponseHasForbiddenOutput",
        ):
            require(literal in page_source, f"{page_id} missing source literal: {literal}")
        for literal in (expected["builder"], page_id, expected["summary_binding"]):
            require(literal in app_source, f"App missing rendered binding for {page_id}: {literal}")


def assert_docs_and_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-formal-ui-readiness-close-v1.py", [])'
        in check_repo,
        "formal UI readiness close checker must be in check-repo.py",
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
        "./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-formal-ui-readiness-close-v1.py",
        "npm run build",
        "./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check",
        "macOS/Linux/WSL: ./scripts/check-repo-fast.sh; Windows/PowerShell: pwsh ./scripts/check-repo.ps1 -Fast",
    }
    missing_commands = sorted(expected_commands - commands)
    require(not missing_commands, f"missing testing strategy commands: {missing_commands}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_fixture(fixture)
    assert_dependencies(fixture)
    assert_app_files(fixture)
    assert_source_boundaries(fixture)
    assert_testing_strategy(fixture)
    assert_docs_and_baseline(fixture)
    print("control-plane-read-formal-ui-readiness-close-v1 check passed")


if __name__ == "__main__":
    main()
