from __future__ import annotations

import hashlib
import struct
import zlib
from dataclasses import dataclass


FAILURE_BINARY_INVALID = "image_artifact_binary_invalid"

PNG_SIGNATURE = b"\x89PNG\r\n\x1a\n"
JPEG_START = b"\xff\xd8"
JPEG_END = b"\xff\xd9"
JPEG_SOF_MARKERS = {
    0xC0,
    0xC1,
    0xC2,
    0xC3,
    0xC5,
    0xC6,
    0xC7,
    0xC9,
    0xCA,
    0xCB,
    0xCD,
    0xCE,
    0xCF,
}


@dataclass(frozen=True)
class ImageBinaryObservation:
    sha256: str
    mime_type: str
    width: int
    height: int
    format: str
    size_bytes: int


def inspect_image_binary(payload: bytes) -> ImageBinaryObservation:
    if payload.startswith(PNG_SIGNATURE):
        image_format, mime_type, width, height = inspect_png(payload)
    elif payload.startswith(JPEG_START):
        image_format, mime_type, width, height = inspect_jpeg(payload)
    elif payload.startswith(b"RIFF") and payload[8:12] == b"WEBP":
        image_format, mime_type, width, height = inspect_webp(payload)
    else:
        raise ValueError(FAILURE_BINARY_INVALID)
    return ImageBinaryObservation(
        sha256=hashlib.sha256(payload).hexdigest(),
        mime_type=mime_type,
        width=width,
        height=height,
        format=image_format,
        size_bytes=len(payload),
    )


def inspect_png(payload: bytes) -> tuple[str, str, int, int]:
    offset = len(PNG_SIGNATURE)
    width = 0
    height = 0
    chunk_index = 0
    saw_image_data = False
    while offset + 12 <= len(payload):
        chunk_length = struct.unpack(">I", payload[offset : offset + 4])[0]
        chunk_type = payload[offset + 4 : offset + 8]
        data_start = offset + 8
        data_end = data_start + chunk_length
        crc_end = data_end + 4
        if crc_end > len(payload):
            raise ValueError(FAILURE_BINARY_INVALID)
        expected_crc = struct.unpack(">I", payload[data_end:crc_end])[0]
        if zlib.crc32(chunk_type + payload[data_start:data_end]) & 0xFFFFFFFF != expected_crc:
            raise ValueError(FAILURE_BINARY_INVALID)
        if chunk_index == 0:
            if chunk_type != b"IHDR" or chunk_length != 13:
                raise ValueError(FAILURE_BINARY_INVALID)
            width, height = struct.unpack(">II", payload[data_start : data_start + 8])
        elif chunk_type == b"IHDR":
            raise ValueError(FAILURE_BINARY_INVALID)
        if chunk_type == b"IDAT":
            saw_image_data = True
        if chunk_type == b"IEND":
            if chunk_length != 0 or crc_end != len(payload) or not saw_image_data:
                raise ValueError(FAILURE_BINARY_INVALID)
            require_positive_dimensions(width, height)
            return "png", "image/png", width, height
        offset = crc_end
        chunk_index += 1
    raise ValueError(FAILURE_BINARY_INVALID)


