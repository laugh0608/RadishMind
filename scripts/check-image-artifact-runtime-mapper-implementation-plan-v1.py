#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-plan-v1.json"
BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json"
)
ENTRY_REVIEW_PATH = (
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
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
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
    "image-artifact-store-binary-reader-boundary-readiness-v1": (
        BOUNDARY_PATH,
        "image_artifact_store_binary_reader_boundary_readiness_defined",
    ),
    "image-artifact-runtime-mapping-implementation-entry-review-v1": (
        ENTRY_REVIEW_PATH,
        "image_artifact_runtime_mapping_entry_review_defined",
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
    "implementation_entry_opened",
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
    "implementation_entry_review_created_in_this_slice",
    "runtime_mapper_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "copilot_response_schema_changed_in_this_slice",
    "runtime_mapping_execution_allowed_now",
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
EXPECTED_PLAN_INPUTS = {
    "image_generation_artifact_metadata",
    "store_boundary_contract",
    "binary_reader_boundary_contract",
    "runtime_mapping_readiness_cases",
    "artifact_return_runbook",
    "image_safety_runbook",
    "backend_adapter_readiness_failure_envelope",
    "no_fake_fallback_policy",
    "no_side_effect_policy",
}
EXPECTED_ARTIFACT_METADATA_FIELDS = {
    "artifact_id",
    "artifact.uri",
    "artifact.sha256",
    "artifact.mime_type",
    "artifact.width",
    "artifact.height",
    "artifact.format",
    "artifact.title",
    "artifact.purpose",
    "generation.backend_id",
    "generation.model",
    "generation.seed",
    "safety.risk_level",
    "safety.requires_confirmation",
    "safety.review_status",
    "provenance.source_request_id",
    "provenance.trace_ids",
    "provenance.backend_request_id",
    "provenance.intent_id",
    "created_at",
}
EXPECTED_FORBIDDEN_INPUTS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "public_url",
    "signed_public_url",
    "storage_write_result",
    "binary_reader_result",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
}
EXPECTED_CITATION_FIELDS = {"id", "kind", "label", "locator", "source_uri"}
EXPECTED_METADATA_REFERENCE_FIELDS = {
    "artifact_id",
    "uri",
    "sha256",
    "mime_type",
    "dimensions.width",
    "dimensions.height",
    "format",
    "title",
    "purpose",
    "backend_id",
    "model",
    "seed",
    "safety.risk_level",
    "safety.requires_confirmation",
    "safety.review_status",
    "provenance.source_request_id",
    "provenance.trace_ids",
    "created_at",
}
EXPECTED_MAPPING_CASES = {
    "generated_not_required": (True, None),
    "generated_reviewed_pass": (True, None),
    "blocked_artifact_status": (False, "image_artifact_safety_blocked"),
    "failed_artifact_status": (False, "image_artifact_invalid_metadata"),
    "pending_review_artifact": (False, "image_artifact_safety_pending_review"),
}
EXPECTED_FAIL_CLOSED_CASES = {
    "invalid_metadata_missing_required_field": "image_artifact_invalid_metadata",
    "hash_mismatch": "image_artifact_hash_mismatch",
    "mime_mismatch": "image_artifact_mime_mismatch",
    "dimension_mismatch": "image_artifact_dimension_mismatch",
    "public_url_claim": "image_artifact_public_url_claim",
    "signed_url_policy_missing": "image_artifact_signed_url_policy_missing",
    "binary_payload_present": "image_artifact_binary_payload_rejected",
    "provider_raw_dump_present": "image_artifact_provider_raw_dump_rejected",
    "artifact_store_missing": "image_artifact_store_missing",
    "artifact_store_unavailable": "image_artifact_store_unavailable",
    "artifact_binary_reader_missing": "image_artifact_binary_reader_missing",
    "artifact_binary_read_forbidden": "image_artifact_binary_read_forbidden",
    "safety_review_not_passed": "image_artifact_safety_review_not_passed",
    "provenance_missing": "image_artifact_provenance_missing",
}
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
    "implementation_plan_promoted_to_runtime_implementation_allowed",
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


def value_at_path(document: dict[str, Any], dotted_path: str) -> Any:
    current: Any = document
    for part in dotted_path.split("."):
        require(isinstance(current, dict) and part in current, f"document missing {dotted_path}")
        current = current[part]
    return current


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(
        fixture.get("kind") == "image_artifact_runtime_mapper_implementation_plan_v1",
        "unexpected fixture kind",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapper-implementation-plan-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "track drifted")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapper_implementation_plan_defined",
        "implementation plan status drifted",
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

    boundary = dependencies["image-artifact-store-binary-reader-boundary-readiness-v1"]
    plan_inputs = boundary.get("runtime_mapping_plan_inputs") or {}
    require(plan_inputs.get("status") == "ready_as_boundary_evidence_only", "boundary plan inputs drifted")
    require(
        plan_inputs.get("runtime_mapping_implementation_plan_review_allowed_next") is True,
        "boundary readiness must allow the plan review next",
    )
    require(
        plan_inputs.get("runtime_mapper_implementation_allowed_now") is False,
        "boundary readiness must not allow runtime mapper implementation",
    )

    entry_review = dependencies["image-artifact-runtime-mapping-implementation-entry-review-v1"]
    entry_policy = entry_review.get("next_development_policy") or {}
    require(
        entry_policy.get("status") == "artifact_store_binary_boundary_readiness_next",
        "entry review must remain a historical boundary-readiness gate",
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
    require(value_at_path(artifact, "provenance.trace_ids"), "artifact provenance trace required")

    citation_def = ((response_schema.get("$defs") or {}).get("citation") or {}).get("properties") or {}
    citation_kind = ((citation_def.get("kind") or {}).get("enum")) or []
    require("artifact" in citation_kind, "CopilotResponse citation kind must still allow artifact")


def assert_plan_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_plan_boundary") or {}
    require(
        boundary.get("status") == "implementation_plan_defined_no_runtime_code",
        "implementation plan boundary status drifted",
    )
    require(boundary.get("decision") == "plan_only_before_runtime_mapper_entry_review", "boundary decision drifted")
    require(
        boundary.get("current_development_mode") == "runtime_mapper_implementation_entry_review_next",
        "current development mode drifted",
    )
    require(
        boundary.get("store_binary_boundary_consumed")
        == "image-artifact-store-binary-reader-boundary-readiness-v1",
        "store/binary boundary dependency drifted",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"implementation_plan_boundary.{field} must remain false")


def assert_mapper_input_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("mapper_input_contract") or {}
    require(contract.get("status") == "planned_not_implemented", "mapper input status drifted")
    require(contract.get("source_kind") == "image_generation_artifact", "mapper source kind drifted")
    require(
        EXPECTED_PLAN_INPUTS.issubset(set(contract.get("required_plan_inputs") or [])),
        "mapper plan inputs drifted",
    )
    require(
        EXPECTED_ARTIFACT_METADATA_FIELDS.issubset(set(contract.get("required_artifact_metadata_fields") or [])),
        "mapper artifact metadata fields drifted",
    )
    require(
        EXPECTED_FORBIDDEN_INPUTS.issubset(set(contract.get("forbidden_inputs") or [])),
        "mapper forbidden inputs drifted",
    )

    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    for dotted_path in EXPECTED_ARTIFACT_METADATA_FIELDS:
        value_at_path(artifact, dotted_path)


def assert_target_response_plan(fixture: dict[str, Any]) -> None:
    plan = fixture.get("target_response_reference_plan") or {}
    require(plan.get("status") == "planned_not_materialized", "target response plan status drifted")
    require(
        plan.get("target_surface") == "future CopilotResponse artifact citation plus metadata reference",
        "target surface drifted",
    )
    require(plan.get("target_citation_kind") == "artifact", "target citation kind drifted")
    require(EXPECTED_CITATION_FIELDS.issubset(set(plan.get("target_citation_fields") or [])), "citation fields drifted")
    require(
        EXPECTED_METADATA_REFERENCE_FIELDS.issubset(set(plan.get("target_metadata_reference_fields") or [])),
        "metadata reference fields drifted",
    )
    for field in (
        "copilot_response_schema_change_required_now",
        "runtime_mapper_required_now",
        "artifact_store_required_now",
        "binary_reader_required_now",
        "public_url_required_now",
    ):
        require(plan.get(field) is False, f"target_response_reference_plan.{field} must remain false")


def assert_mapping_case_plan(fixture: dict[str, Any], runtime_readiness: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("mapping_case_plan") or [], "case_id")
    runtime_rows = rows_by_id(runtime_readiness.get("response_mapping_matrix") or [], "case_id")
    require(set(rows) == set(EXPECTED_MAPPING_CASES), "mapping case plan drifted")
    require(set(EXPECTED_MAPPING_CASES).issubset(set(runtime_rows)), "runtime readiness cases drifted")

    for case_id, (allowed, failure_code) in EXPECTED_MAPPING_CASES.items():
        row = rows[case_id]
        runtime_row = runtime_rows[case_id]
        require(
            row.get("source_case") == "image-artifact-runtime-mapping-readiness-v1",
            f"{case_id} source case drifted",
        )
        require(row.get("success_response_reference_allowed") is allowed, f"{case_id} success policy drifted")
        require(
            row.get("runtime_mapping_implemented_now") is False,
            f"{case_id} runtime mapping must remain false",
        )
        require(
            runtime_row.get("success_response_reference_allowed") is allowed,
            f"{case_id} source readiness success policy drifted",
        )
        require(
            runtime_row.get("runtime_mapping_implemented_now") is False,
            f"{case_id} source readiness must not claim runtime implementation",
        )
        if allowed:
            require(row.get("target_citation_kind") == "artifact", f"{case_id} citation kind drifted")
            require(runtime_row.get("target_citation_kind") == "artifact", f"{case_id} source citation drifted")
        else:
            require(row.get("expected_failure_code") == failure_code, f"{case_id} failure code drifted")
            require(runtime_row.get("expected_failure_code") == failure_code, f"{case_id} source failure drifted")


def assert_fail_closed_plan(
    fixture: dict[str, Any],
    boundary: dict[str, Any],
    runtime_readiness: dict[str, Any],
) -> None:
    rows = rows_by_id(fixture.get("fail_closed_plan") or [], "case_id")
    require(set(rows) == set(EXPECTED_FAIL_CLOSED_CASES), "fail-closed plan drifted")

    boundary_failures = {
        str(row.get("failure_code") or "")
        for row in boundary.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    runtime_failures = {
        str(row.get("failure_code") or "")
        for row in runtime_readiness.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    known_failures = boundary_failures | runtime_failures

    for case_id, expected_failure in EXPECTED_FAIL_CLOSED_CASES.items():
        row = rows[case_id]
        require(row.get("expected_failure_code") == expected_failure, f"{case_id} failure code drifted")
        require(row.get("success_response_reference_allowed") is False, f"{case_id} must not return success")
        require(row.get("fail_closed") is True, f"{case_id} must fail closed")
        require(expected_failure in known_failures, f"{expected_failure} must be covered by upstream taxonomy")


def assert_single_track_and_entry_recommendation(fixture: dict[str, Any]) -> None:
    policy = fixture.get("single_track_implementation_policy") or {}
    require(policy.get("status") == "single_track_policy_defined", "single-track policy status drifted")
    require(
        policy.get("next_review_task_card_allowed")
        == "image-artifact-runtime-mapper-implementation-entry-v1-plan.md",
        "next review task card drifted",
    )
    for field in (
        "runtime_mapper_implementation_task_card_allowed_now",
        "artifact_store_implementation_task_card_allowed_now",
        "artifact_binary_reader_implementation_task_card_allowed_now",
        "public_url_resolver_implementation_task_card_allowed_now",
        "backend_adapter_implementation_task_card_allowed_now",
        "parallel_implementation_tracks_allowed_now",
    ):
        require(policy.get(field) is False, f"single_track_implementation_policy.{field} must remain false")
    require(
        policy.get("if_entry_review_satisfies_runtime_mapper") == "open_runtime_mapper_implementation_only",
        "runtime mapper branch policy drifted",
    )
    require(
        policy.get("if_entry_review_satisfies_store_first") == "open_artifact_store_implementation_only",
        "store-first branch policy drifted",
    )
    require(
        policy.get("if_entry_review_satisfies_reader_first") == "open_binary_reader_implementation_only",
        "reader-first branch policy drifted",
    )

    recommendation = fixture.get("implementation_entry_recommendation") or {}
    require(
        recommendation.get("status") == "plan_defined_entry_not_opened",
        "implementation entry recommendation status drifted",
    )
    require(
        recommendation.get("recommended_next_slice") == "image-artifact-runtime-mapper-implementation-entry-v1",
        "recommended next slice drifted",
    )
    require(
        recommendation.get("runtime_mapper_implementation_allowed_after_this_slice") is False,
        "runtime mapper implementation must remain blocked after this slice",
    )
    require(
        recommendation.get("requires_dedicated_entry_review") is True,
        "dedicated entry review must remain required",
    )
    require(
        recommendation.get("requires_fast_baseline_before_implementation") is True,
        "fast baseline requirement drifted",
    )
    require(
        set(recommendation.get("must_reconsume") or [])
        == {
            "image-artifact-runtime-mapper-implementation-plan-v1",
            "image-artifact-store-binary-reader-boundary-readiness-v1",
            "image-artifact-runtime-mapping-readiness-v1",
        },
        "implementation entry reconsume list drifted",
    )


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
    previous_checker = "check-image-artifact-store-binary-reader-boundary-readiness-v1.py"
    current_checker = "check-image-artifact-runtime-mapper-implementation-plan-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime mapper implementation plan")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime mapper implementation plan checker must run after store/binary boundary readiness",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    dependencies = assert_dependencies()
    assert_slice(fixture)
    assert_schema_contracts()
    assert_plan_boundary(fixture)
    assert_mapper_input_contract(fixture)
    assert_target_response_plan(fixture)
    assert_mapping_case_plan(fixture, dependencies["image-artifact-runtime-mapping-readiness-v1"])
    assert_fail_closed_plan(
        fixture,
        dependencies["image-artifact-store-binary-reader-boundary-readiness-v1"],
        dependencies["image-artifact-runtime-mapping-readiness-v1"],
    )
    assert_single_track_and_entry_recommendation(fixture)
    assert_forbidden_artifacts(fixture)
    assert_execution_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapper implementation plan v1 checks passed.")


if __name__ == "__main__":
    main()
