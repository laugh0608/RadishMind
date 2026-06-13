#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json"
PLAN_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-plan-v1.json"
BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json"
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
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

DEPENDENCY_STATUS_BY_SLICE = {
    "image-artifact-runtime-mapper-implementation-plan-v1": (
        PLAN_PATH,
        "image_artifact_runtime_mapper_implementation_plan_defined",
    ),
    "image-artifact-store-binary-reader-boundary-readiness-v1": (
        BOUNDARY_PATH,
        "image_artifact_store_binary_reader_boundary_readiness_defined",
    ),
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
EXPECTED_DEPENDENCIES = set(DEPENDENCY_STATUS_BY_SLICE) | {
    "contracts/image-generation-artifact.schema.json",
    "contracts/copilot-response.schema.json",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "implementation_task_card_created",
    "runtime_mapper_ready",
    "runtime_mapper_implemented",
    "runtime_mapper_files_created",
    "copilot_response_schema_changed",
    "copilot_response_artifact_runtime_ready",
    "artifact_store_ready",
    "artifact_store_implemented",
    "artifact_binary_reader_ready",
    "artifact_binary_reader_implemented",
    "artifact_binary_read_ready",
    "public_artifact_url_ready",
    "public_url_resolver_ready",
    "signed_url_resolver_ready",
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
    "runtime_mapper_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "copilot_response_schema_changed_in_this_slice",
    "artifact_store_implementation_task_card_allowed_next",
    "artifact_binary_reader_implementation_task_card_allowed_next",
    "public_url_resolver_implementation_task_card_allowed_next",
    "backend_adapter_implementation_task_card_allowed_next",
    "parallel_implementation_tracks_allowed_now",
    "runtime_mapping_execution_allowed_now",
    "runtime_mapper_runtime_code_allowed_now",
    "artifact_store_lookup_allowed_now",
    "artifact_binary_read_allowed_now",
    "artifact_upload_allowed_now",
    "production_storage_write_allowed_now",
    "public_url_allowed_now",
    "signed_url_allowed_now",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_GATE_IDS = {
    "runtime_mapper_implementation_plan_consumed",
    "store_binary_boundary_consumed",
    "runtime_mapping_readiness_consumed",
    "artifact_return_runbook_consumed",
    "safety_runbook_consumed",
    "backend_adapter_readiness_consumed",
    "failure_taxonomy_consumed",
    "single_track_selection_policy_consumed",
    "artifact_runtime_mapper_implementation_track_selected",
    "artifact_store_implementation_deferred",
    "artifact_binary_reader_implementation_deferred",
    "public_url_resolver_implementation_deferred",
    "backend_adapter_implementation_deferred",
    "no_runtime_artifacts_leaked",
}
SATISFIED_GATES = {
    "runtime_mapper_implementation_plan_consumed",
    "store_binary_boundary_consumed",
    "runtime_mapping_readiness_consumed",
    "artifact_return_runbook_consumed",
    "safety_runbook_consumed",
    "backend_adapter_readiness_consumed",
    "failure_taxonomy_consumed",
    "single_track_selection_policy_consumed",
}
DEFERRED_GATES = {
    "artifact_store_implementation_deferred",
    "artifact_binary_reader_implementation_deferred",
    "public_url_resolver_implementation_deferred",
    "backend_adapter_implementation_deferred",
}
EXPECTED_CANDIDATES = {
    "artifact_runtime_mapper_implementation",
    "artifact_store_implementation",
    "artifact_binary_reader_implementation",
    "public_url_resolver_implementation",
    "backend_adapter_implementation",
}
EXPECTED_DEFERRED_CANDIDATES = EXPECTED_CANDIDATES - {"artifact_runtime_mapper_implementation"}
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
    "contracts/image-artifact-store.schema.json",
    "contracts/image-artifact-binary-reader.schema.json",
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
    "IMAGE_ARTIFACT_STORE_URL",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "IMAGE_ARTIFACT_SIGNED_URL_TTL",
    "ReadImageArtifactBinary",
    "ResolveImageArtifactPublicURL",
    "runImageArtifactRuntimeMapping",
    "callImageGenerationBackend",
}
EXPECTED_ALLOWED_SCOPE_NEXT = {
    "metadata-only mapper input validation",
    "image_generation_artifact to CopilotResponse artifact citation projection",
    "metadata reference projection",
    "success mapping tests",
    "fail-closed mapping tests",
    "no side effects smoke",
}
EXPECTED_FORBIDDEN_SCOPE_NEXT = {
    "artifact store implementation",
    "artifact binary reader implementation",
    "public URL resolver implementation",
    "backend adapter implementation",
    "CopilotResponse schema change",
    "real backend call",
    "image generation",
    "artifact upload",
    "binary read",
}
EXPECTED_CONSTRAINT_TRUE_FIELDS = {
    "input_must_be_metadata_only",
    "must_preserve_sha256",
    "must_preserve_mime_type",
    "must_preserve_dimensions",
    "must_preserve_safety_review",
    "must_preserve_provenance",
    "blocked_failed_pending_review_fail_closed",
}
EXPECTED_CONSTRAINT_FALSE_FIELDS = {
    "artifact_uri_is_public_url",
    "artifact_store_lookup_allowed",
    "artifact_binary_read_allowed",
    "public_url_resolution_allowed",
    "backend_call_allowed",
    "copilot_response_schema_change_allowed",
    "retry_or_fallback_allowed",
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
    "entry_review_promoted_to_runtime_implementation_allowed",
    "selected_task_card_promoted_to_runtime_code_allowed",
    "missing_artifact_store_promoted_to_success_allowed",
    "missing_binary_reader_promoted_to_success_allowed",
    "blocked_artifact_promoted_to_success_allowed",
    "pending_review_promoted_to_success_allowed",
    "hash_mismatch_promoted_to_success_allowed",
    "mime_mismatch_promoted_to_success_allowed",
    "dimension_mismatch_promoted_to_success_allowed",
    "public_url_claim_promoted_to_success_allowed",
    "binary_payload_promoted_to_metadata_reference_allowed",
    "provider_raw_dump_promoted_to_metadata_reference_allowed",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "artifact_binary_read_count=0",
    "artifact_store_lookup_count=0",
    "runtime_mapping_execution_count=0",
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
    "artifact_store_lookup",
    "runtime_mapping_execution",
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


def value_at_path(document: dict[str, Any], dotted_path: str) -> Any:
    current: Any = document
    for part in dotted_path.split("."):
        require(isinstance(current, dict) and part in current, f"document missing {dotted_path}")
        current = current[part]
    return current


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_artifact_runtime_mapper_implementation_entry_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapper-implementation-entry-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "track drifted")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapper_implementation_entry_review_defined",
        "entry review status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependency list drifted")


