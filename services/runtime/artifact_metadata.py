from __future__ import annotations


ARTIFACT_SUMMARY_METADATA_FIELDS = (
    ("summary", "summary"),
    ("sanitized_summary", "sanitized summary"),
    ("redaction_summary", "redaction summary"),
)
ARTIFACT_SUMMARY_METADATA_KEYS = tuple(field[0] for field in ARTIFACT_SUMMARY_METADATA_FIELDS)

ARTIFACT_SUPPORTING_METADATA_KEYS = (
    *ARTIFACT_SUMMARY_METADATA_KEYS,
    "redactions",
    "source_scope",
    "summary_variant",
)
ARTIFACT_RETRIEVAL_METADATA_KEYS = (
    "source_type",
    "page_slug",
    "fragment_id",
    "retrieval_rank",
    "is_official",
)
COPILOT_REQUEST_ARTIFACT_METADATA_KEYS = (
    *ARTIFACT_RETRIEVAL_METADATA_KEYS,
    *ARTIFACT_SUPPORTING_METADATA_KEYS,
)

RAW_ARTIFACT_PAYLOAD_METADATA_KEYS = (
    "headers",
    "request_headers",
    "response_headers",
    "request_body",
    "response_body",
    "raw_body",
    "raw_payload",
    "raw_request",
    "raw_response",
    "full_payload",
    "full_response",
    "stack_trace",
    "traceback",
)


def artifact_summary_metadata_hint() -> str:
    return " / ".join(f"metadata.{key}" for key in ARTIFACT_SUMMARY_METADATA_KEYS)
