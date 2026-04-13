from __future__ import annotations

from typing import Any

from .artifact_metadata import (
    ARTIFACT_SUMMARY_METADATA_FIELDS,
    ARTIFACT_SUMMARY_METADATA_KEYS,
    artifact_summary_metadata_hint,
)


def normalize_summary_text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return " ".join(value.split()).strip()
    return " ".join(str(value).split()).strip()


def extract_artifact_summary_metadata(metadata: Any) -> tuple[str, str, str]:
    if not isinstance(metadata, dict):
        return "", "", ""
    for key, label_suffix in ARTIFACT_SUMMARY_METADATA_FIELDS:
        text = normalize_summary_text(metadata.get(key))
        if text:
            return text, key, label_suffix
    return "", "", ""


def has_artifact_summary_metadata(metadata: Any) -> bool:
    text, _, _ = extract_artifact_summary_metadata(metadata)
    return bool(text)
