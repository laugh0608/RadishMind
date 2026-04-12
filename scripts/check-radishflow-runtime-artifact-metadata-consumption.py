#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import build_citations, normalize_citations_from_document  # noqa: E402


def load_input_request(relative_path: str) -> dict:
    document = json.loads((REPO_ROOT / relative_path).read_text(encoding="utf-8"))
    return document["input_request"]


def require_equal(actual: object, expected: object, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, got {actual!r}")


def require_not_contains(text: str, needle: str, label: str) -> None:
    if needle in text:
        raise SystemExit(f"{label}: expected text not to contain {needle!r}, got {text!r}")


def main() -> int:
    uri_summary_request = load_input_request("datasets/eval/radishflow/explain-control-plane-uri-summary-only-001.json")
    uri_summary_citations = build_citations(uri_summary_request)
    first_uri_summary_metadata = uri_summary_request["artifacts"][0]["metadata"]
    require_equal(
        uri_summary_citations[0].get("excerpt"),
        first_uri_summary_metadata["summary"],
        "uri-summary-only excerpt should use metadata.summary",
    )
    require_not_contains(
        str(uri_summary_citations[0].get("excerpt") or ""),
        first_uri_summary_metadata["redaction_summary"],
        "uri-summary-only excerpt should not fall back to metadata.redaction_summary",
    )
    require_not_contains(
        str(uri_summary_citations[0].get("excerpt") or ""),
        str(first_uri_summary_metadata["source_scope"]),
        "uri-summary-only excerpt should not leak metadata.source_scope",
    )

    mixed_summary_request = load_input_request("datasets/eval/radishflow/explain-control-plane-mixed-summary-variants-001.json")
    mixed_summary_citations = build_citations(mixed_summary_request)
    first_mixed_metadata = mixed_summary_request["artifacts"][0]["metadata"]
    second_mixed_metadata = mixed_summary_request["artifacts"][1]["metadata"]
    require_equal(
        mixed_summary_citations[0].get("excerpt"),
        first_mixed_metadata["sanitized_summary"],
        "mixed-summary first excerpt should use metadata.sanitized_summary",
    )
    require_equal(
        mixed_summary_citations[1].get("excerpt"),
        second_mixed_metadata["summary"],
        "mixed-summary second excerpt should use metadata.summary",
    )
    require_not_contains(
        str(mixed_summary_citations[0].get("excerpt") or ""),
        str(first_mixed_metadata["summary_variant"]),
        "mixed-summary first excerpt should not leak metadata.summary_variant",
    )
    require_not_contains(
        str(mixed_summary_citations[0].get("excerpt") or ""),
        str(first_mixed_metadata["source_scope"]),
        "mixed-summary first excerpt should not leak metadata.source_scope",
    )
    normalized_partial = normalize_citations_from_document(
        [{"artifact_id": "private_registry_capture"}],
        mixed_summary_request,
        mixed_summary_citations,
    )
    require_equal(
        normalized_partial[0].get("id"),
        mixed_summary_citations[0].get("id"),
        "normalized partial citation should preserve fallback citation id for metadata locators",
    )

    redacted_request = load_input_request("datasets/eval/radishflow/explain-control-plane-redacted-support-summary-001.json")
    redacted_citations = build_citations(redacted_request)
    first_redacted_metadata = redacted_request["artifacts"][0]["metadata"]
    require_equal(
        redacted_citations[0].get("excerpt"),
        first_redacted_metadata["summary"],
        "redacted-support excerpt should use metadata.summary",
    )
    require_not_contains(
        str(redacted_citations[0].get("excerpt") or ""),
        str(first_redacted_metadata["redactions"]),
        "redacted-support excerpt should not serialize metadata.redactions",
    )
    require_not_contains(
        str(redacted_citations[0].get("label") or ""),
        "redactions",
        "redacted-support label should not treat metadata.redactions as display text",
    )

    audit_only_request = {
        "schema_version": 1,
        "request_id": "rf-runtime-artifact-metadata-audit-only-001",
        "project": "radishflow",
        "task": "explain_control_plane_state",
        "locale": "zh-CN",
        "artifacts": [
            {
                "kind": "attachment_ref",
                "role": "supporting",
                "name": "audit_only_capture",
                "mime_type": "application/json",
                "uri": "capture://radishflow/control-plane/audit-only-001",
                "metadata": {
                    "redactions": ["authorization", "cookie"],
                    "source_scope": "control_plane",
                    "summary_variant": "summary_plus_redaction",
                },
            }
        ],
        "context": {
            "document_revision": 1,
            "control_plane_state": {
                "entitlement_status": "active",
                "lease_status": "stale",
                "sync_status": "warning",
                "manifest_status": "current",
            },
        },
        "tool_hints": {
            "allow_retrieval": False,
            "allow_tool_calls": False,
            "allow_image_reasoning": False,
        },
        "safety": {
            "mode": "advisory",
            "requires_confirmation_for_actions": True,
        },
    }
    audit_only_citations = build_citations(audit_only_request)
    require_equal(
        audit_only_citations[0].get("label"),
        "audit_only_capture",
        "audit-only citation should fall back to artifact name instead of metadata audit fields",
    )
    require_equal(
        audit_only_citations[0].get("locator"),
        "artifact:audit_only_capture",
        "audit-only citation locator should not pretend an excerpt source exists",
    )
    require_equal(
        audit_only_citations[0].get("excerpt"),
        None,
        "audit-only citation should not synthesize excerpt from audit metadata fields",
    )
    normalized_audit_only = normalize_citations_from_document(
        [{"artifact_id": "audit_only_capture"}],
        audit_only_request,
        audit_only_citations,
    )
    require_equal(
        normalized_audit_only[0].get("id"),
        audit_only_citations[0].get("id"),
        "audit-only normalized citation should preserve fallback citation id",
    )
    require_equal(
        normalized_audit_only[0].get("excerpt"),
        None,
        "audit-only normalized citation should still avoid synthesizing excerpt from audit metadata fields",
    )

    print("radishflow runtime artifact metadata consumption regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
