#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
READ_MODEL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-model-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "go_handler_ready",
    "typescript_consumer_ready",
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
EXPECTED_DENIAL_INVARIANTS = {
    "fail_closed",
    "no_items_returned",
    "audit_ref_required",
    "no_executor_invocation",
    "no_database_write",
    "no_business_writeback",
    "no_confirmation_decision",
    "no_replay",
}
EXPECTED_FORBIDDEN_QUERY_PARAMETERS = {
    "execute",
    "replay",
    "confirmation_decision_ref",
    "writeback_payload",
    "raw_tool_payload",
    "include_secret",
}
EXPECTED_FORBIDDEN_EXECUTION_CAPABILITIES = {
    "workflow_execute",
    "tool_execute",
    "confirmation_decision",
    "business_writeback",
    "replay",
}
REQUIRED_SHARED_CASE_TYPES = {
    "forbidden_method",
    "forbidden_query_parameters",
    "forbidden_sensitive_projection",
    "forbidden_fallbacks",
    "forbidden_execution_capability",
}
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-negative-contract-v1",
        "negative contract",
        "不创建数据库 schema",
    ],
    "docs/radishmind-product-scope.md": [
        "control-plane-read-negative-contract-v1",
        "负向契约",
        "fail-closed",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-negative-contract-v1",
        "negative contract",
        "fail-closed",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-negative-contract-v1",
        "control-plane-read-negative-contract-v1.json",
        "check-control-plane-read-negative-contract-v1.py",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-negative-contract-v1",
        "control-plane-read-negative-contract-v1.json",
        "check-control-plane-read-negative-contract-v1.py",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-negative-contract-v1",
        "Control Plane Read Negative Contract",
    ],
    "docs/task-cards/control-plane-read-negative-contract-v1-plan.md": [
        "control-plane-read-negative-contract-v1",
        "negative contract",
        "fail-closed",
    ],
    "scripts/README.md": [
        "check-control-plane-read-negative-contract-v1.py",
        "control-plane-read-negative-contract-v1.json",
        "negative contract",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-negative-contract-v1",
        "control-plane-read-negative-contract-v1.json",
        "check-control-plane-read-negative-contract-v1.py",
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


def iter_keys(value: Any) -> set[str]:
    if isinstance(value, dict):
        keys = set(value)
        for child in value.values():
            keys.update(iter_keys(child))
        return keys
    if isinstance(value, list):
        keys: set[str] = set()
        for item in value:
            keys.update(iter_keys(item))
        return keys
    return set()


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-model-v1",
        "control-plane-read-route-contract-v1",
        "control-plane-read-response-fixtures-v1",
    }
    require(expected.issubset(declared), "fixture must declare read-side dependencies")
    for expected_id, path in {
        "control-plane-read-model-v1": READ_MODEL_FIXTURE_PATH,
        "control-plane-read-route-contract-v1": ROUTE_CONTRACT_FIXTURE_PATH,
        "control-plane-read-response-fixtures-v1": RESPONSE_FIXTURE_PATH,
    }.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(
            slice_info.get("id") == expected_id
            and slice_info.get("status") == "governance_boundary_satisfied",
            f"dependency {expected_id} must remain satisfied",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_negative_contract_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-negative-contract-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "governance_boundary_satisfied",
        "control plane read negative contract must only satisfy a governance boundary",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_route_negative_cases(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    route_by_id = {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    route_cases = {
        str(case.get("route_id") or ""): case
        for case in fixture.get("route_negative_cases") or []
        if isinstance(case, dict)
    }
    require(set(route_cases) == set(route_by_id), "route negative cases must cover every route contract exactly once")

    envelope_fields = set(fixture.get("response_envelope_fields") or [])
    response_envelope_fields = set(response_fixture.get("response_envelope_fields") or [])
    require(envelope_fields == response_envelope_fields, "negative response envelope must match response fixtures")
    require({"request_id", "tenant_ref", "items", "next_cursor", "failure_code", "audit_ref"}.issubset(envelope_fields), "negative envelope fields are incomplete")

    forbidden_output_keys = set(response_fixture.get("forbidden_output_keys") or [])
    required_denial_invariants = set(fixture.get("required_denial_invariants") or [])
    require(EXPECTED_DENIAL_INVARIANTS.issubset(required_denial_invariants), "required denial invariants are incomplete")

    for route_id, case in route_cases.items():
        route = route_by_id[route_id]
        request = case.get("denied_request") or {}
        response = case.get("expected_response") or {}
        require(request.get("method") == route.get("method"), f"{route_id} negative request method must match route contract")
        require(request.get("path") == route.get("path"), f"{route_id} negative request path must match route contract")
        missing_envelope = sorted(envelope_fields - set(response))
        require(not missing_envelope, f"{route_id} negative response missing envelope fields: {missing_envelope}")
        require(response.get("items") == [], f"{route_id} negative response must not return items")
        require(response.get("next_cursor") is None, f"{route_id} negative response must not return a cursor")
        require(isinstance(response.get("audit_ref"), str) and response.get("audit_ref"), f"{route_id} negative response must include audit_ref")
        failure_code = response.get("failure_code")
        require(isinstance(failure_code, str) and failure_code, f"{route_id} negative response must include failure_code")
        require(failure_code in set(route.get("failure_codes") or []), f"{route_id} failure_code must be declared by route contract")
        leaked_keys = sorted(forbidden_output_keys & iter_keys(response))
        require(not leaked_keys, f"{route_id} negative response leaks forbidden keys: {leaked_keys}")

        invariants = set(case.get("expected_invariants") or [])
        missing_invariants = sorted(EXPECTED_DENIAL_INVARIANTS - invariants)
        require(not missing_invariants, f"{route_id} negative case missing invariants: {missing_invariants}")


def assert_shared_negative_cases(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    route_ids = {
        str(route.get("id") or "")
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    shared_contract = route_contract.get("shared_contract") or {}
    forbidden_methods = set(shared_contract.get("forbidden_methods") or [])
    forbidden_fallbacks = set(shared_contract.get("forbidden_fallbacks") or [])
    forbidden_output_keys = set(response_fixture.get("forbidden_output_keys") or [])

    shared_cases = [
        case
        for case in fixture.get("shared_negative_cases") or []
        if isinstance(case, dict)
    ]
    case_by_type = {str(case.get("case_type") or ""): case for case in shared_cases}
    missing_case_types = sorted(REQUIRED_SHARED_CASE_TYPES - set(case_by_type))
    require(not missing_case_types, f"missing shared negative case types: {missing_case_types}")

    for case in shared_cases:
        referenced_route_ids = set(case.get("route_ids") or [])
        unknown_route_ids = sorted(referenced_route_ids - route_ids)
        require(not unknown_route_ids, f"{case.get('id')} references unknown route ids: {unknown_route_ids}")
        require(case.get("expected_side_effects") == "none", f"{case.get('id')} must not allow side effects")
        invariants = set(case.get("expected_invariants") or [])
        missing_invariants = sorted(
            {"fail_closed", "no_executor_invocation", "no_database_write", "no_business_writeback", "no_confirmation_decision", "no_replay"}
            - invariants
        )
        require(not missing_invariants, f"{case.get('id')} missing shared invariants: {missing_invariants}")

    method_case = case_by_type["forbidden_method"]
    require(set(method_case.get("methods") or []) == forbidden_methods, "forbidden method case must cover route contract methods")
    require(set(method_case.get("route_ids") or []) == route_ids, "forbidden method case must cover every route")

    query_case = case_by_type["forbidden_query_parameters"]
    missing_query_parameters = sorted(EXPECTED_FORBIDDEN_QUERY_PARAMETERS - set(query_case.get("query_parameters") or []))
    require(not missing_query_parameters, f"missing forbidden query parameters: {missing_query_parameters}")

    projection_case = case_by_type["forbidden_sensitive_projection"]
    missing_output_keys = sorted(forbidden_output_keys - set(projection_case.get("requested_output_keys") or []))
    require(not missing_output_keys, f"missing forbidden output projection keys: {missing_output_keys}")

    fallback_case = case_by_type["forbidden_fallbacks"]
    require(set(fallback_case.get("requested_fallbacks") or []) == forbidden_fallbacks, "forbidden fallback case must match route contract")
    require(set(fallback_case.get("route_ids") or []) == route_ids, "forbidden fallback case must cover every route")

    execution_case = case_by_type["forbidden_execution_capability"]
    missing_capabilities = sorted(EXPECTED_FORBIDDEN_EXECUTION_CAPABILITIES - set(execution_case.get("requested_capabilities") or []))
    require(not missing_capabilities, f"missing forbidden execution capabilities: {missing_capabilities}")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-negative-contract-v1.py", [])' in check_repo,
        "check-repo.py must run control plane read negative contract check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_route_negative_cases(fixture)
    assert_shared_negative_cases(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read negative contract v1 checks passed.")


if __name__ == "__main__":
    main()