def assert_dependencies() -> dict[str, dict[str, Any]]:
    dependencies: dict[str, dict[str, Any]] = {}
    for slice_id, (path, expected_status) in DEPENDENCY_STATUS_BY_SLICE.items():
        document = load_json(path)
        require(slice_status(document) == expected_status, f"{slice_id} status drifted")
        dependencies[slice_id] = document

    plan = dependencies["image-artifact-runtime-mapper-implementation-plan-v1"]
    recommendation = plan.get("implementation_entry_recommendation") or {}
    require(
        recommendation.get("recommended_next_slice") == "image-artifact-runtime-mapper-implementation-entry-v1",
        "plan must recommend this entry review",
    )
    require(recommendation.get("requires_dedicated_entry_review") is True, "plan must require dedicated entry review")
    require(
        recommendation.get("runtime_mapper_implementation_allowed_after_this_slice") is False,
        "plan must not directly allow implementation",
    )
    policy = plan.get("single_track_implementation_policy") or {}
    require(
        policy.get("next_review_task_card_allowed")
        == "image-artifact-runtime-mapper-implementation-entry-v1-plan.md",
        "plan single-track next review drifted",
    )

    boundary = dependencies["image-artifact-store-binary-reader-boundary-readiness-v1"]
    plan_inputs = boundary.get("runtime_mapping_plan_inputs") or {}
    require(plan_inputs.get("status") == "ready_as_boundary_evidence_only", "boundary plan input status drifted")
    require(
        plan_inputs.get("runtime_mapping_implementation_plan_review_allowed_next") is True,
        "boundary should feed mapper plan review",
    )
    require(
        plan_inputs.get("runtime_mapper_implementation_allowed_now") is False,
        "boundary must not directly allow runtime mapper",
    )

    runtime = dependencies["image-artifact-runtime-mapping-readiness-v1"]
    future_contract = runtime.get("future_mapping_contract") or {}
    require(future_contract.get("uri_scheme") == "artifact://", "runtime readiness URI scheme drifted")
    require(future_contract.get("artifact_store_required_now") is False, "artifact store must not be required now")
    require(future_contract.get("binary_reader_required_now") is False, "binary reader must not be required now")
    require(future_contract.get("runtime_mapping_required_now") is False, "runtime mapping must not be required now")

    artifact_return = dependencies["image-artifact-return-runbook-evidence-v1"]
    metadata_shape = artifact_return.get("metadata_reference_shape") or {}
    require(metadata_shape.get("uri_scheme") == "artifact://", "artifact return URI scheme drifted")
    require(metadata_shape.get("uri_is_public_url") is False, "artifact return must not claim public URL")
    visibility = artifact_return.get("artifact_visibility_policy") or {}
    require(visibility.get("public_url_visible_to_copilot_response") is False, "public URL visibility drifted")
    require(visibility.get("pixel_payload_visible_to_copilot_response") is False, "pixel payload visibility drifted")
    require(
        visibility.get("provider_raw_dump_visible_to_copilot_response") is False,
        "provider raw visibility drifted",
    )

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
    require("image_backend_response_untrusted" in backend_failures, "backend untrusted response failure drifted")
    return dependencies


