from __future__ import annotations

import copy
import json
import math
import re
from collections.abc import Mapping
from dataclasses import dataclass
from functools import lru_cache
from pathlib import Path
from typing import Any, Protocol

import jsonschema

from .image_artifact_runtime_mapper import (
    FAILURE_HASH_MISMATCH,
    ImageArtifactMappingResult,
    map_image_artifact_to_response_reference,
)


REPO_ROOT = Path(__file__).resolve().parents[2]
INTENT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-intent.schema.json"
BACKEND_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-backend-request.schema.json"
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"

FAILURE_INTENT_INVALID = "image_intent_invalid"
FAILURE_INTENT_BUDGET_EXCEEDED = "image_intent_budget_exceeded"
FAILURE_INTENT_SENSITIVE_MATERIAL = "image_intent_sensitive_material_rejected"
FAILURE_INTENT_REQUIRES_CONFIRMATION = "image_intent_requires_confirmation"
FAILURE_INTENT_HIGH_RISK = "image_intent_high_risk"
FAILURE_BACKEND_PROFILE_MISSING = "image_backend_profile_missing"
FAILURE_BACKEND_PROFILE_MISMATCH = "image_backend_profile_mismatch"
FAILURE_BACKEND_SAFETY_BLOCKED = "image_backend_safety_gate_blocked"
FAILURE_BACKEND_TIMEOUT = "image_backend_timeout"
FAILURE_BACKEND_UNAVAILABLE = "image_backend_unavailable"
FAILURE_BACKEND_INVALID_ARTIFACT = "image_backend_invalid_artifact_metadata"
FAILURE_BACKEND_ARTIFACT_LINEAGE = "image_backend_artifact_lineage_mismatch"
FAILURE_BACKEND_ARTIFACT_HASH = "image_backend_artifact_hash_mismatch"
FAILURE_BACKEND_RESPONSE_UNTRUSTED = "image_backend_response_untrusted"

MAX_PROMPT_POSITIVE_BYTES = 12_000
MAX_PROMPT_NEGATIVE_BYTES = 8_000
MAX_PROMPT_TOTAL_BYTES = 16_000
MAX_DIMENSION = 2_048
MAX_PIXELS = 4_194_304
MAX_STEPS = 100
MAX_GUIDANCE_SCALE = 30.0
MAX_REFERENCE_ARTIFACTS = 4
MAX_CONSTRAINT_ITEMS = 16
MAX_REVIEW_NOTES = 8
MAX_TRACE_IDS = 16
MAX_IDENTIFIER_BYTES = 128
MAX_LIST_ITEM_BYTES = 512
MAX_ARTIFACT_TITLE_BYTES = 160
MAX_TIMEOUT_SECONDS = 300.0

IDENTIFIER_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$")
PROFILE_VALUE_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._:/-]{0,127}$")
LOCALE_RE = re.compile(r"^[a-z]{2,3}(?:-[A-Z]{2})?$")
SHA256_RE = re.compile(r"^[a-f0-9]{64}$")
SENSITIVE_PATTERNS = (
    re.compile(r"https?://", re.IGNORECASE),
    re.compile(r"\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis)://", re.IGNORECASE),
    re.compile(
        r"\b(?:authorization|headers?|api[_-]?key|credentials?|token|access[_-]?token|refresh[_-]?token|cookie|password|secret|endpoint|dsn|model[_-]?dir|provider[_-]?(?:config|runtime))\s*[:=]",
        re.IGNORECASE,
    ),
    re.compile(r"\bbearer\s+[A-Za-z0-9._~+/=-]{8,}", re.IGNORECASE),
    re.compile(r"\bsk-[A-Za-z0-9_-]{12,}", re.IGNORECASE),
    re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"),
)
MIME_BY_FORMAT = {
    "png": "image/png",
    "jpg": "image/jpeg",
    "webp": "image/webp",
}
ZERO_SIDE_EFFECTS = {
    "artifact_store_lookup_count": 0,
    "artifact_binary_read_count": 0,
    "artifact_upload_count": 0,
    "public_url_resolution_count": 0,
    "production_storage_write_count": 0,
    "retry_count": 0,
    "fallback_count": 0,
    "executor_call_count": 0,
    "confirmation_call_count": 0,
    "business_writeback_count": 0,
    "replay_call_count": 0,
}


