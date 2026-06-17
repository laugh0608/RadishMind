#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json"
)
RADISH_OIDC_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radish-oidc-client-preconditions.json"
AUTH_CONTEXT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-auth-context-preconditions-v1.json"
)
SELECTOR_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json"
)
SCHEMA_MATERIALIZATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json"
)
ADAPTER_SMOKE_READINESS_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "radish-oidc-client-preconditions": (RADISH_OIDC_FIXTURE_PATH, "governance_boundary_satisfied"),
    "workflow-saved-draft-auth-context-preconditions-v1": (
        AUTH_CONTEXT_FIXTURE_PATH,
        "draft_auth_context_preconditions_defined",
    ),
    "workflow-saved-draft-store-selector-smoke-v1": (
        SELECTOR_IMPLEMENTATION_FIXTURE_PATH,
        "draft_store_selector_smoke_implemented",
    ),
    "workflow-saved-draft-schema-artifact-materialization-v1": (
        SCHEMA_MATERIALIZATION_FIXTURE_PATH,
        "draft_schema_artifact_materialized_static",
    ),
    "workflow-saved-draft-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "draft_repository_adapter_implementation_plan_defined",
    ),
    "workflow-saved-draft-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_FIXTURE_PATH,
        "draft_adapter_smoke_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "production_auth_ready",
    "radish_oidc_ready",
    "oidc_client_implemented",
    "issuer_discovery_runtime_ready",
    "token_validation_ready",
    "auth_middleware_ready",
    "workspace_membership_adapter_ready",
    "repository_interface_implemented",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "durable_store_implemented",
    "durable_persistence_ready",
    "database_schema_ready",
    "database_connection_ready",
    "database_migration_ready",
    "migration_runner_implemented",
    "adapter_smoke_ready",
    "adapter_smoke_executed",
    "production_api_consumer_ready",
    "workflow_publish_ready",
    "workflow_executor_ready",
    "confirmation_decision_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "production_ready",
}
EXPECTED_PLANNED_ARTIFACTS = {
    "contracts/radish-oidc-token-validation.schema.json",
    "services/platform/internal/httpapi/workflow_saved_draft_auth_middleware.go",
    "services/platform/internal/httpapi/workflow_saved_draft_token_validation.go",
    "services/platform/internal/httpapi/workflow_saved_draft_membership_adapter.go",
    "services/platform/internal/httpapi/workflow_saved_draft_auth_middleware_test.go",
    "scripts/checks/fixtures/workflow-saved-draft-production-auth-v1.json",
    "scripts/checks/control_plane/check-workflow-saved-draft-production-auth-v1.py",
}
EXPECTED_GATE_IDS = {
    "radish_oidc_preconditions_consumed",
    "saved_draft_auth_context_consumed",
    "selector_implementation_consumed",
    "schema_artifact_materialization_consumed",
    "repository_adapter_plan_consumed",
    "adapter_smoke_readiness_consumed",
    "issuer_discovery_evidence_contract_defined",
    "token_validation_contract_preconditions_defined",
    "claim_mapping_contract_defined",
    "tenant_workspace_binding_policy_defined",
    "scope_projection_matrix_defined",
    "production_auth_failure_mapping_defined",
    "downstream_adapter_smoke_readiness_assessed",
    "downstream_repository_adapter_entry_prerequisite_assessed",
    "production_auth_smoke_gate",
    "auth_middleware_implementation_gate",
    "membership_adapter_implementation_gate",
    "production_api_consumer_gate",
    "no_production_auth_artifacts_leaked",
}
SATISFIED_GATE_IDS = {
    "radish_oidc_preconditions_consumed",
    "saved_draft_auth_context_consumed",
    "selector_implementation_consumed",
    "schema_artifact_materialization_consumed",
    "repository_adapter_plan_consumed",
    "adapter_smoke_readiness_consumed",
    "issuer_discovery_evidence_contract_defined",
    "token_validation_contract_preconditions_defined",
    "claim_mapping_contract_defined",
    "tenant_workspace_binding_policy_defined",
    "scope_projection_matrix_defined",
    "production_auth_failure_mapping_defined",
    "downstream_adapter_smoke_readiness_assessed",
    "downstream_repository_adapter_entry_prerequisite_assessed",
}
NOT_SATISFIED_GATE_IDS = {
    "production_auth_smoke_gate",
    "auth_middleware_implementation_gate",
    "membership_adapter_implementation_gate",
    "production_api_consumer_gate",
}
EXPECTED_OPERATIONS = {
    "SaveWorkflowDraftRecord",
    "ReadWorkflowDraftRecord",
    "ListWorkflowDraftRecords",
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
EXPECTED_BINDING_FIELDS = {
    "tenant_ref",
    "workspace_id",
    "application_id",
    "actor_subject_ref",
    "owner_subject_ref",
    "scope_grants",
}
EXPECTED_FAILURE_CODES = {
    "draft_identity_context_missing",
    "draft_tenant_binding_missing",
    "draft_workspace_membership_denied",
    "draft_application_scope_denied",
    "draft_owner_scope_denied",
    "draft_scope_grant_missing",
    "draft_auth_context_contract_mismatch",
    "draft_audit_context_missing",
    "draft_scope_denied",
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
    "tenant_binding_missing",
    "tenant_binding_denied",
    "permission_claim_denied",
    "malformed_authorization_header",
    "logout_failed",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "issuer_network_call_count=0",
    "token_validation_call_count=0",
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
    "auth_session_mutation",
    "database_write",
    "repository_write",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "contracts/radish-oidc-token-validation.schema.json",
    "workflow_saved_draft_auth_middleware.go",
    "workflow_saved_draft_token_validation.go",
    "workflow_saved_draft_membership_adapter.go",
    "workflow-saved-draft-production-auth-v1.json",
    "check-workflow-saved-draft-production-auth-v1.py",
    "oidc.Provider",
    "github.com/coreos/go-oidc",
    "ValidateWorkflowDraftToken",
    "WorkflowSavedDraftAuthMiddleware",
    "WorkflowSavedDraftTokenValidator",
    "WorkflowSavedDraftMembershipAdapter",
    "NewSavedWorkflowDraftRepositoryAdapter",
    "type SavedWorkflowDraftRepository interface",
    "database/sql",
    "RADISHMIND_WORKFLOW_SAVED_DRAFT_OIDC",
    "RADISHMIND_WORKFLOW_SAVED_DRAFT_PRODUCTION_AUTH",
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


def rows_by_operation(fixture: dict[str, Any], key: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get("operation") or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_OPERATIONS, f"{key} must cover saved draft save/read/list")
    return rows


def oidc_preconditions_by_id() -> dict[str, dict[str, Any]]:
    oidc = load_json(RADISH_OIDC_FIXTURE_PATH)
    preconditions = {
        str(item.get("id") or ""): item
        for item in oidc.get("preconditions") or []
        if isinstance(item, dict)
    }
    for required in ("issuer-discovery", "claim-mapping", "tenant-binding", "failure-taxonomy"):
        require(required in preconditions, f"Radish OIDC precondition missing: {required}")
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
        fixture.get("kind") == "workflow_saved_draft_production_auth_readiness_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "workflow-saved-draft-production-auth-readiness-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Workflow / Agent Runtime", "unexpected track")
    require(
        slice_info.get("status") == "draft_production_auth_readiness_defined",
        "production auth readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_production_auth_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("production_auth_boundary") or {}
    auth_context = load_json(AUTH_CONTEXT_FIXTURE_PATH).get("auth_context_boundary") or {}
    require(
        boundary.get("status") == "production_auth_readiness_defined_not_implemented",
        "production auth boundary status drifted",
    )
    require(
        boundary.get("current_auth_source") == auth_context.get("current_auth_source"),
        "current auth source must match auth context preconditions",
    )
    require(
        boundary.get("future_auth_source") == auth_context.get("future_auth_source"),
        "future auth source must match auth context preconditions",
    )
    for field in (
        "issuer_network_call_allowed_in_this_slice",
        "oidc_client_created_in_this_slice",
        "token_validation_allowed_in_this_slice",
        "auth_middleware_allowed_in_this_slice",
        "membership_adapter_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "repository_interface_allowed_in_this_slice",
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
        require(gates[gate_id].get("must_cover"), f"{gate_id} must define coverage")
    require(
        gates["no_production_auth_artifacts_leaked"].get("status") == "required_now",
        "artifact leak gate status drifted",
    )


def assert_issuer_and_token_contracts(fixture: dict[str, Any]) -> None:
    oidc_preconditions = oidc_preconditions_by_id()
    issuer = fixture.get("issuer_discovery_evidence_contract") or {}
    require(issuer.get("status") == "required_before_production_auth", "issuer evidence status drifted")
    require(issuer.get("network_call_allowed_now") is False, "issuer network calls must remain false")
    require(issuer.get("committed_raw_discovery_document_allowed") is False, "raw discovery commit must remain false")
    require(issuer.get("committed_raw_jwks_allowed") is False, "raw JWKS commit must remain false")
    require(
        EXPECTED_ISSUER_FIELDS.issubset(set(issuer.get("required_fields") or [])),
        "issuer evidence missing required fields",
    )
    upstream_issuer_fields = set(oidc_preconditions["issuer-discovery"].get("required_fields") or [])
    require(
        {"issuer_url", "discovery_document_url", "jwks_uri"}.issubset(upstream_issuer_fields),
        "upstream issuer discovery fields drifted",
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


def assert_claim_mapping_and_binding(fixture: dict[str, Any]) -> None:
    upstream_required_claims = set(oidc_preconditions_by_id()["claim-mapping"].get("required_claims") or [])
    require(
        (EXPECTED_REQUIRED_CLAIMS | EXPECTED_OPTIONAL_CLAIMS).issubset(upstream_required_claims),
        "upstream claim mapping drifted",
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
        if claim != "tenant_id":
            require(row.get("allowed_in_response") is False, f"{claim} must not be emitted in response")
    for claim in EXPECTED_OPTIONAL_CLAIMS:
        row = claim_rows[claim]
        require(row.get("required") is False, f"{claim} must remain optional")
        require(row.get("allowed_in_response") is False, f"{claim} must not be emitted in response")
    require(claim_rows["tenant_id"].get("allowed_in_response") is True, "tenant ref may remain in envelope only")
    require(
        claim_rows["permissions"].get("failure_code_when_missing") == "draft_scope_grant_missing",
        "permissions failure mapping drifted",
    )

    policy = fixture.get("tenant_workspace_binding_policy") or {}
    require(
        policy.get("status") == "required_before_repository_adapter_or_adapter_smoke",
        "tenant/workspace binding status drifted",
    )
    require(policy.get("tenant_source") == "trusted auth context claim", "tenant source drifted")
    for field in (
        "query_string_tenant_override_allowed",
        "query_string_workspace_override_allowed",
        "query_string_application_override_allowed",
        "draft_payload_scope_override_allowed",
        "fallback_to_default_tenant_allowed",
        "fallback_to_local_admin_allowed",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    expected_failures = {
        "missing_tenant_failure_code": "draft_tenant_binding_missing",
        "workspace_membership_failure_code": "draft_workspace_membership_denied",
        "application_scope_failure_code": "draft_application_scope_denied",
        "owner_scope_failure_code": "draft_owner_scope_denied",
    }
    for field, expected in expected_failures.items():
        require(policy.get(field) == expected, f"{field} drifted")


def assert_operation_scope_projection(fixture: dict[str, Any]) -> None:
    rows = rows_by_operation(fixture, "operation_scope_projection_matrix")
    auth_rows = rows_by_operation(load_json(AUTH_CONTEXT_FIXTURE_PATH), "scope_grant_matrix")
    adapter_rows = rows_by_operation(load_json(ADAPTER_PLAN_FIXTURE_PATH), "operation_adapter_matrix")
    for operation, row in rows.items():
        require(row.get("required_scope") == auth_rows[operation].get("required_scope"), f"{operation} auth scope drifted")
        require(
            row.get("required_scope") == adapter_rows[operation].get("required_scope"),
            f"{operation} adapter scope drifted",
        )
        require(
            row.get("scope_source") == "permissions claim projected into saved draft actor context",
            f"{operation} scope source drifted",
        )
        require(
            EXPECTED_BINDING_FIELDS.issubset(set(row.get("binding_fields_required") or [])),
            f"{operation} binding fields missing",
        )
        require(row.get("missing_identity_failure_code") == "draft_identity_context_missing", f"{operation} identity failure drifted")
        require(row.get("tenant_binding_failure_code") == "draft_tenant_binding_missing", f"{operation} tenant failure drifted")
        require(
            row.get("workspace_membership_failure_code") == "draft_workspace_membership_denied",
            f"{operation} workspace membership failure drifted",
        )
        require(row.get("scope_denied_failure_code") == "draft_scope_grant_missing", f"{operation} scope failure drifted")
        require(row.get("ready_for_downstream_adapter_smoke") is True, f"{operation} downstream readiness must be true")
        for field in (
            "adapter_smoke_execution_allowed_now",
            "repository_adapter_allowed_now",
            "fake_auth_header_allowed_in_production",
            "fallback_to_test_auth_allowed",
            "fallback_to_memory_dev_store_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{operation} {field} must remain false")


def assert_downstream_review(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("id") or ""): row
        for row in fixture.get("downstream_readiness_review") or []
        if isinstance(row, dict)
    }
    expected = {
        "adapter_smoke_readiness_review": (
            "workflow-saved-draft-adapter-smoke-readiness-v1",
            "draft_adapter_smoke_readiness_defined",
            "adapter_smoke_execution_allowed_now",
        ),
        "repository_adapter_plan_review": (
            "workflow-saved-draft-repository-adapter-implementation-plan-v1",
            "draft_repository_adapter_implementation_plan_defined",
            "repository_adapter_implementation_allowed_now",
        ),
    }
    require(set(rows) == set(expected), "downstream readiness review rows drifted")
    for row_id, (source, status, blocked_field) in expected.items():
        row = rows[row_id]
        require(row.get("source") == source, f"{row_id} source drifted")
        require(row.get("source_status") == status, f"{row_id} source status drifted")
        require(row.get("production_auth_readiness_defined_now") is True, f"{row_id} readiness must be defined")
        require(row.get(blocked_field) is False, f"{row_id} {blocked_field} must remain false")
        require(str(row.get("reason") or "").strip(), f"{row_id} must explain readiness conclusion")


def assert_failure_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "production auth failure mapping missing codes")
    oidc_failures = set(oidc_preconditions_by_id()["failure-taxonomy"].get("required_failures") or [])
    require(oidc_failures.issubset(failures), "production auth readiness must include OIDC failures")
    auth_failures = set((load_json(AUTH_CONTEXT_FIXTURE_PATH).get("failure_policy") or {}).get("required_failure_codes") or [])
    require(auth_failures.issubset(failures), "production auth readiness must include saved draft auth failures")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "dev_fake_auth_header_allowed_in_production",
        "missing_bearer_token_fallback_to_fake_auth_allowed",
        "invalid_token_fallback_to_fake_auth_allowed",
        "scope_denied_fallback_to_admin_allowed",
        "tenant_binding_failure_fallback_allowed",
        "workspace_membership_failure_fallback_allowed",
        "repository_mode_fallback_to_memory_dev_store_allowed",
        "repository_mode_fallback_to_sample_allowed",
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


def assert_implementation_artifact_guard(fixture: dict[str, Any]) -> None:
    guard = fixture.get("implementation_artifact_guard") or {}
    require(
        guard.get("status") == "forbid_production_auth_implementation_artifacts",
        "artifact guard status drifted",
    )
    for relative_path in guard.get("future_files_must_not_exist") or []:
        require(not (REPO_ROOT / str(relative_path)).exists(), f"future auth artifact exists early: {relative_path}")
    configured_literals = set(guard.get("future_literals_must_not_appear_in_source") or [])
    require(EXPECTED_SOURCE_ABSENT_LITERALS.issubset(configured_literals), "source absent literals drifted")
    for source_path in guard.get("source_files_to_scan") or []:
        path = REPO_ROOT / str(source_path)
        require(path.exists(), f"source file missing: {source_path}")
        text = path.read_text(encoding="utf-8")
        for literal in configured_literals:
            require(literal not in text, f"{source_path} contains future auth literal: {literal}")


def assert_required_docs_testing_and_check_repo(fixture: dict[str, Any]) -> None:
    references = fixture.get("required_doc_references") or []
    require(references, "required doc references must be declared")
    for reference in references:
        path = str(reference.get("path") or "")
        require(path, "doc reference path missing")
        content = read(path)
        for needle in reference.get("must_contain") or []:
            require(str(needle) in content, f"{path} missing required text: {needle}")

    strategy = fixture.get("testing_strategy") or {}
    require(strategy.get("status") == "static_checker_only", "testing strategy status drifted")
    required_checks = set(strategy.get("required_checks") or [])
    for expected_check in (
        "workflow_saved_draft_production_auth_readiness_checker",
        "workflow_saved_draft_auth_context_preconditions_checker",
        "workflow_saved_draft_adapter_smoke_readiness_checker",
        "workflow_saved_draft_repository_adapter_implementation_plan_checker",
        "workflow_saved_draft_schema_artifact_materialization_checker",
        "./scripts/check-repo.sh --fast",
    ):
        require(expected_check in required_checks, f"missing required check: {expected_check}")
    for field in (
        "does_not_start_service",
        "does_not_call_oidc",
        "does_not_validate_token",
        "does_not_create_auth_middleware",
        "does_not_create_membership_adapter",
        "does_not_create_repository_interface",
        "does_not_create_repository_adapter",
        "does_not_create_sql_or_migration",
        "does_not_connect_database",
    ):
        require(strategy.get(field) is True, f"{field} must remain true")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py"
    current_checker = "checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py"
    next_checker = "checks/control_plane/check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    require(current_checker in check_repo, "check-repo.py must run saved draft production auth readiness check")
    require(previous_checker in check_repo, "schema artifact materialization checker missing from check-repo.py")
    require(next_checker in check_repo, "product surface recheck missing from check-repo.py")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker) < check_repo.index(next_checker),
        "production auth readiness check must run after schema materialization and before product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_production_auth_boundary(fixture)
    assert_planned_artifacts(fixture)
    assert_readiness_gates(fixture)
    assert_issuer_and_token_contracts(fixture)
    assert_claim_mapping_and_binding(fixture)
    assert_operation_scope_projection(fixture)
    assert_downstream_review(fixture)
    assert_failure_fallback_and_side_effects(fixture)
    assert_implementation_artifact_guard(fixture)
    assert_required_docs_testing_and_check_repo(fixture)
    print("workflow saved draft production auth readiness v1 checks passed.")


if __name__ == "__main__":
    main()
