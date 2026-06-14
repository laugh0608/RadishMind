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
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-runtime-mapper-response-consumer-integration-review-v1.json"
)
RUNTIME_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
RESPONSE_CONSUMER_RUNTIME_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-response-consumer-runtime-implementation-v1.json"
)
RESPONSE_CONSUMER_MODULE_PATH = "services/runtime/image_artifact_response_consumer.py"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_DEPENDENCIES = {
    "image-artifact-runtime-mapper-runtime-implementation-v1",
    "contracts/copilot-response.schema.json",
    "contracts/image-generation-artifact.schema.json",
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/inference_response.py",
    "services/runtime/inference_support.py",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "response_consumer_implemented",
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
EXPECTED_GATE_STATUS = {
    "runtime_mapper_implementation_consumed": "satisfied",
    "copilot_response_citation_schema_accepts_artifact": "satisfied",
    "success_result_can_supply_artifact_citation": "satisfied",
    "metadata_reference_remains_internal_metadata_only": "satisfied",
    "blocked_failed_pending_review_rejected_before_success_response": "satisfied",
    "fail_closed_errors_do_not_emit_success_citation": "satisfied",
    "response_consumer_implementation_task_card": "deferred",
    "artifact_store_binary_reader_backend_adapter": "blocked",
}
EXPECTED_ENTRY_CANDIDATES = {
    "copilot_response_citations_schema": "selected_schema_surface",
    "runtime_response_coercion": "future_implementation_candidate",
    "runtime_context_citation_builder": "future_implementation_candidate",
    "platform_http_response_route": "blocked_for_this_slice",
}
EXPECTED_PROPAGATION = {
    "generated_not_required_success": (True, None),
    "generated_reviewed_pass_success": (True, None),
    "blocked_artifact_status": (False, "image_artifact_safety_blocked"),
    "failed_artifact_status": (False, "image_artifact_invalid_metadata"),
    "pending_review_artifact": (False, "image_artifact_safety_pending_review"),
    "public_url_claim": (False, "image_artifact_public_url_claim"),
    "binary_payload_present": (False, "image_artifact_binary_payload_rejected"),
    "provider_raw_dump_present": (False, "image_artifact_provider_raw_dump_rejected"),
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_response_consumer.py",
    "services/runtime/copilot_response_artifact_mapper.py",
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "docs/task-cards/image-artifact-runtime-mapper-response-consumer-integration-v1-plan.md",
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


def response_consumer_runtime_allows_artifact(relative_path: str) -> bool:
    if relative_path != RESPONSE_CONSUMER_MODULE_PATH:
        return False
    if not RESPONSE_CONSUMER_RUNTIME_PATH.exists():
        return False
    runtime = load_json(RESPONSE_CONSUMER_RUNTIME_PATH)
    return (
        (runtime.get("slice") or {}).get("status")
        == "image_artifact_response_consumer_runtime_implemented"
    )


def artifact_fixture() -> dict[str, Any]:
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.validate(artifact, load_json(ARTIFACT_SCHEMA_PATH))
    return artifact


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "image_artifact_runtime_mapper_response_consumer_integration_review_v1",
        "fixture kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(
        slice_info.get("id") == "image-artifact-runtime-mapper-response-consumer-integration-review-v1",
        "slice id drifted",
    )
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status")
        == "image_artifact_runtime_mapper_response_consumer_integration_review_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependency list drifted")

    runtime_fixture = load_json(RUNTIME_IMPLEMENTATION_PATH)
    runtime_status = (runtime_fixture.get("slice") or {}).get("status")
    require(
        runtime_status == "image_artifact_runtime_mapper_runtime_implemented",
        "runtime mapper implementation status drifted",
    )


def assert_integration_review(fixture: dict[str, Any]) -> None:
    review = fixture.get("integration_review") or {}
    require(review.get("status") == "response_consumer_integration_entry_review_defined", "review status drifted")
    require(
        review.get("conclusion")
        == "future_consumer_must_use_existing_copilot_response_citations_path_and_metadata_only_mapper",
        "review conclusion drifted",
    )
    for key in (
        "implementation_created_in_this_slice",
        "response_builder_changed_in_this_slice",
        "copilot_response_schema_change_allowed",
        "metadata_reference_public_output_allowed",
        "artifact_store_lookup_allowed",
        "artifact_binary_read_allowed",
        "public_url_resolution_allowed",
        "backend_call_allowed",
        "image_generation_allowed",
    ):
        require(review.get(key) is False, f"integration_review.{key} must remain false")

    candidates = rows_by_id(fixture.get("consumer_entry_candidates") or [], "candidate_id")
    require(set(candidates) == set(EXPECTED_ENTRY_CANDIDATES), "consumer entry candidates drifted")
    for candidate_id, expected_decision in EXPECTED_ENTRY_CANDIDATES.items():
        candidate = candidates[candidate_id]
        require(candidate.get("decision") == expected_decision, f"{candidate_id} decision drifted")
        require(candidate.get("implementation_created") is False, f"{candidate_id} must not be implemented")
        surface = str(candidate.get("surface") or "")
        require(surface, f"{candidate_id} surface missing")
        if "#" not in surface:
            require((REPO_ROOT / surface).exists(), f"{candidate_id} surface file missing: {surface}")

    gates = rows_by_id(fixture.get("consumer_gate_matrix") or [], "gate_id")
    require(set(gates) == set(EXPECTED_GATE_STATUS), "consumer gate matrix drifted")
    for gate_id, expected_status in EXPECTED_GATE_STATUS.items():
        require(gates[gate_id].get("status") == expected_status, f"{gate_id} status drifted")


