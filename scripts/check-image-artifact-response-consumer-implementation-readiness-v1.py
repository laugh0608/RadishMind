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

from services.runtime.image_artifact_runtime_mapper import (  # noqa: E402
    map_image_artifact_to_response_reference,
    runtime_mapper_side_effect_counters,
)


FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-implementation-readiness-v1.json"
)
INTEGRATION_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-runtime-mapper-response-consumer-integration-review-v1.json"
)
RUNTIME_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-runtime-mapper-response-consumer-integration-review-v1",
    "image-artifact-runtime-mapper-runtime-implementation-v1",
    "contracts/copilot-response.schema.json",
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/inference_response.py",
    "services/runtime/inference_support.py",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "response_consumer_implemented",
    "response_consumer_task_card_created",
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
    "implementation_created_in_this_slice",
    "task_card_created_in_this_slice",
    "response_builder_changed_in_this_slice",
    "copilot_response_schema_change_allowed",
    "metadata_reference_public_output_allowed",
    "artifact_store_lookup_allowed",
    "artifact_binary_read_allowed",
    "public_url_resolution_allowed",
    "backend_call_allowed",
    "image_generation_allowed",
}
EXPECTED_GATES = {
    "integration_review_consumed": "satisfied",
    "runtime_mapper_implementation_consumed": "satisfied",
    "copilot_response_citation_schema_preserved": "satisfied",
    "future_consumer_function_contract_defined": "satisfied",
    "failure_propagation_test_plan_defined": "satisfied",
    "existing_response_builder_unmodified": "satisfied",
    "response_consumer_implementation_task_card": "deferred",
    "artifact_store_binary_reader_backend_adapter": "blocked",
}
EXPECTED_ALLOWED_NEXT_SCOPE = {
    "create image-artifact-response-consumer-implementation-v1 task card",
    "implement metadata-only response consumer",
    "merge artifact citation into CopilotResponse.citations",
    "keep metadata_reference internal",
    "add success and fail-closed unit tests",
    "add no side effects smoke",
}
EXPECTED_FORBIDDEN_NEXT_SCOPE = {
    "CopilotResponse schema change",
    "artifact store implementation",
    "artifact binary reader implementation",
    "public URL resolver implementation",
    "backend adapter implementation",
    "real backend call",
    "image generation",
    "artifact upload",
    "UI implementation",
    "platform HTTP route implementation",
    "executor implementation",
    "confirmation implementation",
    "writeback implementation",
    "replay implementation",
}
EXPECTED_TEST_CASES = {
    "append_generated_not_required_citation",
    "append_generated_reviewed_pass_citation",
    "reject_blocked_artifact_status",
    "reject_failed_artifact_status",
    "reject_pending_review_artifact",
    "reject_public_url_claim",
    "reject_binary_payload_present",
    "reject_provider_raw_dump_present",
    "reject_duplicate_citation_id",
    "reject_metadata_reference_leak",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "docs/task-cards/image-artifact-response-consumer-implementation-v1-plan.md",
    "services/runtime/image_artifact_response_consumer.py",
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
    "executor_call_count=0",
    "confirmation_call_count=0",
    "business_writeback_count=0",
    "replay_call_count=0",
    "response_consumer_call_count=0",
    "response_builder_change_count=0",
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


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(fixture.get("kind") == "image_artifact_response_consumer_implementation_readiness_v1", "kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-response-consumer-implementation-readiness-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_response_consumer_implementation_readiness_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependencies drifted")

    integration = load_json(INTEGRATION_REVIEW_PATH)
    require(
        (integration.get("slice") or {}).get("status")
        == "image_artifact_runtime_mapper_response_consumer_integration_review_defined",
        "integration review status drifted",
    )
    runtime = load_json(RUNTIME_IMPLEMENTATION_PATH)
    require(
        (runtime.get("slice") or {}).get("status") == "image_artifact_runtime_mapper_runtime_implemented",
        "runtime mapper status drifted",
    )


def assert_readiness_contract(fixture: dict[str, Any]) -> None:
    readiness = fixture.get("implementation_readiness") or {}
    require(readiness.get("status") == "response_consumer_implementation_readiness_defined", "readiness status drifted")
    require(readiness.get("selected_track") == "metadata_only_response_consumer", "selected track drifted")
    require(readiness.get("future_module") == "services/runtime/image_artifact_response_consumer.py", "future module drifted")
    require(
        readiness.get("future_function") == "apply_image_artifact_reference_to_response",
        "future function drifted",
    )
    require(readiness.get("future_result_type") == "ImageArtifactResponseConsumerResult", "future result type drifted")
    for field in EXPECTED_FALSE_FIELDS:
        require(readiness.get(field) is False, f"implementation_readiness.{field} must remain false")

    contract = fixture.get("future_function_contract") or {}
    for field in (
        "input_response_document",
        "input_mapping_result",
        "success_output",
        "internal_output",
        "failure_output",
        "citation_id_conflict_failure_code",
        "schema_failure_code",
        "metadata_reference_leak_failure_code",
    ):
        require(str(contract.get(field) or ""), f"future_function_contract.{field} missing")
    require(
        contract.get("citation_id_conflict_failure_code") == "image_artifact_citation_id_conflict",
        "citation conflict failure code drifted",
    )
    require(
        contract.get("metadata_reference_leak_failure_code") == "image_artifact_metadata_reference_leak",
        "metadata leak failure code drifted",
    )

    gates = rows_by_id(fixture.get("readiness_gate_matrix") or [], "gate_id")
    require(set(gates) == set(EXPECTED_GATES), "readiness gates drifted")
    for gate_id, status in EXPECTED_GATES.items():
        require(gates[gate_id].get("status") == status, f"{gate_id} status drifted")

    require(set(fixture.get("allowed_next_scope") or []) == EXPECTED_ALLOWED_NEXT_SCOPE, "allowed next scope drifted")
    require(set(fixture.get("forbidden_next_scope") or []) == EXPECTED_FORBIDDEN_NEXT_SCOPE, "forbidden next scope drifted")


def assert_future_test_plan(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("future_test_plan") or [], "case_id")
    require(set(rows) == EXPECTED_TEST_CASES, "future test plan drifted")
    base = artifact_fixture()

    success = map_image_artifact_to_response_reference(base)
    require(success.ok is True and success.citation is not None, "generated_not_required must support future success test")
    jsonschema.validate(success.citation, citation_schema())
    require_no_forbidden_payload(success.metadata_reference or {}, "generated_not_required")

    reviewed = copy.deepcopy(base)
    reviewed["safety"]["requires_confirmation"] = True
    reviewed["safety"]["review_status"] = "reviewed_pass"
    reviewed_success = map_image_artifact_to_response_reference(reviewed)
    require(reviewed_success.ok is True and reviewed_success.citation is not None, "reviewed_pass must support future success test")
    jsonschema.validate(reviewed_success.citation, citation_schema())

    blocked = copy.deepcopy(base)
    blocked["status"] = "blocked"
    blocked["safety"]["review_status"] = "blocked"
    assert_mapper_failure(blocked, "image_artifact_safety_blocked", "blocked")

    failed = copy.deepcopy(base)
    failed["status"] = "failed"
    assert_mapper_failure(failed, "image_artifact_invalid_metadata", "failed")

    pending = copy.deepcopy(base)
    pending["safety"]["review_status"] = "pending_review"
    assert_mapper_failure(pending, "image_artifact_safety_pending_review", "pending_review")

    public_url = copy.deepcopy(base)
    public_url["artifact"]["uri"] = "https://example.invalid/image.png"
    assert_mapper_failure(public_url, "image_artifact_public_url_claim", "public_url")

    binary = copy.deepcopy(base)
    binary["base64_image"] = "not-real-binary"
    assert_mapper_failure(binary, "image_artifact_binary_payload_rejected", "binary")

    raw = copy.deepcopy(base)
    raw["provider_raw_response"] = {"raw": True}
    assert_mapper_failure(raw, "image_artifact_provider_raw_dump_rejected", "provider_raw")

    for case_id in EXPECTED_TEST_CASES:
        action = str(rows[case_id].get("expected_consumer_action") or "")
        require(action in {"merge_artifact_citation", "reject_success_response"}, f"{case_id} action drifted")


def citation_schema() -> dict[str, Any]:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    properties = set((schema.get("properties") or {}).keys())
    require("citations" in properties, "CopilotResponse.citations missing")
    require("artifacts" not in properties, "CopilotResponse.artifacts must not be added")
    require("artifact_metadata" not in properties, "CopilotResponse.artifact_metadata must not be added")
    citation = ((schema.get("$defs") or {}).get("citation") or {})
    kind_enum = ((((citation.get("properties") or {}).get("kind") or {}).get("enum")) or [])
    require("artifact" in kind_enum, "CopilotResponse citation kind must allow artifact")
    return citation


def assert_mapper_failure(artifact: dict[str, Any], expected_code: str, label: str) -> None:
    result = map_image_artifact_to_response_reference(artifact)
    require(result.ok is False, f"{label} must fail closed")
    require(result.failure_code == expected_code, f"{label} failure code drifted: {result.failure_code}")
    require(result.citation is None, f"{label} must not return citation")
    require(result.metadata_reference is None, f"{label} must not return metadata reference")


def require_no_forbidden_payload(metadata_reference: dict[str, Any], case_id: str) -> None:
    text = json.dumps(metadata_reference, sort_keys=True)
    for literal in (
        "base64_image",
        "pixel_payload",
        "binary_payload",
        "provider_raw_response",
        "provider_raw_dump",
        "public_url",
        "signed_public_url",
        "signed_url",
    ):
        require(literal not in text, f"{case_id} metadata reference leaked {literal}")


def assert_no_implementation_artifacts(fixture: dict[str, Any]) -> None:
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
    counters = runtime_mapper_side_effect_counters()
    counters["response_consumer_call_count"] = 0
    counters["response_builder_change_count"] = 0
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
    previous_checker = "check-image-artifact-runtime-mapper-response-consumer-integration-review-v1.py"
    current_checker = "check-image-artifact-response-consumer-implementation-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run response consumer readiness checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "response consumer readiness checker must run after integration review checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_readiness_contract(fixture)
    assert_future_test_plan(fixture)
    assert_no_implementation_artifacts(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact response consumer implementation readiness v1 checks passed.")


if __name__ == "__main__":
    main()
