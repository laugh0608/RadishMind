#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-implementation-readiness-v1.json"
)
PRODUCT_SURFACE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/product-surface-v1-boundary.json"
AUTH_DB_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
)
CONSUMER_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json"
FORMAL_UI_BOUNDARY_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-formal-ui-boundary-v1.json"
)
TS_CONTRACT_PATH = REPO_ROOT / "contracts/typescript/control-plane-read-api.ts"
CONSOLE_PACKAGE_PATH = REPO_ROOT / "apps/radishmind-console/package.json"
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
EXPECTED_SEQUENCE = [
    "shared-read-shell",
    "admin-tenant-overview",
    "workspace-applications",
    "workspace-api-keys",
    "workspace-usage-quota",
    "workspace-workflow-definitions",
    "workspace-run-history",
    "admin-audit-log",
]
EXPECTED_CONTRACT_IMPORTS = {
    "CONTROL_PLANE_READ_ROUTES",
    "CONTROL_PLANE_READ_ROUTE_DEFINITIONS",
    "CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS",
    "ControlPlaneReadCollectionViewModel",
    "ControlPlaneReadRouteCatalogViewModel",
    "ControlPlaneReadRouteId",
    "ControlPlaneReadResponseByRoute",
    "controlPlaneReadResponseHasForbiddenOutput",
    "isControlPlaneReadEnvelope",
    "listControlPlaneReadRouteCatalog",
    "toControlPlaneReadCollectionViewModel",
}
EXPECTED_REUSE_RULES = {
    "pages_must_not_duplicate_route_paths",
    "pages_must_not_parse_raw_fixture_json_directly",
    "pages_must_map_envelopes_through_toControlPlaneReadCollectionViewModel",
    "pages_must_check_controlPlaneReadResponseHasForbiddenOutput_before_render",
    "pages_must_preserve_failureCode_and_auditRef",
    "denied_pages_must_render_no_items",
    "stale_pages_must_keep_previous_read_only_data_only",
}
EXPECTED_TEST_IDS = {
    "readiness_checker",
    "consumer_smoke",
    "repo_fast",
    "future_product_app_build",
    "future_component_state_tests",
    "future_visual_smoke",
}
EXPECTED_NOW_TEST_IDS = {
    "readiness_checker",
    "consumer_smoke",
    "repo_fast",
}
EXPECTED_STOP_LINES = {
    "do not create React pages in this slice",
    "do not modify apps/radishmind-console into formal user workspace",
    "do not modify apps/radishmind-console into production admin console",
    "do not request live backend from readiness checker",
    "do not add database schema migration query or repository",
    "do not add Radish OIDC middleware token validation login or logout",
    "do not add API key lifecycle quota enforcement rate limiting billing or cost ledger",
    "do not add workflow builder executor confirmation writeback replay or materialized result reader",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "react_ui_implemented",
    "formal_user_workspace_ui_ready",
    "production_admin_console_ready",
    "apps_radishmind_console_upgraded",
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
FORBIDDEN_CONSOLE_SCRIPT_NAMES = {
    "deploy",
    "publish",
    "release",
    "package",
    "docker",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "正式 UI 实现 readiness",
    ],
    "docs/contracts/README.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "正式 UI 实现 readiness",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "apps/radishmind-web/",
        "apps/radishmind-console/",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "Control Plane Read Formal UI Implementation Readiness",
    ],
    "docs/task-cards/control-plane-read-formal-ui-implementation-readiness-v1-plan.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "apps/radishmind-web/",
        "contracts/typescript/control-plane-read-api.ts",
    ],
    "scripts/README.md": [
        "check-control-plane-read-formal-ui-implementation-readiness-v1.py",
        "control-plane-read-formal-ui-implementation-readiness-v1.json",
    ],
    "services/platform/README.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "apps/radishmind-web/",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-formal-ui-implementation-readiness-v1",
        "control-plane-read-formal-ui-implementation-readiness-v1.json",
        "check-control-plane-read-formal-ui-implementation-readiness-v1.py",
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


