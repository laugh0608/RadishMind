#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-runtime-mapping-implementation-entry-review-v1.json"
)
RUNTIME_MAPPING_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json"
)
ARTIFACT_RETURN_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
)
SAFETY_PATH = REPO_ROOT / "scripts/checks/fixtures/image-safety-runbook-evidence-v1.json"
BACKEND_ADAPTER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json"
)
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD = (
    "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md"
)
RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json"
)
RUNTIME_MAPPER_MODULE_PATH = "services/runtime/image_artifact_runtime_mapper.py"
RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)

DEPENDENCY_STATUS_BY_SLICE = {
    "image-artifact-runtime-mapping-readiness-v1": (
        RUNTIME_MAPPING_READINESS_PATH,
        "image_artifact_runtime_mapping_readiness_defined",
    ),
    "image-artifact-return-runbook-evidence-v1": (
        ARTIFACT_RETURN_PATH,
        "image_artifact_return_runbook_evidence_defined",
    ),
    "image-safety-runbook-evidence-v1": (
        SAFETY_PATH,
        "image_safety_runbook_evidence_defined",
    ),
    "image-backend-adapter-readiness-evidence-v1": (
        BACKEND_ADAPTER_PATH,
        "image_backend_adapter_readiness_defined",
    ),
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implementation_entry_opened",
    "implementation_task_card_created",
    "runtime_mapper_ready",
    "runtime_mapper_implemented",
    "copilot_response_schema_changed",
    "copilot_response_artifact_runtime_ready",
    "artifact_store_ready",
    "artifact_store_implemented",
    "artifact_binary_reader_ready",
    "artifact_binary_reader_implemented",
    "public_artifact_url_ready",
    "public_url_resolver_ready",
    "production_artifact_storage_ready",
    "image_backend_adapter_implemented",
    "image_backend_client_ready",
    "real_backend_call_ready",
    "image_generation_ready",
    "image_pixel_payload_ready",
    "provider_raw_dump_ready",
    "automatic_retry_ready",
    "fallback_execution_ready",
    "executor_ready",
    "confirmation_decision_ready",
    "writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "implementation_task_card_created_in_this_slice",
    "parallel_implementation_tracks_allowed_now",
    "runtime_mapper_files_allowed_now",
    "artifact_store_files_allowed_now",
    "artifact_binary_reader_files_allowed_now",
    "public_url_resolver_files_allowed_now",
    "backend_adapter_files_allowed_now",
    "copilot_response_schema_change_allowed_now",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "artifact_upload_allowed_now",
    "artifact_binary_read_allowed_now",
    "public_url_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_GATE_IDS = {
    "runtime_mapping_readiness_consumed",
    "artifact_return_runbook_consumed",
    "safety_runbook_consumed",
    "backend_adapter_readiness_consumed",
    "artifact_store_binary_boundary_missing",
    "runtime_mapper_implementation_plan_missing",
    "backend_adapter_implementation_missing",
    "storage_url_failure_gate_missing",
    "single_track_selection_policy_defined",
    "no_runtime_artifacts_leaked",
}
SATISFIED_GATES = {
    "runtime_mapping_readiness_consumed",
    "artifact_return_runbook_consumed",
    "safety_runbook_consumed",
    "backend_adapter_readiness_consumed",
    "single_track_selection_policy_defined",
}
BLOCKED_GATES = {
    "artifact_store_binary_boundary_missing",
    "runtime_mapper_implementation_plan_missing",
    "backend_adapter_implementation_missing",
    "storage_url_failure_gate_missing",
}
EXPECTED_CANDIDATES = {
    "artifact_runtime_mapper_implementation",
    "artifact_store_implementation",
    "artifact_binary_reader_implementation",
    "public_url_resolver_implementation",
    "backend_adapter_implementation",
}
EXPECTED_REVIEW_ORDER = [
    "artifact_store_binary_boundary_readiness",
    "artifact_runtime_mapper_implementation",
    "backend_adapter_implementation",
    "artifact_store_implementation",
    "artifact_binary_reader_implementation",
    "public_url_resolver_implementation",
]
EXPECTED_FORBIDDEN_TASK_CARDS = {
    "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-store-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-binary-reader-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-public-url-resolver-implementation-v1-plan.md",
    "docs/task-cards/image-backend-adapter-implementation-v1-plan.md",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/copilot_response_artifact_mapper.py",
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "contracts/image-artifact-runtime-mapping.schema.json",
    "services/platform/internal/httpapi/image_artifacts.go",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactRuntimeMappingPanel.tsx",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageArtifactRuntimeMapper",
    "CopilotResponseArtifactMapper",
    "ImageArtifactStore",
    "ImageArtifactBinaryReader",
    "ImageArtifactPublicUrlResolver",
    "ImageBackendAdapter",
    "image_artifact_runtime_mapper",
    "copilot_response_artifact_mapper",
    "image_artifact_store",
    "image_artifact_binary_reader",
    "image_artifact_public_url_resolver",
    "image_backend_adapter",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "IMAGE_ARTIFACT_STORE_URL",
    "ImageArtifactRuntimeMappingPanel",
    "runImageArtifactRuntimeMapping",
    "ReadImageArtifactBinary",
    "callImageGenerationBackend",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_upload_artifacts",
    "does_not_read_artifact_binary",
    "does_not_write_production_storage",
    "does_not_create_public_url",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_NO_FAKE_FALSE_FIELDS = {
    "readiness_promoted_to_implementation_allowed",
    "blocked_entry_promoted_to_task_card_allowed",
    "missing_artifact_store_promoted_to_success_allowed",
    "missing_binary_reader_promoted_to_success_allowed",
    "public_url_claim_promoted_to_success_allowed",
    "backend_adapter_readiness_promoted_to_backend_call_allowed",
    "runtime_mapper_missing_promoted_to_response_reference_allowed",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "artifact_binary_read_count=0",
    "production_storage_write_count=0",
    "public_url_resolution_count=0",
    "dev_server_start_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}
EXPECTED_FORBIDDEN_SIDE_EFFECTS = {
    "real_backend_call",
    "image_pixel_generation",
    "model_download",
    "artifact_upload",
    "artifact_binary_read",
    "production_storage_write",
    "public_url_resolution",
    "dev_server_start",
    "browser_plugin_open",
    "executor_call",
    "confirmation_decision",
    "business_writeback",
    "replay_execution",
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


def later_entry_allows_task_card(relative_path: str) -> bool:
    if relative_path != RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD:
        return False
    if not RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH.exists():
        return False

    entry = load_json(RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH)
    boundary = entry.get("implementation_entry_boundary") or {}
    selected = entry.get("selected_track_contract") or {}
    return (
        boundary.get("decision") == "runtime_mapper_implementation_task_card_allowed_next"
        and selected.get("next_task_card") == RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD
    )


def later_runtime_mapper_implementation_allows_artifact(relative_path: str) -> bool:
    if relative_path != RUNTIME_MAPPER_MODULE_PATH:
        return False
    if not RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH.exists():
        return False

    implementation = load_json(RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH)
    runtime = implementation.get("runtime_implementation") or {}
    return (
        (implementation.get("slice") or {}).get("status")
        == "image_artifact_runtime_mapper_runtime_implemented"
        and runtime.get("status") == "metadata_only_runtime_mapper_implemented"
        and runtime.get("module") == RUNTIME_MAPPER_MODULE_PATH
    )


def rows_by_id(rows: list[Any], key: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for row in rows:
        require(isinstance(row, dict), f"{key} rows must be JSON objects")
        row_id = str(row.get(key) or "")
        require(row_id, f"{key} row missing id")
        result[row_id] = row
    return result


def slice_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    require(isinstance(slice_info, dict), "slice must be a JSON object")
    return str(slice_info.get("status") or "")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "image_artifact_runtime_mapping_implementation_entry_review_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapping-implementation-entry-review-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "track drifted")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapping_entry_review_defined",
        "entry review status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")
    require(set(fixture.get("depends_on") or []) == set(DEPENDENCY_STATUS_BY_SLICE), "dependency list drifted")


def assert_dependencies() -> dict[str, dict[str, Any]]:
    dependencies: dict[str, dict[str, Any]] = {}
    for slice_id, (path, expected_status) in DEPENDENCY_STATUS_BY_SLICE.items():
        document = load_json(path)
        require(slice_status(document) == expected_status, f"{slice_id} status drifted")
        dependencies[slice_id] = document

    artifact_return = dependencies["image-artifact-return-runbook-evidence-v1"]
    metadata_shape = artifact_return.get("metadata_reference_shape") or {}
    require(metadata_shape.get("uri_scheme") == "artifact://", "artifact return URI scheme drifted")
    require(metadata_shape.get("uri_is_public_url") is False, "artifact return must not claim public URL")
    visibility = artifact_return.get("artifact_visibility_policy") or {}
    require(visibility.get("pixel_payload_visible_to_copilot_response") is False, "pixel payload visibility drifted")
    require(visibility.get("provider_raw_dump_visible_to_copilot_response") is False, "provider raw visibility drifted")

    safety = dependencies["image-safety-runbook-evidence-v1"]
    safety_cases = rows_by_id(safety.get("safety_decision_matrix") or [], "case_id")
    require(
        safety_cases["artifact_review_pending"].get("metadata_reference_allowed_by_runbook") is False,
        "pending review must stay blocked",
    )
    require(
        safety_cases["artifact_review_blocked"].get("metadata_reference_allowed_by_runbook") is False,
        "blocked review must stay blocked",
    )

    backend = dependencies["image-backend-adapter-readiness-evidence-v1"]
    backend_failures = {
        str(row.get("failure_code") or "")
        for row in backend.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require("image_backend_invalid_artifact_metadata" in backend_failures, "backend invalid metadata failure drifted")
    require("image_backend_artifact_hash_mismatch" in backend_failures, "backend hash mismatch failure drifted")
    require("image_backend_response_untrusted" in backend_failures, "backend untrusted failure drifted")
    return dependencies


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_entry_boundary") or {}
    require(
        boundary.get("status") == "implementation_entry_review_defined_no_runtime_change",
        "entry boundary status drifted",
    )
    require(
        boundary.get("decision") == "runtime_mapping_implementation_entry_not_opened",
        "entry boundary decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "artifact_store_binary_boundary_readiness_next",
        "current development mode drifted",
    )
    require(boundary.get("selected_implementation_track_now") == "none", "implementation track must not be selected")
    require(
        boundary.get("if_entry_remains_blocked") == "create_artifact_store_binary_boundary_readiness_before_any_runtime_mapper",
        "blocked-entry policy drifted",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"implementation_entry_boundary.{field} must remain false")


def assert_entry_gates(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture.get("entry_gate_matrix") or [], "id")
    require(set(gates) == EXPECTED_GATE_IDS, "entry gate matrix drifted")
    for gate_id in SATISFIED_GATES:
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must stay satisfied")
    for gate_id in BLOCKED_GATES:
        require(gates[gate_id].get("status") == "blocked", f"{gate_id} must stay blocked")
    require(gates["no_runtime_artifacts_leaked"].get("status") == "required_now", "runtime leak gate drifted")
    required_runtime_locks = {
        "no runtime mapper",
        "no artifact store",
        "no binary reader",
        "no public URL resolver",
        "no backend adapter implementation",
        "no response schema change",
    }
    require(
        required_runtime_locks.issubset(set(gates["no_runtime_artifacts_leaked"].get("must_cover") or [])),
        "runtime leak gate coverage drifted",
    )


def assert_candidates(fixture: dict[str, Any], dependencies: dict[str, dict[str, Any]]) -> None:
    candidates = rows_by_id(fixture.get("implementation_entry_candidates") or [], "candidate_id")
    require(set(candidates) == EXPECTED_CANDIDATES, "implementation candidate matrix drifted")
    for candidate_id, row in candidates.items():
        source = str(row.get("source") or "")
        require(source in dependencies, f"{candidate_id} source dependency missing")
        require(row.get("source_status") == slice_status(dependencies[source]), f"{candidate_id} source status drifted")
        require(row.get("entry_decision") == "blocked", f"{candidate_id} must stay blocked")
        require(row.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(
            row.get("future_task_card_allowed_after_trigger") is True,
            f"{candidate_id} must keep future task card path",
        )
        require(row.get("task_card_allowed_now") is False, f"{candidate_id} task card must not be allowed now")
        require(
            row.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} implementation artifacts must not be allowed now",
        )
        require(
            row.get("direct_runtime_implementation_allowed_now") is False,
            f"{candidate_id} direct implementation must not be allowed now",
        )
        require(row.get("current_blockers"), f"{candidate_id} must keep blockers")


def assert_readiness_reconciliation(fixture: dict[str, Any], runtime_readiness: dict[str, Any]) -> None:
    reconciliation = fixture.get("runtime_mapping_readiness_reconciliation") or {}
    require(
        reconciliation.get("status") == "readiness_defined_not_implementation",
        "readiness reconciliation status drifted",
    )
    require(
        reconciliation.get("runtime_mapping_readiness_status") == slice_status(runtime_readiness),
        "readiness reconciliation source status drifted",
    )

    mapping_rows = runtime_readiness.get("response_mapping_matrix") or []
    success_cases = [
        row for row in mapping_rows if isinstance(row, dict) and row.get("success_response_reference_allowed") is True
    ]
    blocked_cases = [
        row for row in mapping_rows if isinstance(row, dict) and row.get("success_response_reference_allowed") is False
    ]
    implemented_now = [
        row for row in mapping_rows if isinstance(row, dict) and row.get("runtime_mapping_implemented_now") is not False
    ]
    require(not implemented_now, "runtime mapping readiness rows must not claim implementation")
    require(reconciliation.get("success_mapping_cases") == len(success_cases), "success mapping case count drifted")
    require(reconciliation.get("blocked_mapping_cases") == len(blocked_cases), "blocked mapping case count drifted")
    require(
        reconciliation.get("fail_closed_mapping_cases")
        == len(runtime_readiness.get("fail_closed_mapping_cases") or []),
        "fail-closed mapping case count drifted",
    )
    for field in (
        "future_mapping_contract_materialized_now",
        "runtime_mapper_ready_now",
        "artifact_store_ready_now",
        "artifact_binary_reader_ready_now",
        "public_url_resolver_ready_now",
        "implementation_trigger_satisfied_now",
    ):
        require(reconciliation.get(field) is False, f"readiness_reconciliation.{field} must remain false")


def assert_next_policy_and_order(fixture: dict[str, Any]) -> None:
    policy = fixture.get("next_development_policy") or {}
    require(policy.get("status") == "artifact_store_binary_boundary_readiness_next", "next policy status drifted")
    for field in (
        "artifact_store_binary_boundary_task_card_allowed_next",
        "runtime_mapping_implementation_requires_satisfied_entry",
        "implementation_entry_selects_one_track_only",
        "no_same_level_image_ui_panel_by_default",
        "no_real_backend_call_by_default",
        "no_response_schema_change_without_task_card",
        "no_public_url_or_binary_reader_without_boundary",
        "ordinary_doc_or_fixture_drift_uses_existing_gates",
    ):
        require(policy.get(field) is True, f"next_development_policy.{field} must remain true")

    ordered = fixture.get("entry_review_order") or []
    candidate_ids = [str(row.get("candidate_id") or "") for row in ordered if isinstance(row, dict)]
    require(candidate_ids == EXPECTED_REVIEW_ORDER, "entry review order drifted")
    for index, row in enumerate(ordered, start=1):
        require(isinstance(row, dict), "entry review rows must be JSON objects")
        require(row.get("step") == index, "entry review step sequence drifted")


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    task_cards = rows_by_id(fixture.get("forbidden_task_card_matrix") or [], "path")
    require(set(task_cards) == EXPECTED_FORBIDDEN_TASK_CARDS, "forbidden task cards drifted")
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if not later_entry_allows_task_card(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifacts drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if not later_runtime_mapper_implementation_allows_artifact(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    configured_absent_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_ABSENT_LITERALS.issubset(configured_absent_literals), "source absent literals drifted")


def assert_execution_and_side_effects(fixture: dict[str, Any]) -> None:
    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in EXPECTED_NO_FAKE_FALSE_FIELDS:
        require(fallback.get(field) is False, f"no_fake_fallback_policy.{field} must remain false")

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
        require(not missing, f"{relative_path} missing doc literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-image-artifact-runtime-mapping-readiness-v1.py"
    current_checker = "check-image-artifact-runtime-mapping-implementation-entry-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image artifact runtime mapping entry review")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "entry review checker must run after runtime mapping readiness",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    dependencies = assert_dependencies()
    assert_slice(fixture)
    assert_entry_boundary(fixture)
    assert_entry_gates(fixture)
    assert_candidates(fixture, dependencies)
    assert_readiness_reconciliation(fixture, dependencies["image-artifact-runtime-mapping-readiness-v1"])
    assert_next_policy_and_order(fixture)
    assert_forbidden_artifacts(fixture)
    assert_execution_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapping implementation entry review v1 checks passed.")


if __name__ == "__main__":
    main()
