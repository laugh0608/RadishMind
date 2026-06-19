#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-entry-review-v1.json"
)
TRIGGER_REVIEW_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json"
)
SCHEMA_EVIDENCE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-evidence-v1.json"
)
PRODUCT_SURFACE_RECHECK_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/product-surface-readiness-implementation-trigger-recheck-v1.json"
)
SCHEMA_ARTIFACT_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json"
)
SELECTOR_SMOKE_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json"
)
PRODUCTION_AUTH_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json"
)
ADAPTER_SMOKE_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-adapter-smoke-readiness-v1.json"
)
ADAPTER_PLAN_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-implementation-trigger-review-v1": (
        TRIGGER_REVIEW_PATH,
        "implementation_trigger_review_defined",
    ),
    "control-plane-read-schema-artifact-evidence-v1": (
        SCHEMA_EVIDENCE_PATH,
        "schema_artifact_evidence_defined",
    ),
    "product-surface-readiness-implementation-trigger-recheck-v1": (
        PRODUCT_SURFACE_RECHECK_PATH,
        "product_surface_readiness_trigger_recheck_defined",
    ),
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        SCHEMA_ARTIFACT_READINESS_PATH,
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_READINESS_PATH,
        "store_selector_smoke_readiness_defined",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        PRODUCTION_AUTH_READINESS_PATH,
        "production_auth_readiness_defined",
    ),
    "control-plane-read-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_READINESS_PATH,
        "adapter_smoke_readiness_defined",
    ),
    "control-plane-read-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_PATH,
        "repository_adapter_implementation_plan_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implementation_trigger_satisfied",
    "implementation_entry_opened",
    "implementation_task_card_created",
    "schema_artifact_manifest_ready",
    "schema_artifact_files_created",
    "schema_artifact_implementation_ready",
    "store_selector_implementation_ready",
    "store_selector_smoke_ready",
    "production_auth_implementation_ready",
    "production_auth_ready",
    "adapter_smoke_ready",
    "repository_interface_ready",
    "repository_adapter_ready",
    "repository_implementation_ready",
    "database_schema_ready",
    "database_query_ready",
    "database_migration_ready",
    "migration_files_created",
    "migration_manifest_created",
    "migration_runner_ready",
    "store_selector_implemented",
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
EXPECTED_CANDIDATE_IDS = {
    "schema_artifact_manifest_implementation",
    "store_selector_smoke_implementation",
    "production_auth_implementation",
    "adapter_smoke_execution",
}
EXPECTED_CANDIDATE_SOURCES = {
    "schema_artifact_manifest_implementation": "control-plane-read-schema-artifact-manifest-readiness-v1",
    "store_selector_smoke_implementation": "control-plane-read-store-selector-smoke-readiness-v1",
    "production_auth_implementation": "control-plane-read-production-auth-readiness-v1",
    "adapter_smoke_execution": "control-plane-read-adapter-smoke-readiness-v1",
}
EXPECTED_GATE_IDS = {
    "implementation_trigger_review_consumed",
    "schema_artifact_evidence_consumed",
    "product_surface_recheck_consumed",
    "schema_artifact_manifest_candidate_blocked",
    "store_selector_smoke_candidate_blocked",
    "production_auth_candidate_blocked",
    "adapter_smoke_candidate_blocked",
    "single_track_selection_policy_defined",
    "no_runtime_artifacts_leaked",
    "no_same_level_ui_panel_added",
}
EXPECTED_EVIDENCE_KINDS = {
    "ddl_review_evidence",
    "rollback_fixture_evidence",
    "schema_version_evidence",
    "tenant_index_evidence",
    "read_only_role_evidence",
    "route_schema_artifact_mapping",
}
EXPECTED_FORBIDDEN_TASK_CARDS = {
    "docs/task-cards/control-plane-read-schema-artifact-manifest-implementation-v1-plan.md",
    "docs/task-cards/control-plane-read-store-selector-smoke-implementation-v1-plan.md",
    "docs/task-cards/control-plane-read-production-auth-implementation-v1-plan.md",
    "docs/task-cards/control-plane-read-adapter-smoke-execution-v1-plan.md",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/platform/migrations/control_plane_read/manifest.json",
    "services/platform/migrations/control_plane_read/ddl-review.md",
    "services/platform/migrations/control_plane_read/rollback-fixture.json",
    "services/platform/internal/httpapi/control_plane_read_store_selector.go",
    "contracts/radish-oidc-token-validation.schema.json",
    "services/platform/internal/httpapi/control_plane_read_auth_middleware.go",
    "scripts/checks/fixtures/control-plane-read-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-control-plane-read-adapter-smoke-v1.py",
    "services/platform/internal/httpapi/control_plane_read_repository_interface.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
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
    "control_plane_read_repository_interface.go",
    "control_plane_read_repository_adapter.go",
    "control_plane_read_store_selector.go",
    "control_plane_read_auth_middleware.go",
    "control-plane-read-adapter-smoke-v1.json",
    "check-control-plane-read-adapter-smoke-v1.py",
    "type ControlPlaneReadRepository interface",
    "func NewControlPlaneReadRepository",
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


def trigger_rows() -> dict[str, dict[str, Any]]:
    trigger = load_json(TRIGGER_REVIEW_PATH)
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in trigger.get("implementation_trigger_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_CANDIDATE_IDS, "upstream trigger review candidate ids drifted")
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
    require(
        fixture.get("kind") == "control_plane_read_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-implementation-entry-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "implementation_entry_review_defined",
        "implementation entry review status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_entry_boundary") or {}
    require(
        boundary.get("status") == "implementation_entry_review_defined_no_runtime_change",
        "entry boundary status drifted",
    )
    require(boundary.get("decision") == "implementation_entry_not_opened", "entry decision drifted")
    require(
        boundary.get("current_development_mode") == "targeted_correction_or_wait_for_trigger",
        "development mode drifted",
    )
    require(boundary.get("selected_implementation_track_now") == "none", "selected track must remain none")
    require(
        boundary.get("if_trigger_remains_not_satisfied") == "do_not_open_implementation_task_card",
        "not satisfied branch drifted",
    )
    require(
        boundary.get("if_single_trigger_satisfied") == "open_only_that_candidate_task_card_after_review",
        "single trigger branch drifted",
    )
    require(
        boundary.get("if_multiple_triggers_satisfied")
        == "choose_first_review_order_candidate_and_document_deferral",
        "multiple trigger branch drifted",
    )
    for field in (
        "implementation_task_card_created_in_this_slice",
        "parallel_implementation_tracks_allowed_now",
        "schema_artifact_files_allowed_now",
        "selector_files_allowed_now",
        "production_auth_files_allowed_now",
        "adapter_smoke_files_allowed_now",
        "repository_adapter_files_allowed_now",
        "database_connection_allowed_now",
        "dev_server_started_in_this_slice",
        "write_allowed_now",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_gate_matrix(fixture: dict[str, Any]) -> None:
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get("entry_gate_matrix") or []
        if isinstance(gate, dict)
    }
    require(set(gates) == EXPECTED_GATE_IDS, "entry gate ids drifted")
    for gate_id in (
        "implementation_trigger_review_consumed",
        "schema_artifact_evidence_consumed",
        "product_surface_recheck_consumed",
        "single_track_selection_policy_defined",
    ):
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must remain satisfied")
    blocked_gate_to_candidate = {
        "schema_artifact_manifest_candidate_blocked": "schema_artifact_manifest_implementation",
        "store_selector_smoke_candidate_blocked": "store_selector_smoke_implementation",
        "production_auth_candidate_blocked": "production_auth_implementation",
        "adapter_smoke_candidate_blocked": "adapter_smoke_execution",
    }
    for gate_id, candidate_id in blocked_gate_to_candidate.items():
        require(gates[gate_id].get("status") == "blocked", f"{gate_id} must remain blocked")
        require(gates[gate_id].get("candidate_id") == candidate_id, f"{gate_id} candidate drifted")
    for gate_id in ("no_runtime_artifacts_leaked", "no_same_level_ui_panel_added"):
        require(gates[gate_id].get("status") == "required_now", f"{gate_id} status drifted")
        require(gates[gate_id].get("must_cover"), f"{gate_id} must list coverage")


def assert_entry_candidates(fixture: dict[str, Any]) -> None:
    upstream = trigger_rows()
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in fixture.get("implementation_entry_candidates") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_CANDIDATE_IDS, "implementation entry candidate ids drifted")
    for candidate_id, row in rows.items():
        expected_source = EXPECTED_CANDIDATE_SOURCES[candidate_id]
        source_path, source_status = EXPECTED_DEPENDENCIES[expected_source]
        require(row.get("source") == expected_source, f"{candidate_id} source drifted")
        require(row.get("source_status") == source_status, f"{candidate_id} source status drifted")
        require(load_json(source_path).get("slice", {}).get("status") == source_status, f"{expected_source} drifted")
        require(upstream[candidate_id].get("decision") == "not_satisfied", f"{candidate_id} upstream decision")
        require(row.get("upstream_review_decision") == "not_satisfied", f"{candidate_id} upstream decision drifted")
        require(row.get("entry_decision") == "blocked", f"{candidate_id} entry must remain blocked")
        require(row.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(row.get("future_task_card_allowed_after_trigger") is True, f"{candidate_id} future branch")
        for field in (
            "task_card_allowed_now",
            "implementation_artifacts_allowed_now",
            "direct_runtime_implementation_allowed_now",
        ):
            require(row.get(field) is False, f"{candidate_id} {field} must remain false")
        require(row.get("current_blockers") == upstream[candidate_id].get("current_blockers"), f"{candidate_id} blockers")
        if candidate_id in {"schema_artifact_manifest_implementation", "adapter_smoke_execution"}:
            require(
                row.get("schema_evidence_ref") == "control-plane-read-schema-artifact-evidence-v1",
                f"{candidate_id} schema evidence ref drifted",
            )


def assert_schema_evidence_reconciliation(fixture: dict[str, Any]) -> None:
    reconciliation = fixture.get("schema_evidence_reconciliation") or {}
    schema_evidence = load_json(SCHEMA_EVIDENCE_PATH)
    schema_slice = schema_evidence.get("slice") or {}
    require(
        reconciliation.get("status") == "schema_evidence_defined_not_materialized",
        "schema reconciliation status drifted",
    )
    require(
        reconciliation.get("schema_evidence_status") == schema_slice.get("status"),
        "schema evidence status drifted",
    )
    require(
        reconciliation.get("route_mapping_count") == len(schema_evidence.get("route_schema_artifact_mapping") or []),
        "route mapping count drifted",
    )
    require(set(reconciliation.get("evidence_kinds") or []) == EXPECTED_EVIDENCE_KINDS, "evidence kinds drifted")
    for field in (
        "evidence_materialized_now",
        "schema_artifact_manifest_ready_now",
        "schema_artifact_manifest_task_card_allowed_now",
        "schema_artifact_implementation_trigger_satisfied_now",
    ):
        require(reconciliation.get(field) is False, f"{field} must remain false")

    gates = {
        str(gate.get("id") or ""): gate
        for gate in schema_evidence.get("evidence_gate_matrix") or []
        if isinstance(gate, dict)
    }
    for gate_id in (
        "ddl_review_evidence_defined",
        "rollback_fixture_evidence_defined",
        "schema_version_evidence_defined",
        "tenant_index_evidence_defined",
        "read_only_role_evidence_defined",
        "route_schema_artifact_mapping_defined",
    ):
        require(gates.get(gate_id, {}).get("status") == "defined_not_materialized", f"{gate_id} drifted")
    require(
        gates.get("implementation_trigger_remains_not_satisfied", {}).get("status") == "required_now",
        "schema evidence must keep implementation trigger blocked",
    )


def assert_next_development_policy(fixture: dict[str, Any]) -> None:
    policy = fixture.get("next_development_policy") or {}
    require(policy.get("status") == "targeted_correction_or_wait_for_trigger", "next policy status drifted")
    for field in (
        "no_same_level_read_only_panel_by_default",
        "no_new_governance_slice_without_new_boundary",
        "ordinary_copy_or_layout_uses_existing_gates",
        "real_reading_gap_allows_targeted_correction",
        "schema_evidence_gap_allows_targeted_correction",
        "implementation_requires_satisfied_trigger",
        "implementation_entry_selects_one_track_only",
        "trigger_recheck_required_before_task_card",
    ):
        require(policy.get(field) is True, f"{field} must remain true")


def assert_review_order(fixture: dict[str, Any]) -> None:
    rows = [row for row in fixture.get("entry_review_order") or [] if isinstance(row, dict)]
    require(len(rows) == len(EXPECTED_CANDIDATE_IDS), "review order must cover every candidate")
    ordered = [str(row.get("candidate_id") or "") for row in sorted(rows, key=lambda row: row.get("step"))]
    require(
        ordered
        == [
            "schema_artifact_manifest_implementation",
            "store_selector_smoke_implementation",
            "production_auth_implementation",
            "adapter_smoke_execution",
        ],
        "entry review order drifted",
    )
    for index, row in enumerate(sorted(rows, key=lambda row: row.get("step")), start=1):
        require(row.get("step") == index, "review order step must be contiguous")
        require(str(row.get("future_track") or "").strip(), "future track is required")
        require(str(row.get("reason") or "").strip(), "review order reason is required")


def assert_forbidden_tasks_artifacts_and_sources(fixture: dict[str, Any]) -> None:
    task_cards = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_task_card_matrix") or []
        if isinstance(row, dict)
    }
    require(set(task_cards) == EXPECTED_FORBIDDEN_TASK_CARDS, "forbidden task card paths drifted")
    for relative_path, row in task_cards.items():
        require(row.get("candidate_id") in EXPECTED_CANDIDATE_IDS, f"{relative_path} candidate drifted")
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    artifacts = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact paths drifted")
    for relative_path, row in artifacts.items():
        require(row.get("candidate_id") in EXPECTED_CANDIDATE_IDS, f"{relative_path} candidate drifted")
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")
    require(not (REPO_ROOT / "services/platform/migrations/control_plane_read").exists(), "migration root leaked")
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this entry review slice",
                )


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "missing_trigger_success_allowed",
        "schema_evidence_promoted_to_manifest_allowed",
        "readiness_only_promoted_to_implementation_allowed",
        "blocked_candidate_promoted_to_task_card_allowed",
        "repository_mode_fallback_to_fixture_fake_store_allowed",
        "database_mode_fallback_to_fixture_fake_store_allowed",
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
    previous_checker = "check-control-plane-read-schema-artifact-evidence-v1.py"
    current_checker = "check-control-plane-read-implementation-entry-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run implementation entry review check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "implementation entry review check must run after schema artifact evidence",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_boundary(fixture)
    assert_gate_matrix(fixture)
    assert_entry_candidates(fixture)
    assert_schema_evidence_reconciliation(fixture)
    assert_next_development_policy(fixture)
    assert_review_order(fixture)
    assert_forbidden_tasks_artifacts_and_sources(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_references_and_check_repo(fixture)
    print("control plane read implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
