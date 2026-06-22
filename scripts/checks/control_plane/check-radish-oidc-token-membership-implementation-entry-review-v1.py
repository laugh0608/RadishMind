#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/radish-oidc-token-membership-implementation-entry-review-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "radish-oidc-token-membership-readiness-v1": (
        "scripts/checks/fixtures/radish-oidc-token-membership-readiness-v1.json",
        "radish_oidc_token_membership_readiness_defined",
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
    "token_validation_schema_created",
    "token_validator_created",
    "auth_middleware_ready",
    "auth_middleware_created",
    "login_flow_ready",
    "logout_flow_ready",
    "session_cookie_ready",
    "membership_adapter_ready",
    "membership_adapter_created",
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
EXPECTED_FALSE_ENTRY_FLAGS = {
    "runtime_task_card_created_in_this_slice",
    "token_validation_schema_created_in_this_slice",
    "token_validator_created_in_this_slice",
    "auth_middleware_created_in_this_slice",
    "membership_adapter_created_in_this_slice",
    "negative_auth_smoke_created_in_this_slice",
    "production_api_consumer_created_in_this_slice",
    "repository_mode_enabled_in_this_slice",
    "database_query_created_in_this_slice",
    "executor_created_in_this_slice",
}
EXPECTED_CANDIDATES = {
    "token_validation_schema",
    "auth_middleware",
    "membership_adapter",
    "negative_auth_smoke",
    "production_api_consumer_binding",
}
EXPECTED_BLOCKED_CONDITIONS = {
    "reviewed_issuer_evidence_missing",
    "jwks_pin_policy_missing",
    "token_validation_schema_missing",
    "auth_middleware_not_created",
    "membership_adapter_not_created",
    "negative_auth_smoke_missing",
    "membership_data_source_not_defined",
    "production_api_consumer_gate_blocked",
    "repository_mode_disabled",
}
EXPECTED_REQUIREMENTS = {
    "reviewed issuer discovery evidence",
    "JWKS pin and refresh policy",
    "token validation schema and sanitized failure envelope",
    "auth middleware route ownership and dev fake auth isolation",
    "membership adapter data source ownership and cache expiry policy",
    "negative auth smoke matrix",
    "production API consumer gate",
}
EXPECTED_FAILURE_CODES = {
    "radish_oidc_token_membership_entry_review_readiness_missing",
    "radish_oidc_token_membership_entry_review_issuer_evidence_missing",
    "radish_oidc_token_membership_entry_review_token_schema_missing",
    "radish_oidc_token_membership_entry_review_membership_source_missing",
    "radish_oidc_token_membership_entry_review_runtime_artifact_forbidden",
}
EXPECTED_ALLOWED_DIAGNOSTICS = {
    "entry_decision",
    "candidate_id",
    "candidate_status",
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
    "token_validation_call_count=0",
    "membership_query_count=0",
    "database_write_count=0",
    "repository_write_count=0",
    "production_api_call_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
REQUIRED_DOC_REFERENCES = {
    "docs/integrations/README.md": [
        "Radish OIDC Token / Membership Implementation Entry Review v1",
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ],
    "docs/platform/README.md": [
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ],
    "docs/integrations/radish-oidc-token-membership-readiness-v1.md": [
        "radish_oidc_token_membership_implementation_entry_review_defined",
    ],
    "docs/radishmind-current-focus.md": [
        "radish_oidc_token_membership_implementation_entry_review_defined",
        "radish-oidc-token-membership-implementation-entry-review-v1.json",
    ],
    "docs/task-cards/README.md": [
        "Radish OIDC Token / Membership Implementation Entry Review",
        "radish-oidc-token-membership-implementation-entry-review-v1",
    ],
    "scripts/README.md": [
        "check-radish-oidc-token-membership-implementation-entry-review-v1.py",
        "radish-oidc-token-membership-implementation-entry-review-v1.json",
    ],
    "docs/devlogs/2026-W26.md": [
        "radish-oidc-token-membership-implementation-entry-review-v1",
        "radish_oidc_token_membership_implementation_entry_review_defined",
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
    require(fixture.get("kind") == "radish_oidc_token_membership_implementation_entry_review_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "radish-oidc-token-membership-implementation-entry-review-v1", "unexpected id")
    require(
        slice_info.get("status") == "radish_oidc_token_membership_implementation_entry_review_defined",
        "unexpected status",
    )
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


def assert_entry_review(fixture: dict[str, Any]) -> None:
    entry = fixture.get("entry_review") or {}
    require(entry.get("status") == "implementation_entry_review_defined", "entry status drifted")
    require(entry.get("entry_decision") == "blocked_before_runtime_task_card", "entry decision must stay blocked")
    for flag in EXPECTED_FALSE_ENTRY_FLAGS:
        require(entry.get(flag) is False, f"{flag} must remain false")


def assert_candidate_matrix(fixture: dict[str, Any]) -> None:
    candidates = rows_by_id(fixture, "candidate_matrix", "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, f"unexpected candidates: {sorted(candidates)}")
    for candidate_id, item in candidates.items():
        require(item.get("entry_decision") == "blocked_before_task_card", f"{candidate_id} must stay blocked")
        require(item.get("runtime_artifact_allowed_now") is False, f"{candidate_id} runtime artifact must be forbidden")
        require(item.get("current_blockers"), f"{candidate_id} blockers are required")

    missing_blocked = sorted(EXPECTED_BLOCKED_CONDITIONS - set(fixture.get("blocked_conditions") or []))
    require(not missing_blocked, f"missing blocked conditions: {missing_blocked}")
    missing_requirements = sorted(EXPECTED_REQUIREMENTS - set(fixture.get("future_task_card_requirements") or []))
    require(not missing_requirements, f"missing future task card requirements: {missing_requirements}")


def assert_failure_and_diagnostics(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture, "failure_mapping", "code")
    missing_failures = sorted(EXPECTED_FAILURE_CODES - set(failures))
    require(not missing_failures, f"missing failure codes: {missing_failures}")
    for code, item in failures.items():
        require(str(item.get("failure_boundary") or "").strip(), f"{code} failure boundary is required")
        diagnostic = str(item.get("sanitized_diagnostic") or "")
        require(diagnostic.strip(), f"{code} sanitized diagnostic is required")
        lowered = diagnostic.lower()
        for forbidden in ("token", "secret", "cookie", "authorization_header", "raw_claim"):
            if forbidden in {"token", "secret"}:
                continue
            require(forbidden not in lowered, f"{code} diagnostic leaks forbidden term: {forbidden}")

    policy = fixture.get("diagnostic_policy") or {}
    missing_allowed = sorted(EXPECTED_ALLOWED_DIAGNOSTICS - set(policy.get("allowed_fields") or []))
    require(not missing_allowed, f"missing allowed diagnostics: {missing_allowed}")
    missing_forbidden = sorted(EXPECTED_FORBIDDEN_DIAGNOSTICS - set(policy.get("forbidden_fields") or []))
    require(not missing_forbidden, f"missing forbidden diagnostics: {missing_forbidden}")


def assert_artifact_guard(fixture: dict[str, Any]) -> None:
    for item in fixture.get("planned_artifacts") or []:
        path_text = str(item.get("path") or "")
        require(path_text, "planned artifact path is required")
        require(item.get("created_in_this_slice") is False, f"{path_text} must not be created in this slice")
        require(not (REPO_ROOT / path_text).exists(), f"future artifact exists too early: {path_text}")

    counters = set((fixture.get("no_side_effect_policy") or {}).get("side_effect_counters_must_remain") or [])
    missing_counters = sorted(EXPECTED_SIDE_EFFECT_COUNTERS - counters)
    require(not missing_counters, f"missing side effect counters: {missing_counters}")


def assert_docs_and_registration() -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    require(
        'run_python_script("checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py", [])'
        in check_repo,
        "check-repo.py must run Radish OIDC token / membership implementation entry review check",
    )
    for relative_path, literals in REQUIRED_DOC_REFERENCES.items():
        text = (REPO_ROOT / relative_path).read_text(encoding="utf-8")
        missing = [literal for literal in literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies(fixture)
    assert_entry_review(fixture)
    assert_candidate_matrix(fixture)
    assert_failure_and_diagnostics(fixture)
    assert_artifact_guard(fixture)
    assert_docs_and_registration()
    print("Radish OIDC token / membership implementation entry review checks passed.")


if __name__ == "__main__":
    main()
