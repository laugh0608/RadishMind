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
    ImageArtifactResponseConsumerResult,
    apply_image_artifact_reference_to_response,
    response_consumer_side_effect_counters,
)
from services.runtime.image_artifact_runtime_mapper import (  # noqa: E402
    ImageArtifactMappingResult,
    map_image_artifact_to_response_reference,
    runtime_mapper_side_effect_counters,
)


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json"
IMPLEMENTATION_TASK_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-implementation-v1.json"
READINESS_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-response-consumer-implementation-readiness-v1.json"
)
RUNTIME_MAPPER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
RESPONSE_CONSUMER_MODULE = REPO_ROOT / "services/runtime/image_artifact_response_consumer.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-response-consumer-implementation-v1",
    "image-artifact-response-consumer-implementation-readiness-v1",
    "image-artifact-runtime-mapper-runtime-implementation-v1",
    "contracts/copilot-response.schema.json",
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/image_artifact_response_consumer.py",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "response_builder_changed",
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
    "copilot_response_schema_changed_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "metadata_reference_public_output_allowed",
    "store_reader_public_url_backend_parallel_tracks_allowed",
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
EXPECTED_FAILURE_CODES = {
    "image_artifact_mapper_failed",
    "image_artifact_citation_id_conflict",
    "image_artifact_citation_schema_invalid",
    "image_artifact_metadata_reference_leak",
}
EXPECTED_RUNTIME_TEST_CASES = {
    "append_generated_not_required_citation": (True, "merge_artifact_citation"),
    "append_generated_reviewed_pass_citation": (True, "merge_artifact_citation"),
    "reject_mapper_blocked_failure": (False, "reject_success_response"),
    "reject_duplicate_citation_id": (False, "reject_success_response"),
    "reject_schema_invalid_citation": (False, "reject_success_response"),
    "reject_metadata_reference_leak": (False, "reject_success_response"),
    "preserve_no_side_effects": (True, "preserve_no_side_effects"),
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/copilot_response_artifact_mapper.py",
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "contracts/image-artifact-response-reference.schema.json",
    "services/platform/internal/httpapi/image_artifacts.go",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactResponseConsumerPanel.tsx",
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
    citation = ((schema.get("$defs") or {}).get("citation") or {})
    kind_enum = ((((citation.get("properties") or {}).get("kind") or {}).get("enum")) or [])
    require("artifact" in kind_enum, "CopilotResponse citation kind must allow artifact")
    return schema


def valid_response_document() -> dict[str, Any]:
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "image artifact response consumer runtime check",
        "summary": "Image artifact metadata is ready.",
        "answers": [{"text": "The generated artifact is available as metadata-only citation."}],
        "issues": [],
        "proposed_actions": [],
        "citations": [],
        "confidence": 0.8,
        "risk_level": "low",
        "requires_confirmation": False,
    }


def reviewed_artifact(base: dict[str, Any]) -> dict[str, Any]:
    artifact = copy.deepcopy(base)
    artifact["safety"]["requires_confirmation"] = True
    artifact["safety"]["review_status"] = "reviewed_pass"
    return artifact