@dataclass(frozen=True)
class ImageBackendProfile:
    backend_id: str
    model: str
    adapter_profile: str
    enabled: bool = True


@dataclass(frozen=True)
class ImageBackendInvocationResult:
    artifact_document: Mapping[str, Any]
    observed_sha256: str
    observed_mime_type: str
    observed_width: int
    observed_height: int


class ImageBackendClient(Protocol):
    def invoke(
        self,
        request_document: Mapping[str, Any],
        timeout_seconds: float,
    ) -> ImageBackendInvocationResult: ...


@dataclass(frozen=True)
class ImageGenerationAdapterResult:
    ok: bool
    backend_request: dict[str, Any] | None = None
    artifact_document: dict[str, Any] | None = None
    citation: dict[str, Any] | None = None
    metadata_reference: dict[str, Any] | None = None
    failure_code: str | None = None
    failure_message: str = ""
    backend_call_count: int = 0
    image_generation_count: int = 0


def adapter_side_effect_counters(result: ImageGenerationAdapterResult) -> dict[str, int]:
    return {
        "backend_call_count": result.backend_call_count,
        "image_generation_count": result.image_generation_count,
        **ZERO_SIDE_EFFECTS,
    }


def invoke_image_generation(
    intent_document: Mapping[str, Any],
    *,
    profile: ImageBackendProfile | None,
    client: ImageBackendClient,
    backend_request_id: str,
    timeout_seconds: float,
) -> ImageGenerationAdapterResult:
    intent = copy_mapping(intent_document)
    if intent is None:
        return fail(FAILURE_INTENT_INVALID, "image intent must be an object")

    schema_failure = validate_schema(intent, INTENT_SCHEMA_PATH)
    if schema_failure:
        return fail(FAILURE_INTENT_INVALID, schema_failure)
    if not contains_only_valid_utf8(intent):
        return fail(FAILURE_INTENT_INVALID, "image intent must contain valid UTF-8 text")

    failure = validate_intent_budget(intent, backend_request_id, timeout_seconds)
    if failure:
        return failure
    if contains_sensitive_material(intent) or (profile is not None and contains_sensitive_material(profile_document(profile))):
        return fail(FAILURE_INTENT_SENSITIVE_MATERIAL, "image intent or profile contains sensitive material")

    safety = mapping_at(intent, "safety")
    if safety["requires_confirmation"] is True:
        return fail(FAILURE_INTENT_REQUIRES_CONFIRMATION, "image intent requires human confirmation before backend submission")
    if safety["risk_level"] == "high":
        return fail(FAILURE_INTENT_HIGH_RISK, "high-risk image intent is not eligible for backend submission")
    if safety["risk_level"] != "low":
        return fail(FAILURE_BACKEND_SAFETY_BLOCKED, "image intent did not pass the low-risk backend safety gate")

    if profile is None or not profile.enabled or not profile_is_valid(profile):
        return fail(FAILURE_BACKEND_PROFILE_MISSING, "an enabled image backend profile is required")
    preferred_backend = mapping_at(intent, "backend")["preferred"]
    if preferred_backend != profile.backend_id:
        return fail(FAILURE_BACKEND_PROFILE_MISMATCH, "image backend profile does not match intent preference")

    backend_request = compile_image_backend_request(intent, profile=profile, backend_request_id=backend_request_id)
    schema_failure = validate_schema(backend_request, BACKEND_REQUEST_SCHEMA_PATH)
    if schema_failure:
        return fail(FAILURE_INTENT_INVALID, "compiled backend request did not satisfy the canonical contract")

    try:
        backend_result = client.invoke(copy.deepcopy(backend_request), float(timeout_seconds))
    except TimeoutError:
        return fail_after_call(
            FAILURE_BACKEND_TIMEOUT,
            "image backend request timed out",
        )
    except ConnectionError:
        return fail_after_call(
            FAILURE_BACKEND_UNAVAILABLE,
            "image backend is unavailable",
        )
    except Exception:
        return fail_after_call(
            FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            "image backend invocation failed without a trusted result",
        )

    return validate_backend_result(
        intent,
        profile,
        backend_request,
        backend_result,
    )


