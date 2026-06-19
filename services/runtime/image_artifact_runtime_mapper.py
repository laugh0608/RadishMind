from __future__ import annotations

import re
from collections.abc import Mapping
from dataclasses import dataclass
from typing import Any


ARTIFACT_URI_SCHEME = "artifact://"
ALLOWED_MIME_TYPES = {"image/png", "image/jpeg", "image/webp"}
ALLOWED_ARTIFACT_STATUSES = {"generated", "blocked", "failed"}
SUCCESS_REVIEW_STATUSES = {"not_required", "reviewed_pass"}
HASH_RE = re.compile(r"^[a-f0-9]{64}$")

FAILURE_INVALID_METADATA = "image_artifact_invalid_metadata"
FAILURE_HASH_MISMATCH = "image_artifact_hash_mismatch"
FAILURE_MIME_MISMATCH = "image_artifact_mime_mismatch"
FAILURE_DIMENSION_MISMATCH = "image_artifact_dimension_mismatch"
FAILURE_PUBLIC_URL_CLAIM = "image_artifact_public_url_claim"
FAILURE_SIGNED_URL_POLICY_MISSING = "image_artifact_signed_url_policy_missing"
FAILURE_BINARY_PAYLOAD_REJECTED = "image_artifact_binary_payload_rejected"
FAILURE_PROVIDER_RAW_DUMP_REJECTED = "image_artifact_provider_raw_dump_rejected"
FAILURE_ARTIFACT_STORE_MISSING = "image_artifact_store_missing"
FAILURE_ARTIFACT_STORE_UNAVAILABLE = "image_artifact_store_unavailable"
FAILURE_ARTIFACT_BINARY_READER_MISSING = "image_artifact_binary_reader_missing"
FAILURE_ARTIFACT_BINARY_READ_FORBIDDEN = "image_artifact_binary_read_forbidden"
FAILURE_SAFETY_REVIEW_NOT_PASSED = "image_artifact_safety_review_not_passed"
FAILURE_PROVENANCE_MISSING = "image_artifact_provenance_missing"
FAILURE_SAFETY_PENDING_REVIEW = "image_artifact_safety_pending_review"
FAILURE_SAFETY_BLOCKED = "image_artifact_safety_blocked"

ZERO_SIDE_EFFECT_COUNTERS = {
    "backend_call_count": 0,
    "image_generation_count": 0,
    "model_download_count": 0,
    "artifact_upload_count": 0,
    "artifact_binary_read_count": 0,
    "artifact_store_lookup_count": 0,
    "runtime_mapping_execution_count": 0,
    "production_storage_write_count": 0,
    "public_url_resolution_count": 0,
    "executor_call_count": 0,
    "confirmation_call_count": 0,
    "business_writeback_count": 0,
    "replay_call_count": 0,
}

FORBIDDEN_KEY_FAILURES = (
    ({"provider_raw_response", "provider_raw_dump", "raw_provider_response"}, FAILURE_PROVIDER_RAW_DUMP_REJECTED),
    ({"pixel_payload", "base64_image", "binary_payload", "image_bytes"}, FAILURE_BINARY_PAYLOAD_REJECTED),
    ({"signed_public_url", "signed_url"}, FAILURE_SIGNED_URL_POLICY_MISSING),
    ({"public_url"}, FAILURE_PUBLIC_URL_CLAIM),
)


@dataclass(frozen=True)
class ImageArtifactMappingResult:
    ok: bool
    citation: dict[str, Any] | None = None
    metadata_reference: dict[str, Any] | None = None
    failure_code: str | None = None
    failure_message: str = ""


def runtime_mapper_side_effect_counters() -> dict[str, int]:
    return dict(ZERO_SIDE_EFFECT_COUNTERS)


