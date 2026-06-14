#!/usr/bin/env python3
from __future__ import annotations

import copy
import json
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_artifact_response_consumer import (  # noqa: E402
    apply_image_artifact_reference_to_response,
    response_consumer_side_effect_counters,
)
from services.runtime.image_artifact_runtime_mapper import (  # noqa: E402
    map_image_artifact_to_response_reference,
    runtime_mapper_side_effect_counters,
)


FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-entry-review-v1.json"
)
RESPONSE_CONSUMER_RUNTIME_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json"
)
RESPONSE_CONSUMER_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-implementation-v1.json"
)
RUNTIME_MAPPER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
SUCCESSOR_INTEGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json"
)

EXPECTED_DEPENDENCIES = {
    "image-artifact-response-consumer-runtime-implementation-v1",
    "image-artifact-response-consumer-implementation-v1",
    "image-artifact-runtime-mapper-runtime-implementation-v1",
    "contracts/copilot-response.schema.json",
    "services/runtime/image_artifact_response_consumer.py",
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/inference_response.py",
    "services/runtime/inference_support.py",
    "services/gateway/copilot_gateway.py",
    "services/platform/internal/httpapi/responses.go",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "response_builder_changed",
    "response_builder_connected",
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
EXPECTED_ENTRY_CANDIDATES = {
    "python_response_coercion_post_citations": (
        "selected_future_hook",
        "services/runtime/inference_response.py#coerce_response_document",
    ),
    "inference_support_context_citation_builder": (
        "not_selected_context_citation_builder",
        "services/runtime/inference_support.py#build_citations",
    ),
    "gateway_envelope_handler": (
        "blocked_gateway_envelope_boundary",
        "services/gateway/copilot_gateway.py#handle_copilot_request",
    ),
    "platform_responses_route": (
        "blocked_northbound_protocol_boundary",
        "services/platform/internal/httpapi/responses.go#buildOpenAIResponsesResponse",
    ),
}
EXPECTED_GATE_STATUS = {
    "runtime_mapper_success_required": "satisfied",
    "response_consumer_success_required": "satisfied",
    "copilot_response_schema_validation_required": "satisfied",
    "citation_id_conflict_fail_closed_required": "satisfied",
    "metadata_reference_internal_only_required": "satisfied",
    "current_response_builder_unmodified": "satisfied",
    "store_reader_public_url_backend_tracks": "blocked",
    "response_builder_connection_task_card": "deferred",
}
EXPECTED_FAILURE_CODES = {
    "image_artifact_mapper_failed",
    "image_artifact_citation_id_conflict",
    "image_artifact_citation_schema_invalid",
    "image_artifact_metadata_reference_leak",
}
EXPECTED_ALLOWED_NEXT_SCOPE = {
    "create image-artifact-response-builder-integration-v1 task card",
    "define exact coerce_response_document hook placement",
    "define request artifact metadata discovery input",
    "define post-merge CopilotResponse schema validation",
    "define source-level no side effects checks",
}
EXPECTED_FORBIDDEN_NEXT_SCOPE = {
    "direct response builder code connection in this slice",
    "CopilotResponse schema change",
    "gateway route change",
    "platform HTTP route change",
    "artifact store implementation",
    "artifact binary reader implementation",
    "public URL resolver implementation",
    "backend adapter implementation",
    "real backend call",
    "image generation",
    "artifact upload",
    "executor implementation",
    "confirmation implementation",
    "writeback implementation",
    "replay implementation",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "docs/task-cards/image-artifact-response-builder-integration-v1-plan.md",
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "services/platform/internal/httpapi/image_artifacts.go",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactResponseBuilderPanel.tsx",
}
EXPECTED_ZERO_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "artifact_binary_read_count=0",
    "artifact_store_lookup_count=0",
    "runtime_mapping_execution_count=0",
    "production_storage_write_count=0",
    "public_url_resolution_count=0",
    "response_consumer_call_count=0",
    "response_builder_change_count=0",
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def rows_by_id(rows: list[Any], key: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for row in rows:
        require(isinstance(row, dict), f"{key} rows must be JSON objects")
        row_id = str(row.get(key) or "")
        require(row_id, f"{key} row missing id")
        result[row_id] = row
    return result


def artifact_fixture() -> dict[str, Any]:
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.validate(artifact, load_json(ARTIFACT_SCHEMA_PATH))
    return artifact


def copilot_response_schema() -> dict[str, Any]:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    properties = set((schema.get("properties") or {}).keys())
    require("citations" in properties, "CopilotResponse.citations missing")
    require("artifacts" not in properties, "CopilotResponse.artifacts must not be added")
    require("artifact_metadata" not in properties, "CopilotResponse.artifact_metadata must not be added")
    citation_schema = ((schema.get("$defs") or {}).get("citation") or {})
    kind_enum = ((((citation_schema.get("properties") or {}).get("kind") or {}).get("enum")) or [])
    require("artifact" in kind_enum, "CopilotResponse citation kind must allow artifact")
    return schema


def valid_response_document() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "image artifact response builder integration entry review",
        "summary": "Image artifact metadata can be cited without public payload exposure.",
        "answers": [{"text": "The generated artifact is available through an artifact citation."}],
        "issues": [],
        "proposed_actions": [],
        "citations": [
            {"id": "existing-rule", "kind": "rule", "label": "Existing rule citation"},
        ],
        "confidence": 0.82,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def source_path(surface: str) -> Path:
    path_text = surface.split("#", 1)[0]
    return REPO_ROOT / path_text


def successor_integration_task_card_allows(relative_path: str) -> bool:
    if relative_path != "docs/task-cards/image-artifact-response-builder-integration-v1-plan.md":
        return False
    if not SUCCESSOR_INTEGRATION_FIXTURE_PATH.exists():
        return False
    successor = load_json(SUCCESSOR_INTEGRATION_FIXTURE_PATH)
    return (
        (successor.get("slice") or {}).get("status")
        == "image_artifact_response_builder_integration_task_card_defined"
    )


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(fixture.get("kind") == "image_artifact_response_builder_integration_entry_review_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "image-artifact-response-builder-integration-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_response_builder_integration_entry_review_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")

    consumer_runtime = load_json(RESPONSE_CONSUMER_RUNTIME_PATH)
    require(
        (consumer_runtime.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_runtime_implemented",
        "response consumer runtime status drifted",
    )
    consumer_task = load_json(RESPONSE_CONSUMER_IMPLEMENTATION_PATH)
    require(
        (consumer_task.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_implementation_task_card_defined",
        "response consumer implementation task card status drifted",
    )
    runtime_mapper = load_json(RUNTIME_MAPPER_PATH)
    require(
        (runtime_mapper.get("slice") or {}).get("status")
        == "image_artifact_runtime_mapper_runtime_implemented",
        "runtime mapper status drifted",
    )


def assert_integration_entry_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("integration_entry_review") or {}
    require(review.get("status") == "response_builder_integration_entry_review_defined", "review status drifted")
    require(
        review.get("conclusion")
        == "future_connection_must_use_python_response_normalization_boundary_with_schema_validation",
        "review conclusion drifted",
    )
    require(
        review.get("selected_future_hook") == "services/runtime/inference_response.py#coerce_response_document",
        "selected future hook drifted",
    )
    for key in (
        "implementation_created_in_this_slice",
        "response_builder_changed_in_this_slice",
        "response_builder_connected_in_this_slice",
        "copilot_response_schema_change_allowed",
        "metadata_reference_public_output_allowed",
        "artifact_store_lookup_allowed",
        "artifact_binary_read_allowed",
        "public_url_resolution_allowed",
        "backend_call_allowed",
        "image_generation_allowed",
        "gateway_route_change_allowed",
        "platform_route_change_allowed",
    ):
        require(review.get(key) is False, f"integration_entry_review.{key} must remain false")

    candidates = rows_by_id(fixture.get("entry_candidate_matrix") or [], "candidate_id")
    require(set(candidates) == set(EXPECTED_ENTRY_CANDIDATES), "entry candidate matrix drifted")
    for candidate_id, (expected_decision, expected_surface) in EXPECTED_ENTRY_CANDIDATES.items():
        candidate = candidates[candidate_id]
        require(candidate.get("decision") == expected_decision, f"{candidate_id} decision drifted")
        require(candidate.get("surface") == expected_surface, f"{candidate_id} surface drifted")
        require(candidate.get("implementation_created") is False, f"{candidate_id} must not be implemented")
        require(source_path(expected_surface).exists(), f"{candidate_id} surface path missing")

    gates = rows_by_id(fixture.get("future_connection_gates") or [], "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "future connection gates drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")

    require(set(fixture.get("future_failure_propagation") or []) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    require(set(fixture.get("allowed_next_scope") or []) == EXPECTED_ALLOWED_NEXT_SCOPE, "allowed next scope drifted")
    require(set(fixture.get("forbidden_next_scope") or []) == EXPECTED_FORBIDDEN_NEXT_SCOPE, "forbidden next scope drifted")


def assert_mapper_consumer_chain_still_schema_valid() -> None:
    schema = copilot_response_schema()
    mapping = map_image_artifact_to_response_reference(artifact_fixture())
    require(mapping.ok is True and mapping.citation is not None, "runtime mapper success citation missing")
    require(mapping.metadata_reference is not None, "runtime mapper metadata reference missing")

    response = valid_response_document()
    original_response = copy.deepcopy(response)
    result = apply_image_artifact_reference_to_response(response, mapping)
    require(result.ok is True, f"response consumer success failed: {result.failure_code}")
    require(response == original_response, "response consumer must not mutate draft response")
    require(result.metadata_reference == mapping.metadata_reference, "metadata reference handoff drifted")
    require("metadata_reference" not in result.response_document, "metadata_reference leaked into response document")
    require("artifact_metadata" not in result.response_document, "artifact_metadata leaked into response document")
    require(
        [item["id"] for item in result.response_document["citations"]] == ["existing-rule", mapping.citation["id"]],
        "response consumer must append generated artifact citation after existing citations",
    )
    jsonschema.validate(result.response_document, schema)

    duplicate_response = valid_response_document()
    duplicate_response["citations"].append(copy.deepcopy(mapping.citation))
    duplicate = apply_image_artifact_reference_to_response(duplicate_response, mapping)
    require(duplicate.ok is False, "duplicate citation id must fail closed")
    require(
        duplicate.failure_code == "image_artifact_citation_id_conflict",
        "duplicate citation failure code drifted",
    )


def assert_forbidden_artifacts_and_no_source_connection(fixture: dict[str, Any]) -> None:
    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact matrix drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if row.get("allowed_next") is True:
            require(
                relative_path == "docs/task-cards/image-artifact-response-builder-integration-v1-plan.md",
                f"{relative_path} must not be allowed next",
            )
        else:
            require(row.get("allowed_next") is False, f"{relative_path} allowed_next drifted")
        if not successor_integration_task_card_allows(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    forbidden_literals = set(fixture.get("forbidden_source_literals") or [])
    require(forbidden_literals, "forbidden source literal list missing")
    for relative_path in fixture.get("source_files_that_must_not_be_connected") or []:
        source = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        leaked = sorted(literal for literal in forbidden_literals if literal in source)
        require(not leaked, f"{relative_path} must not be connected to image artifact response builder: {leaked}")


def assert_side_effects(fixture: dict[str, Any]) -> None:
    mapper_counters = runtime_mapper_side_effect_counters()
    consumer_counters = response_consumer_side_effect_counters()
    for key, value in mapper_counters.items():
        require(consumer_counters.get(key) == value == 0, f"{key} side-effect counter drifted")
    rendered = {f"{key}={value}" for key, value in consumer_counters.items()}
    require(rendered == EXPECTED_ZERO_COUNTERS, "side-effect counters drifted")
    require(set(fixture.get("side_effect_counters_must_remain") or []) == EXPECTED_ZERO_COUNTERS, "fixture counters drifted")


def assert_references_and_check_repo(fixture: dict[str, Any]) -> None:
    for relative_path in fixture.get("evidence") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing evidence path: {relative_path}")
    for relative_path in fixture.get("required_consumers") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing consumer path: {relative_path}")

    for relative_path, required_literals in (fixture.get("required_doc_references") or {}).items():
        text = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        missing = [literal for literal in required_literals if literal not in text]
        require(not missing, f"{relative_path} missing doc literals: {missing}")

    check_repo = CHECK_REPO_PATH.read_text(encoding="utf-8")
    previous_checker = "check-image-artifact-response-consumer-runtime-implementation-v1.py"
    current_checker = "check-image-artifact-response-builder-integration-entry-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run response builder integration entry review checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "response builder integration entry review checker must run after response consumer runtime checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_integration_entry_review(fixture)
    assert_mapper_consumer_chain_still_schema_valid()
    assert_forbidden_artifacts_and_no_source_connection(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response builder integration entry review v1 checks passed.")


if __name__ == "__main__":
    main()