def compile_image_backend_request(
    intent_document: Mapping[str, Any],
    *,
    profile: ImageBackendProfile,
    backend_request_id: str,
) -> dict[str, Any]:
    intent = copy.deepcopy(dict(intent_document))
    backend = mapping_at(intent, "backend")
    prompt = mapping_at(intent, "prompt")
    output = mapping_at(intent, "output")
    style = mapping_at(intent, "style")
    constraints = mapping_at(intent, "constraints")
    safety = mapping_at(intent, "safety")
    artifact_metadata = mapping_at(intent, "artifact_metadata")
    trace_ids = canonical_strings(
        [
            intent["source_request_id"],
            intent["intent_id"],
            backend_request_id,
            *artifact_metadata["trace_ids"],
        ]
    )
    return {
        "schema_version": 1,
        "kind": "image_generation_backend_request",
        "request_id": backend_request_id,
        "intent_id": intent["intent_id"],
        "backend": {
            "id": profile.backend_id,
            "model": profile.model,
            "adapter_profile": profile.adapter_profile,
        },
        "prompt": {
            "positive": prompt["positive"],
            "negative": prompt["negative"],
            "locale": prompt["locale"],
            "transformed_from_intent": False,
        },
        "output": copy.deepcopy(output),
        "parameters": {
            "seed": backend["seed"],
            "steps": backend["steps"],
            "guidance_scale": backend["guidance_scale"],
        },
        "inputs": {
            "reference_artifact_ids": copy.deepcopy(style["reference_artifact_ids"]),
            "edit_artifact_id": constraints["edit_artifact_id"],
            "mask_artifact_id": constraints["mask_artifact_id"],
        },
        "constraints": {
            "must_include": copy.deepcopy(constraints["must_include"]),
            "must_avoid": copy.deepcopy(constraints["must_avoid"]),
            "style_preset": style["preset"],
        },
        "safety": {
            "gate": "approved_for_backend",
            "requires_confirmation": False,
            "risk_level": "low",
            "review_notes": copy.deepcopy(safety["review_notes"]),
        },
        "trace": {
            "source_request_id": intent["source_request_id"],
            "trace_ids": trace_ids,
        },
    }


def validate_backend_result(
    intent: dict[str, Any],
    profile: ImageBackendProfile,
    backend_request: dict[str, Any],
    backend_result: Any,
) -> ImageGenerationAdapterResult:
    if not isinstance(backend_result, ImageBackendInvocationResult):
        return fail_after_call(
            FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            "image backend returned an unsupported result envelope",
        )

    artifact = copy_mapping(backend_result.artifact_document)
    if artifact is None:
        return fail_after_call(
            FAILURE_BACKEND_INVALID_ARTIFACT,
            "image backend artifact metadata must be an object",
        )
    schema_failure = validate_schema(artifact, ARTIFACT_SCHEMA_PATH)
    if schema_failure:
        return fail_after_call(
            FAILURE_BACKEND_INVALID_ARTIFACT,
            schema_failure,
        )
    if not contains_only_valid_utf8(artifact):
        return fail_after_call(
            FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            "image backend artifact metadata must contain valid UTF-8 text",
        )
    if contains_sensitive_material(artifact):
        return fail_after_call(
            FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            "image backend artifact metadata contains sensitive material",
        )

    lineage_failure = validate_artifact_lineage(intent, profile, backend_request, artifact)
    if lineage_failure:
        return fail_after_call(
            FAILURE_BACKEND_ARTIFACT_LINEAGE,
            lineage_failure,
        )
    if not observed_result_is_valid(backend_result):
        return fail_after_call(
            FAILURE_BACKEND_RESPONSE_UNTRUSTED,
            "image backend transport observation is invalid",
        )

    mapping = map_image_artifact_to_response_reference(
        artifact,
        expected_sha256=backend_result.observed_sha256,
        expected_mime_type=backend_result.observed_mime_type,
        expected_width=backend_result.observed_width,
        expected_height=backend_result.observed_height,
    )
    if not mapping.ok:
        failure_code = FAILURE_BACKEND_ARTIFACT_HASH if mapping.failure_code == FAILURE_HASH_MISMATCH else FAILURE_BACKEND_RESPONSE_UNTRUSTED
        return fail_after_call(
            failure_code,
            "image backend artifact metadata does not match the trusted transport observation",
        )

    return success(backend_request, artifact, mapping)


