from __future__ import annotations

import copy
import hashlib
import json
import struct
import sys
import unittest
import zlib
from dataclasses import asdict
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_backend_contract_fixture_client import (  # noqa: E402
    FAILURE_FIXTURE_BINARY_INVALID,
    FAILURE_FIXTURE_PROFILE_INVALID,
    FAILURE_FIXTURE_REQUEST_INVALID,
    FAILURE_FIXTURE_REQUEST_MISMATCH,
    ContractFixtureClientError,
    ContractFixtureImageBackendClient,
)
from services.runtime.image_backend_profile_configuration import (  # noqa: E402
    ImageBackendProfile,
    compile_image_backend_profile,
)
from services.runtime.image_generation_adapter import (  # noqa: E402
    FAILURE_BACKEND_RESPONSE_UNTRUSTED,
    adapter_side_effect_counters,
    compile_image_backend_request,
    invoke_image_generation,
)


INTENT_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
PROFILE_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-backend-profile-source-basic.json"
CONTRACT_FIXTURE_PROFILE = (
    REPO_ROOT
    / "scripts/checks/fixtures/image-backend-profile-source-contract-fixture.json"
)
BACKEND_REQUEST_ID = "image-backend-request-fixture-001"
CREATED_AT = "2026-07-25T00:00:00Z"


def load_intent() -> dict[str, Any]:
    return json.loads(INTENT_FIXTURE.read_text(encoding="utf-8"))


def load_profile_source() -> dict[str, Any]:
    return json.loads(CONTRACT_FIXTURE_PROFILE.read_text(encoding="utf-8"))


def compile_profile(source: dict[str, Any] | None = None) -> ImageBackendProfile:
    result = compile_image_backend_profile(source or load_profile_source())
    if not result.ok or result.profile is None:
        raise AssertionError(f"{result.failure_code}: {result.failure_message}")
    return result.profile


def png_chunk(kind: bytes, data: bytes) -> bytes:
    return (
        struct.pack(">I", len(data))
        + kind
        + data
        + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)
    )


def make_png(width: int = 1_024, height: int = 1_024) -> bytes:
    header = struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0)
    return (
        b"\x89PNG\r\n\x1a\n"
        + png_chunk(b"IHDR", header)
        + png_chunk(b"IDAT", zlib.compress(b"contract-fixture"))
        + png_chunk(b"IEND", b"")
    )


def make_jpeg(width: int = 1_024, height: int = 1_024) -> bytes:
    start_of_frame = (
        b"\xff\xc0"
        + struct.pack(">H", 11)
        + b"\x08"
        + struct.pack(">HH", height, width)
        + b"\x01\x01\x11\x00"
    )
    scan = b"\xff\xda\x00\x08\x01\x01\x00\x00\x3f\x00\x00"
    return b"\xff\xd8" + start_of_frame + scan + b"\xff\xd9"


def make_webp(width: int = 1_024, height: int = 1_024) -> bytes:
    packed = (width - 1) | ((height - 1) << 14)
    chunk_data = b"\x2f" + packed.to_bytes(4, "little") + b"\x00\x00\x00\x00"
    chunk = b"VP8L" + struct.pack("<I", len(chunk_data)) + chunk_data + b"\x00"
    payload = b"WEBP" + chunk
    return b"RIFF" + struct.pack("<I", len(payload)) + payload


def compile_request(
    profile: ImageBackendProfile,
    intent: dict[str, Any] | None = None,
) -> dict[str, Any]:
    return compile_image_backend_request(
        intent or load_intent(),
        profile=profile,
        backend_request_id=BACKEND_REQUEST_ID,
    )


