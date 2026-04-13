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


def main() -> int:
    uri_summary_request = load_input_request("datasets/eval/radishflow/explain-control-plane-uri-summary-only-001.json")
    uri_summary_citations = build_citations(uri_summary_request)
    require_equal(
        uri_summary_citations[0].get("locator"),
        "artifact:lease_refresh_capture.metadata.summary",
        "uri-summary-only first citation locator",
    )
    require_equal(
        uri_summary_citations[0].get("excerpt"),
        "lease refresh capture shows a stale local lease handle after a 401 response; authorization and cookie fields were removed",
        "uri-summary-only first citation excerpt",
    )
    require_equal(
        uri_summary_citations[1].get("locator"),
        "artifact:package_sync_capture.metadata.summary",
        "uri-summary-only second citation locator",
    )

    mixed_summary_request = load_input_request("datasets/eval/radishflow/explain-control-plane-mixed-summary-variants-001.json")
    mixed_summary_citations = build_citations(mixed_summary_request)
    require_equal(
        mixed_summary_citations[0].get("locator"),
        "artifact:private_registry_capture.metadata.sanitized_summary",
        "mixed-summary first citation locator",
    )
    require_equal(
        mixed_summary_citations[0].get("label"),
        "private_registry_capture sanitized summary",
        "mixed-summary first citation label",
    )
    require_equal(
        mixed_summary_citations[0].get("excerpt"),
        "private package source returned repeated 429 responses; request and response bodies plus auth headers were stripped",
        "mixed-summary first citation excerpt",
    )
    require_equal(
        mixed_summary_citations[1].get("locator"),
        "artifact:lease_cache_capture.metadata.summary",
        "mixed-summary second citation locator",
    )
    require_equal(
        mixed_summary_citations[1].get("label"),
        "lease_cache_capture summary",
        "mixed-summary second citation label",
    )
    require_equal(
        mixed_summary_citations[2].get("locator"),
        "artifact:source_scope_rollup",
        "mixed-summary json rollup locator",
    )
    require_equal(
        mixed_summary_citations[2].get("excerpt"),
        "{\"public_manifest_status\": \"ok\", \"private_source_status\": \"429\", \"cached_lease_minutes_remaining\": 14, \"scope_split\": true}",
        "mixed-summary json rollup excerpt",
    )

    normalized_partial = normalize_citations_from_document(
        [{"artifact_id": "private_registry_capture"}],
        mixed_summary_request,
        mixed_summary_citations,
    )
    require_equal(
        normalized_partial[0].get("locator"),
        "artifact:private_registry_capture.metadata.sanitized_summary",
        "normalized partial citation locator",
    )
    require_equal(
        normalized_partial[0].get("excerpt"),
        "private package source returned repeated 429 responses; request and response bodies plus auth headers were stripped",
        "normalized partial citation excerpt",
    )

    print("radishflow artifact summary consumption regression passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
