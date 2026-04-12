from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
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
PENDING_SAME_SAMPLE_SUMMARY_NAME = "pending.recommended-negative-replay.summary.json"
PENDING_CROSS_SAMPLE_SUMMARY_NAME = "pending.cross-sample-recommended-negative-replay.summary.json"
EXPECTED_REAL_BATCH_FAILURE_NOTE = (
    "INFO: the next Radish docs QA real-batch audit output intentionally replays known failing "
    "captured candidate responses to build recommended negative replay summaries; printed FAIL/WARNING "
    "lines below are expected input signals unless this checker itself exits non-zero."
)


@dataclass(frozen=True)
class RealBatchSummaryPaths:
    temp_root: Path
    audit_report_path: Path
    replay_index_path: Path
    cross_sample_replay_index_path: Path
    artifact_summary_path: Path
    same_sample_negative_output_dir: Path
    same_sample_summary_path: Path
    cross_sample_summary_path: Path


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def run_command(command: list[str]) -> None:
    print(EXPECTED_REAL_BATCH_FAILURE_NOTE, file=sys.stderr)
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


def require_false(value: Any, label: str) -> None:
    if value is not False:
        raise SystemExit(f"{label} must be false")


def require_non_empty_list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list) or not value:
        raise SystemExit(f"{label} must be a non-empty json array")
    return value


def require_missing_path(path: Path, label: str) -> None:
    if path.exists():
        raise SystemExit(f"{label} should not exist: {path}")


def build_real_batch_summary_paths(
    temp_root: Path,
    *,
    same_sample_top: int = 0,
    cross_sample_top: int = 0,
) -> RealBatchSummaryPaths:
    same_sample_summary_path = (
        temp_root / f"{REAL_BATCH_COLLECTION}.recommended-negative-replay-top{same_sample_top}-same_sample.summary.json"
        if same_sample_top > 0
        else temp_root / PENDING_SAME_SAMPLE_SUMMARY_NAME
    )
    cross_sample_summary_path = (
        temp_root / f"{REAL_BATCH_COLLECTION}.recommended-negative-replay-top{cross_sample_top}-cross_sample.summary.json"
        if cross_sample_top > 0
        else temp_root / PENDING_CROSS_SAMPLE_SUMMARY_NAME
    )
    return RealBatchSummaryPaths(
        temp_root=temp_root,
        audit_report_path=temp_root / f"{REAL_BATCH_COLLECTION}.audit.json",
        replay_index_path=temp_root / f"{REAL_BATCH_COLLECTION}.negative-replay-index.json",
        cross_sample_replay_index_path=temp_root / f"{REAL_BATCH_COLLECTION}.cross-sample-replay-index.json",
        artifact_summary_path=temp_root / f"{REAL_BATCH_COLLECTION}.artifacts.json",
        same_sample_negative_output_dir=temp_root / "same-sample-negative-replay",
        same_sample_summary_path=same_sample_summary_path,
        cross_sample_summary_path=cross_sample_summary_path,
    )


def build_real_batch_summary_command(
    paths: RealBatchSummaryPaths,
    *,
    same_sample_top: int = 0,
    cross_sample_top: int = 0,
    recommended_replay_mode: str = "",
) -> list[str]:
    command = [
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
        str(paths.audit_report_path),
        "--replay-index-output",
        str(paths.replay_index_path),
        "--cross-sample-replay-index-output",
        str(paths.cross_sample_replay_index_path),
        "--artifact-summary-output",
        str(paths.artifact_summary_path),
        "--build-negative-replay",
        "--negative-output-dir",
        str(paths.same_sample_negative_output_dir),
        "--build-recommended-negative-replay-summary",
        "--cross-sample-negative-sample-dir",
        make_repo_relative(REAL_BATCH_CROSS_SAMPLE_NEGATIVE_DIR),
        "--fail-on-recommended-replay-violation",
    ]
    if recommended_replay_mode:
        command.extend(["--recommended-replay-mode", recommended_replay_mode])
    if same_sample_top > 0:
        command.extend(["--recommended-groups-top", str(same_sample_top)])
    if cross_sample_top > 0:
        command.extend(["--cross-sample-recommended-groups-top", str(cross_sample_top)])

    if recommended_replay_mode == "cross_sample":
        command.extend(["--recommended-summary-output", str(paths.cross_sample_summary_path)])
    else:
        command.extend(["--recommended-summary-output", str(paths.same_sample_summary_path)])

    if cross_sample_top > 0 and recommended_replay_mode != "cross_sample":
        command.extend(
            [
                "--cross-sample-recommended-summary-output",
                str(paths.cross_sample_summary_path),
            ]
        )
    return command


def require_artifact_entry_state(
    artifacts: dict[str, Any],
    *,
    field_name: str,
    expected_path: Path,
    expected_requested: bool,
    expected_exists: bool,
    label: str,
) -> dict[str, Any]:
    artifact = expect_object(artifacts.get(field_name), label)
    if expected_requested:
        require_true(artifact.get("requested"), f"{label}.requested")
    else:
        require_false(artifact.get("requested"), f"{label}.requested")
    if expected_exists:
        require_true(artifact.get("exists"), f"{label}.exists")
    else:
        require_false(artifact.get("exists"), f"{label}.exists")
    require_equal(artifact.get("path"), str(expected_path), f"{label}.path")
    return artifact


def require_summary_selection(
    summary: dict[str, Any],
    *,
    replay_mode: str,
    requested_top: int,
    label: str,
) -> None:
    selection = expect_object(summary.get("selection"), f"{label} selection")
    require_equal(selection.get("replay_mode"), replay_mode, f"{label} replay_mode")
    require_equal(selection.get("requested_top"), requested_top, f"{label} requested_top")


def load_summary(path: Path, schema_path: Path) -> dict[str, Any]:
    document = expect_object(load_json_document(path), make_repo_relative(path))
    ensure_schema(document, schema_path, make_repo_relative(path))
    return document
