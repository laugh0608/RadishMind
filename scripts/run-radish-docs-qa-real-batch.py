#!/usr/bin/env python3
from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    RECORD_BATCH_SCHEMA_PATH,
    ensure_schema,
    load_json_document,
    make_repo_relative,
    write_json_document,
)

SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json"
DEFAULT_NEGATIVE_OUTPUT_DIR = "datasets/eval/radish-negative"


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
        "--artifact-summary-output",
        default="",
        help="Optional structured artifact summary output path override.",
    )
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


def load_object_if_exists(path: Path) -> dict[str, Any] | None:
    if not path.is_file():
        return None
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"{make_repo_relative(path)} must be a json object")
    return document


def count_same_sample_entries(index_document: dict[str, Any]) -> tuple[int, int]:
    linked_same_sample_count = 0
    for group in index_document.get("violation_groups") or []:
        if not isinstance(group, dict):
            raise SystemExit("negative replay index violation_groups entries must be json objects")
        for entry in group.get("entries") or []:
            if not isinstance(entry, dict):
                raise SystemExit("negative replay index entries must be json objects")
            if str(entry.get("replay_mode") or "").strip() == "same_sample":
                linked_same_sample_count += 1
    unlinked_same_sample_count = len(list(index_document.get("unlinked_failed_samples") or []))
    return linked_same_sample_count, unlinked_same_sample_count


def build_artifact_summary_document(
    *,
    args: argparse.Namespace,
    output_root: Path,
    manifest_path: Path,
    audit_report_path: Path,
    replay_index_path: Path,
    inference_exit_code: int,
    audit_exit_code: int | None,
) -> dict[str, Any]:
    manifest = load_json_document(manifest_path)
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))

    if not isinstance(manifest, dict):
        raise SystemExit(f"{make_repo_relative(manifest_path)} must be a json object")

    audit_report = load_object_if_exists(audit_report_path)
    replay_index = load_object_if_exists(replay_index_path)

    linked_same_sample_count = 0
    unlinked_same_sample_count = 0
    violation_group_count = 0
    if replay_index is not None:
        linked_same_sample_count, unlinked_same_sample_count = count_same_sample_entries(replay_index)
        violation_group_count = len(list(replay_index.get("violation_groups") or []))

    negative_output_dir = (
        resolve_relative_to_repo(args.negative_output_dir)
        if args.negative_output_dir.strip()
        else resolve_relative_to_repo(
            str((replay_index or {}).get("negative_sample_dir") or DEFAULT_NEGATIVE_OUTPUT_DIR)
        )
    )

    summary_document: dict[str, Any] = {
        "schema_version": 1,
        "pipeline": "radish-docs-qa-real-batch",
        "project": str(manifest.get("project") or "").strip(),
        "task": str(manifest.get("task") or "").strip(),
        "eval_task": "radish-docs-qa",
        "provider": args.provider,
        "collection_batch": str(manifest.get("collection_batch") or "").strip(),
        "source": str(manifest.get("source") or "").strip(),
        "output_root": make_repo_relative(output_root),
        "execution": {
            "inference_exit_code": inference_exit_code,
            "audit_exit_code": audit_exit_code,
        },
        "artifacts": {
            "manifest": {
                "path": make_repo_relative(manifest_path),
                "exists": manifest_path.is_file(),
                "record_count": len(list(manifest.get("records") or [])),
            },
            "audit_report": {
                "requested": not args.skip_audit,
                "path": make_repo_relative(audit_report_path),
                "exists": audit_report_path.is_file(),
            },
            "negative_replay_index": {
                "requested": not args.skip_audit,
                "path": make_repo_relative(replay_index_path),
                "exists": replay_index_path.is_file(),
            },
            "same_sample_negative_replay": {
                "requested": args.build_negative_replay,
                "output_dir": make_repo_relative(negative_output_dir),
                "expected_fixture_count": linked_same_sample_count + unlinked_same_sample_count,
            },
        },
        "summary": {
            "captured_record_count": len(list(manifest.get("records") or [])),
            "matched_sample_count": int((audit_report or {}).get("matched_sample_count") or 0),
            "passed_sample_count": int((audit_report or {}).get("passed_count") or 0),
            "failed_sample_count": int((audit_report or {}).get("failed_count") or 0),
            "violation_count": int((audit_report or {}).get("violation_count") or 0),
            "violation_group_count": violation_group_count,
            "linked_same_sample_negative_count": linked_same_sample_count,
            "unlinked_same_sample_negative_count": unlinked_same_sample_count,
            "expected_same_sample_negative_count": linked_same_sample_count + unlinked_same_sample_count,
        },
    }

    model = args.model.strip()
    if model:
        summary_document["model"] = model

    capture_origin = str(manifest.get("capture_origin") or "").strip()
    if capture_origin:
        summary_document["capture_origin"] = capture_origin

    ensure_schema(summary_document, SUMMARY_SCHEMA_PATH, "batch orchestration summary")
    return summary_document


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
    artifact_summary_path = (
        resolve_repo_relative(args.artifact_summary_output)
        if args.artifact_summary_output.strip()
        else output_root / f"{args.collection_batch}.artifacts.json"
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

    if not manifest_path.is_file():
        if inference_result.returncode != 0:
            return inference_result.returncode
        raise SystemExit(f"manifest was not produced: {manifest_path}")

    audit_exit_code: int | None = None

    if not args.skip_audit:
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
        audit_exit_code = audit_result.returncode

    summary_document = build_artifact_summary_document(
        args=args,
        output_root=output_root,
        manifest_path=manifest_path,
        audit_report_path=audit_report_path,
        replay_index_path=replay_index_path,
        inference_exit_code=inference_result.returncode,
        audit_exit_code=audit_exit_code,
    )
    write_json_document(artifact_summary_path, summary_document)
    print(f"wrote artifact summary to {make_repo_relative(artifact_summary_path)}", file=sys.stderr)

    if inference_result.returncode != 0:
        return inference_result.returncode
    if args.skip_audit:
        return 0
    return audit_exit_code or 0


if __name__ == "__main__":
    raise SystemExit(main())
