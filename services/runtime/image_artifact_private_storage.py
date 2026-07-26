from __future__ import annotations

import hashlib
import io
import json
import os
import re
import stat
import tempfile
from collections.abc import Callable, Mapping
from dataclasses import dataclass
from functools import lru_cache
from pathlib import Path
from typing import Any, BinaryIO

import jsonschema

from services.runtime.image_artifact_binary_inspection import (
    FAILURE_BINARY_INVALID,
    ImageBinaryObservation,
    inspect_image_binary,
)
from services.runtime.image_artifact_runtime_mapper import (
    FAILURE_DIMENSION_MISMATCH,
    FAILURE_HASH_MISMATCH,
    FAILURE_INVALID_METADATA,
    FAILURE_MIME_MISMATCH,
    ImageArtifactMappingResult,
    map_image_artifact_to_response_reference,
)


REPO_ROOT = Path(__file__).resolve().parents[2]
ARTIFACT_SCHEMA_PATH = REPO_ROOT / "contracts/image-generation-artifact.schema.json"

FAILURE_STORE_ROOT_INVALID = "image_artifact_private_store_root_invalid"
FAILURE_BINARY_TOO_LARGE = "image_artifact_binary_too_large"
FAILURE_STORE_UNAVAILABLE = "image_artifact_store_unavailable"
FAILURE_STORE_CONFLICT = "image_artifact_store_conflict"
FAILURE_STORE_INTEGRITY = "image_artifact_store_integrity_failure"
FAILURE_STORE_REFERENCE_MISSING = "image_artifact_store_reference_missing"
FAILURE_BINARY_READ_FORBIDDEN = "image_artifact_binary_read_forbidden"
FAILURE_BINARY_READER_UNAVAILABLE = "image_artifact_binary_reader_unavailable"
FAILURE_BINARY_CONSUMER_FAILED = "image_artifact_binary_consumer_failed"

MAX_ARTIFACT_BYTES = 32 * 1024 * 1024
MAX_REFERENCE_RECORD_BYTES = 8 * 1024
MAX_DIMENSION = 2_048
MAX_PIXELS = 4_194_304

IDENTIFIER_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$")
SHA256_RE = re.compile(r"^[a-f0-9]{64}$")
MIME_BY_FORMAT = {
    "png": "image/png",
    "jpg": "image/jpeg",
    "webp": "image/webp",
}
FORMAT_BY_MIME = {mime_type: image_format for image_format, mime_type in MIME_BY_FORMAT.items()}
ZERO_EXTERNAL_SIDE_EFFECTS = {
    "backend_call_count": 0,
    "image_generation_count": 0,
    "model_download_count": 0,
    "artifact_upload_count": 0,
    "production_storage_write_count": 0,
    "public_url_resolution_count": 0,
    "retry_count": 0,
    "fallback_count": 0,
    "executor_call_count": 0,
    "confirmation_call_count": 0,
    "business_writeback_count": 0,
    "replay_call_count": 0,
}
REFERENCE_RECORD_FIELDS = {
    "schema_version",
    "kind",
    "artifact_id",
    "artifact_uri",
    "storage_ref",
    "sha256",
    "mime_type",
    "width",
    "height",
    "format",
    "size_bytes",
}
FORBIDDEN_REFERENCE_KEYS = {
    "absolute_path",
    "base64_image",
    "binary_payload",
    "file_path",
    "image_bytes",
    "pixel_payload",
    "provider_raw_dump",
    "provider_raw_response",
    "public_url",
    "raw_provider_response",
    "signed_public_url",
    "signed_url",
    "storage_ref",
}


@dataclass(frozen=True)
class StoredImageArtifact:
    artifact_id: str
    artifact_uri: str
    storage_ref: str
    sha256: str
    mime_type: str
    width: int
    height: int
    format: str
    size_bytes: int


@dataclass(frozen=True)
class ImageArtifactStoreResult:
    ok: bool
    stored_artifact: StoredImageArtifact | None = None
    created: bool = False
    failure_code: str | None = None
    failure_message: str = ""
    artifact_store_lookup_count: int = 0
    local_artifact_store_write_count: int = 0
    artifact_binary_revalidation_count: int = 0
    artifact_binary_read_count: int = 0


