#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-readiness-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "radish-oidc-client-preconditions": (
        REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-auth-store-transition-preconditions-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-auth-store-transition-preconditions-v1.json",
        "auth_store_transition_preconditions_defined",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json",
        "production_auth_readiness_defined",
    ),
    "workflow-saved-draft-production-auth-readiness-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json",
        "draft_production_auth_readiness_defined",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "radish_oidc_ready",
    "oidc_client_implemented",
    "issuer_discovery_runtime_ready",
    "token_validation_ready",
    "auth_middleware_ready",
    "login_flow_ready",
    "logout_flow_ready",
    "session_cookie_ready",
    "membership_adapter_ready",
    "workspace_membership_adapter_ready",
    "tenant_permission_enforcement_ready",
    "repository_mode_ready",
    "database_connection_ready",
    "database_query_ready",
    "production_api_consumer_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
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
EXPECTED_VALIDATION_GATES = {
    "issuer-metadata-pinned",
    "jwks-refresh-policy",
    "algorithm-allowlist",
    "issuer-audience-client-checks",
    "time-window-checks",
    "required-claim-checks",
    "permission-claim-checks",
    "sanitized-failure-envelope",
}
EXPECTED_REQUIRED_CLAIMS = {"sub", "tenant_id", "roles", "permissions"}
EXPECTED_OPTIONAL_CLAIMS = {"email", "preferred_username"}
EXPECTED_MEMBERSHIP_INPUT_FIELDS = {
    "request_id",
    "issuer_ref",
    "actor_subject_ref",
    "tenant_ref",
    "role_refs",
    "permission_refs",
    "workspace_id",
    "application_id",
    "owner_subject_ref",
    "audit_ref",
}
EXPECTED_MEMBERSHIP_RESULT_FIELDS = {
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
    "membership_verified",
    "application_scope_verified",
    "owner_scope_verified",
    "audit_ref",
}
EXPECTED_MEMBERSHIP_CHECKS = {
    "tenant-binding",
    "workspace-membership",
    "application-scope",
    "owner-scope",
    "operation-scope-grant",
    "audit-context",
}
EXPECTED_CONSUMERS = {
    "control-plane-read",
    "saved-workflow-draft",
    "admin-control-plane",
    "model-gateway-api-distribution",
}
EXPECTED_FAILURE_CODES = {
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
    "tenant_binding_missing",
    "tenant_binding_denied",
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
    "draft_auth_context_contract_mismatch",
    "draft_audit_context_missing",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "issuer_network_call_count=0",
    "token_validation_call_count=0",
    "membership_query_count=0",
    "database_write_count=0",
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "issuer_network_call",
    "token_validation_runtime_call",
    "membership_database_query",
    "auth_session_mutation",
    "database_write",
    "repository_write",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "run_replay",
}
REQUIRED_DOC_REFERENCES = {
    "docs/integrations/README.md": [
        "Radish OIDC Token / Membership Readiness v1",
        "radish_oidc_token_membership_readiness_defined",
    ],
    "docs/platform/README.md": [
        "Radish OIDC token / membership readiness",
        "radish_oidc_token_membership_readiness_defined",
    ],
    "docs/features/workflow/saved-workflow-draft-v1.md": [
        "radish_oidc_token_membership_readiness_defined",
        "Radish OIDC token / membership readiness",
    ],
    "docs/radishmind-current-focus.md": [
        "radish_oidc_token_membership_readiness_defined",
        "radish-oidc-token-membership-readiness-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Radish OIDC Token / Membership Readiness",
        "radish-oidc-token-membership-readiness-v1",
    ],
    "scripts/README.md": [
        "check-radish-oidc-token-membership-readiness-v1.py",
        "radish-oidc-token-membership-readiness-v1.json",
    ],
    "docs/devlogs/2026-W26.md": [
        "radish-oidc-token-membership-readiness-v1",
        "radish_oidc_token_membership_readiness_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def assert_dependencies(fixture: dict[str, Any]) -> None:
    depends_on = set(fixture.get("depends_on") or [])
    missing_dependencies = sorted(set(EXPECTED_DEPENDENCIES) - depends_on)
    require(not missing_dependencies, f"missing dependencies: {missing_dependencies}")

    for dependency_id, (path, expected_status) in EXPECTED_DEPENDENCIES.items():
        dependency = load_json(path)
        slice_info = dependency.get("slice") or {}
        require(slice_info.get("id") == dependency_id, f"{dependency_id} fixture id drifted")
        require(
            slice_info.get("status") == expected_status,
            f"{dependency_id} must remain {expected_status}",
        )


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "radish_oidc_token_membership_readiness_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "radish-oidc-token-membership-readiness-v1", "unexpected slice id")
    require(slice_info.get("track") == "Integrations / Radish OIDC", "unexpected track")
    require(
        slice_info.get("status") == "radish_oidc_token_membership_readiness_defined",
        "unexpected readiness status",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_token_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("token_validation_contract") or {}
    require(contract.get("status") == "contract_defined_not_implemented", "token contract must not be implemented")
    require(
        EXPECTED_ISSUER_FIELDS.issubset(set(contract.get("issuer_evidence_required_fields") or [])),
        "issuer evidence fields are incomplete",
    )
    require(
        EXPECTED_VALIDATION_GATES.issubset(set(contract.get("validation_gates") or [])),
        "token validation gates are incomplete",
    )
    require(
        EXPECTED_REQUIRED_CLAIMS.issubset(set(contract.get("required_claims") or [])),
        "required claims are incomplete",
    )
    require(
        EXPECTED_OPTIONAL_CLAIMS.issubset(set(contract.get("optional_claims") or [])),
        "optional claims are incomplete",
    )
    require(contract.get("runtime_created") is False, "token runtime must not be created")
    require(contract.get("issuer_network_call_allowed") is False, "issuer network calls must stay disabled")
    require(contract.get("raw_claim_dump_allowed") is False, "raw claim dumps must stay forbidden")


def assert_membership_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("membership_contract") or {}
    require(contract.get("status") == "contract_defined_not_implemented", "membership contract must not be implemented")
    require(
        EXPECTED_MEMBERSHIP_INPUT_FIELDS.issubset(set(contract.get("input_context_fields") or [])),
        "membership input fields are incomplete",
    )
    require(
        EXPECTED_MEMBERSHIP_RESULT_FIELDS.issubset(set(contract.get("result_fields") or [])),
        "membership result fields are incomplete",
    )
    require(
        EXPECTED_MEMBERSHIP_CHECKS.issubset(set(contract.get("required_checks") or [])),
        "membership checks are incomplete",
    )
    require(contract.get("adapter_created") is False, "membership adapter must not be created")
    require(contract.get("database_query_allowed") is False, "membership database query must stay disabled")
    require(contract.get("local_admin_fallback_allowed") is False, "local admin fallback must stay forbidden")
    require(contract.get("dev_fake_auth_fallback_allowed") is False, "dev fake auth fallback must stay forbidden")


def assert_consumer_and_failure_mapping(fixture: dict[str, Any]) -> None:
    consumers = {str(item.get("consumer") or ""): item for item in fixture.get("consumer_matrix") or []}
    require(set(consumers) == EXPECTED_CONSUMERS, f"unexpected consumers: {sorted(consumers)}")
    for consumer_id, item in consumers.items():
        require(item.get("readiness_conclusion") == "defined_not_implemented", f"{consumer_id} must not be implemented")
        require(item.get("implementation_opened") is False, f"{consumer_id} implementation must stay closed")

    failure_mapping = fixture.get("failure_mapping") or {}
    require(failure_mapping.get("status") == "failure_mapping_defined", "failure mapping status drifted")
    missing_failure_codes = sorted(
        EXPECTED_FAILURE_CODES - set(failure_mapping.get("required_failure_codes") or [])
    )
    require(not missing_failure_codes, f"missing failure codes: {missing_failure_codes}")
    require(failure_mapping.get("sanitized_envelope_required") is True, "sanitized envelope is required")
    require(failure_mapping.get("raw_token_leak_allowed") is False, "raw token leakage must stay forbidden")
    require(failure_mapping.get("raw_claim_leak_allowed") is False, "raw claim leakage must stay forbidden")
    require(failure_mapping.get("database_detail_leak_allowed") is False, "database detail leakage must stay forbidden")


def assert_artifacts_and_side_effects(fixture: dict[str, Any]) -> None:
    for item in fixture.get("planned_artifacts") or []:
        require(item.get("created_in_this_slice") is False, f"{item.get('path')} must not be created in this slice")
        path = REPO_ROOT / str(item.get("path") or "")
        require(not path.exists(), f"future artifact exists too early: {path.relative_to(REPO_ROOT)}")

    fallback_policy = fixture.get("no_fallback_policy") or {}
    for key, value in fallback_policy.items():
        require(value is False, f"{key} must remain false")

    side_effect_policy = fixture.get("no_side_effect_policy") or {}
    require(side_effect_policy.get("status") == "readiness_only", "side effect policy status drifted")
    missing_counters = sorted(
        EXPECTED_SIDE_EFFECT_COUNTERS - set(side_effect_policy.get("side_effect_counters_must_remain") or [])
    )
    require(not missing_counters, f"missing side effect counters: {missing_counters}")
    missing_side_effects = sorted(
        EXPECTED_FORBIDDEN_SIDE_EFFECTS - set(side_effect_policy.get("forbidden_side_effects") or [])
    )
    require(not missing_side_effects, f"missing forbidden side effects: {missing_side_effects}")


def assert_evidence_and_docs(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing required consumer: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-radish-oidc-token-membership-readiness-v1.py", [])'
        in check_repo,
        "check-repo.py must run Radish OIDC token / membership readiness check",
    )

    for relative_path, required_literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing_literals = [literal for literal in required_literals if literal not in text]
        require(not missing_literals, f"{relative_path} missing literals: {missing_literals}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_token_contract(fixture)
    assert_membership_contract(fixture)
    assert_consumer_and_failure_mapping(fixture)
    assert_artifacts_and_side_effects(fixture)
    assert_evidence_and_docs(fixture)
    print("Radish OIDC token / membership readiness checks passed.")


if __name__ == "__main__":
    main()
