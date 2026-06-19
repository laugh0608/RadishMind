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
)


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-backend-adapter-readiness-evidence-v1.json"
HANDSHAKE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json"
ARTIFACT_RETURN_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
SAFETY_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-safety-runbook-evidence-v1.json"
EVAL_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-eval-manifest-v0.json"
INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
INTENT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
BACKEND_REQUEST_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-backend-request-basic.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "image_backend_adapter_implemented",
    "image_backend_adapter_ready",
    "image_backend_client_ready",
    "backend_profile_registry_ready",
    "credential_resolver_ready",
    "model_dir_resolver_ready",
    "endpoint_health_probe_ready",
    "real_backend_call_ready",
    "image_generation_ready",
    "image_pixel_quality_evaluated",
    "model_download_ready",
    "artifact_upload_ready",
    "artifact_store_ready",
    "production_artifact_storage_ready",
    "public_artifact_url_ready",
    "copilot_response_artifact_runtime_ready",
    "runtime_mapping_ready",
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
    "image-artifact-return-runbook-evidence-v1",
    "image-safety-runbook-evidence-v1",
    "image-generation-eval-manifest-v0",
    "contracts/image-generation-backend-request.schema.json",
    "contracts/image-generation-artifact.schema.json",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "backend_adapter_created_in_this_slice",
    "backend_client_created_in_this_slice",
    "backend_profile_registry_created_in_this_slice",
    "credential_resolver_created_in_this_slice",
    "model_dir_resolver_created_in_this_slice",
    "endpoint_health_probe_created_in_this_slice",
    "backend_queue_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "copilot_response_runtime_mapping_created_in_this_slice",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "model_download_allowed_now",
    "model_dir_read_allowed_now",
    "credential_resolution_allowed_now",
    "endpoint_health_probe_allowed_now",
    "artifact_upload_allowed_now",
    "public_url_allowed_now",
    "production_storage_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_CONTRACT_INPUT_FIELDS = {
    ("backend_request", "request_id"),
    ("backend_request", "intent_id"),
    ("backend_request", "backend.id"),
    ("backend_request", "backend.model"),
    ("backend_request", "backend.adapter_profile"),
    ("backend_request", "prompt.positive"),
    ("backend_request", "prompt.negative"),
    ("backend_request", "output.width"),
    ("backend_request", "output.height"),
    ("backend_request", "output.count"),
    ("backend_request", "output.format"),
    ("backend_request", "parameters.seed"),
    ("backend_request", "parameters.steps"),
    ("backend_request", "parameters.guidance_scale"),
    ("backend_request", "safety.gate"),
    ("backend_request", "safety.requires_confirmation"),
    ("backend_request", "safety.risk_level"),
    ("backend_request", "trace.source_request_id"),
    ("backend_request", "trace.trace_ids"),
    ("artifact", "artifact.uri"),
    ("artifact", "artifact.sha256"),
    ("artifact", "generation.backend_id"),
    ("artifact", "generation.model"),
    ("artifact", "provenance.trace_ids"),
}
EXPECTED_FORBIDDEN_FIELDS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "backend_auth_token",
    "backend_secret_value",
    "resolved_model_path",
    "signed_public_url",
    "storage_write_result",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
}
EXPECTED_PREREQUISITES = {
    "backend_profile_contract",
    "credential_reference_policy",
    "endpoint_or_model_dir_binding",
    "request_timeout_budget",
    "safety_gate_integration",
    "response_artifact_metadata_mapping",
    "artifact_hash_validation",
    "failure_envelope_mapping",
    "offline_contract_smoke",
}
EXPECTED_PROFILE_IDS = {
    "image-backend:contract-fixture",
    "image-backend:local-model-dir",
    "image-backend:http-service",
}
EXPECTED_GATE_CASES = {
    "profile_binding_missing": "image_backend_profile_missing",
    "credential_reference_missing": "image_backend_credential_missing",
    "model_dir_missing": "image_backend_model_dir_missing",
    "endpoint_unavailable": "image_backend_endpoint_unavailable",
    "timeout_budget_missing": "image_backend_timeout",
    "safety_gate_not_approved": "image_backend_safety_gate_blocked",
    "response_metadata_invalid": "image_backend_invalid_artifact_metadata",
    "artifact_hash_missing_or_mismatch": "image_backend_artifact_hash_mismatch",
}
EXPECTED_FAILURE_CODES = {
    "image_backend_profile_missing",
    "image_backend_credential_missing",
    "image_backend_model_dir_missing",
    "image_backend_endpoint_unavailable",
    "image_backend_timeout",
    "image_backend_safety_gate_blocked",
    "image_backend_invalid_artifact_metadata",
    "image_backend_artifact_hash_mismatch",
    "image_backend_response_untrusted",
}
EXPECTED_SMOKE_ASSERTIONS = {
    "validate_backend_request_schema",
    "consume_safety_runbook_gate",
    "map_backend_failure_to_fail_closed_code",
    "validate_artifact_metadata_schema",
    "validate_artifact_hash",
    "reject_provider_raw_response",
    "reject_pixel_payload",
    "preserve_provenance_trace",
}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "adapters/image/image_backend_adapter.py",
    "services/runtime/image_backend_adapter.py",
    "services/runtime/image_backend_client.py",
    "services/runtime/image_backend_profile_resolver.py",
    "services/platform/internal/httpapi/image_backend.go",
    "contracts/image-backend-adapter.schema.json",
    "apps/radishmind-web/src/features/image-generation/ImageBackendAdapterPanel.tsx",
    "scripts/run-image-backend-adapter-smoke.py",
    "deploy/image-backend-adapter.yaml",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageBackendAdapter",
    "ImageBackendClient",
    "ImageBackendProfileResolver",
    "ImageBackendCredentialResolver",
    "image_backend_adapter",
    "image_backend_client",
    "image_backend_profile_resolver",
    "IMAGE_BACKEND_ENDPOINT_URL",
    "IMAGE_BACKEND_CREDENTIAL_REF",
    "IMAGE_BACKEND_MODEL_DIR",
    "ImageBackendAdapterPanel",
    "runImageBackendAdapterSmoke",
    "callImageGenerationBackend",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_resolve_credentials",
    "does_not_read_model_dir",
    "does_not_probe_endpoint",
    "does_not_upload_artifacts",
    "does_not_write_production_storage",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "credential_resolve_count=0",
    "model_dir_read_count=0",
    "endpoint_health_probe_count=0",
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
    "credential_resolution",
    "model_dir_read",
    "endpoint_health_probe",
    "image_pixel_generation",
    "model_download",
    "artifact_upload",
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


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_backend_adapter_readiness_evidence_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-backend-adapter-readiness-evidence-v1", "unexpected slice id")
    require(slice_info.get("track") == "Image Path", "unexpected track")
    require(
        slice_info.get("status") == "image_backend_adapter_readiness_defined",
        "image backend adapter readiness status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")
    missing_dependencies = sorted(EXPECTED_DEPENDENCIES - set(fixture.get("depends_on") or []))
    require(not missing_dependencies, f"missing dependencies: {missing_dependencies}")