@dataclass(frozen=True)
class ImageArtifactBinaryReadResult:
    ok: bool
    observation: ImageBinaryObservation | None = None
    failure_code: str | None = None
    failure_message: str = ""
    artifact_store_lookup_count: int = 0
    artifact_binary_read_count: int = 0
    binary_consumer_call_count: int = 0


class LocalPrivateImageArtifactStore:
    def __init__(self, root: str | Path, *, max_artifact_bytes: int = MAX_ARTIFACT_BYTES) -> None:
        candidate = Path(root)
        if not candidate.is_absolute():
            raise ValueError(FAILURE_STORE_ROOT_INVALID)
        if (
            not isinstance(max_artifact_bytes, int)
            or isinstance(max_artifact_bytes, bool)
            or max_artifact_bytes <= 0
            or max_artifact_bytes > MAX_ARTIFACT_BYTES
        ):
            raise ValueError(FAILURE_STORE_ROOT_INVALID)
        self._root = candidate
        self._max_artifact_bytes = max_artifact_bytes

    @property
    def max_artifact_bytes(self) -> int:
        return self._max_artifact_bytes

    def put(
        self,
        artifact_document: Mapping[str, Any],
        payload: bytes | bytearray | memoryview,
    ) -> ImageArtifactStoreResult:
        payload_bytes = normalize_payload(payload)
        if payload_bytes is None:
            return store_failure(FAILURE_BINARY_INVALID, "artifact payload must be bytes-like")
        if not payload_bytes:
            return store_failure(FAILURE_BINARY_INVALID, "artifact payload must not be empty")
        if len(payload_bytes) > self._max_artifact_bytes:
            return store_failure(FAILURE_BINARY_TOO_LARGE, "artifact payload exceeds the private store budget")

        try:
            observation = inspect_image_binary(payload_bytes)
        except ValueError:
            return store_failure(
                FAILURE_BINARY_INVALID,
                "artifact payload is not a supported image container",
                revalidation_count=1,
            )
        if observation.width > MAX_DIMENSION or observation.height > MAX_DIMENSION:
            return store_failure(
                FAILURE_BINARY_INVALID,
                "artifact dimensions exceed the private store budget",
                revalidation_count=1,
            )
        if observation.width * observation.height > MAX_PIXELS:
            return store_failure(
                FAILURE_BINARY_INVALID,
                "artifact pixel count exceeds the private store budget",
                revalidation_count=1,
            )

        validation = validate_artifact_for_storage(artifact_document, observation)
        if not validation.ok:
            return store_failure(
                validation.failure_code or FAILURE_INVALID_METADATA,
                validation.failure_message or "artifact metadata failed private store validation",
                revalidation_count=1,
            )
        artifact = dict(artifact_document)
        artifact_fields = dict(artifact["artifact"])
        record = build_reference_record(artifact, artifact_fields, observation)

        blob_created = False
        try:
            self._prepare_root()
            blob_path = self._path_for_storage_ref(record["storage_ref"])
            reference_path = self._reference_path(record["artifact_id"])
            self._ensure_private_parent_chain(blob_path)
            self._validate_private_parent_chain(reference_path)
            existing_record = self._read_reference_record(reference_path, missing_allowed=True)
            if existing_record is not None and existing_record != record:
                return store_failure(
                    FAILURE_STORE_CONFLICT,
                    "artifact id is already bound to different private storage metadata",
                    revalidation_count=1,
                )

            blob_created = write_immutable_file(
                blob_path,
                payload_bytes,
                expected_sha256=observation.sha256,
            )
            reference_created = write_immutable_file(
                reference_path,
                canonical_json_bytes(record),
            )
            created = blob_created or reference_created
            stored = stored_artifact_from_record(record)
            return ImageArtifactStoreResult(
                ok=True,
                stored_artifact=stored,
                created=created,
                local_artifact_store_write_count=1 if created else 0,
                artifact_binary_revalidation_count=1,
                artifact_binary_read_count=0 if blob_created else 1,
            )
        except FileExistsError:
            return store_failure(
                FAILURE_STORE_CONFLICT,
                "private artifact storage is immutable",
                write_count=1 if blob_created else 0,
                revalidation_count=1,
            )
        except StoreIntegrityError:
            return store_failure(
                FAILURE_STORE_INTEGRITY,
                "existing private artifact storage failed integrity validation",
                write_count=1 if blob_created else 0,
                revalidation_count=1,
            )
        except (OSError, ValueError):
            return store_failure(
                FAILURE_STORE_UNAVAILABLE,
                "private artifact store is unavailable",
                write_count=1 if blob_created else 0,
                revalidation_count=1,
            )

    def lookup(self, metadata_reference: Mapping[str, Any]) -> ImageArtifactStoreResult:
        expected_record = reference_record_from_metadata_reference(metadata_reference)
        if expected_record is None:
            return store_failure(FAILURE_INVALID_METADATA, "artifact metadata reference is invalid")

        try:
            if not self._root.exists():
                return store_failure(
                    FAILURE_STORE_REFERENCE_MISSING,
                    "private artifact reference does not exist",
                    lookup_count=1,
                )
            self._assert_private_directory(self._root)
            reference_path = self._reference_path(expected_record["artifact_id"])
            self._validate_private_parent_chain(reference_path)
            stored_record = self._read_reference_record(reference_path, missing_allowed=False)
            if any(stored_record.get(key) != value for key, value in expected_record.items()):
                return store_failure(
                    FAILURE_STORE_INTEGRITY,
                    "private artifact reference does not match requested metadata",
                    lookup_count=1,
                )
            blob_path = self._path_for_storage_ref(stored_record["storage_ref"])
            self._validate_private_parent_chain(blob_path)
            file_stat = os.lstat(blob_path)
            if stat.S_ISLNK(file_stat.st_mode) or not stat.S_ISREG(file_stat.st_mode):
                raise StoreIntegrityError
            if file_stat.st_size != stored_record["size_bytes"] or file_stat.st_size > self._max_artifact_bytes:
                raise StoreIntegrityError
            return ImageArtifactStoreResult(
                ok=True,
                stored_artifact=stored_artifact_from_record(stored_record),
                artifact_store_lookup_count=1,
            )
        except FileNotFoundError:
            return store_failure(
                FAILURE_STORE_REFERENCE_MISSING,
                "private artifact reference does not exist",
                lookup_count=1,
            )
        except StoreIntegrityError:
            return store_failure(
                FAILURE_STORE_INTEGRITY,
                "private artifact storage failed integrity validation",
                lookup_count=1,
            )
        except (OSError, ValueError):
            return store_failure(
                FAILURE_STORE_UNAVAILABLE,
                "private artifact store is unavailable",
                lookup_count=1,
            )

    def _prepare_root(self) -> None:
        self._root.mkdir(mode=0o700, parents=True, exist_ok=True)
        self._assert_private_directory(self._root)
        for child in (self._root / "blobs", self._root / "refs"):
            child.mkdir(mode=0o700, exist_ok=True)
            self._assert_private_directory(child)

    def _assert_private_directory(self, path: Path) -> None:
        path_stat = os.lstat(path)
        if stat.S_ISLNK(path_stat.st_mode) or not stat.S_ISDIR(path_stat.st_mode):
            raise StoreIntegrityError
        if os.name != "nt" and path_stat.st_mode & 0o077:
            raise StoreIntegrityError

    def _ensure_private_parent_chain(self, path: Path) -> None:
        relative = path.relative_to(self._root)
        current = self._root
        for part in relative.parts[:-1]:
            current = current / part
            current.mkdir(mode=0o700, exist_ok=True)
            self._assert_private_directory(current)

    def _validate_private_parent_chain(self, path: Path) -> None:
        relative = path.relative_to(self._root)
        current = self._root
        for part in relative.parts[:-1]:
            current = current / part
            self._assert_private_directory(current)

    def _path_for_storage_ref(self, storage_ref: str) -> Path:
        parts = storage_ref.split("/")
        if len(parts) != 4 or parts[0] != "blobs" or parts[1] != "sha256":
            raise ValueError(FAILURE_STORE_INTEGRITY)
        digest = parts[3].split(".", 1)[0]
        if parts[2] != digest[:2] or not SHA256_RE.fullmatch(digest):
            raise ValueError(FAILURE_STORE_INTEGRITY)
        image_format = parts[3].removeprefix(f"{digest}.")
        if image_format not in MIME_BY_FORMAT:
            raise ValueError(FAILURE_STORE_INTEGRITY)
        return self._root / parts[0] / parts[1] / parts[2] / parts[3]

    def _reference_path(self, artifact_id: str) -> Path:
        if not IDENTIFIER_RE.fullmatch(artifact_id):
            raise ValueError(FAILURE_INVALID_METADATA)
        return self._root / "refs" / f"{artifact_id}.json"

    def _read_reference_record(self, path: Path, *, missing_allowed: bool) -> dict[str, Any] | None:
        try:
            record_bytes = read_regular_file(path, MAX_REFERENCE_RECORD_BYTES)
        except FileNotFoundError:
            if missing_allowed:
                return None
            raise
        try:
            record = json.loads(record_bytes)
        except (UnicodeDecodeError, json.JSONDecodeError) as exc:
            raise StoreIntegrityError from exc
        if not reference_record_is_valid(record):
            raise StoreIntegrityError
        return record

    def _read_stored_payload(self, stored: StoredImageArtifact) -> bytes:
        blob_path = self._path_for_storage_ref(stored.storage_ref)
        self._validate_private_parent_chain(blob_path)
        return read_regular_file(blob_path, self._max_artifact_bytes)


