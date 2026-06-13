#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-evidence-v1.json"
SCHEMA_ARTIFACT_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json"
)
SCHEMA_IMPLEMENTATION_PRECONDITIONS_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/control-plane-read-schema-migration-implementation-preconditions-v1.json"
)
IMPLEMENTATION_TRIGGER_REVIEW_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json"
)
PRODUCT_SURFACE_RECHECK_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/product-surface-readiness-implementation-trigger-recheck-v1.json"
)
ROUTE_CONTRACT_PATH = REPO_ROOT / "scripts/checks/fixtures/control-plane-read-route-contract-v1.json"
RUNNER_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-contract-smoke-runner-implementation-v1.json"
)
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
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        SCHEMA_ARTIFACT_READINESS_PATH,
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-schema-migration-implementation-preconditions-v1": (
        SCHEMA_IMPLEMENTATION_PRECONDITIONS_PATH,
        "schema_migration_implementation_preconditions_defined",
    ),
    "control-plane-read-implementation-trigger-review-v1": (
        IMPLEMENTATION_TRIGGER_REVIEW_PATH,
        "implementation_trigger_review_defined",
    ),
    "product-surface-readiness-implementation-trigger-recheck-v1": (
        PRODUCT_SURFACE_RECHECK_PATH,
        "product_surface_readiness_trigger_recheck_defined",
    ),
    "control-plane-read-route-contract-v1": (
        ROUTE_CONTRACT_PATH,
        "governance_boundary_satisfied",
    ),
    "control-plane-read-repository-contract-smoke-runner-implementation-v1": (
        RUNNER_IMPLEMENTATION_PATH,
        "repository_contract_smoke_runner_implemented",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implementation_trigger_satisfied",
    "schema_artifact_manifest_ready",
    "schema_artifact_files_created",
    "ddl_review_ready",
    "ddl_review_artifact_created",
    "rollback_fixture_ready",
    "rollback_fixture_created",
    "schema_version_smoke_ready",
    "schema_version_smoke_executed",
    "tenant_index_smoke_ready",
    "tenant_index_smoke_executed",
    "read_only_role_smoke_ready",
    "read_only_role_smoke_executed",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_manifest_created",
    "migration_runner_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "repository_migration_ready",
    "store_selector_implemented",
    "formal_read_store_config_ready",
    "radish_oidc_client_ready",
    "auth_middleware_ready",
    "token_validation_ready",
    "production_api_consumer_ready",
    "api_key_lifecycle_ready",
    "quota_enforcement_ready",
    "billing_ready",
    "cost_ledger_ready",
    "workflow_executor_ready",
    "confirmation_flow_ready",
    "business_writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_EVIDENCE_KINDS = {
    "ddl_review_evidence",
    "rollback_fixture_evidence",
    "schema_version_evidence",
    "tenant_index_evidence",
    "read_only_role_evidence",
    "route_schema_artifact_mapping",
}
EXPECTED_GATE_IDS = {
    "schema_artifact_manifest_readiness_consumed",
    "schema_migration_implementation_preconditions_consumed",
    "implementation_trigger_review_consumed",
    "product_surface_recheck_consumed",
    "ddl_review_evidence_defined",
    "rollback_fixture_evidence_defined",
    "schema_version_evidence_defined",
    "tenant_index_evidence_defined",
    "read_only_role_evidence_defined",
    "route_schema_artifact_mapping_defined",
    "implementation_trigger_remains_not_satisfied",
    "no_schema_artifacts_leaked",
}
EXPECTED_TRIGGER_CANDIDATES = {
    "schema_artifact_manifest_implementation",
    "store_selector_smoke_implementation",
    "production_auth_implementation",
    "adapter_smoke_execution",
}
EXPECTED_PLANNED_EVIDENCE_PATHS = {
    "services/platform/migrations/control_plane_read/ddl-review.md",
    "services/platform/migrations/control_plane_read/rollback-fixture.json",
    "services/platform/migrations/control_plane_read/schema-version-smoke.json",
    "services/platform/migrations/control_plane_read/tenant-index-smoke.json",
    "services/platform/migrations/control_plane_read/read-only-role-smoke.json",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/platform/migrations/control_plane_read/manifest.json",
    "services/platform/migrations/control_plane_read/ddl-review.md",
    "services/platform/migrations/control_plane_read/rollback-fixture.json",
    "services/platform/migrations/control_plane_read/schema-version-smoke.json",
    "services/platform/migrations/control_plane_read/tenant-index-smoke.json",
    "services/platform/migrations/control_plane_read/read-only-role-smoke.json",
    "services/platform/internal/httpapi/control_plane_read_store_selector.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
}
EXPECTED_FAILURE_CODES = {
    "database_read_disabled",
    "invalid_read_store_mode",
    "read_store_unavailable",
    "read_store_contract_mismatch",
    "schema_migration_not_applied",
    "schema_version_mismatch",
    "tenant_binding_missing",
    "scope_denied",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "dev_server_start_count=0",
    "repository_write_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "dev_server_start",
    "database_write",
    "api_key_issue",
    "quota_mutation",
    "workflow_execution",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
}
EXPECTED_SOURCE_ABSENT_LITERALS = {
    "ControlPlaneReadSchemaMigrationRunner",
    "control_plane_read_schema_migration_runner.go",
    "database/sql",
    "CREATE TABLE",
    "ALTER TABLE",
    "CREATE INDEX",
    "DROP TABLE",
    "SELECT *",
    "INSERT INTO",
    "UPDATE ",
    "DELETE FROM",
    "services/platform/migrations/control_plane_read",
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_store_selector.go",
    "RADISHMIND_CONTROL_PLANE_READ_STORE",
    "SelectControlPlaneReadStore",
    "ControlPlaneReadStoreSelector",
    "oidc.Provider",
    "github.com/coreos/go-oidc",
    "ValidateToken",
    "VerifyToken",
    "NewRadishOIDCClient",
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


def rows_by_route(fixture: dict[str, Any], key: str) -> dict[str, dict[str, Any]]:
    rows = {
        str(row.get("route_id") or row.get("id") or ""): row
        for row in fixture.get(key) or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_ROUTE_IDS, f"{key} must cover every read route")
    return rows


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
    require(fixture.get("kind") == "control_plane_read_schema_artifact_evidence_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "control-plane-read-schema-artifact-evidence-v1", "unexpected slice id")
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "schema_artifact_evidence_defined",
        "schema artifact evidence status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    schema_boundary = load_json(SCHEMA_ARTIFACT_READINESS_PATH).get("schema_artifact_boundary") or {}
    migration_boundary = load_json(SCHEMA_IMPLEMENTATION_PRECONDITIONS_PATH).get(
        "implementation_preconditions_boundary"
    ) or {}
    product_boundary = load_json(PRODUCT_SURFACE_RECHECK_PATH).get("recheck_boundary") or {}

    require(boundary.get("status") == "schema_artifact_evidence_defined_no_artifact", "boundary status drifted")
    require(
        boundary.get("decision") == "evidence_chain_defined_without_schema_artifact",
        "boundary decision drifted",
    )
    require(boundary.get("current_store_source") == schema_boundary.get("current_store_source"), "store source drifted")
    require(boundary.get("future_migration_root") == schema_boundary.get("future_migration_root"), "root drifted")
    require(boundary.get("future_migration_root") == migration_boundary.get("future_migration_root"), "root drifted")
    require(boundary.get("future_manifest_file") == schema_boundary.get("future_manifest_file"), "manifest drifted")
    require(
        boundary.get("future_schema_version_table") == schema_boundary.get("future_schema_version_table"),
        "schema version table drifted",
    )
    require(product_boundary.get("dev_server_started_by_ai") is False, "product recheck must not start dev server")
    for field in (
        "schema_artifact_manifest_created_in_this_slice",
        "schema_artifact_files_created_in_this_slice",
        "ddl_review_artifact_created_in_this_slice",
        "rollback_fixture_created_in_this_slice",
        "schema_version_smoke_created_in_this_slice",
        "schema_version_smoke_executed_in_this_slice",
        "tenant_index_smoke_created_in_this_slice",
        "tenant_index_smoke_executed_in_this_slice",
        "read_only_role_smoke_created_in_this_slice",
        "read_only_role_smoke_executed_in_this_slice",
        "migration_root_created_in_this_slice",
        "sql_files_created_in_this_slice",
        "migration_runner_created_in_this_slice",
        "repository_adapter_allowed_in_this_slice",
        "store_selector_allowed_in_this_slice",
        "database_query_allowed_in_this_slice",
        "database_migration_allowed_in_this_slice",
        "oidc_validation_allowed_in_this_slice",
        "production_api_consumer_allowed_in_this_slice",
        "write_allowed_in_this_slice",
        "dev_server_started_in_this_slice",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_evidence_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("schema_artifact_evidence_contract") or {}
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    require(contract.get("status") == "evidence_contract_defined_not_materialized", "contract status drifted")
    require(contract.get("required_before_durable_read_store") is True, "durable read store gate required")
    require(
        contract.get("required_before_schema_artifact_manifest_ready") is True,
        "schema artifact manifest ready gate required",
    )
    require(contract.get("future_manifest_path") == boundary.get("future_manifest_file"), "manifest path drifted")
    require(set(contract.get("evidence_kinds") or []) == EXPECTED_EVIDENCE_KINDS, "evidence kinds drifted")

    artifacts = {
        str(item.get("path") or ""): item
        for item in contract.get("planned_evidence_artifacts") or []
        if isinstance(item, dict)
    }
    require(set(artifacts) == EXPECTED_PLANNED_EVIDENCE_PATHS, "planned evidence paths drifted")
    for relative_path, artifact in artifacts.items():
        require(artifact.get("evidence_kind") in EXPECTED_EVIDENCE_KINDS, f"{relative_path} kind drifted")
        require(artifact.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("evidence_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "evidence gate ids drifted")
    for gate_id in (
        "schema_artifact_manifest_readiness_consumed",
        "schema_migration_implementation_preconditions_consumed",
        "implementation_trigger_review_consumed",
        "product_surface_recheck_consumed",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
        require(gates[gate_id].get("evidence"), f"{gate_id} must include evidence")
    for gate_id in (
        "ddl_review_evidence_defined",
        "rollback_fixture_evidence_defined",
        "schema_version_evidence_defined",
        "tenant_index_evidence_defined",
        "read_only_role_evidence_defined",
        "route_schema_artifact_mapping_defined",
    ):
        require(gates[gate_id].get("status") == "defined_not_materialized", f"{gate_id} status drifted")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")
    for gate_id in ("implementation_trigger_remains_not_satisfied", "no_schema_artifacts_leaked"):
        require(gates[gate_id].get("status") == "required_now", f"{gate_id} status drifted")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")


def assert_evidence_sections(fixture: dict[str, Any]) -> None:
    ddl = fixture.get("ddl_review_evidence") or {}
    require(ddl.get("status") == "defined_not_materialized", "DDL evidence status drifted")
    require(ddl.get("reviewer_requirement") == "human_review_required", "DDL review must require human review")
    require(ddl.get("manual_apply_required") is True, "manual apply must be required")
    require(ddl.get("startup_auto_migration_allowed") is False, "startup auto migration must stay blocked")
    require(ddl.get("destructive_change_without_review_allowed") is False, "destructive DDL review gate drifted")
    require(ddl.get("artifact_created_now") is False, "DDL review artifact must not be created")

    rollback = fixture.get("rollback_fixture_evidence") or {}
    require(rollback.get("status") == "defined_not_materialized", "rollback evidence status drifted")
    require(rollback.get("backup_required_before_apply") is True, "backup requirement drifted")
    require(rollback.get("failed_migration_recovery_required") is True, "recovery requirement drifted")
    require(rollback.get("migration_lock_release_required") is True, "lock release requirement drifted")
    require(rollback.get("artifact_created_now") is False, "rollback fixture must not be created")

    version = fixture.get("schema_version_evidence") or {}
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    require(version.get("status") == "defined_not_materialized", "schema version evidence status drifted")
    require(version.get("expected_schema_version") == boundary.get("future_schema_version"), "schema version drifted")
    require(
        version.get("future_schema_version_table") == boundary.get("future_schema_version_table"),
        "schema version table drifted",
    )
    require(version.get("missing_version_failure_code") == "schema_migration_not_applied", "missing version code")
    require(version.get("mismatched_version_failure_code") == "schema_version_mismatch", "mismatched version code")
    require(version.get("schema_version_smoke_executed_now") is False, "schema version smoke must not run")

    tenant_index = fixture.get("tenant_index_evidence") or {}
    require(tenant_index.get("tenant_predicate_key") == "tenant_ref", "tenant predicate drifted")
    require(tenant_index.get("leading_index_required") is True, "tenant leading index requirement drifted")
    require(tenant_index.get("cross_tenant_read_denial_required") is True, "cross tenant denial required")
    require(
        tenant_index.get("invalid_tenant_predicate_fail_closed_required") is True,
        "invalid tenant predicate must fail closed",
    )
    require(tenant_index.get("tenant_index_smoke_executed_now") is False, "tenant index smoke must not run")

    role = fixture.get("read_only_role_evidence") or {}
    require(role.get("credential_policy") == "secret_reference_only", "runtime role credential policy drifted")
    require(role.get("allowed_privileges") == ["read sanitized projections"], "allowed privileges drifted")
    forbidden = set(role.get("forbidden_privileges") or [])
    require(
        {"insert", "update", "delete", "ddl", "migration_apply", "secret_value_read"}.issubset(forbidden),
        "read-only role forbidden privileges drifted",
    )
    require(role.get("read_only_role_smoke_executed_now") is False, "read-only role smoke must not run")


def assert_route_schema_artifact_mapping(fixture: dict[str, Any]) -> None:
    rows = rows_by_route(fixture, "route_schema_artifact_mapping")
    schema_rows = rows_by_route(load_json(SCHEMA_ARTIFACT_READINESS_PATH), "route_schema_artifact_matrix")
    route_rows = rows_by_route(load_json(ROUTE_CONTRACT_PATH), "route_contracts")
    runner_rows = rows_by_route(load_json(RUNNER_IMPLEMENTATION_PATH), "route_runner_matrix")
    boundary = fixture.get("schema_artifact_evidence_boundary") or {}
    evidence_ids = {
        "ddl_review_evidence_ref": "ddl-review-evidence",
        "rollback_fixture_evidence_ref": "rollback-fixture-evidence",
        "schema_version_evidence_ref": "schema-version-evidence",
        "tenant_index_evidence_ref": "tenant-index-evidence",
        "read_only_role_evidence_ref": "read-only-role-evidence",
    }

    for route_id, row in rows.items():
        require(row.get("operation") == schema_rows[route_id].get("operation"), f"{route_id} operation drifted")
        require(row.get("operation") == runner_rows[route_id].get("operation"), f"{route_id} runner operation drifted")
        require(row.get("logical_entity") == schema_rows[route_id].get("logical_entity"), f"{route_id} entity drifted")
        require(route_rows[route_id].get("tenant_scoped") is True, f"{route_id} must remain tenant scoped")
        require(row.get("tenant_predicate") == "tenant_ref", f"{route_id} tenant predicate drifted")
        require(
            row.get("future_schema_artifact_id") == boundary.get("future_schema_artifact_id"),
            f"{route_id} schema artifact id drifted",
        )
        require(str(row.get("future_projection_name") or "").startswith("control_plane_read_"), f"{route_id} projection")
        for field, expected in evidence_ids.items():
            require(row.get(field) == expected, f"{route_id} {field} drifted")
        for field in (
            "schema_artifact_ready",
            "runtime_mapping_implemented",
            "sql_allowed_now",
            "repository_adapter_allowed_now",
            "fake_fallback_allowed",
            "side_effect_allowed",
        ):
            require(row.get(field) is False, f"{route_id} {field} must remain false")


def assert_trigger_review_still_blocked() -> None:
    trigger = load_json(IMPLEMENTATION_TRIGGER_REVIEW_PATH)
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in trigger.get("implementation_trigger_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_TRIGGER_CANDIDATES, "trigger candidate ids drifted")
    for candidate_id, row in rows.items():
        require(row.get("decision") == "not_satisfied", f"{candidate_id} must remain not_satisfied")
        require(row.get("implementation_task_card_allowed_now") is False, f"{candidate_id} task card gate drifted")
        require(row.get("implementation_artifacts_allowed_now") is False, f"{candidate_id} artifact gate drifted")
        require(row.get("direct_runtime_implementation_allowed_now") is False, f"{candidate_id} runtime gate drifted")


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    failures = set(fixture.get("failure_mapping") or [])
    require(EXPECTED_FAILURE_CODES.issubset(failures), "failure mapping missing required codes")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "missing_ddl_review_success_allowed",
        "missing_rollback_fixture_success_allowed",
        "missing_schema_version_success_allowed",
        "missing_tenant_index_success_allowed",
        "missing_read_only_role_success_allowed",
        "schema_evidence_promoted_to_schema_artifact_allowed",
        "database_mode_fallback_to_fixture_fake_store_allowed",
        "repository_mode_fallback_to_fixture_fake_store_allowed",
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


def assert_forbidden_artifacts_and_sources(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifacts drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    migration_root = REPO_ROOT / "services/platform/migrations/control_plane_read"
    require(not migration_root.exists(), "migration root must not be created")
    sql_files = list((REPO_ROOT / "services/platform").rglob("*.sql"))
    require(not sql_files, f"SQL files must not be introduced: {sql_files}")

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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this evidence slice",
                )


def assert_references_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-product-surface-readiness-implementation-trigger-recheck-v1.py"
    current_checker = "check-control-plane-read-schema-artifact-evidence-v1.py"
    require(current_checker in check_repo, "check-repo.py must run schema artifact evidence check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "schema artifact evidence check must run after product surface recheck",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_evidence_contract(fixture)
    assert_gate_matrix(fixture)
    assert_evidence_sections(fixture)
    assert_route_schema_artifact_mapping(fixture)
    assert_trigger_review_still_blocked()
    assert_failure_and_safety_policies(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_references_and_check_repo(fixture)
    print("control plane read schema artifact evidence v1 checks passed.")


if __name__ == "__main__":
    main()
