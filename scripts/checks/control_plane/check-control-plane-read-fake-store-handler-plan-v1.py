#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-plan-v1.json"
IMPLEMENTATION_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-fake-store-handler-implementation-v1.json"
PRECONDITIONS_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-preconditions-v1.json"
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RESPONSE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-response-fixtures-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "control_plane_api_ready",
    "go_handler_ready",
    "http_route_registered",
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
EXPECTED_PHASE_IDS = {
    "phase-1-shared-read-shell",
    "phase-2-single-resource-summary-routes",
    "phase-3-cursor-list-summary-routes",
    "phase-4-negative-route-smoke",
}
EXPECTED_FORBIDDEN_STORE_MODES = {
    "database_query",
    "database_migration",
    "durable_store",
    "production_data_source",
    "secret_value_storage",
    "provider_call",
    "workflow_executor",
    "business_writeback",
}
EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS = {
    "public_fake_auth_header",
    "query_scope_override",
    "anonymous_read",
    "cross_tenant_read",
    "ops_console_elevation",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "fake_store_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_ROUTE_REGISTRATION_LITERALS = {
    "tenant-summary-route": "controlPlaneTenantSummaryRoute",
    "application-summary-list-route": "controlPlaneApplicationSummaryListRoute",
    "api-key-summary-list-route": "controlPlaneAPIKeySummaryListRoute",
    "quota-summary-route": "controlPlaneQuotaSummaryRoute",
    "workflow-definition-summary-list-route": "controlPlaneWorkflowDefinitionSummaryListRoute",
    "run-record-summary-list-route": "controlPlaneRunRecordSummaryListRoute",
    "audit-summary-list-route": "controlPlaneAuditSummaryListRoute",
}
EXPECTED_FORBIDDEN_SCOPE = {
    "Go route handler implementation",
    "HTTP route registration",
    "TypeScript consumer contract",
    "database schema",
    "database migration",
    "database query implementation",
    "durable store",
    "Radish OIDC integration",
    "public fake auth header",
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
        "control-plane-read-fake-store-handler-plan-v1",
        "fake-store-backed read handler plan",
        "不实现 Go handler",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "fake-store-backed read handler plan",
        "plan-only",
    ],
    "docs/radishmind-current-focus.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "fake-store-backed read handler plan",
        "不直接实现数据库、OIDC、executor、confirmation、writeback 或 replay",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "services/platform/internal/httpapi",
        "fake store",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "fake-store-backed read handler plan",
        "不默认进入数据库",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "fake-store-backed",
        "handler plan",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "Control Plane Read Fake-Store Handler Plan",
    ],
    "docs/task-cards/control-plane-read-fake-store-handler-plan-v1-plan.md": [
        "test-only fake auth context",
        "fake-store-backed handler",
        "不创建或修改 Go route handler",
    ],
    "scripts/README.md": [
        "check-control-plane-read-fake-store-handler-plan-v1.py",
        "control-plane-read-fake-store-handler-plan-v1.json",
        "fake-store-backed read handler plan",
    ],
    "services/platform/README.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "plan-only",
        "/v1/control-plane/tenants/{tenant_ref}/summary",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-plan-v1.json",
        "check-control-plane-read-fake-store-handler-plan-v1.py",
    ],
}
IMPLEMENTED_DOC_REFERENCES = {
    "docs/README.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "不实现完整 read-side API",
    ],
    "docs/contracts/control-plane-read-side.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed",
    ],
    "docs/radishmind-current-focus.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "不直接实现数据库、OIDC、executor、confirmation、writeback 或 replay",
    ],
    "docs/radishmind-architecture.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "services/platform/internal/httpapi",
        "fake store",
    ],
    "docs/radishmind-roadmap.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "不默认进入数据库",
    ],
    "docs/radishmind-capability-matrix.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "fake-store-backed",
    ],
    "docs/task-cards/README.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "Control Plane Read Fake-Store Handler Plan",
    ],
    "docs/task-cards/control-plane-read-fake-store-handler-plan-v1-plan.md": [
        "test-only fake auth context",
        "fake-store-backed handler",
        "control-plane-read-fake-store-handler-implementation-v1",
    ],
    "scripts/README.md": [
        "check-control-plane-read-fake-store-handler-plan-v1.py",
        "control-plane-read-fake-store-handler-plan-v1.json",
        "fake-store-backed read handler plan",
    ],
    "services/platform/README.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-implementation-v1",
        "/v1/control-plane/tenants/{tenant_ref}/summary",
    ],
    "docs/devlogs/2026-W22.md": [
        "control-plane-read-fake-store-handler-plan-v1",
        "control-plane-read-fake-store-handler-plan-v1.json",
        "check-control-plane-read-fake-store-handler-plan-v1.py",
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
    require(routes, "route contract fixture must contain routes")
    return routes


def implemented_route_ids() -> set[str]:
    if not IMPLEMENTATION_FIXTURE_PATH.exists():
        return set()
    implementation = load_json(IMPLEMENTATION_FIXTURE_PATH)
    return {
        str(route.get("route_id") or "")
        for route in implementation.get("implemented_routes") or []
        if isinstance(route, dict)
    }


def precondition_fake_store_bindings_by_route() -> dict[str, dict[str, Any]]:
    preconditions = load_json(PRECONDITIONS_FIXTURE_PATH)
    fake_store = preconditions.get("fake_store_strategy") or {}
    bindings = {
        str(item.get("route_id") or ""): item
        for item in fake_store.get("route_bindings") or []
        if isinstance(item, dict)
    }
    require(bindings, "preconditions fixture must expose fake store bindings")
    return bindings


def response_envelope_and_forbidden_keys() -> tuple[set[str], set[str], set[str]]:
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


def package_go_text(package_path: Path) -> str:
    texts: list[str] = []
    for go_file in sorted(package_path.glob("*.go")):
        texts.append(go_file.read_text(encoding="utf-8"))
    require(texts, f"{package_path.relative_to(REPO_ROOT)} must contain Go files")
    return "\n".join(texts)


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    require(
        "control-plane-read-implementation-preconditions-v1" in declared,
        "fixture must depend on read implementation preconditions",
    )
    preconditions = load_json(PRECONDITIONS_FIXTURE_PATH)
    slice_info = preconditions.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-implementation-preconditions-v1",
        "preconditions dependency id drifted",
    )
    require(
        slice_info.get("status") == "implementation_preconditions_defined",
        "preconditions dependency must remain implementation_preconditions_defined",
    )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "control_plane_read_fake_store_handler_plan_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-fake-store-handler-plan-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "fake_store_handler_plan_defined",
        "fake-store handler plan must stay plan-only",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_go_ownership(fixture: dict[str, Any]) -> None:
    ownership = fixture.get("planned_go_ownership") or {}
    require(ownership.get("status") == "plan_only_not_implemented", "Go ownership must stay plan-only")
    require(ownership.get("handler_owner") == "go_platform_control_plane_read_side", "handler owner drifted")
    implementation_route_ids = implemented_route_ids()

    package_path = REPO_ROOT / str(ownership.get("package") or "")
    require(package_path.is_dir(), "planned Go package must exist")
    route_registration = REPO_ROOT / str(ownership.get("route_registration_file") or "")
    require(route_registration.is_file(), "route registration file must exist")

    for key in ("future_handler_file", "future_fake_store_file", "future_test_file"):
        future_path = REPO_ROOT / str(ownership.get(key) or "")
        require(future_path.parent == package_path, f"{key} must stay under planned Go package")
        if implementation_route_ids:
            require(future_path.exists(), f"{future_path.relative_to(REPO_ROOT)} must exist once implementation slice is present")
        else:
            require(not future_path.exists(), f"{future_path.relative_to(REPO_ROOT)} must remain unimplemented in this slice")

    all_go = package_go_text(package_path)
    for helper in ownership.get("existing_helpers") or []:
        require(str(helper) in all_go, f"planned helper does not exist in Go package: {helper}")

    server_go = route_registration.read_text(encoding="utf-8")
    for route_id, route in route_contracts_by_id().items():
        route_path = str(route.get("path") or "")
        if route_id in implementation_route_ids:
            registration_literal = EXPECTED_ROUTE_REGISTRATION_LITERALS.get(route_id, route_path)
            require(route_path in all_go, f"implemented route path must be declared by implementation slice: {route_id}")
            require(registration_literal in server_go, f"implemented route must be registered by implementation slice: {route_id}")
        else:
            require(route_path not in server_go, f"unimplemented read route must not be registered: {route_id}")