class ContractFixtureImageBackendClientTest(unittest.TestCase):
    def test_fixture_client_is_deterministic_and_observes_real_container(self) -> None:
        profile = compile_profile()
        cases = (
            ("png", "image/png", make_png()),
            ("jpg", "image/jpeg", make_jpeg()),
            ("webp", "image/webp", make_webp()),
        )
        for image_format, mime_type, payload in cases:
            with self.subTest(image_format=image_format):
                client = ContractFixtureImageBackendClient(
                    profile,
                    payload,
                    created_at=CREATED_AT,
                )
                request = compile_request(profile)
                request["output"]["format"] = image_format
                original = copy.deepcopy(request)

                first = client.invoke(request, profile.timeout_seconds)
                second = client.invoke(
                    copy.deepcopy(request),
                    profile.timeout_seconds,
                )

                self.assertEqual(first, second)
                self.assertEqual(request, original)
                self.assertEqual(
                    first.observed_sha256,
                    hashlib.sha256(payload).hexdigest(),
                )
                self.assertEqual(first.observed_mime_type, mime_type)
                self.assertEqual(
                    (first.observed_width, first.observed_height),
                    (1_024, 1_024),
                )
                self.assertTrue(first.artifact_id.startswith("image-fixture-"))
                serialized = json.dumps(asdict(first), sort_keys=True)
                self.assertNotIn("payload", serialized)
                self.assertNotIn("base64", serialized)
                self.assertNotIn("endpoint", serialized)
                self.assertNotIn("credential", serialized)
                self.assertNotIn("title", serialized)
                self.assertNotIn("purpose", serialized)
                self.assertNotIn("provenance", serialized)
                self.assertNotIn("safety", serialized)

    def test_adapter_owns_canonical_artifact_metadata(self) -> None:
        profile = compile_profile()
        client = ContractFixtureImageBackendClient(
            profile,
            make_png(),
            created_at=CREATED_AT,
        )
        intent = load_intent()

        result = invoke_image_generation(
            intent,
            profile=profile,
            client=client,
            backend_request_id=BACKEND_REQUEST_ID,
        )

        self.assertTrue(result.ok)
        artifact = result.artifact_document
        self.assertEqual(
            artifact["artifact"]["title"],
            intent["artifact_metadata"]["proposed_title"],
        )
        self.assertEqual(
            artifact["artifact"]["purpose"],
            intent["artifact_metadata"]["purpose"],
        )
        self.assertEqual(
            artifact["provenance"]["trace_ids"],
            result.backend_request["trace"]["trace_ids"],
        )
        self.assertEqual(artifact["created_at"], CREATED_AT)
        self.assertEqual(result.backend_call_count, 1)
        self.assertEqual(result.image_generation_count, 0)
        counters = adapter_side_effect_counters(result)
        self.assertEqual(counters["retry_count"], 0)
        self.assertEqual(counters["fallback_count"], 0)
        self.assertEqual(counters["artifact_upload_count"], 0)
        self.assertEqual(counters["public_url_resolution_count"], 0)

    def test_client_requires_test_only_contract_fixture_profile(self) -> None:
        remote = json.loads(PROFILE_FIXTURE.read_text(encoding="utf-8"))
        remote_profile = compile_profile(remote)
        development = load_profile_source()
        development["environment"] = "development"
        invalid_development = compile_image_backend_profile(development)

        with self.assertRaisesRegex(
            ContractFixtureClientError,
            FAILURE_FIXTURE_PROFILE_INVALID,
        ):
            ContractFixtureImageBackendClient(
                remote_profile,
                make_png(),
                created_at=CREATED_AT,
            )
        self.assertFalse(invalid_development.ok)

    def test_client_rejects_invalid_binary_and_timestamp(self) -> None:
        profile = compile_profile()
        for name, payload in (
            ("empty", b""),
            ("invalid", b"not-an-image"),
            ("oversized-dimensions", make_png(width=2_049, height=1)),
        ):
            with self.subTest(name=name):
                with self.assertRaisesRegex(
                    ContractFixtureClientError,
                    FAILURE_FIXTURE_BINARY_INVALID,
                ):
                    ContractFixtureImageBackendClient(
                        profile,
                        payload,
                        created_at=CREATED_AT,
                    )

        with self.assertRaisesRegex(
            ContractFixtureClientError,
            FAILURE_FIXTURE_PROFILE_INVALID,
        ):
            ContractFixtureImageBackendClient(
                profile,
                make_png(),
                created_at="2026-07-25 00:00:00",
            )

    def test_client_rejects_strict_or_sensitive_request_before_fixture_use(self) -> None:
        profile = compile_profile()
        client = ContractFixtureImageBackendClient(
            profile,
            make_png(),
            created_at=CREATED_AT,
        )
        cases = []
        unknown = compile_request(profile)
        unknown["provider_raw_response"] = {}
        cases.append(("unknown", unknown))
        sensitive = compile_request(profile)
        sensitive["prompt"]["positive"] = "token=secret-value"
        cases.append(("sensitive", sensitive))
        invalid_utf8 = compile_request(profile)
        invalid_utf8["prompt"]["positive"] = "\ud800"
        cases.append(("utf8", invalid_utf8))
        oversized = compile_request(profile)
        oversized["prompt"]["positive"] = "x" * (64 * 1_024)
        cases.append(("byte-budget", oversized))

        for name, request in cases:
            with self.subTest(name=name):
                with self.assertRaisesRegex(
                    ContractFixtureClientError,
                    FAILURE_FIXTURE_REQUEST_INVALID,
                ):
                    client.invoke(request, profile.timeout_seconds)

    def test_client_rejects_profile_timeout_output_input_and_safety_drift(self) -> None:
        profile = compile_profile()
        client = ContractFixtureImageBackendClient(
            profile,
            make_png(),
            created_at=CREATED_AT,
        )
        cases = []
        backend = compile_request(profile)
        backend["backend"]["model"] = "other-model"
        cases.append(("backend", backend, profile.timeout_seconds))
        output = compile_request(profile)
        output["output"]["width"] = 512
        cases.append(("output", output, profile.timeout_seconds))
        image_format = compile_request(profile)
        image_format["output"]["format"] = "jpg"
        cases.append(("format", image_format, profile.timeout_seconds))
        inputs = compile_request(profile)
        inputs["inputs"]["reference_artifact_ids"] = ["artifact-reference-001"]
        cases.append(("inputs", inputs, profile.timeout_seconds))
        safety = compile_request(profile)
        safety["safety"]["risk_level"] = "medium"
        cases.append(("safety", safety, profile.timeout_seconds))
        timeout = compile_request(profile)
        cases.append(("timeout", timeout, profile.timeout_seconds + 1))

        for name, request, timeout_seconds in cases:
            with self.subTest(name=name):
                with self.assertRaisesRegex(
                    ContractFixtureClientError,
                    FAILURE_FIXTURE_REQUEST_MISMATCH,
                ):
                    client.invoke(request, timeout_seconds)

    def test_adapter_sanitizes_fixture_mismatch_without_retry_or_fallback(self) -> None:
        profile = compile_profile()
        client = ContractFixtureImageBackendClient(
            profile,
            make_png(),
            created_at=CREATED_AT,
        )
        intent = load_intent()
        intent["output"]["width"] = 512

        result = invoke_image_generation(
            intent,
            profile=profile,
            client=client,
            backend_request_id=BACKEND_REQUEST_ID,
        )

        self.assertEqual(result.failure_code, FAILURE_BACKEND_RESPONSE_UNTRUSTED)
        self.assertEqual(result.backend_call_count, 1)
        self.assertIsNone(result.backend_request)
        self.assertIsNone(result.artifact_document)
        self.assertNotIn(FAILURE_FIXTURE_REQUEST_MISMATCH, result.failure_message)
        counters = adapter_side_effect_counters(result)
        self.assertEqual(counters["retry_count"], 0)
        self.assertEqual(counters["fallback_count"], 0)
        self.assertEqual(counters["artifact_store_lookup_count"], 0)


if __name__ == "__main__":
    unittest.main()
