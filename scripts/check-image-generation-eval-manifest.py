#!/usr/bin/env python3
from __future__ import annotations

import argparse
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


DEFAULT_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/image-generation-eval-manifest-v0.json"
INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"
REQUIRED_EVAL_DIMENSIONS = {
    "structured_intent",
    "backend_request_mapping",
    "artifact_metadata",
    "safety_gate",
    "provenance",
}
DISALLOWED_COMMITTED_ARTIFACTS = {
    "image_pixels",
    "provider_raw_dump",
    "model_weights",
    "checkpoint",
    "large_jsonl",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", default=str(DEFAULT_MANIFEST_PATH.relative_to(REPO_ROOT)))
    return parser.parse_args()


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_existing_fixture(path_value: str | None, *, field_name: str) -> Path | None:
    if path_value is None:
        return None
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a clean non-empty path")
    path = REPO_ROOT / path_value
    require(path.is_file(), f"{field_name} points to a missing file: {path_value}")
    return path


def load_schemas() -> tuple[dict[str, Any], dict[str, Any], dict[str, Any]]:
    intent_schema = load_json_document(INTENT_SCHEMA_PATH)
    backend_schema = load_json_document(BACKEND_REQUEST_SCHEMA_PATH)
    artifact_schema = load_json_document(ARTIFACT_SCHEMA_PATH)
    for schema in (intent_schema, backend_schema, artifact_schema):
        jsonschema.Draft202012Validator.check_schema(schema)
    return intent_schema, backend_schema, artifact_schema


def check_manifest_policy(manifest: dict[str, Any]) -> None:
    require(manifest.get("schema_version") == 1, "image generation eval manifest schema_version must be 1")
    require(manifest.get("kind") == "image_generation_eval_manifest", "image generation eval manifest kind mismatch")
    require(manifest.get("status") == "draft", "image generation eval manifest must remain draft")

    scope = manifest.get("scope")
    require(isinstance(scope, dict), "manifest scope must be an object")
    dimensions = set(scope.get("eval_dimensions") or [])
    require(dimensions == REQUIRED_EVAL_DIMENSIONS, "image generation eval dimensions drifted")
    excluded_dimensions = set(scope.get("excluded_dimensions") or [])
    for excluded in ("image_pixel_quality", "real_backend_latency", "provider_specific_rendering"):
        require(excluded in excluded_dimensions, f"excluded_dimensions must include {excluded}")

    execution_policy = manifest.get("execution_policy")
    require(isinstance(execution_policy, dict), "execution_policy must be an object")
    for flag in ("does_not_call_backend", "does_not_generate_images", "does_not_download_models", "does_not_start_training"):
        require(execution_policy.get(flag) is True, f"execution_policy.{flag} must be true")

    artifact_policy = manifest.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "artifact_policy must be an object")
    disallowed = set(artifact_policy.get("committed_disallowed") or [])
    require(DISALLOWED_COMMITTED_ARTIFACTS <= disallowed, "artifact_policy committed_disallowed is incomplete")


def build_blocked_backend_request(backend_request: dict[str, Any]) -> dict[str, Any]:
    return {
        **backend_request,
        "safety": {
            **backend_request["safety"],
            "gate": "blocked_requires_confirmation",
            "requires_confirmation": True,
            "risk_level": "medium",
            "review_notes": ["blocked by image generation eval manifest because intent requires confirmation"],
        },
    }


def check_case(
    case: dict[str, Any],
    *,
    intent_schema: dict[str, Any],
    backend_schema: dict[str, Any],
    artifact_schema: dict[str, Any],
) -> dict[str, int]:
    case_id = str(case.get("case_id") or "")
    require(case_id, "image generation eval case is missing case_id")
    intent_path = require_existing_fixture(str(case.get("intent_fixture") or ""), field_name=f"{case_id}.intent_fixture")
    backend_path = require_existing_fixture(str(case.get("backend_request_fixture") or ""), field_name=f"{case_id}.backend_request_fixture")
    artifact_path = require_existing_fixture(case.get("artifact_fixture"), field_name=f"{case_id}.artifact_fixture")

    intent = load_json_document(intent_path)
    backend_request = load_json_document(backend_path)
    if bool(case.get("simulate_confirmation_required")):
        intent = with_confirmation_required(intent)
        backend_request = build_blocked_backend_request(backend_request)

    jsonschema.validate(intent, intent_schema)
    jsonschema.validate(backend_request, backend_schema)
    check_image_generation_intent(intent)
    expected_gate = str(case.get("expected_safety_gate") or "")
    check_backend_request(intent, backend_request, expected_gate=expected_gate)

    expected_backend_submittable = bool(case.get("expected_backend_submittable"))
    require(
        (expected_gate == "approved_for_backend") == expected_backend_submittable,
        f"{case_id} backend submittable expectation does not match safety gate",
    )

    artifact_case_count = 0
    if artifact_path is not None:
        artifact = load_json_document(artifact_path)
        jsonschema.validate(artifact, artifact_schema)
        check_artifact(intent, backend_request, artifact)
        expected_status = case.get("expected_artifact_status")
        require(artifact.get("status") == expected_status, f"{case_id} artifact status mismatch")
        artifact_case_count = 1
    else:
        require(case.get("expected_artifact_status") is None, f"{case_id} must not expect artifact status without fixture")

    coverage = set(case.get("coverage") or [])
    require(coverage <= REQUIRED_EVAL_DIMENSIONS, f"{case_id} has unknown coverage dimensions")
    require(coverage, f"{case_id} must include coverage dimensions")

    return {
        "approved_for_backend_count": 1 if expected_gate == "approved_for_backend" else 0,
        "blocked_requires_confirmation_count": 1 if expected_gate == "blocked_requires_confirmation" else 0,
        "artifact_metadata_case_count": artifact_case_count,
    }


def main() -> int:
    args = parse_args()
    manifest_path = REPO_ROOT / args.manifest
    manifest = load_json_document(manifest_path)
    require(isinstance(manifest, dict), "image generation eval manifest must be an object")
    check_manifest_policy(manifest)

    intent_schema, backend_schema, artifact_schema = load_schemas()
    cases = manifest.get("cases")
    require(isinstance(cases, list) and cases, "image generation eval manifest must include cases")

    summary = {
        "case_count": len(cases),
        "approved_for_backend_count": 0,
        "blocked_requires_confirmation_count": 0,
        "artifact_metadata_case_count": 0,
        **manifest["execution_policy"],
    }
    seen_case_ids: set[str] = set()
    for case in cases:
        require(isinstance(case, dict), "image generation eval case must be an object")
        case_id = str(case.get("case_id") or "")
        require(case_id not in seen_case_ids, f"duplicate image generation eval case_id: {case_id}")
        seen_case_ids.add(case_id)
        case_summary = check_case(
            case,
            intent_schema=intent_schema,
            backend_schema=backend_schema,
            artifact_schema=artifact_schema,
        )
        for key, value in case_summary.items():
            summary[key] += value

    require(summary == manifest.get("expected_summary"), "image generation eval manifest summary mismatch")
    print("image generation eval manifest smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