def assert_upstream_evidence() -> None:
    handshake = load_json(HANDSHAKE_FIXTURE_PATH)
    handshake_slice = handshake.get("slice") or {}
    require(handshake_slice.get("status") == "image_adapter_handshake_safety_gate_defined", "handshake status drifted")
    policy = handshake.get("artifact_return_policy") or {}
    require(policy.get("status") == "metadata_reference_only", "handshake artifact policy drifted")
    require(policy.get("artifact_upload_allowed_now") is False, "handshake must still block artifact upload")

    artifact_return = load_json(ARTIFACT_RETURN_FIXTURE_PATH)
    return_slice = artifact_return.get("slice") or {}
    require(
        return_slice.get("status") == "image_artifact_return_runbook_evidence_defined",
        "artifact return status drifted",
    )
    return_failures = {str(row.get("failure_code") or "") for row in artifact_return.get("failure_taxonomy") or []}
    require("image_backend_unavailable" in return_failures, "artifact return must keep backend unavailable failure")
    require("image_artifact_hash_mismatch" in return_failures, "artifact return must keep hash mismatch failure")

    safety = load_json(SAFETY_FIXTURE_PATH)
    safety_slice = safety.get("slice") or {}
    require(safety_slice.get("status") == "image_safety_runbook_evidence_defined", "safety runbook status drifted")
    safety_failures = {str(row.get("failure_code") or "") for row in safety.get("failure_taxonomy") or []}
    require("image_backend_safety_gate_blocked" in safety_failures, "safety runbook must keep backend safety block")
    safety_execution = safety.get("execution_policy") or {}
    require(safety_execution.get("does_not_call_backend") is True, "safety runbook must not call backend")

    manifest = load_json(EVAL_MANIFEST_PATH)
    dimensions = set((manifest.get("scope") or {}).get("eval_dimensions") or [])
    for dimension in ("backend_request_mapping", "artifact_metadata", "safety_gate", "provenance"):
        require(dimension in dimensions, f"image eval manifest must keep {dimension}")
    execution = manifest.get("execution_policy") or {}
    for field in ("does_not_call_backend", "does_not_generate_images", "does_not_download_models"):
        require(execution.get(field) is True, f"eval manifest {field} must remain true")