class PrivateImageArtifactBinaryReader:
    def __init__(self, store: LocalPrivateImageArtifactStore) -> None:
        self._store = store

    def consume(
        self,
        metadata_reference: Mapping[str, Any],
        consumer: Callable[[BinaryIO, ImageBinaryObservation], None],
        *,
        allow_binary_read: bool = False,
    ) -> ImageArtifactBinaryReadResult:
        if allow_binary_read is not True:
            return read_failure(
                FAILURE_BINARY_READ_FORBIDDEN,
                "private artifact binary read requires explicit authorization",
            )
        if not callable(consumer):
            return read_failure(
                FAILURE_BINARY_CONSUMER_FAILED,
                "private artifact binary consumer must be callable",
            )

        lookup = self._store.lookup(metadata_reference)
        if not lookup.ok or lookup.stored_artifact is None:
            return read_failure(
                lookup.failure_code or FAILURE_STORE_UNAVAILABLE,
                lookup.failure_message or "private artifact lookup failed",
                lookup_count=lookup.artifact_store_lookup_count,
            )
        stored = lookup.stored_artifact

        try:
            payload = self._store._read_stored_payload(stored)
            observation = inspect_image_binary(payload)
        except (OSError, ValueError, StoreIntegrityError):
            return read_failure(
                FAILURE_BINARY_READER_UNAVAILABLE,
                "private artifact binary could not be read or inspected",
                lookup_count=1,
                binary_read_count=1,
            )

        mismatch = observation_mismatch(stored, observation)
        if mismatch is not None:
            return read_failure(
                mismatch,
                "private artifact binary does not match stored metadata",
                lookup_count=1,
                binary_read_count=1,
            )

        stream = io.BytesIO(payload)
        try:
            consumer(stream, observation)
        except Exception:
            return read_failure(
                FAILURE_BINARY_CONSUMER_FAILED,
                "private artifact binary consumer failed",
                lookup_count=1,
                binary_read_count=1,
                consumer_call_count=1,
            )
        finally:
            stream.close()
        return ImageArtifactBinaryReadResult(
            ok=True,
            observation=observation,
            artifact_store_lookup_count=1,
            artifact_binary_read_count=1,
            binary_consumer_call_count=1,
        )


