#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import ensure_schema  # noqa: E402
from services.runtime.real_derived_negative_index import load_real_derived_negative_index  # noqa: E402

INDEX_PATH = REPO_ROOT / "datasets/eval/candidate-records/radishflow/batches/2026-04/rfb-ee35765bdc06/real-derived-index.json"
INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/real-derived-negative-index.schema.json"
MANIFEST_PATH = REPO_ROOT / "datasets/eval/candidate-records/radishflow/batches/2026-04/rfb-ee35765bdc06/manifest.json"
NEGATIVE_SAMPLE_DIR = REPO_ROOT / "datasets/eval/radishflow-negative"
SOURCE_MANIFEST_PATH = "datasets/eval/candidate-records/radishflow/batches/2026-04/rfb-22d033e634d2/manifest.json"


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
            "radishflow-ghost-completion-negative",
            "--sample-paths",
            "datasets/eval/radishflow-negative/radishflow-suggest-ghost-completion-negative-real-derived-flash-vapor-accept-key-drift-001.json",
            "datasets/eval/radishflow-negative/radishflow-suggest-ghost-completion-negative-real-derived-valve-ambiguous-accept-key-drift-001.json",
            "datasets/eval/radishflow-negative/radishflow-suggest-ghost-completion-negative-real-derived-empty-candidates-boundary-drift-001.json",
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

    document = load_real_derived_negative_index(INDEX_PATH)
    ensure_schema(document, INDEX_SCHEMA_PATH, str(INDEX_PATH.relative_to(REPO_ROOT)))
    index_document = expect_object(document, "real-derived negative index")

    summary = expect_object(index_document.get("summary"), "real-derived negative index summary")
    require_equal(summary.get("derived_record_count"), 3, "summary.derived_record_count")
    require_equal(summary.get("linked_negative_sample_count"), 3, "summary.linked_negative_sample_count")
    require_equal(summary.get("source_manifest_count"), 1, "summary.source_manifest_count")
    require_equal(summary.get("source_record_count"), 3, "summary.source_record_count")
    require_equal(summary.get("source_record_group_count"), 3, "summary.source_record_group_count")
    require_equal(summary.get("violation_group_count"), 3, "summary.violation_group_count")
    require_equal(summary.get("pattern_group_count"), 2, "summary.pattern_group_count")
    require_equal(summary.get("unlinked_derived_record_count"), 0, "summary.unlinked_derived_record_count")

    source_record_groups = index_document.get("source_record_groups")
    if not isinstance(source_record_groups, list) or len(source_record_groups) != 3:
        raise SystemExit("source_record_groups must contain exactly 3 groups")
    source_manifest_paths = sorted(
        {
            str(expect_object(group, "source_record_group").get("source_manifest_path") or "")
            for group in source_record_groups
        }
    )
    require_equal(source_manifest_paths, [SOURCE_MANIFEST_PATH], "source_record_groups source_manifest_path set")
    source_sample_ids = sorted(
        str(expect_object(group, "source_record_group").get("source_sample_id") or "") for group in source_record_groups
    )
    require_equal(
        source_sample_ids,
        [
            "radishflow-suggest-ghost-completion-chain-feed-valve-flash-stop-no-legal-outlet-001",
            "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
            "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
        ],
        "source_record_groups source_sample_id set",
    )
    for group in source_record_groups:
        require_equal(expect_object(group, "source_record_group").get("entry_count"), 1, "source_record_group.entry_count")

    violation_groups = index_document.get("violation_groups")
    if not isinstance(violation_groups, list) or len(violation_groups) != 3:
        raise SystemExit("violation_groups must contain exactly 3 groups")
    violation_signatures = sorted(
        tuple(expect_object(group, "violation_group").get("expected_candidate_violations") or [])
        for group in violation_groups
    )
    require_equal(
        violation_signatures,
        [
            (
                "Tab accept key must map to a legal candidate marked is_tab_default=true",
                "Tab accept key must map to a high-confidence legal candidate",
                "Tab accept key must not map to a legal candidate with conflict_flags",
                "json path '$.proposed_actions[0].preview.accept_key' should not equal 'Tab'",
            ),
            (
                "ghost_completion patch.candidate_ref must come from context.legal_candidate_completions",
                "json path '$.proposed_actions[0]' should not exist",
            ),
            (
                "json path '$.proposed_actions[0].preview.accept_key' expected 'Tab' but found 'manual_only'",
            ),
        ],
        "violation_groups signature set",
    )

    pattern_groups = index_document.get("pattern_groups")
    if not isinstance(pattern_groups, list) or len(pattern_groups) != 2:
        raise SystemExit("pattern_groups must contain exactly 2 groups")
    pattern_entry_counts: dict[tuple[str, ...], object] = {}
    for group in pattern_groups:
        group_object = expect_object(group, "pattern_group")
        derived_patterns = tuple(group_object.get("derived_patterns") or [])
        pattern_entry_counts[derived_patterns] = group_object.get("entry_count")
    require_equal(
        pattern_entry_counts,
        {
            ("ghost_accept_key_policy_drift",): 2,
            ("ghost_empty_candidate_boundary_drift",): 1,
        },
        "pattern_groups entry_count map",
    )
    require_contains(
        list(pattern_entry_counts.keys()),
        ("ghost_accept_key_policy_drift",),
        "pattern_groups derived_patterns",
    )
    require_contains(
        list(pattern_entry_counts.keys()),
        ("ghost_empty_candidate_boundary_drift",),
        "pattern_groups derived_patterns",
    )

    unlinked_derived_records = index_document.get("unlinked_derived_records")
    if not isinstance(unlinked_derived_records, list):
        raise SystemExit("unlinked_derived_records must be a list")
    require_equal(len(unlinked_derived_records), 0, "unlinked_derived_records length")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
