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
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-attachment-mixed-missing-read-only-check-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-attachment-mixed-missing-read-only-check-issue-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-faq-missing-read-only-check-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-faq-missing-read-only-check-confirmation-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-faq-missing-read-only-check-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-faq-missing-read-only-check-issue-confirmation-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-forum-conflict-citation-drift-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-attachments-forum-conflict-missing-read-only-check-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-faq-forum-conflict-citation-drift-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-faq-forum-conflict-missing-read-only-check-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-docs-faq-mixed-missing-read-only-check-issue-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-direct-answer-missing-answer-issue-action-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-direct-answer-missing-answer-issue-action-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-direct-answer-missing-read-only-check-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-evidence-gap-multi-issues-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-evidence-gap-unconfirmed-operation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-evidence-gap-unconfirmed-operation-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-forum-supplement-citation-drift-action-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-forum-supplement-citation-drift-issue-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-forum-supplement-missing-answer-issue-action-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-navigation-missing-read-only-check-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-role-boundary-missing-read-only-check-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-role-boundary-multi-issues-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-role-boundary-multi-issues-confirmation-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-wiki-faq-citation-drift-action-confirmation-001.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-wiki-faq-citation-drift-action-confirmation-002.json",
            "datasets/eval/radish-negative/answer-docs-question-negative-real-derived-wiki-faq-mixed-missing-read-only-check-issue-001.json",
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
    require_equal(summary.get("derived_record_count"), 27, "summary.derived_record_count")
    require_equal(summary.get("linked_negative_sample_count"), 27, "summary.linked_negative_sample_count")
    require_equal(summary.get("source_manifest_count"), 2, "summary.source_manifest_count")
    require_equal(summary.get("source_record_count"), 17, "summary.source_record_count")
    require_equal(summary.get("source_record_group_count"), 17, "summary.source_record_group_count")
    require_equal(summary.get("violation_group_count"), 13, "summary.violation_group_count")
    require_equal(summary.get("pattern_group_count"), 8, "summary.pattern_group_count")
    require_equal(summary.get("unlinked_derived_record_count"), 0, "summary.unlinked_derived_record_count")

    source_record_groups = index_document.get("source_record_groups")
    if not isinstance(source_record_groups, list) or len(source_record_groups) != 17:
        raise SystemExit("source_record_groups must contain exactly 17 groups")

    source_manifest_paths = sorted(
        {
            str(expect_object(group, "source_record_group").get("source_manifest_path") or "")
            for group in source_record_groups
        }
    )
    require_equal(
        source_manifest_paths,
        [
            "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
            "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
        ],
        "source_record_groups source_manifest_path set",
    )

    source_sample_ids = sorted(
        str(expect_object(group, "source_record_group").get("source_sample_id") or "") for group in source_record_groups
    )
    require_equal(
        sorted(set(source_sample_ids)),
        [
            "radish-answer-docs-question-attachment-mixed-001",
            "radish-answer-docs-question-direct-answer-001",
            "radish-answer-docs-question-docs-attachments-faq-001",
            "radish-answer-docs-question-docs-attachments-forum-conflict-001",
            "radish-answer-docs-question-docs-faq-forum-conflict-001",
            "radish-answer-docs-question-docs-faq-mixed-001",
            "radish-answer-docs-question-evidence-gap-001",
            "radish-answer-docs-question-forum-supplement-001",
            "radish-answer-docs-question-navigation-001",
            "radish-answer-docs-question-role-example-boundary-001",
            "radish-answer-docs-question-wiki-faq-mixed-001",
        ],
        "source_record_groups unique source_sample_id set",
    )
    source_group_entry_counts = {
        (
            str(expect_object(group, "source_record_group").get("source_manifest_path") or ""),
            str(expect_object(group, "source_record_group").get("source_sample_id") or ""),
        ): expect_object(group, "source_record_group").get("entry_count")
        for group in source_record_groups
    }
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-attachment-mixed-001",
            )
        ),
        2,
        "source_record_groups 2026-04-05 attachment-mixed entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-direct-answer-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 direct-answer entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-direct-answer-001",
            )
        ),
        1,
        "source_record_groups 2026-04-05 direct-answer entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-docs-attachments-forum-conflict-001",
            )
        ),
        2,
        "source_record_groups 2026-04-05 docs-attachments-forum-conflict entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-docs-attachments-faq-001",
            )
        ),
        2,
        "source_record_groups 2026-04-05 docs-attachments-faq entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-docs-faq-forum-conflict-001",
            )
        ),
        2,
        "source_record_groups 2026-04-05 docs-faq-forum-conflict entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1/2026-04-05-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-forum-supplement-001",
            )
        ),
        1,
        "source_record_groups 2026-04-05 forum-supplement entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-docs-attachments-faq-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 docs-attachments-faq entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-evidence-gap-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 evidence-gap entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-forum-supplement-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 forum-supplement entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-role-example-boundary-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 role-example-boundary entry_count",
    )
    require_equal(
        source_group_entry_counts.get(
            (
                "datasets/eval/candidate-records/radish/2026-04-04-radish-docs-qa-real-batch-v1.manifest.json",
                "radish-answer-docs-question-wiki-faq-mixed-001",
            )
        ),
        2,
        "source_record_groups 2026-04-04 wiki-faq-mixed entry_count",
    )

    violation_groups = index_document.get("violation_groups")
    if not isinstance(violation_groups, list) or len(violation_groups) != 13:
        raise SystemExit("violation_groups must contain exactly 13 groups")

    pattern_groups = index_document.get("pattern_groups")
    if not isinstance(pattern_groups, list) or len(pattern_groups) != 8:
        raise SystemExit("pattern_groups must contain exactly 8 groups")

    pattern_entry_counts = {
        tuple(list(expect_object(group, "pattern_group").get("derived_patterns") or [])): expect_object(
            group,
            "pattern_group",
        ).get("entry_count")
        for group in pattern_groups
    }
    require_equal(
        pattern_entry_counts.get(("missing_read_only_check_confirmation_drift",)),
        6,
        "pattern_groups missing_read_only_check_confirmation_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("missing_read_only_check_issue_confirmation_drift",)),
        4,
        "pattern_groups missing_read_only_check_issue_confirmation_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("missing_read_only_check_issue_drift",)),
        3,
        "pattern_groups missing_read_only_check_issue_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("missing_answer_issue_action_drift",)),
        3,
        "pattern_groups missing_answer_issue_action_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("citation_drift_issue_read_only_confirmation_drift",)),
        3,
        "pattern_groups citation_drift_issue_read_only_confirmation_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("evidence_gap_confirmation_operation_drift",)),
        2,
        "pattern_groups evidence_gap_confirmation_operation_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("multi_issue_confirmation_drift",)),
        3,
        "pattern_groups multi_issue_confirmation_drift entry_count",
    )
    require_equal(
        pattern_entry_counts.get(("citation_drift_issue_action_confirmation_drift",)),
        3,
        "pattern_groups citation_drift_issue_action_confirmation_drift entry_count",
    )

    flattened_patterns = [
        pattern
        for group in pattern_groups
        for pattern in list(expect_object(group, "pattern_group").get("derived_patterns") or [])
    ]
    require_contains(
        flattened_patterns,
        "missing_answer_issue_action_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "citation_drift_issue_read_only_confirmation_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "missing_read_only_check_confirmation_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "missing_read_only_check_issue_confirmation_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "missing_read_only_check_issue_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "evidence_gap_confirmation_operation_drift",
        "pattern_groups derived_patterns",
    )
    require_contains(
        flattened_patterns,
        "multi_issue_confirmation_drift",
        "pattern_groups derived_patterns",
    )

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
        "is missing required action kind 'read_only_check'",
        "violation_groups expected_candidate_violations",
    )
    require_contains(
        flattened_violations,
        "referenced citation id 'forum-1' is missing from candidate_response.citations",
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
