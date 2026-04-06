#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
)

SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json"
RECOMMENDED_SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json"
REAL_BATCH_OUTPUT_ROOT = REPO_ROOT / "datasets/eval/candidate-records/radish/2026-04-05-radish-docs-qa-real-batch-v1"
REAL_BATCH_COLLECTION = "2026-04-05-radish-docs-qa-real-batch-v1"
REAL_BATCH_CROSS_SAMPLE_NEGATIVE_DIR = REPO_ROOT / "datasets/eval/radish-negative"
SAME_SAMPLE_TOP = 2
CROSS_SAMPLE_TOP = 1


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def run_command(command: list[str]) -> None:
    result = subprocess.run(
        command,
        cwd=REPO_ROOT,
        capture_output=True,
        text=True,
    )
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    if result.returncode != 0:
        raise SystemExit(result.returncode)


def require_equal(actual: Any, expected: Any, label: str) -> None:
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected!r}, got {actual!r}")


def require_true(value: Any, label: str) -> None:
    if value is not True:
        raise SystemExit(f"{label} must be true")


def require_non_empty_list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list) or not value:
        raise SystemExit(f"{label} must be a non-empty json array")
    return value


def load_summary(path: Path, schema_path: Path) -> dict[str, Any]:
    document = expect_object(load_json_document(path), make_repo_relative(path))
    ensure_schema(document, schema_path, make_repo_relative(path))
    return document


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="radish-docs-qa-real-batch-dual-summary-") as temp_dir:
        temp_root = Path(temp_dir)
        audit_report_path = temp_root / f"{REAL_BATCH_COLLECTION}.audit.json"
        replay_index_path = temp_root / f"{REAL_BATCH_COLLECTION}.negative-replay-index.json"
        cross_sample_replay_index_path = temp_root / f"{REAL_BATCH_COLLECTION}.cross-sample-replay-index.json"
        artifact_summary_path = temp_root / f"{REAL_BATCH_COLLECTION}.artifacts.json"
        same_sample_summary_path = (
            temp_root / f"{REAL_BATCH_COLLECTION}.recommended-negative-replay-top{SAME_SAMPLE_TOP}-same_sample.summary.json"
        )
        cross_sample_summary_path = (
            temp_root / f"{REAL_BATCH_COLLECTION}.recommended-negative-replay-top{CROSS_SAMPLE_TOP}-cross_sample.summary.json"
        )
        same_sample_negative_output_dir = temp_root / "same-sample-negative-replay"

        run_command(
            [
                sys.executable,
                str(REPO_ROOT / "scripts" / "run-radish-docs-qa-real-batch.py"),
                "--provider",
                "openai-compatible",
                "--output-root",
                make_repo_relative(REAL_BATCH_OUTPUT_ROOT),
                "--collection-batch",
                REAL_BATCH_COLLECTION,
                "--resume",
                "--audit-report-output",
                str(audit_report_path),
                "--replay-index-output",
                str(replay_index_path),
                "--cross-sample-replay-index-output",
                str(cross_sample_replay_index_path),
                "--artifact-summary-output",
                str(artifact_summary_path),
                "--build-negative-replay",
                "--negative-output-dir",
                str(same_sample_negative_output_dir),
                "--build-recommended-negative-replay-summary",
                "--recommended-groups-top",
                str(SAME_SAMPLE_TOP),
                "--recommended-summary-output",
                str(same_sample_summary_path),
                "--cross-sample-recommended-groups-top",
                str(CROSS_SAMPLE_TOP),
                "--cross-sample-recommended-summary-output",
                str(cross_sample_summary_path),
                "--cross-sample-negative-sample-dir",
                make_repo_relative(REAL_BATCH_CROSS_SAMPLE_NEGATIVE_DIR),
                "--fail-on-recommended-replay-violation",
            ]
        )

        artifact_summary = load_summary(artifact_summary_path, SUMMARY_SCHEMA_PATH)
        same_sample_summary = load_summary(same_sample_summary_path, RECOMMENDED_SUMMARY_SCHEMA_PATH)
        cross_sample_summary = load_summary(cross_sample_summary_path, RECOMMENDED_SUMMARY_SCHEMA_PATH)

        require_equal(artifact_summary.get("provider"), "openai-compatible", "artifact summary provider")
        execution = expect_object(artifact_summary.get("execution"), "artifact summary execution")
        require_equal(execution.get("recommended_negative_replay_exit_code"), 0, "same-sample replay exit code")
        require_equal(
            execution.get("cross_sample_recommended_negative_replay_exit_code"),
            0,
            "cross-sample replay exit code",
        )

        artifacts = expect_object(artifact_summary.get("artifacts"), "artifact summary artifacts")
        same_sample_artifact = expect_object(
            artifacts.get("recommended_negative_replay_summary"),
            "artifact summary artifacts.recommended_negative_replay_summary",
        )
        cross_sample_artifact = expect_object(
            artifacts.get("cross_sample_recommended_negative_replay_summary"),
            "artifact summary artifacts.cross_sample_recommended_negative_replay_summary",
        )
        require_true(
            same_sample_artifact.get("requested"),
            "artifact summary artifacts.recommended_negative_replay_summary.requested",
        )
        require_true(
            same_sample_artifact.get("exists"),
            "artifact summary artifacts.recommended_negative_replay_summary.exists",
        )
        require_equal(
            same_sample_artifact.get("path"),
            str(same_sample_summary_path),
            "artifact summary same-sample summary path",
        )
        require_true(
            cross_sample_artifact.get("requested"),
            "artifact summary artifacts.cross_sample_recommended_negative_replay_summary.requested",
        )
        require_true(
            cross_sample_artifact.get("exists"),
            "artifact summary artifacts.cross_sample_recommended_negative_replay_summary.exists",
        )
        require_equal(
            cross_sample_artifact.get("path"),
            str(cross_sample_summary_path),
            "artifact summary cross-sample summary path",
        )

        summary = expect_object(artifact_summary.get("summary"), "artifact summary summary")
        require_equal(
            summary.get("recommended_replay_group_count"),
            SAME_SAMPLE_TOP,
            "artifact summary same-sample recommended group count",
        )
        require_equal(
            summary.get("cross_sample_recommended_replay_group_count"),
            CROSS_SAMPLE_TOP,
            "artifact summary cross-sample recommended group count",
        )

        recommended_negative_replays = expect_object(
            artifact_summary.get("recommended_negative_replays"),
            "artifact summary recommended_negative_replays",
        )
        same_sample_group_ids = require_non_empty_list(
            recommended_negative_replays.get("recommended_group_ids"),
            "artifact summary recommended_negative_replays.recommended_group_ids",
        )
        cross_sample_group_ids = require_non_empty_list(
            recommended_negative_replays.get("cross_sample_recommended_group_ids"),
            "artifact summary recommended_negative_replays.cross_sample_recommended_group_ids",
        )
        if len(same_sample_group_ids) < SAME_SAMPLE_TOP:
            raise SystemExit("artifact summary same-sample recommended_group_ids is shorter than the requested top")
        if len(cross_sample_group_ids) < CROSS_SAMPLE_TOP:
            raise SystemExit("artifact summary cross-sample recommended_group_ids is shorter than the requested top")

        require_equal(
            same_sample_summary.get("batch_artifact_summary_path"),
            str(artifact_summary_path),
            "same-sample summary batch_artifact_summary_path",
        )
        require_equal(
            cross_sample_summary.get("batch_artifact_summary_path"),
            str(artifact_summary_path),
            "cross-sample summary batch_artifact_summary_path",
        )

        same_sample_selection = expect_object(same_sample_summary.get("selection"), "same-sample summary selection")
        require_equal(same_sample_selection.get("replay_mode"), "same_sample", "same-sample summary replay_mode")
        require_equal(same_sample_selection.get("requested_top"), SAME_SAMPLE_TOP, "same-sample summary requested_top")

        cross_sample_selection = expect_object(cross_sample_summary.get("selection"), "cross-sample summary selection")
        require_equal(cross_sample_selection.get("replay_mode"), "cross_sample", "cross-sample summary replay_mode")
        require_equal(
            cross_sample_selection.get("requested_top"),
            CROSS_SAMPLE_TOP,
            "cross-sample summary requested_top",
        )

    print("radish docs qa real batch dual recommended summary check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
