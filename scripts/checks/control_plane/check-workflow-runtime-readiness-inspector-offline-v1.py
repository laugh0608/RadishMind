#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/workflow-runtime-readiness-inspector-offline-v1.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
EXPECTED_DEPENDENCIES = {
    "workflow-execution-plan-preview-offline-v1": (
        REPO_ROOT / "scripts/checks/fixtures/workflow-execution-plan-preview-offline-v1.json",
        "workflow_execution_plan_preview_offline_defined",
    ),
    "control-plane-read-implementation-trigger-review-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json",
        "implementation_trigger_review_defined",
    ),
    "control-plane-read-production-auth-readiness-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-production-auth-readiness-v1.json",
        "production_auth_readiness_defined",
    ),
    "control-plane-read-schema-artifact-manifest-readiness-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-schema-artifact-manifest-readiness-v1.json",
        "schema_artifact_manifest_readiness_defined",
    ),
    "control-plane-read-store-selector-smoke-readiness-v1": (
        REPO_ROOT / "scripts/checks/fixtures/control-plane-read-store-selector-smoke-readiness-v1.json",
        "store_selector_smoke_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_api_implemented",
    "production_api_consumer_ready",
    "workflow_builder_ready",
    "runtime_readiness_persistence_ready",
    "workflow_executor_ready",
    "node_executor_ready",
    "tool_executor_ready",
    "confirmation_flow_ready",
    "confirmation_decision_ready",
    "confirmation_decision_store_ready",
    "execution_unlock_ready",
    "business_writeback_ready",
    "run_replay_ready",
    "run_resume_ready",
    "durable_run_store_ready",
    "durable_result_store_ready",
    "materialized_result_reader_ready",
    "database_ready",
    "repository_adapter_ready",
    "store_selector_ready",
    "radish_oidc_ready",
    "production_ready",
}
EXPECTED_CAPABILITIES = {
    "workflow_runtime_readiness_inspector_view_model",
    "execution_plan_readiness_projection",
    "runtime_prerequisite_matrix_display",
    "readiness_blocker_display",
    "implementation_gate_display",
    "auth_db_repository_gate_display",
    "provider_binding_readiness_display",
    "durable_store_readiness_display",
    "writeback_replay_policy_display",
    "route_request_audit_metadata_display",
    "forbidden_output_guard_reused",
    "no_live_backend_request",
    "no_runtime_execution_controls",
}
EXPECTED_REQUIRED_FIELDS = {
    "selectedDraftId",
    "summary",
    "runtimePrerequisites",
    "readinessBlockers",
    "implementationGates",
    "auditMetadata",
    "routePath",
    "requestId",
    "auditRef",
    "canRenderRuntimeReadinessInspector",
    "canInspectReadinessLocally",
    "canRequestLiveBackend",
    "canPersistReadinessResult",
    "canStartWorkflowRuntime",
    "canExecuteWorkflow",
    "canSubmitConfirmationDecision",
    "canWriteBusinessTruth",
    "canReplayRun",
    "canEnableProductionAuth",
    "canAttachDatabase",
    "canImplementRepositoryAdapter",
}
EXPECTED_STATUSES = {"satisfied", "review_required", "blocked"}
EXPECTED_PREREQUISITES = {
    "runtime_executor_implementation",
    "provider_profile_binding",
    "confirmation_decision_store",
    "durable_run_result_store",
    "audit_policy_projection",
    "writeback_policy_gate",
    "replay_policy_gate",
    "auth_db_repository_gate",
    "workflow_publish_lifecycle_gate",
}
EXPECTED_BLOCKERS = {
    "blocker_runtime_executor_implementation",
    "blocker_provider_profile_binding",
    "blocker_confirmation_decision_store",
    "blocker_durable_run_result_store",
    "blocker_writeback_policy_gate",
    "blocker_replay_policy_gate",
    "blocker_auth_db_repository_gate",
    "blocker_workflow_publish_lifecycle_gate",
}
EXPECTED_GATES = {
    "gate_runtime_implementation_trigger_review",
    "gate_executor_before_runtime_start",
    "gate_confirmation_store_before_tool_execution",
    "gate_auth_repository_before_live_backend",
    "gate_writeback_replay_policy_before_unblock",
}
EXPECTED_STOP_LINES = {
    "no new runtime API",
    "no live backend request",
    "no runtime readiness persistence",
    "no workflow executor",
    "no node executor",
    "no tool executor",
    "no confirmation decision",
    "no confirmation decision store",
    "no execution unlock",
    "no durable run store",
    "no durable result store",
    "no materialized result reader",
    "no business writeback",
    "no replay or resume",
    "no database attach",
    "no Radish OIDC implementation",
    "no repository adapter implementation",
    "no production readiness claim",
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


def assert_slice_and_page(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "workflow_runtime_readiness_inspector_offline_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "workflow-runtime-readiness-inspector-offline-v1", "unexpected slice id")
    require(slice_info.get("track") == "Workflow / Agent Runtime Function Surface v1", "unexpected track")
    require(
        slice_info.get("status") == "workflow_runtime_readiness_inspector_offline_defined",
        "workflow runtime readiness inspector status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")

    page = fixture.get("page") or {}
    require(page.get("id") == "workflow-runtime-readiness-inspector-offline", "page id drifted")
    require(page.get("source_page_id") == "workflow-execution-plan-preview-offline", "source page drifted")
    require(page.get("source_route_id") == "workflow-definition-summary-list-route", "source route drifted")
    require(
        page.get("plan_route_id") == "workflow-execution-plan-preview-offline-draft",
        "plan route id drifted",
    )
    require(
        page.get("readiness_route_id") == "workflow-runtime-readiness-inspector-offline-draft",
        "readiness route id drifted",
    )
    require(page.get("status") == "offline_runtime_readiness_inspector_surface", "page status drifted")
    require(
        page.get("default_selected_draft_id") == "draft_wf_radishflow_copilot_latest_offline",
        "default selected draft id drifted",
    )
    require(page.get("local_inspection_allowed_now") is True, "local inspection must remain true")
    for field in (
        "live_backend_allowed_now",
        "runtime_api_allowed_now",
        "runtime_readiness_persistence_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
        "production_auth_allowed_now",
        "database_attach_allowed_now",
        "repository_adapter_allowed_now",
    ):
        require(page.get(field) is False, f"{field} must remain false")


def assert_capability_contract(fixture: dict[str, Any]) -> None:
    missing_capabilities = sorted(EXPECTED_CAPABILITIES - set(fixture.get("implemented_capabilities") or []))
    require(not missing_capabilities, f"missing implemented capabilities: {missing_capabilities}")
    missing_fields = sorted(EXPECTED_REQUIRED_FIELDS - set(fixture.get("required_view_model_fields") or []))
    require(not missing_fields, f"missing required view model fields: {missing_fields}")

    policy = fixture.get("readiness_policy") or {}
    require(EXPECTED_STATUSES.issubset(set(policy.get("required_statuses") or [])), "statuses drifted")
    require(
        EXPECTED_PREREQUISITES.issubset(set(policy.get("required_prerequisites") or [])),
        "prerequisites drifted",
    )
    require(EXPECTED_BLOCKERS.issubset(set(policy.get("required_blockers") or [])), "blockers drifted")
    require(EXPECTED_GATES.issubset(set(policy.get("required_gates") or [])), "gates drifted")
    require(policy.get("minimum_prerequisite_count") >= 9, "prerequisite count drifted")
    require(policy.get("minimum_blocker_count") >= 7, "blocker count drifted")
    require(policy.get("minimum_gate_count") >= 5, "gate count drifted")
    require(policy.get("local_inspection_allowed_now") is True, "local inspection must remain true")
    for field in (
        "live_backend_allowed_now",
        "persist_readiness_allowed_now",
        "workflow_execution_allowed_now",
        "confirmation_submission_allowed_now",
        "business_writeback_allowed_now",
        "run_replay_allowed_now",
        "production_auth_allowed_now",
        "database_attach_allowed_now",
        "repository_adapter_allowed_now",
    ):
        require(policy.get(field) is False, f"{field} must remain false")
    require(EXPECTED_STOP_LINES.issubset(set(fixture.get("stop_lines") or [])), "stop lines drifted")


def assert_required_files(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("required_files") or []:
        require((REPO_ROOT / str(relative_path)).is_file(), f"missing required file: {relative_path}")


def assert_source_boundaries(fixture: dict[str, Any]) -> None:
    readiness_source = read("apps/radishmind-web/src/features/control-plane-read/workflowRuntimeReadinessInspector.ts")
    app_source = read("apps/radishmind-web/src/app/App.tsx")
    styles = read("apps/radishmind-web/src/styles.css")
    checked_source = "\n".join([readiness_source, app_source])

    for literal in (
        "buildWorkflowRuntimeReadinessInspectorViewModel",
        "WorkflowRuntimeReadinessInspectorViewModel",
        "WorkflowRuntimeReadinessPrerequisite",
        "WorkflowRuntimeReadinessBlocker",
        "WorkflowRuntimeReadinessGate",
        "workflow-runtime-readiness-inspector-offline",
        "workflow-runtime-readiness-inspector-offline-draft",
        "workflow-execution-plan-preview-offline",
        "runtime_executor_implementation",
        "provider_profile_binding",
        "confirmation_decision_store",
        "durable_run_result_store",
        "audit_policy_projection",
        "writeback_policy_gate",
        "replay_policy_gate",
        "auth_db_repository_gate",
        "workflow_publish_lifecycle_gate",
        "gate_runtime_implementation_trigger_review",
        "gate_executor_before_runtime_start",
        "gate_confirmation_store_before_tool_execution",
        "gate_auth_repository_before_live_backend",
        "gate_writeback_replay_policy_before_unblock",
        "control-plane-read-implementation-trigger-review-v1",
        "control-plane-read-production-auth-readiness-v1",
        "control-plane-read-schema-artifact-manifest-readiness-v1",
        "canRenderRuntimeReadinessInspector",
        "canInspectReadinessLocally: true",
        "canRequestLiveBackend: false",
        "canPersistReadinessResult: false",
        "canStartWorkflowRuntime: false",
        "canExecuteWorkflow: false",
        "canSubmitConfirmationDecision: false",
        "canWriteBusinessTruth: false",
        "canReplayRun: false",
        "canEnableProductionAuth: false",
        "canAttachDatabase: false",
        "canImplementRepositoryAdapter: false",
    ):
        require(literal in readiness_source, f"runtime readiness source missing literal: {literal}")

    for literal in (
        "WorkflowRuntimeReadinessInspectorPanel",
        "WorkflowRuntimeReadinessSummaryCard",
        "WorkflowRuntimeReadinessPrerequisiteCard",
        "WorkflowRuntimeReadinessBlockerCard",
        "WorkflowRuntimeReadinessGateCard",
        "workflowRuntimeReadinessInspector",
        "workflow-runtime-readiness-inspector",
    ):
        require(literal in app_source, f"App.tsx missing runtime readiness literal: {literal}")

    for literal in (
        ".workflow-runtime-readiness-inspector",
        ".workflow-runtime-readiness-summary-grid",
        ".workflow-runtime-readiness-prerequisite-grid",
        ".workflow-runtime-readiness-blocker-grid",
        ".workflow-runtime-readiness-gate-grid",
        ".workflow-runtime-readiness-card",
        ".workflow-runtime-readiness-prerequisite",
        ".workflow-runtime-readiness-blocker",
        ".workflow-runtime-readiness-gate",
        ".workflow-runtime-readiness-row-main",
    ):
        require(literal in styles, f"styles missing runtime readiness literal: {literal}")

    for status in EXPECTED_STATUSES:
        require(f'"{status}"' in readiness_source, f"missing readiness status literal: {status}")
    for prerequisite_id in EXPECTED_PREREQUISITES:
        require(f'"{prerequisite_id}"' in readiness_source, f"missing prerequisite id: {prerequisite_id}")
    for gate_id in EXPECTED_GATES:
        require(f'"{gate_id}"' in readiness_source, f"missing gate id: {gate_id}")

    for forbidden in fixture.get("forbidden_source_literals") or []:
        require(str(forbidden) not in checked_source, f"runtime readiness source contains forbidden literal: {forbidden}")


def assert_docs_and_fast_baseline(fixture: dict[str, Any]) -> None:
    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-workflow-execution-plan-preview-offline-v1.py"
    current_checker = "check-workflow-runtime-readiness-inspector-offline-v1.py"
    require(current_checker in check_repo, "check-repo.py must run workflow runtime readiness inspector check")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime readiness inspector check must run after execution plan preview check",
    )

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = read(str(relative_path))
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing literals: {missing}")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_dependencies(fixture)
    assert_slice_and_page(fixture)
    assert_capability_contract(fixture)
    assert_required_files(fixture)
    assert_source_boundaries(fixture)
    assert_docs_and_fast_baseline(fixture)
    print("workflow runtime readiness inspector offline v1 checks passed.")


if __name__ == "__main__":
    main()
