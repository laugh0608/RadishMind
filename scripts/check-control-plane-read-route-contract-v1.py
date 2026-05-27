#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
READ_MODEL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-model-v1.json"
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
EXPECTED_READ_MODELS = {
    "tenant-summary",
    "application-summary",
    "api-key-summary",
    "quota-summary",
    "workflow-definition-summary",
    "run-record-summary",
    "audit-summary",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "go_handler_ready",
    "typescript_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "radish_oidc_client_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "Go route handler",
    "TypeScript consumer contract",
    "database schema",
    "database migration",
    "database query implementation",
    "Radish OIDC integration",
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
    "production admin console",
    "user workspace UI",
}
EXPECTED_FORBIDDEN_OUTPUTS = {
    "raw_secret_value",
    "api_key_value",
    "api_key_hash",
    "authorization_header",
    "bearer_token",
    "cookie_value",
    "raw_request_body_dump",
    "raw_tool_payload",
    "business_writeback_payload",
    "full_prompt_dump_with_secret",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-route-contract-v1",
        "read-only route",
        "不创建数据库 schema",
    ],
    "docs/radishmind-current-focus.md": [
        "control-plane-read-route-contract-v1",
        "control-plane-read-route-contract-v1.json",
        "check-control-plane-read-route-contract-v1.py",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-route-contract-v1",
        "GET /v1/user-workspace/runs",
        "tenant-scoped",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-route-contract-v1",
        "read-only route",
        "tenant-scoped",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-route-contract-v1",
        "control-plane-read-route-contract-v1.json",
        "check-control-plane-read-route-contract-v1.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-route-contract-v1",
        "control-plane-read-route-contract-v1.json",
        "check-control-plane-read-route-contract-v1.py",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-route-contract-v1",
        "Control Plane Read Route Contract",
    ],
    "docs/task-cards/control-plane-read-route-contract-v1-plan.md": [
        "control-plane-read-route-contract-v1",
        "GET /v1/user-workspace/runs",
        "route contract",
    ],
    "scripts/README.md": [
        "check-control-plane-read-route-contract-v1.py",
        "control-plane-read-route-contract-v1.json",
        "read-only route",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-route-contract-v1",
        "control-plane-read-route-contract-v1.json",
        "check-control-plane-read-route-contract-v1.py",
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


def assert_dependency(fixture: dict[str, Any]) -> None:
    require("control-plane-read-model-v1" in set(fixture.get("depends_on") or []), "fixture must depend on read model v1")
    read_model_fixture = load_json(READ_MODEL_FIXTURE_PATH)
    slice_info = read_model_fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-model-v1"
        and slice_info.get("status") == "governance_boundary_satisfied",
        "control plane read route contract depends on satisfied read model v1",
    )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_route_contract_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-route-contract-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "control plane read route contract must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_routes(fixture: dict[str, Any]) -> None:
    routes = {
        str(item.get("id") or ""): item
        for item in fixture.get("route_contracts") or []
        if isinstance(item, dict)
    }
    require(set(routes) == EXPECTED_ROUTE_IDS, f"unexpected route contracts: {sorted(routes)}")
    require(
        {str(route.get("read_model") or "") for route in routes.values()} == EXPECTED_READ_MODELS,
        "route contracts must cover every read model exactly once",
    )

    for route_id, route in routes.items():
        require(route.get("method") == "GET", f"{route_id} must be GET only")
        require(str(route.get("path") or "").startswith("/v1/"), f"{route_id}.path must be a v1 route")
        require(route.get("tenant_scoped") is True, f"{route_id} must remain tenant scoped")
        require(route.get("status") == "defined_not_implemented", f"{route_id} must remain defined_not_implemented")
        require(str(route.get("required_scope") or "").endswith(":read"), f"{route_id} must require a read scope")
        failure_codes = set(route.get("failure_codes") or [])
        require("identity_context_missing" in failure_codes, f"{route_id} must fail on missing identity")
        require("tenant_binding_missing" in failure_codes, f"{route_id} must fail on missing tenant binding")
        require("scope_denied" in failure_codes, f"{route_id} must fail on missing scope")
        if str(route.get("pagination")) == "cursor_required":
            require(isinstance(route.get("allowed_filters"), list), f"{route_id}.allowed_filters must be a list")

    require(
        routes["run-record-summary-list-route"].get("path") == "/v1/user-workspace/runs",
        "run record route path drifted",
    )
    require(
        "api_key_value" not in set(routes["api-key-summary-list-route"].get("allowed_filters") or []),
        "API key route must not filter on raw key values",
    )


def assert_shared_contract_and_sanitization(fixture: dict[str, Any]) -> None:
    shared_contract = fixture.get("shared_contract") or {}
    require(
        {"limit", "cursor", "sort"}.issubset(set(shared_contract.get("list_parameters") or [])),
        "shared contract must define list parameters",
    )
    require(
        {"request_id", "tenant_ref", "items", "failure_code", "audit_ref"}.issubset(
            set(shared_contract.get("response_envelope_fields") or [])
        ),
        "shared contract must define response envelope fields",
    )
    require(
        {"identity_context", "tenant_binding", "subject_binding", "scope_grant"}.issubset(
            set(shared_contract.get("fail_closed_when_missing") or [])
        ),
        "shared contract must fail closed on identity, tenant, subject and scope",
    )
    require(
        {"POST", "PUT", "PATCH", "DELETE"}.issubset(set(shared_contract.get("forbidden_methods") or [])),
        "shared contract must forbid write methods",
    )
    require(
        {"anonymous_read", "cross_tenant_read", "mock_provider_fallback"}.issubset(
            set(shared_contract.get("forbidden_fallbacks") or [])
        ),
        "shared contract must forbid unsafe fallbacks",
    )

    sanitization_policy = fixture.get("sanitization_policy") or {}
    require("redacted_secret_ref" in set(sanitization_policy.get("allowed_outputs") or []), "must allow redacted secret refs")
    missing_forbidden_outputs = sorted(EXPECTED_FORBIDDEN_OUTPUTS - set(sanitization_policy.get("forbidden_outputs") or []))
    require(not missing_forbidden_outputs, f"missing forbidden outputs: {missing_forbidden_outputs}")

    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("check-control-plane-read-route-contract-v1.py", [])' in check_repo,
        "check-repo.py must run control plane read route contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependency(fixture)
    assert_slice(fixture)
    assert_routes(fixture)
    assert_shared_contract_and_sanitization(fixture)
    assert_evidence_and_docs(fixture)
    print("control plane read route contract v1 checks passed.")


if __name__ == "__main__":
    main()