def assert_copilot_response_schema() -> None:
    schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    citation_schema = (((schema.get("$defs") or {}).get("citation") or {}).get("properties") or {})
    kind_enum = (((citation_schema.get("kind") or {}).get("enum")) or [])
    require("artifact" in kind_enum, "CopilotResponse citation kind must continue to allow artifact")
    properties = set((schema.get("properties") or {}).keys())
    require("citations" in properties, "CopilotResponse citations property missing")
    require("artifacts" not in properties, "this review must not add CopilotResponse.artifacts")
    require("artifact_metadata" not in properties, "this review must not add CopilotResponse.artifact_metadata")


def assert_propagation_cases(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("propagation_cases") or [], "case_id")
    require(set(rows) == set(EXPECTED_PROPAGATION), "propagation cases drifted")
    base = artifact_fixture()

    cases: dict[str, tuple[dict[str, Any], dict[str, Any]]] = {
        "generated_not_required_success": (copy.deepcopy(base), {}),
        "generated_reviewed_pass_success": (reviewed_artifact(base), {}),
        "blocked_artifact_status": (mutated_artifact(base, status="blocked", review_status="blocked"), {}),
        "failed_artifact_status": (mutated_artifact(base, status="failed"), {}),
        "pending_review_artifact": (mutated_artifact(base, review_status="pending_review"), {}),
        "public_url_claim": (mutated_artifact(base, uri="https://example.invalid/image.png"), {}),
        "binary_payload_present": (mutated_artifact(base, extra_root={"base64_image": "not-real-binary"}), {}),
        "provider_raw_dump_present": (mutated_artifact(base, extra_root={"provider_raw_response": {"raw": True}}), {}),
    }

    citation_schema = ((load_json(COPILOT_RESPONSE_SCHEMA_PATH).get("$defs") or {}).get("citation") or {})
    for case_id, (expected_ok, expected_failure_code) in EXPECTED_PROPAGATION.items():
        artifact, kwargs = cases[case_id]
        result = map_image_artifact_to_response_reference(artifact, **kwargs)
        require(result.ok is expected_ok, f"{case_id} ok state drifted")
        require(rows[case_id].get("mapper_ok") is expected_ok, f"{case_id} fixture mapper_ok drifted")
        if expected_ok:
            require(result.citation is not None, f"{case_id} citation missing")
            require(result.metadata_reference is not None, f"{case_id} metadata reference missing")
            jsonschema.validate(result.citation, citation_schema)
            assert_no_forbidden_payload(result.metadata_reference, case_id)
            require(
                rows[case_id].get("expected_consumer_action") == "allow_future_success_response_citation",
                f"{case_id} consumer action drifted",
            )
        else:
            require(result.failure_code == expected_failure_code, f"{case_id} failure code drifted")
            require(result.citation is None, f"{case_id} must not return success citation")
            require(result.metadata_reference is None, f"{case_id} must not return metadata reference")
            require(rows[case_id].get("expected_failure_code") == expected_failure_code, f"{case_id} fixture failure drifted")
            require(
                rows[case_id].get("expected_consumer_action") == "reject_success_response",
                f"{case_id} consumer action drifted",
            )


def reviewed_artifact(base: dict[str, Any]) -> dict[str, Any]:
    artifact = copy.deepcopy(base)
    artifact["safety"]["requires_confirmation"] = True
    artifact["safety"]["review_status"] = "reviewed_pass"
    return artifact


def mutated_artifact(
    base: dict[str, Any],
    *,
    status: str | None = None,
    review_status: str | None = None,
    uri: str | None = None,
    extra_root: dict[str, Any] | None = None,
) -> dict[str, Any]:
    artifact = copy.deepcopy(base)
    if status is not None:
        artifact["status"] = status
    if review_status is not None:
        artifact["safety"]["review_status"] = review_status
    if uri is not None:
        artifact["artifact"]["uri"] = uri
    if extra_root:
        artifact.update(extra_root)
    return artifact


def assert_no_forbidden_payload(metadata_reference: dict[str, Any], case_id: str) -> None:
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
        if not response_consumer_runtime_allows_artifact(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    forbidden_literals = set(fixture.get("forbidden_source_literals") or [])
    require(forbidden_literals, "forbidden source literals missing")
    for relative_path in fixture.get("source_files_that_must_not_be_connected") or []:
        source = (REPO_ROOT / str(relative_path)).read_text(encoding="utf-8")
        leaked = sorted(literal for literal in forbidden_literals if literal in source)
        require(not leaked, f"{relative_path} must not be connected to image artifact mapper: {leaked}")


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
    previous_checker = "check-image-artifact-runtime-mapper-runtime-implementation-v1.py"
    current_checker = "check-image-artifact-runtime-mapper-response-consumer-integration-review-v1.py"
    require(current_checker in check_repo, "check-repo.py must run response consumer integration review checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "response consumer integration review checker must run after runtime mapper implementation checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_integration_review(fixture)
    assert_copilot_response_schema()
    assert_propagation_cases(fixture)
    assert_no_implementation_artifacts(fixture)
    assert_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapper response consumer integration review v1 checks passed.")


if __name__ == "__main__":
    main()
