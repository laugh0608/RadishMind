#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json"
)
OIDC_PRECONDITIONS_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
AUTH_STORE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json"
)
ROUTE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
NEGATIVE_CONTRACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-negative-contract-v1.json"
SELECTOR_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json"
)
DATA_BOUNDARY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-data-boundary.json"
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
EXPECTED_DEPENDENCIES = {
    "radish-oidc-client-preconditions": (OIDC_PRECONDITIONS_FIXTURE_PATH, "governance_boundary_satisfied"),
    "control-plane-read-auth-store-transition-preconditions-v1": (
        AUTH_STORE_FIXTURE_PATH,
        "auth_store_transition_preconditions_defined",
    ),
    "control-plane-read-route-contract-v1": (ROUTE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
    "control-plane-read-negative-contract-v1": (NEGATIVE_CONTRACT_FIXTURE_PATH, "governance_boundary_satisfied"),
    "control-plane-read-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_READINESS_FIXTURE_PATH,
        "store_selector_smoke_readiness_defined",
    ),
    "control-plane-data-boundary": (DATA_BOUNDARY_FIXTURE_PATH, "governance_boundary_satisfied"),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_auth_ready",
    "radish_oidc_client_ready",
    "oidc_client_implemented",
    "issuer_discovery_runtime_ready",
    "token_validation_ready",
    "auth_middleware_ready",
    "login_flow_ready",
    "logout_flow_ready",
    "session_cookie_ready",
    "tenant_permission_enforcement_ready",
    "production_api_consumer_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "store_selector_implemented",
    "formal_read_store_config_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "rate_limit_ready",
    "billing_ready",
    "cost_ledger_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_PLANNED_ARTIFACTS = {
    "contracts/radish-oidc-token-validation.schema.json",
    "services/platform/internal/httpapi/control_plane_read_auth_middleware.go",
    "services/platform/internal/httpapi/control_plane_read_auth_middleware_test.go",
    "scripts/checks/fixtures/control-plane-read-production-auth-v1.json",
    "scripts/checks/control_plane/check-control-plane-read-production-auth-v1.py",
}
EXPECTED_GATE_IDS = {
    "radish_oidc_preconditions_consumed",
    "auth_store_transition_preconditions_consumed",
    "route_contract_consumed",
    "negative_contract_consumed",
    "store_selector_smoke_readiness_consumed",
    "issuer_discovery_evidence_contract_defined",
    "token_validation_contract_preconditions_defined",
    "claim_mapping_contract_defined",
    "tenant_binding_policy_defined",
    "scope_projection_matrix_defined",
    "production_auth_smoke_gate",
    "auth_middleware_implementation_gate",
    "production_api_consumer_gate",
    "no_production_auth_artifacts_leaked",
}
SATISFIED_GATE_IDS = {
    "radish_oidc_preconditions_consumed",
    "auth_store_transition_preconditions_consumed",
    "route_contract_consumed",
    "negative_contract_consumed",
    "store_selector_smoke_readiness_consumed",
    "issuer_discovery_evidence_contract_defined",
    "token_validation_contract_preconditions_defined",
    "claim_mapping_contract_defined",
    "tenant_binding_policy_defined",
    "scope_projection_matrix_defined",
}
NOT_SATISFIED_GATE_IDS = {
    "production_auth_smoke_gate",
    "auth_middleware_implementation_gate",
    "production_api_consumer_gate",
}
EXPECTED_ISSUER_FIELDS = {
    "issuer_url",
    "discovery_document_url",
    "jwks_uri",
    "supported_signing_algorithms",
    "supported_scopes",
    "fetched_at",
    "expires_at",
    "environment",
    "operator_review_ref",
    "sanitized_evidence_ref",
}
EXPECTED_TOKEN_GATE_IDS = {
    "issuer-metadata-pinned",
    "jwks-refresh-policy",
    "algorithm-allowlist",
    "issuer-audience-checks",
    "time-window-checks",
    "required-claim-checks",
    "sanitized-failure-envelope",
}
EXPECTED_REQUIRED_CLAIMS = {"sub", "tenant_id", "roles", "permissions"}
EXPECTED_OPTIONAL_CLAIMS = {"email", "preferred_username"}
EXPECTED_FAILURE_CODES = {
    "identity_context_missing",
    "tenant_binding_missing",
    "tenant_binding_denied",
    "scope_denied",
    "auth_context_contract_mismatch",
    "issuer_unavailable",
    "issuer_metadata_invalid",
    "jwks_unavailable",
    "token_signature_invalid",
    "token_expired",
    "token_not_yet_valid",
    "token_audience_invalid",
    "token_issuer_invalid",
    "unsupported_token_algorithm",
    "required_claim_missing",
    "permission_claim_denied",
    "malformed_authorization_header",
    "logout_failed",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "database_write",
    "api_key_issue",
    "quota_mutation",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "control_plane_read_auth_middleware.go",
    "control_plane_read_auth_middleware_test.go",
    "control-plane-read-production-auth-v1.json",
    "check-control-plane-read-production-auth-v1.py",
    "oidc.Provider",
    "github.com/coreos/go-oidc",
    "ValidateToken",
    "VerifyToken",
    "NewRadishOIDCClient",
    "TokenValidationContract",
    "RADISHMIND_RADISH_OIDC_ISSUER",
    "RADISHMIND_CONTROL_PLANE_PRODUCTION_AUTH",
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


def oidc_preconditions_by_id() -> dict[str, dict[str, Any]]:
    oidc = load_json(OIDC_PRECONDITIONS_FIXTURE_PATH)
    preconditions = {
        str(item.get("id") or ""): item
        for item in oidc.get("preconditions") or []
        if isinstance(item, dict)
    }
    require("issuer-discovery" in preconditions, "OIDC issuer preconditions missing")
    require("claim-mapping" in preconditions, "OIDC claim mapping preconditions missing")
    require("failure-taxonomy" in preconditions, "OIDC failure taxonomy missing")
    return preconditions


def assert_dependencies(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(set(EXPECTED_DEPENDENCIES) - declared)
    require(not missing, f"missing dependencies: {missing}")
    for expected_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
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
        fixture.get("kind") == "control_plane_read_production_auth_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-production-auth-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "production_auth_readiness_defined",
        "production auth readiness status drifted",
    )
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_forbidden, f"missing forbidden claims: {missing_forbidden}")


