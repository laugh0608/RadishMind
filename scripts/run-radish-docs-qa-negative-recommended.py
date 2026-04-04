#!/usr/bin/env python3
from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run the recommended Radish docs QA negative replay groups from a batch artifact summary. "
            "This is a thin wrapper over run-eval-regression.py radish-docs-qa-negative."
        ),
    )
    parser.add_argument("--batch-artifact-summary", required=True, help="Batch artifact summary json path.")
    parser.add_argument("--top", type=int, default=1, help="Replay the top N recommended groups. Default: 1")
    parser.add_argument("--sample-dir", default="", help="Optional negative sample directory override.")
    parser.add_argument(
        "--replay-mode",
        choices=["same_sample", "cross_sample"],
        default="",
        help="Optional replay mode override. Defaults to the batch artifact summary recommendation.",
    )
    parser.add_argument(
        "--fail-on-violation",
        action="store_true",
        help="Fail if the negative replay regression finds violations.",
    )
    return parser.parse_args()


def run_command(command: list[str]) -> int:
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
    return result.returncode


def main() -> int:
    args = parse_args()
    if args.top <= 0:
        raise SystemExit("--top must be greater than 0")

    command = [
        sys.executable,
        str(REPO_ROOT / "scripts" / "run-eval-regression.py"),
        "radish-docs-qa-negative",
        "--batch-artifact-summary",
        args.batch_artifact_summary,
        "--recommended-groups-top",
        str(args.top),
    ]
    if args.sample_dir.strip():
        command.extend(["--sample-dir", args.sample_dir.strip()])
    if args.replay_mode.strip():
        command.extend(["--replay-mode", args.replay_mode.strip()])
    if args.fail_on_violation:
        command.append("--fail-on-violation")
    return run_command(command)


if __name__ == "__main__":
    raise SystemExit(main())