def map_image_artifact_to_response_reference(
    artifact_document: Mapping[str, Any],
    *,
    expected_sha256: str | None = None,
    expected_mime_type: str | None = None,
    expected_width: int | None = None,
    expected_height: int | None = None,
    artifact_store_state: str = "not_required",
    artifact_binary_reader_state: str = "not_required",
) -> ImageArtifactMappingResult:
    if not isinstance(artifact_document, Mapping):
        return fail(FAILURE_INVALID_METADATA, "artifact document must be a mapping")

    forbidden_failure = find_forbidden_metadata_failure(artifact_document)
    if forbidden_failure:
        return fail(forbidden_failure, "artifact metadata contains forbidden payload or URL claim")

    if artifact_store_state == "missing":
        return fail(FAILURE_ARTIFACT_STORE_MISSING, "artifact store is required by caller context but missing")
    if artifact_store_state == "unavailable":
        return fail(FAILURE_ARTIFACT_STORE_UNAVAILABLE, "artifact store is required by caller context but unavailable")
    if artifact_binary_reader_state == "missing":
        return fail(FAILURE_ARTIFACT_BINARY_READER_MISSING, "artifact binary reader is required by caller context but missing")
    if artifact_binary_reader_state == "forbidden":
        return fail(FAILURE_ARTIFACT_BINARY_READ_FORBIDDEN, "artifact binary reader use is forbidden by policy")

    if artifact_document.get("schema_version") != 1:
        return fail(FAILURE_INVALID_METADATA, "artifact schema_version must be 1")
    if artifact_document.get("kind") != "image_generation_artifact":
        return fail(FAILURE_INVALID_METADATA, "artifact kind must be image_generation_artifact")

    status = normalized_string(artifact_document.get("status"))
    if status not in ALLOWED_ARTIFACT_STATUSES:
        return fail(FAILURE_INVALID_METADATA, "artifact status is invalid")
    if status == "blocked":
        return fail(FAILURE_SAFETY_BLOCKED, "blocked artifact must not enter a success response")
    if status == "failed":
        return fail(FAILURE_INVALID_METADATA, "failed artifact must not enter a success response")

    artifact = mapping_value(artifact_document, "artifact")
    generation = mapping_value(artifact_document, "generation")
    safety = mapping_value(artifact_document, "safety")
    provenance = mapping_value(artifact_document, "provenance")
    if artifact is None or generation is None or safety is None or provenance is None:
        return fail(FAILURE_INVALID_METADATA, "artifact metadata sections are incomplete")

    artifact_id = normalized_string(artifact_document.get("artifact_id"))
    created_at = normalized_string(artifact_document.get("created_at"))
    if not artifact_id or not created_at:
        return fail(FAILURE_INVALID_METADATA, "artifact_id and created_at are required")

    artifact_validation = validate_artifact_metadata(
        artifact,
        expected_sha256=expected_sha256,
        expected_mime_type=expected_mime_type,
        expected_width=expected_width,
        expected_height=expected_height,
    )
    if artifact_validation is not None:
        return artifact_validation

    generation_validation = validate_generation_metadata(generation)
    if generation_validation is not None:
        return generation_validation

    safety_validation = validate_safety_metadata(safety)
    if safety_validation is not None:
        return safety_validation

    provenance_validation = validate_provenance_metadata(provenance)
    if provenance_validation is not None:
        return provenance_validation

    artifact_uri = normalized_string(artifact["uri"])
    title = normalized_string(artifact["title"])
    citation = {
        "id": artifact_id,
        "kind": "artifact",
        "label": title,
        "locator": artifact_uri,
        "source_uri": artifact_uri,
    }
    metadata_reference = {
        "artifact_id": artifact_id,
        "uri": artifact_uri,
        "sha256": normalized_string(artifact["sha256"]),
        "mime_type": normalized_string(artifact["mime_type"]),
        "dimensions": {
            "width": int(artifact["width"]),
            "height": int(artifact["height"]),
        },
        "format": normalized_string(artifact["format"]),
        "title": title,
        "purpose": normalized_string(artifact["purpose"]),
        "backend_id": normalized_string(generation["backend_id"]),
        "model": normalized_string(generation["model"]),
        "seed": int(generation["seed"]),
        "safety": {
            "risk_level": normalized_string(safety["risk_level"]),
            "requires_confirmation": bool(safety["requires_confirmation"]),
            "review_status": normalized_string(safety["review_status"]),
        },
        "provenance": {
            "source_request_id": normalized_string(provenance["source_request_id"]),
            "trace_ids": [normalized_string(item) for item in provenance["trace_ids"]],
        },
        "created_at": created_at,
    }
    return ImageArtifactMappingResult(ok=True, citation=citation, metadata_reference=metadata_reference)


def validate_artifact_metadata(
    artifact: Mapping[str, Any],
    *,
    expected_sha256: str | None,
    expected_mime_type: str | None,
    expected_width: int | None,
    expected_height: int | None,
) -> ImageArtifactMappingResult | None:
    for key in ("uri", "mime_type", "width", "height", "format", "sha256", "title", "purpose"):
        if key not in artifact:
            return fail(FAILURE_INVALID_METADATA, f"artifact.{key} is required")

    artifact_uri = normalized_string(artifact.get("uri"))
    if not artifact_uri:
        return fail(FAILURE_INVALID_METADATA, "artifact.uri is required")
    if artifact_uri.startswith(("http://", "https://")):
        return fail(FAILURE_PUBLIC_URL_CLAIM, "artifact.uri must not be a public URL")
    if not artifact_uri.startswith(ARTIFACT_URI_SCHEME):
        return fail(FAILURE_INVALID_METADATA, "artifact.uri must use artifact://")

    sha256 = normalized_string(artifact.get("sha256"))
    if not HASH_RE.match(sha256):
        return fail(FAILURE_INVALID_METADATA, "artifact.sha256 must be lowercase sha256 hex")
    if expected_sha256 is not None and normalized_string(expected_sha256) != sha256:
        return fail(FAILURE_HASH_MISMATCH, "artifact.sha256 does not match expected digest")

    mime_type = normalized_string(artifact.get("mime_type"))
    if mime_type not in ALLOWED_MIME_TYPES:
        return fail(FAILURE_INVALID_METADATA, "artifact.mime_type is unsupported")
    if expected_mime_type is not None and normalized_string(expected_mime_type) != mime_type:
        return fail(FAILURE_MIME_MISMATCH, "artifact.mime_type does not match expected MIME type")

    width = positive_integer(artifact.get("width"))
    height = positive_integer(artifact.get("height"))
    if width is None or height is None:
        return fail(FAILURE_INVALID_METADATA, "artifact dimensions must be positive integers")
    if expected_width is not None and int(expected_width) != width:
        return fail(FAILURE_DIMENSION_MISMATCH, "artifact.width does not match expected dimension")
    if expected_height is not None and int(expected_height) != height:
        return fail(FAILURE_DIMENSION_MISMATCH, "artifact.height does not match expected dimension")

    if not normalized_string(artifact.get("title")) or not normalized_string(artifact.get("purpose")):
        return fail(FAILURE_INVALID_METADATA, "artifact title and purpose are required")
    return None