def assert_production_auth_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("production_auth_boundary") or {}
    transition_boundary = load_json(AUTH_STORE_FIXTURE_PATH).get("transition_boundary") or {}
    require(
        boundary.get("status") == "production_auth_readiness_defined_not_implemented",
        "production auth boundary status drifted",
    )
    require(
        boundary.get("current_auth_source") == transition_boundary.get("current_auth_source"),
        "current auth source must match auth/store transition",
    )
    require(
        boundary.get("future_auth_source") == transition_boundary.get("future_auth_source"),
        "future auth source must match auth/store transition",
    )
    require(
        boundary.get("dev_live_source_must_remain_opt_in")
        == transition_boundary.get("dev_live_source_must_remain_opt_in"),
        "dev live opt-in env drifted",
    )
    require(
        boundary.get("backend_dev_auth_must_remain_opt_in")
        == transition_boundary.get("backend_dev_auth_must_remain_opt_in"),
        "backend dev auth env drifted",
    )
    for field in (
        "issuer_network_call_allowed_in_this_slice",
        "oidc_client_created_in_this_slice",
        "token_validation_allowed_in_this_slice",
        "auth_middleware_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "write_allowed_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_planned_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(item.get("path") or ""): item
        for item in fixture.get("planned_production_auth_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_ARTIFACTS, "planned production auth artifacts drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if relative_path == "contracts/radish-oidc-token-validation.schema.json":
            continue
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist in this readiness slice")


def assert_readiness_gates(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("readiness_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "production auth readiness gate ids drifted")
    for gate_id in SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    for gate_id in NOT_SATISFIED_GATE_IDS:
        require(gates[gate_id].get("status") == "not_satisfied", f"{gate_id} must remain not_satisfied")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    require(
        gates["no_production_auth_artifacts_leaked"].get("status") == "required_now",
        "artifact leak gate status drifted",
    )


def assert_issuer_and_token_contracts(fixture: dict[str, Any]) -> None:
    oidc_preconditions = oidc_preconditions_by_id()
    issuer = fixture.get("issuer_discovery_evidence_contract") or {}
    require(issuer.get("status") == "required_before_production_auth", "issuer evidence status drifted")
    require(issuer.get("network_call_allowed_now") is False, "issuer network call must remain false")
    require(
        issuer.get("committed_raw_discovery_document_allowed") is False,
        "raw discovery document must not be committed",
    )
    require(issuer.get("committed_raw_jwks_allowed") is False, "raw JWKS must not be committed")
    required_fields = set(issuer.get("required_fields") or [])
    require(EXPECTED_ISSUER_FIELDS.issubset(required_fields), "issuer evidence missing required fields")
    upstream_fields = set(oidc_preconditions["issuer-discovery"].get("required_fields") or [])
    require(
        {"issuer_url", "discovery_document_url", "jwks_uri"}.issubset(upstream_fields),
        "upstream issuer preconditions drifted",
    )

    token_gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("token_validation_contract_preconditions") or []
        if isinstance(gate, dict)
    }
    require(set(token_gates) == EXPECTED_TOKEN_GATE_IDS, "token validation precondition ids drifted")
    for gate_id, gate in token_gates.items():
        require(
            gate.get("status") == "required_before_token_validation",
            f"{gate_id} must remain required before token validation",
        )
        require(str(gate.get("rule") or "").strip(), f"{gate_id}.rule is required")


def assert_claim_mapping_and_tenant_policy(fixture: dict[str, Any]) -> None:
    upstream_required_claims = set(oidc_preconditions_by_id()["claim-mapping"].get("required_claims") or [])
    require(
        EXPECTED_REQUIRED_CLAIMS.issubset(upstream_required_claims),
        "upstream required claims must retain subject, tenant, roles and permissions",
    )
    claim_rows = {
        str(row.get("claim") or ""): row
        for row in fixture.get("claim_mapping_matrix") or []
        if isinstance(row, dict)
    }
    require(
        set(claim_rows) == EXPECTED_REQUIRED_CLAIMS | EXPECTED_OPTIONAL_CLAIMS,
        "claim mapping matrix drifted",
    )
    for claim in EXPECTED_REQUIRED_CLAIMS:
        row = claim_rows[claim]
        require(row.get("required") is True, f"{claim} must remain required")
        require(row.get("failure_code_when_missing"), f"{claim} must define missing failure")
    for claim in EXPECTED_OPTIONAL_CLAIMS:
        row = claim_rows[claim]
        require(row.get("required") is False, f"{claim} must remain optional")
        require(row.get("allowed_in_response") is False, f"{claim} must not be emitted in response")
    require(claim_rows["tenant_id"].get("allowed_in_response") is True, "tenant ref may remain in envelope only")
    require(claim_rows["permissions"].get("failure_code_when_missing") == "scope_denied", "scope failure drifted")

    policy = fixture.get("tenant_binding_policy") or {}
    require(policy.get("status") == "required_before_production_auth", "tenant binding status drifted")
    require(policy.get("tenant_source") == "trusted auth context claim", "tenant source drifted")
    for field in (
        "query_string_tenant_override_allowed",
        "path_tenant_override_allowed",
        "fallback_to_default_tenant_allowed",
        "fallback_to_local_admin_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    require(policy.get("missing_tenant_failure_code") == "tenant_binding_missing", "missing tenant failure drifted")
    require(policy.get("cross_tenant_failure_code") == "tenant_binding_denied", "cross tenant failure drifted")


def assert_route_scope_projection(fixture: dict[str, Any]) -> None:
    routes = route_contracts_by_id()
    rows = {
        str(row.get("route_id") or ""): row
        for row in fixture.get("route_scope_projection_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_ROUTE_IDS, "route scope projection matrix must cover every read route")
    for route_id, row in rows.items():
        route = routes[route_id]
        require(row.get("method") == route.get("method"), f"{route_id} method drifted")
        require(row.get("path") == route.get("path"), f"{route_id} path drifted")
        require(row.get("required_scope") == route.get("required_scope"), f"{route_id} required scope drifted")
        require(
            row.get("scope_source") == "permissions claim projected into read auth context",
            f"{route_id} scope source drifted",
        )
        require(row.get("missing_identity_failure_code") == "identity_context_missing", f"{route_id} identity drifted")
        require(row.get("tenant_binding_failure_code") == "tenant_binding_missing", f"{route_id} tenant drifted")
        require(row.get("scope_denied_failure_code") == "scope_denied", f"{route_id} scope failure drifted")
        require(row.get("production_auth_smoke_required") is True, f"{route_id} auth smoke must be required")
        for field in ("production_auth_ready", "fake_auth_header_allowed_in_production", "side_effect_allowed"):
            require(row.get(field) is False, f"{route_id} {field} must remain false")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "production auth failure mapping missing codes")
    oidc_failures = set(oidc_preconditions_by_id()["failure-taxonomy"].get("required_failures") or [])
    require(oidc_failures.issubset(failures), "production auth readiness must include OIDC failures")
    negative_failures = {
        str(case.get("expected_response", {}).get("failure_code") or "")
        for case in load_json(NEGATIVE_CONTRACT_FIXTURE_PATH).get("route_negative_cases") or []
        if isinstance(case, dict)
    }
    require(
        {"identity_context_missing", "tenant_binding_missing", "scope_denied"}.issubset(negative_failures),
        "negative contract auth failures drifted",
    )
    require(negative_failures & failures, "production auth readiness must share negative contract failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "dev_fake_auth_header_allowed_in_production",
        "missing_bearer_token_fallback_to_fake_auth_allowed",
        "invalid_token_fallback_to_fake_auth_allowed",
        "scope_denied_fallback_to_admin_allowed",
        "tenant_binding_failure_fallback_allowed",
        "local_admin_bypass_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

    side_effects = fixture.get("no_side_effect_policy") or {}
    require(
        EXPECTED_SIDE_EFFECT_COUNTERS.issubset(set(side_effects.get("side_effect_counters_must_remain") or [])),
        "missing zero side-effect counters",
    )
    require(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS.issubset(set(side_effects.get("forbidden_side_effects") or [])),
        "missing forbidden side effects",
    )


def assert_docs_check_repo_and_no_leak(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-control-plane-read-store-selector-smoke-readiness-v1.py"
    current_checker = "check-control-plane-read-production-auth-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run production auth readiness check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "production auth readiness check must run after store selector smoke readiness",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured), "source absent literals drifted")
    for root in (REPO_ROOT / "services/platform/internal", REPO_ROOT / "apps/radishmind-web/src"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(
                    literal not in text,
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this readiness slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_production_auth_boundary(fixture)
    assert_planned_artifacts(fixture)
    assert_readiness_gates(fixture)
    assert_issuer_and_token_contracts(fixture)
    assert_claim_mapping_and_tenant_policy(fixture)
    assert_route_scope_projection(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_check_repo_and_no_leak(fixture)
    print("control plane read production auth readiness v1 checks passed.")


if __name__ == "__main__":
    main()
