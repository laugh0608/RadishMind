#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-upstream-evidence-refresh-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "radish-oidc-token-membership-readiness-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-readiness-v1.json",
        "radish_oidc_token_membership_readiness_defined",
    ),
    "radish-oidc-token-membership-implementation-entry-review-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-implementation-entry-review-v1.json",
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ),
    "radish-oidc-client-preconditions": (
        "scripts/checks/fixtures/radish-oidc-client-preconditions.json",
        "governance_boundary_satisfied",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json",
        "production_auth_readiness_defined",
    ),
    "workflow-saved-draft-production-auth-readiness-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json",
        "draft_production_auth_readiness_defined",
    ),
    "workflow-saved-draft-production-auth-runtime-v1": (
        "scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json",
        "draft_production_auth_runtime_bridge_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "radish_oidc_ready",
    "oidc_client_implemented",
    "issuer_discovery_fetched",
    "jwks_fetched",
    "jwks_cache_created",
    "client_registration_loaded_into_runtime",
    "token_validation_schema_created",
    "token_validator_created",
    "auth_middleware_ready",
    "auth_middleware_created",
    "login_flow_ready",
    "logout_flow_ready",
    "session_cookie_ready",
    "membership_adapter_ready",
    "membership_adapter_created",
    "membership_cache_created",
    "negative_auth_smoke_created",
    "production_api_consumer_ready",
    "production_api_consumer_created",
    "repository_mode_ready",
    "database_connection_ready",
    "database_query_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_REFRESH_FALSE_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "issuer_network_call_allowed_in_this_slice",
    "jwks_fetch_allowed_in_this_slice",
    "token_validation_allowed_in_this_slice",
    "auth_middleware_created_in_this_slice",
    "membership_adapter_created_in_this_slice",
    "membership_cache_created_in_this_slice",
    "negative_auth_smoke_created_in_this_slice",
    "production_api_consumer_created_in_this_slice",
    "repository_mode_enabled_in_this_slice",
}
EXPECTED_EVIDENCE_STATUS = {
    "reviewed_issuer_evidence": "static_evidence_contract_defined_no_network",
    "jwks_pin_refresh_policy": "static_policy_defined_no_jwks_fetch",
    "client_registration_evidence": "static_contract_defined_no_client_runtime",
    "auth_middleware_ownership": "static_owner_defined_no_middleware",
    "membership_source_ownership": "static_owner_defined_no_adapter",
    "membership_cache_policy": "static_policy_defined_no_cache_runtime",
    "negative_auth_smoke_matrix": "static_matrix_defined_no_runtime_smoke",
}
EXPECTED_REQUIRED_FIELDS = {
    "reviewed_issuer_evidence": {
        "issuer_url",
        "discovery_document_url",
        "jwks_uri",
        "supported_signing_algorithms",
        "supported_scopes",
        "environment",
        "fetched_at",
        "expires_at",
        "operator_review_ref",
        "sanitized_evidence_ref",
        "policy_version",
    },
    "jwks_pin_refresh_policy": {
        "jwks_uri_ref",
        "pin_set_ref",
        "refresh_interval_policy",
        "cache_expiry_policy",
        "rotation_failure_policy",
        "key_id_allowlist_ref",
        "operator_review_ref",
    },
    "client_registration_evidence": {
        "client_id_ref",
        "allowed_audiences",
        "allowed_scopes",
        "redirect_policy_ref",
        "environment",
        "operator_review_ref",
        "secret_ref_status",
    },
    "auth_middleware_ownership": {
        "route_owner_ref",
        "dev_fake_auth_isolation_policy",
        "public_fake_auth_forbidden",
        "failure_envelope_owner_ref",
        "audit_ref",
    },
    "membership_source_ownership": {
        "source_owner_ref",
        "tenant_binding_owner_ref",
        "workspace_membership_owner_ref",
        "application_scope_owner_ref",
        "owner_scope_owner_ref",
        "cache_policy_ref",
        "audit_ref",
    },
    "membership_cache_policy": {
        "cache_scope",
        "cache_key_inputs",
        "expiry_policy",
        "revocation_policy",
        "stale_result_failure_policy",
        "audit_ref",
    },
    "negative_auth_smoke_matrix": {
        "case_id",
        "failure_boundary",
        "expected_failure_code",
        "sanitized_diagnostic",
        "audit_ref",
    },
}
EXPECTED_NEGATIVE_CASES = {
    "missing_authorization_header",
    "malformed_bearer_header",
    "unknown_issuer",
    "jwks_unavailable",
    "invalid_signature",
    "expired_token",
    "missing_required_claim",
    "tenant_binding_denied",
    "workspace_membership_denied",
    "application_scope_denied",
    "owner_scope_denied",
    "scope_grant_missing",
    "membership_source_unavailable",
}
EXPECTED_RUNTIME_BLOCKERS = {
    "token_validation_schema_missing",
    "auth_middleware_not_created",
    "membership_adapter_not_created",
    "negative_auth_smoke_runtime_missing",
    "repository_mode_disabled",
    "production_api_consumer_gate_blocked",
}
EXPECTED_FAILURE_CODES = {
    "radish_oidc_upstream_evidence_missing",
    "radish_oidc_jwks_policy_missing",
    "radish_oidc_client_registration_missing",
    "radish_oidc_membership_source_missing",
    "radish_oidc_negative_auth_smoke_matrix_missing",
    "radish_oidc_runtime_artifact_forbidden",
}
EXPECTED_ALLOWED_DIAGNOSTICS = {
    "entry_decision",
    "evidence_id",
    "case_id",
    "failure_code",
    "failure_boundary",
    "sanitized_diagnostic",
    "request_id",
    "audit_ref",
    "policy_version",
}
EXPECTED_FORBIDDEN_DIAGNOSTICS = {
    "raw_token",
    "authorization_header",
    "cookie",
    "client_secret",
    "jwks_raw_dump",
    "raw_claim_dump",
    "database_detail",
    "membership_raw_record",
    "full_user_claim",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "issuer_network_call_count=0",
    "jwks_fetch_count=0",
    "token_validation_call_count=0",
    "membership_query_count=0",
    "membership_cache_create_count=0",
    "database_write_count=0",
    "repository_write_count=0",
    "production_api_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_TEST_FLAGS = {
    "does_not_start_service",
    "does_not_call_oidc",
    "does_not_fetch_issuer_discovery",
    "does_not_fetch_jwks",
    "does_not_validate_token",
    "does_not_query_membership",
    "does_not_create_membership_cache",
    "does_not_create_auth_middleware",
    "does_not_create_membership_adapter",
    "does_not_create_production_api",
}
REQUIRED_DOC_REFERENCES = {
    "docs/integrations/README.md": [
        "Radish OIDC Token / Membership Upstream Evidence Refresh v1",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
    "docs/platform/README.md": [
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
    "docs/features/README.md": [
        "Radish OIDC Token / Membership Upstream Evidence Refresh v1",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
    "docs/features/workflow/saved-workflow-draft-v1.md": [
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
    "docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md": [
        "Radish OIDC Token / Membership Upstream Evidence Refresh v1",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
    "docs/radishmind-current-focus.md": [
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
        "radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Radish OIDC Token / Membership Upstream Evidence Refresh",
        "radish-oidc-token-membership-upstream-evidence-refresh-v1",
    ],
    "scripts/README.md": [
        "check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py",
        "radish-oidc-token-membership-upstream-evidence-refresh-v1.json",
    ],
    "docs/devlogs/2026-W26.md": [
        "radish-oidc-token-membership-upstream-evidence-refresh-v1",
        "radish_oidc_token_membership_upstream_evidence_refresh_defined",
    ],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must contain a JSON object")
    return document


def rows_by_id(fixture: dict[str, Any], key: str, id_field: str) -> dict[str, dict[str, Any]]:
    rows = {str(row.get(id_field) or ""): row for row in fixture.get(key) or [] if isinstance(row, dict)}
    require(rows, f"{key} must not be empty")
    return rows


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "radish_oidc_token_membership_upstream_evidence_refresh_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "radish-oidc-token-membership-upstream-evidence-refresh-v1", "unexpected id")
    require(
        slice_info.get("status") == "radish_oidc_token_membership_upstream_evidence_refresh_defined",
        "unexpected status",
    )
    require(slice_info.get("entry_decision") == "blocked_before_runtime_task_card", "entry decision drifted")
    for key in ("task_card", "integration_topic"):
        value = str(slice_info.get(key) or "")
        require(value, f"{key} is required")
        require((REPO_ROOT / value).exists(), f"{key} path missing: {value}")
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")


def assert_dependencies(fixture: dict[str, Any]) -> None:
    dependencies = rows_by_id(fixture, "depends_on", "id")
    missing = sorted(set(EXPECTED_DEPENDENCIES) - set(dependencies))
    require(not missing, f"missing dependencies: {missing}")
    for dependency_id, (relative_path, expected_status) in EXPECTED_DEPENDENCIES.items():
        item = dependencies[dependency_id]
        require(item.get("status") == expected_status, f"{dependency_id} fixture status drifted")
        evidence = str(item.get("evidence") or "")
        require(evidence == relative_path, f"{dependency_id} evidence path drifted")
        evidence_doc = load_json(REPO_ROOT / evidence)
        require((evidence_doc.get("slice") or {}).get("status") == expected_status, f"{dependency_id} evidence drifted")


def assert_refresh_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("upstream_evidence_refresh") or {}
    require(
        boundary.get("status") == "static_upstream_evidence_contract_defined_no_runtime",
        "refresh status drifted",
    )
    require(boundary.get("entry_decision") == "blocked_before_runtime_task_card", "refresh decision drifted")
    for flag in EXPECTED_REFRESH_FALSE_FLAGS:
        require(boundary.get(flag) is False, f"{flag} must remain false")


def assert_evidence_contracts(fixture: dict[str, Any]) -> None:
    contracts = rows_by_id(fixture, "evidence_contract_matrix", "evidence_id")
    require(set(contracts) == set(EXPECTED_EVIDENCE_STATUS), "evidence ids drifted")
    for evidence_id, expected_status in EXPECTED_EVIDENCE_STATUS.items():
        item = contracts[evidence_id]
        require(item.get("status") == expected_status, f"{evidence_id} status drifted")
        require(item.get("runtime_artifact_allowed_now") is False, f"{evidence_id} runtime must remain forbidden")
        missing_fields = sorted(EXPECTED_REQUIRED_FIELDS[evidence_id] - set(item.get("required_fields") or []))
        require(not missing_fields, f"{evidence_id} missing required fields: {missing_fields}")
        forbidden_fields = set(item.get("must_not_include") or [])
        require(forbidden_fields, f"{evidence_id} must declare forbidden fields")
        leaked_overlap = sorted(set(item.get("required_fields") or []) & forbidden_fields)
        require(not leaked_overlap, f"{evidence_id} required fields include forbidden fields: {leaked_overlap}")


def assert_negative_cases_and_blockers(fixture: dict[str, Any]) -> None:
    cases = rows_by_id(fixture, "negative_auth_smoke_cases", "case_id")
    require(set(cases) == EXPECTED_NEGATIVE_CASES, f"negative auth case ids drifted: {sorted(cases)}")
    for case_id, item in cases.items():
        require(str(item.get("failure_boundary") or "").strip(), f"{case_id} failure boundary is required")

    missing_blockers = sorted(EXPECTED_RUNTIME_BLOCKERS - set(fixture.get("remaining_runtime_blockers") or []))
    require(not missing_blockers, f"missing runtime blockers: {missing_blockers}")


def assert_failure_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(str(item.get("failure_boundary") or "").strip(), f"{code} failure boundary is required")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic.strip(), f"{code} sanitized diagnostic is required")
        lowered = diagnostic.lower()
        for forbidden in ("authorization_header", "cookie", "client_secret", "raw_claim", "jwks_raw_dump"):
            require(forbidden not in lowered, f"{code} diagnostic leaks forbidden term: {forbidden}")

    policy = fixture.get("diagnostic_policy") or {}
    missing_allowed = sorted(EXPECTED_ALLOWED_DIAGNOSTICS - set(policy.get("allowed_fields") or []))
    require(not missing_allowed, f"missing allowed diagnostics: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(policy.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostics: {missing_forbidden}")


def assert_artifact_guard_and_testing(fixture: dict[str, Any]) -> None:
    for item in fixture.get("planned_artifacts") or []:
        path_text = str(item.get("path") or "")
        require(path_text, "planned artifact path is required")
        require(item.get("created_in_this_slice") is False, f"{path_text} must not be created in this slice")
        if path_text == "contracts/radish-oidc-token-validation.schema.json":
            continue
        require(not (REPO_ROOT / path_text).exists(), f"future artifact exists too early: {path_text}")

    counters = set((fixture.get("no_side_effect_policy") or {}).get("side_effect_counters_must_remain") or [])
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - counters)
    require(not missing_counters, f"missing side effect counters: {missing_counters}")

    strategy = fixture.get("testing_strategy") or {}
    require(
        strategy.get("status") == "upstream_evidence_refresh_checker_only_no_runtime",
        "testing strategy status drifted",
    )
    for flag in EXPECTED_TEST_FLAGS:
        require(strategy.get(flag) is True, f"{flag} must remain true")
    for expected_check in (
        "radish_oidc_token_membership_upstream_evidence_refresh_checker",
        "radish_oidc_token_membership_implementation_entry_review_checker",
        "radish_oidc_token_membership_readiness_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in set(strategy.get("required_checks") or []), f"missing required check: {expected_check}")


def assert_docs_and_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py"
    current_checker = "checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py"
    next_checker = "checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py"
    require(previous_checker in check_repo, "implementation entry review checker missing from check-repo")
    require(current_checker in check_repo, "upstream evidence refresh checker missing from check-repo")
    require(next_checker in check_repo, "repository mode enablement checker missing from check-repo")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "upstream evidence refresh checker order drifted",
    )

    for relative_path, literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_refresh_boundary(fixture)
    assert_evidence_contracts(fixture)
    assert_negative_cases_and_blockers(fixture)
    assert_failure_and_diagnostics(fixture)
    assert_artifact_guard_and_testing(fixture)
    assert_docs_and_registration()
    print("Radish OIDC token / membership upstream evidence refresh checks passed.")


if __name__ == "__main__":
    main()