def assert_phases(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    phases = {
        str(item.get("id") or ""): item
        for item in fixture.get("implementation_phases") or []
        if isinstance(item, dict)
    }
    require(set(phases) == EXPECTED_PHASE_IDS, "implementation phases drifted")
    for phase_id, phase in phases.items():
        require(phase.get("status") == "planned_not_implemented", f"{phase_id} must stay planned")
        route_ids = set(phase.get("route_ids") or [])
        require(route_ids.issubset(set(routes)), f"{phase_id} has unknown route ids")
        require(phase.get("planned_outputs"), f"{phase_id} must describe planned outputs")

    require(set(phases["phase-2-single-resource-summary-routes"].get("route_ids") or []) == {
        "tenant-summary-route",
        "quota-summary-route",
    }, "single-resource phase route ids drifted")
    require(set(phases["phase-4-negative-route-smoke"].get("route_ids") or []) == set(routes), "negative smoke phase must cover all routes")


def assert_route_plan(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    fake_store_bindings = precondition_fake_store_bindings_by_route()
    phases = {
        str(item.get("id") or ""): item
        for item in fixture.get("implementation_phases") or []
        if isinstance(item, dict)
    }
    route_plan = {
        str(item.get("route_id") or ""): item
        for item in fixture.get("route_plan") or []
        if isinstance(item, dict)
    }
    require(set(route_plan) == set(routes), "route plan must cover every read-only route")
    require(set(route_plan) == set(fake_store_bindings), "route plan must align with fake store bindings")

    handler_names: set[str] = set()
    for route_id, plan in route_plan.items():
        route = routes[route_id]
        require(plan.get("method") == route.get("method"), f"{route_id} method drifted")
        require(plan.get("path") == route.get("path"), f"{route_id} path drifted")
        require(plan.get("read_model") == route.get("read_model"), f"{route_id} read model drifted")
        require(plan.get("fake_store_mode") == fake_store_bindings[route_id].get("store_fixture_mode"), f"{route_id} fake store mode drifted")
        phase_id = str(plan.get("implementation_phase") or "")
        require(phase_id in phases, f"{route_id} references unknown phase")
        require(route_id in set(phases[phase_id].get("route_ids") or []), f"{route_id} phase membership drifted")
        handler = str(plan.get("future_handler") or "")
        require(handler.startswith("handle"), f"{route_id} future handler name must be explicit")
        require(handler not in handler_names, f"duplicate future handler name: {handler}")
        handler_names.add(handler)


def assert_fake_store_plan(fixture: dict[str, Any]) -> None:
    fake_store = fixture.get("fake_store_plan") or {}
    require(fake_store.get("status") == "planned_not_implemented", "fake store plan must not claim implementation")
    require(fake_store.get("strategy") == "fixture_backed_fake_store_only", "fake store strategy drifted")
    require(fake_store.get("future_binding") == "control_plane_read_fake_store.go", "fake store future binding drifted")
    for relative_path in fake_store.get("source_fixtures") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"fake store source fixture missing: {relative_path}")
    require({"single_resource_summary", "cursor_list_summary"} == set(fake_store.get("allowed_modes") or []), "allowed fake store modes drifted")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_STORE_MODES - set(fake_store.get("does_not_allow") or []))
    require(not missing_forbidden, f"fake store missing forbidden modes: {missing_forbidden}")


