from __future__ import annotations

import copy
from collections.abc import Mapping
from dataclasses import dataclass
from typing import Any

from services.runtime.image_artifact_runtime_mapper import ImageArtifactMappingResult


ARTIFACT_URI_SCHEME = "artifact://"

FAILURE_MAPPER_FAILED = "image_artifact_mapper_failed"
FAILURE_CITATION_ID_CONFLICT = "image_artifact_citation_id_conflict"
FAILURE_CITATION_SCHEMA_INVALID = "image_artifact_citation_schema_invalid"
FAILURE_METADATA_REFERENCE_LEAK = "image_artifact_metadata_reference_leak"

FORBIDDEN_RESPONSE_METADATA_FIELDS = {
    "artifact_metadata",
    "artifacts",
    "image_artifact_metadata",
    "metadata_reference",
}
FORBIDDEN_METADATA_REFERENCE_KEYS = {
    "base64_image",
    "binary_payload",
    "image_bytes",
    "pixel_payload",
    "provider_raw_dump",
    "provider_raw_response",
    "public_url",
    "raw_provider_response",
    "signed_public_url",
    "signed_url",
}
ALLOWED_CITATION_FIELDS = {"id", "kind", "label", "locator", "source_uri", "excerpt"}
REQUIRED_ARTIFACT_CITATION_FIELDS = ("id", "kind", "label", "locator", "source_uri")

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
    "response_consumer_call_count": 0,
    "response_builder_change_count": 0,
    "executor_call_count": 0,
    "confirmation_call_count": 0,
    "business_writeback_count": 0,
    "replay_call_count": 0,
}


@dataclass(frozen=True)
class ImageArtifactResponseConsumerResult:
    ok: bool
    response_document: dict[str, Any]
    metadata_reference: dict[str, Any] | None = None
    failure_code: str | None = None
    failure_message: str = ""


def response_consumer_side_effect_counters() -> dict[str, int]:
    return dict(ZERO_SIDE_EFFECT_COUNTERS)


def apply_image_artifact_reference_to_response(
    response_document: Mapping[str, Any],
    mapping_result: ImageArtifactMappingResult,
) -> ImageArtifactResponseConsumerResult:
    if not isinstance(response_document, Mapping):
        return fail(FAILURE_CITATION_SCHEMA_INVALID, {}, "response document must be a mapping")

    response_copy = copy_response_document(response_document)
    if response_top_level_has_metadata_leak(response_copy):
        return fail(
            FAILURE_METADATA_REFERENCE_LEAK,
            response_copy,
            "response document must not expose artifact metadata handoff fields",
        )

    if mapping_result.ok is not True or mapping_result.citation is None:
        message = mapping_result.failure_message or "image artifact mapper did not return a success citation"
        return fail(FAILURE_MAPPER_FAILED, response_copy, message)
    if not isinstance(mapping_result.metadata_reference, Mapping):
        return fail(
            FAILURE_MAPPER_FAILED,
            response_copy,
            "image artifact mapper did not return an internal metadata reference",
        )

    metadata_reference = copy.deepcopy(dict(mapping_result.metadata_reference))
    if metadata_reference_has_public_or_binary_leak(metadata_reference):
        return fail(
            FAILURE_METADATA_REFERENCE_LEAK,
            response_copy,
            "metadata reference contains public URL, signed URL, binary payload, or provider raw data",
        )

    citation = copy.deepcopy(mapping_result.citation)
    if not artifact_citation_is_valid(citation):
        return fail(FAILURE_CITATION_SCHEMA_INVALID, response_copy, "artifact citation shape is invalid")

    existing_citations = response_copy.get("citations", [])
    if not isinstance(existing_citations, list):
        return fail(FAILURE_CITATION_SCHEMA_INVALID, response_copy, "response citations must be a list")

    existing_ids = citation_ids(existing_citations)
    if existing_ids is None:
        return fail(FAILURE_CITATION_SCHEMA_INVALID, response_copy, "existing citations must have valid ids")
    if citation["id"] in existing_ids:
        return fail(FAILURE_CITATION_ID_CONFLICT, response_copy, "artifact citation id already exists")

    next_response = copy_response_document(response_copy)
    next_citations = copy.deepcopy(existing_citations)
    next_citations.append(citation)
    next_response["citations"] = next_citations
    if response_top_level_has_metadata_leak(next_response):
        return fail(
            FAILURE_METADATA_REFERENCE_LEAK,
            response_copy,
            "response document must not expose artifact metadata handoff fields",
        )
    return ImageArtifactResponseConsumerResult(
        ok=True,
        response_document=next_response,
        metadata_reference=metadata_reference,
    )


def copy_response_document(response_document: Mapping[str, Any]) -> dict[str, Any]:
    return copy.deepcopy(dict(response_document))


def fail(failure_code: str, response_document: dict[str, Any], message: str) -> ImageArtifactResponseConsumerResult:
    return ImageArtifactResponseConsumerResult(
        ok=False,
        response_document=response_document,
        failure_code=failure_code,
        failure_message=message,
    )


def response_top_level_has_metadata_leak(response_document: Mapping[str, Any]) -> bool:
    return any(key in response_document for key in FORBIDDEN_RESPONSE_METADATA_FIELDS)


def metadata_reference_has_public_or_binary_leak(value: Any) -> bool:
    if isinstance(value, Mapping):
        if any(key in value for key in FORBIDDEN_METADATA_REFERENCE_KEYS):
            return True
        return any(metadata_reference_has_public_or_binary_leak(item) for item in value.values())
    if isinstance(value, list):
        return any(metadata_reference_has_public_or_binary_leak(item) for item in value)
    if isinstance(value, str):
        return value.startswith(("http://", "https://"))
    return False


def artifact_citation_is_valid(citation: Any) -> bool:
    if not isinstance(citation, Mapping):
        return False
    if any(key not in ALLOWED_CITATION_FIELDS for key in citation):
        return False
    for field in REQUIRED_ARTIFACT_CITATION_FIELDS:
        if not normalized_string(citation.get(field)):
            return False
    if normalized_string(citation.get("kind")) != "artifact":
        return False
    if not artifact_uri_string(citation.get("locator")):
        return False
    if not artifact_uri_string(citation.get("source_uri")):
        return False
    if "excerpt" in citation and not isinstance(citation.get("excerpt"), str):
        return False
    return True


def citation_ids(citations: list[Any]) -> set[str] | None:
    ids: set[str] = set()
    for citation in citations:
        if not isinstance(citation, Mapping):
            return None
        citation_id = normalized_string(citation.get("id"))
        if not citation_id:
            return None
        ids.add(citation_id)
    return ids


def artifact_uri_string(value: Any) -> str:
    text = normalized_string(value)
    if text.startswith(ARTIFACT_URI_SCHEME):
        return text
    return ""


def normalized_string(value: Any) -> str:
    if isinstance(value, str):
        return " ".join(value.split()).strip()
    return ""
