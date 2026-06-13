#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
HANDSHAKE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json"
EVAL_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-eval-manifest-v0.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "copilot_response_schema_changed",
    "copilot_response_artifact_runtime_ready",
    "artifact_store_ready",
    "artifact_upload_ready",
    "artifact_binary_reader_ready",
    "production_artifact_storage_ready",
    "public_artifact_url_ready",
    "image_backend_client_ready",
    "real_backend_call_ready",
    "image_generation_ready",
    "image_pixel_payload_ready",
    "provider_raw_dump_ready",
    "image_safety_runbook_complete",
    "automatic_retry_ready",
    "fallback_execution_ready",
    "executor_ready",
    "confirmation_decision_ready",
    "writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_DEPENDENCIES = {
    "image-adapter-handshake-safety-gate-v1",
    "image-generation-eval-manifest-v0",
    "contracts/image-generation-artifact.schema.json",
    "contracts/copilot-response.schema.json",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "copilot_response_schema_changed_in_this_slice",
    "copilot_response_runtime_mapping_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_upload_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "production_storage_created_in_this_slice",
    "image_backend_client_created_in_this_slice",
    "image_generation_allowed_now",
    "pixel_payload_allowed_now",
    "provider_raw_dump_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_REQUIRED_FIELDS = {
    "artifact_id",
    "intent_id",
    "backend_request_id",
    "status",
    "artifact.uri",
    "artifact.mime_type",
    "artifact.width",
    "artifact.height",
    "artifact.format",
    "artifact.sha256",
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
    "created_at",
}
EXPECTED_FORBIDDEN_FIELDS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "signed_public_url",
    "storage_write_result",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
}
EXPECTED_RUNBOOK_STEPS = [
    "consume_safety_gated_artifact_metadata",
    "validate_metadata_reference_fields",
    "redact_non_metadata_payload",
    "map_to_future_response_reference",
    "defer_binary_retrieval_to_future_artifact_store",
]
EXPECTED_FAILURE_CODES = {
    "image_backend_unavailable",
    "image_artifact_metadata_missing",
    "image_artifact_hash_mismatch",
    "image_artifact_safety_blocked",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_return_mapper.py",
    "services/runtime/image_artifact_store.py",
    "services/platform/internal/httpapi/image_artifacts.go",
    "services/platform/internal/httpapi/image_artifact_store.go",
    "contracts/image-artifact-return.schema.json",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactPanel.tsx",
    "scripts/run-image-artifact-upload.py",
    "deploy/image-artifact-store.yaml",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageArtifactReturnMapper",
    "ImageArtifactStore",
    "NewImageArtifactStore",
    "image_artifact_return_mapper",
    "image_artifact_store",
    "IMAGE_ARTIFACT_STORE_URL",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "ImageArtifactPanel",
    "uploadImageArtifact",
    "resolvePublicImageArtifactURL",
    "ReadImageArtifactBinary",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_upload_artifacts",
    "does_not_write_production_storage",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "artifact_binary_read_count=0",
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
    "artifact_binary_read",
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


def value_at_path(document: dict[str, Any], dotted_path: str) -> Any:
    current: Any = document
    for part in dotted_path.split("."):
        require(isinstance(current, dict) and part in current, f"artifact fixture missing {dotted_path}")
        current = current[part]
    return current


def assert_schema_and_artifact_fixture() -> None:
    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)
    copilot_response_schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.Draft202012Validator.check_schema(artifact_schema)
    jsonschema.Draft202012Validator.check_schema(copilot_response_schema)
    jsonschema.validate(artifact, artifact_schema)
    require(artifact.get("kind") == "image_generation_artifact", "artifact fixture kind drifted")
    uri = str(value_at_path(artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "artifact URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "artifact URI must not be public URL")
    require(value_at_path(artifact, "artifact.sha256"), "artifact hash required")
    require(value_at_path(artifact, "provenance.trace_ids"), "artifact provenance trace required")

    citation = ((copilot_response_schema.get("$defs") or {}).get("citation") or {}).get("properties") or {}
    citation_kind = (((citation.get("kind") or {}).get("enum")) or [])
    require("artifact" in citation_kind, "CopilotResponse citation kind must still allow artifact references")


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_artifact_return_runbook_evidence_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-return-runbook-evidence-v1", "unexpected slice id")
    require(slice_info.get("track") == "Image Path", "unexpected track")
    require(
        slice_info.get("status") == "image_artifact_return_runbook_evidence_defined",
        "artifact return runbook status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")
    declared = set(fixture.get("depends_on") or [])
    missing_dependencies = sorted(EXPECTED_DEPENDENCIES - declared)
    require(not missing_dependencies, f"missing dependencies: {missing_dependencies}")


def assert_upstream_evidence() -> None:
    handshake = load_json(HANDSHAKE_FIXTURE_PATH)
    handshake_slice = handshake.get("slice") or {}
    require(handshake_slice.get("id") == "image-adapter-handshake-safety-gate-v1", "handshake id drifted")
    require(
        handshake_slice.get("status") == "image_adapter_handshake_safety_gate_defined",
        "handshake status drifted",
    )
    policy = handshake.get("artifact_return_policy") or {}
    require(policy.get("status") == "metadata_reference_only", "handshake artifact policy drifted")
    require(policy.get("artifact_uri_is_public_url") is False, "handshake must block public URL")
    require(policy.get("artifact_upload_allowed_now") is False, "handshake must block upload")

    manifest = load_json(EVAL_MANIFEST_PATH)
    require(manifest.get("status") == "draft", "image eval manifest must remain draft")
    dimensions = set((manifest.get("scope") or {}).get("eval_dimensions") or [])
    require("artifact_metadata" in dimensions, "image eval manifest must keep artifact metadata dimension")
    execution = manifest.get("execution_policy") or {}
    for field in ("does_not_call_backend", "does_not_generate_images", "does_not_download_models"):
        require(execution.get(field) is True, f"eval manifest {field} must remain true")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("artifact_return_boundary") or {}
    require(
        boundary.get("status") == "artifact_return_runbook_defined_no_runtime_mapping",
        "artifact return boundary status drifted",
    )
    require(boundary.get("decision") == "metadata_reference_contract_only", "artifact return decision drifted")
    require(boundary.get("source_artifact_schema") == "contracts/image-generation-artifact.schema.json", "schema ref")
    require(boundary.get("source_artifact_fixture") == "scripts/checks/fixtures/image-generation-artifact-basic.json", "fixture ref")
    require(
        boundary.get("upstream_handshake_fixture")
        == "scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json",
        "handshake fixture ref",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_metadata_reference_shape(fixture: dict[str, Any]) -> None:
    shape = fixture.get("metadata_reference_shape") or {}
    require(shape.get("status") == "defined_not_materialized", "metadata shape status drifted")
    require(shape.get("future_reference_kind") == "image_artifact_metadata_reference", "reference kind drifted")
    require(shape.get("source_kind") == "image_generation_artifact", "source kind drifted")
    require(shape.get("uri_scheme") == "artifact://", "URI scheme drifted")
    require(shape.get("uri_is_public_url") is False, "artifact URI must not be public URL")
    require(shape.get("copilot_response_schema_change_required_now") is False, "schema change must not be required")
    require(shape.get("runtime_mapping_required_now") is False, "runtime mapping must not be required")
    require(EXPECTED_REQUIRED_FIELDS.issubset(set(shape.get("required_fields") or [])), "required fields drifted")
    require(EXPECTED_FORBIDDEN_FIELDS.issubset(set(shape.get("forbidden_fields") or [])), "forbidden fields drifted")

    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    for dotted_path in EXPECTED_REQUIRED_FIELDS:
        value_at_path(artifact, dotted_path)


def assert_runbook_steps_and_failures(fixture: dict[str, Any]) -> None:
    steps = fixture.get("return_runbook_steps") or []
    ordered_ids = [str(step.get("step_id") or "") for step in sorted(steps, key=lambda row: int(row.get("order") or 0))]
    require(ordered_ids == EXPECTED_RUNBOOK_STEPS, "return runbook step order drifted")
    for step in steps:
        require(step.get("runtime_implemented_now") is False, f"{step.get('step_id')} runtime must stay false")
    for step_id in ("map_to_future_response_reference", "defer_binary_retrieval_to_future_artifact_store"):
        row = next(step for step in steps if step.get("step_id") == step_id)
        require(row.get("allowed_now") is False, f"{step_id} must remain blocked")

    failures = {
        str(row.get("failure_code") or ""): row
        for row in fixture.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    for failure_code, row in failures.items():
        require(row.get("retry_or_fallback_allowed_now") is False, f"{failure_code} retry must stay blocked")
        require(row.get("artifact_reference_returned") is False, f"{failure_code} must not return artifact reference")


def assert_visibility_and_execution(fixture: dict[str, Any]) -> None:
    visibility = fixture.get("artifact_visibility_policy") or {}
    require(visibility.get("metadata_visible_to_copilot_response") is True, "metadata visibility drifted")
    for field in (
        "pixel_payload_visible_to_copilot_response",
        "provider_raw_dump_visible_to_copilot_response",
        "public_url_visible_to_copilot_response",
        "artifact_binary_download_allowed_now",
        "artifact_upload_allowed_now",
        "production_storage_allowed_now",
    ):
        require(visibility.get(field) is False, f"{field} must remain false")

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


def assert_fallback_and_side_effects(fixture: dict[str, Any]) -> None:
    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "missing_metadata_promoted_to_success_allowed",
        "hash_mismatch_promoted_to_success_allowed",
        "backend_unavailable_promoted_to_retry_loop_allowed",
        "artifact_uri_promoted_to_public_url_allowed",
        "metadata_reference_promoted_to_binary_payload_allowed",
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
    previous_checker = "check-image-adapter-handshake-safety-gate-v1.py"
    current_checker = "check-image-artifact-return-runbook-evidence-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image artifact return runbook evidence")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "artifact return must run after handshake")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_schema_and_artifact_fixture()
    assert_slice(fixture)
    assert_upstream_evidence()
    assert_boundary(fixture)
    assert_metadata_reference_shape(fixture)
    assert_runbook_steps_and_failures(fixture)
    assert_visibility_and_execution(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_fallback_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact return runbook evidence v1 checks passed.")


if __name__ == "__main__":
    main()
