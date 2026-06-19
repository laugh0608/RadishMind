#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-v1.json"
ENTRY_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-entry-v1.json"
PLAN_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-implementation-plan-v1.json"
BOUNDARY_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-store-binary-reader-boundary-readiness-v1.json"
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
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"
TASK_CARD_PATH = "docs/task-cards/image-artifact-runtime-mapper-implementation-v1-plan.md"
RUNTIME_MAPPER_MODULE_PATH = "services/runtime/image_artifact_runtime_mapper.py"
RUNTIME_MAPPER_RUNTIME_IMPLEMENTATION_PATH = (
    REPO_ROOT / "scripts/checks/fixtures/image-artifact-runtime-mapper-runtime-implementation-v1.json"
)

DEPENDENCY_STATUS_BY_SLICE = {
    "image-artifact-runtime-mapper-implementation-entry-v1": (
        ENTRY_PATH,
        "image_artifact_runtime_mapper_implementation_entry_review_defined",
    ),
    "image-artifact-runtime-mapper-implementation-plan-v1": (
        PLAN_PATH,
        "image_artifact_runtime_mapper_implementation_plan_defined",
    ),
    "image-artifact-store-binary-reader-boundary-readiness-v1": (
        BOUNDARY_PATH,
        "image_artifact_store_binary_reader_boundary_readiness_defined",
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
RECONCILIATION_STATUS_FIELDS = {
    "image-artifact-runtime-mapper-implementation-entry-v1": "implementation_entry_status",
    "image-artifact-runtime-mapper-implementation-plan-v1": "implementation_plan_status",
    "image-artifact-store-binary-reader-boundary-readiness-v1": "store_binary_boundary_status",
    "image-artifact-runtime-mapping-readiness-v1": "runtime_mapping_readiness_status",
    "image-artifact-return-runbook-evidence-v1": "artifact_return_runbook_status",
    "image-safety-runbook-evidence-v1": "safety_runbook_status",
    "image-backend-adapter-readiness-evidence-v1": "backend_adapter_readiness_status",
}
EXPECTED_DEPENDENCIES = set(DEPENDENCY_STATUS_BY_SLICE) | {
    "contracts/image-generation-artifact.schema.json",
    "contracts/copilot-response.schema.json",
}
EXPECTED_FORBIDDEN_CLAIMS = {
    "runtime_mapper_implemented",
    "runtime_mapper_files_created",
    "runtime_mapping_execution_ready",
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
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "runtime_mapper_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "artifact_binary_reader_created_in_this_slice",
    "public_url_resolver_created_in_this_slice",
    "backend_adapter_created_in_this_slice",
    "copilot_response_schema_changed_in_this_slice",
    "artifact_store_implementation_task_card_allowed_next",
    "artifact_binary_reader_implementation_task_card_allowed_next",
    "public_url_resolver_implementation_task_card_allowed_next",
    "backend_adapter_implementation_task_card_allowed_next",
    "parallel_implementation_tracks_allowed_now",
    "runtime_mapping_execution_allowed_now",
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
EXPECTED_SCOPE_FALSE_FIELDS = {
    "copilot_response_schema_change_required_now",
    "artifact_store_required_now",
    "binary_reader_required_now",
    "public_url_required_now",
    "backend_adapter_required_now",
    "backend_call_allowed",
    "binary_payload_allowed",
    "provider_raw_dump_allowed",
}
EXPECTED_ARTIFACT_METADATA_FIELDS = {
    "artifact_id",
    "artifact.uri",
    "artifact.sha256",
    "artifact.mime_type",
    "artifact.width",
    "artifact.height",
    "artifact.format",
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
    "provenance.backend_request_id",
    "provenance.intent_id",
    "created_at",
}
EXPECTED_FORBIDDEN_INPUTS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "public_url",
    "signed_public_url",
    "storage_write_result",
    "binary_reader_result",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
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
EXPECTED_MAPPING_CASES = {
    "generated_not_required": (True, None),
    "generated_reviewed_pass": (True, None),
    "blocked_artifact_status": (False, "image_artifact_safety_blocked"),
    "failed_artifact_status": (False, "image_artifact_invalid_metadata"),
    "pending_review_artifact": (False, "image_artifact_safety_pending_review"),
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
EXPECTED_FORBIDDEN_TASK_CARDS = {
    "docs/task-cards/image-artifact-store-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-binary-reader-implementation-v1-plan.md",
    "docs/task-cards/image-artifact-public-url-resolver-implementation-v1-plan.md",
    "docs/task-cards/image-backend-adapter-implementation-v1-plan.md",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "services/runtime/image_artifact_runtime_mapper.py",
    "services/runtime/copilot_response_artifact_mapper.py",
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
EXPECTED_ABSENT_LITERALS = {
    "ImageArtifactRuntimeMapper",
    "CopilotResponseArtifactMapper",
    "ImageArtifactStore",
    "ImageArtifactBinaryReader",
    "ImageArtifactPublicUrlResolver",
    "ImageBackendAdapter",
    "image_artifact_runtime_mapper",
    "copilot_response_artifact_mapper",
    "image_artifact_store",
    "image_artifact_binary_reader",
    "image_artifact_public_url_resolver",
    "image_backend_adapter",
    "IMAGE_ARTIFACT_STORE_URL",
    "IMAGE_ARTIFACT_PUBLIC_BASE_URL",
    "IMAGE_ARTIFACT_SIGNED_URL_TTL",
    "ReadImageArtifactBinary",
    "ResolveImageArtifactPublicURL",
    "runImageArtifactRuntimeMapping",
    "callImageGenerationBackend",
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
    "task_card_promoted_to_runtime_code_without_dedicated_commit_allowed",
    "invalid_metadata_promoted_to_success_allowed",
    "missing_artifact_store_promoted_to_success_allowed",
    "missing_binary_reader_promoted_to_success_allowed",
    "blocked_artifact_promoted_to_success_allowed",
    "pending_review_promoted_to_success_allowed",
    "hash_mismatch_promoted_to_success_allowed",
    "mime_mismatch_promoted_to_success_allowed",
    "dimension_mismatch_promoted_to_success_allowed",
    "public_url_claim_promoted_to_success_allowed",
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
    "runtime_mapping_execution_count=0",
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
    "runtime_mapping_execution",
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


def nested(schema: dict[str, Any], *keys: str) -> Any:
    current: Any = schema
    for key in keys:
        require(isinstance(current, dict), f"missing schema node before {key}")
        current = current.get(key)
    return current


def assert_dependencies() -> dict[str, dict[str, Any]]:
    dependencies: dict[str, dict[str, Any]] = {}
    for slice_id, (path, expected_status) in DEPENDENCY_STATUS_BY_SLICE.items():
        document = load_json(path)
        status = (document.get("slice") or {}).get("status")
        require(status == expected_status, f"{slice_id} status drifted: {status}")
        dependencies[slice_id] = document

    artifact_schema = load_json(ARTIFACT_SCHEMA_PATH)
    artifact_fixture = load_json(ARTIFACT_FIXTURE_PATH)
    jsonschema.validate(artifact_fixture, artifact_schema)
    require(artifact_fixture["kind"] == "image_generation_artifact", "artifact fixture kind drifted")
    require(artifact_fixture["artifact"]["uri"].startswith("artifact://"), "artifact fixture URI must use artifact://")
    require(artifact_fixture["artifact"]["sha256"] == "a" * 64, "artifact fixture sha256 drifted")

    copilot_schema = load_json(COPILOT_RESPONSE_SCHEMA_PATH)
    citation_kind = nested(copilot_schema, "$defs", "citation", "properties", "kind")
    require("artifact" in set(citation_kind.get("enum") or []), "CopilotResponse citation.kind must allow artifact")
    citation_properties = nested(copilot_schema, "$defs", "citation", "properties")
    require({"locator", "source_uri"}.issubset(set(citation_properties)), "citation locator/source_uri missing")
    return dependencies


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "fixture schema_version drifted")
    require(fixture.get("kind") == "image_artifact_runtime_mapper_implementation_v1", "fixture kind drifted")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-artifact-runtime-mapper-implementation-v1", "slice id drifted")
    require(slice_info.get("track") == "Image Path", "slice track drifted")
    require(
        slice_info.get("status") == "image_artifact_runtime_mapper_implementation_task_card_defined",
        "slice status drifted",
    )
    require(set(slice_info.get("does_not_claim") or []) == EXPECTED_FORBIDDEN_CLAIMS, "forbidden claims drifted")
    require(set(fixture.get("depends_on") or []) == EXPECTED_DEPENDENCIES, "dependency list drifted")


def assert_entry_dependency(entry: dict[str, Any]) -> None:
    boundary = entry.get("implementation_entry_boundary") or {}
    require(
        boundary.get("decision") == "runtime_mapper_implementation_task_card_allowed_next",
        "entry decision must allow implementation task card next",
    )
    require(boundary.get("implementation_task_card_created_in_this_slice") is False, "entry must not create task card")
    require(boundary.get("runtime_mapper_runtime_code_allowed_now") is False, "entry must not allow runtime code")

    selected = entry.get("selected_track_contract") or {}
    require(selected.get("next_task_card") == TASK_CARD_PATH, "entry selected task card drifted")
    require(selected.get("selected_candidate") == "artifact_runtime_mapper_implementation", "entry selected track drifted")

    constraints = entry.get("next_implementation_constraints") or {}
    require(constraints.get("artifact_uri_scheme") == "artifact://", "entry artifact URI scheme drifted")
    for field in (
        "input_must_be_metadata_only",
        "must_preserve_sha256",
        "must_preserve_mime_type",
        "must_preserve_dimensions",
        "must_preserve_safety_review",
        "must_preserve_provenance",
        "blocked_failed_pending_review_fail_closed",
    ):
        require(constraints.get(field) is True, f"entry constraints.{field} must remain true")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("implementation_task_card_boundary") or {}
    require(boundary.get("status") == "implementation_task_card_defined_no_runtime_code", "boundary status drifted")
    require(
        boundary.get("decision") == "metadata_only_runtime_mapper_runtime_code_allowed_next",
        "boundary decision drifted",
    )
    require(
        boundary.get("current_development_mode") == "metadata_only_runtime_mapper_runtime_code_next",
        "development mode drifted",
    )
    require(
        boundary.get("source_entry_slice") == "image-artifact-runtime-mapper-implementation-entry-v1",
        "source entry slice drifted",
    )
    require(boundary.get("task_card_created_in_this_slice") is True, "this slice must create task card")
    require(
        boundary.get("runtime_mapper_runtime_code_allowed_after_this_slice") is True,
        "runtime mapper code should be allowed after this task card",
    )
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"implementation_task_card_boundary.{field} must remain false")
    require(
        boundary.get("next_runtime_code_scope") == "metadata_only_runtime_mapper_only",
        "next runtime code scope drifted",
    )


def assert_scope(fixture: dict[str, Any]) -> None:
    scope = fixture.get("implementation_scope") or {}
    require(scope.get("status") == "metadata_only_runtime_mapper_scope_defined", "scope status drifted")
    require(scope.get("selected_track") == "artifact_runtime_mapper_implementation", "selected track drifted")
    require(scope.get("future_module") == "services/runtime/image_artifact_runtime_mapper.py", "future module drifted")
    require(scope.get("input_kind") == "image_generation_artifact", "input kind drifted")
    require(scope.get("artifact_uri_scheme") == "artifact://", "scope artifact URI scheme drifted")
    require(scope.get("hash_algorithm") == "sha256", "scope hash algorithm drifted")
    require(scope.get("input_must_be_metadata_only") is True, "scope must require metadata-only input")
    require(scope.get("output_must_be_metadata_only") is True, "scope must require metadata-only output")
    require(scope.get("no_side_effects_required") is True, "scope must require no side effects")
    for field in EXPECTED_SCOPE_FALSE_FIELDS:
        require(scope.get(field) is False, f"implementation_scope.{field} must remain false")


def assert_input_contract(fixture: dict[str, Any], plan: dict[str, Any]) -> None:
    contract = fixture.get("input_metadata_contract") or {}
    require(
        contract.get("status") == "metadata_contract_required_for_future_runtime_code",
        "input contract status drifted",
    )
    require(set(contract.get("required_artifact_metadata_fields") or []) == EXPECTED_ARTIFACT_METADATA_FIELDS, "metadata fields drifted")
    require(set(contract.get("forbidden_inputs") or []) == EXPECTED_FORBIDDEN_INPUTS, "forbidden inputs drifted")
    require(contract.get("required_uri_scheme") == "artifact://", "required URI scheme drifted")
    require(contract.get("hash_algorithm") == "sha256", "hash algorithm drifted")
    require(contract.get("dimensions_must_be_positive") is True, "dimensions must remain positive")
    require(contract.get("safety_review_required_for_success") is True, "safety review requirement drifted")
    require(contract.get("provenance_required_for_success") is True, "provenance requirement drifted")

    plan_input = plan.get("mapper_input_contract") or {}
    require(
        set(plan_input.get("required_artifact_metadata_fields") or []) == EXPECTED_ARTIFACT_METADATA_FIELDS,
        "plan metadata fields drifted",
    )
    require(set(plan_input.get("forbidden_inputs") or []) == EXPECTED_FORBIDDEN_INPUTS, "plan forbidden inputs drifted")


def assert_target_response_contract(fixture: dict[str, Any], plan: dict[str, Any]) -> None:
    target = fixture.get("target_response_reference_contract") or {}
    require(
        target.get("status") == "future_response_reference_contract_required_for_future_runtime_code",
        "target response contract status drifted",
    )
    require(target.get("target_citation_kind") == "artifact", "target citation kind drifted")
    require(set(target.get("target_citation_fields") or []) == EXPECTED_CITATION_FIELDS, "citation fields drifted")
    require(
        set(target.get("target_metadata_reference_fields") or []) == EXPECTED_METADATA_REFERENCE_FIELDS,
        "metadata reference fields drifted",
    )
    require(target.get("copilot_response_schema_change_required_now") is False, "response schema must not change")
    require(target.get("citation_kind_supported_by_existing_schema") is True, "citation kind support drifted")
    require(target.get("metadata_reference_schema_materialized_now") is False, "metadata reference schema must not materialize")
    require(target.get("blocked_failed_pending_review_success_reference_allowed") is False, "blocked states must fail closed")

    plan_target = plan.get("target_response_reference_plan") or {}
    require(set(plan_target.get("target_citation_fields") or []) == EXPECTED_CITATION_FIELDS, "plan citation fields drifted")
    require(
        set(plan_target.get("target_metadata_reference_fields") or []) == EXPECTED_METADATA_REFERENCE_FIELDS,
        "plan metadata reference fields drifted",
    )


def assert_mapping_case_requirements(fixture: dict[str, Any], plan: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("mapping_case_requirements") or [], "case_id")
    require(set(rows) == set(EXPECTED_MAPPING_CASES), "mapping case requirements drifted")
    plan_rows = rows_by_id(plan.get("mapping_case_plan") or [], "case_id")

    for case_id, (success_allowed, failure_code) in EXPECTED_MAPPING_CASES.items():
        row = rows[case_id]
        require(row.get("source_case") == "image-artifact-runtime-mapping-readiness-v1", f"{case_id} source drifted")
        require(row.get("success_response_reference_allowed") is success_allowed, f"{case_id} success policy drifted")
        require(row.get("runtime_mapping_implemented_now") is False, f"{case_id} must not be implemented now")
        require(row.get("runtime_test_required_next") is True, f"{case_id} must require runtime test next")
        if success_allowed:
            require(row.get("target_citation_kind") == "artifact", f"{case_id} target citation kind drifted")
            require("expected_failure_code" not in row, f"{case_id} must not have failure code")
        else:
            require(row.get("expected_failure_code") == failure_code, f"{case_id} failure code drifted")

        plan_row = plan_rows[case_id]
        require(
            plan_row.get("success_response_reference_allowed") is success_allowed,
            f"{case_id} plan success policy drifted",
        )


def assert_fail_closed_requirements(fixture: dict[str, Any], plan: dict[str, Any]) -> None:
    rows = rows_by_id(fixture.get("fail_closed_requirements") or [], "case_id")
    require(set(rows) == set(EXPECTED_FAIL_CLOSED_CASES), "fail-closed requirements drifted")
    plan_rows = rows_by_id(plan.get("fail_closed_plan") or [], "case_id")

    for case_id, failure_code in EXPECTED_FAIL_CLOSED_CASES.items():
        row = rows[case_id]
        require(row.get("expected_failure_code") == failure_code, f"{case_id} failure code drifted")
        require(row.get("success_response_reference_allowed") is False, f"{case_id} must not allow success")
        require(row.get("fail_closed") is True, f"{case_id} must fail closed")
        require(row.get("runtime_test_required_next") is True, f"{case_id} must require runtime test next")

        plan_row = plan_rows[case_id]
        require(plan_row.get("expected_failure_code") == failure_code, f"{case_id} plan failure code drifted")
        require(plan_row.get("fail_closed") is True, f"{case_id} plan must fail closed")


def assert_dependency_reconciliation(fixture: dict[str, Any], dependencies: dict[str, dict[str, Any]]) -> None:
    reconciliation = fixture.get("dependency_reconciliation") or {}
    require(
        reconciliation.get("status") == "implementation_task_card_dependencies_reconciled",
        "dependency reconciliation status drifted",
    )
    for slice_id, field in RECONCILIATION_STATUS_FIELDS.items():
        document = dependencies[slice_id]
        status = (document.get("slice") or {}).get("status")
        require(reconciliation.get(field) == status, f"{field} drifted")

    require(reconciliation.get("implementation_entry_decision") == "runtime_mapper_implementation_task_card_allowed_next", "entry decision drifted")
    require(reconciliation.get("success_mapping_cases") == 2, "success case count drifted")
    require(reconciliation.get("blocked_mapping_cases") == 3, "blocked case count drifted")
    require(reconciliation.get("fail_closed_requirement_cases") == 14, "fail-closed case count drifted")
    require(reconciliation.get("runtime_mapper_runtime_code_allowed_next") is True, "runtime code should be allowed next")
    require(reconciliation.get("runtime_mapper_runtime_code_created_now") is False, "runtime code must not be created now")
    for field in (
        "artifact_store_ready_now",
        "artifact_binary_reader_ready_now",
        "public_url_resolver_ready_now",
        "backend_adapter_ready_now",
        "copilot_response_schema_change_allowed_now",
    ):
        require(reconciliation.get(field) is False, f"{field} must remain false")


def assert_future_runtime_test_plan(fixture: dict[str, Any]) -> None:
    plan = fixture.get("future_runtime_test_plan") or {}
    require(plan.get("status") == "runtime_mapper_tests_required_next", "runtime test plan status drifted")
    require(set(plan.get("success_tests_required") or []) == {"generated_not_required", "generated_reviewed_pass"}, "success tests drifted")
    require(
        set(plan.get("blocked_tests_required") or [])
        == {"blocked_artifact_status", "failed_artifact_status", "pending_review_artifact"},
        "blocked tests drifted",
    )
    require(set(plan.get("fail_closed_tests_required") or []) == set(EXPECTED_FAIL_CLOSED_CASES), "fail-closed tests drifted")
    require(
        {"copilot_response_schema_unchanged", "artifact_schema_fixture_still_valid"}.issubset(
            set(plan.get("schema_guard_tests_required") or [])
        ),
        "schema guard tests drifted",
    )
    require(
        {
            "backend_call_count=0",
            "artifact_binary_read_count=0",
            "artifact_store_lookup_count=0",
            "public_url_resolution_count=0",
            "artifact_upload_count=0",
        }.issubset(set(plan.get("side_effect_tests_required") or [])),
        "side-effect tests drifted",
    )


def assert_forbidden_artifacts(fixture: dict[str, Any]) -> None:
    task_cards = rows_by_id(fixture.get("forbidden_task_card_matrix") or [], "path")
    require(set(task_cards) == EXPECTED_FORBIDDEN_TASK_CARDS, "forbidden task cards drifted")
    for relative_path, row in task_cards.items():
        require(row.get("created_in_this_slice") is False, f"{relative_path} must not be created")
        require(row.get("allowed_next") is False, f"{relative_path} must not be allowed next")
        require(not (REPO_ROOT / relative_path).exists(), f"{relative_path} must not exist")

    artifacts = rows_by_id(fixture.get("forbidden_artifact_matrix") or [], "path")
    require(set(artifacts) == EXPECTED_FORBIDDEN_ARTIFACTS, "forbidden artifacts drifted")
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
    previous_checker = "check-image-artifact-runtime-mapper-implementation-entry-v1.py"
    current_checker = "check-image-artifact-runtime-mapper-implementation-v1.py"
    require(current_checker in check_repo, "check-repo.py must run runtime mapper implementation task card checker")
    require(
        check_repo.index(previous_checker) < check_repo.index(current_checker),
        "runtime mapper implementation checker must run after implementation entry",
    )


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    dependencies = assert_dependencies()
    assert_slice(fixture)
    assert_entry_dependency(dependencies["image-artifact-runtime-mapper-implementation-entry-v1"])
    assert_boundary(fixture)
    assert_scope(fixture)
    assert_input_contract(fixture, dependencies["image-artifact-runtime-mapper-implementation-plan-v1"])
    assert_target_response_contract(fixture, dependencies["image-artifact-runtime-mapper-implementation-plan-v1"])
    assert_mapping_case_requirements(fixture, dependencies["image-artifact-runtime-mapper-implementation-plan-v1"])
    assert_fail_closed_requirements(fixture, dependencies["image-artifact-runtime-mapper-implementation-plan-v1"])
    assert_dependency_reconciliation(fixture, dependencies)
    assert_future_runtime_test_plan(fixture)
    assert_forbidden_artifacts(fixture)
    assert_execution_and_side_effects(fixture)
    assert_references_and_check_repo(fixture)
    print("image artifact runtime mapper implementation v1 checks passed.")


if __name__ == "__main__":
    main()
