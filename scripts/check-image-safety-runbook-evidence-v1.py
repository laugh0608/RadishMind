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


FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-safety-runbook-evidence-v1.json"
HANDSHAKE_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-adapter-handshake-safety-gate-v1.json"
ARTIFACT_RETURN_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-artifact-return-runbook-evidence-v1.json"
EVAL_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-eval-manifest-v0.json"
INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
INTENT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
BACKEND_REQUEST_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-backend-request-basic.json"
ARTIFACT_FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"
CHECK_REPO_PATH = REPO_ROOT / "scripts/check-repo.py"

EXPECTED_FORBIDDEN_CLAIMS = {
    "image_safety_runtime_ready",
    "image_safety_classifier_ready",
    "runtime_policy_engine_ready",
    "moderation_provider_ready",
    "image_backend_client_ready",
    "real_backend_call_ready",
    "image_generation_ready",
    "image_pixel_quality_evaluated",
    "model_download_ready",
    "artifact_upload_ready",
    "production_artifact_storage_ready",
    "public_artifact_url_ready",
    "copilot_response_artifact_runtime_ready",
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
    "image-generation-eval-manifest-v0",
    "contracts/image-generation-intent.schema.json",
    "contracts/image-generation-backend-request.schema.json",
    "contracts/image-generation-artifact.schema.json",
}
EXPECTED_BOUNDARY_FALSE_FIELDS = {
    "copilot_response_schema_changed_in_this_slice",
    "image_safety_runtime_created_in_this_slice",
    "image_safety_classifier_created_in_this_slice",
    "runtime_policy_engine_created_in_this_slice",
    "moderation_provider_created_in_this_slice",
    "image_backend_client_created_in_this_slice",
    "artifact_store_created_in_this_slice",
    "real_backend_call_allowed_now",
    "image_generation_allowed_now",
    "model_download_allowed_now",
    "artifact_upload_allowed_now",
    "public_url_allowed_now",
    "production_storage_allowed_now",
    "dev_server_started_in_this_slice",
}
EXPECTED_POLICY_INPUT_FIELDS = {
    ("intent", "prompt.positive"),
    ("intent", "prompt.negative"),
    ("intent", "constraints.must_include"),
    ("intent", "constraints.must_avoid"),
    ("intent", "safety.requires_confirmation"),
    ("intent", "safety.risk_level"),
    ("intent", "safety.review_notes"),
    ("backend_request", "safety.gate"),
    ("backend_request", "safety.requires_confirmation"),
    ("backend_request", "safety.risk_level"),
    ("backend_request", "trace.trace_ids"),
    ("artifact", "safety.risk_level"),
    ("artifact", "safety.requires_confirmation"),
    ("artifact", "safety.review_status"),
    ("artifact", "safety.review_notes"),
    ("artifact", "artifact.sha256"),
    ("artifact", "provenance.trace_ids"),
}
EXPECTED_FORBIDDEN_FIELDS = {
    "pixel_payload",
    "base64_image",
    "provider_raw_response",
    "external_moderation_result",
    "signed_public_url",
    "storage_write_result",
    "human_confirmation_decision",
    "executor_ref",
    "writeback_ref",
    "replay_ref",
}
EXPECTED_RUNBOOK_STEPS = [
    "core_intent_safety_precheck",
    "adapter_safety_gate_before_backend_request",
    "block_confirmation_required_or_high_risk",
    "backend_request_safety_trace_review",
    "artifact_metadata_safety_review",
    "route_failed_or_blocked_safety_metadata",
    "defer_runtime_policy_engine_and_moderation_provider",
]
EXPECTED_DECISION_CASES = {
    "low_risk_no_confirmation",
    "confirmation_required",
    "high_risk_or_policy_unknown",
    "backend_unavailable_after_gate",
    "artifact_review_pending",
    "artifact_review_blocked",
}
EXPECTED_FAILURE_CODES = {
    "image_prompt_policy_unknown",
    "image_intent_requires_confirmation",
    "image_intent_high_risk",
    "image_backend_safety_gate_blocked",
    "image_backend_unavailable",
    "image_artifact_safety_pending_review",
    "image_artifact_safety_blocked",
}
EXPECTED_ARTIFACT_REVIEW_FIELDS = {
    "safety.risk_level",
    "safety.requires_confirmation",
    "safety.review_status",
    "safety.review_notes",
    "artifact.sha256",
    "artifact.uri",
    "provenance.trace_ids",
}
EXPECTED_REVIEW_STATUSES = {"not_required", "pending_review", "reviewed_pass", "blocked"}
EXPECTED_BLOCKED_REVIEW_STATUSES = {"pending_review", "blocked"}
EXPECTED_FORBIDDEN_ARTIFACTS = {
    "adapters/image/image_safety_policy.py",
    "services/runtime/image_safety_runtime.py",
    "services/runtime/image_safety_classifier.py",
    "services/platform/internal/httpapi/image_safety.go",
    "contracts/image-safety-runbook.schema.json",
    "apps/radishmind-web/src/features/image-generation/ImageSafetyPanel.tsx",
    "scripts/run-image-safety-classifier.py",
    "deploy/image-safety-service.yaml",
}
EXPECTED_ABSENT_LITERALS = {
    "ImageSafetyRuntime",
    "ImageSafetyClassifier",
    "ImageSafetyPolicyEngine",
    "ImageModerationProvider",
    "image_safety_runtime",
    "image_safety_classifier",
    "image_safety_policy_engine",
    "IMAGE_SAFETY_PROVIDER_URL",
    "IMAGE_SAFETY_MODEL_DIR",
    "ImageSafetyPanel",
    "runImageSafetyClassifier",
    "callImageModerationProvider",
}
EXPECTED_EXECUTION_TRUE_FIELDS = {
    "does_not_call_backend",
    "does_not_call_moderation_provider",
    "does_not_generate_images",
    "does_not_download_models",
    "does_not_upload_artifacts",
    "does_not_write_production_storage",
    "does_not_start_dev_server",
    "does_not_use_browser_plugin",
}
EXPECTED_SIDE_EFFECT_COUNTERS = {
    "backend_call_count=0",
    "moderation_provider_call_count=0",
    "policy_engine_call_count=0",
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
    "moderation_provider_call",
    "runtime_policy_engine_call",
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


def build_blocked_backend_request(
    backend_request: dict[str, Any],
    *,
    requires_confirmation: bool,
    risk_level: str,
    review_note: str,
) -> dict[str, Any]:
    return {
        **backend_request,
        "safety": {
            **backend_request["safety"],
            "gate": "blocked_requires_confirmation",
            "requires_confirmation": requires_confirmation,
            "risk_level": risk_level,
            "review_notes": [review_note],
        },
    }


def build_high_risk_intent(intent: dict[str, Any]) -> dict[str, Any]:
    return {
        **intent,
        "safety": {
            **intent["safety"],
            "requires_confirmation": True,
            "risk_level": "high",
            "review_notes": ["blocked by image safety runbook before backend submission"],
        },
    }


def build_blocked_artifact(artifact: dict[str, Any]) -> dict[str, Any]:
    return {
        **artifact,
        "status": "blocked",
        "safety": {
            **artifact["safety"],
            "risk_level": "high",
            "requires_confirmation": True,
            "review_status": "blocked",
            "review_notes": ["blocked by image safety runbook"],
        },
    }


def assert_slice(fixture: dict[str, Any]) -> None:
    require(fixture.get("schema_version") == 1, "unexpected schema_version")
    require(fixture.get("kind") == "image_safety_runbook_evidence_v1", "unexpected fixture kind")
    slice_info = fixture.get("slice") or {}
    require(slice_info.get("id") == "image-safety-runbook-evidence-v1", "unexpected slice id")
    require(slice_info.get("track") == "Image Path", "unexpected track")
    require(
        slice_info.get("status") == "image_safety_runbook_evidence_defined",
        "image safety runbook status drifted",
    )
    missing = sorted(EXPECTED_FORBIDDEN_CLAIMS - set(slice_info.get("does_not_claim") or []))
    require(not missing, f"missing forbidden claims: {missing}")
    declared = set(fixture.get("depends_on") or [])
    missing_dependencies = sorted(EXPECTED_DEPENDENCIES - declared)
    require(not missing_dependencies, f"missing dependencies: {missing_dependencies}")


def assert_upstream_evidence() -> None:
    handshake = load_json(HANDSHAKE_FIXTURE_PATH)
    handshake_slice = handshake.get("slice") or {}
    require(handshake_slice.get("status") == "image_adapter_handshake_safety_gate_defined", "handshake status drifted")
    gate_ids = {str(row.get("gate_id") or "") for row in handshake.get("safety_gate_matrix") or []}
    for gate_id in (
        "allow_low_risk_structured_intent_review",
        "block_requires_confirmation",
        "block_high_risk_or_policy_unknown",
        "backend_unavailable",
        "artifact_metadata_only_return",
    ):
        require(gate_id in gate_ids, f"handshake missing gate {gate_id}")
    policy = handshake.get("artifact_return_policy") or {}
    require(policy.get("status") == "metadata_reference_only", "handshake artifact policy drifted")
    require(policy.get("artifact_uri_is_public_url") is False, "handshake must block public URL")

    artifact_return = load_json(ARTIFACT_RETURN_FIXTURE_PATH)
    return_slice = artifact_return.get("slice") or {}
    require(
        return_slice.get("status") == "image_artifact_return_runbook_evidence_defined",
        "artifact return status drifted",
    )
    return_failures = {str(row.get("failure_code") or "") for row in artifact_return.get("failure_taxonomy") or []}
    require("image_artifact_safety_blocked" in return_failures, "artifact return must keep safety blocked failure")
    visibility = artifact_return.get("artifact_visibility_policy") or {}
    require(visibility.get("public_url_visible_to_copilot_response") is False, "artifact return must block public URL")

    manifest = load_json(EVAL_MANIFEST_PATH)
    dimensions = set((manifest.get("scope") or {}).get("eval_dimensions") or [])
    require("safety_gate" in dimensions, "image eval manifest must keep safety gate dimension")
    execution = manifest.get("execution_policy") or {}
    for field in ("does_not_call_backend", "does_not_generate_images", "does_not_download_models"):
        require(execution.get(field) is True, f"eval manifest {field} must remain true")


def assert_existing_contract_safety_chain() -> None:
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
    require(value_at_path(intent, "safety.risk_level") == "low", "basic intent must remain low risk")
    require(value_at_path(backend_request, "safety.gate") == "approved_for_backend", "basic backend gate drifted")
    require(value_at_path(artifact, "safety.review_status") == "not_required", "basic artifact review drifted")

    confirmation_intent = with_confirmation_required(intent)
    confirmation_backend = build_blocked_backend_request(
        backend_request,
        requires_confirmation=True,
        risk_level="medium",
        review_note="blocked by image safety runbook because intent requires confirmation",
    )
    jsonschema.validate(confirmation_intent, intent_schema)
    jsonschema.validate(confirmation_backend, backend_schema)
    check_image_generation_intent(confirmation_intent)
    check_backend_request(confirmation_intent, confirmation_backend, expected_gate="blocked_requires_confirmation")

    high_risk_intent = build_high_risk_intent(intent)
    high_risk_backend = build_blocked_backend_request(
        backend_request,
        requires_confirmation=True,
        risk_level="high",
        review_note="blocked by image safety runbook because intent is high risk",
    )
    jsonschema.validate(high_risk_intent, intent_schema)
    jsonschema.validate(high_risk_backend, backend_schema)
    check_image_generation_intent(high_risk_intent)
    check_backend_request(high_risk_intent, high_risk_backend, expected_gate="blocked_requires_confirmation")

    blocked_artifact = build_blocked_artifact(artifact)
    jsonschema.validate(blocked_artifact, artifact_schema)
    require(blocked_artifact["status"] == "blocked", "blocked artifact status drifted")
    require(blocked_artifact["safety"]["review_status"] == "blocked", "blocked artifact review drifted")
    uri = str(value_at_path(blocked_artifact, "artifact.uri"))
    require(uri.startswith("artifact://"), "blocked artifact URI must remain artifact://")
    require(not uri.startswith(("http://", "https://")), "blocked artifact URI must not become public URL")


def assert_boundary(fixture: dict[str, Any]) -> None:
    boundary = fixture.get("safety_runbook_boundary") or {}
    require(
        boundary.get("status") == "image_safety_runbook_defined_no_runtime_policy_engine",
        "safety runbook boundary status drifted",
    )
    require(boundary.get("decision") == "safety_policy_runbook_only", "safety runbook decision drifted")
    require(boundary.get("source_intent_schema") == "contracts/image-generation-intent.schema.json", "intent schema ref")
    require(
        boundary.get("source_backend_request_schema") == "contracts/image-generation-backend-request.schema.json",
        "backend request schema ref",
    )
    require(boundary.get("source_artifact_schema") == "contracts/image-generation-artifact.schema.json", "artifact schema ref")
    for field in EXPECTED_BOUNDARY_FALSE_FIELDS:
        require(boundary.get(field) is False, f"{field} must remain false")


def assert_policy_input_fields(fixture: dict[str, Any]) -> None:
    policy = fixture.get("policy_input_fields") or {}
    require(policy.get("status") == "defined_not_executable", "policy input status drifted")
    require(policy.get("runtime_policy_engine_required_now") is False, "policy engine must not be required now")
    configured = {
        (str(row.get("source") or ""), str(row.get("path") or ""))
        for row in policy.get("required_fields") or []
        if isinstance(row, dict)
    }
    require(EXPECTED_POLICY_INPUT_FIELDS.issubset(configured), "policy input fields drifted")
    require(EXPECTED_FORBIDDEN_FIELDS.issubset(set(policy.get("forbidden_fields") or [])), "forbidden fields drifted")

    documents = {
        "intent": load_json(INTENT_FIXTURE_PATH),
        "backend_request": load_json(BACKEND_REQUEST_FIXTURE_PATH),
        "artifact": load_json(ARTIFACT_FIXTURE_PATH),
    }
    for source, dotted_path in EXPECTED_POLICY_INPUT_FIELDS:
        value_at_path(documents[source], dotted_path)


def assert_runbook_steps_and_decisions(fixture: dict[str, Any]) -> None:
    steps = fixture.get("safety_runbook_steps") or []
    ordered_ids = [str(step.get("step_id") or "") for step in sorted(steps, key=lambda row: int(row.get("order") or 0))]
    require(ordered_ids == EXPECTED_RUNBOOK_STEPS, "safety runbook step order drifted")
    for step in steps:
        require(step.get("runtime_implemented_now") is False, f"{step.get('step_id')} runtime must stay false")
    blocked_step = next(step for step in steps if step.get("step_id") == "defer_runtime_policy_engine_and_moderation_provider")
    require(blocked_step.get("allowed_now") is False, "runtime policy engine step must remain blocked")

    decisions = {
        str(row.get("case_id") or ""): row
        for row in fixture.get("safety_decision_matrix") or []
        if isinstance(row, dict)
    }
    require(set(decisions) == EXPECTED_DECISION_CASES, "safety decision matrix drifted")
    require(decisions["low_risk_no_confirmation"].get("expected_backend_gate") == "approved_for_backend", "low risk gate")
    require(
        decisions["low_risk_no_confirmation"].get("metadata_reference_allowed_by_runbook") is True,
        "low risk metadata reference runbook allowance drifted",
    )
    for case_id, row in decisions.items():
        require(row.get("backend_call_allowed_now") is False, f"{case_id} backend call must stay blocked")
        require(row.get("runtime_return_implemented_now") is False, f"{case_id} runtime return must stay false")
        if case_id != "low_risk_no_confirmation":
            require(
                row.get("metadata_reference_allowed_by_runbook") is False,
                f"{case_id} must not allow metadata reference",
            )


def assert_failures_and_artifact_review(fixture: dict[str, Any]) -> None:
    failures = {
        str(row.get("failure_code") or ""): row
        for row in fixture.get("failure_taxonomy") or []
        if isinstance(row, dict)
    }
    require(set(failures) == EXPECTED_FAILURE_CODES, "failure taxonomy drifted")
    for failure_code, row in failures.items():
        require(row.get("retry_or_fallback_allowed_now") is False, f"{failure_code} retry must stay blocked")
        require(row.get("artifact_reference_returned") is False, f"{failure_code} must not return artifact reference")

    review = fixture.get("artifact_safety_review_requirements") or {}
    require(EXPECTED_ARTIFACT_REVIEW_FIELDS.issubset(set(review.get("required_fields") or [])), "artifact review fields drifted")
    require(set(review.get("allowed_review_statuses") or []) == EXPECTED_REVIEW_STATUSES, "review statuses drifted")
    require(set(review.get("blocked_review_statuses") or []) == EXPECTED_BLOCKED_REVIEW_STATUSES, "blocked statuses drifted")
    blocked_when = set(review.get("artifact_reference_return_blocked_when") or [])
    for condition in ("review_status=pending_review", "review_status=blocked", "risk_level=high", "requires_confirmation=true"):
        require(condition in blocked_when, f"missing blocked artifact condition {condition}")

    artifact = load_json(ARTIFACT_FIXTURE_PATH)
    for dotted_path in EXPECTED_ARTIFACT_REVIEW_FIELDS:
        value_at_path(artifact, dotted_path)


def assert_visibility_execution_and_side_effects(fixture: dict[str, Any]) -> None:
    visibility = fixture.get("artifact_visibility_policy") or {}
    require(visibility.get("metadata_visible_to_safety_runbook") is True, "metadata visibility drifted")
    for field in (
        "pixel_payload_visible_to_safety_runbook",
        "provider_raw_dump_visible_to_safety_runbook",
        "external_moderation_dump_visible_to_safety_runbook",
        "public_url_visible_to_safety_runbook",
        "artifact_binary_download_allowed_now",
        "artifact_upload_allowed_now",
        "production_storage_allowed_now",
    ):
        require(visibility.get(field) is False, f"{field} must remain false")

    execution = fixture.get("execution_policy") or {}
    for field in EXPECTED_EXECUTION_TRUE_FIELDS:
        require(execution.get(field) is True, f"execution_policy.{field} must remain true")

    fallback = fixture.get("no_fake_fallback_policy") or {}
    for field in (
        "policy_unknown_promoted_to_success_allowed",
        "requires_confirmation_promoted_to_backend_allowed",
        "high_risk_promoted_to_backend_allowed",
        "pending_review_promoted_to_success_allowed",
        "blocked_artifact_promoted_to_reference_allowed",
        "backend_unavailable_promoted_to_retry_loop_allowed",
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
    previous_checker = "check-image-artifact-return-runbook-evidence-v1.py"
    current_checker = "check-image-safety-runbook-evidence-v1.py"
    require(current_checker in check_repo, "check-repo.py must run image safety runbook evidence")
    require(check_repo.index(previous_checker) < check_repo.index(current_checker), "image safety must run after artifact return")


def main() -> None:
    fixture = load_json(FIXTURE_PATH)
    assert_slice(fixture)
    assert_upstream_evidence()
    assert_existing_contract_safety_chain()
    assert_boundary(fixture)
    assert_policy_input_fields(fixture)
    assert_runbook_steps_and_decisions(fixture)
    assert_failures_and_artifact_review(fixture)
    assert_visibility_execution_and_side_effects(fixture)
    assert_forbidden_artifacts_and_sources(fixture)
    assert_references_and_check_repo(fixture)
    print("image safety runbook evidence v1 checks passed.")


if __name__ == "__main__":
    main()
