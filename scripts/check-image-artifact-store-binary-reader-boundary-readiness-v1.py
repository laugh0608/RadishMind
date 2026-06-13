#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json"
)
ENTRY_REVIEW_PATH = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-artifact-runtime-mapping-implementation-entry-review-v1.json"
)
RUNTIME_MAPPING_READINESS_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapping-readiness-v1.json"
)
ARTIFACT_RETURN_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
)
SAFETY_PATH = REPO_ROOT / "scripts/checks/fixtures/image-safety-runbook-evidence-v1.json"
BACKEND_ADAPTER_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json"
)
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD = (
    "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md"
)
RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json"
)
RUNTIME_MAPPER_MODULE_PATH = "services/runtime/image_artifact_runtime_mapper.py"
RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)

DEPENDENCY_STATUS_BY_SLICE = {
    "image-artifact-runtime-mapping-implementation-entry-review-v1": (
        ENTRY_REVIEW_PATH,
        "image_artifact_runtime_mapping_entry_review_defined",
    ),
    "image-artifact-runtime-mapping-readiness-v1": (
        RUNTIME_MAPPING_READINESS_PATH,
        "image_artifact_runtime_mapping_readiness_defined",
    ),
    "image-artifact-return-runbook-evidence-v1": (
        ARTIFACT_RETURN_PATH,
        "image_artifact_return_runbook_evidence_defined",
    ),
    "image-safety-runbook-evidence-v1": (
        SAFETY_PATH,
        "image_safety_runbook_evidence_defined",
    ),
    "image-backend-adapter-readiness-evidence-v1": (
        BACKEND_ADAPTER_PATH,
        "image_backend_adapter_readiness_defined",
    ),
}
EXPECTED_DEPENDENCIES = set(DEPENDENCY_STATUS_BY_SLICE) | {"contracts/image-generation-artifact.schema.json"}
EXPECTED_FORBIDDEN_CLAIMS = {
    "artifact_store_ready",
    "artifact_store_implemented",
    "artifact_upload_ready",
    "artifact_binary_reader_ready",
    "artifact_binary_reader_implemented",
    "artifact_binary_read_ready",
    "public_artifact_url_ready",
    "public_url_resolver_ready",
    "signed_url_resolver_ready",
    "production_artifact_storage_ready",
    "runtime_mapper_ready",
    "runtime_mapper_implemented",
    "copilot_response_schema_changed",
    "copilot_response_artifact_runtime_ready",
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
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "artifact_upload_created_in_this_slice",
    "production_storage_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "signed_url_resolver_created_in_this_slice",
    "runtime_mapper_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "copilot_response_schema_changed_in_this_slice",
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
EXPECTED_STORE_INPUTS = {
    "artifact_id",
    "artifact.uri",
    "artifact.sha256",
    "artifact.mime_type",
    "artifact.width",
    "artifact.height",
    "safety.review_status",
    "provenance.source_request_id",
    "provenance.trace_ids",
}
EXPECTED_STORE_OUTPUTS = {
    "artifact_id",
    "artifact_uri",
    "storage_ref",
    "verified_sha256",
    "verified_mime_type",
    "verified_width",
    "verified_height",
    "verified_at",
}
EXPECTED_FORBIDDEN_LOOKUP_INPUTS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "public_url",
    "signed_public_url",
}
EXPECTED_READER_INPUTS = {
    "artifact_id",
    "artifact_uri",
    "storage_ref",
    "expected_sha256",
    "expected_mime_type",
    "expected_width",
    "expected_height",
    "safety.review_status",
}
EXPECTED_PRE_READ_GATES = {
    "artifact_store_lookup_verified",
    "artifact_uri_scheme_verified",
    "sha256_expected",
    "mime_type_expected",
    "dimensions_expected",
    "safety_review_allows_read",
    "public_url_policy_satisfied",
}
EXPECTED_VALIDATION_IDS = {
    "artifact_uri_scheme",
    "artifact_sha256",
    "mime_type",
    "dimensions",
    "safety_review",
    "provenance",
}
EXPECTED_FAILURE_CODES = {
    "image_artifact_store_missing",
    "image_artifact_store_unavailable",
    "image_artifact_binary_reader_missing",
    "image_artifact_binary_read_forbidden",
    "image_artifact_invalid_uri",
    "image_artifact_hash_mismatch",
    "image_artifact_mime_mismatch",
    "image_artifact_dimension_mismatch",
    "image_artifact_public_url_claim",
    "image_artifact_signed_url_policy_missing",
    "image_artifact_binary_payload_rejected",
    "image_artifact_provider_raw_dump_rejected",
    "image_artifact_safety_review_not_passed",
    "image_artifact_provenance_missing",
}
EXPECTED_RUNTIME_PLAN_INPUTS = {
    "store_boundary_contract",
    "binary_reader_boundary_contract",
    "validation_matrix",
    "public_url_policy",
    "failure_taxonomy",
    "no_fake_fallback_policy",
    "no_side_effect_policy",
}
EXPECTED_FORBIDDEN_TASK_CARDS = {
    "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-store-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-binary-reader-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-public-url-resolver-implementation-v1-plan.md",
    "docs/task-cards/image-backend-adapter-implementation-v1-plan.md",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_store.py",
    "services/runtime/image_artifact_binary_reader.py",
    "services/runtime/image_artifact_public_url_resolver.py",
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/copilot_response_artifact_mapper.py",
    "services/runtime/image_backend_adapter.py",
    "contracts/image-artifact-store.schema.json",
    "contracts/image-artifact-binary-reader.schema.json",
    "contracts/image-artifact-runtime-mapping.schema.json",
    "services/platform/internal/httpapi/image_artifacts.go",
    "deploy/image-artifact-store.yaml",
    "apps/radishmind-web/src/features/image-generation/ImageArtifactStorePanel.tsx",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageArtifactStore",
    "ImageArtifactBinaryReader",
    "ImageArtifactPublicUrlResolver",
    "ImageArtifactRuntimeMapper",
    "CopilotResponseArtifactMapper",
    "image_artifact_store",
    "image_artifact_binary_reader",
    "image_artifact_public_url_resolver",
    "image_artifact_runtime_mapper",
    "copilot_response_artifact_mapper",
    "IMAGE_ARTIFACT_STORE_URL",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "IMAGE_ARTIFACT_SIGNED_URL_TTL",
    "ReadImageArtifactBinary",
    "ResolveImageArtifactPublicURL",
    "uploadImageArtifact",
    "runImageArtifactRuntimeMapping",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_upload_artifacts",
    "does_not_read_artifact_binary",
    "does_not_write_production_storage",
    "does_not_create_public_url",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_NO_FAKE_FALSE_FIELDS = {
    "boundary_readiness_promoted_to_runtime_implementation_allowed",
    "missing_artifact_store_promoted_to_success_allowed",
    "missing_binary_reader_promoted_to_success_allowed",
    "hash_mismatch_promoted_to_success_allowed",
    "mime_mismatch_promoted_to_success_allowed",
    "dimension_mismatch_promoted_to_success_allowed",
    "public_url_claim_promoted_to_success_allowed",
    "signed_url_policy_missing_promoted_to_success_allowed",
    "binary_payload_promoted_to_metadata_reference_allowed",
    "provider_raw_dump_promoted_to_metadata_reference_allowed",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "image_generation_count=0",
    "model_download_count=0",
    "artifact_upload_count=0",
    "artifact_binary_read_count=0",
    "artifact_store_lookup_count=0",
    "production_storage_write_count=0",
    "public_url_resolution_count=0",
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
    "artifact_store_lookup",
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


def later_entry_allows_task_card(relative_path: str) -> bool:
    if relative_path != RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD:
        return False
    if not RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH.exists():
        return False

    entry = load_json(RUNTIME_MAPPER_IMPLEMENTATION_ENTRY_PATH)
    boundary = entry.get("implementation_entry_boundary") or {}
    selected = entry.get("selected_track_contract") or {}
    return (
        boundary.get("decision") == "runtime_mapper_implementation_task_card_allowed_next"
        and selected.get("next_task_card") == RUNTIME_MAPPER_IMPLEMENTATION_TASK_CARD
    )


def later_runtime_mapper_implementation_allows_artifact(relative_path: str) -> bool:
    if relative_path != RUNTIME_MAPPER_MODULE_PATH:
        return False
    if not RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH.exists():
        return False

    implementation = load_json(RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH)
    runtime = implementation.get("runtime_implementation") or {}
    return (
        (implementation.get("slice") or {}).get("status")
        == "image_artifact_runtime_mapper_runtime_implemented"
        and runtime.get("status") == "metadata_only_runtime_mapper_implemented"
        and runtime.get("module") == RUNTIME_MAPPER_MODULE_PATH
    )


def rows_by_id(rows: list[Any], key: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for row in rows:
        require(isinstance(row, dict), f"{key} rows must be JSON objects")
        row_id = str(row.get(key) or "")
        require(row_id, f"{key} row missing id")
        result[row_id] = row
    return result


def slice_status(document: dict[str, Any]) -> str:
    slice_info = document.get("slice") or {}
    require(isinstance(slice_info, dict), "slice must be a JSON object")
    return str(slice_info.get("status") or "")


def value_at_path(document: dict[str, Any], dotted_path: str) -> Any:
    current: Any = document
    for part in dotted_path.split("."):
        require(isinstance(current, dict) and part in current, f"document missing {dotted_path}")
        current = current[part]
    return current


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_artifact_store_binary_reader_boundary_readiness_v1", "unexpected kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-store-binary-reader-boundary-readiness-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "track drifted")
    require(
        slice_info.get("status") == "image_artifact_store_binary_reader_boundary_readiness_defined",
        "boundary readiness status drifted",
    )
    missing_claims = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing_claims, f"missing forbidden claims: {missing_claims}")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependency list drifted")


def assert_dependencies() -> dict[str, dict[str, Any]]:
    dependencies: dict[str, dict[str, Any]] = {}
    for slice_id, (path, expected_status) in DEPENDENCY_STATUS_BY_SLICE.items():
        document = load_json(path)
        require(slice_status(document) == expected_status, f"{slice_id} status drifted")
        dependencies[slice_id] = document

    entry_review = dependencies["image-artifact-runtime-mapping-implementation-entry-review-v1"]
    policy = entry_review.get("next_development_policy") or {}
    require(
        policy.get("status") == "artifact_store_binary_boundary_readiness_next",
        "entry review must still point to store/binary boundary readiness",
    )
    ordered = [
        str(row.get("candidate_id") or "")
        for row in entry_review.get("entry_review_order") or []
        if isinstance(row, dict)
    ]
    require(ordered and ordered[0] == "artifact_store_binary_boundary_readiness", "entry review order drifted")

    runtime = dependencies["image-artifact-runtime-mapping-readiness-v1"]
    future_contract = runtime.get("future_mapping_contract") or {}
    require(future_contract.get("uri_scheme") == "artifact://", "runtime readiness URI scheme drifted")
    require(future_contract.get("artifact_store_required_now") is False, "artifact store must not be required now")
    require(future_contract.get("binary_reader_required_now") is False, "binary reader must not be required now")
    require(future_contract.get("runtime_mapping_required_now") is False, "runtime mapping must not be required now")

    artifact_return = dependencies["image-artifact-return-runbook-evidence-v1"]
    visibility = artifact_return.get("artifact_visibility_policy") or {}
    require(visibility.get("artifact_binary_download_allowed_now") is False, "binary download policy drifted")
    require(visibility.get("production_storage_allowed_now") is False, "production storage policy drifted")
    require(visibility.get("public_url_visible_to_copilot_response") is False, "public URL visibility drifted")

    safety = dependencies["image-safety-runbook-evidence-v1"]
    safety_cases = rows_by_id(safety.get("safety_decision_matrix") or [], "case_id")
    require(
        safety_cases["artifact_review_pending"].get("metadata_reference_allowed_by_runbook") is False,
        "pending review must stay blocked",
    )
    require(
        safety_cases["artifact_review_blocked"].get("metadata_reference_allowed_by_runbook") is False,
        "blocked review must stay blocked",
    )

    backend = dependencies["image-backend-adapter-readiness-evidence-v1"]
    backend_failures = {
        str(row.get("failure_code") or "")
        for row in backend.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require("image_backend_invalid_artifact_metadata" in backend_failures, "backend invalid metadata failure drifted")
    require("image_backend_artifact_hash_mismatch" in backend_failures, "backend hash mismatch failure drifted")
    require("image_backend_response_untrusted" in backend_failures, "backend untrusted response failure drifted")
    return dependencies


def assert_artifact_schema_fixture() -> None:
    schema = load_json(ARTIFACT_SCHEMA_PATH)
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(artifact, schema)
    uri = str(value_at_path(artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "artifact fixture URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "artifact fixture must not use public URL")
    require(len(str(value_at_path(artifact, "artifact.sha256"))) == 64, "artifact sha256 must remain 64 chars")
    require(str(value_at_path(artifact, "artifact.mime_type")).startswith("image/"), "artifact mime must remain image/*")
    require(int(value_at_path(artifact, "artifact.width")) > 0, "artifact width required")
    require(int(value_at_path(artifact, "artifact.height")) > 0, "artifact height required")
    require(value_at_path(artifact, "provenance.trace_ids"), "artifact provenance trace required")


def assert_boundary_readiness(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("boundary_readiness") or {}
    require(
        boundary.get("status") == "store_binary_reader_boundary_defined_no_runtime_store",
        "boundary readiness status drifted",
    )
    require(boundary.get("decision") == "storage_and_binary_contract_gate_only", "boundary decision drifted")
    require(
        boundary.get("current_development_mode") == "runtime_mapping_implementation_plan_review_next",
        "current development mode drifted",
    )
    require(
        boundary.get("entry_review_consumed") == "image-artifact-runtime-mapping-implementation-entry-review-v1",
        "entry review dependency drifted",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"boundary_readiness.{field} must remain false")


def assert_store_and_reader_contracts(fixture: dict[str, Any]) -> None:
    store = fixture.get("store_boundary_contract") or {}
    require(store.get("status") == "defined_not_implemented", "store contract status drifted")
    require(store.get("artifact_uri_scheme") == "artifact://", "store URI scheme drifted")
    require(store.get("artifact_uri_is_public_url") is False, "store URI must not be public URL")
    require(EXPECTED_STORE_INPUTS.issubset(set(store.get("required_lookup_inputs") or [])), "store inputs drifted")
    require(EXPECTED_STORE_OUTPUTS.issubset(set(store.get("required_lookup_outputs") or [])), "store outputs drifted")
    require(
        EXPECTED_FORBIDDEN_LOOKUP_INPUTS.issubset(set(store.get("forbidden_lookup_inputs") or [])),
        "store forbidden inputs drifted",
    )
    for field in ("store_manifest_required_before_implementation",):
        require(store.get(field) is True, f"store_boundary_contract.{field} must remain true")
    for field in ("store_lookup_allowed_now", "store_write_allowed_now", "production_storage_allowed_now"):
        require(store.get(field) is False, f"store_boundary_contract.{field} must remain false")

    reader = fixture.get("binary_reader_boundary_contract") or {}
    require(reader.get("status") == "defined_not_implemented", "reader contract status drifted")
    require(EXPECTED_READER_INPUTS.issubset(set(reader.get("reader_inputs") or [])), "reader inputs drifted")
    require(EXPECTED_PRE_READ_GATES.issubset(set(reader.get("pre_read_gates") or [])), "pre-read gates drifted")
    for field in (
        "copilot_response_binary_payload_allowed",
        "provider_raw_dump_allowed",
        "binary_read_allowed_now",
        "reader_implementation_allowed_now",
    ):
        require(reader.get(field) is False, f"binary_reader_boundary_contract.{field} must remain false")


def assert_validation_and_public_url_policy(fixture: dict[str, Any]) -> None:
    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    matrix = rows_by_id(fixture.get("validation_matrix") or [], "id")
    require(set(matrix) == EXPECTED_VALIDATION_IDS, "validation matrix drifted")
    for row_id, row in matrix.items():
        require(row.get("required") is True, f"{row_id} must remain required")
        failure = str(row.get("failure_code") or "")
        require(failure in EXPECTED_FAILURE_CODES, f"{row_id} failure code drifted")
    require(matrix["artifact_uri_scheme"].get("required_scheme") == "artifact://", "URI scheme validation drifted")
    require(matrix["artifact_uri_scheme"].get("public_url_allowed") is False, "public URL validation drifted")
    require(matrix["artifact_sha256"].get("revalidate_before_binary_read") is True, "hash revalidation drifted")
    require(matrix["mime_type"].get("revalidate_before_binary_read") is True, "mime revalidation drifted")
    require(matrix["dimensions"].get("revalidate_before_binary_read") is True, "dimension revalidation drifted")

    value_at_path(artifact, "artifact.uri")
    value_at_path(artifact, "artifact.sha256")
    value_at_path(artifact, "artifact.mime_type")
    value_at_path(artifact, "artifact.width")
    value_at_path(artifact, "artifact.height")
    value_at_path(artifact, "safety.review_status")
    value_at_path(artifact, "provenance.source_request_id")
    value_at_path(artifact, "provenance.trace_ids")

    public_url = fixture.get("public_url_policy") or {}
    require(public_url.get("status") == "public_url_policy_defined_no_resolver", "public URL policy status drifted")
    for field in (
        "artifact_uri_promoted_to_public_url_allowed",
        "signed_url_allowed_now",
        "public_url_resolver_allowed_now",
    ):
        require(public_url.get(field) is False, f"public_url_policy.{field} must remain false")
    for field in ("production_storage_policy_required", "signed_url_expiry_policy_required"):
        require(public_url.get(field) is True, f"public_url_policy.{field} must remain true")
    require(public_url.get("public_url_claim_failure_code") == "image_artifact_public_url_claim", "public URL failure drifted")
    require(
        public_url.get("signed_url_policy_missing_failure_code") == "image_artifact_signed_url_policy_missing",
        "signed URL failure drifted",
    )


def assert_failures_and_runtime_plan(fixture: dict[str, Any]) -> None:
    failures = rows_by_id(fixture.get("failure_taxonomy") or [], "failure_code")
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    for failure_code, row in failures.items():
        require(row.get("artifact_reference_returned") is False, f"{failure_code} must not return artifact reference")
        require(row.get("retry_or_fallback_allowed_now") is False, f"{failure_code} retry must remain blocked")

    reconciliation = fixture.get("entry_review_reconciliation") or {}
    require(
        reconciliation.get("status") == "boundary_readiness_resolves_first_blocker_only",
        "entry reconciliation status drifted",
    )
    require(
        reconciliation.get("entry_review_status") == "image_artifact_runtime_mapping_entry_review_defined",
        "entry reconciliation source status drifted",
    )
    require(
        reconciliation.get("resolved_blocker") == "artifact_store_binary_boundary_readiness_missing",
        "resolved blocker drifted",
    )
    for field in (
        "runtime_mapper_implementation_plan_ready_now",
        "runtime_mapper_implementation_allowed_now",
        "artifact_store_implementation_allowed_now",
        "artifact_binary_reader_implementation_allowed_now",
        "public_url_resolver_implementation_allowed_now",
        "backend_adapter_implementation_allowed_now",
        "implementation_trigger_satisfied_now",
    ):
        require(reconciliation.get(field) is False, f"entry_review_reconciliation.{field} must remain false")

    plan = fixture.get("runtime_mapping_plan_inputs") or {}
    require(plan.get("status") == "ready_as_boundary_evidence_only", "runtime plan input status drifted")
    require(
        EXPECTED_RUNTIME_PLAN_INPUTS.issubset(set(plan.get("inputs_available_to_future_plan") or [])),
        "runtime plan inputs drifted",
    )
    require(plan.get("runtime_mapping_implementation_plan_review_allowed_next") is True, "plan review should be next")
    require(plan.get("runtime_mapper_implementation_allowed_now") is False, "runtime mapper must remain blocked")
    require(
        plan.get("requires_dedicated_task_card_before_runtime_code") is True,
        "runtime code must require a dedicated task card",
    )


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    task_cards = rows_by_id(fixture.get("forbidden_task_card_matrix") or [], "path")
    require(set(task_cards) == EXPECTED_FORBIDDEN_TASK_CARDS, "forbidden task cards drifted")
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if not later_entry_allows_task_card(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifact paths drifted")
    for relative_path, row in artifacts.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        if not later_runtime_mapper_implementation_allows_artifact(relative_path):
            require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    configured_absent_literals = set(fixture.get("source_absent_literals") or [])
    require(EXPECTED_ABSENT_LITERALS.issubset(configured_absent_literals), "source absent literals drifted")


def assert_execution_and_side_effects(fixture: dict[str, Any]) -> None:
    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in EXPECTED_NO_FAKE_FALSE_FIELDS:
        require(fallback.get(field) is False, f"no_fake_fallback_policy.{field} must remain false")

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
    previous_checker = "check-image-artifact-runtime-mapping-implementation-entry-review-v1.py"
    current_checker = "check-image-artifact-store-binary-reader-boundary-readiness-v1.py"
    require(current_checker in check_repo, "check-repo.py must run store/binary boundary readiness")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "store/binary boundary checker must run after entry review",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_dependencies()
    assert_artifact_schema_fixture()
    assert_boundary_readiness(fixture)
    assert_store_and_reader_contracts(fixture)
    assert_validation_and_public_url_policy(fixture)
    assert_failures_and_runtime_plan(fixture)
    assert_forbidden_artifacts(fixture)
    assert_execution_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact store / binary reader boundary readiness v1 checks passed.")


if __name__ == "__main__":
    main()
