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
    / "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-entry-review-v1.json"
)
INTEGRATION_FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-v1.json"
)
INTEGRATION_ENTRY_REVIEW_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-builder-integration-entry-review-v1.json"
)
RESPONSE_CONSUMER_RUNTIME_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json"
)
RUNTIME_MAPPER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
SUCCESSOR_IMPLEMENTATION_FIXTURE_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json"
)
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
TASK_CARD_PATH = (
    REPO_ROOT / "docs/task-cards/image-artifact-response-builder-runtime-integration-entry-review-v1-plan.md"
)
INFERENCE_RESPONSE_PATH = REPO_ROOT / "services/runtime/inference_response.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-response-builder-integration-v1",
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
    "response_builder_runtime_integration_task_card_created",
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
    "implementation_task_card_created_in_this_slice",
    "runtime_code_allowed_in_this_slice",
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
EXPECTED_PREREQUISITES = {
    "runtime_mapper_runtime": "image_artifact_runtime_mapper_runtime_implemented",
    "response_consumer_runtime": "image_artifact_response_consumer_runtime_implemented",
    "response_builder_integration_task_card": "image_artifact_response_builder_integration_task_card_defined",
    "response_builder_integration_entry_review": "image_artifact_response_builder_integration_entry_review_defined",
}
EXPECTED_ALLOWED_NEXT_SCOPE = {
    "create image-artifact-response-builder-runtime-integration-implementation-v1 task card",
    "define coerce_response_document runtime integration implementation tests",
    "define request artifact metadata discovery helper placement",
    "define mapper and consumer failure propagation into response builder failure path",
    "define post-merge CopilotResponse schema validation in runtime implementation",
}
EXPECTED_FORBIDDEN_NEXT_SCOPE = {
    "implement runtime code in this slice",
    "change CopilotResponse schema",
    "change gateway route",
    "change platform HTTP route",
    "implement artifact store",
    "implement artifact binary reader",
    "implement public URL resolver",
    "implement signed URL resolver",
    "implement backend adapter",
    "perform real backend call",
    "generate image pixels",
    "upload artifact",
    "read artifact binary",
    "implement executor",
    "implement confirmation",
    "implement writeback",
    "implement replay",
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
    "docs/task-cards/image-artifact-response-builder-runtime-integration-implementation-v1-plan.md",
    "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json",
    "scripts/check-image-artifact-response-builder-runtime-integration-implementation-v1.py",
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "services/runtime/image_artifact_response_builder_integration.py",
    "contracts/image-artifact-response-reference.schema.json",
    "services/platform/internal/httpapi/image_artifacts.go",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactResponseBuilderPanel.tsx",
}
EXPECTED_ALLOWED_NEXT_ARTIFACTS = {
    "docs/task-cards/image-artifact-response-builder-runtime-integration-implementation-v1-plan.md",
    "scripts/checks/fixtures/image-artifact-response-builder-runtime-integration-implementation-v1.json",
    "scripts/check-image-artifact-response-builder-runtime-integration-implementation-v1.py",
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
        "task": "image artifact response builder runtime integration entry review",
        "summary": "Image artifact metadata can be connected by a future response builder task.",
        "answers": [{"text": "The response builder entry review keeps runtime integration deferred."}],
        "issues": [],
        "proposed_actions": [],
        "citations": [{"id": "existing-rule", "kind": "rule", "label": "Existing rule citation"}],
        "confidence": 0.84,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def request_with_image_artifact_metadata(artifact_document: dict[str, Any] | None) -> dict[str, Any]:
    metadata: dict[str, Any] = {}
    if artifact_document is not None:
        metadata["image_generation_artifact"] = copy.deepcopy(artifact_document)
    request = {
        "schema_version": 1,
        "request_id": "image-builder-runtime-integration-entry-review-check",
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


def successor_task_card_allows(relative_path: str) -> bool:
    if relative_path not in EXPECTED_ALLOWED_NEXT_ARTIFACTS:
        return False
    if not SUCCESSOR_IMPLEMENTATION_FIXTURE_PATH.exists():
        return False
    successor = load_json(SUCCESSOR_IMPLEMENTATION_FIXTURE_PATH)
    return (
        (successor.get("slice") or {}).get("status")
        in {
            "image_artifact_response_builder_runtime_integration_implementation_task_card_defined",
            "image_artifact_response_builder_runtime_integration_implemented",
        }
    )


def successor_runtime_integration_implemented() -> bool:
    if not SUCCESSOR_IMPLEMENTATION_FIXTURE_PATH.exists():
        return False
    successor = load_json(SUCCESSOR_IMPLEMENTATION_FIXTURE_PATH)
    return (
        (successor.get("slice") or {}).get("status")
        == "image_artifact_response_builder_runtime_integration_implemented"
    )


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "image_artifact_response_builder_runtime_integration_entry_review_v1",
        "kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "image-artifact-response-builder-runtime-integration-entry-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_response_builder_runtime_integration_entry_review_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")
    require(TASK_CARD_PATH.exists(), "runtime integration entry review task card must exist")

    prerequisites = {
        "runtime_mapper_runtime": load_json(RUNTIME_MAPPER_PATH),
        "response_consumer_runtime": load_json(RESPONSE_CONSUMER_RUNTIME_PATH),
        "response_builder_integration_task_card": load_json(INTEGRATION_FIXTURE_PATH),
        "response_builder_integration_entry_review": load_json(INTEGRATION_ENTRY_REVIEW_PATH),
    }
    for gate_id, required_status in EXPECTED_PREREQUISITES.items():
        require(
            (prerequisites[gate_id].get("slice") or {}).get("status") == required_status,
            f"{gate_id} status drifted",
        )

    prerequisite_rows = rows_by_id(fixture.get("prerequisite_review") or [], "gate_id")
    require(set(prerequisite_rows) == set(EXPECTED_PREREQUISITES), "prerequisite review rows drifted")
    for gate_id, required_status in EXPECTED_PREREQUISITES.items():
        row = prerequisite_rows[gate_id]
        require(row.get("required_status") == required_status, f"{gate_id} required status drifted")
        require(row.get("satisfied") is True, f"{gate_id} must be satisfied")
        require((REPO_ROOT / str(row.get("evidence") or "")).exists(), f"{gate_id} evidence missing")


def assert_entry_review_decision(fixture: dict[str, Any]) -> None:
    decision = fixture.get("entry_review_decision") or {}
    require(
        decision.get("status") == "runtime_integration_entry_review_defined_no_source_change",
        "entry review status drifted",
    )
    require(
        decision.get("decision") == "allow_single_runtime_integration_implementation_task_card_next",
        "entry review decision drifted",
    )
    require(
        decision.get("next_slice_allowed") == "image-artifact-response-builder-runtime-integration-implementation-v1",
        "next slice drifted",
    )
    require(decision.get("implementation_task_card_allowed_next") is True, "implementation task card must be allowed next")
    for field in EXPECTED_FALSE_FIELDS:
        require(decision.get(field) is False, f"entry_review_decision.{field} must remain false")

    boundary = fixture.get("single_runtime_task_boundary") or {}
    require(
        boundary.get("allowed_next_task_card")
        == "docs/task-cards/image-artifact-response-builder-runtime-integration-implementation-v1-plan.md",
        "allowed next task card drifted",
    )
    require(set(boundary.get("allowed_next_scope") or []) == EXPECTED_ALLOWED_NEXT_SCOPE, "allowed scope drifted")
    require(
        set(boundary.get("forbidden_next_scope") or []) == EXPECTED_FORBIDDEN_NEXT_SCOPE,
        "forbidden next scope drifted",
    )


def assert_future_runtime_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("future_runtime_connection_contract") or {}
    require(
        contract.get("function") == "services/runtime/inference_response.py#coerce_response_document",
        "runtime integration function drifted",
    )
    require(
        contract.get("function_signature") == ["document", "copilot_request", "raw_text"],
        "function signature contract drifted",
    )
    require(
        contract.get("placement") == "after_response_top_level_filter_before_existing_schema_validation",
        "hook placement drifted",
    )
    require(
        contract.get("request_metadata_source") == "copilot_request.artifacts[*].metadata.image_generation_artifact",
        "request metadata source drifted",
    )
    require(contract.get("missing_metadata_policy") == "no_op_preserve_response", "missing metadata policy drifted")
    require(contract.get("metadata_schema") == "contracts/image-generation-artifact.schema.json", "metadata schema drifted")
    require(contract.get("mapper_function") == "map_image_artifact_to_response_reference", "mapper function drifted")
    require(
        contract.get("consumer_function") == "apply_image_artifact_reference_to_response",
        "consumer function drifted",
    )
    require(contract.get("post_merge_validation") == "validate_response_document", "post-merge validation drifted")
    require(contract.get("post_merge_schema") == "contracts/copilot-response.schema.json", "post-merge schema drifted")
    for field in (
        "signature_change_allowed",
        "implementation_created_in_this_slice",
        "gateway_route_allowed",
        "platform_route_allowed",
        "inference_support_build_citations_allowed",
    ):
        require(contract.get(field) is False, f"future_runtime_connection_contract.{field} must remain false")

    source = INFERENCE_RESPONSE_PATH.read_text(encoding="utf-8")
    tree = ast.parse(source)
    function = next((node for node in tree.body if isinstance(node, ast.FunctionDef) and node.name == "coerce_response_document"), None)
    require(function is not None, "coerce_response_document missing")
    require(
        [arg.arg for arg in function.args.args] == contract.get("function_signature"),
        "coerce_response_document signature drifted",
    )
    after_anchor = str(contract.get("after_anchor") or "")
    before_anchor = str(contract.get("before_anchor") or "")
    require(after_anchor in source, "future hook after anchor missing")
    require(before_anchor in source, "future hook before anchor missing")
    require(source.index(after_anchor) < source.index(before_anchor), "future hook anchors are in the wrong order")


def assert_mapper_consumer_chain_and_current_noop() -> None:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    request = request_with_image_artifact_metadata(artifact_fixture())
    discovered = discover_request_image_artifact_metadata(request)
    require(len(discovered) == 1, "request metadata discovery must find exactly one image artifact")
    jsonschema.validate(discovered[0], load_json(ARTIFACT_SCHEMA_PATH))

    mapping = map_image_artifact_to_response_reference(discovered[0])
    require(mapping.ok is True and mapping.citation is not None, "mapper success citation missing")
    response = valid_response_document()
    consumer_result = apply_image_artifact_reference_to_response(response, mapping)
    require(isinstance(consumer_result, ImageArtifactResponseConsumerResult), "consumer result type drifted")
    require(consumer_result.ok is True and consumer_result.response_document is not response, "consumer success merge drifted")
    require("metadata_reference" not in consumer_result.response_document, "metadata reference leaked into response")
    jsonschema.validate(consumer_result.response_document, schema)

    coerced = coerce_response_document(valid_response_document(), request, raw_text="{}")
    jsonschema.validate(coerced, schema)
    artifact_citation_ids = {
        str(citation.get("id") or "")
        for citation in coerced.get("citations") or []
        if isinstance(citation, dict) and citation.get("kind") == "artifact"
    }
    expected_citation_id = str(mapping.citation.get("id") or "")
    if successor_runtime_integration_implemented():
        require(
            expected_citation_id in artifact_citation_ids,
            "implemented response builder must connect image artifact metadata",
        )
    else:
        require(
            expected_citation_id not in artifact_citation_ids,
            "current response builder must not already connect image artifact metadata",
        )

    no_metadata = request_with_image_artifact_metadata(None)
    require(discover_request_image_artifact_metadata(no_metadata) == [], "missing metadata must remain no-op")


def assert_failure_propagation(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("failure_propagation_required") or [], "case_id")
    require(set(rows) == EXPECTED_FAILURE_CASES, "failure propagation cases drifted")
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
    require(merged.ok is True, "consumer should leave post-merge schema validation to response builder")
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
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created in this slice")
        require(
            bool(row.get("allowed_next")) is (relative_path in EXPECTED_ALLOWED_NEXT_ARTIFACTS),
            f"{relative_path} allowed_next drifted",
        )
        if not successor_task_card_allows(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist yet")

    forbidden_literals = set(fixture.get("forbidden_source_literals") or [])
    require(forbidden_literals, "forbidden source literal list missing")
    for relative_path in fixture.get("source_files_that_must_not_be_connected") or []:
        source = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        leaked = sorted(literal for literal in forbidden_literals if literal in source)
        require(not leaked, f"{relative_path} must not be connected to image artifact runtime integration: {leaked}")


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
    previous_checker = "check-image-artifact-response-builder-integration-v1.py"
    current_checker = "check-image-artifact-response-builder-runtime-integration-entry-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime integration entry review checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime integration entry review checker must run after response builder integration checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_entry_review_decision(fixture)
    assert_future_runtime_contract(fixture)
    assert_mapper_consumer_chain_and_current_noop()
    assert_failure_propagation(fixture)
    assert_forbidden_artifacts_and_no_source_connection(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response builder runtime integration entry review v1 checks passed.")


if __name__ == "__main__":
    main()
