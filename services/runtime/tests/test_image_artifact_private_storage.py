from __future__ import annotations

import copy
import hashlib
import json
import struct
import sys
import tempfile
import unittest
import zlib
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.image_artifact_private_storage import (  # noqa: E402
    FAILURE_BINARY_CONSUMER_FAILED,
    FAILURE_BINARY_INVALID,
    FAILURE_BINARY_READ_FORBIDDEN,
    FAILURE_BINARY_READER_UNAVAILABLE,
    FAILURE_BINARY_TOO_LARGE,
    FAILURE_STORE_CONFLICT,
    FAILURE_STORE_INTEGRITY,
    FAILURE_STORE_REFERENCE_MISSING,
    LocalPrivateImageArtifactStore,
    PrivateImageArtifactBinaryReader,
    inspect_image_binary,
    private_storage_side_effect_counters,
)
from services.runtime.image_artifact_runtime_mapper import (  # noqa: E402
    FAILURE_DIMENSION_MISMATCH,
    FAILURE_HASH_MISMATCH,
    FAILURE_INVALID_METADATA,
    FAILURE_MIME_MISMATCH,
    FAILURE_SAFETY_PENDING_REVIEW,
    map_image_artifact_to_response_reference,
)


ARTIFACT_FIXTURE = REPO_ROOT / "scripts/checks/fixtures/image-generation-artifact-basic.json"


def png_chunk(chunk_type: bytes, data: bytes) -> bytes:
    checksum = zlib.crc32(chunk_type + data) & 0xFFFFFFFF
    return struct.pack(">I", len(data)) + chunk_type + data + struct.pack(">I", checksum)


def make_png(width: int = 2, height: int = 3, *, channel: int = 0) -> bytes:
    signature = b"\x89PNG\r\n\x1a\n"
    ihdr = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    row = bytes([0]) + bytes([channel, 32, 64]) * width
    image_data = zlib.compress(row * height)
    return signature + png_chunk(b"IHDR", ihdr) + png_chunk(b"IDAT", image_data) + png_chunk(b"IEND", b"")


def make_jpeg(width: int = 2, height: int = 3) -> bytes:
    sof = (
        b"\xff\xc0"
        + struct.pack(">H", 11)
        + b"\x08"
        + struct.pack(">HH", height, width)
        + b"\x01\x01\x11\x00"
    )
    scan = b"\xff\xda\x00\x08\x01\x01\x00\x00\x3f\x00\x00"
    return b"\xff\xd8" + sof + scan + b"\xff\xd9"


def make_webp(width: int = 2, height: int = 3) -> bytes:
    packed = (width - 1) | ((height - 1) << 14)
    chunk_data = b"\x2f" + packed.to_bytes(4, "little") + b"\x00\x00\x00\x00"
    chunk = b"VP8L" + struct.pack("<I", len(chunk_data)) + chunk_data + b"\x00"
    payload = b"WEBP" + chunk
    return b"RIFF" + struct.pack("<I", len(payload)) + payload


def artifact_for_payload(
    payload: bytes,
    *,
    artifact_id: str = "image-artifact-private-001",
) -> dict[str, Any]:
    artifact = json.loads(ARTIFACT_FIXTURE.read_text(encoding="utf-8"))
    observation = inspect_image_binary(payload)
    artifact["artifact_id"] = artifact_id
    artifact["artifact"].update(
        {
            "uri": f"artifact://radishmind/generated/{artifact_id}.{observation.format}",
            "mime_type": observation.mime_type,
            "width": observation.width,
            "height": observation.height,
            "format": observation.format,
            "sha256": observation.sha256,
        }
    )
    return artifact


def metadata_reference(artifact: dict[str, Any]) -> dict[str, Any]:
    mapping = map_image_artifact_to_response_reference(artifact)
    if not mapping.ok or mapping.metadata_reference is None:
        raise AssertionError(mapping.failure_message)
    return copy.deepcopy(mapping.metadata_reference)


class ImageArtifactPrivateStorageTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary_directory.name) / "private-artifacts"
        self.store = LocalPrivateImageArtifactStore(self.root)

    def tearDown(self) -> None:
        self.temporary_directory.cleanup()

    def test_png_jpeg_and_webp_observations_are_deterministic(self) -> None:
        cases = (
            ("png", make_png(), "image/png"),
            ("jpg", make_jpeg(), "image/jpeg"),
            ("webp", make_webp(), "image/webp"),
        )
        for expected_format, payload, expected_mime in cases:
            with self.subTest(image_format=expected_format):
                first = inspect_image_binary(payload)
                second = inspect_image_binary(bytes(payload))
                self.assertEqual(first, second)
                self.assertEqual(first.format, expected_format)
                self.assertEqual(first.mime_type, expected_mime)
                self.assertEqual((first.width, first.height), (2, 3))
                self.assertEqual(first.sha256, hashlib.sha256(payload).hexdigest())
                artifact = artifact_for_payload(
                    payload,
                    artifact_id=f"image-artifact-private-{expected_format}",
                )
                self.assertTrue(self.store.put(artifact, payload).ok)

    def test_store_configuration_requires_absolute_private_root(self) -> None:
        with self.assertRaisesRegex(ValueError, "image_artifact_private_store_root_invalid"):
            LocalPrivateImageArtifactStore(Path("relative-artifacts"))
        with self.assertRaisesRegex(ValueError, "image_artifact_private_store_root_invalid"):
            LocalPrivateImageArtifactStore(self.root, max_artifact_bytes=0)

        target = Path(self.temporary_directory.name) / "symlink-target"
        target.mkdir(mode=0o700)
        self.root.symlink_to(target, target_is_directory=True)
        payload = make_png()
        result = self.store.put(artifact_for_payload(payload), payload)
        self.assertEqual(result.failure_code, FAILURE_STORE_INTEGRITY)
        self.assertEqual(list(target.iterdir()), [])

    def test_store_lookup_and_explicit_binary_consume_form_private_round_trip(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)

        stored = self.store.put(artifact, payload)

        self.assertTrue(stored.ok)
        self.assertTrue(stored.created)
        self.assertEqual(stored.local_artifact_store_write_count, 1)
        self.assertEqual(stored.artifact_binary_read_count, 0)
        self.assertIsNotNone(stored.stored_artifact)
        self.assertTrue(stored.stored_artifact.storage_ref.startswith("blobs/sha256/"))
        self.assertNotIn(str(self.root), repr(stored.stored_artifact))
        self.assertFalse(hasattr(stored.stored_artifact, "_blob_path"))

        reference = metadata_reference(artifact)
        lookup = self.store.lookup(reference)
        self.assertTrue(lookup.ok)
        self.assertEqual(lookup.artifact_store_lookup_count, 1)
        self.assertEqual(lookup.artifact_binary_read_count, 0)

        reader = PrivateImageArtifactBinaryReader(self.store)
        denied = reader.consume(reference, lambda stream, observation: None)
        self.assertEqual(denied.failure_code, FAILURE_BINARY_READ_FORBIDDEN)
        self.assertEqual(denied.artifact_store_lookup_count, 0)
        self.assertEqual(denied.artifact_binary_read_count, 0)

        consumed: list[tuple[bytes, str]] = []
        captured_streams = []

        def consume(stream, observation) -> None:
            captured_streams.append(stream)
            consumed.append((stream.read(), observation.sha256))

        read = reader.consume(reference, consume, allow_binary_read=True)
        self.assertTrue(read.ok)
        self.assertEqual(consumed, [(payload, hashlib.sha256(payload).hexdigest())])
        self.assertTrue(captured_streams[0].closed)
        self.assertFalse(hasattr(read, "payload"))
        self.assertFalse(hasattr(read, "file_path"))
        counters = private_storage_side_effect_counters(read)
        self.assertEqual(counters["artifact_store_lookup_count"], 1)
        self.assertEqual(counters["artifact_binary_read_count"], 1)
        self.assertEqual(counters["binary_consumer_call_count"], 1)
        self.assertEqual(counters["backend_call_count"], 0)
        self.assertEqual(counters["artifact_upload_count"], 0)
        self.assertEqual(counters["public_url_resolution_count"], 0)
        self.assertEqual(counters["production_storage_write_count"], 0)

    def test_same_artifact_is_idempotent_and_never_overwritten(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)

        first = self.store.put(artifact, payload)
        second = self.store.put(copy.deepcopy(artifact), bytes(payload))

        self.assertTrue(first.ok)
        self.assertTrue(second.ok)
        self.assertFalse(second.created)
        self.assertEqual(second.local_artifact_store_write_count, 0)
        self.assertEqual(second.artifact_binary_read_count, 1)
        self.assertEqual(first.stored_artifact.storage_ref, second.stored_artifact.storage_ref)

    def test_concurrent_identical_writes_converge_without_partial_files(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)

        with ThreadPoolExecutor(max_workers=8) as executor:
            results = list(executor.map(lambda _: self.store.put(artifact, payload), range(16)))

        self.assertTrue(all(result.ok for result in results))
        self.assertEqual(
            {result.stored_artifact.storage_ref for result in results},
            {results[0].stored_artifact.storage_ref},
        )
        self.assertEqual(len(list((self.root / "refs").glob("*.json"))), 1)
        self.assertEqual(len(list((self.root / "blobs").glob("sha256/*/*"))), 1)
        self.assertEqual(list(self.root.rglob(".artifact-*")), [])

    def test_existing_artifact_id_cannot_be_rebound(self) -> None:
        first_payload = make_png(channel=0)
        second_payload = make_png(channel=255)
        first = artifact_for_payload(first_payload)
        second = artifact_for_payload(second_payload)
        self.assertTrue(self.store.put(first, first_payload).ok)

        conflict = self.store.put(second, second_payload)

        self.assertEqual(conflict.failure_code, FAILURE_STORE_CONFLICT)
        lookup = self.store.lookup(metadata_reference(first))
        self.assertTrue(lookup.ok)

    def test_payload_type_empty_invalid_and_size_budgets_fail_before_write(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        cases = (
            ("string", "base64-payload", FAILURE_BINARY_INVALID),
            ("empty", b"", FAILURE_BINARY_INVALID),
            ("truncated", payload[:24], FAILURE_BINARY_INVALID),
            ("jpeg-premature-eoi", b"\xff\xd8\xff\xd9" + make_jpeg()[2:], FAILURE_BINARY_INVALID),
        )
        for name, candidate, expected_failure in cases:
            with self.subTest(name=name):
                result = self.store.put(artifact, candidate)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertFalse(self.root.exists())

        small_store = LocalPrivateImageArtifactStore(self.root, max_artifact_bytes=64)
        oversized = small_store.put(artifact, payload)
        self.assertEqual(oversized.failure_code, FAILURE_BINARY_TOO_LARGE)
        self.assertFalse(self.root.exists())

    def test_metadata_hash_mime_dimensions_format_and_uri_drift_fail_closed(self) -> None:
        payload = make_png()
        base = artifact_for_payload(payload)
        cases = []
        hash_drift = copy.deepcopy(base)
        hash_drift["artifact"]["sha256"] = "b" * 64
        cases.append(("hash", hash_drift, FAILURE_HASH_MISMATCH))
        mime_drift = copy.deepcopy(base)
        mime_drift["artifact"]["mime_type"] = "image/jpeg"
        cases.append(("mime", mime_drift, FAILURE_MIME_MISMATCH))
        dimension_drift = copy.deepcopy(base)
        dimension_drift["artifact"]["width"] = 9
        cases.append(("dimensions", dimension_drift, FAILURE_DIMENSION_MISMATCH))
        format_drift = copy.deepcopy(base)
        format_drift["artifact"]["format"] = "jpg"
        format_drift["artifact"]["uri"] = (
            f"artifact://radishmind/generated/{format_drift['artifact_id']}.jpg"
        )
        cases.append(("format", format_drift, FAILURE_MIME_MISMATCH))
        uri_drift = copy.deepcopy(base)
        uri_drift["artifact"]["uri"] = "artifact://other/private.png"
        cases.append(("uri", uri_drift, FAILURE_INVALID_METADATA))

        for name, artifact, expected_failure in cases:
            with self.subTest(name=name):
                result = self.store.put(artifact, payload)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertFalse(self.root.exists())

    def test_schema_safety_and_provenance_fail_before_write(self) -> None:
        payload = make_png()
        base = artifact_for_payload(payload)
        cases = []
        unknown = copy.deepcopy(base)
        unknown["public_url"] = "https://public.example/image.png"
        cases.append(("unknown", unknown, FAILURE_INVALID_METADATA))
        pending = copy.deepcopy(base)
        pending["safety"]["review_status"] = "pending_review"
        cases.append(("pending", pending, FAILURE_SAFETY_PENDING_REVIEW))
        missing_provenance = copy.deepcopy(base)
        del missing_provenance["provenance"]["source_request_id"]
        cases.append(("provenance", missing_provenance, FAILURE_INVALID_METADATA))

        for name, artifact, expected_failure in cases:
            with self.subTest(name=name):
                result = self.store.put(artifact, payload)
                self.assertEqual(result.failure_code, expected_failure)
                self.assertFalse(self.root.exists())

    def test_lookup_rejects_missing_public_or_caller_selected_storage_reference(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        reference = metadata_reference(artifact)

        missing = self.store.lookup(reference)
        self.assertEqual(missing.failure_code, FAILURE_STORE_REFERENCE_MISSING)
        self.assertEqual(missing.artifact_store_lookup_count, 1)

        public = copy.deepcopy(reference)
        public["public_url"] = "https://public.example/image.png"
        self.assertEqual(self.store.lookup(public).failure_code, FAILURE_INVALID_METADATA)

        selected_path = copy.deepcopy(reference)
        selected_path["storage_ref"] = "../../outside"
        self.assertEqual(self.store.lookup(selected_path).failure_code, FAILURE_INVALID_METADATA)

        invalid_id = copy.deepcopy(reference)
        invalid_id["artifact_id"] = "../outside"
        self.assertEqual(self.store.lookup(invalid_id).failure_code, FAILURE_INVALID_METADATA)

    def test_reference_record_tampering_fails_integrity_without_binary_read(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        self.assertTrue(self.store.put(artifact, payload).ok)
        reference_path = self.root / "refs" / f"{artifact['artifact_id']}.json"
        record = json.loads(reference_path.read_text(encoding="utf-8"))
        record["width"] = 99
        reference_path.write_text(json.dumps(record), encoding="utf-8")

        result = self.store.lookup(metadata_reference(artifact))

        self.assertEqual(result.failure_code, FAILURE_STORE_INTEGRITY)
        self.assertEqual(result.artifact_store_lookup_count, 1)
        self.assertEqual(result.artifact_binary_read_count, 0)

    def test_blob_symlink_is_rejected_by_lookup(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        stored = self.store.put(artifact, payload)
        blob_path = self.root / stored.stored_artifact.storage_ref
        replacement = self.root / "replacement.bin"
        replacement.write_bytes(payload)
        blob_path.unlink()
        blob_path.symlink_to(replacement)

        result = self.store.lookup(metadata_reference(artifact))

        self.assertEqual(result.failure_code, FAILURE_STORE_INTEGRITY)

    def test_binary_reader_detects_valid_container_tampering_before_consumer(self) -> None:
        original = make_png(channel=1)
        replacement = make_png(channel=2)
        self.assertEqual(len(original), len(replacement))
        artifact = artifact_for_payload(original)
        stored = self.store.put(artifact, original)
        blob_path = self.root / stored.stored_artifact.storage_ref
        blob_path.write_bytes(replacement)
        consumer_calls: list[bool] = []

        result = PrivateImageArtifactBinaryReader(self.store).consume(
            metadata_reference(artifact),
            lambda stream, observation: consumer_calls.append(True),
            allow_binary_read=True,
        )

        self.assertEqual(result.failure_code, FAILURE_HASH_MISMATCH)
        self.assertEqual(result.artifact_store_lookup_count, 1)
        self.assertEqual(result.artifact_binary_read_count, 1)
        self.assertEqual(result.binary_consumer_call_count, 0)
        self.assertEqual(consumer_calls, [])

    def test_binary_reader_sanitizes_consumer_failures_and_never_retries(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        self.assertTrue(self.store.put(artifact, payload).ok)
        calls: list[int] = []

        def failing_consumer(stream, observation) -> None:
            calls.append(1)
            raise RuntimeError("consumer-private-path-secret")

        result = PrivateImageArtifactBinaryReader(self.store).consume(
            metadata_reference(artifact),
            failing_consumer,
            allow_binary_read=True,
        )

        self.assertEqual(result.failure_code, FAILURE_BINARY_CONSUMER_FAILED)
        self.assertNotIn("consumer-private-path-secret", result.failure_message)
        self.assertEqual(calls, [1])
        counters = private_storage_side_effect_counters(result)
        self.assertEqual(counters["binary_consumer_call_count"], 1)
        self.assertEqual(counters["retry_count"], 0)
        self.assertEqual(counters["fallback_count"], 0)

    def test_malformed_binary_after_lookup_never_reaches_consumer(self) -> None:
        payload = make_png()
        artifact = artifact_for_payload(payload)
        stored = self.store.put(artifact, payload)
        blob_path = self.root / stored.stored_artifact.storage_ref
        blob_path.write_bytes(b"x" * len(payload))
        calls: list[int] = []

        result = PrivateImageArtifactBinaryReader(self.store).consume(
            metadata_reference(artifact),
            lambda stream, observation: calls.append(1),
            allow_binary_read=True,
        )

        self.assertEqual(result.failure_code, FAILURE_BINARY_READER_UNAVAILABLE)
        self.assertEqual(calls, [])


if __name__ == "__main__":
    unittest.main()