def assert_fake_auth_plan(fixture: dict[str, Any]) -> None:
    preconditions = load_json(PRECONDITIONS_FIXTURE_PATH)
    auth_dependency = preconditions.get("auth_middleware_dependency") or {}
    auth_plan = fixture.get("fake_auth_context_plan") or {}
    require(auth_plan.get("status") == "test_only_plan_not_middleware", "fake auth must stay test-only plan")
    require(auth_plan.get("injection_mode") == "go_test_request_context_only", "fake auth injection mode drifted")
    require("Radish OIDC" in str(auth_plan.get("production_truth_source") or ""), "auth production truth source drifted")
    require(
        set(auth_dependency.get("required_context_fields") or []).issubset(set(auth_plan.get("required_context_fields") or [])),
        "fake auth plan must include precondition auth context fields",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_PUBLIC_AUTH_INPUTS - set(auth_plan.get("forbidden_public_inputs") or []))
    require(not missing_forbidden, f"fake auth missing forbidden public inputs: {missing_forbidden}")


def assert_response_conformance_gate(fixture: dict[str, Any]) -> None:
    _, envelope, forbidden_keys = response_envelope_and_forbidden_keys()
    gate = fixture.get("response_conformance_gate") or {}
    require(gate.get("status") == "planned_not_implemented", "response conformance must stay planned")
    require(gate.get("response_source") == "control-plane-read-response-fixtures-v1", "response source drifted")
    require(set(gate.get("envelope_fields") or []) == envelope, "response envelope drifted")
    require(set(gate.get("forbidden_output_keys") or []) == forbidden_keys, "forbidden output keys drifted")
    require(
        gate.get("failure_code_source") == "control-plane-read-route-contract-v1.failure_codes",
        "failure code source drifted",
    )


