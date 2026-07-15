#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json"
PRODUCT_SURFACE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
AUTH_DB_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
)
CONSUMER_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json"
TS_CONTRACT_PATH = REPO_ROOT / "contracts/typescript/control-plane-read-api.ts"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_ROUTE_IDS = {
    "tenant-summary-route",
    "application-summary-list-route",
    "api-key-summary-list-route",
    "quota-summary-route",
    "workflow-definition-summary-list-route",
    "run-record-summary-list-route",
    "audit-summary-list-route",
}
EXPECTED_PAGE_IDS = {
    "admin-tenant-overview",
    "admin-audit-log",
    "workspace-applications",
    "workspace-api-keys",
    "workspace-usage-quota",
    "workspace-workflow-definitions",
    "workspace-run-history",
}
EXPECTED_REQUIRED_VIEW_STATES = {
    "loading",
    "ready",
    "empty",
    "denied",
    "stale",
    "partial_failure",
    "forbidden_projection",
}
EXPECTED_READ_ONLY_INVARIANTS = {
    "all_pages_consume_control_plane_read_api_view_models",
    "all_pages_keep_read_only_commands",
    "denied_views_render_no_items",
    "stale_views_may_only_preserve_previous_read_only_data",
    "partial_failure_views_keep_failure_code_and_audit_ref",
    "forbidden_projection_blocks_sensitive_fields",
    "api_key_views_never_reveal_key_value_or_hash",
    "run_history_views_never_execute_or_replay_runs",
    "workflow_definition_views_never_create_update_or_execute_workflows",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "formal_user_workspace_ui_ready",
    "production_admin_console_ready",
    "react_ui_implementation_ready",
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
EXPECTED_FORBIDDEN_SCOPE = {
    "formal React implementation",
    "production admin console",
    "formal user workspace UI",
    "real API consumer",
    "database schema",
    "database migration",
    "database query implementation",
    "repository implementation",
    "Radish OIDC integration",
    "auth middleware implementation",
    "tenant or user CRUD",
    "API key generation",
    "API key reveal",
    "quota enforcement",
    "rate limiting implementation",
    "billing ledger implementation",
    "workflow builder",
    "workflow executor",
    "confirmation flow wiring",
    "business writeback",
    "replay execution",
    "production ready",
}
FORBIDDEN_ALLOWED_COMMAND_TOKENS = {
    "create",
    "edit",
    "delete",
    "issue",
    "reveal",
    "rotate",
    "revoke",
    "change",
    "execute",
    "enforce",
    "write",
    "confirm",
    "start",
    "cancel",
    "resume",
    "replay",
    "materialize",
    "export_raw",
}
REQUIRED_TS_STOP_LINES = {
    "databaseBacked: false",
    "formalUiReady: false",
    "canMutate: false",
    "canExecuteWorkflow: false",
    "canWriteBusinessTruth: false",
    "canRevealSecrets: false",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
        "不直接实现 OIDC、数据库、API key / quota、workflow executor、confirmation、writeback 或 replay",
    ],
    "docs/contracts/README.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
        "formal UI",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
    ],
    "docs/radishmind-integration-contracts.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "contracts/typescript/control-plane-read-api.ts",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "Control Plane Read Formal UI Boundary",
    ],
    "docs/task-cards/control-plane-read-formal-ui-boundary-v1-plan.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "正式 UI 边界",
        "ControlPlaneReadCollectionViewModel",
    ],
    "scripts/README.md": [
        "check-control-plane-read-formal-ui-boundary-v1.py",
        "control-plane-read-formal-ui-boundary-v1.json",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-formal-ui-boundary-v1",
        "control-plane-read-formal-ui-boundary-v1.json",
        "check-control-plane-read-formal-ui-boundary-v1.py",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def read(relative_path: str) -> str:
    return (REPO_ROOT / relative_path).read_text(encoding="utf-8")


def route_contracts_by_id() -> dict[str, dict[str, Any]]:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    routes = {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    require(set(routes) == EXPECTED_ROUTE_IDS, "route contract ids drifted")
    return routes


def response_forbidden_output_keys() -> set[str]:
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    keys = {str(item) for item in response_fixture.get("forbidden_output_keys") or []}
    require(keys, "response fixture must declare forbidden output keys")
    return keys


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "product-surface-v1-boundary",
        "control-plane-read-route-contract-v1",
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-negative-contract-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "control-plane-read-auth-db-preconditions-v1",
        "control-plane-read-consumer-contract-v1",
    }
    require(expected.issubset(declared), "formal UI boundary fixture must declare every dependency")

    expected_statuses = {
        "product-surface-v1-boundary": (PRODUCT_SURFACE_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-response-fixtures-v1": (RESPONSE_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-fake-store-handler-implementation-v1": (
            FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH,
            "fake_store_handler_implemented",
        ),
        "control-plane-read-auth-db-preconditions-v1": (
            AUTH_DB_PRECONDITIONS_FIXTURE_PATH,
            "auth_db_preconditions_defined",
        ),
        "control-plane-read-consumer-contract-v1": (CONSUMER_CONTRACT_FIXTURE_PATH, "consumer_contract_defined"),
    }
    for expected_id, (path, expected_status) in expected_statuses.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"dependency {expected_id} status must be {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_formal_ui_boundary_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-formal-ui-boundary-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "formal_ui_boundary_defined",
        "formal UI boundary must only define a boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_ui_contract_source(fixture: dict[str, Any]) -> None:
    source = fixture.get("ui_contract_source") or {}
    require(source.get("typescript_contract") == "contracts/typescript/control-plane-read-api.ts", "TS source drifted")
    require(source.get("consumer_smoke") == "scripts/run-control-plane-read-consumer-smoke.py --check", "smoke source drifted")
    require(source.get("view_model") == "ControlPlaneReadCollectionViewModel", "view model source drifted")
    require(source.get("route_catalog") == "CONTROL_PLANE_READ_ROUTES", "route catalog source drifted")
    require(source.get("status") == "design_boundary_only", "formal UI source must stay design boundary only")

    ts_content = TS_CONTRACT_PATH.read_text(encoding="utf-8")
    for literal in REQUIRED_TS_STOP_LINES:
        require(literal in ts_content, f"TypeScript contract missing stop-line literal: {literal}")
    require("ControlPlaneReadCollectionViewModel" in ts_content, "TypeScript contract missing view model")


def assert_ui_pages(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    pages = {
        str(page.get("id") or ""): page
        for page in fixture.get("ui_pages") or []
        if isinstance(page, dict)
    }
    require(set(pages) == EXPECTED_PAGE_IDS, f"unexpected UI page ids: {sorted(pages)}")

    required_states = set(fixture.get("required_view_states") or [])
    missing_states = sorted(EXPECTED_REQUIRED_VIEW_STATES - required_states)
    require(not missing_states, f"missing required view states: {missing_states}")

    route_to_page: dict[str, str] = {}
    for page_id, page in pages.items():
        require(page.get("status") == "boundary_defined_not_implemented", f"{page_id} must not claim implementation")
        require(page.get("read_only") is True, f"{page_id} must stay read-only")
        require(str(page.get("surface") or "") in {"admin_control_plane", "user_workspace"}, f"{page_id} surface drifted")
        require(required_states.issubset(set(page.get("allowed_view_states") or [])), f"{page_id} missing view states")

        route_ids = [str(route_id) for route_id in page.get("consumes_route_ids") or []]
        require(route_ids, f"{page_id} must consume at least one route")
        for route_id in route_ids:
            require(route_id in routes, f"{page_id} consumes unknown route {route_id}")
            require(route_id not in route_to_page, f"{route_id} is assigned to multiple pages")
            route_to_page[route_id] = page_id

            expected_surface = (
                "admin_control_plane"
                if routes[route_id].get("visibility") == "admin_control_plane"
                else "user_workspace"
            )
            require(page.get("surface") == expected_surface, f"{route_id} assigned to wrong UI surface")

        allowed_commands = {str(command) for command in page.get("allowed_commands") or []}
        blocked_commands = {str(command) for command in page.get("blocked_commands") or []}
        require(allowed_commands, f"{page_id} must define read-only allowed commands")
        require(blocked_commands, f"{page_id} must define blocked commands")
        for command in allowed_commands:
            for forbidden_token in FORBIDDEN_ALLOWED_COMMAND_TOKENS:
                require(
                    forbidden_token not in command,
                    f"{page_id} allowed command must stay read-only: {command}",
                )

    require(set(route_to_page) == EXPECTED_ROUTE_IDS, "UI pages must cover every read route exactly once")


def assert_read_only_policy(fixture: dict[str, Any]) -> None:
    missing_invariants = sorted(EXPECTED_READ_ONLY_INVARIANTS - set(fixture.get("read_only_invariants") or []))
    require(not missing_invariants, f"missing read-only invariants: {missing_invariants}")

    forbidden_keys = response_forbidden_output_keys()
    declared_must_not_render = set(fixture.get("must_not_render_fields") or [])
    missing_keys = sorted(forbidden_keys - declared_must_not_render)
    require(not missing_keys, f"UI boundary missing forbidden output keys: {missing_keys}")

    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-formal-ui-boundary-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read formal UI boundary check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_ui_contract_source(fixture)
    assert_ui_pages(fixture)
    assert_read_only_policy(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read formal UI boundary v1 checks passed.")


if __name__ == "__main__":
    main()