def validate_intent_budget(
    intent: dict[str, Any],
    backend_request_id: str,
    timeout_seconds: float,
) -> ImageGenerationAdapterResult | None:
    for identifier in (intent.get("intent_id"), intent.get("source_request_id"), backend_request_id):
        if not valid_identifier(identifier):
            return fail(FAILURE_INTENT_INVALID, "image intent and backend request identifiers must be canonical")
    if not isinstance(timeout_seconds, (int, float)) or isinstance(timeout_seconds, bool):
        return fail(FAILURE_INTENT_INVALID, "image backend timeout must be numeric")
    if not math.isfinite(float(timeout_seconds)) or timeout_seconds <= 0 or timeout_seconds > MAX_TIMEOUT_SECONDS:
        return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image backend timeout exceeds the development budget")

    prompt = mapping_at(intent, "prompt")
    if not LOCALE_RE.fullmatch(prompt["locale"]):
        return fail(FAILURE_INTENT_INVALID, "image prompt locale must use canonical language or language-region form")
    positive_bytes = utf8_size(prompt["positive"])
    negative_bytes = utf8_size(prompt["negative"])
    if (
        positive_bytes > MAX_PROMPT_POSITIVE_BYTES
        or negative_bytes > MAX_PROMPT_NEGATIVE_BYTES
        or positive_bytes + negative_bytes > MAX_PROMPT_TOTAL_BYTES
    ):
        return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image prompt exceeds the UTF-8 byte budget")

    output = mapping_at(intent, "output")
    width = output["width"]
    height = output["height"]
    if width > MAX_DIMENSION or height > MAX_DIMENSION or width * height > MAX_PIXELS or output["count"] != 1:
        return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image output exceeds the dimension, pixel, or count budget")

    backend = mapping_at(intent, "backend")
    guidance_scale = float(backend["guidance_scale"])
    if backend["steps"] > MAX_STEPS or not math.isfinite(guidance_scale) or guidance_scale > MAX_GUIDANCE_SCALE:
        return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image generation parameters exceed the development budget")

    style = mapping_at(intent, "style")
    constraints = mapping_at(intent, "constraints")
    safety = mapping_at(intent, "safety")
    artifact_metadata = mapping_at(intent, "artifact_metadata")
    artifact_ids = [
        *style["reference_artifact_ids"],
        *[value for value in (constraints["edit_artifact_id"], constraints["mask_artifact_id"]) if value is not None],
        *artifact_metadata["trace_ids"],
    ]
    if any(not valid_identifier(value) for value in artifact_ids):
        return fail(FAILURE_INTENT_INVALID, "image artifact and trace identifiers must be canonical")
    if utf8_size(artifact_metadata["proposed_title"]) > MAX_ARTIFACT_TITLE_BYTES:
        return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image artifact title exceeds the UTF-8 byte budget")
    collections = (
        (style["reference_artifact_ids"], MAX_REFERENCE_ARTIFACTS),
        (constraints["must_include"], MAX_CONSTRAINT_ITEMS),
        (constraints["must_avoid"], MAX_CONSTRAINT_ITEMS),
        (safety["review_notes"], MAX_REVIEW_NOTES),
        (artifact_metadata["trace_ids"], MAX_TRACE_IDS),
    )
    for values, limit in collections:
        if len(values) > limit or any(utf8_size(item) > MAX_LIST_ITEM_BYTES for item in values):
            return fail(FAILURE_INTENT_BUDGET_EXCEEDED, "image intent list exceeds its item or byte budget")
    if constraints["mask_artifact_id"] is not None and constraints["edit_artifact_id"] is None:
        return fail(FAILURE_INTENT_INVALID, "mask artifact requires an edit artifact")
    return None