def assert_route_smoke_plan(fixture: dict[str, Any]) -> None:
    negative_contract = load_json(NEGATIVE_CONTRACT_FIXTURE_PATH)
    smoke = fixture.get("route_smoke_plan") or {}
    require(smoke.get("status") == "planned_not_implemented", "route smoke must stay planned")
    require(
        smoke.get("positive_case_source") == "one_success_example_per_route_from_control-plane-read-response-fixtures-v1",
        "positive case source drifted",
    )
    require(set(smoke.get("negative_case_sources") or []) == {
        "control-plane-read-negative-contract-v1.route_negative_cases",
        "control-plane-read-negative-contract-v1.shared_negative_cases",
    }, "negative case sources drifted")
    require(
        set(negative_contract.get("required_denial_invariants") or []).issubset(
            set(smoke.get("required_denial_invariants") or [])
        ),
        "route smoke plan must include negative contract denial invariants",
    )
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - set(smoke.get("side_effect_counters") or []))
    require(not missing_counters, f"route smoke missing side-effect counters: {missing_counters}")


def assert_policy_and_docs(fixture: dict[str, Any]) -> None:
    missing_scope = sorted(EXPECTED_FORBIDDEN_SCOPE - set(fixture.get("forbidden_current_scope") or []))
    require(not missing_scope, f"missing forbidden current scope: {missing_scope}")

    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-control-plane-read-fake-store-handler-plan-v1.py", [])'
        in check_repo,
        "check-repo.py must run control plane read fake-store handler plan check",
    )

    doc_references = IMPLEMENTED_DOC_REFERENCES if implemented_route_ids() else REQUIRED_DOC_REFERENCES
    for relative_path, required_literals in doc_references.items():
        text = read(relative_path)
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_go_ownership(fixture)
    assert_phases(fixture)
    assert_route_plan(fixture)
    assert_fake_store_plan(fixture)
    assert_fake_auth_plan(fixture)
    assert_response_conformance_gate(fixture)
    assert_route_smoke_plan(fixture)
    assert_policy_and_docs(fixture)
    print("control plane read fake-store handler plan v1 checks passed.")


if __name__ == "__main__":
    main()