class StoreIntegrityError(Exception):
    pass


def private_storage_side_effect_counters(
    result: ImageArtifactStoreResult | ImageArtifactBinaryReadResult,
) -> dict[str, int]:
    return {
        "artifact_store_lookup_count": result.artifact_store_lookup_count,
        "local_artifact_store_write_count": (
            result.local_artifact_store_write_count
            if isinstance(result, ImageArtifactStoreResult)
            else 0
        ),
        "artifact_binary_revalidation_count": (
            result.artifact_binary_revalidation_count
            if isinstance(result, ImageArtifactStoreResult)
            else 0
        ),
        "artifact_binary_read_count": result.artifact_binary_read_count,
        "binary_consumer_call_count": (
            result.binary_consumer_call_count
            if isinstance(result, ImageArtifactBinaryReadResult)
            else 0
        ),
        **ZERO_EXTERNAL_SIDE_EFFECTS,
    }


def validate_artifact_for_storage(
    artifact_document: Mapping[str, Any],
    observation: ImageBinaryObservation,
) -> ImageArtifactMappingResult:
    if not isinstance(artifact_document, Mapping):
        return mapping_failure(FAILURE_INVALID_METADATA, "artifact document must be a mapping")
    schema_error = next(iter(artifact_schema_validator().iter_errors(artifact_document)), None)
    if schema_error is not None:
        return mapping_failure(FAILURE_INVALID_METADATA, "artifact document failed the strict schema")
    mapping = map_image_artifact_to_response_reference(
        artifact_document,
        expected_sha256=observation.sha256,
        expected_mime_type=observation.mime_type,
        expected_width=observation.width,
        expected_height=observation.height,
    )
    if not mapping.ok:
        return mapping
    artifact = dict(artifact_document)
    artifact_fields = dict(artifact["artifact"])
    safety = dict(artifact["safety"])
    expected_uri = (
        f"artifact://radishmind/generated/{artifact['artifact_id']}."
        f"{artifact_fields['format']}"
    )
    if artifact_fields["uri"] != expected_uri:
        return mapping_failure(FAILURE_INVALID_METADATA, "artifact URI is not canonical")
    if artifact_fields["format"] != observation.format:
        return mapping_failure(FAILURE_MIME_MISMATCH, "artifact format does not match binary")
    if (
        safety["risk_level"] != "low"
        or safety["requires_confirmation"] is not False
        or safety["review_status"] not in {"not_required", "reviewed_pass"}
    ):
        return mapping_failure(
            "image_artifact_safety_review_not_passed",
            "artifact safety state does not allow private storage",
        )
    return mapping


