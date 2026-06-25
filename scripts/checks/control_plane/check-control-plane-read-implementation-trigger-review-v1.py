#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json"
)
ADAPTER_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-adapter-smoke-readiness-v1.json"
)
SCHEMA_ARTIFACT_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json"
)
SELECTOR_SMOKE_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json"
)
PRODUCTION_AUTH_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json"
)
ADAPTER_PLAN_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/control-plane-read-repository-adapter-implementation-plan-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "control-plane-read-adapter-smoke-readiness-v1": (
        ADAPTER_SMOKE_FIXTURE_PATH,
        "adapter_smoke_readiness_defined",
    ),
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        SCHEMA_ARTIFACT_FIXTURE_PATH,
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-store-selector-smoke-readiness-v1": (
        SELECTOR_SMOKE_FIXTURE_PATH,
        "store_selector_smoke_readiness_defined",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        PRODUCTION_AUTH_FIXTURE_PATH,
        "production_auth_readiness_defined",
    ),
    "control-plane-read-repository-adapter-implementation-plan-v1": (
        ADAPTER_PLAN_FIXTURE_PATH,
        "repository_adapter_implementation_plan_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implementation_trigger_satisfied",
    "schema_artifact_implementation_ready",
    "schema_artifact_manifest_ready",
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
EXPECTED_BLOCKERS = {
    "schema_artifact_manifest_implementation": {
        "ddl_review_evidence_gate",
        "rollback_fixture_evidence_gate",
        "schema_version_smoke_evidence_gate",
        "tenant_index_smoke_evidence_gate",
        "read_only_role_smoke_evidence_gate",
        "durable_adapter_smoke_dependency_gate",
    },
    "store_selector_smoke_implementation": {
        "selector_config_entry_gate",
        "selector_code_gate",
        "fixture_default_smoke_gate",
        "reserved_mode_fail_closed_smoke_gate",
        "unknown_mode_fail_closed_smoke_gate",
        "durable_mode_enablement_gate",
    },
    "production_auth_implementation": {
        "production_auth_smoke_gate",
        "auth_middleware_implementation_gate",
        "production_api_consumer_gate",
    },
    "adapter_smoke_execution": {
        "schema_artifact_manifest_gate",
        "store_selector_smoke_gate",
        "production_auth_gate",
        "repository_adapter_implementation_gate",
        "adapter_smoke_execution_gate",
        "production_api_consumer_gate",
    },
}
UPSTREAM_GATE_KEYS = {
    "schema_artifact_manifest_implementation": (SCHEMA_ARTIFACT_FIXTURE_PATH, "readiness_gate_matrix"),
    "store_selector_smoke_implementation": (SELECTOR_SMOKE_FIXTURE_PATH, "smoke_readiness_gate_matrix"),
    "production_auth_implementation": (PRODUCTION_AUTH_FIXTURE_PATH, "readiness_gate_matrix"),
    "adapter_smoke_execution": (ADAPTER_SMOKE_FIXTURE_PATH, "readiness_gate_matrix"),
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/platform/migrations/control_plane_read/manifest.json",
    "services/platform/internal/httpapi/control_plane_read_store_selector.go",
    "services/platform/internal/httpapi/control_plane_read_auth_middleware.go",
    "scripts/checks/fixtures/control-plane-read-adapter-smoke-v1.json",
    "scripts/checks/control_plane/check-control-plane-read-adapter-smoke-v1.py",
    "services/platform/internal/httpapi/control_plane_read_repository_interface.go",
    "services/platform/internal/httpapi/control_plane_read_repository_adapter.go",
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


def gates_by_id(path: Path, key: str) -> dict[str, dict[str, Any]]:
    fixture = load_json(path)
    gates = {
        str(gate.get("id") or ""): gate
        for gate in fixture.get(key) or []
        if isinstance(gate, dict)
    }
    require(gates, f"{path.relative_to(REPO_ROOT)} {key} must not be empty")
    return gates


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
        fixture.get("kind") == "control_plane_read_implementation_trigger_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "control-plane-read-implementation-trigger-review-v1",
        "unexpected slice id",
    )
    require(slice_info.get("track") == "Control Plane / User Workspace / Workflow v1", "unexpected track")
    require(
        slice_info.get("status") == "implementation_trigger_review_defined",
        "implementation trigger review status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_trigger_review_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("trigger_review_boundary") or {}
    require(
        boundary.get("status") == "implementation_trigger_review_defined_no_implementation",
        "trigger review boundary status drifted",
    )
    require(
        boundary.get("decision") == "no_read_side_implementation_trigger_satisfied",
        "trigger review decision must remain no implementation trigger",
    )
    for field in (
        "implementation_task_card_allowed_now",
        "direct_runtime_implementation_allowed_now",
        "schema_artifact_files_allowed_now",
        "selector_files_allowed_now",
        "production_auth_files_allowed_now",
        "adapter_smoke_files_allowed_now",
        "repository_adapter_files_allowed_now",
        "database_connection_allowed_now",
        "write_allowed_now",
    ):
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_trigger_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("candidate_id") or ""): row
        for row in fixture.get("implementation_trigger_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == EXPECTED_CANDIDATE_IDS, "implementation trigger candidate ids drifted")
    for candidate_id, row in rows.items():
        source = EXPECTED_CANDIDATE_SOURCES[candidate_id]
        source_path, source_status = EXPECTED_DEPENDENCIES[source]
        require(row.get("source") == source, f"{candidate_id} source drifted")
        require(row.get("source_status") == source_status, f"{candidate_id} source status drifted")
        require(load_json(source_path).get("slice", {}).get("status") == source_status, f"{source} fixture drifted")
        require(row.get("decision") == "not_satisfied", f"{candidate_id} must remain not_satisfied")
        blockers = set(row.get("current_blockers") or [])
        require(blockers == EXPECTED_BLOCKERS[candidate_id], f"{candidate_id} blockers drifted")
        require(row.get("required_before_task_card"), f"{candidate_id} must list required evidence")
        for field in (
            "implementation_task_card_allowed_now",
            "implementation_artifacts_allowed_now",
            "direct_runtime_implementation_allowed_now",
        ):
            require(row.get(field) is False, f"{candidate_id} {field} must remain false")

        gate_path, gate_key = UPSTREAM_GATE_KEYS[candidate_id]
        upstream = gates_by_id(gate_path, gate_key)
        missing_upstream = sorted(blockers - set(upstream))
        require(not missing_upstream, f"{candidate_id} upstream blockers missing: {missing_upstream}")
        for blocker in blockers:
            require(
                upstream[blocker].get("status") == "not_satisfied",
                f"{candidate_id} blocker {blocker} must remain not_satisfied upstream",
            )


def assert_review_order(fixture: dict[str, Any]) -> None:
    rows = [row for row in fixture.get("review_order") or [] if isinstance(row, dict)]
    require(len(rows) == len(EXPECTED_CANDIDATE_IDS), "review order must cover every candidate")
    ordered_ids = [str(row.get("candidate_id") or "") for row in sorted(rows, key=lambda row: row.get("step"))]
    require(
        ordered_ids
        == [
            "schema_artifact_manifest_implementation",
            "store_selector_smoke_implementation",
            "production_auth_implementation",
            "adapter_smoke_execution",
        ],
        "review order drifted",
    )
    for index, row in enumerate(sorted(rows, key=lambda row: row.get("step")), start=1):
        require(row.get("step") == index, "review order step must be contiguous")
        require(str(row.get("reason") or "").strip(), "review order reason is required")


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
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


def assert_failure_and_safety_policies(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "missing_trigger_success_allowed",
        "readiness_only_promoted_to_implementation_allowed",
        "schema_readiness_promoted_to_manifest_allowed",
        "selector_readiness_promoted_to_selector_allowed",
        "auth_readiness_promoted_to_middleware_allowed",
        "adapter_smoke_readiness_promoted_to_execution_allowed",
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


def assert_docs_check_repo_and_no_leak(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-control-plane-read-adapter-smoke-readiness-v1.py"
    current_checker = "check-control-plane-read-implementation-trigger-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run implementation trigger review check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "implementation trigger review check must run after adapter smoke readiness",
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
                    f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r} in this trigger review slice",
                )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice(fixture)
    assert_trigger_review_boundary(fixture)
    assert_trigger_matrix(fixture)
    assert_review_order(fixture)
    assert_forbidden_artifacts(fixture)
    assert_failure_and_safety_policies(fixture)
    assert_docs_check_repo_and_no_leak(fixture)
    print("control plane read implementation trigger review v1 checks passed.")


if __name__ == "__main__":
    main()
