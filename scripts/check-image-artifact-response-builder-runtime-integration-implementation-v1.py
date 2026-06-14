#!/usr/bin/env python3
from __future__ import annotations

import ast
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
    runtime_mapper_side_effect_counters,
)
from services.runtime.inference_response import coerce_response_document  # noqa: E402


FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json"
)
ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-entry-review-v1.json"
)
INTEGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json"
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
TASK_CARD_PATH = (
    REPO_ROOT / "docs/task-cards/image-artifact-response-builder-runtime-integration-implementation-v1-plan.md"
)
INFERENCE_RESPONSE_PATH = REPO_ROOT / "services/runtime/inference_response.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-response-builder-runtime-integration-entry-review-v1",
    "image-artifact-response-builder-integration-v1",
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
    "runtime_code_created_in_this_slice",
    "response_builder_changed_in_this_slice",
    "response_builder_connected_in_this_slice",
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
EXPECTED_PREDECESSORS = {
    "runtime_integration_entry_review": "image_artifact_response_builder_runtime_integration_entry_review_defined",
    "response_builder_integration_task_card": "image_artifact_response_builder_integration_task_card_defined",
    "response_consumer_runtime": "image_artifact_response_consumer_runtime_implemented",
    "runtime_mapper_runtime": "image_artifact_runtime_mapper_runtime_implemented",
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
    "artifact_metadata_schema_invalid",
    "mapper_failure",
    "consumer_citation_id_conflict",
    "consumer_citation_schema_invalid",
    "consumer_metadata_reference_leak",
    "post_merge_schema_validation_failure",
}
EXPECTED_TEST_CASES = {
    "single_artifact_success_appends_citation",
    "multiple_artifacts_preserve_request_order",
    "missing_metadata_no_op_preserves_response",
    "artifact_metadata_schema_invalid_fail_closed",
    "mapper_blocked_failure_fail_closed",
    "consumer_citation_id_conflict_fail_closed",
    "consumer_citation_schema_invalid_fail_closed",
    "consumer_metadata_reference_leak_fail_closed",
    "post_merge_schema_invalid_fail_closed",
    "no_retry_backend_upload_or_binary_side_effects",
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


def copied_artifact(*, artifact_id: str, uri_suffix: str) -> dict[str, Any]:
    artifact = artifact_fixture()
    artifact["artifact_id"] = artifact_id
    artifact["artifact"]["uri"] = f"artifact://radishmind/generated/{uri_suffix}"
    artifact["artifact"]["title"] = f"Generated image {artifact_id}"
    return artifact


def valid_response_document() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "image artifact response builder runtime integration implementation",
        "summary": "Image artifact metadata can be merged by a future response builder runtime hook.",
        "answers": [{"text": "The future runtime integration must append artifact citations deterministically."}],
        "issues": [],
        "proposed_actions": [],
        "citations": [{"id": "existing-rule", "kind": "rule", "label": "Existing rule citation"}],
        "confidence": 0.86,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def request_with_image_artifact_metadata(artifact_documents: list[dict[str, Any] | None]) -> dict[str, Any]:
    artifacts: list[dict[str, Any]] = []
    for index, artifact_document in enumerate(artifact_documents, start=1):
        metadata: dict[str, Any] = {}
        if artifact_document is not None:
            metadata["image_generation_artifact"] = copy.deepcopy(artifact_document)
        artifacts.append(
            {
                "kind": "json",
                "role": "supporting",
                "name": f"image_generation_artifact_metadata_{index}",
                "mime_type": "application/json",
                "content": {},
                "metadata": metadata,
            }
        )
    request = {
        "schema_version": 1,
        "request_id": "image-builder-runtime-integration-implementation-check",
        "project": "radishflow",
        "task": "inspect_canvas_snapshot",
        "locale": "zh-CN",
        "artifacts": artifacts,
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
    require(
        fixture.get("kind") == "image_artifact_response_builder_runtime_integration_implementation_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "image-artifact-response-builder-runtime-integration-implementation-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status")
        == "image_artifact_response_builder_runtime_integration_implementation_task_card_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")
    require(TASK_CARD_PATH.exists(), "runtime integration implementation task card must exist")

    predecessor_documents = {
        "runtime_integration_entry_review": load_json(ENTRY_REVIEW_PATH),
        "response_builder_integration_task_card": load_json(INTEGRATION_FIXTURE_PATH),
        "response_consumer_runtime": load_json(RESPONSE_CONSUMER_RUNTIME_PATH),
        "runtime_mapper_runtime": load_json(RUNTIME_MAPPER_PATH),
    }
    for gate_id, required_status in EXPECTED_PREDECESSORS.items():
        require(
            (predecessor_documents[gate_id].get("slice") or {}).get("status") == required_status,
            f"{gate_id} status drifted",
        )

    rows = rows_by_id(fixture.get("predecessor_review") or [], "gate_id")
    require(set(rows) == set(EXPECTED_PREDECESSORS), "predecessor review rows drifted")
    for gate_id, required_status in EXPECTED_PREDECESSORS.items():
        row = rows[gate_id]
        require(row.get("required_status") == required_status, f"{gate_id} required status drifted")
        require(row.get("satisfied") is True, f"{gate_id} must be satisfied")
        require((REPO_ROOT / str(row.get("evidence") or "")).exists(), f"{gate_id} evidence missing")


def assert_task_card_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_task_card_boundary") or {}
    require(
        boundary.get("status") == "runtime_integration_implementation_task_card_defined_no_runtime_code",
        "task card boundary status drifted",
    )
    require(
        boundary.get("decision") == "coerce_response_document_runtime_integration_code_allowed_next",
        "task card boundary decision drifted",
    )
    require(boundary.get("selected_track") == "single_python_response_builder_hook", "selected track drifted")
    require(boundary.get("task_card_created_in_this_slice") is True, "task card must be created")
    require(boundary.get("runtime_code_allowed_after_this_slice") is True, "runtime code must be allowed next")
    for field in EXPECTED_FALSE_FIELDS:
        require(boundary.get(field) is False, f"implementation_task_card_boundary.{field} must remain false")


def assert_future_runtime_hook_contract(fixture: dict[str, Any]) -> None:
    hook = fixture.get("future_runtime_hook_contract") or {}
    require(hook.get("target_file") == "services/runtime/inference_response.py", "target file drifted")
    require(
        hook.get("function") == "services/runtime/inference_response.py#coerce_response_document",
        "function drifted",
    )
    require(hook.get("function_signature") == ["document", "copilot_request", "raw_text"], "signature drifted")
    require(
        hook.get("placement") == "after_response_top_level_filter_before_existing_schema_validation",
        "hook placement drifted",
    )
    require(
        hook.get("allowed_future_source_change") == "single_hook_inside_coerce_response_document",
        "allowed future source change drifted",
    )
    for field in (
        "signature_change_allowed",
        "gateway_route_allowed",
        "platform_route_allowed",
        "inference_support_build_citations_allowed",
        "implementation_created_in_this_slice",
    ):
        require(hook.get(field) is False, f"future_runtime_hook_contract.{field} must remain false")

    source = INFERENCE_RESPONSE_PATH.read_text(encoding="utf-8")
    tree = ast.parse(source)
    function = next((node for node in tree.body if isinstance(node, ast.FunctionDef) and node.name == "coerce_response_document"), None)
    require(function is not None, "coerce_response_document missing")
    require([arg.arg for arg in function.args.args] == hook.get("function_signature"), "function signature drifted")
    after_anchor = str(hook.get("after_anchor") or "")
    before_anchor = str(hook.get("before_anchor") or "")
    require(after_anchor in source, "after anchor missing from inference_response.py")
    require(before_anchor in source, "before anchor missing from inference_response.py")
    require(source.index(after_anchor) < source.index(before_anchor), "hook anchors are in the wrong order")


def assert_metadata_discovery_and_merge_pipeline(fixture: dict[str, Any]) -> None:
    discovery = fixture.get("metadata_discovery_contract") or {}
    require(discovery.get("source") == "copilot_request", "discovery source drifted")
    require(discovery.get("path") == "artifacts[*].metadata.image_generation_artifact", "discovery path drifted")
    require(discovery.get("ordering") == "copilot_request.artifacts order", "discovery ordering drifted")
    require(discovery.get("schema") == "contracts/image-generation-artifact.schema.json", "discovery schema drifted")
    require(discovery.get("missing_metadata_policy") == "no_op_preserve_response", "missing metadata policy drifted")
    require(
        discovery.get("schema_invalid_failure_code") == "image_artifact_metadata_schema_invalid",
        "schema invalid failure code drifted",
    )
    require(
        set(discovery.get("forbidden_sources") or []) == EXPECTED_DISCOVERY_FORBIDDEN_SOURCES,
        "discovery forbidden sources drifted",
    )

    pipeline = fixture.get("merge_pipeline_contract") or {}
    require(pipeline.get("mapper_function") == "map_image_artifact_to_response_reference", "mapper function drifted")
    require(
        pipeline.get("consumer_function") == "apply_image_artifact_reference_to_response",
        "consumer function drifted",
    )
    require(pipeline.get("post_merge_validation") == "validate_response_document", "post-merge validation drifted")
    require(pipeline.get("post_merge_schema") == "contracts/copilot-response.schema.json", "post-merge schema drifted")
    require(pipeline.get("multiple_metadata_policy") == "merge_in_request_artifact_order", "multiple metadata policy drifted")
    require(pipeline.get("partial_success_allowed") is False, "partial success must remain false")
    require(pipeline.get("metadata_reference_public_output_allowed") is False, "metadata reference must remain internal")

    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    first = copied_artifact(artifact_id="image-artifact-first", uri_suffix="first")
    second = copied_artifact(artifact_id="image-artifact-second", uri_suffix="second")
    request = request_with_image_artifact_metadata([first, None, second])
    discovered = discover_request_image_artifact_metadata(request)
    require([item["artifact_id"] for item in discovered] == ["image-artifact-first", "image-artifact-second"], "metadata discovery order drifted")

    response = valid_response_document()
    for artifact_document in discovered:
        jsonschema.validate(artifact_document, load_json(ARTIFACT_SCHEMA_PATH))
        mapping = map_image_artifact_to_response_reference(artifact_document)
        require(mapping.ok is True and mapping.citation is not None, "mapper success citation missing")
        result = apply_image_artifact_reference_to_response(response, mapping)
        require(isinstance(result, ImageArtifactResponseConsumerResult), "consumer result type drifted")
        require(result.ok is True, f"consumer success failed: {result.failure_code}")
        response = result.response_document
    jsonschema.validate(response, schema)
    require(
        [citation["id"] for citation in response["citations"]]
        == ["existing-rule", "image-artifact-first", "image-artifact-second"],
        "artifact citation ordering drifted",
    )

    current_builder_response = coerce_response_document(valid_response_document(), request, raw_text="{}")
    jsonschema.validate(current_builder_response, schema)
    require(
        all(citation.get("kind") != "artifact" for citation in current_builder_response.get("citations") or []),
        "current response builder must not already append image artifact citations",
    )


def assert_failure_propagation_and_test_plan(fixture: dict[str, Any]) -> None:
    failure_rows = rows_by_id(fixture.get("failure_propagation_required") or [], "case_id")
    require(set(failure_rows) == EXPECTED_FAILURE_CASES, "failure propagation cases drifted")
    require(
        failure_rows["artifact_metadata_schema_invalid"].get("expected_failure_code")
        == "image_artifact_metadata_schema_invalid",
        "metadata schema invalid failure code drifted",
    )

    test_rows = rows_by_id(fixture.get("future_runtime_test_plan") or [], "case_id")
    require(set(test_rows) == EXPECTED_TEST_CASES, "future runtime test plan drifted")
    for case_id, row in test_rows.items():
        require(row.get("runtime_test_required_next") is True, f"{case_id} must require runtime test next")

    invalid_metadata = artifact_fixture()
    invalid_metadata.pop("artifact")
    try:
        jsonschema.validate(invalid_metadata, load_json(ARTIFACT_SCHEMA_PATH))
    except jsonschema.ValidationError:
        pass
    else:
        raise SystemExit("invalid image artifact metadata should fail schema validation")

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
    require(merged.ok is True, "consumer should not replace response builder post-merge validation")
    try:
        jsonschema.validate(merged.response_document, load_json(COPILOT_RESPONSE_SCHEMA_PATH))
    except jsonschema.ValidationError:
        pass
    else:
        raise SystemExit("post-merge schema validation failure was not detected")


def assert_response_schema_unchanged() -> None:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    properties = set((schema.get("properties") or {}).keys())
    require("citations" in properties, "CopilotResponse.citations missing")
    require("artifacts" not in properties, "CopilotResponse.artifacts must not be added")
    require("artifact_metadata" not in properties, "CopilotResponse.artifact_metadata must not be added")
    citation = (schema.get("$defs") or {}).get("citation") or {}
    kind_enum = ((((citation.get("properties") or {}).get("kind") or {}).get("enum")) or [])
    require("artifact" in kind_enum, "CopilotResponse citation kind must continue allowing artifact")


def assert_forbidden_artifacts_and_no_source_connection(fixture: dict[str, Any]) -> None:
    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact matrix drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(row.get("allowed_next") is False, f"{relative_path} must not be allowed next")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    forbidden_literals = set(fixture.get("forbidden_source_literals") or [])
    require(forbidden_literals, "forbidden source literal list missing")
    for relative_path in fixture.get("source_files_that_must_not_be_connected") or []:
        source = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        leaked = sorted(literal for literal in forbidden_literals if literal in source)
        require(not leaked, f"{relative_path} must not be connected in task-card-only slice: {leaked}")


def assert_side_effects(fixture: dict[str, Any]) -> None:
    mapper_counters = runtime_mapper_side_effect_counters()
    consumer_counters = response_consumer_side_effect_counters()
    require(all(value == 0 for value in mapper_counters.values()), "runtime mapper side-effect counters must remain zero")
    require(all(value == 0 for value in consumer_counters.values()), "response consumer side-effect counters must remain zero")
    rendered_consumer = {f"{key}={value}" for key, value in consumer_counters.items()}
    require(rendered_consumer == EXPECTED_ZERO_COUNTERS, "response consumer side-effect counter surface drifted")
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
    previous_checker = "check-image-artifact-response-builder-runtime-integration-entry-review-v1.py"
    current_checker = "check-image-artifact-response-builder-runtime-integration-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime integration implementation checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime integration implementation checker must run after entry review checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_task_card_boundary(fixture)
    assert_future_runtime_hook_contract(fixture)
    assert_metadata_discovery_and_merge_pipeline(fixture)
    assert_failure_propagation_and_test_plan(fixture)
    assert_response_schema_unchanged()
    assert_forbidden_artifacts_and_no_source_connection(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response builder runtime integration implementation v1 checks passed.")


if __name__ == "__main__":
    main()