def blocked_artifact(base: dict[str, Any]) -> dict[str, Any]:
    artifact = copy.deepcopy(base)
    artifact["status"] = "blocked"
    artifact["safety"]["review_status"] = "blocked"
    return artifact


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(fixture.get("kind") == "image_artifact_response_consumer_runtime_implementation_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-response-consumer-runtime-implementation-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_response_consumer_runtime_implemented",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")
    require(RESPONSE_CONSUMER_MODULE.exists(), "response consumer module must exist")

    implementation_task = load_json(IMPLEMENTATION_TASK_PATH)
    require(
        (implementation_task.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_implementation_task_card_defined",
        "implementation task card status drifted",
    )
    readiness = load_json(READINESS_PATH)
    require(
        (readiness.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_implementation_readiness_defined",
        "response consumer readiness status drifted",
    )
    runtime_mapper = load_json(RUNTIME_MAPPER_PATH)
    require(
        (runtime_mapper.get("slice") or {}).get("status") == "image_artifact_runtime_mapper_runtime_implemented",
        "runtime mapper status drifted",
    )


def assert_runtime_contract(fixture: dict[str, Any]) -> None:
    implementation = fixture.get("runtime_implementation") or {}
    require(implementation.get("status") == "metadata_only_response_consumer_implemented", "runtime status drifted")
    require(implementation.get("module") == "services/runtime/image_artifact_response_consumer.py", "module drifted")
    require(
        implementation.get("function") == "apply_image_artifact_reference_to_response",
        "function drifted",
    )
    require(
        implementation.get("result_type") == "ImageArtifactResponseConsumerResult",
        "result type drifted",
    )
    require(implementation.get("response_consumer_created_in_this_slice") is True, "consumer must be created")
    require(implementation.get("output_surface") == "CopilotResponse.citations", "output surface drifted")
    require(implementation.get("target_citation_kind") == "artifact", "citation kind drifted")
    for field in EXPECTED_FALSE_FIELDS:
        require(implementation.get(field) is False, f"runtime_implementation.{field} must remain false")
    require(set(fixture.get("failure_taxonomy") or []) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")

    cases = rows_by_id(fixture.get("runtime_test_cases") or [], "case_id")
    require(set(cases) == set(EXPECTED_RUNTIME_TEST_CASES), "runtime test cases drifted")
    for case_id, (expected_ok, expected_action) in EXPECTED_RUNTIME_TEST_CASES.items():
        require(cases[case_id].get("expected_ok") is expected_ok, f"{case_id} ok drifted")
        require(cases[case_id].get("expected_consumer_action") == expected_action, f"{case_id} action drifted")


def assert_success_cases() -> None:
    schema = copilot_response_schema()
    base_artifact = artifact_fixture()
    for artifact in (base_artifact, reviewed_artifact(base_artifact)):
        mapping = map_image_artifact_to_response_reference(artifact)
        require(mapping.ok is True and mapping.citation is not None, "mapper success citation missing")
        require(mapping.metadata_reference is not None, "mapper success metadata reference missing")

        response = valid_response_document()
        original_response = copy.deepcopy(response)
        original_mapping_citation = copy.deepcopy(mapping.citation)
        result = apply_image_artifact_reference_to_response(response, mapping)

        require(isinstance(result, ImageArtifactResponseConsumerResult), "consumer result type drifted")
        require(result.ok is True, f"consumer success failed: {result.failure_code}")
        require(result.failure_code is None, "success must not carry failure code")
        require(result.metadata_reference == mapping.metadata_reference, "metadata reference handoff drifted")
        require(response == original_response, "consumer must not mutate input response")
        require(mapping.citation == original_mapping_citation, "consumer must not mutate mapper citation")
        require(len(result.response_document["citations"]) == 1, "artifact citation must be appended once")
        require(result.response_document["citations"][0] == mapping.citation, "appended citation drifted")
        require("metadata_reference" not in result.response_document, "metadata_reference leaked into response")
        require("artifact_metadata" not in result.response_document, "artifact_metadata leaked into response")
        jsonschema.validate(result.response_document, schema)

        response_with_existing = valid_response_document()
        response_with_existing["citations"].append(
            {"id": "existing-rule", "kind": "rule", "label": "Existing rule citation"}
        )
        merged = apply_image_artifact_reference_to_response(response_with_existing, mapping)
        require(merged.ok is True, "consumer must preserve existing citations")
        require(
            [item["id"] for item in merged.response_document["citations"]] == ["existing-rule", mapping.citation["id"]],
            "consumer must append artifact citation after existing citations",
        )
        jsonschema.validate(merged.response_document, schema)


def assert_failure_cases() -> None:
    base_artifact = artifact_fixture()
    blocked_mapping = map_image_artifact_to_response_reference(blocked_artifact(base_artifact))
    require(blocked_mapping.ok is False, "blocked artifact must fail in mapper")
    assert_failure_result(
        apply_image_artifact_reference_to_response(valid_response_document(), blocked_mapping),
        "image_artifact_mapper_failed",
        valid_response_document(),
    )

    success_mapping = map_image_artifact_to_response_reference(base_artifact)
    require(success_mapping.ok is True, "base artifact must map successfully")
    duplicate_response = valid_response_document()
    duplicate_response["citations"].append(copy.deepcopy(success_mapping.citation))
    assert_failure_result(
        apply_image_artifact_reference_to_response(duplicate_response, success_mapping),
        "image_artifact_citation_id_conflict",
        duplicate_response,
    )

    invalid_citation = copy.deepcopy(success_mapping.citation)
    invalid_citation.pop("locator")
    invalid_mapping = ImageArtifactMappingResult(
        ok=True,
        citation=invalid_citation,
        metadata_reference=success_mapping.metadata_reference,
    )
    assert_failure_result(
        apply_image_artifact_reference_to_response(valid_response_document(), invalid_mapping),
        "image_artifact_citation_schema_invalid",
        valid_response_document(),
    )

    leaked_metadata = copy.deepcopy(success_mapping.metadata_reference)
    leaked_metadata["public_url"] = "https://example.invalid/generated.png"
    leaked_mapping = ImageArtifactMappingResult(
        ok=True,
        citation=success_mapping.citation,
        metadata_reference=leaked_metadata,
    )
    assert_failure_result(
        apply_image_artifact_reference_to_response(valid_response_document(), leaked_mapping),
        "image_artifact_metadata_reference_leak",
        valid_response_document(),
    )

    response_leak = valid_response_document()
    response_leak["metadata_reference"] = {"artifact_id": "leaked"}
    result = apply_image_artifact_reference_to_response(response_leak, success_mapping)
    require(result.ok is False, "response metadata leak must fail")
    require(result.failure_code == "image_artifact_metadata_reference_leak", "response leak failure code drifted")
    require(result.response_document == response_leak, "response leak failure must preserve response draft")


def assert_failure_result(
    result: ImageArtifactResponseConsumerResult,
    expected_failure_code: str,
    expected_response: dict[str, Any],
) -> None:
    require(result.ok is False, f"{expected_failure_code} must fail")
    require(result.failure_code == expected_failure_code, f"{expected_failure_code} drifted: {result.failure_code}")
    require(result.metadata_reference is None, f"{expected_failure_code} must not return metadata reference")
    require(result.response_document == expected_response, f"{expected_failure_code} must preserve response draft")


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
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
        require(not leaked, f"{relative_path} must not be connected to image artifact consumer: {leaked}")


def assert_side_effects(fixture: dict[str, Any]) -> None:
    mapper_counters = runtime_mapper_side_effect_counters()
    consumer_counters = response_consumer_side_effect_counters()
    for key, value in mapper_counters.items():
        require(consumer_counters.get(key) == value == 0, f"{key} side-effect counter drifted")
    rendered = {f"{key}={value}" for key, value in consumer_counters.items()}
    require(rendered == EXPECTED_ZERO_COUNTERS, "consumer side-effect counters drifted")
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
    previous_checker = "check-image-artifact-response-consumer-implementation-v1.py"
    current_checker = "check-image-artifact-response-consumer-runtime-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run response consumer runtime implementation checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "response consumer runtime implementation checker must run after implementation checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_runtime_contract(fixture)
    assert_success_cases()
    assert_failure_cases()
    assert_forbidden_artifacts(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response consumer runtime implementation v1 checks passed.")


if __name__ == "__main__":
    main()