def reference_record_from_metadata_reference(
    metadata_reference: Mapping[str, Any],
) -> dict[str, Any] | None:
    if not isinstance(metadata_reference, Mapping):
        return None
    if reference_has_forbidden_material(metadata_reference):
        return None
    artifact_id = metadata_reference.get("artifact_id")
    artifact_uri = metadata_reference.get("uri")
    digest = metadata_reference.get("sha256")
    mime_type = metadata_reference.get("mime_type")
    image_format = metadata_reference.get("format")
    dimensions = metadata_reference.get("dimensions")
    safety = metadata_reference.get("safety")
    provenance = metadata_reference.get("provenance")
    if (
        not isinstance(artifact_id, str)
        or not IDENTIFIER_RE.fullmatch(artifact_id)
        or not isinstance(digest, str)
        or not SHA256_RE.fullmatch(digest)
        or mime_type not in FORMAT_BY_MIME
        or image_format != FORMAT_BY_MIME[mime_type]
        or not isinstance(dimensions, Mapping)
        or not isinstance(safety, Mapping)
        or not isinstance(provenance, Mapping)
    ):
        return None
    width = positive_integer(dimensions.get("width"))
    height = positive_integer(dimensions.get("height"))
    if width is None or height is None or width > MAX_DIMENSION or height > MAX_DIMENSION:
        return None
    if width * height > MAX_PIXELS:
        return None
    expected_uri = f"artifact://radishmind/generated/{artifact_id}.{image_format}"
    if artifact_uri != expected_uri:
        return None
    if (
        safety.get("risk_level") != "low"
        or safety.get("requires_confirmation") is not False
        or safety.get("review_status") not in {"not_required", "reviewed_pass"}
    ):
        return None
    trace_ids = provenance.get("trace_ids")
    if (
        not isinstance(provenance.get("source_request_id"), str)
        or not provenance["source_request_id"]
        or not isinstance(trace_ids, list)
        or not trace_ids
        or any(not isinstance(item, str) or not item for item in trace_ids)
    ):
        return None
    expected_record = {
        "schema_version": 1,
        "kind": "image_artifact_private_store_record",
        "artifact_id": artifact_id,
        "artifact_uri": artifact_uri,
        "storage_ref": storage_ref_for(digest, image_format),
        "sha256": digest,
        "mime_type": mime_type,
        "width": width,
        "height": height,
        "format": image_format,
    }
    if "size_bytes" in metadata_reference:
        size_bytes = positive_integer(metadata_reference.get("size_bytes"))
        if size_bytes is None or size_bytes > MAX_ARTIFACT_BYTES:
            return None
        expected_record["size_bytes"] = size_bytes
    return expected_record