def inspect_jpeg(payload: bytes) -> tuple[str, str, int, int]:
    if len(payload) < 12 or not payload.endswith(JPEG_END):
        raise ValueError(FAILURE_BINARY_INVALID)
    offset = 2
    width = 0
    height = 0
    saw_scan = False
    while offset < len(payload) - 2:
        if payload[offset] != 0xFF:
            raise ValueError(FAILURE_BINARY_INVALID)
        while offset < len(payload) and payload[offset] == 0xFF:
            offset += 1
        if offset >= len(payload):
            raise ValueError(FAILURE_BINARY_INVALID)
        marker = payload[offset]
        offset += 1
        if marker == 0xDA:
            if offset + 2 > len(payload) - 2:
                raise ValueError(FAILURE_BINARY_INVALID)
            segment_length = struct.unpack(">H", payload[offset : offset + 2])[0]
            if segment_length < 6 or offset + segment_length > len(payload) - 2:
                raise ValueError(FAILURE_BINARY_INVALID)
            saw_scan = True
            break
        if marker == 0x01:
            continue
        if 0xD0 <= marker <= 0xD9:
            raise ValueError(FAILURE_BINARY_INVALID)
        if offset + 2 > len(payload):
            raise ValueError(FAILURE_BINARY_INVALID)
        segment_length = struct.unpack(">H", payload[offset : offset + 2])[0]
        if segment_length < 2 or offset + segment_length > len(payload):
            raise ValueError(FAILURE_BINARY_INVALID)
        if marker in JPEG_SOF_MARKERS:
            if segment_length < 8:
                raise ValueError(FAILURE_BINARY_INVALID)
            height = struct.unpack(">H", payload[offset + 3 : offset + 5])[0]
            width = struct.unpack(">H", payload[offset + 5 : offset + 7])[0]
        offset += segment_length
    if not saw_scan:
        raise ValueError(FAILURE_BINARY_INVALID)
    require_positive_dimensions(width, height)
    return "jpg", "image/jpeg", width, height


def inspect_webp(payload: bytes) -> tuple[str, str, int, int]:
    if len(payload) < 30:
        raise ValueError(FAILURE_BINARY_INVALID)
    declared_size = struct.unpack("<I", payload[4:8])[0] + 8
    if declared_size != len(payload):
        raise ValueError(FAILURE_BINARY_INVALID)
    offset = 12
    canvas_dimensions: tuple[int, int] | None = None
    image_dimensions: tuple[int, int] | None = None
    chunk_index = 0
    while offset + 8 <= len(payload):
        chunk_type = payload[offset : offset + 4]
        chunk_size = struct.unpack("<I", payload[offset + 4 : offset + 8])[0]
        data_start = offset + 8
        data_end = data_start + chunk_size
        padded_end = data_end + (chunk_size % 2)
        if padded_end > len(payload):
            raise ValueError(FAILURE_BINARY_INVALID)
        chunk = payload[data_start:data_end]
        if chunk_type == b"VP8X":
            if len(chunk) != 10 or chunk_index != 0 or canvas_dimensions is not None:
                raise ValueError(FAILURE_BINARY_INVALID)
            canvas_dimensions = (
                1 + int.from_bytes(chunk[4:7], "little"),
                1 + int.from_bytes(chunk[7:10], "little"),
            )
        elif chunk_type == b"VP8L":
            if len(chunk) < 5 or chunk[0] != 0x2F or image_dimensions is not None:
                raise ValueError(FAILURE_BINARY_INVALID)
            packed = int.from_bytes(chunk[1:5], "little")
            image_dimensions = (1 + (packed & 0x3FFF), 1 + ((packed >> 14) & 0x3FFF))
        elif chunk_type == b"VP8 ":
            if (
                len(chunk) < 10
                or chunk[3:6] != b"\x9d\x01\x2a"
                or image_dimensions is not None
            ):
                raise ValueError(FAILURE_BINARY_INVALID)
            image_dimensions = (
                struct.unpack("<H", chunk[6:8])[0] & 0x3FFF,
                struct.unpack("<H", chunk[8:10])[0] & 0x3FFF,
            )
        offset = padded_end
        chunk_index += 1
    if offset != len(payload) or image_dimensions is None:
        raise ValueError(FAILURE_BINARY_INVALID)
    if canvas_dimensions is not None and canvas_dimensions != image_dimensions:
        raise ValueError(FAILURE_BINARY_INVALID)
    width, height = image_dimensions
    require_positive_dimensions(width, height)
    return "webp", "image/webp", width, height


def require_positive_dimensions(width: int, height: int) -> None:
    if width <= 0 or height <= 0:
        raise ValueError(FAILURE_BINARY_INVALID)
