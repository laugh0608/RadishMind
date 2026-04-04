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
            "Run the Radish docs QA batch capture pipeline: "
            "inference batch -> manifest -> audit report -> replay index -> optional same-sample negatives."
        ),
    )
    parser.add_argument("--sample-dir", default="datasets/eval/radish", help="Eval sample directory. Default: datasets/eval/radish")
    parser.add_argument("--sample-pattern", default="*.json", help="Glob used with --sample-dir. Default: *.json")
    parser.add_argument("--provider", choices=["mock", "openai-compatible"], default="openai-compatible")
    parser.add_argument("--model", default="", help="Provider model name override.")
    parser.add_argument("--base-url", default="", help="Provider base URL override.")
    parser.add_argument("--api-key", default="", help="Provider API key override.")
    parser.add_argument("--temperature", type=float, default=0.0)
    parser.add_argument("--output-root", required=True, help="Batch output root.")
    parser.add_argument("--collection-batch", required=True, help="Batch collection name.")
    parser.add_argument("--capture-origin", default="", help="Optional capture origin override.")
    parser.add_argument("--capture-tag", action="append", default=[], help="Additional capture tag. Can be repeated.")
    parser.add_argument("--capture-notes", default="", help="Optional capture notes.")
    parser.add_argument("--max-attempts", type=int, default=3, help="Maximum attempts for retryable provider failures.")
    parser.add_argument(
        "--retry-base-delay-seconds",
        type=float,
        default=5.0,
        help="Base delay in seconds before retrying retryable provider failures.",
    )
    parser.add_argument("--manifest-output", default="", help="Optional manifest output path override.")
    parser.add_argument("--manifest-description", default="", help="Optional manifest description.")
    parser.add_argument("--resume", action="store_true", help="Skip already captured samples when batch outputs exist.")
    parser.add_argument(
        "--continue-on-error",
        action="store_true",
        help="Continue later samples after retry exhaustion and still attempt audit if a manifest is produced.",
    )
    parser.add_argument("--skip-audit", action="store_true", help="Only run inference batch capture; skip audit and replay governance.")
    parser.add_argument("--audit-report-output", default="", help="Optional audit report output path override.")
    parser.add_argument("--replay-index-output", default="", help="Optional replay index output path override.")
    parser.add_argument(
        "--build-negative-replay",
        action="store_true",
        help="After audit, also build same-sample replay negatives. Recommended for real provider batches.",
    )
    parser.add_argument("--negative-output-dir", default="", help="Optional output directory override for generated negatives.")
    parser.add_argument(
        "--fail-on-audit-violation",
        action="store_true",
        help="Fail the pipeline if audit finds violations.",
    )
    return parser.parse_args()


def run_command(command: list[str]) -> subprocess.CompletedProcess[str]:
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
    return result


def resolve_repo_relative(path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = (REPO_ROOT / path).resolve()
    return path


def main() -> int:
    args = parse_args()
    output_root = resolve_repo_relative(args.output_root)
    manifest_path = (
        resolve_repo_relative(args.manifest_output)
        if args.manifest_output.strip()
        else output_root / f"{args.collection_batch}.manifest.json"
    )
    audit_report_path = (
        resolve_repo_relative(args.audit_report_output)
        if args.audit_report_output.strip()
        else output_root / f"{args.collection_batch}.audit.json"
    )
    replay_index_path = (
        resolve_repo_relative(args.replay_index_output)
        if args.replay_index_output.strip()
        else output_root / f"{args.collection_batch}.negative-replay-index.json"
    )

    inference_command = [
        sys.executable,
        str(REPO_ROOT / "scripts/run-copilot-inference.py"),
        "--sample-dir",
        args.sample_dir,
        "--sample-pattern",
        args.sample_pattern,
        "--provider",
        args.provider,
        "--temperature",
        str(args.temperature),
        "--output-root",
        str(output_root),
        "--collection-batch",
        args.collection_batch,
        "--manifest-output",
        str(manifest_path),
    ]
    if args.model.strip():
        inference_command.extend(["--model", args.model.strip()])
    if args.base_url.strip():
        inference_command.extend(["--base-url", args.base_url.strip()])
    if args.api_key.strip():
        inference_command.extend(["--api-key", args.api_key.strip()])
    if args.capture_origin.strip():
        inference_command.extend(["--capture-origin", args.capture_origin.strip()])
    for tag in args.capture_tag:
        if str(tag).strip():
            inference_command.extend(["--capture-tag", str(tag).strip()])
    if args.capture_notes.strip():
        inference_command.extend(["--capture-notes", args.capture_notes.strip()])
    if args.max_attempts != 3:
        inference_command.extend(["--max-attempts", str(args.max_attempts)])
    if args.retry_base_delay_seconds != 5.0:
        inference_command.extend(["--retry-base-delay-seconds", str(args.retry_base_delay_seconds)])
    if args.manifest_description.strip():
        inference_command.extend(["--manifest-description", args.manifest_description.strip()])
    if args.resume:
        inference_command.append("--resume")
    if args.continue_on_error:
        inference_command.append("--continue-on-error")

    inference_result = run_command(inference_command)

    if args.skip_audit:
        return inference_result.returncode

    if not manifest_path.is_file():
        if inference_result.returncode != 0:
            return inference_result.returncode
        raise SystemExit(f"manifest was not produced: {manifest_path}")

    if inference_result.returncode != 0:
        print(
            "inference batch exited non-zero but produced a manifest; continuing with audit on captured records",
            file=sys.stderr,
        )

    audit_command = [
        sys.executable,
        str(REPO_ROOT / "scripts/audit-candidate-record-batch.py"),
        "radish-docs-qa",
        "--manifest",
        str(manifest_path),
        "--report-output",
        str(audit_report_path),
        "--replay-index-output",
        str(replay_index_path),
    ]
    if args.build_negative_replay:
        audit_command.append("--build-negative-replay")
    if args.negative_output_dir.strip():
        audit_command.extend(["--negative-output-dir", args.negative_output_dir.strip()])
    if args.fail_on_audit_violation:
        audit_command.append("--fail-on-violation")

    audit_result = run_command(audit_command)
    if inference_result.returncode != 0:
        return inference_result.returncode
    return audit_result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
