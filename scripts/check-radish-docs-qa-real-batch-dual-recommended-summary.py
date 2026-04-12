#!/usr/bin/env python3
from __future__ import annotations

import sys
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.checks.radish_docs_qa_real_batch_summary_common import (  # noqa: E402
    RECOMMENDED_SUMMARY_SCHEMA_PATH,
    SUMMARY_SCHEMA_PATH,
    build_real_batch_summary_command,
    build_real_batch_summary_paths,
    expect_object,
    load_summary,
    require_artifact_entry_state,
    require_equal,
    require_non_empty_list,
    require_summary_selection,
    run_command,
)

SAME_SAMPLE_TOP = 2
CROSS_SAMPLE_TOP = 1


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="radish-docs-qa-real-batch-dual-summary-") as temp_dir:
        paths = build_real_batch_summary_paths(
            Path(temp_dir),
            same_sample_top=SAME_SAMPLE_TOP,
            cross_sample_top=CROSS_SAMPLE_TOP,
        )
        run_command(
            build_real_batch_summary_command(
                paths,
                same_sample_top=SAME_SAMPLE_TOP,
                cross_sample_top=CROSS_SAMPLE_TOP,
            )
        )

        artifact_summary = load_summary(paths.artifact_summary_path, SUMMARY_SCHEMA_PATH)
        same_sample_summary = load_summary(paths.same_sample_summary_path, RECOMMENDED_SUMMARY_SCHEMA_PATH)
        cross_sample_summary = load_summary(paths.cross_sample_summary_path, RECOMMENDED_SUMMARY_SCHEMA_PATH)

        require_equal(artifact_summary.get("provider"), "openai-compatible", "artifact summary provider")
        execution = expect_object(artifact_summary.get("execution"), "artifact summary execution")
        require_equal(execution.get("recommended_negative_replay_exit_code"), 0, "same-sample replay exit code")
        require_equal(
            execution.get("cross_sample_recommended_negative_replay_exit_code"),
            0,
            "cross-sample replay exit code",
        )

        artifacts = expect_object(artifact_summary.get("artifacts"), "artifact summary artifacts")
        require_artifact_entry_state(
            artifacts,
            field_name="recommended_negative_replay_summary",
            expected_path=paths.same_sample_summary_path,
            expected_requested=True,
            expected_exists=True,
            label="artifact summary artifacts.recommended_negative_replay_summary",
        )
        require_artifact_entry_state(
            artifacts,
            field_name="cross_sample_recommended_negative_replay_summary",
            expected_path=paths.cross_sample_summary_path,
            expected_requested=True,
            expected_exists=True,
            label="artifact summary artifacts.cross_sample_recommended_negative_replay_summary",
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
            str(paths.artifact_summary_path),
            "same-sample summary batch_artifact_summary_path",
        )
        require_equal(
            cross_sample_summary.get("batch_artifact_summary_path"),
            str(paths.artifact_summary_path),
            "cross-sample summary batch_artifact_summary_path",
        )

        require_summary_selection(
            same_sample_summary,
            replay_mode="same_sample",
            requested_top=SAME_SAMPLE_TOP,
            label="same-sample summary",
        )
        require_summary_selection(
            cross_sample_summary,
            replay_mode="cross_sample",
            requested_top=CROSS_SAMPLE_TOP,
            label="cross-sample summary",
        )

    print("radish docs qa real batch dual recommended summary check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
