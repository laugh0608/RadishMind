#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.inference import coerce_response_document  # noqa: E402


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
    mixed_summary_request = load_input_request("datasets/eval/radishflow/explain-control-plane-mixed-summary-variants-001.json")
    mixed_summary_response = {
        "status": "partial",
        "summary": "control plane summary remains partially degraded",
        "answers": [
            {
                "kind": "state_summary",
                "text": "private registry capture should stay grounded on the sanitized metadata summary",
                "citation_ids": [{"artifact_index": 0}],
            }
        ],
        "issues": [
            {
                "code": "CONTROL_PLANE_SCOPE_SPLIT",
                "message": "package source and cached lease freshness remain split",
                "severity": "warning",
                "citation_ids": [{"artifact_index": 0}],
            }
        ],
        "proposed_actions": [
            {
                "kind": "read_only_check",
                "title": "Review capture summary",
                "rationale": "keep action grounded on the same sanitized artifact summary",
                "citation_ids": [{"artifact_index": 0}],
                "risk_level": "low",
                "requires_confirmation": False,
                "target": {
                    "type": "control_plane",
                    "id": "review-private-registry-capture",
                },
            }
        ],
        "citations": [
            {
                "artifact_index": 0,
            }
        ],
        "risk_level": "low",
        "requires_confirmation": False,
    }
    coerced_mixed_summary = coerce_response_document(mixed_summary_response, mixed_summary_request, raw_text="")
    require_equal(
        coerced_mixed_summary["answers"][0]["citation_ids"],
        ["ref-1"],
        "mixed-summary answer artifact_index should normalize to fallback citation id",
    )
    require_equal(
        coerced_mixed_summary["issues"][0]["citation_ids"],
        ["ref-1"],
        "mixed-summary issue artifact_index should normalize to fallback citation id",
    )
    require_equal(
        coerced_mixed_summary["proposed_actions"][0]["citation_ids"],
        ["ref-1"],
        "mixed-summary action artifact_index should normalize to fallback citation id",
    )
    require_equal(
        coerced_mixed_summary["citations"][0]["id"],
        "ref-1",
        "mixed-summary citation artifact_index should preserve fallback citation id",
    )
    require_equal(
        coerced_mixed_summary["citations"][0]["locator"],
        "artifact:private_registry_capture.metadata.sanitized_summary",
        "mixed-summary citation artifact_index should preserve metadata locator",
    )
    require_equal(
        coerced_mixed_summary["citations"][0]["excerpt"],
        mixed_summary_request["artifacts"][0]["metadata"]["sanitized_summary"],
        "mixed-summary citation artifact_index should preserve sanitized_summary excerpt",
    )
    require_not_contains(
        str(coerced_mixed_summary["citations"][0]["excerpt"]),
        str(mixed_summary_request["artifacts"][0]["metadata"]["source_scope"]),
        "mixed-summary coerced citation excerpt should not leak source_scope",
    )
    require_not_contains(
        str(coerced_mixed_summary["citations"][0]["excerpt"]),
        str(mixed_summary_request["artifacts"][0]["metadata"]["summary_variant"]),
        "mixed-summary coerced citation excerpt should not leak summary_variant",
    )

    redacted_request = load_input_request("datasets/eval/radishflow/explain-control-plane-redacted-support-summary-001.json")
    redacted_response = {
        "status": "partial",
        "summary": "refresh path remains blocked",
        "answers": [
            {
                "kind": "conflict_explanation",
                "text": "refresh capture should stay grounded on the redacted summary only",
                "citation_ids": [{"artifact_id": "lease_refresh_capture"}],
            }
        ],
        "issues": [
            {
                "code": "CONTROL_PLANE_REFRESH_BLOCKED",
                "message": "refresh chain remains blocked",
                "severity": "warning",
                "citation_ids": [{"artifact_id": "lease_refresh_capture"}],
            }
        ],
        "proposed_actions": [
            {
                "kind": "candidate_operation",
                "title": "Review refresh capture",
                "rationale": "the action should cite the existing redacted artifact citation id",
                "citation_ids": [{"artifact_id": "lease_refresh_capture"}],
                "risk_level": "medium",
                "requires_confirmation": True,
                "target": {
                    "type": "control_plane",
                    "id": "review-redacted-refresh-capture",
                },
            }
        ],
        "citations": [
            {
                "artifact_id": "lease_refresh_capture",
            }
        ],
        "risk_level": "medium",
        "requires_confirmation": True,
    }
    coerced_redacted = coerce_response_document(redacted_response, redacted_request, raw_text="")
    require_equal(
        coerced_redacted["answers"][0]["citation_ids"],
        ["ref-1"],
        "redacted answer artifact_id should normalize to fallback citation id",
    )
    require_equal(
        coerced_redacted["issues"][0]["citation_ids"],
        ["ref-1"],
        "redacted issue artifact_id should normalize to fallback citation id",
    )
    require_equal(
        coerced_redacted["proposed_actions"][0]["citation_ids"],
        ["ref-1"],
        "redacted action artifact_id should normalize to fallback citation id",
    )
    require_equal(
        coerced_redacted["citations"][0]["id"],
        "ref-1",
        "redacted citation artifact_id should preserve fallback citation id",
    )
    require_equal(
        coerced_redacted["citations"][0]["excerpt"],
        redacted_request["artifacts"][0]["metadata"]["summary"],
        "redacted citation artifact_id should preserve summary excerpt",
    )
    require_not_contains(
        str(coerced_redacted["citations"][0]["excerpt"]),
        str(redacted_request["artifacts"][0]["metadata"]["redactions"]),
        "redacted coerced citation excerpt should not serialize redactions metadata",
    )

    print("radishflow runtime artifact metadata response coercion regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