def formal_ui_pages_by_id() -> dict[str, dict[str, Any]]:
    boundary = load_json(FORMAL_UI_BOUNDARY_FIXTURE_PATH)
    pages = {
        str(page.get("id") or ""): page
        for page in boundary.get("ui_pages") or []
        if isinstance(page, dict)
    }
    require(set(pages) == EXPECTED_PAGE_IDS, "formal UI boundary page ids drifted")
    return pages


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "product-surface-v1-boundary",
        "control-plane-read-auth-db-preconditions-v1",
        "control-plane-read-consumer-contract-v1",
        "control-plane-read-formal-ui-boundary-v1",
    }
    require(expected.issubset(declared), "formal UI implementation readiness fixture must declare dependencies")

    expected_statuses = {
        "product-surface-v1-boundary": (PRODUCT_SURFACE_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-auth-db-preconditions-v1": (
            AUTH_DB_PRECONDITIONS_FIXTURE_PATH,
            "auth_db_preconditions_defined",
        ),
        "control-plane-read-consumer-contract-v1": (CONSUMER_CONTRACT_FIXTURE_PATH, "consumer_contract_defined"),
        "control-plane-read-formal-ui-boundary-v1": (
            FORMAL_UI_BOUNDARY_FIXTURE_PATH,
            "formal_ui_boundary_defined",
        ),
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
    require(
        fixture.get("kind") == "control_plane_read_formal_ui_implementation_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-formal-ui-implementation-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "formal_ui_implementation_readiness_defined",
        "readiness slice must not claim implementation",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_engineering_landing(fixture: dict[str, Any]) -> None:
    landing = fixture.get("engineering_landing") or {}
    require(
        landing.get("status") == "readiness_only_no_react_files_created",
        "engineering landing must stay readiness-only",
    )
    require(landing.get("future_app_root") == "apps/radishmind-web/", "future product app root drifted")
    require(
        landing.get("future_feature_root") == "apps/radishmind-web/src/features/control-plane-read/",
        "future feature root drifted",
    )
    require(landing.get("shared_contract") == "contracts/typescript/control-plane-read-api.ts", "contract source drifted")
    require(landing.get("consumer_smoke") == "scripts/run-control-plane-read-consumer-smoke.py --check", "smoke drifted")
    require(landing.get("local_ops_console_root") == "apps/radishmind-console/", "ops console root drifted")
    require(
        landing.get("local_ops_console_boundary") == "remain_local_ops_surface_only",
        "ops console boundary drifted",
    )
    require(
        str(landing.get("future_app_root")) != str(landing.get("local_ops_console_root")),
        "formal product app root must not be the local ops console root",
    )


def assert_app_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("app_boundary") or {}
    formal_app = boundary.get("formal_product_app") or {}
    local_console = boundary.get("local_ops_console") or {}

    require(formal_app.get("root") == "apps/radishmind-web/", "formal product app root drifted")
    require(
        formal_app.get("status") == "reserved_not_created_in_this_slice",
        "formal product app must remain reserved in this slice",
    )
    require(
        set(formal_app.get("surfaces") or []) == {"admin_control_plane", "user_workspace"},
        "formal app surfaces drifted",
    )
    require(local_console.get("root") == "apps/radishmind-console/", "local console root drifted")
    require(local_console.get("status") == "local_ops_surface_only", "local console status drifted")
    require(
        {"formal user workspace", "production admin console", "control plane write surface"}.issubset(
            set(local_console.get("must_not_become") or [])
        ),
        "local console must keep formal UI and production admin stop-lines",
    )

    package_json = load_json(CONSOLE_PACKAGE_PATH)
    require(package_json.get("private") is True, "apps/radishmind-console package must remain private")
    scripts = package_json.get("scripts") or {}
    require(isinstance(scripts, dict), "apps/radishmind-console package scripts must be an object")
    forbidden_scripts = sorted(FORBIDDEN_CONSOLE_SCRIPT_NAMES & set(scripts))
    require(not forbidden_scripts, f"apps/radishmind-console must not add production scripts: {forbidden_scripts}")


def assert_implementation_sequence(fixture: dict[str, Any]) -> None:
    pages = formal_ui_pages_by_id()
    sequence = [item for item in fixture.get("implementation_sequence") or [] if isinstance(item, dict)]
    require(len(sequence) == len(EXPECTED_SEQUENCE), "implementation sequence length drifted")

    orders = [item.get("order") for item in sequence]
    require(orders == list(range(1, len(sequence) + 1)), "implementation sequence order must be contiguous")

    observed_sequence: list[str] = []
    route_coverage: set[str] = set()
    page_coverage: set[str] = set()
    for item in sequence:
        require(item.get("status") == "planned_not_implemented", "sequence steps must remain planned only")
        page_id = item.get("page_id")
        if page_id is None:
            require(item.get("step_id") == "shared-read-shell", "first non-page step id drifted")
            require(item.get("consumes_route_ids") == [], "shared shell must not consume a route directly")
            observed_sequence.append(str(item.get("step_id") or ""))
            continue

        page_id = str(page_id)
        require(page_id in pages, f"sequence references unknown page: {page_id}")
        boundary_page = pages[page_id]
        expected_routes = [str(route_id) for route_id in boundary_page.get("consumes_route_ids") or []]
        route_ids = [str(route_id) for route_id in item.get("consumes_route_ids") or []]
        require(route_ids == expected_routes, f"{page_id} route assignment drifted")
        require(item.get("surface") == boundary_page.get("surface"), f"{page_id} surface drifted")
        require(set(route_ids).issubset(EXPECTED_ROUTE_IDS), f"{page_id} references unknown route")
        route_coverage.update(route_ids)
        page_coverage.add(page_id)
        observed_sequence.append(page_id)

    require(observed_sequence == EXPECTED_SEQUENCE, "implementation sequence order drifted")
    require(page_coverage == EXPECTED_PAGE_IDS, "implementation sequence must cover every formal UI page")
    require(route_coverage == EXPECTED_ROUTE_IDS, "implementation sequence must cover every read route")


def assert_contract_reuse(fixture: dict[str, Any]) -> None:
    reuse = fixture.get("consumer_contract_reuse") or {}
    require(reuse.get("status") == "required_before_react_pages", "contract reuse status drifted")
    missing_imports = sorted(EXPECTED_CONTRACT_IMPORTS - set(reuse.get("imports") or []))
    require(not missing_imports, f"missing required TypeScript imports: {missing_imports}")
    missing_rules = sorted(EXPECTED_REUSE_RULES - set(reuse.get("required_reuse_rules") or []))
    require(not missing_rules, f"missing consumer reuse rules: {missing_rules}")

    ts_content = TS_CONTRACT_PATH.read_text(encoding="utf-8")
    for literal in EXPECTED_CONTRACT_IMPORTS:
        require(literal in ts_content, f"TypeScript control-plane read contract missing literal: {literal}")
    for literal in (
        "databaseBacked: false",
        "formalUiReady: false",
        "canMutate: false",
        "canExecuteWorkflow: false",
        "canWriteBusinessTruth: false",
        "canRevealSecrets: false",
    ):
        require(literal in ts_content, f"TypeScript stop-line literal missing: {literal}")


def assert_testing_strategy(fixture: dict[str, Any]) -> None:
    tests = {
        str(item.get("id") or ""): item
        for item in fixture.get("testing_strategy") or []
        if isinstance(item, dict)
    }
    require(set(tests) == EXPECTED_TEST_IDS, f"testing strategy ids drifted: {sorted(tests)}")
    for test_id in EXPECTED_NOW_TEST_IDS:
        require(tests[test_id].get("status") == "required_now", f"{test_id} must be required now")
    require(
        tests["readiness_checker"].get("command")
        == "./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-formal-ui-implementation-readiness-v1.py",
        "readiness checker command drifted",
    )
    require(
        tests["consumer_smoke"].get("command") == "./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check",
        "consumer smoke command drifted",
    )
    require(
        tests["repo_fast"].get("command")
        == "macOS/Linux/WSL: ./scripts/check-repo-fast.sh; Windows/PowerShell: pwsh ./scripts/check-repo.ps1 -Fast",
        "fast command drifted",
    )

    future_ids = EXPECTED_TEST_IDS - EXPECTED_NOW_TEST_IDS
    for test_id in future_ids:
        require(
            tests[test_id].get("status") == "future_react_implementation_only",
            f"{test_id} must stay future-only",
        )


def assert_stop_lines_and_docs(fixture: dict[str, Any]) -> None:
    missing_stop_lines = sorted(EXPECTED_STOP_LINES - set(fixture.get("implementation_stop_lines") or []))
    require(not missing_stop_lines, f"missing implementation stop-lines: {missing_stop_lines}")
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-formal-ui-implementation-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read formal UI implementation readiness check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_engineering_landing(fixture)
    assert_app_boundary(fixture)
    assert_implementation_sequence(fixture)
    assert_contract_reuse(fixture)
    assert_testing_strategy(fixture)
    assert_stop_lines_and_docs(fixture)
    print("control plane read formal UI implementation readiness v1 checks passed.")


if __name__ == "__main__":
    main()
