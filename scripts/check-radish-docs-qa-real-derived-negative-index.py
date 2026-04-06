#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import ensure_schema, load_json_document  # noqa: E402

INDEX_PATH = REPO_ROOT / (
    "datasets/eval/candidate-records/radish-negative/"
    "2026-04-04-radish-docs-qa-simulated-negatives-v1.real-derived-index.json"
)
INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/real-derived-negative-index.schema.json"
MANIFEST_PATH = REPO_ROOT / (
    "datasets/eval/candidate-records/radish-negative/2026-04-04-radish-docs-qa-simulated-negatives-v1.manifest.json"
)
NEGATIVE_SAMPLE_DIR = REPO_ROOT / "datasets/eval/radish-negative"


def expect_object(document: object, label: str) -> dict[str, object]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def require_equal(actual: object, expected: object, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, found {actual!r}")


def require_contains(values: list[object], expected: object, label: str) -> None:
    if expected not in values:
        raise SystemExit(f"{label}: missing expected value {expected!r}")


def main() -> int:
    subprocess.run(
        [
            sys.executable,
            str(REPO_ROOT / "scripts/run-eval-regression.py"),
            "radish-docs-qa-negative",
            "--sample-paths",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-evidence-gap-unconfirmed-operation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-forum-supplement-missing-answer-issue-action-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-role-boundary-multi-issues-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-wiki-faq-citation-drift-action-confirmation-001.json",
            "--fail-on-violation",
        ],
        cwd=REPO_ROOT,
        check=True,
    )

    subprocess.run(
        [
            sys.executable,
            str(REPO_ROOT / "scripts/build-real-derived-negative-index.py"),
            "--manifest",
            str(MANIFEST_PATH),
            "--negative-sample-dir",
            str(NEGATIVE_SAMPLE_DIR),
            "--output",
            str(INDEX_PATH),
            "--check",
        ],
        cwd=REPO_ROOT,
        check=True,
    )

    document = load_json_document(INDEX_PATH)
    ensure_schema(document, INDEX_SCHEMA_PATH, str(INDEX_PATH.relative_to(REPO_ROOT)))
    index_document = expect_object(document, "real-derived negative index")

    summary = expect_object(index_document.get("summary"), "real-derived negative index summary")
    require_equal(summary.get("derived_record_count"), 4, "summary.derived_record_count")
    require_equal(summary.get("linked_negative_sample_count"), 4, "summary.linked_negative_sample_count")
    require_equal(summary.get("source_manifest_count"), 1, "summary.source_manifest_count")
    require_equal(summary.get("source_record_count"), 4, "summary.source_record_count")
    require_equal(summary.get("source_record_group_count"), 4, "summary.source_record_group_count")
    require_equal(summary.get("violation_group_count"), 4, "summary.violation_group_count")
    require_equal(summary.get("unlinked_derived_record_count"), 0, "summary.unlinked_derived_record_count")

    source_record_groups = index_document.get("source_record_groups")
    if not isinstance(source_record_groups, list) or len(source_record_groups) != 4:
        raise SystemExit("source_record_groups must contain exactly 4 groups")

    source_sample_ids = sorted(
        str(expect_object(group, "source_record_group").get("source_sample_id") or "") for group in source_record_groups
    )
    require_equal(
        source_sample_ids,
        [
            "radish-answer-docs-question-evidence-gap-001",
            "radish-answer-docs-question-forum-supplement-001",
            "radish-answer-docs-question-role-example-boundary-001",
            "radish-answer-docs-question-wiki-faq-mixed-001",
        ],
        "source_record_groups source_sample_id set",
    )

    violation_groups = index_document.get("violation_groups")
    if not isinstance(violation_groups, list) or len(violation_groups) != 4:
        raise SystemExit("violation_groups must contain exactly 4 groups")

    flattened_violations = [
        violation
        for group in violation_groups
        for violation in list(expect_object(group, "violation_group").get("expected_candidate_violations") or [])
    ]
    require_contains(
        flattened_violations,
        "must contain at least 1 answer",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "should not contain issues",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "should not contain proposed_actions",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "referenced citation id 'faq-1' is missing from candidate_response.citations",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "first answer must cite at least one primary artifact for official_source_precedence samples",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "json path '$.requires_confirmation' should not equal 'True'",
        "violation_groups expected_candidate_violations",
    )

    print("radish docs qa real-derived negative index check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