def validate_artifact_lineage(
    intent: dict[str, Any],
    profile: ImageBackendProfile,
    backend_request: dict[str, Any],
    artifact: dict[str, Any],
) -> str | None:
    if artifact["status"] != "generated":
        return "image backend did not return a generated artifact"
    if artifact["intent_id"] != intent["intent_id"] or artifact["backend_request_id"] != backend_request["request_id"]:
        return "image artifact intent or backend request lineage does not match"

    artifact_metadata = mapping_at(intent, "artifact_metadata")
    artifact_fields = mapping_at(artifact, "artifact")
    if not valid_identifier(artifact["artifact_id"]):
        return "image artifact id is not canonical"
    expected_uri = f"artifact://radishmind/generated/{artifact['artifact_id']}.{artifact_fields['format']}"
    if artifact_fields["uri"] != expected_uri:
        return "image artifact URI does not match the canonical artifact reference"
    if artifact_fields["title"] != artifact_metadata["proposed_title"] or artifact_fields["purpose"] != artifact_metadata["purpose"]:
        return "image artifact title or purpose does not match intent"
    output = mapping_at(backend_request, "output")
    for key in ("width", "height", "format"):
        if artifact_fields[key] != output[key]:
            return f"image artifact {key} does not match backend request"
    if artifact_fields["mime_type"] != MIME_BY_FORMAT[artifact_fields["format"]]:
        return "image artifact MIME type does not match format"

    generation = mapping_at(artifact, "generation")
    parameters = mapping_at(backend_request, "parameters")
    expected_generation = {
        "backend_id": profile.backend_id,
        "model": profile.model,
        "seed": parameters["seed"],
        "steps": parameters["steps"],
        "guidance_scale": parameters["guidance_scale"],
    }
    if generation != expected_generation:
        return "image artifact generation metadata does not match backend request"

    safety = mapping_at(artifact, "safety")
    if (
        safety["risk_level"] != "low"
        or safety["requires_confirmation"] is not False
        or safety["review_status"] not in {"not_required", "reviewed_pass"}
    ):
        return "image artifact did not pass the expected safety state"

    provenance = mapping_at(artifact, "provenance")
    if (
        provenance["source_request_id"] != intent["source_request_id"]
        or provenance["backend_request_id"] != backend_request["request_id"]
        or provenance["intent_id"] != intent["intent_id"]
        or provenance["trace_ids"] != mapping_at(backend_request, "trace")["trace_ids"]
    ):
        return "image artifact provenance does not match the canonical trace"
    return None


def observed_result_is_valid(result: ImageBackendInvocationResult) -> bool:
    return (
        bool(SHA256_RE.fullmatch(result.observed_sha256))
        and result.observed_mime_type in MIME_BY_FORMAT.values()
        and isinstance(result.observed_width, int)
        and not isinstance(result.observed_width, bool)
        and result.observed_width > 0
        and isinstance(result.observed_height, int)
        and not isinstance(result.observed_height, bool)
        and result.observed_height > 0
    )


def profile_is_valid(profile: ImageBackendProfile) -> bool:
    return isinstance(profile.enabled, bool) and all(
        valid_profile_value(value) for value in (profile.backend_id, profile.model, profile.adapter_profile)
    )


