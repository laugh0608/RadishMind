from __future__ import annotations

import copy
import json
import struct
import sys
import tempfile
import unittest
import zlib
from concurrent.futures import ThreadPoolExecutor
from dataclasses import asdict
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_artifact_delivery_coordinator import (  # noqa: E402
    FAILURE_BINARY_DELIVERY_ALREADY_CONSUMED,
    FAILURE_BINARY_DELIVERY_MISMATCH,
    FAILURE_PRIVATE_STORAGE,
    ImageArtifactDeliveryCoordinatorResult,
    coordinator_side_effect_counters,
    invoke_and_store_fixture_image,
)
from services.runtime.image_artifact_private_storage import (  # noqa: E402
    ImageArtifactStoreResult,
    LocalPrivateImageArtifactStore,
)
from services.runtime.image_backend_contract_fixture_client import (  # noqa: E402
    FAILURE_FIXTURE_DELIVERY_ALREADY_CONSUMED,
    FAILURE_FIXTURE_DELIVERY_MISMATCH,
    FAILURE_FIXTURE_DELIVERY_UNAVAILABLE,
    ContractFixtureImageBackendClient,
    FixtureBinaryDeliveryResult,
)
from services.runtime.image_backend_profile_configuration import (  # noqa: E402
    ImageBackendProfile,
    compile_image_backend_profile,
)
from services.runtime.image_generation_adapter import (  # noqa: E402
    FAILURE_BACKEND_RESPONSE_UNTRUSTED,
    FAILURE_BACKEND_SAFETY_BLOCKED,
    invoke_image_generation,
)


INTENT_FIXTURE = (
    REPO_ROOT / "scripts/checks/fixtures/image-generation-intent-basic.json"
)
PROFILE_FIXTURE = (
    REPO_ROOT
    / "scripts/checks/fixtures/"
    "image-backend-profile-source-contract-fixture.json"
)
BACKEND_REQUEST_ID = "image-backend-request-delivery-001"
CREATED_AT = "2026-07-26T00:00:00Z"


def load_intent() -> dict[str, Any]:
    return json.loads(INTENT_FIXTURE.read_text(encoding="utf-8"))


def compile_profile() -> ImageBackendProfile:
    source = json.loads(PROFILE_FIXTURE.read_text(encoding="utf-8"))
    result = compile_image_backend_profile(source)
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