def assert_schema_contracts() -> None:
    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)
    response_schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.Draft202012Validator.check_schema(artifact_schema)
    jsonschema.Draft202012Validator.check_schema(response_schema)
    jsonschema.validate(artifact, artifact_schema)

    uri = str(value_at_path(artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "artifact fixture URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "artifact fixture must not use public URL")
    require(len(str(value_at_path(artifact, "artifact.sha256"))) == 64, "artifact sha256 must remain 64 chars")
    require(str(value_at_path(artifact, "artifact.mime_type")).startswith("image/"), "artifact mime must remain image/*")
    require(int(value_at_path(artifact, "artifact.width")) > 0, "artifact width required")
    require(int(value_at_path(artifact, "artifact.height")) > 0, "artifact height required")
    require(value_at_path(artifact, "safety.review_status") == "not_required", "basic safety review drifted")
    require(value_at_path(artifact, "provenance.trace_ids"), "artifact provenance trace required")

    citation_def = ((response_schema.get("$defs") or {}).get("citation") or {}).get("properties") or {}
    citation_kind = ((citation_def.get("kind") or {}).get("enum")) or []
    require("artifact" in citation_kind, "CopilotResponse citation kind must still allow artifact")


def assert_entry_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_entry_boundary") or {}
    require(
        boundary.get("status") == "implementation_entry_review_defined_no_runtime_code",
        "entry boundary status drifted",
    )
    require(
        boundary.get("decision") == "runtime_mapper_implementation_task_card_allowed_next",
        "entry decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "runtime_mapper_implementation_task_card_next",
        "development mode drifted",
    )
    require(
        boundary.get("selected_implementation_track_now") == "artifact_runtime_mapper_implementation",
        "selected implementation track drifted",
    )
    require(
        boundary.get("runtime_mapper_implementation_task_card_allowed_next") is True,
        "runtime mapper implementation task card should be allowed next",
    )
    require(
        boundary.get("next_task_card_allowed") == "image-artifact-runtime-mapper-implementation-v1-plan.md",
        "next task card drifted",
    )
    require(
        boundary.get("if_next_task_card_created") == "metadata_only_runtime_mapper_implementation_only",
        "next task card scope policy drifted",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"implementation_entry_boundary.{field} must remain false")


def assert_entry_gates(fixture: dict[str, Any]) -> None:
    gates = rows_by_id(fixture.get("entry_gate_matrix") or [], "id")
    require(set(gates) == EXPECTED_GATE_IDS, "entry gate matrix drifted")
    for gate_id in SATISFIED_GATES:
        require(gates[gate_id].get("status") == "satisfied", f"{gate_id} must stay satisfied")
    for gate_id in DEFERRED_GATES:
        require(gates[gate_id].get("status") == "deferred", f"{gate_id} must stay deferred")
    selected = gates["artifact_runtime_mapper_implementation_track_selected"]
    require(selected.get("status") == "selected_next", "selected track gate drifted")
    require(
        selected.get("candidate_id") == "artifact_runtime_mapper_implementation",
        "selected track candidate drifted",
    )
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
    require(set(candidates) == EXPECTED_CANDIDATES, "candidate matrix drifted")
    selected = candidates["artifact_runtime_mapper_implementation"]
    require(selected.get("source") == "image-artifact-runtime-mapper-implementation-plan-v1", "selected source drifted")
    require(selected.get("source_status") == slice_status(dependencies[selected["source"]]), "selected status drifted")
    require(selected.get("entry_decision") == "selected_next_task_card", "selected decision drifted")
    require(selected.get("trigger_satisfied_now") is True, "selected trigger must be satisfied")
    require(selected.get("task_card_allowed_next") is True, "selected task card should be allowed next")
    require(selected.get("task_card_created_now") is False, "selected task card must not be created now")
    require(selected.get("implementation_artifacts_allowed_now") is False, "implementation artifacts must stay blocked")
    require(selected.get("direct_runtime_implementation_allowed_now") is False, "direct runtime implementation blocked")
    require(selected.get("runtime_code_allowed_now") is False, "runtime code must stay blocked in this slice")
    require(
        selected.get("future_task_card_after_trigger") == "image-artifact-runtime-mapper-implementation-v1-plan.md",
        "selected future task card drifted",
    )

    for candidate_id in EXPECTED_DEFERRED_CANDIDATES:
        row = candidates[candidate_id]
        source = str(row.get("source") or "")
        require(source in dependencies, f"{candidate_id} source dependency missing")
        require(row.get("source_status") == slice_status(dependencies[source]), f"{candidate_id} source status drifted")
        require(row.get("entry_decision") == "deferred", f"{candidate_id} must remain deferred")
        require(row.get("trigger_satisfied_now") is False, f"{candidate_id} trigger must remain false")
        require(row.get("task_card_allowed_next") is False, f"{candidate_id} task card must not be allowed next")
        require(row.get("task_card_created_now") is False, f"{candidate_id} task card must not be created")
        require(
            row.get("implementation_artifacts_allowed_now") is False,
            f"{candidate_id} implementation artifacts must stay blocked",
        )
        require(row.get("direct_runtime_implementation_allowed_now") is False, f"{candidate_id} direct impl blocked")
        require(row.get("runtime_code_allowed_now") is False, f"{candidate_id} runtime code blocked")
        require(row.get("current_blockers"), f"{candidate_id} must keep blockers")


def assert_reconciliation(fixture: dict[str, Any], plan: dict[str, Any], runtime: dict[str, Any]) -> None:
    reconciliation = fixture.get("implementation_preconditions_reconciliation") or {}
    require(
        reconciliation.get("status") == "entry_review_satisfied_for_runtime_mapper_task_card_only",
        "reconciliation status drifted",
    )
    require(
        reconciliation.get("runtime_mapper_implementation_plan_status") == slice_status(plan),
        "plan reconciliation status drifted",
    )
    require(
        reconciliation.get("runtime_mapping_readiness_status") == slice_status(runtime),
        "runtime readiness reconciliation status drifted",
    )

    mapping_rows = runtime.get("response_mapping_matrix") or []
    success_cases = [
        row for row in mapping_rows if isinstance(row, dict) and row.get("success_response_reference_allowed") is True
    ]
    blocked_cases = [
        row for row in mapping_rows if isinstance(row, dict) and row.get("success_response_reference_allowed") is False
    ]
    require(reconciliation.get("success_mapping_cases") == len(success_cases), "success case count drifted")
    require(reconciliation.get("blocked_mapping_cases") == len(blocked_cases), "blocked case count drifted")
    require(
        reconciliation.get("fail_closed_plan_cases") == len(plan.get("fail_closed_plan") or []),
        "fail-closed plan case count drifted",
    )
    require(
        reconciliation.get("runtime_mapper_implementation_task_card_allowed_next") is True,
        "runtime mapper task card should be allowed next",
    )
    for field in (
        "runtime_mapper_runtime_code_allowed_now",
        "runtime_mapper_ready_now",
        "artifact_store_ready_now",
        "artifact_binary_reader_ready_now",
        "public_url_resolver_ready_now",
        "backend_adapter_ready_now",
        "copilot_response_schema_change_allowed_now",
    ):
        require(reconciliation.get(field) is False, f"implementation_preconditions_reconciliation.{field} false")


def assert_selected_track_contract_and_constraints(fixture: dict[str, Any]) -> None:
    selected = fixture.get("selected_track_contract") or {}
    require(
        selected.get("status") == "runtime_mapper_track_selected_for_next_task_card",
        "selected track contract status drifted",
    )
    require(
        selected.get("selected_candidate") == "artifact_runtime_mapper_implementation",
        "selected candidate drifted",
    )
    require(
        selected.get("next_task_card") == "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md",
        "selected next task card drifted",
    )
    require(
        EXPECTED_ALLOWED_SCOPE_NEXT.issubset(set(selected.get("allowed_scope_next") or [])),
        "allowed next scope drifted",
    )
    require(
        EXPECTED_FORBIDDEN_SCOPE_NEXT.issubset(set(selected.get("forbidden_scope_next") or [])),
        "forbidden next scope drifted",
    )
    require(
        selected.get("required_next_checker_order_after")
        == "check-image-artifact-runtime-mapper-implementation-entry-v1.py",
        "next checker order hint drifted",
    )

    constraints = fixture.get("next_implementation_constraints") or {}
    require(
        constraints.get("status") == "metadata_only_runtime_mapper_constraints_defined",
        "next constraints status drifted",
    )
    require(constraints.get("artifact_uri_scheme") == "artifact://", "artifact URI scheme drifted")
    for field in EXPECTED_CONSTRAINT_TRUE_FIELDS:
        require(constraints.get(field) is True, f"next_implementation_constraints.{field} must remain true")
    for field in EXPECTED_CONSTRAINT_FALSE_FIELDS:
        require(constraints.get(field) is False, f"next_implementation_constraints.{field} must remain false")


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    task_cards = rows_by_id(fixture.get("forbidden_task_card_matrix") or [], "path")
    require(set(task_cards) == EXPECTED_FORBIDDEN_TASK_CARDS, "forbidden task cards drifted")
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")
    require(
        task_cards["docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md"].get("allowed_next")
        is True,
        "runtime mapper implementation task card should be the only allowed next card",
    )
    for relative_path, row in task_cards.items():
        if relative_path != "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md":
            require(row.get("allowed_next") is False, f"{relative_path} must not be allowed next")

    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifacts drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
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
    previous_checker = "check-image-artifact-runtime-mapper-implementation-plan-v1.py"
    current_checker = "check-image-artifact-runtime-mapper-implementation-entry-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime mapper implementation entry review")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime mapper implementation entry checker must run after implementation plan",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    dependencies = assert_dependencies()
    assert_slice(fixture)
    assert_schema_contracts()
    assert_entry_boundary(fixture)
    assert_entry_gates(fixture)
    assert_candidates(fixture, dependencies)
    assert_reconciliation(
        fixture,
        dependencies["image-artifact-runtime-mapper-implementation-plan-v1"],
        dependencies["image-artifact-runtime-mapping-readiness-v1"],
    )
    assert_selected_track_contract_and_constraints(fixture)
    assert_forbidden_artifacts(fixture)
    assert_execution_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapper implementation entry v1 checks passed.")


if __name__ == "__main__":
    main()