def valid_identifier(value: Any) -> bool:
    return isinstance(value, str) and utf8_size(value) <= MAX_IDENTIFIER_BYTES and bool(IDENTIFIER_RE.fullmatch(value))


def valid_profile_value(value: Any) -> bool:
    return (
        isinstance(value, str)
        and utf8_size(value) <= MAX_IDENTIFIER_BYTES
        and bool(PROFILE_VALUE_RE.fullmatch(value))
        and "://" not in value
        and ".." not in value
    )


def contains_sensitive_material(value: Any) -> bool:
    if isinstance(value, Mapping):
        return any(contains_sensitive_material(key) or contains_sensitive_material(item) for key, item in value.items())
    if isinstance(value, (list, tuple)):
        return any(contains_sensitive_material(item) for item in value)
    if isinstance(value, str):
        return any(pattern.search(value) for pattern in SENSITIVE_PATTERNS)
    return False


def contains_only_valid_utf8(value: Any) -> bool:
    if isinstance(value, Mapping):
        return all(contains_only_valid_utf8(key) and contains_only_valid_utf8(item) for key, item in value.items())
    if isinstance(value, (list, tuple)):
        return all(contains_only_valid_utf8(item) for item in value)
    if isinstance(value, str):
        try:
            value.encode("utf-8")
        except UnicodeEncodeError:
            return False
    return True


def profile_document(profile: ImageBackendProfile) -> dict[str, Any]:
    return {
        "backend_id": profile.backend_id,
        "model": profile.model,
        "adapter_profile": profile.adapter_profile,
    }


def canonical_strings(values: list[str]) -> list[str]:
    return list(dict.fromkeys(values))


def mapping_at(document: Mapping[str, Any], key: str) -> dict[str, Any]:
    value = document[key]
    if not isinstance(value, Mapping):
        raise TypeError(f"{key} must be a mapping")
    return dict(value)


def copy_mapping(value: Any) -> dict[str, Any] | None:
    if not isinstance(value, Mapping):
        return None
    return copy.deepcopy(dict(value))


def utf8_size(value: str) -> int:
    return len(value.encode("utf-8"))


@lru_cache(maxsize=3)
def schema_validator(path: Path) -> jsonschema.Draft202012Validator:
    schema = json.loads(path.read_text(encoding="utf-8"))
    jsonschema.Draft202012Validator.check_schema(schema)
    return jsonschema.Draft202012Validator(schema, format_checker=jsonschema.FormatChecker())


def validate_schema(document: Mapping[str, Any], path: Path) -> str:
    error = next(iter(schema_validator(path).iter_errors(document)), None)
    if error is None:
        return ""
    location = ".".join(str(item) for item in error.absolute_path) or "document"
    return f"{path.name} rejected {location}"


def success(
    backend_request: dict[str, Any],
    artifact_document: dict[str, Any],
    mapping: ImageArtifactMappingResult,
) -> ImageGenerationAdapterResult:
    return ImageGenerationAdapterResult(
        ok=True,
        backend_request=copy.deepcopy(backend_request),
        artifact_document=copy.deepcopy(artifact_document),
        citation=copy.deepcopy(mapping.citation),
        metadata_reference=copy.deepcopy(mapping.metadata_reference),
        backend_call_count=1,
        image_generation_count=1,
    )


def fail_after_call(
    code: str,
    message: str,
) -> ImageGenerationAdapterResult:
    # Do not expose the compiled prompt or untrusted artifact metadata after a
    # failed backend exchange.
    return fail(
        code,
        message,
        backend_call_count=1,
    )


def fail(
    code: str,
    message: str,
    *,
    backend_call_count: int = 0,
) -> ImageGenerationAdapterResult:
    return ImageGenerationAdapterResult(
        ok=False,
        failure_code=code,
        failure_message=message,
        backend_call_count=backend_call_count,
        image_generation_count=0,
    )
