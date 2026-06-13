#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.checks.image_generation import (  # noqa: E402
    check_artifact,
    check_backend_request,
    check_image_generation_intent,
    load_json_document,
    with_confirmation_required,
)


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json"
INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
INTENT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
BACKEND_REQUEST_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-backend-request-basic.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
EVAL_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-eval-manifest-v0.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "image_adapter_runtime_ready",
    "image_backend_client_ready",
    "real_backend_call_ready",
    "image_generation_ready",
    "image_pixel_quality_evaluated",
    "model_download_ready",
    "model_weights_committed",
    "artifact_upload_ready",
    "production_artifact_storage_ready",
    "public_artifact_url_ready",
    "copilot_response_artifact_runtime_ready",
    "image_safety_runbook_complete",
    "main_model_pixel_generation_ready",
    "executor_ready",
    "confirmation_decision_ready",
    "writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_DEPENDENCIES = {
    "image-generation-intent-contract",
    "image-generation-eval-manifest-v0",
    "contracts/image-generation-intent.schema.json",
    "contracts/image-generation-backend-request.schema.json",
    "contracts/image-generation-artifact.schema.json",
}
EXPECTED_CONTRACT_MAPPING = {
    "intent_schema": "contracts/image-generation-intent.schema.json",
    "backend_request_schema": "contracts/image-generation-backend-request.schema.json",
    "artifact_schema": "contracts/image-generation-artifact.schema.json",
    "intent_fixture": "scripts/checks/fixtures/image-generation-intent-basic.json",
    "backend_request_fixture": "scripts/checks/fixtures/image-generation-backend-request-basic.json",
    "artifact_fixture": "scripts/checks/fixtures/image-generation-artifact-basic.json",
    "eval_manifest": "scripts/checks/fixtures/image-generation-eval-manifest-v0.json",
}
EXPECTED_MANIFEST_DIMENSIONS = {
    "structured_intent",
    "backend_request_mapping",
    "artifact_metadata",
    "safety_gate",
    "provenance",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "real_adapter_runtime_created_in_this_slice",
    "real_backend_client_created_in_this_slice",
    "backend_queue_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "copilot_response_runtime_mapping_created_in_this_slice",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "model_download_allowed_now",
    "artifact_upload_allowed_now",
    "production_storage_allowed_now",
    "main_model_pixel_generation_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_PHASE_ORDER = [
    "core_intent_to_adapter",
    "adapter_safety_gate_before_backend_request",
    "blocked_confirmation_required_before_backend",
    "backend_result_to_artifact_metadata",
    "artifact_metadata_to_copilot_response",
]
EXPECTED_SAFETY_GATES = {
    "allow_low_risk_structured_intent_review": "approved_for_backend",
    "block_requires_confirmation": "blocked_requires_confirmation",
    "block_high_risk_or_policy_unknown": "blocked_requires_confirmation",
    "backend_unavailable": "backend_unavailable",
    "artifact_metadata_only_return": "approved_for_backend",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_start_training",
    "does_not_upload_artifacts",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "adapters/image/radishmind_image_adapter.py",
    "services/runtime/image_adapter_runtime.py",
    "services/runtime/image_generation_backend.py",
    "services/platform/internal/httpapi/image_generation.go",
    "services/platform/internal/httpapi/image_generation_artifact_store.go",
    "apps/radishmind-web/src/features/image-generation/ImageGenerationPanel.tsx",
    "scripts/run-image-generation-backend.py",
    "deploy/image-generation-backend.yaml",
}
EXPECTED_ABSENT_LITERALS = {
    "RadishMindImageAdapterRuntime",
    "ImageGenerationBackendClient",
    "NewImageGenerationBackend",
    "runImageGenerationBackend",
    "image_generation_backend_client",
    "image_generation_artifact_store",
    "IMAGE_GENERATION_BACKEND_URL",
    "IMAGE_GENERATION_MODEL_DIR",
    "ImageGenerationPanel",
    "generateImagePixels",
    "storeImageArtifact",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "production_storage_write_count=0",
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
    "production_storage_write",
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


def build_blocked_backend_request(backend_request: dict[str, Any]) -> dict[str, Any]:
    return {
        **backend_request,
        "safety": {
            **backend_request["safety"],
            "gate": "blocked_requires_confirmation",
            "requires_confirmation": True,
            "risk_level": "medium",
            "review_notes": ["blocked by image adapter safety gate because intent requires confirmation"],
        },
    }


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_adapter_handshake_safety_gate_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-adapter-handshake-safety-gate-v1", "unexpected slice id")
    require(slice_info.get("track") == "Image Path", "unexpected track")
    require(
        slice_info.get("status") == "image_adapter_handshake_safety_gate_defined",
        "image adapter handshake status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")


def assert_dependencies_and_contract_mapping(fixture: dict[str, Any]) -> None:
    declared = set(fixture.get("depends_on") or [])
    missing = sorted(EXPECTED_DEPENDENCIES - declared)
    require(not missing, f"missing dependencies: {missing}")

    mapping = fixture.get("contract_mapping") or {}
    for key, expected_path in EXPECTED_CONTRACT_MAPPING.items():
        require(mapping.get(key) == expected_path, f"contract mapping drifted for {key}")
        require((REPO_ROOT / expected_path).is_file(), f"missing mapped file: {expected_path}")
    require(
        set(mapping.get("manifest_required_dimensions") or []) == EXPECTED_MANIFEST_DIMENSIONS,
        "manifest required dimensions drifted",
    )


def assert_existing_image_contract_chain() -> None:
    intent_schema = load_json_document(INTENT_SCHEMA_PATH)
    backend_schema = load_json_document(BACKEND_REQUEST_SCHEMA_PATH)
    artifact_schema = load_json_document(ARTIFACT_SCHEMA_PATH)
    intent_fixture = load_json_document(INTENT_FIXTURE_PATH)
    backend_fixture = load_json_document(BACKEND_REQUEST_FIXTURE_PATH)
    artifact_fixture = load_json_document(ARTIFACT_FIXTURE_PATH)

    for schema in (intent_schema, backend_schema, artifact_schema):
        jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(intent_fixture, intent_schema)
    jsonschema.validate(backend_fixture, backend_schema)
    jsonschema.validate(artifact_fixture, artifact_schema)

    check_image_generation_intent(intent_fixture)
    check_backend_request(intent_fixture, backend_fixture, expected_gate="approved_for_backend")
    check_artifact(intent_fixture, backend_fixture, artifact_fixture)

    blocked_intent = with_confirmation_required(intent_fixture)
    blocked_backend = build_blocked_backend_request(backend_fixture)
    jsonschema.validate(blocked_intent, intent_schema)
    jsonschema.validate(blocked_backend, backend_schema)
    check_image_generation_intent(blocked_intent)
    check_backend_request(blocked_intent, blocked_backend, expected_gate="blocked_requires_confirmation")


def assert_eval_manifest_policy() -> None:
    manifest = load_json(EVAL_MANIFEST_PATH)
    require(manifest.get("kind") == "image_generation_eval_manifest", "eval manifest kind drifted")
    require(manifest.get("status") == "draft", "eval manifest must remain draft")
    scope = manifest.get("scope") or {}
    require(set(scope.get("eval_dimensions") or []) == EXPECTED_MANIFEST_DIMENSIONS, "eval dimensions drifted")
    execution = manifest.get("execution_policy") or {}
    for field in ("does_not_call_backend", "does_not_generate_images", "does_not_download_models"):
        require(execution.get(field) is True, f"eval manifest {field} must remain true")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("handshake_boundary") or {}
    require(boundary.get("status") == "handshake_defined_no_runtime_adapter", "boundary status drifted")
    require(boundary.get("decision") == "schema_fixture_handshake_only", "boundary decision drifted")
    require(boundary.get("core_owner") == "RadishMind-Core", "core owner drifted")
    require(boundary.get("adapter_owner") == "RadishMind-Image Adapter", "adapter owner drifted")
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_handshake_phase_matrix(fixture: dict[str, Any]) -> None:
    rows = fixture.get("handshake_phase_matrix") or []
    require(isinstance(rows, list) and len(rows) == len(EXPECTED_PHASE_ORDER), "phase matrix size drifted")
    ordered_ids = [str(row.get("phase_id") or "") for row in sorted(rows, key=lambda row: int(row.get("order") or 0))]
    require(ordered_ids == EXPECTED_PHASE_ORDER, "phase matrix order drifted")

    by_id = {str(row.get("phase_id") or ""): row for row in rows if isinstance(row, dict)}
    for phase_id, row in by_id.items():
        source_contract = str(row.get("source_contract") or "")
        fixture_path = str(row.get("fixture") or "")
        require((REPO_ROOT / source_contract).is_file(), f"{phase_id} source contract missing")
        require((REPO_ROOT / fixture_path).is_file(), f"{phase_id} fixture missing")
        require(row.get("runtime_implemented_now") is False, f"{phase_id} runtime must remain unimplemented")
        if phase_id in {
            "core_intent_to_adapter",
            "adapter_safety_gate_before_backend_request",
            "blocked_confirmation_required_before_backend",
        }:
            require(row.get("requires_safety_gate_before_next") is True, f"{phase_id} must require safety gate")
            require(row.get("backend_call_allowed_now") is False, f"{phase_id} must not call backend now")

    approved = by_id["adapter_safety_gate_before_backend_request"]
    require(approved.get("expected_gate") == "approved_for_backend", "approved phase gate drifted")
    require(approved.get("future_backend_submittable_after_gate") is True, "approved future submittable drifted")
    blocked = by_id["blocked_confirmation_required_before_backend"]
    require(blocked.get("expected_gate") == "blocked_requires_confirmation", "blocked phase gate drifted")
    require(blocked.get("future_backend_submittable_after_gate") is False, "blocked future submittable drifted")
    artifact = by_id["backend_result_to_artifact_metadata"]
    require(artifact.get("pixel_payload_committed_allowed") is False, "pixel payload must stay uncommitted")
    returned = by_id["artifact_metadata_to_copilot_response"]
    require(returned.get("public_url_claim_allowed_now") is False, "public URL claim must stay blocked")
    require(returned.get("artifact_upload_allowed_now") is False, "artifact upload must stay blocked")


def assert_safety_gate_matrix(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("gate_id") or ""): row
        for row in fixture.get("safety_gate_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_SAFETY_GATES), "safety gate ids drifted")
    for gate_id, expected_gate in EXPECTED_SAFETY_GATES.items():
        row = rows[gate_id]
        require(row.get("expected_backend_gate") == expected_gate, f"{gate_id} expected gate drifted")
        require(row.get("backend_call_allowed_now") is False, f"{gate_id} must not call backend now")
        require((REPO_ROOT / str(row.get("source_fixture") or "")).is_file(), f"{gate_id} fixture missing")

    require(rows["allow_low_risk_structured_intent_review"].get("requires_confirmation") is False, "low risk gate")
    for gate_id in ("block_requires_confirmation", "block_high_risk_or_policy_unknown"):
        require(rows[gate_id].get("requires_confirmation") is True, f"{gate_id} must require confirmation")


def assert_artifact_and_execution_policies(fixture: dict[str, Any]) -> None:
    policy = fixture.get("artifact_return_policy") or {}
    require(policy.get("status") == "metadata_reference_only", "artifact return policy status drifted")
    require(policy.get("artifact_uri_scheme") == "artifact://", "artifact URI scheme drifted")
    for field in (
        "artifact_uri_is_public_url",
        "pixel_payload_committed_allowed",
        "provider_raw_dump_committed_allowed",
        "artifact_upload_allowed_now",
        "production_storage_allowed_now",
        "copilot_response_runtime_mapping_ready_now",
    ):
        require(policy.get(field) is False, f"{field} must remain false")

    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    uri = str((artifact.get("artifact") or {}).get("uri") or "")
    require(uri.startswith("artifact://"), "basic artifact fixture must keep artifact:// URI")
    require(not uri.startswith(("http://", "https://")), "artifact URI must not be a public URL")

    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")


def assert_forbidden_artifacts_and_sources(fixture: dict[str, Any]) -> None:
    artifacts = {
        str(row.get("path") or ""): row
        for row in fixture.get("forbidden_artifact_matrix") or []
        if isinstance(row, dict)
    }
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact paths drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    configured = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_ABSENT_LITERALS.issubset(configured), "source absent literals drifted")
    for root in (REPO_ROOT / "services", REPO_ROOT / "apps", REPO_ROOT / "adapters"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".py", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(literal not in text, f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r}")


def assert_fallback_and_side_effect_policies(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "approved_gate_promoted_to_backend_call_allowed",
        "backend_unavailable_promoted_to_retry_loop_allowed",
        "artifact_metadata_promoted_to_public_url_allowed",
        "image_eval_manifest_promoted_to_quality_eval_allowed",
        "main_model_promoted_to_pixel_generator_allowed",
    ):
        require(fallback.get(field) is False, f"{field} must remain false")

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
    previous_checker = "check-image-generation-eval-manifest.py"
    current_checker = "check-image-adapter-handshake-safety-gate-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image adapter handshake safety gate")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "handshake gate must run after eval manifest")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies_and_contract_mapping(fixture)
    assert_existing_image_contract_chain()
    assert_eval_manifest_policy()
    assert_boundary(fixture)
    assert_handshake_phase_matrix(fixture)
    assert_safety_gate_matrix(fixture)
    assert_artifact_and_execution_policies(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_fallback_and_side_effect_policies(fixture)
    assert_references_and_check_repo(fixture)
    print("image adapter handshake safety gate v1 checks passed.")


if __name__ == "__main__":
    main()