def build_reference_record(
    artifact_document: Mapping[str, Any],
    artifact_fields: Mapping[str, Any],
    observation: ImageBinaryObservation,
) -> dict[str, Any]:
    return {
        "schema_version": 1,
        "kind": "image_artifact_private_store_record",
        "artifact_id": artifact_document["artifact_id"],
        "artifact_uri": artifact_fields["uri"],
        "storage_ref": storage_ref_for(observation.sha256, observation.format),
        "sha256": observation.sha256,
        "mime_type": observation.mime_type,
        "width": observation.width,
        "height": observation.height,
        "format": observation.format,
        "size_bytes": observation.size_bytes,
    }


def stored_artifact_from_record(record: Mapping[str, Any]) -> StoredImageArtifact:
    return StoredImageArtifact(
        artifact_id=str(record["artifact_id"]),
        artifact_uri=str(record["artifact_uri"]),
        storage_ref=str(record["storage_ref"]),
        sha256=str(record["sha256"]),
        mime_type=str(record["mime_type"]),
        width=int(record["width"]),
        height=int(record["height"]),
        format=str(record["format"]),
        size_bytes=int(record["size_bytes"]),
    )


def reference_record_is_valid(record: Any) -> bool:
    if not isinstance(record, dict) or set(record) != REFERENCE_RECORD_FIELDS:
        return False
    if record.get("schema_version") != 1 or record.get("kind") != "image_artifact_private_store_record":
        return False
    artifact_id = record.get("artifact_id")
    digest = record.get("sha256")
    image_format = record.get("format")
    mime_type = record.get("mime_type")
    width = positive_integer(record.get("width"))
    height = positive_integer(record.get("height"))
    size_bytes = positive_integer(record.get("size_bytes"))
    if (
        not isinstance(artifact_id, str)
        or not IDENTIFIER_RE.fullmatch(artifact_id)
        or not isinstance(digest, str)
        or not SHA256_RE.fullmatch(digest)
        or image_format not in MIME_BY_FORMAT
        or mime_type != MIME_BY_FORMAT[image_format]
        or width is None
        or height is None
        or width > MAX_DIMENSION
        or height > MAX_DIMENSION
        or width * height > MAX_PIXELS
        or size_bytes is None
        or size_bytes > MAX_ARTIFACT_BYTES
    ):
        return False
    return (
        record.get("artifact_uri")
        == f"artifact://radishmind/generated/{artifact_id}.{image_format}"
        and record.get("storage_ref") == storage_ref_for(digest, image_format)
    )


def observation_mismatch(
    stored: StoredImageArtifact,
    observation: ImageBinaryObservation,
) -> str | None:
    if observation.sha256 != stored.sha256:
        return FAILURE_HASH_MISMATCH
    if observation.mime_type != stored.mime_type or observation.format != stored.format:
        return FAILURE_MIME_MISMATCH
    if observation.width != stored.width or observation.height != stored.height:
        return FAILURE_DIMENSION_MISMATCH
    if observation.size_bytes != stored.size_bytes:
        return FAILURE_STORE_INTEGRITY
    return None


def write_immutable_file(
    path: Path,
    content: bytes,
    *,
    expected_sha256: str | None = None,
) -> bool:
    parent_stat = os.lstat(path.parent)
    if stat.S_ISLNK(parent_stat.st_mode) or not stat.S_ISDIR(parent_stat.st_mode):
        raise StoreIntegrityError
    if os.name != "nt" and parent_stat.st_mode & 0o077:
        raise StoreIntegrityError
    if path.is_symlink():
        raise StoreIntegrityError
    if path.exists():
        existing = read_regular_file(path, max(len(content), MAX_REFERENCE_RECORD_BYTES))
        if len(existing) != len(content):
            raise StoreIntegrityError
        if expected_sha256 is not None:
            if hashlib.sha256(existing).hexdigest() != expected_sha256:
                raise StoreIntegrityError
        elif existing != content:
            raise StoreIntegrityError
        return False

    descriptor, temporary_name = tempfile.mkstemp(prefix=".artifact-", dir=path.parent)
    temporary_path = Path(temporary_name)
    try:
        os.fchmod(descriptor, 0o600)
        offset = 0
        while offset < len(content):
            written = os.write(descriptor, content[offset:])
            if written <= 0:
                raise OSError("private artifact write made no progress")
            offset += written
        os.fsync(descriptor)
        os.close(descriptor)
        descriptor = -1
        try:
            os.link(temporary_path, path)
        except FileExistsError:
            existing = read_regular_file(path, max(len(content), MAX_REFERENCE_RECORD_BYTES))
            if existing != content:
                raise StoreIntegrityError
            return False
        return True
    finally:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            temporary_path.unlink()
        except FileNotFoundError:
            pass


