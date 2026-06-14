#!/usr/bin/env python3
from __future__ import annotations

import copy
import json
import sys
from collections.abc import Mapping
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_artifact_response_consumer import (  # noqa: E402
    ImageArtifactResponseConsumerResult,
    apply_image_artifact_reference_to_response,
    response_consumer_side_effect_counters,
)
from services.runtime.image_artifact_runtime_mapper import (  # noqa: E402
    ImageArtifactMappingResult,
    map_image_artifact_to_response_reference,
)


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json"
ENTRY_REVIEW_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-entry-review-v1.json"
)
RESPONSE_CONSUMER_RUNTIME_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json"
)
RUNTIME_MAPPER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
TASK_CARD_PATH = REPO_ROOT / "docs/task-cards/image-artifact-response-builder-integration-v1-plan.md"
INFERENCE_RESPONSE_PATH = REPO_ROOT / "services/runtime/inference_response.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-response-builder-integration-entry-review-v1",
    "image-artifact-response-consumer-runtime-implementation-v1",
    "image-artifact-runtime-mapper-runtime-implementation-v1",
    "contracts/copilot-request.schema.json",
    "contracts/copilot-response.schema.json",
    "contracts/image-generation-artifact.schema.json",
    "services/runtime/inference_response.py",
    "services/runtime/image_artifact_response_consumer.py",
    "services/runtime/image_artifact_runtime_mapper.py",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "response_builder_changed",
    "response_builder_connected",
    "response_builder_runtime_implemented",
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
EXPECTED_FALSE_FIELDS = {
    "response_builder_changed_in_this_slice",
    "response_builder_connected_in_this_slice",
    "response_builder_runtime_code_allowed_in_this_slice",
    "copilot_response_schema_changed_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "metadata_reference_public_output_allowed",
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
EXPECTED_DISCOVERY_FORBIDDEN_SOURCES = {
    "response_document",
    "raw_text",
    "environment_variable",
    "artifact_store_lookup",
    "artifact_binary_reader",
    "public_url_resolver",
    "signed_url_resolver",
    "backend_adapter_response",
}
EXPECTED_FAILURE_CASES = {
    "missing_request_artifact_metadata",
    "mapper_failure",
    "consumer_citation_id_conflict",
    "consumer_citation_schema_invalid",
    "consumer_metadata_reference_leak",
    "post_merge_schema_validation_failure",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "services/runtime/image_artifact_response_builder_integration.py",
    "contracts/image-artifact-response-reference.schema.json",
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


def valid_response_document() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "image artifact response builder integration",
        "summary": "Image artifact metadata can be cited after response normalization.",
        "answers": [{"text": "The generated image artifact is available as metadata-only citation."}],
        "issues": [],
        "proposed_actions": [],
        "citations": [{"id": "existing-rule", "kind": "rule", "label": "Existing rule citation"}],
        "confidence": 0.83,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def request_with_image_artifact_metadata(artifact_document: dict[str, Any] | None) -> dict[str, Any]:
    metadata: dict[str, Any] = {}
    if artifact_document is not None:
        metadata["image_generation_artifact"] = copy.deepcopy(artifact_document)
    request = {
        "schema_version": 1,
        "request_id": "image-builder-integration-check",
        "project": "radishflow",
        "task": "inspect_canvas_snapshot",
        "locale": "zh-CN",
        "artifacts": [
            {
                "kind": "json",
                "role": "supporting",
                "name": "image_generation_artifact_metadata",
                "mime_type": "application/json",
                "content": {},
                "metadata": metadata,
            }
        ],
        "context": {},
        "tool_hints": {
            "allow_retrieval": False,
            "allow_tool_calls": False,
            "allow_image_reasoning": True,
        },
        "safety": {
            "mode": "advisory",
            "requires_confirmation_for_actions": True,
        },
    }
    jsonschema.validate(request, load_json(COPILOT_REQUEST_SCHEMA_PATH))
    return request


def discover_request_image_artifact_metadata(copilot_request: Mapping[str, Any]) -> list[dict[str, Any]]:
    candidates: list[dict[str, Any]] = []
    for artifact in copilot_request.get("artifacts") or []:
        if not isinstance(artifact, Mapping):
            continue
        metadata = artifact.get("metadata") or {}
        if not isinstance(metadata, Mapping):
            continue
        candidate = metadata.get("image_generation_artifact")
        if isinstance(candidate, dict):
            candidates.append(copy.deepcopy(candidate))
    return candidates


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(fixture.get("kind") == "image_artifact_response_builder_integration_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-response-builder-integration-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_response_builder_integration_task_card_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")
    require(TASK_CARD_PATH.exists(), "response builder integration task card must exist")

    entry_review = load_json(ENTRY_REVIEW_PATH)
    require(
        (entry_review.get("slice") or {}).get("status")
        == "image_artifact_response_builder_integration_entry_review_defined",
        "entry review status drifted",
    )
    response_consumer = load_json(RESPONSE_CONSUMER_RUNTIME_PATH)
    require(
        (response_consumer.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_runtime_implemented",
        "response consumer runtime status drifted",
    )
    runtime_mapper = load_json(RUNTIME_MAPPER_PATH)
    require(
        (runtime_mapper.get("slice") or {}).get("status") == "image_artifact_runtime_mapper_runtime_implemented",
        "runtime mapper status drifted",
    )


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("integration_task_card_boundary") or {}
    require(
        boundary.get("status") == "response_builder_integration_task_card_defined_no_runtime_code",
        "task card boundary status drifted",
    )
    require(
        boundary.get("decision") == "future_coerce_response_document_hook_defined",
        "task card boundary decision drifted",
    )
    require(boundary.get("task_card_created_in_this_slice") is True, "task card must be created")
    for field in EXPECTED_FALSE_FIELDS:
        require(boundary.get(field) is False, f"integration_task_card_boundary.{field} must remain false")


def assert_hook_placement(fixture: dict[str, Any]) -> None:
    hook = fixture.get("exact_hook_placement") or {}
    require(
        hook.get("function") == "services/runtime/inference_response.py#coerce_response_document",
        "hook function drifted",
    )
    require(
        hook.get("placement") == "after_response_top_level_filter_before_existing_schema_validation",
        "hook placement drifted",
    )
    require(hook.get("implementation_created_in_this_slice") is False, "hook must not be implemented")
    source = INFERENCE_RESPONSE_PATH.read_text(encoding="utf-8")
    function_marker = "def coerce_response_document("
    require(function_marker in source, "coerce_response_document missing")
    after_anchor = str(hook.get("after_anchor") or "")
    before_anchor = str(hook.get("before_anchor") or "")
    require(after_anchor in source, "after anchor missing from inference_response.py")
    require(before_anchor in source, "before anchor missing from inference_response.py")
    require(source.index(after_anchor) < source.index(before_anchor), "hook anchors are in the wrong order")
    require("normalize_citations_from_document" in source, "citation normalization anchor missing")
    require("RESPONSE_TOP_LEVEL_KEYS" in source, "top-level filtering anchor missing")


def assert_discovery_and_validation_plan(fixture: dict[str, Any]) -> None:
    discovery = fixture.get("request_artifact_metadata_discovery_input") or {}
    require(discovery.get("source") == "copilot_request", "discovery source drifted")
    require(
        discovery.get("path") == "artifacts[*].metadata.image_generation_artifact",
        "discovery path drifted",
    )
    require(discovery.get("schema") == "contracts/image-generation-artifact.schema.json", "discovery schema drifted")
    require(discovery.get("missing_metadata_policy") == "no_op_preserve_response", "missing metadata policy drifted")
    require(
        set(discovery.get("forbidden_sources") or []) == EXPECTED_DISCOVERY_FORBIDDEN_SOURCES,
        "discovery forbidden sources drifted",
    )

    plan = fixture.get("post_merge_validation_plan") or {}
    require(plan.get("mapper_function") == "map_image_artifact_to_response_reference", "mapper function drifted")
    require(plan.get("consumer_function") == "apply_image_artifact_reference_to_response", "consumer function drifted")
    require(plan.get("schema_validation") == "validate_response_document", "schema validation function drifted")
    require(plan.get("schema") == "contracts/copilot-response.schema.json", "response schema drifted")
    require(plan.get("post_merge_validation_required") is True, "post-merge validation must be required")
    require(
        plan.get("metadata_reference_public_output_allowed") is False,
        "metadata_reference public output must remain false",
    )


def assert_mapper_consumer_chain_from_request_metadata() -> None:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    request = request_with_image_artifact_metadata(artifact_fixture())
    discovered = discover_request_image_artifact_metadata(request)
    require(len(discovered) == 1, "request artifact metadata discovery must find exactly one candidate")
    jsonschema.validate(discovered[0], load_json(ARTIFACT_SCHEMA_PATH))

    mapping = map_image_artifact_to_response_reference(discovered[0])
    require(mapping.ok is True and mapping.citation is not None, "mapper success citation missing")
    response = valid_response_document()
    result = apply_image_artifact_reference_to_response(response, mapping)
    require(result.ok is True and result.response_document is not response, "consumer merge must return a new response")
    require("metadata_reference" not in result.response_document, "metadata_reference leaked into response")
    jsonschema.validate(result.response_document, schema)

    no_metadata = request_with_image_artifact_metadata(None)
    require(discover_request_image_artifact_metadata(no_metadata) == [], "missing request metadata must be no-op")


def assert_failure_propagation(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("failure_propagation") or [], "case_id")
    require(set(rows) == EXPECTED_FAILURE_CASES, "failure propagation matrix drifted")
    require(
        rows["missing_request_artifact_metadata"].get("expected_action") == "no_op_preserve_response",
        "missing metadata action drifted",
    )

    blocked = artifact_fixture()
    blocked["status"] = "blocked"
    blocked["safety"]["review_status"] = "blocked"
    mapper_failure = map_image_artifact_to_response_reference(blocked)
    require(mapper_failure.ok is False, "blocked artifact must fail in mapper")
    propagated = apply_image_artifact_reference_to_response(valid_response_document(), mapper_failure)
    require(propagated.ok is False, "mapper failure must propagate through consumer")
    require(propagated.failure_code == "image_artifact_mapper_failed", "mapper failure propagation drifted")

    success_mapping = map_image_artifact_to_response_reference(artifact_fixture())
    require(success_mapping.ok is True and success_mapping.citation is not None, "success mapping missing")
    duplicate_response = valid_response_document()
    duplicate_response["citations"].append(copy.deepcopy(success_mapping.citation))
    duplicate = apply_image_artifact_reference_to_response(duplicate_response, success_mapping)
    require(duplicate.ok is False, "duplicate citation id must fail closed")
    require(duplicate.failure_code == "image_artifact_citation_id_conflict", "duplicate failure code drifted")

    invalid_mapping = ImageArtifactMappingResult(
        ok=True,
        citation={"id": "bad-artifact", "kind": "artifact", "label": "bad"},
        metadata_reference=copy.deepcopy(success_mapping.metadata_reference),
    )
    invalid = apply_image_artifact_reference_to_response(valid_response_document(), invalid_mapping)
    require(invalid.ok is False, "invalid citation must fail closed")
    require(invalid.failure_code == "image_artifact_citation_schema_invalid", "invalid citation failure drifted")

    leak_mapping = ImageArtifactMappingResult(
        ok=True,
        citation=copy.deepcopy(success_mapping.citation),
        metadata_reference={"artifact_id": "image-artifact-id", "public_url": "https://example.invalid/image.png"},
    )
    leak = apply_image_artifact_reference_to_response(valid_response_document(), leak_mapping)
    require(leak.ok is False, "metadata reference leak must fail closed")
    require(leak.failure_code == "image_artifact_metadata_reference_leak", "metadata leak failure drifted")

    invalid_response = valid_response_document()
    invalid_response["confidence"] = 2.0
    merged = apply_image_artifact_reference_to_response(invalid_response, success_mapping)
    require(merged.ok is True, "consumer should not replace post-merge schema validation")
    try:
        jsonschema.validate(merged.response_document, load_json(COPILOT_RESPONSE_SCHEMA_PATH))
    except jsonschema.ValidationError:
        pass
    else:
        raise SystemExit("post-merge schema validation failure was not detected")
    require(
        rows["post_merge_schema_validation_failure"].get("expected_failure_code")
        == "image_artifact_response_schema_invalid",
        "post-merge failure code drifted",
    )


def assert_forbidden_artifacts_and_no_source_connection(fixture: dict[str, Any]) -> None:
    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact matrix drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    forbidden_literals = set(fixture.get("forbidden_source_literals") or [])
    require(forbidden_literals, "forbidden source literal list missing")
    for relative_path in fixture.get("source_files_that_must_not_be_connected") or []:
        source = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        leaked = sorted(literal for literal in forbidden_literals if literal in source)
        require(not leaked, f"{relative_path} must not be connected to image artifact response builder: {leaked}")


def assert_side_effects(fixture: dict[str, Any]) -> None:
    result = apply_image_artifact_reference_to_response(
        valid_response_document(),
        map_image_artifact_to_response_reference(artifact_fixture()),
    )
    require(isinstance(result, ImageArtifactResponseConsumerResult), "consumer result type drifted")
    counters = response_consumer_side_effect_counters()
    rendered = {f"{key}={value}" for key, value in counters.items()}
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
    previous_checker = "check-image-artifact-response-builder-integration-entry-review-v1.py"
    current_checker = "check-image-artifact-response-builder-integration-v1.py"
    require(current_checker in check_repo, "check-repo.py must run response builder integration checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "response builder integration checker must run after entry review checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_task_card_boundary(fixture)
    assert_hook_placement(fixture)
    assert_discovery_and_validation_plan(fixture)
    assert_mapper_consumer_chain_from_request_metadata()
    assert_failure_propagation(fixture)
    assert_forbidden_artifacts_and_no_source_connection(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response builder integration v1 checks passed.")


if __name__ == "__main__":
    main()