def assert_existing_contract_chain() -> None:
    intent_schema = load_json_document(INTENT_SCHEMA_PATH)
    backend_schema = load_json_document(BACKEND_REQUEST_SCHEMA_PATH)
    artifact_schema = load_json_document(ARTIFACT_SCHEMA_PATH)
    intent = load_json_document(INTENT_FIXTURE_PATH)
    backend_request = load_json_document(BACKEND_REQUEST_FIXTURE_PATH)
    artifact = load_json_document(ARTIFACT_FIXTURE_PATH)

    for schema in (intent_schema, backend_schema, artifact_schema):
        jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(intent, intent_schema)
    jsonschema.validate(backend_request, backend_schema)
    jsonschema.validate(artifact, artifact_schema)

    check_image_generation_intent(intent)
    check_backend_request(intent, backend_request, expected_gate="approved_for_backend")
    check_artifact(intent, backend_request, artifact)
    require(value_at_path(backend_request, "backend.id"), "backend id required")
    require(value_at_path(backend_request, "backend.model"), "backend model required")
    require(value_at_path(backend_request, "backend.adapter_profile"), "backend adapter profile required")
    require(value_at_path(backend_request, "safety.gate") == "approved_for_backend", "basic backend gate drifted")
    require(value_at_path(artifact, "generation.backend_id") == value_at_path(backend_request, "backend.id"), "backend id trace")
    require(value_at_path(artifact, "generation.model") == value_at_path(backend_request, "backend.model"), "model trace")
    uri = str(value_at_path(artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "artifact URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "artifact URI must not become public URL")
    require(str(value_at_path(artifact, "artifact.sha256")).strip(), "artifact hash required")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("adapter_readiness_boundary") or {}
    require(
        boundary.get("status") == "backend_adapter_readiness_defined_no_backend_client",
        "adapter readiness boundary status drifted",
    )
    require(boundary.get("decision") == "adapter_implementation_gate_only", "adapter readiness decision drifted")
    require(
        boundary.get("source_backend_request_schema") == "contracts/image-generation-backend-request.schema.json",
        "backend request schema ref",
    )
    require(boundary.get("source_artifact_schema") == "contracts/image-generation-artifact.schema.json", "artifact schema ref")
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_contract_inputs(fixture: dict[str, Any]) -> None:
    contract_inputs = fixture.get("adapter_contract_input_fields") or {}
    require(contract_inputs.get("status") == "defined_not_executable", "contract input status drifted")
    require(
        contract_inputs.get("runtime_backend_adapter_required_now") is False,
        "runtime backend adapter must not be required now",
    )
    configured = {
        (str(row.get("source") or ""), str(row.get("path") or ""))
        for row in contract_inputs.get("required_fields") or []
        if isinstance(row, dict)
    }
    require(EXPECTED_CONTRACT_INPUT_FIELDS.issubset(configured), "adapter contract input fields drifted")
    require(EXPECTED_FORBIDDEN_FIELDS.issubset(set(contract_inputs.get("forbidden_fields") or [])), "forbidden fields drifted")

    documents = {
        "backend_request": load_json(BACKEND_REQUEST_FIXTURE_PATH),
        "artifact": load_json(ARTIFACT_FIXTURE_PATH),
    }
    for source, dotted_path in EXPECTED_CONTRACT_INPUT_FIELDS:
        value_at_path(documents[source], dotted_path)


def assert_prerequisites_profiles_and_gates(fixture: dict[str, Any]) -> None:
    prerequisites = {
        str(row.get("gate_id") or ""): row
        for row in fixture.get("readiness_prerequisites") or []
        if isinstance(row, dict)
    }
    require(set(prerequisites) == EXPECTED_PREREQUISITES, "readiness prerequisites drifted")
    for gate_id, row in prerequisites.items():
        require(row.get("gate_status") == "required_before_implementation", f"{gate_id} gate status drifted")
        require(row.get("satisfied_now") is False, f"{gate_id} must remain unsatisfied")
        require(row.get("implementation_allowed_now") is False, f"{gate_id} must not allow implementation")

    profiles = {
        str(row.get("profile_id") or ""): row
        for row in fixture.get("backend_profile_matrix") or []
        if isinstance(row, dict)
    }
    require(set(profiles) == EXPECTED_PROFILE_IDS, "backend profile matrix drifted")
    require(profiles["image-backend:local-model-dir"].get("model_dir_required") is True, "local model dir profile drifted")
    require(profiles["image-backend:http-service"].get("endpoint_required") is True, "http endpoint profile drifted")
    require(profiles["image-backend:http-service"].get("credential_required") is True, "http credential profile drifted")
    for profile_id, row in profiles.items():
        require(row.get("current_binding_ready") is False, f"{profile_id} must not be binding-ready")
        require(row.get("real_backend_call_allowed_now") is False, f"{profile_id} backend call must stay blocked")

    gates = {
        str(row.get("case_id") or ""): row
        for row in fixture.get("readiness_gate_matrix") or []
        if isinstance(row, dict)
    }
    require(set(gates) == set(EXPECTED_GATE_CASES), "readiness gate matrix drifted")
    for case_id, expected_failure in EXPECTED_GATE_CASES.items():
        row = gates[case_id]
        require(row.get("expected_failure_code") == expected_failure, f"{case_id} failure code drifted")
        require(row.get("adapter_implementation_allowed_now") is False, f"{case_id} must not allow adapter implementation")
        require(row.get("retry_or_fallback_allowed_now") is False, f"{case_id} retry/fallback must stay blocked")


def assert_failures_and_future_smoke(fixture: dict[str, Any]) -> None:
    failures = {
        str(row.get("failure_code") or ""): row
        for row in fixture.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    for failure_code, row in failures.items():
        require(row.get("retry_or_fallback_allowed_now") is False, f"{failure_code} retry must stay blocked")
        require(row.get("artifact_reference_returned") is False, f"{failure_code} must not return artifact reference")

    smoke = fixture.get("future_adapter_smoke_contract") or {}
    require(smoke.get("status") == "readiness_only", "future adapter smoke status drifted")
    require(smoke.get("runtime_smoke_required_now") is False, "runtime smoke must not be required now")
    require(smoke.get("backend_call_allowed_now") is False, "future smoke must not call backend now")
    for relative_path in smoke.get("required_inputs") or []:
        require((REPO_ROOT / str(relative_path)).exists(), f"missing future smoke input: {relative_path}")
    require(EXPECTED_SMOKE_ASSERTIONS.issubset(set(smoke.get("required_assertions") or [])), "future smoke assertions drifted")


def assert_execution_and_side_effects(fixture: dict[str, Any]) -> None:
    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "missing_profile_promoted_to_success_allowed",
        "missing_credential_promoted_to_backend_allowed",
        "missing_model_dir_promoted_to_backend_allowed",
        "endpoint_unavailable_promoted_to_retry_loop_allowed",
        "safety_gate_blocked_promoted_to_backend_allowed",
        "invalid_artifact_metadata_promoted_to_reference_allowed",
        "hash_mismatch_promoted_to_reference_allowed",
        "timeout_promoted_to_success_allowed",
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
    for root in (REPO_ROOT / "services", REPO_ROOT / "apps", REPO_ROOT / "adapters"):
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.suffix not in {".go", ".py", ".ts", ".tsx"}:
                continue
            text = path.read_text(encoding="utf-8")
            for literal in configured:
                require(literal not in text, f"{path.relative_to(REPO_ROOT)} must not introduce {literal!r}")


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
    previous_checker = "check-image-safety-runbook-evidence-v1.py"
    current_checker = "check-image-backend-adapter-readiness-evidence-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image backend adapter readiness evidence")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "backend adapter readiness must run after safety")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_upstream_evidence()
    assert_existing_contract_chain()
    assert_boundary(fixture)
    assert_contract_inputs(fixture)
    assert_prerequisites_profiles_and_gates(fixture)
    assert_failures_and_future_smoke(fixture)
    assert_execution_and_side_effects(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_references_and_check_repo(fixture)
    print("image backend adapter readiness evidence v1 checks passed.")


if __name__ == "__main__":
    main()