def read_regular_file(path: Path, max_bytes: int) -> bytes:
    flags = os.O_RDONLY
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    descriptor = os.open(path, flags)
    try:
        file_stat = os.fstat(descriptor)
        if not stat.S_ISREG(file_stat.st_mode) or file_stat.st_size <= 0 or file_stat.st_size > max_bytes:
            raise StoreIntegrityError
        if os.name != "nt" and file_stat.st_mode & 0o077:
            raise StoreIntegrityError
        chunks: list[bytes] = []
        remaining = file_stat.st_size
        while remaining:
            chunk = os.read(descriptor, min(64 * 1024, remaining))
            if not chunk:
                raise StoreIntegrityError
            chunks.append(chunk)
            remaining -= len(chunk)
        if os.read(descriptor, 1):
            raise StoreIntegrityError
        return b"".join(chunks)
    finally:
        os.close(descriptor)


def storage_ref_for(digest: str, image_format: str) -> str:
    return f"blobs/sha256/{digest[:2]}/{digest}.{image_format}"


def canonical_json_bytes(value: Mapping[str, Any]) -> bytes:
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode("utf-8")


def normalize_payload(payload: Any) -> bytes | None:
    if isinstance(payload, bytes):
        return payload
    if isinstance(payload, (bytearray, memoryview)):
        return bytes(payload)
    return None


def reference_has_forbidden_material(value: Any) -> bool:
    if isinstance(value, Mapping):
        if any(str(key) in FORBIDDEN_REFERENCE_KEYS for key in value):
            return True
        return any(reference_has_forbidden_material(item) for item in value.values())
    if isinstance(value, list):
        return any(reference_has_forbidden_material(item) for item in value)
    if isinstance(value, str):
        return value.startswith(("http://", "https://"))
    return False


def positive_integer(value: Any) -> int | None:
    if isinstance(value, int) and not isinstance(value, bool) and value > 0:
        return value
    return None


@lru_cache(maxsize=1)
def artifact_schema_validator() -> jsonschema.Draft202012Validator:
    schema = json.loads(ARTIFACT_SCHEMA_PATH.read_text(encoding="utf-8"))
    jsonschema.Draft202012Validator.check_schema(schema)
    return jsonschema.Draft202012Validator(schema, format_checker=jsonschema.FormatChecker())


def mapping_failure(code: str, message: str) -> ImageArtifactMappingResult:
    return ImageArtifactMappingResult(ok=False, failure_code=code, failure_message=message)


def store_failure(
    code: str,
    message: str,
    *,
    lookup_count: int = 0,
    write_count: int = 0,
    revalidation_count: int = 0,
) -> ImageArtifactStoreResult:
    return ImageArtifactStoreResult(
        ok=False,
        failure_code=code,
        failure_message=message,
        artifact_store_lookup_count=lookup_count,
        local_artifact_store_write_count=write_count,
        artifact_binary_revalidation_count=revalidation_count,
    )


def read_failure(
    code: str,
    message: str,
    *,
    lookup_count: int = 0,
    binary_read_count: int = 0,
    consumer_call_count: int = 0,
) -> ImageArtifactBinaryReadResult:
    return ImageArtifactBinaryReadResult(
        ok=False,
        failure_code=code,
        failure_message=message,
        artifact_store_lookup_count=lookup_count,
        artifact_binary_read_count=binary_read_count,
        binary_consumer_call_count=consumer_call_count,
    )
