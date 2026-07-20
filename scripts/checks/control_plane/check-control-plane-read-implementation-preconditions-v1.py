#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json"
READ_MODEL_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-model-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "go_handler_ready",
    "typescript_consumer_ready",
    "database_schema_ready",
    "database_query_ready",
    "radish_oidc_client_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "route_smoke_ready",
    "workflow_executor_ready",
    "production_ready",
}
EXPECTED_FORBIDDEN_DEPENDENCIES = {
    "database_query",
    "durable_store",
    "workflow_executor",
    "tool_executor",
    "confirmation_flow",
    "business_writeback",
    "replay_executor",
}
EXPECTED_ALLOWED_DEPENDENCIES = {
    "route_contract",
    "auth_context",
    "fixture_backed_fake_store",
    "response_fixture_builder",
    "audit_context",
}
EXPECTED_AUTH_CONTEXT_FIELDS = {
    "identity_context",
    "tenant_binding",
    "subject_binding",
    "scope_grants",
    "audit_context",
}
EXPECTED_NEGATIVE_SMOKE_CASE_TYPES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "scope_denied",
    "invalid_filter_or_route_specific_failure",
    "forbidden_method",
    "forbidden_query_parameters",
    "no_side_effects",
}
EXPECTED_FORBIDDEN_QUERY_PARAMETERS = {
    "execute",
    "replay",
    "confirmation_decision_ref",
    "writeback_payload",
    "raw_tool_payload",
    "include_secret",
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
EXPECTED_FORBIDDEN_SCOPE = {
    "Go route handler",
    "HTTP route registration",
    "TypeScript consumer contract",
    "database schema",
    "database migration",
    "database query implementation",
    "durable store",
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
REQUIRED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-implementation-preconditions-v1",
        "fake store",
        "不实现完整 read-side API",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-implementation-preconditions-v1",
        "handler ownership",
        "fake store",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-implementation-preconditions-v1",
        "fake store",
        "auth middleware",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-implementation-preconditions-v1",
        "fake store",
        "handler ownership",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-implementation-preconditions-v1",
        "Control Plane Read Implementation Preconditions",
    ],
    "docs/task-cards/control-plane-read-implementation-preconditions-v1-plan.md": [
        "negative route smoke readiness",
        "handler ownership",
        "fake store strategy",
    ],
    "scripts/README.md": [
        "check-control-plane-read-implementation-preconditions-v1.py",
        "control-plane-read-implementation-preconditions-v1.json",
        "fake store",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-implementation-preconditions-v1",
        "control-plane-read-implementation-preconditions-v1.json",
        "check-control-plane-read-implementation-preconditions-v1.py",
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


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    expected = {
        "control-plane-read-model-v1",
        "control-plane-read-route-contract-v1",
        "control-plane-read-response-fixtures-v1",
        "control-plane-read-negative-contract-v1",
    }
    require(expected.issubset(declared), "fixture must declare every read-side dependency")

    for expected_id, path in {
        "control-plane-read-model-v1": READ_MODEL_FIXTURE_PATH,
        "control-plane-read-route-contract-v1": ROUTE_CONTRACT_FIXTURE_PATH,
        "control-plane-read-response-fixtures-v1": RESPONSE_FIXTURE_PATH,
        "control-plane-read-negative-contract-v1": NEGATIVE_CONTRACT_FIXTURE_PATH,
    }.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == expected_id, f"dependency {expected_id} id drifted")
        require(
            slice_info.get("status") == "governance_boundary_satisfied",
            f"dependency {expected_id} must remain governance-boundary satisfied",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_implementation_preconditions_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-implementation-preconditions-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "implementation_preconditions_defined",
        "implementation preconditions must not claim route implementation",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def route_contracts_by_id() -> dict[str, dict[str, Any]]:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    routes = {
        str(route.get("id") or ""): route
        for route in route_contract.get("route_contracts") or []
        if isinstance(route, dict)
    }
    require(routes, "route contract fixture must contain routes")
    return routes


def response_route_ids_and_envelope() -> tuple[set[str], set[str], set[str]]:
    response_fixture = load_json(RESPONSE_FIXTURE_PATH)
    route_ids = {
        str(example.get("route_id") or "")
        for example in response_fixture.get("response_examples") or []
        if isinstance(example, dict)
    }
    envelope = set(response_fixture.get("response_envelope_fields") or [])
    forbidden_keys = set(response_fixture.get("forbidden_output_keys") or [])
    require(route_ids and envelope and forbidden_keys, "response fixture must expose route ids, envelope and forbidden keys")
    return route_ids, envelope, forbidden_keys


def assert_handler_ownership(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    handlers = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("handler_ownership") or []
        if isinstance(item, dict)
    }
    require(set(handlers) == set(routes), "handler ownership must cover every read-only route exactly once")

    for route_id, handler in handlers.items():
        route = routes[route_id]
        require(handler.get("method") == route.get("method"), f"{route_id} handler method must match route contract")
        require(handler.get("path") == route.get("path"), f"{route_id} handler path must match route contract")
        require(handler.get("owner") == "go_platform_control_plane_read_side", f"{route_id} handler owner drifted")
        require(
            handler.get("handler_status") == "future_go_handler_not_implemented",
            f"{route_id} handler must remain not implemented",
        )
        missing_allowed = sorted(EXPECTED_ALLOWED_DEPENDENCIES - set(handler.get("allowed_dependencies") or []))
        require(not missing_allowed, f"{route_id} missing allowed dependencies: {missing_allowed}")
        missing_forbidden = sorted(EXPECTED_FORBIDDEN_DEPENDENCIES - set(handler.get("forbidden_dependencies") or []))
        require(not missing_forbidden, f"{route_id} missing forbidden dependencies: {missing_forbidden}")


def assert_fake_store_strategy(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    fake_store = fixture.get("fake_store_strategy") or {}
    require(fake_store.get("strategy") == "fixture_backed_fake_store_only", "fake store strategy drifted")
    require(
        fake_store.get("status") == "precondition_defined_not_implemented",
        "fake store must remain precondition-only",
    )
    for relative_path in fake_store.get("source_fixtures") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"fake store source fixture missing: {relative_path}")

    bindings = {
        str(item.get("route_id") or ""): item
        for item in fake_store.get("route_bindings") or []
        if isinstance(item, dict)
    }
    require(set(bindings) == set(routes), "fake store bindings must cover every route")
    for route_id, binding in bindings.items():
        require(binding.get("read_model") == routes[route_id].get("read_model"), f"{route_id} read_model binding drifted")
        require(
            binding.get("store_fixture_mode") in {"single_resource_summary", "cursor_list_summary"},
            f"{route_id} store fixture mode is invalid",
        )

    forbidden_store_modes = {
        "database_query",
        "database_migration",
        "durable_store",
        "production_data_source",
        "secret_value_storage",
        "business_writeback",
    }
    missing_forbidden = sorted(forbidden_store_modes - set(fake_store.get("does_not_allow") or []))
    require(not missing_forbidden, f"fake store missing forbidden modes: {missing_forbidden}")


def assert_auth_dependency(fixture: dict[str, Any]) -> None:
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    shared_contract = route_contract.get("shared_contract") or {}
    auth = fixture.get("auth_middleware_dependency") or {}
    require(auth.get("status") == "dependency_defined_not_implemented", "auth dependency must remain not implemented")
    require("Radish OIDC" in str(auth.get("future_truth_source") or ""), "auth truth source must remain Radish OIDC")
    missing_context = sorted(EXPECTED_AUTH_CONTEXT_FIELDS - set(auth.get("required_context_fields") or []))
    require(not missing_context, f"missing auth context fields: {missing_context}")
    require(
        set(shared_contract.get("fail_closed_when_missing") or []).issubset(set(auth.get("fail_closed_when_missing") or [])),
        "auth dependency must fail closed for route contract missing fields",
    )
    require(
        auth.get("temporary_route_smoke_input") == "explicit_fake_auth_context_only",
        "temporary route smoke input must stay explicit fake auth only",
    )
    require(
        set(auth.get("forbidden_fallbacks") or []) == set(shared_contract.get("forbidden_fallbacks") or []),
        "auth forbidden fallbacks must match route contract",
    )


def assert_response_conformance(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    response_route_ids, envelope, forbidden_keys = response_route_ids_and_envelope()
    conformance = fixture.get("response_fixture_conformance") or {}
    require(
        conformance.get("status") == "precondition_defined_not_implemented",
        "response conformance must remain precondition-only",
    )
    require(set(conformance.get("route_ids") or []) == set(routes), "response conformance route ids must cover routes")
    require(response_route_ids == set(routes), "response fixture examples must cover routes")
    require(set(conformance.get("envelope_fields") or []) == envelope, "response conformance envelope drifted")
    require(
        conformance.get("failure_code_source") == "control-plane-read-route-contract-v1.failure_codes",
        "failure code source must remain route contract",
    )
    require(set(conformance.get("forbidden_output_keys") or []) == forbidden_keys, "forbidden output keys drifted")


def assert_negative_route_smoke_readiness(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    route_contract = load_json(ROUTE_CONTRACT_FIXTURE_PATH)
    negative_contract = load_json(NEGATIVE_CONTRACT_FIXTURE_PATH)
    shared_contract = route_contract.get("shared_contract") or {}
    negative = fixture.get("negative_route_smoke_readiness") or {}
    require(
        negative.get("status") == "readiness_defined_not_implemented",
        "negative route smoke readiness must not claim smoke implementation",
    )
    require(set(negative.get("route_ids") or []) == set(routes), "negative smoke readiness must cover routes")
    missing_cases = sorted(EXPECTED_NEGATIVE_SMOKE_CASE_TYPES - set(negative.get("required_case_types") or []))
    require(not missing_cases, f"missing negative smoke case types: {missing_cases}")
    require(set(negative.get("forbidden_methods") or []) == set(shared_contract.get("forbidden_methods") or []), "forbidden methods drifted")
    missing_query = sorted(EXPECTED_FORBIDDEN_QUERY_PARAMETERS - set(negative.get("forbidden_query_parameters") or []))
    require(not missing_query, f"missing forbidden query parameters: {missing_query}")
    missing_invariants = sorted(EXPECTED_DENIAL_INVARIANTS - set(negative.get("required_denial_invariants") or []))
    require(not missing_invariants, f"missing denial invariants: {missing_invariants}")
    require(
        set(negative_contract.get("required_denial_invariants") or []).issubset(
            set(negative.get("required_denial_invariants") or [])
        ),
        "negative smoke readiness must include read negative contract denial invariants",
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
        '"check-control-plane-read-implementation-preconditions-v1.py"'
        in check_repo,
        "check-repo.py must catalog control plane read implementation preconditions check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_handler_ownership(fixture)
    assert_fake_store_strategy(fixture)
    assert_auth_dependency(fixture)
    assert_response_conformance(fixture)
    assert_negative_route_smoke_readiness(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read implementation preconditions v1 checks passed.")


if __name__ == "__main__":
    main()