def make_png(
    width: int = 1_024,
    height: int = 1_024,
    *,
    marker: bytes = b"delivery-fixture",
) -> bytes:
    header = struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0)
    return (
        b"\x89PNG\r\n\x1a\n"
        + png_chunk(b"IHDR", header)
        + png_chunk(b"IDAT", zlib.compress(marker))
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


class RaisingStore:
    def put(
        self,
        artifact_document: dict[str, Any],
        payload: bytes,
    ) -> ImageArtifactStoreResult:
        raise OSError("private path must not escape")


class MismatchDeliveryClient(ContractFixtureImageBackendClient):
    def deliver_binary(
        self,
        artifact_document: dict[str, Any],
        consumer,
    ) -> FixtureBinaryDeliveryResult:
        return FixtureBinaryDeliveryResult(
            ok=False,
            failure_code=FAILURE_FIXTURE_DELIVERY_MISMATCH,
            failure_message="internal mismatch detail",
        )


class ImageArtifactDeliveryCoordinatorTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary_directory.name) / "private-artifacts"
        self.profile = compile_profile()

    def tearDown(self) -> None:
        self.temporary_directory.cleanup()

    def client(
        self,
        payload: bytes | bytearray,
    ) -> ContractFixtureImageBackendClient:
        return ContractFixtureImageBackendClient(
            self.profile,
            payload,
            created_at=CREATED_AT,
        )

    def invoke(
        self,
        payload: bytes,
        *,
        image_format: str = "png",
        store: Any | None = None,
        client: ContractFixtureImageBackendClient | None = None,
    ) -> ImageArtifactDeliveryCoordinatorResult:
        intent = load_intent()
        intent["output"]["format"] = image_format
        return invoke_and_store_fixture_image(
            intent,
            profile=self.profile,
            client=client or self.client(payload),
            store=store or LocalPrivateImageArtifactStore(self.root),
            backend_request_id=BACKEND_REQUEST_ID,
        )

    def test_png_jpeg_and_webp_persist_before_references_are_released(
        self,
    ) -> None:
        cases = (
            ("png", make_png()),
            ("jpg", make_jpeg()),
            ("webp", make_webp()),
        )
        for image_format, payload in cases:
            with self.subTest(image_format=image_format):
                root = self.root / image_format
                store = LocalPrivateImageArtifactStore(root)
                result = self.invoke(
                    payload,
                    image_format=image_format,
                    store=store,
                )

                self.assertTrue(result.ok)
                self.assertIsNotNone(result.citation)
                self.assertIsNotNone(result.metadata_reference)
                self.assertTrue(store.lookup(result.metadata_reference).ok)
                self.assertEqual(result.backend_call_count, 1)
                self.assertEqual(result.image_generation_count, 0)
                self.assertEqual(result.artifact_binary_delivery_count, 1)
                self.assertEqual(result.local_artifact_store_write_count, 1)
                self.assertEqual(
                    result.artifact_binary_revalidation_count,
                    1,
                )
                serialized = json.dumps(asdict(result), sort_keys=True)
                for forbidden in (
                    "base64",
                    "binary_payload",
                    "file_path",
                    "image_bytes",
                    "storage_ref",
                    str(root),
                ):
                    self.assertNotIn(forbidden, serialized)
                counters = coordinator_side_effect_counters(result)
                for key in (
                    "retry_count",
                    "fallback_count",
                    "artifact_upload_count",
                    "public_url_resolution_count",
                    "production_storage_write_count",
                ):
                    self.assertEqual(counters[key], 0)

    def test_same_input_and_fresh_client_reuse_immutable_storage_binding(
        self,
    ) -> None:
        payload = make_png()
        first = self.invoke(payload)
        second = self.invoke(payload)

        self.assertTrue(first.ok)
        self.assertTrue(second.ok)
        self.assertEqual(first.citation, second.citation)
        self.assertEqual(first.metadata_reference, second.metadata_reference)
        self.assertEqual(first.local_artifact_store_write_count, 1)
        self.assertEqual(second.local_artifact_store_write_count, 0)
        self.assertEqual(second.artifact_binary_revalidation_count, 1)

    def test_adapter_failures_never_request_delivery_or_release_references(
        self,
    ) -> None:
        unsafe_intent = load_intent()
        unsafe_intent["safety"]["risk_level"] = "medium"
        unsafe = invoke_and_store_fixture_image(
            unsafe_intent,
            profile=self.profile,
            client=self.client(make_png()),
            store=LocalPrivateImageArtifactStore(self.root),
            backend_request_id=BACKEND_REQUEST_ID,
        )
        mismatched_intent = load_intent()
        mismatched_intent["output"]["width"] = 512
        backend_failure = invoke_and_store_fixture_image(
            mismatched_intent,
            profile=self.profile,
            client=self.client(make_png()),
            store=LocalPrivateImageArtifactStore(self.root),
            backend_request_id=BACKEND_REQUEST_ID,
        )

        self.assertEqual(unsafe.failure_code, FAILURE_BACKEND_SAFETY_BLOCKED)
        self.assertEqual(unsafe.backend_call_count, 0)
        self.assertEqual(backend_failure.failure_code, FAILURE_BACKEND_RESPONSE_UNTRUSTED)
        self.assertEqual(backend_failure.backend_call_count, 1)
        for result in (unsafe, backend_failure):
            self.assertFalse(result.ok)
            self.assertIsNone(result.citation)
            self.assertIsNone(result.metadata_reference)
            self.assertEqual(result.artifact_binary_delivery_count, 0)
            self.assertEqual(result.local_artifact_store_write_count, 0)
            self.assertEqual(result.artifact_binary_revalidation_count, 0)

    def test_client_rejects_unbound_mismatched_and_repeated_delivery(
        self,
    ) -> None:
        client = self.client(make_png())
        invoked_client = self.client(make_png())
        invoked_adapter = invoke_image_generation(
            load_intent(),
            profile=self.profile,
            client=invoked_client,
            backend_request_id=BACKEND_REQUEST_ID,
        )
        unbound = client.deliver_binary(
            invoked_adapter.artifact_document,
            lambda payload: None,
        )
        self.assertEqual(
            unbound.failure_code,
            FAILURE_FIXTURE_DELIVERY_UNAVAILABLE,
        )

        adapter = invoke_image_generation(
            load_intent(),
            profile=self.profile,
            client=client,
            backend_request_id=BACKEND_REQUEST_ID,
        )
        self.assertTrue(adapter.ok)
        drifted = copy.deepcopy(adapter.artifact_document)
        drifted["artifact"]["sha256"] = "0" * 64
        mismatch = client.deliver_binary(drifted, lambda payload: None)
        self.assertEqual(
            mismatch.failure_code,
            FAILURE_FIXTURE_DELIVERY_MISMATCH,
        )
        wrong_artifact = copy.deepcopy(adapter.artifact_document)
        wrong_artifact["artifact_id"] = "image-fixture-wrong-artifact"
        wrong_artifact["artifact"]["uri"] = (
            "artifact://radishmind/generated/"
            "image-fixture-wrong-artifact.png"
        )
        wrong = client.deliver_binary(
            wrong_artifact,
            lambda payload: None,
        )
        self.assertEqual(
            wrong.failure_code,
            FAILURE_FIXTURE_DELIVERY_MISMATCH,
        )
        delivered: list[int] = []
        success = client.deliver_binary(
            adapter.artifact_document,
            lambda payload: delivered.append(len(payload)),
        )
        repeated = client.deliver_binary(
            adapter.artifact_document,
            lambda payload: None,
        )

        self.assertTrue(success.ok)
        self.assertEqual(len(delivered), 1)
        self.assertEqual(
            repeated.failure_code,
            FAILURE_FIXTURE_DELIVERY_ALREADY_CONSUMED,
        )
        self.assertEqual(repeated.artifact_binary_delivery_count, 0)

    def test_delivery_consumer_failure_is_consumed_without_retry(
        self,
    ) -> None:
        client = self.client(make_png())
        adapter = invoke_image_generation(
            load_intent(),
            profile=self.profile,
            client=client,
            backend_request_id=BACKEND_REQUEST_ID,
        )

        def fail_consumer(payload: bytes) -> None:
            raise RuntimeError("consumer detail must not escape")

        failed = client.deliver_binary(
            adapter.artifact_document,
            fail_consumer,
        )
        repeated = client.deliver_binary(
            adapter.artifact_document,
            lambda payload: None,
        )

        self.assertEqual(
            failed.failure_code,
            FAILURE_FIXTURE_DELIVERY_UNAVAILABLE,
        )
        self.assertEqual(failed.artifact_binary_delivery_count, 1)
        self.assertEqual(failed.binary_consumer_call_count, 1)
        self.assertNotIn("consumer detail", failed.failure_message)
        self.assertEqual(
            repeated.failure_code,
            FAILURE_FIXTURE_DELIVERY_ALREADY_CONSUMED,
        )

    def test_coordinator_sanitizes_delivery_mismatch(self) -> None:
        client = MismatchDeliveryClient(
            self.profile,
            make_png(),
            created_at=CREATED_AT,
        )

        result = self.invoke(make_png(), client=client)

        self.assertFalse(result.ok)
        self.assertEqual(
            result.failure_code,
            FAILURE_BINARY_DELIVERY_MISMATCH,
        )
        self.assertNotIn("internal mismatch", result.failure_message)
        self.assertIsNone(result.citation)
        self.assertIsNone(result.metadata_reference)
        self.assertEqual(result.backend_call_count, 1)
        self.assertEqual(result.artifact_binary_delivery_count, 0)

    def test_store_failure_and_payload_drift_fail_without_references(
        self,
    ) -> None:
        unavailable = self.invoke(make_png(), store=RaisingStore())
        drifted_client = self.client(make_png())
        drifted_client._fixture_payload = make_png(marker=b"drifted")
        drifted = self.invoke(
            make_png(),
            client=drifted_client,
        )

        for result in (unavailable, drifted):
            self.assertFalse(result.ok)
            self.assertEqual(result.failure_code, FAILURE_PRIVATE_STORAGE)
            self.assertIsNone(result.citation)
            self.assertIsNone(result.metadata_reference)
            self.assertEqual(result.backend_call_count, 1)
            self.assertEqual(result.artifact_binary_delivery_count, 1)
            self.assertEqual(result.local_artifact_store_write_count, 0)
        self.assertEqual(
            unavailable.artifact_binary_revalidation_count,
            0,
        )
        self.assertEqual(drifted.artifact_binary_revalidation_count, 1)
        self.assertNotIn(
            str(self.root),
            json.dumps(asdict(unavailable), sort_keys=True),
        )

    def test_store_conflict_is_sanitized_and_preserves_exact_counts(
        self,
    ) -> None:
        first = self.invoke(make_png(marker=b"first"))
        conflicting = self.invoke(make_png(marker=b"second"))

        self.assertTrue(first.ok)
        self.assertFalse(conflicting.ok)
        self.assertEqual(conflicting.failure_code, FAILURE_PRIVATE_STORAGE)
        self.assertEqual(conflicting.artifact_binary_delivery_count, 1)
        self.assertEqual(conflicting.local_artifact_store_write_count, 0)
        self.assertEqual(
            conflicting.artifact_binary_revalidation_count,
            1,
        )
        self.assertIsNone(conflicting.citation)
        self.assertIsNone(conflicting.metadata_reference)

    def test_repeated_coordinator_call_fails_closed_after_delivery(
        self,
    ) -> None:
        payload = make_png()
        client = self.client(payload)
        first = self.invoke(payload, client=client)
        repeated = self.invoke(payload, client=client)

        self.assertTrue(first.ok)
        self.assertFalse(repeated.ok)
        self.assertEqual(
            repeated.failure_code,
            FAILURE_BINARY_DELIVERY_ALREADY_CONSUMED,
        )
        self.assertEqual(repeated.backend_call_count, 1)
        self.assertEqual(repeated.artifact_binary_delivery_count, 0)
        self.assertIsNone(repeated.citation)
        self.assertIsNone(repeated.metadata_reference)

    def test_concurrent_delivery_has_exactly_one_consumer_and_store_write(
        self,
    ) -> None:
        payload = make_png()
        client = self.client(payload)
        store = LocalPrivateImageArtifactStore(self.root)

        def invoke_once(_: int):
            return self.invoke(payload, client=client, store=store)

        with ThreadPoolExecutor(max_workers=8) as executor:
            results = list(executor.map(invoke_once, range(16)))

        successful = [result for result in results if result.ok]
        rejected = [result for result in results if not result.ok]
        self.assertEqual(len(successful), 1)
        self.assertEqual(len(rejected), 15)
        self.assertEqual(
            sum(result.artifact_binary_delivery_count for result in results),
            1,
        )
        self.assertEqual(
            sum(result.local_artifact_store_write_count for result in results),
            1,
        )
        self.assertTrue(
            all(
                result.failure_code
                == FAILURE_BINARY_DELIVERY_ALREADY_CONSUMED
                for result in rejected
            )
        )

    def test_mutable_fixture_input_is_copied_before_delivery(self) -> None:
        source = bytearray(make_png())
        client = self.client(source)
        source[:] = b"corrupted-after-construction"

        result = self.invoke(make_png(), client=client)

        self.assertTrue(result.ok)
        self.assertEqual(result.artifact_binary_delivery_count, 1)
        self.assertEqual(result.artifact_binary_revalidation_count, 1)


if __name__ == "__main__":
    unittest.main()
