#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-consumer-contract-v1.json"
FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
)
AUTH_DB_PRECONDITIONS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-db-preconditions-v1.json"
)
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
TS_CONTRACT_PATH = REPO_ROOT / "contracts/typescript/control-plane-read-api.ts"
CONSUMER_SMOKE_PATH = REPO_ROOT / "scripts/run-control-plane-read-consumer-smoke.py"
GO_ROUTE_SOURCE_PATH = REPO_ROOT / "services/platform/internal/httpapi/control_plane_read.go"
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
EXPECTED_FORBIDDEN_CLAIMS = {
    "formal_user_workspace_ui_ready",
    "production_admin_console_ready",
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
EXPECTED_TS_EXPORTS = {
    "CONTROL_PLANE_READ_ROUTES",
    "CONTROL_PLANE_READ_ROUTE_IDS",
    "CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS",
    "ControlPlaneReadEnvelope",
    "ControlPlaneReadResponseByRoute",
    "ControlPlaneReadRequestByRoute",
    "ControlPlaneReadRouteCatalogViewModel",
    "ControlPlaneReadCollectionViewModel",
    "listControlPlaneReadRouteCatalog",
    "isControlPlaneReadEnvelope",
    "toControlPlaneReadCollectionViewModel",
    "controlPlaneReadResponseHasForbiddenOutput",
}
EXPECTED_VIEW_STOP_LINES = {
    "databaseBacked: false",
    "formalUiReady: false",
    "canMutate: false",
    "canExecuteWorkflow: false",
    "canWriteBusinessTruth: false",
    "canRevealSecrets: false",
}
FORBIDDEN_TS_LITERALS = {
    "databaseBacked: true",
    "formalUiReady: true",
    "canMutate: true",
    "canExecuteWorkflow: true",
    "canWriteBusinessTruth: true",
    "canRevealSecrets: true",
    "database_backed: true",
    "formal_ui_ready: true",
    "can_execute_workflow: true",
    "can_write_business_truth: true",
}
EXPECTED_CONSUMER_INVARIANTS = {
    "all_routes_consumed",
    "all_envelopes_complete",
    "failure_views_are_denied",
    "failure_views_return_no_items",
    "audit_ref_required",
    "forbidden_output_absent",
    "read_only_views",
    "secret_projection_disabled",
    "negative_contract_invariants_preserved",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "formal user workspace UI",
    "production admin console",
    "database schema",
    "database migration",
    "database query implementation",
    "repository implementation",
    "Radish OIDC integration",
    "auth middleware implementation",
    "tenant or user CRUD",
    "API key generation",
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
REQUIRED_DOC_REFERENCES = {
    "contracts/README.md": [
        "typescript/control-plane-read-api.ts",
        "ControlPlaneReadCollectionViewModel",
        "run-control-plane-read-consumer-smoke.py --check",
    ],
    "docs/README.md": [
        "control-plane-read-consumer-contract-v1",
        "TypeScript consumer contract",
        "不实现完整 read-side API",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-consumer-contract-v1",
        "contracts/typescript/control-plane-read-api.ts",
        "run-control-plane-read-consumer-smoke.py",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-consumer-contract-v1",
        "contracts/typescript/control-plane-read-api.ts",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-consumer-contract-v1",
        "TypeScript consumer contract",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-consumer-contract-v1",
        "正式用户端",
    ],
    "docs/radishmind-integration-contracts.md": [
        "control-plane-read-consumer-contract-v1",
        "contracts/typescript/control-plane-read-api.ts",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-consumer-contract-v1",
        "Control Plane Read Consumer Contract",
    ],
    "docs/task-cards/control-plane-read-consumer-contract-v1-plan.md": [
        "control-plane-read-consumer-contract-v1",
        "TypeScript consumer contract",
        "run-control-plane-read-consumer-smoke.py",
    ],
    "scripts/README.md": [
        "check-control-plane-read-consumer-contract-v1.py",
        "control-plane-read-consumer-contract-v1.json",
        "run-control-plane-read-consumer-smoke.py",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-consumer-contract-v1",
        "control-plane-read-consumer-contract-v1.json",
        "check-control-plane-read-consumer-contract-v1.py",
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


def response_examples_by_route_id() -> dict[str, dict[str, Any]]:
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    examples = {
        str(example.get("route_id") or ""): example
        for example in response_fixture.get("response_examples") or []
        if isinstance(example, dict)
    }
    require(set(examples) == EXPECTED_ROUTE_IDS, "response fixture route ids drifted")
    return examples


def implemented_routes_by_id() -> dict[str, dict[str, Any]]:
    implementation = load_json(FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH)
    implemented = {
        str(route.get("route_id") or ""): route
        for route in implementation.get("implemented_routes") or []
        if isinstance(route, dict)
    }
    require(set(implemented) == EXPECTED_ROUTE_IDS, "fake-store implementation route ids drifted")
    return implemented


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-fake-store-handler-implementation-v1",
        "control-plane-read-auth-db-preconditions-v1",
        "control-plane-read-route-contract-v1",
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-negative-contract-v1",
    }
    require(expected.issubset(declared), "consumer contract fixture must declare every dependency")

    expected_statuses = {
        "control-plane-read-fake-store-handler-implementation-v1": (
            FAKE_STORE_IMPLEMENTATION_FIXTURE_PATH,
            "fake_store_handler_implemented",
        ),
        "control-plane-read-auth-db-preconditions-v1": (
            AUTH_DB_PRECONDITIONS_FIXTURE_PATH,
            "auth_db_preconditions_defined",
        ),
        "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-response-fixtures-v1": (RESPONSE_FIXTURE_PATH, "governance_boundary_satisfied"),
        "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
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
    require(fixture.get("kind") == "control_plane_read_consumer_contract_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-consumer-contract-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "consumer_contract_defined",
        "consumer contract must not claim formal UI or production readiness",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_route_catalog(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    implemented = implemented_routes_by_id()
    catalog = {
        str(route.get("route_id") or ""): route
        for route in fixture.get("route_catalog") or []
        if isinstance(route, dict)
    }
    require(set(catalog) == set(routes), "consumer route catalog must cover every read route")
    for route_id, item in catalog.items():
        route = routes[route_id]
        implementation = implemented[route_id]
        require(item.get("method") == route.get("method"), f"{route_id} method drifted")
        require(item.get("path") == route.get("path"), f"{route_id} path drifted")
        require(item.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(item.get("required_scope") == route.get("required_scope"), f"{route_id} scope drifted")
        require(item.get("pagination") == route.get("pagination"), f"{route_id} pagination drifted")
        require(set(item.get("allowed_filters") or []) == set(route.get("allowed_filters") or []), f"{route_id} filters drifted")
        require(
            implementation.get("status") == "fake_store_backed_handler_implemented",
            f"{route_id} must already have fake-store handler implementation",
        )


def assert_typescript_contract(fixture: dict[str, Any]) -> None:
    ts_content = TS_CONTRACT_PATH.read_text(encoding="utf-8")
    go_content = GO_ROUTE_SOURCE_PATH.read_text(encoding="utf-8")
    routes = route_contracts_by_id()
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)

    ts_contract = fixture.get("typescript_contract") or {}
    require(ts_contract.get("file") == "contracts/typescript/control-plane-read-api.ts", "typescript contract file drifted")
    require(set(ts_contract.get("route_ids") or []) == EXPECTED_ROUTE_IDS, "typescript contract route ids drifted")
    missing_exports = sorted(EXPECTED_TS_EXPORTS - set(ts_contract.get("exports") or []))
    require(not missing_exports, f"fixture missing TS exports: {missing_exports}")

    for literal in EXPECTED_TS_EXPORTS | EXPECTED_VIEW_STOP_LINES:
        require(literal in ts_content, f"TypeScript contract missing literal: {literal}")
    for literal in FORBIDDEN_TS_LITERALS:
        require(literal not in ts_content, f"TypeScript contract enables forbidden state: {literal}")

    for route_id, route in routes.items():
        for literal in (route_id, str(route.get("path")), str(route.get("read_model")), str(route.get("required_scope"))):
            require(literal in ts_content, f"TypeScript contract missing route literal {literal}")
        require(str(route.get("path")) in go_content, f"Go read route missing path: {route.get('path')}")

    for field in response_fixture.get("response_envelope_fields") or []:
        require(str(field) in ts_content, f"TypeScript contract missing envelope field: {field}")
    for key in response_fixture.get("forbidden_output_keys") or []:
        require(str(key) in ts_content, f"TypeScript contract missing forbidden output key: {key}")


def assert_consumer_smoke(fixture: dict[str, Any]) -> None:
    smoke = fixture.get("consumer_smoke") or {}
    require(smoke.get("script") == "scripts/run-control-plane-read-consumer-smoke.py", "consumer smoke script drifted")
    missing_invariants = sorted(EXPECTED_CONSUMER_INVARIANTS - set(smoke.get("required_invariants") or []))
    require(not missing_invariants, f"missing consumer smoke invariants: {missing_invariants}")
    for relative_path in smoke.get("source_fixtures") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"consumer smoke source fixture missing: {relative_path}")

    smoke_result = subprocess.run(
        [sys.executable, str(CONSUMER_SMOKE_PATH), "--check"],
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if smoke_result.returncode != 0:
        if smoke_result.stdout:
            print(smoke_result.stdout, end="")
        if smoke_result.stderr:
            print(smoke_result.stderr, end="", file=sys.stderr)
        raise SystemExit(smoke_result.returncode)


def assert_response_and_negative_contracts() -> None:
    examples = response_examples_by_route_id()
    negative_contract = load_json(NEGATIVE_CONTRACT_FIXTURE_PATH)
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    envelope_fields = set(response_fixture.get("response_envelope_fields") or [])
    forbidden_keys = set(response_fixture.get("forbidden_output_keys") or [])
    require(envelope_fields == {"request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"}, "envelope fields drifted")
    require(forbidden_keys, "forbidden output keys must be declared")

    for route_id, example in examples.items():
        for response_kind in ("success", "failure"):
            response = example.get(response_kind) or {}
            require(envelope_fields.issubset(set(response)), f"{route_id} {response_kind} response missing envelope fields")
            require(str(response.get("audit_ref") or "").strip(), f"{route_id} {response_kind} response missing audit_ref")
        failure = example.get("failure") or {}
        require(failure.get("items") == [], f"{route_id} failure response must return no items")
        require(str(failure.get("failure_code") or "").strip(), f"{route_id} failure response must include failure_code")

    expected_denial_invariants = {
        "fail_closed",
        "no_items_returned",
        "audit_ref_required",
        "no_executor_invocation",
        "no_database_write",
        "no_business_writeback",
        "no_confirmation_decision",
        "no_replay",
    }
    require(
        expected_denial_invariants.issubset(set(negative_contract.get("required_denial_invariants") or [])),
        "negative contract denial invariants drifted",
    )


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-consumer-contract-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read consumer contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_route_catalog(fixture)
    assert_typescript_contract(fixture)
    assert_consumer_smoke(fixture)
    assert_response_and_negative_contracts()
    assert_policy_and_docs(fixture)
    print("control plane read consumer contract v1 checks passed.")


if __name__ == "__main__":
    main()
