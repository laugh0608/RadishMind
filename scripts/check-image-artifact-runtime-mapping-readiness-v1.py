#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json"
ARTIFACT_RETURN_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
SAFETY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-safety-runbook-evidence-v1.json"
BACKEND_ADAPTER_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "copilot_response_schema_changed",
    "copilot_response_artifact_runtime_ready",
    "runtime_mapper_implemented",
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
    "automatic_retry_ready",
    "fallback_execution_ready",
    "executor_ready",
    "confirmation_decision_ready",
    "writeback_ready",
    "replay_ready",
    "production_ready",
}
EXPECTED_DEPENDENCIES = {
    "image-artifact-return-runbook-evidence-v1",
    "image-safety-runbook-evidence-v1",
    "image-backend-adapter-readiness-evidence-v1",
    "image-generation-eval-manifest-v0",
    "contracts/image-generation-artifact.schema.json",
    "contracts/copilot-response.schema.json",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "copilot_response_schema_changed_in_this_slice",
    "runtime_mapper_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_upload_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "production_storage_created_in_this_slice",
    "image_backend_client_created_in_this_slice",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "model_download_allowed_now",
    "artifact_upload_allowed_now",
    "public_url_allowed_now",
    "provider_raw_dump_allowed_now",
    "binary_payload_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_TARGET_CITATION_FIELDS = {"id", "kind", "label", "locator", "source_uri"}
EXPECTED_TARGET_METADATA_FIELDS = {
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
EXPECTED_FORBIDDEN_FIELDS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "signed_public_url",
    "public_url",
    "storage_write_result",
    "binary_reader_result",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
}
EXPECTED_METADATA_FIELDS = {
    "artifact_id": ("artifact_id",),
    "artifact_uri": ("artifact.uri",),
    "artifact_sha256": ("artifact.sha256",),
    "mime_type": ("artifact.mime_type",),
    "dimensions": ("artifact.width", "artifact.height"),
    "safety_review": (
        "safety.risk_level",
        "safety.requires_confirmation",
        "safety.review_status",
        "safety.review_notes",
    ),
    "provenance": (
        "provenance.source_request_id",
        "provenance.trace_ids",
        "provenance.backend_request_id",
        "provenance.intent_id",
    ),
}
EXPECTED_RESPONSE_CASES = {
    "generated_not_required": (True, None),
    "generated_reviewed_pass": (True, None),
    "blocked_artifact_status": (False, "image_artifact_safety_blocked"),
    "failed_artifact_status": (False, "image_artifact_invalid_metadata"),
    "pending_review_artifact": (False, "image_artifact_safety_pending_review"),
}
EXPECTED_FAIL_CLOSED_CASES = {
    "invalid_metadata_missing_required_field": "image_artifact_invalid_metadata",
    "hash_mismatch": "image_artifact_hash_mismatch",
    "public_url_claim": "image_artifact_public_url_claim",
    "binary_payload_present": "image_artifact_binary_payload_rejected",
    "provider_raw_dump_present": "image_artifact_provider_raw_dump_rejected",
}
EXPECTED_FAILURE_CODES = {
    "image_artifact_invalid_metadata",
    "image_artifact_hash_mismatch",
    "image_artifact_public_url_claim",
    "image_artifact_binary_payload_rejected",
    "image_artifact_provider_raw_dump_rejected",
    "image_artifact_safety_pending_review",
    "image_artifact_safety_blocked",
    "image_backend_response_untrusted",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/copilot_response_artifact_mapper.py",
    "services/runtime/image_artifact_store.py",
    "services/platform/internal/httpapi/image_artifacts.go",
    "contracts/image-artifact-runtime-mapping.schema.json",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactRuntimeMappingPanel.tsx",
    "scripts/run-image-artifact-runtime-mapping.py",
    "deploy/image-artifact-runtime-mapping.yaml",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageArtifactRuntimeMapper",
    "CopilotResponseArtifactMapper",
    "ImageArtifactStore",
    "image_artifact_runtime_mapper",
    "copilot_response_artifact_mapper",
    "image_artifact_store",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "IMAGE_ARTIFACT_STORE_URL",
    "ImageArtifactRuntimeMappingPanel",
    "runImageArtifactRuntimeMapping",
    "ReadImageArtifactBinary",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_upload_artifacts",
    "does_not_read_artifact_binary",
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
        require(isinstance(current, dict) and part in current, f"document missing {dotted_path}")
        current = current[part]
    return current


def artifact_with_status(artifact: dict[str, Any], *, status: str, review_status: str) -> dict[str, Any]:
    return {
        **artifact,
        "status": status,
        "safety": {
            **artifact["safety"],
            "review_status": review_status,
        },
    }


def build_success_response(artifact: dict[str, Any]) -> dict[str, Any]:
    return {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "image_generation_artifact_reference_preview",
        "summary": "Image artifact metadata reference is available for future response mapping.",
        "answers": [
            {
                "kind": "artifact_reference",
                "text": "Generated image artifact metadata is available.",
                "citation_ids": ["image-artifact-citation-1"],
            }
        ],
        "issues": [],
        "proposed_actions": [],
        "citations": [
            {
                "id": "image-artifact-citation-1",
                "kind": "artifact",
                "label": artifact["artifact"]["title"],
                "locator": artifact["artifact"]["uri"],
                "source_uri": artifact["artifact"]["uri"],
            }
        ],
        "confidence": 0.8,
        "risk_level": artifact["safety"]["risk_level"],
        "requires_confirmation": artifact["safety"]["requires_confirmation"],
    }


def assert_schema_and_fixture_contracts() -> None:
    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)
    response_schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.Draft202012Validator.check_schema(artifact_schema)
    jsonschema.Draft202012Validator.check_schema(response_schema)
    jsonschema.validate(artifact, artifact_schema)

    uri = str(value_at_path(artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "basic image artifact URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "basic image artifact URI must not be public URL")
    require(value_at_path(artifact, "artifact.mime_type").startswith("image/"), "artifact mime type required")
    require(len(str(value_at_path(artifact, "artifact.sha256"))) == 64, "artifact sha256 must be 64 hex chars")
    require(int(value_at_path(artifact, "artifact.width")) > 0, "artifact width required")
    require(int(value_at_path(artifact, "artifact.height")) > 0, "artifact height required")
    require(value_at_path(artifact, "safety.review_status") == "not_required", "basic artifact review status drifted")
    require(value_at_path(artifact, "provenance.trace_ids"), "artifact provenance trace required")

    citation_def = ((response_schema.get("$defs") or {}).get("citation") or {}).get("properties") or {}
    citation_kind = (((citation_def.get("kind") or {}).get("enum")) or [])
    require("artifact" in citation_kind, "CopilotResponse citation kind must still allow artifact")
    response = build_success_response(artifact)
    jsonschema.validate(response, response_schema)


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_artifact_runtime_mapping_readiness_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapping-readiness-v1", "unexpected slice id")
    require(slice_info.get("track") == "Image Path", "unexpected track")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapping_readiness_defined",
        "image artifact runtime mapping readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")
    missing_dependencies = sorted(EXPECTED_DEPENDENCIES - set(fixture.get("depends_on") or []))
    require(not missing_dependencies, f"missing dependencies: {missing_dependencies}")


def assert_upstream_evidence() -> None:
    artifact_return = load_json(ARTIFACT_RETURN_FIXTURE_PATH)
    return_slice = artifact_return.get("slice") or {}
    require(
        return_slice.get("status") == "image_artifact_return_runbook_evidence_defined",
        "artifact return runbook status drifted",
    )
    return_shape = artifact_return.get("metadata_reference_shape") or {}
    require(return_shape.get("uri_scheme") == "artifact://", "artifact return URI scheme drifted")
    require(return_shape.get("uri_is_public_url") is False, "artifact return must still block public URLs")
    visibility = artifact_return.get("artifact_visibility_policy") or {}
    require(visibility.get("pixel_payload_visible_to_copilot_response") is False, "pixel payload visibility drifted")
    require(visibility.get("provider_raw_dump_visible_to_copilot_response") is False, "provider raw visibility drifted")

    safety = load_json(SAFETY_FIXTURE_PATH)
    safety_slice = safety.get("slice") or {}
    require(safety_slice.get("status") == "image_safety_runbook_evidence_defined", "safety status drifted")
    safety_cases = {str(row.get("case_id") or ""): row for row in safety.get("safety_decision_matrix") or []}
    require(safety_cases["artifact_review_pending"].get("metadata_reference_allowed_by_runbook") is False, "pending review must stay blocked")
    require(safety_cases["artifact_review_blocked"].get("metadata_reference_allowed_by_runbook") is False, "blocked review must stay blocked")

    backend = load_json(BACKEND_ADAPTER_FIXTURE_PATH)
    backend_slice = backend.get("slice") or {}
    require(
        backend_slice.get("status") == "image_backend_adapter_readiness_defined",
        "backend adapter readiness status drifted",
    )
    backend_failures = {str(row.get("failure_code") or "") for row in backend.get("failure_taxonomy") or []}
    require("image_backend_invalid_artifact_metadata" in backend_failures, "backend readiness must keep invalid metadata failure")
    require("image_backend_artifact_hash_mismatch" in backend_failures, "backend readiness must keep hash mismatch failure")
    require("image_backend_response_untrusted" in backend_failures, "backend readiness must keep untrusted response failure")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("runtime_mapping_boundary") or {}
    require(
        boundary.get("status") == "runtime_mapping_readiness_defined_no_mapper",
        "runtime mapping boundary status drifted",
    )
    require(boundary.get("decision") == "metadata_reference_mapping_gate_only", "runtime mapping decision drifted")
    require(boundary.get("source_artifact_schema") == "contracts/image-generation-artifact.schema.json", "artifact schema ref")
    require(boundary.get("target_response_schema") == "contracts/copilot-response.schema.json", "response schema ref")
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_future_mapping_contract(fixture: dict[str, Any]) -> None:
    contract = fixture.get("future_mapping_contract") or {}
    require(contract.get("status") == "defined_not_materialized", "future mapping contract status drifted")
    require(contract.get("target_citation_kind") == "artifact", "target citation kind drifted")
    require(contract.get("uri_scheme") == "artifact://", "future mapping URI scheme drifted")
    require(contract.get("uri_is_public_url") is False, "future mapping must reject public URL")
    require(contract.get("hash_algorithm") == "sha256", "hash algorithm drifted")
    for field in (
        "response_schema_change_required_now",
        "runtime_mapping_required_now",
        "artifact_store_required_now",
        "binary_reader_required_now",
    ):
        require(contract.get(field) is False, f"{field} must remain false")
    require(EXPECTED_TARGET_CITATION_FIELDS.issubset(set(contract.get("target_citation_fields") or [])), "citation fields drifted")
    require(
        EXPECTED_TARGET_METADATA_FIELDS.issubset(set(contract.get("target_metadata_reference_fields") or [])),
        "metadata reference fields drifted",
    )
    require(EXPECTED_FORBIDDEN_FIELDS.issubset(set(contract.get("forbidden_fields") or [])), "forbidden fields drifted")


def assert_metadata_fields(fixture: dict[str, Any]) -> None:
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    rows = {
        str(row.get("field") or ""): row
        for row in fixture.get("metadata_field_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_METADATA_FIELDS), "metadata field matrix drifted")
    for field, paths in EXPECTED_METADATA_FIELDS.items():
        row = rows[field]
        require(row.get("required_for_success_reference") is True, f"{field} must be required")
        for path in paths:
            value_at_path(artifact, path)
    uri_row = rows["artifact_uri"]
    require(uri_row.get("required_scheme") == "artifact://", "artifact URI required scheme drifted")
    require(uri_row.get("public_url_allowed") is False, "artifact URI public URL must stay blocked")
    require(rows["artifact_sha256"].get("hash_mismatch_fail_closed") is True, "hash mismatch must fail closed")


def assert_mapping_cases(fixture: dict[str, Any]) -> None:
    rows = {
        str(row.get("case_id") or ""): row
        for row in fixture.get("response_mapping_matrix") or []
        if isinstance(row, dict)
    }
    require(set(rows) == set(EXPECTED_RESPONSE_CASES), "response mapping matrix drifted")
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)

    for case_id, (allowed, failure_code) in EXPECTED_RESPONSE_CASES.items():
        row = rows[case_id]
        require(row.get("success_response_reference_allowed") is allowed, f"{case_id} success policy drifted")
        require(row.get("runtime_mapping_implemented_now") is False, f"{case_id} runtime mapping must stay false")
        if allowed:
            require(row.get("target_citation_kind") == "artifact", f"{case_id} citation kind drifted")
        else:
            require(row.get("expected_failure_code") == failure_code, f"{case_id} failure code drifted")

        candidate = artifact_with_status(
            artifact,
            status=str(row.get("artifact_status")),
            review_status=str(row.get("artifact_review_status")),
        )
        jsonschema.validate(candidate, artifact_schema)

    pending = artifact_with_status(artifact, status="generated", review_status="pending_review")
    require(pending["safety"]["review_status"] == "pending_review", "pending review fixture candidate drifted")
    require(rows["pending_review_artifact"].get("success_response_reference_allowed") is False, "pending review must fail closed")


def assert_fail_closed_cases_and_failures(fixture: dict[str, Any]) -> None:
    cases = {
        str(row.get("case_id") or ""): row
        for row in fixture.get("fail_closed_mapping_cases") or []
        if isinstance(row, dict)
    }
    require(set(cases) == set(EXPECTED_FAIL_CLOSED_CASES), "fail-closed cases drifted")
    for case_id, expected_failure in EXPECTED_FAIL_CLOSED_CASES.items():
        row = cases[case_id]
        require(row.get("expected_failure_code") == expected_failure, f"{case_id} failure code drifted")
        require(row.get("success_response_reference_allowed") is False, f"{case_id} must not return success")
        require(row.get("fail_closed") is True, f"{case_id} must fail closed")

    failures = {
        str(row.get("failure_code") or ""): row
        for row in fixture.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    for failure_code, row in failures.items():
        require(row.get("artifact_reference_returned") is False, f"{failure_code} must not return artifact reference")
        require(row.get("retry_or_fallback_allowed_now") is False, f"{failure_code} retry must stay blocked")


def assert_execution_and_side_effects(fixture: dict[str, Any]) -> None:
    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "blocked_artifact_promoted_to_success_allowed",
        "failed_artifact_promoted_to_success_allowed",
        "pending_review_promoted_to_success_allowed",
        "invalid_metadata_promoted_to_success_allowed",
        "hash_mismatch_promoted_to_success_allowed",
        "artifact_uri_promoted_to_public_url_allowed",
        "binary_payload_promoted_to_metadata_reference_allowed",
        "provider_raw_dump_promoted_to_metadata_reference_allowed",
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
    previous_checker = "check-image-backend-adapter-readiness-evidence-v1.py"
    current_checker = "check-image-artifact-runtime-mapping-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image artifact runtime mapping readiness")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "runtime mapping readiness must run after backend adapter readiness")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_schema_and_fixture_contracts()
    assert_slice(fixture)
    assert_upstream_evidence()
    assert_boundary(fixture)
    assert_future_mapping_contract(fixture)
    assert_metadata_fields(fixture)
    assert_mapping_cases(fixture)
    assert_fail_closed_cases_and_failures(fixture)
    assert_execution_and_side_effects(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapping readiness v1 checks passed.")


if __name__ == "__main__":
    main()
