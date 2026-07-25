from __future__ import annotations

import copy
import hashlib
import json
import math
from collections.abc import Mapping
from typing import Any

from .image_artifact_binary_inspection import (
    ImageBinaryObservation,
    inspect_image_binary,
)
from .image_backend_profile_configuration import (
    ImageBackendProfile,
    compiled_profile_is_valid,
)
from .image_generation_adapter import (
    BACKEND_REQUEST_SCHEMA_PATH,
    ImageBackendInvocationResult,
    contains_only_valid_utf8,
    contains_sensitive_material,
    valid_created_at,
    validate_schema,
)


FAILURE_FIXTURE_PROFILE_INVALID = "image_backend_fixture_profile_invalid"
FAILURE_FIXTURE_BINARY_INVALID = "image_backend_fixture_binary_invalid"
FAILURE_FIXTURE_REQUEST_INVALID = "image_backend_fixture_request_invalid"
FAILURE_FIXTURE_REQUEST_MISMATCH = "image_backend_fixture_request_mismatch"

MAX_FIXTURE_BYTES = 32 * 1024 * 1024
MAX_BACKEND_REQUEST_BYTES = 64 * 1024
MAX_DIMENSION = 2_048
MAX_PIXELS = 4_194_304
FORMAT_BY_MIME = {
    "image/png": "png",
    "image/jpeg": "jpg",
    "image/webp": "webp",
}


class ContractFixtureClientError(ValueError):
    pass


class ContractFixtureImageBackendClient:
    """Test-only concrete client backed by one inspected image fixture."""

    def __init__(
        self,
        profile: ImageBackendProfile,
        fixture_payload: bytes | bytearray | memoryview,
        *,
        created_at: str,
    ) -> None:
        if (
            not compiled_profile_is_valid(profile)
            or not profile.enabled
            or profile.environment != "test"
            or profile.runtime_mode != "contract_fixture"
        ):
            raise ContractFixtureClientError(FAILURE_FIXTURE_PROFILE_INVALID)
        payload = normalize_payload(fixture_payload)
        if payload is None or not payload or len(payload) > MAX_FIXTURE_BYTES:
            raise ContractFixtureClientError(FAILURE_FIXTURE_BINARY_INVALID)
        try:
            observation = inspect_image_binary(payload)
        except ValueError as error:
            raise ContractFixtureClientError(
                FAILURE_FIXTURE_BINARY_INVALID
            ) from error
        if (
            observation.width > MAX_DIMENSION
            or observation.height > MAX_DIMENSION
            or observation.width * observation.height > MAX_PIXELS
        ):
            raise ContractFixtureClientError(FAILURE_FIXTURE_BINARY_INVALID)
        if not valid_created_at(created_at):
            raise ContractFixtureClientError(FAILURE_FIXTURE_PROFILE_INVALID)

        self._profile = profile
        self._observation = observation
        self._created_at = created_at

    def invoke(
        self,
        request_document: Mapping[str, Any],
        timeout_seconds: float,
    ) -> ImageBackendInvocationResult:
        request = copy_mapping(request_document)
        if request is None or not contains_only_valid_utf8(request):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_INVALID)
        request_bytes = canonical_json_bytes(request)
        if request_bytes is None or len(request_bytes) > MAX_BACKEND_REQUEST_BYTES:
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_INVALID)
        if validate_schema(request, BACKEND_REQUEST_SCHEMA_PATH):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_INVALID)
        if contains_sensitive_material(request):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_INVALID)
        if not valid_timeout(timeout_seconds):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_INVALID)
        if float(timeout_seconds) != self._profile.timeout_seconds:
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_MISMATCH)
        if not request_matches_profile(request, self._profile):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_MISMATCH)
        if not request_matches_fixture(request, self._observation):
            raise ContractFixtureClientError(FAILURE_FIXTURE_REQUEST_MISMATCH)

        request_digest = hashlib.sha256(request_bytes).hexdigest()
        return ImageBackendInvocationResult(
            artifact_id=f"image-fixture-{request_digest[:24]}",
            created_at=self._created_at,
            observed_sha256=self._observation.sha256,
            observed_mime_type=self._observation.mime_type,
            observed_width=self._observation.width,
            observed_height=self._observation.height,
        )


def request_matches_profile(
    request: Mapping[str, Any],
    profile: ImageBackendProfile,
) -> bool:
    backend = request["backend"]
    return backend == {
        "id": profile.backend_id,
        "model": profile.model,
        "adapter_profile": profile.adapter_profile,
    }


def request_matches_fixture(
    request: Mapping[str, Any],
    observation: ImageBinaryObservation,
) -> bool:
    output = request["output"]
    inputs = request["inputs"]
    safety = request["safety"]
    return (
        output["count"] == 1
        and output["width"] == observation.width
        and output["height"] == observation.height
        and output["format"] == FORMAT_BY_MIME[observation.mime_type]
        and inputs["reference_artifact_ids"] == []
        and inputs["edit_artifact_id"] is None
        and inputs["mask_artifact_id"] is None
        and safety["gate"] == "approved_for_backend"
        and safety["requires_confirmation"] is False
        and safety["risk_level"] == "low"
    )


def valid_timeout(value: Any) -> bool:
    return (
        isinstance(value, (int, float))
        and not isinstance(value, bool)
        and math.isfinite(float(value))
        and float(value) > 0
    )


def normalize_payload(value: Any) -> bytes | None:
    if isinstance(value, bytes):
        return value
    if isinstance(value, (bytearray, memoryview)):
        return bytes(value)
    return None


def copy_mapping(value: Any) -> dict[str, Any] | None:
    if not isinstance(value, Mapping):
        return None
    return copy.deepcopy(dict(value))


def canonical_json_bytes(value: Mapping[str, Any]) -> bytes | None:
    try:
        return json.dumps(
            value,
            ensure_ascii=False,
            sort_keys=True,
            separators=(",", ":"),
            allow_nan=False,
        ).encode("utf-8")
    except (TypeError, ValueError, UnicodeEncodeError):
        return None