def validate_generation_metadata(generation: Mapping[str, Any]) -> ImageArtifactMappingResult | None:
    for key in ("backend_id", "model", "seed"):
        if key not in generation:
            return fail(FAILURE_INVALID_METADATA, f"generation.{key} is required")
    if not normalized_string(generation.get("backend_id")) or not normalized_string(generation.get("model")):
        return fail(FAILURE_INVALID_METADATA, "generation backend_id and model are required")
    if not isinstance(generation.get("seed"), int) or generation["seed"] < 0:
        return fail(FAILURE_INVALID_METADATA, "generation.seed must be a non-negative integer")
    return None


def validate_safety_metadata(safety: Mapping[str, Any]) -> ImageArtifactMappingResult | None:
    for key in ("risk_level", "requires_confirmation", "review_status"):
        if key not in safety:
            return fail(FAILURE_INVALID_METADATA, f"safety.{key} is required")

    review_status = normalized_string(safety.get("review_status"))
    requires_confirmation = safety.get("requires_confirmation")
    if requires_confirmation is not True and requires_confirmation is not False:
        return fail(FAILURE_INVALID_METADATA, "safety.requires_confirmation must be boolean")
    if review_status == "pending_review":
        return fail(FAILURE_SAFETY_PENDING_REVIEW, "pending review artifact must not enter a success response")
    if review_status == "blocked":
        return fail(FAILURE_SAFETY_REVIEW_NOT_PASSED, "blocked review artifact must not enter a success response")
    if review_status not in SUCCESS_REVIEW_STATUSES:
        return fail(FAILURE_INVALID_METADATA, "safety.review_status is invalid")
    if requires_confirmation is True and review_status != "reviewed_pass":
        return fail(FAILURE_SAFETY_REVIEW_NOT_PASSED, "confirmation-required artifact needs reviewed_pass")
    if normalized_string(safety.get("risk_level")) not in {"low", "medium", "high"}:
        return fail(FAILURE_INVALID_METADATA, "safety.risk_level is invalid")
    return None


def validate_provenance_metadata(provenance: Mapping[str, Any]) -> ImageArtifactMappingResult | None:
    for key in ("source_request_id", "trace_ids", "backend_request_id", "intent_id"):
        if key not in provenance:
            return fail(FAILURE_PROVENANCE_MISSING, f"provenance.{key} is required")
    for key in ("source_request_id", "backend_request_id", "intent_id"):
        if not normalized_string(provenance.get(key)):
            return fail(FAILURE_PROVENANCE_MISSING, f"provenance.{key} must be non-empty")

    trace_ids = provenance.get("trace_ids")
    if not isinstance(trace_ids, list) or not trace_ids:
        return fail(FAILURE_PROVENANCE_MISSING, "provenance.trace_ids must be a non-empty list")
    if any(not normalized_string(item) for item in trace_ids):
        return fail(FAILURE_PROVENANCE_MISSING, "provenance.trace_ids must contain non-empty strings")
    return None


def mapping_value(document: Mapping[str, Any], key: str) -> Mapping[str, Any] | None:
    value = document.get(key)
    if isinstance(value, Mapping):
        return value
    return None


def normalized_string(value: Any) -> str:
    if isinstance(value, str):
        return " ".join(value.split()).strip()
    return ""


def positive_integer(value: Any) -> int | None:
    if isinstance(value, bool):
        return None
    if isinstance(value, int) and value > 0:
        return value
    return None


def find_forbidden_metadata_failure(value: Any) -> str | None:
    if isinstance(value, Mapping):
        for keys, failure_code in FORBIDDEN_KEY_FAILURES:
            if any(key in value for key in keys):
                return failure_code
        for item in value.values():
            failure_code = find_forbidden_metadata_failure(item)
            if failure_code:
                return failure_code
    if isinstance(value, list):
        for item in value:
            failure_code = find_forbidden_metadata_failure(item)
            if failure_code:
                return failure_code
    return None


def fail(failure_code: str, message: str) -> ImageArtifactMappingResult:
    return ImageArtifactMappingResult(ok=False, failure_code=failure_code, failure_message=message)
