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
    ImageArtifactMappingResult,
    map_image_artifact_to_response_reference,
    runtime_mapper_side_effect_counters,
)

FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)
TASK_CARD_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-v1.json"
RUNTIME_MAPPING_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json"
)
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
MAPPER_PATH = REPO_ROOT / "services/runtime/image_artifact_runtime_mapper.py"

EXPECTED_FORBIDDEN_CLAIMS = {
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
EXPECTED_DEPENDENCIES = {
    "image-artifact-runtime-mapper-implementation-v1",
    "image-artifact-runtime-mapping-readiness-v1",
    "contracts/image-generation-artifact.schema.json",
    "contracts/copilot-response.schema.json",
}
EXPECTED_CITATION_FIELDS = {"id", "kind", "label", "locator", "source_uri"}
EXPECTED_METADATA_REFERENCE_FIELDS = {
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
EXPECTED_SUCCESS_CASES = {
    "generated_not_required",
    "generated_reviewed_pass",
}
EXPECTED_BLOCKED_CASES = {
    "blocked_artifact_status": "image_artifact_safety_blocked",
    "failed_artifact_status": "image_artifact_invalid_metadata",
    "pending_review_artifact": "image_artifact_safety_pending_review",
}
EXPECTED_FAIL_CLOSED_CASES = {
    "invalid_metadata_missing_required_field": "image_artifact_invalid_metadata",
    "hash_mismatch": "image_artifact_hash_mismatch",
    "mime_mismatch": "image_artifact_mime_mismatch",
    "dimension_mismatch": "image_artifact_dimension_mismatch",
    "public_url_claim": "image_artifact_public_url_claim",
    "signed_url_policy_missing": "image_artifact_signed_url_policy_missing",
    "binary_payload_present": "image_artifact_binary_payload_rejected",
    "provider_raw_dump_present": "image_artifact_provider_raw_dump_rejected",
    "artifact_store_missing": "image_artifact_store_missing",
    "artifact_store_unavailable": "image_artifact_store_unavailable",
    "artifact_binary_reader_missing": "image_artifact_binary_reader_missing",
    "artifact_binary_read_forbidden": "image_artifact_binary_read_forbidden",
    "safety_review_not_passed": "image_artifact_safety_review_not_passed",
    "provenance_missing": "image_artifact_provenance_missing",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_backend_adapter.py",
    "contracts/image-artifact-runtime-mapping.schema.json",
    "contracts/image-artifact-store.schema.json",
    "contracts/image-artifact-binary-reader.schema.json",
    "services/platform/internal/httpapi/image_artifacts.go",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactRuntimeMappingPanel.tsx",
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
}
FORBIDDEN_SOURCE_LITERALS = {
    "requests.",
    "urllib.request",
    "urlopen(",
    "subprocess.",
    "open(",
    "Path(",
    "IMAGE_ARTIFACT_STORE_URL",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "callImageGenerationBackend",
    "ReadImageArtifactBinary",
    "ResolveImageArtifactPublicURL",
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


def rows_by_id(rows: list[Any], key: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for row in rows:
        require(isinstance(row, dict), f"{key} rows must be JSON objects")
        row_id = str(row.get(key) or "")
        require(row_id, f"{key} row missing id")
        result[row_id] = row
    return result


def nested(document: dict[str, Any], *keys: str) -> Any:
    current: Any = document
    for key in keys:
        require(isinstance(current, dict), f"missing node before {key}")
        current = current.get(key)
    return current


def artifact_fixture() -> dict[str, Any]:
    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.validate(artifact, artifact_schema)
    return artifact


def assert_slice_and_dependencies(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "schema_version drifted")
    require(
        fixture.get("kind") == "image_artifact_runtime_mapper_runtime_implementation_v1",
        "fixture kind drifted",
    )
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapper-runtime-implementation-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapper_runtime_implemented",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependency list drifted")

    task_card_fixture = load_json(TASK_CARD_FIXTURE_PATH)
    require(
        (task_card_fixture.get("slice") or {}).get("status")
        == "image_artifact_runtime_mapper_implementation_task_card_defined",
        "implementation task card status drifted",
    )
    readiness_fixture = load_json(RUNTIME_MAPPING_READINESS_PATH)
    require(
        (readiness_fixture.get("slice") or {}).get("status") == "image_artifact_runtime_mapping_readiness_defined",
        "runtime mapping readiness status drifted",
    )


def assert_runtime_implementation_contract(fixture: dict[str, Any]) -> None:
    implementation = fixture.get("runtime_implementation") or {}
    require(implementation.get("status") == "metadata_only_runtime_mapper_implemented", "implementation status drifted")
    require(implementation.get("module") == "services/runtime/image_artifact_runtime_mapper.py", "module path drifted")
    require(implementation.get("function") == "map_image_artifact_to_response_reference", "function name drifted")
    require(implementation.get("result_type") == "ImageArtifactMappingResult", "result type drifted")
    require(implementation.get("input_kind") == "image_generation_artifact", "input kind drifted")
    for field in (
        "input_must_be_metadata_only",
        "output_must_be_metadata_only",
    ):
        require(implementation.get(field) is True, f"runtime_implementation.{field} must remain true")
    for field in (
        "copilot_response_schema_changed",
        "artifact_store_lookup_allowed",
        "artifact_binary_read_allowed",
        "public_url_resolution_allowed",
        "backend_call_allowed",
        "image_generation_allowed",
        "artifact_upload_allowed",
    ):
        require(implementation.get(field) is False, f"runtime_implementation.{field} must remain false")

    source = MAPPER_PATH.read_text(encoding="utf-8")
    for literal in FORBIDDEN_SOURCE_LITERALS:
        require(literal not in source, f"mapper source must not contain {literal}")


def assert_success_result(case_id: str, artifact: dict[str, Any]) -> None:
    result = map_image_artifact_to_response_reference(artifact)
    require(isinstance(result, ImageArtifactMappingResult), f"{case_id} result type drifted")
    require(result.ok is True, f"{case_id} must succeed: {result.failure_code}")
    require(result.failure_code is None, f"{case_id} must not expose failure code")
    require(isinstance(result.citation, dict), f"{case_id} citation missing")
    require(isinstance(result.metadata_reference, dict), f"{case_id} metadata reference missing")

    citation = result.citation
    metadata = result.metadata_reference
    require(set(citation) == EXPECTED_CITATION_FIELDS, f"{case_id} citation fields drifted")
    require(citation["kind"] == "artifact", f"{case_id} citation kind drifted")
    require(str(citation["locator"]).startswith("artifact://"), f"{case_id} locator must use artifact://")
    require(str(citation["source_uri"]).startswith("artifact://"), f"{case_id} source_uri must use artifact://")

    require(metadata["artifact_id"] == artifact["artifact_id"], f"{case_id} artifact_id drifted")
    require(metadata["uri"] == artifact["artifact"]["uri"], f"{case_id} uri drifted")
    require(metadata["sha256"] == artifact["artifact"]["sha256"], f"{case_id} sha256 drifted")
    require(metadata["mime_type"] == artifact["artifact"]["mime_type"], f"{case_id} mime drifted")
    require(metadata["dimensions"]["width"] == artifact["artifact"]["width"], f"{case_id} width drifted")
    require(metadata["dimensions"]["height"] == artifact["artifact"]["height"], f"{case_id} height drifted")
    require(metadata["safety"]["review_status"] == artifact["safety"]["review_status"], f"{case_id} safety drifted")
    require(metadata["provenance"]["source_request_id"] == artifact["provenance"]["source_request_id"], f"{case_id} provenance drifted")
    require(set(flatten_metadata_fields(metadata)) == EXPECTED_METADATA_REFERENCE_FIELDS, f"{case_id} metadata fields drifted")
    require_no_forbidden_payload(metadata, case_id)

    copilot_schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    jsonschema.validate(citation, nested(copilot_schema, "$defs", "citation"))


def assert_success_cases(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("success_cases") or [], "case_id")
    require(set(rows) == EXPECTED_SUCCESS_CASES, "success cases drifted")
    base = artifact_fixture()
    assert_success_result("generated_not_required", base)

    reviewed = copy.deepcopy(base)
    reviewed["safety"]["review_status"] = "reviewed_pass"
    reviewed["safety"]["requires_confirmation"] = True
    assert_success_result("generated_reviewed_pass", reviewed)


def assert_failure(case_id: str, artifact: dict[str, Any], expected_code: str, **kwargs: Any) -> None:
    result = map_image_artifact_to_response_reference(artifact, **kwargs)
    require(isinstance(result, ImageArtifactMappingResult), f"{case_id} result type drifted")
    require(result.ok is False, f"{case_id} must fail closed")
    require(result.failure_code == expected_code, f"{case_id} failure code drifted: {result.failure_code}")
    require(result.citation is None, f"{case_id} must not return success citation")
    require(result.metadata_reference is None, f"{case_id} must not return metadata reference")


def assert_blocked_cases(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("blocked_cases") or [], "case_id")
    require(set(rows) == set(EXPECTED_BLOCKED_CASES), "blocked cases drifted")
    base = artifact_fixture()

    blocked = copy.deepcopy(base)
    blocked["status"] = "blocked"
    blocked["safety"]["review_status"] = "blocked"
    assert_failure("blocked_artifact_status", blocked, EXPECTED_BLOCKED_CASES["blocked_artifact_status"])

    failed = copy.deepcopy(base)
    failed["status"] = "failed"
    assert_failure("failed_artifact_status", failed, EXPECTED_BLOCKED_CASES["failed_artifact_status"])

    pending = copy.deepcopy(base)
    pending["safety"]["review_status"] = "pending_review"
    assert_failure("pending_review_artifact", pending, EXPECTED_BLOCKED_CASES["pending_review_artifact"])


def assert_fail_closed_cases(fixture: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("fail_closed_cases") or [], "case_id")
    require(set(rows) == set(EXPECTED_FAIL_CLOSED_CASES), "fail-closed cases drifted")
    base = artifact_fixture()

    missing = copy.deepcopy(base)
    del missing["artifact"]["mime_type"]
    assert_failure("invalid_metadata_missing_required_field", missing, EXPECTED_FAIL_CLOSED_CASES["invalid_metadata_missing_required_field"])

    assert_failure("hash_mismatch", base, EXPECTED_FAIL_CLOSED_CASES["hash_mismatch"], expected_sha256="b" * 64)
    assert_failure("mime_mismatch", base, EXPECTED_FAIL_CLOSED_CASES["mime_mismatch"], expected_mime_type="image/webp")
    assert_failure("dimension_mismatch", base, EXPECTED_FAIL_CLOSED_CASES["dimension_mismatch"], expected_width=512)

    public_url = copy.deepcopy(base)
    public_url["artifact"]["uri"] = "https://example.invalid/image.png"
    assert_failure("public_url_claim", public_url, EXPECTED_FAIL_CLOSED_CASES["public_url_claim"])

    signed_url = copy.deepcopy(base)
    signed_url["artifact"]["signed_public_url"] = "https://example.invalid/signed.png"
    assert_failure("signed_url_policy_missing", signed_url, EXPECTED_FAIL_CLOSED_CASES["signed_url_policy_missing"])

    binary = copy.deepcopy(base)
    binary["base64_image"] = "not-real-binary"
    assert_failure("binary_payload_present", binary, EXPECTED_FAIL_CLOSED_CASES["binary_payload_present"])

    provider_raw = copy.deepcopy(base)
    provider_raw["provider_raw_response"] = {"raw": "provider dump"}
    assert_failure("provider_raw_dump_present", provider_raw, EXPECTED_FAIL_CLOSED_CASES["provider_raw_dump_present"])

    assert_failure("artifact_store_missing", base, EXPECTED_FAIL_CLOSED_CASES["artifact_store_missing"], artifact_store_state="missing")
    assert_failure("artifact_store_unavailable", base, EXPECTED_FAIL_CLOSED_CASES["artifact_store_unavailable"], artifact_store_state="unavailable")
    assert_failure("artifact_binary_reader_missing", base, EXPECTED_FAIL_CLOSED_CASES["artifact_binary_reader_missing"], artifact_binary_reader_state="missing")
    assert_failure("artifact_binary_read_forbidden", base, EXPECTED_FAIL_CLOSED_CASES["artifact_binary_read_forbidden"], artifact_binary_reader_state="forbidden")

    safety_blocked = copy.deepcopy(base)
    safety_blocked["safety"]["review_status"] = "blocked"
    assert_failure("safety_review_not_passed", safety_blocked, EXPECTED_FAIL_CLOSED_CASES["safety_review_not_passed"])

    provenance_missing = copy.deepcopy(base)
    del provenance_missing["provenance"]["source_request_id"]
    assert_failure("provenance_missing", provenance_missing, EXPECTED_FAIL_CLOSED_CASES["provenance_missing"])


def flatten_metadata_fields(metadata: dict[str, Any], prefix: str = "") -> list[str]:
    fields: list[str] = []
    for key, value in metadata.items():
        field = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            fields.extend(flatten_metadata_fields(value, field))
        else:
            fields.append(field)
    return fields


def require_no_forbidden_payload(metadata: dict[str, Any], case_id: str) -> None:
    text = json.dumps(metadata, sort_keys=True)
    for literal in ("base64_image", "pixel_payload", "provider_raw_response", "public_url", "signed_public_url"):
        require(literal not in text, f"{case_id} metadata reference leaked {literal}")


def assert_side_effects(fixture: dict[str, Any]) -> None:
    counters = runtime_mapper_side_effect_counters()
    rendered = {f"{key}={value}" for key, value in counters.items()}
    require(rendered == EXPECTED_ZERO_COUNTERS, "side-effect counters drifted")
    require(set(fixture.get("side_effect_counters_must_remain") or []) == EXPECTED_ZERO_COUNTERS, "fixture counters drifted")


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifacts drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")


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
    previous_checker = "check-image-artifact-runtime-mapper-implementation-v1.py"
    current_checker = "check-image-artifact-runtime-mapper-runtime-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime mapper runtime implementation checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime mapper runtime implementation checker must run after implementation task card checker",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice_and_dependencies(fixture)
    assert_runtime_implementation_contract(fixture)
    assert_success_cases(fixture)
    assert_blocked_cases(fixture)
    assert_fail_closed_cases(fixture)
    assert_side_effects(fixture)
    assert_forbidden_artifacts(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapper runtime implementation v1 checks passed.")


if __name__ == "__main__":
    main()
