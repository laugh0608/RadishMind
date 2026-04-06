#!/usr/bin/env python3
from __future__ import annotations

import argparse
import shlex
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
RECOMMENDED_NEGATIVE_REPLAY_SUMMARY_SCHEMA_PATH = (
    REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json"
)
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
        "--cross-sample-negative-sample-dir",
        default="",
        help="Optional negative sample directory used to build a cross-sample replay index after audit.",
    )
    parser.add_argument(
        "--cross-sample-replay-index-output",
        default="",
        help="Optional cross-sample replay index output path override.",
    )
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
        "--build-recommended-negative-replay-summary",
        action="store_true",
        help="After writing artifacts.json, also replay recommended negative groups and write a structured summary.",
    )
    parser.add_argument(
        "--recommended-groups-top",
        type=int,
        default=0,
        help="Optional top N recommended groups to replay. Default: 0 (replay all recommended groups).",
    )
    parser.add_argument(
        "--cross-sample-recommended-groups-top",
        type=int,
        default=0,
        help="Optional top N cross-sample recommended groups to replay. Default: 0 (replay all cross-sample groups).",
    )
    parser.add_argument(
        "--recommended-replay-mode",
        choices=["same_sample", "cross_sample"],
        default="",
        help="Optional replay mode override for the recommended negative replay summary.",
    )
    parser.add_argument(
        "--recommended-summary-output",
        default="",
        help="Optional recommended negative replay summary output path override.",
    )
    parser.add_argument(
        "--cross-sample-recommended-summary-output",
        default="",
        help="Optional cross-sample recommended negative replay summary output path override.",
    )
    parser.add_argument(
        "--fail-on-recommended-replay-violation",
        action="store_true",
        help="Fail the pipeline if the recommended negative replay summary finds violations.",
    )
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


def count_replay_entries(index_document: dict[str, Any], replay_mode: str) -> tuple[int, int]:
    linked_replay_count = 0
    for group in index_document.get("violation_groups") or []:
        if not isinstance(group, dict):
            raise SystemExit("negative replay index violation_groups entries must be json objects")
        for entry in group.get("entries") or []:
            if not isinstance(entry, dict):
                raise SystemExit("negative replay index entries must be json objects")
            if str(entry.get("replay_mode") or "").strip() == replay_mode:
                linked_replay_count += 1
    unlinked_replay_count = len(list(index_document.get("unlinked_failed_samples") or []))
    return linked_replay_count, unlinked_replay_count


def derive_recommended_summary_output_path(artifact_summary_path: Path, top_n: int, replay_mode: str) -> Path:
    name = artifact_summary_path.name
    if name.endswith(".artifacts.json"):
        base_name = name[: -len(".artifacts.json")]
    elif name.endswith(".json"):
        base_name = name[: -len(".json")]
    else:
        base_name = name
    return artifact_summary_path.with_name(
        f"{base_name}.recommended-negative-replay-top{top_n}-{replay_mode}.summary.json"
    )


def load_recommended_summary_if_exists(path: Path) -> dict[str, Any] | None:
    summary_document = load_object_if_exists(path)
    if summary_document is not None:
        ensure_schema(
            summary_document,
            RECOMMENDED_NEGATIVE_REPLAY_SUMMARY_SCHEMA_PATH,
            make_repo_relative(path),
        )
    return summary_document


def extract_recommended_summary_metrics(summary_document: dict[str, Any] | None) -> dict[str, int]:
    summary = (summary_document or {}).get("summary") or {}
    if not isinstance(summary, dict):
        raise SystemExit("recommended negative replay summary summary must be a json object")
    return {
        "group_count": int(summary.get("selected_group_count") or 0),
        "passed_group_count": int(summary.get("passed_group_count") or 0),
        "failed_group_count": int(summary.get("failed_group_count") or 0),
        "sample_count": int(summary.get("selected_sample_count") or 0),
        "violation_count": int(summary.get("violation_count") or 0),
    }


def get_recommended_group_ids(recommended_negative_replays: dict[str, Any], replay_mode: str) -> list[str]:
    field_name = "cross_sample_recommended_group_ids" if replay_mode == "cross_sample" else "recommended_group_ids"
    values = recommended_negative_replays.get(field_name) or []
    if not isinstance(values, list):
        raise SystemExit(f"recommended_negative_replays.{field_name} must be a json array")
    return [str(value).strip() for value in values if str(value).strip()]


def select_recommended_top(group_ids: list[str], requested_top: int) -> int:
    return requested_top or len(group_ids)


def build_recommended_negative_replays(
    *,
    artifact_summary_path: Path,
    replay_index: dict[str, Any] | None,
    cross_sample_replay_index: dict[str, Any] | None,
) -> dict[str, Any]:
    summary_path_value = make_repo_relative(artifact_summary_path)
    def build_groups(index_document: dict[str, Any] | None, replay_mode: str) -> list[dict[str, Any]]:
        if index_document is None:
            return []
        groups: list[dict[str, Any]] = []
        for group in index_document.get("violation_groups") or []:
            if not isinstance(group, dict):
                raise SystemExit("negative replay index violation_groups entries must be json objects")
            group_id = str(group.get("group_id") or "").strip()
            if not group_id:
                raise SystemExit("negative replay index violation_groups entries must include group_id")
            entries = group.get("entries") or []
            entry_count = 0
            for entry in entries:
                if not isinstance(entry, dict):
                    raise SystemExit(f"negative replay index group '{group_id}' entries must be json objects")
                if str(entry.get("replay_mode") or "").strip() == replay_mode:
                    entry_count += 1
            if entry_count == 0:
                continue

            bash_command = " ".join(
                [
                    "bash",
                    shlex.quote("./scripts/run-radish-docs-qa-negative-regression.sh"),
                    "--batch-artifact-summary",
                    shlex.quote(summary_path_value),
                    "--group-id",
                    shlex.quote(group_id),
                    "--replay-mode",
                    replay_mode,
                    "--fail-on-violation",
                ]
            )
            python_command = " ".join(
                [
                    "python3",
                    shlex.quote("./scripts/run-eval-regression.py"),
                    "radish-docs-qa-negative",
                    "--batch-artifact-summary",
                    shlex.quote(summary_path_value),
                    "--group-id",
                    shlex.quote(group_id),
                    "--replay-mode",
                    replay_mode,
                    "--fail-on-violation",
                ]
            )
            group_document = {
                "group_id": group_id,
                "replay_mode": replay_mode,
                "entry_count": entry_count,
                "expected_candidate_violations": list(group.get("expected_candidate_violations") or []),
                "bash_command": bash_command,
                "python_command": python_command,
            }
            if replay_mode == "same_sample":
                group_document["same_sample_entry_count"] = entry_count
            else:
                group_document["cross_sample_entry_count"] = entry_count
            groups.append(group_document)
        return sorted(groups, key=lambda item: (-int(item["entry_count"]), item["group_id"]))

    same_sample_groups = build_groups(replay_index, "same_sample")
    cross_sample_groups = build_groups(cross_sample_replay_index, "cross_sample")
    document = {
        "default_replay_mode": "same_sample",
        "recommended_group_ids": [group["group_id"] for group in same_sample_groups],
        "groups": same_sample_groups,
    }
    if cross_sample_groups:
        document["cross_sample_recommended_group_ids"] = [group["group_id"] for group in cross_sample_groups]
        document["cross_sample_groups"] = cross_sample_groups
    return document


def build_artifact_summary_document(
    *,
    args: argparse.Namespace,
    output_root: Path,
    manifest_path: Path,
    audit_report_path: Path,
    replay_index_path: Path,
    cross_sample_replay_index_path: Path,
    artifact_summary_path: Path,
    inference_exit_code: int,
    audit_exit_code: int | None,
    recommended_negative_replay_summary_path: Path,
    recommended_negative_replay_requested: bool,
    recommended_negative_replay_exit_code: int | None,
    cross_sample_recommended_negative_replay_summary_path: Path,
    cross_sample_recommended_negative_replay_requested: bool,
    cross_sample_recommended_negative_replay_exit_code: int | None,
) -> dict[str, Any]:
    manifest = load_json_document(manifest_path)
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))

    if not isinstance(manifest, dict):
        raise SystemExit(f"{make_repo_relative(manifest_path)} must be a json object")

    audit_report = load_object_if_exists(audit_report_path)
    replay_index = load_object_if_exists(replay_index_path)
    cross_sample_replay_index = load_object_if_exists(cross_sample_replay_index_path)
    recommended_negative_replay_summary = load_recommended_summary_if_exists(recommended_negative_replay_summary_path)
    cross_sample_recommended_negative_replay_summary = load_recommended_summary_if_exists(
        cross_sample_recommended_negative_replay_summary_path
    )
    recommended_summary_metrics = extract_recommended_summary_metrics(recommended_negative_replay_summary)
    cross_sample_recommended_summary_metrics = extract_recommended_summary_metrics(
        cross_sample_recommended_negative_replay_summary
    )

    linked_same_sample_count = 0
    unlinked_same_sample_count = 0
    violation_group_count = 0
    if replay_index is not None:
        linked_same_sample_count, unlinked_same_sample_count = count_replay_entries(replay_index, "same_sample")
        violation_group_count = len(list(replay_index.get("violation_groups") or []))

    linked_cross_sample_count = 0
    unlinked_cross_sample_count = 0
    cross_sample_violation_group_count = 0
    if cross_sample_replay_index is not None:
        linked_cross_sample_count, unlinked_cross_sample_count = count_replay_entries(
            cross_sample_replay_index, "cross_sample"
        )
        cross_sample_violation_group_count = len(list(cross_sample_replay_index.get("violation_groups") or []))

    negative_output_dir = (
        resolve_repo_relative(args.negative_output_dir)
        if args.negative_output_dir.strip()
        else resolve_repo_relative(
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
            "recommended_negative_replay_exit_code": recommended_negative_replay_exit_code,
            "cross_sample_recommended_negative_replay_exit_code": cross_sample_recommended_negative_replay_exit_code,
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
            "cross_sample_negative_replay_index": {
                "requested": bool(args.cross_sample_negative_sample_dir.strip()),
                "path": make_repo_relative(cross_sample_replay_index_path),
                "exists": cross_sample_replay_index_path.is_file(),
            },
            "same_sample_negative_replay": {
                "requested": args.build_negative_replay,
                "output_dir": make_repo_relative(negative_output_dir),
                "expected_fixture_count": linked_same_sample_count + unlinked_same_sample_count,
            },
            "recommended_negative_replay_summary": {
                "requested": recommended_negative_replay_requested,
                "path": make_repo_relative(recommended_negative_replay_summary_path),
                "exists": recommended_negative_replay_summary_path.is_file(),
            },
            "cross_sample_recommended_negative_replay_summary": {
                "requested": cross_sample_recommended_negative_replay_requested,
                "path": make_repo_relative(cross_sample_recommended_negative_replay_summary_path),
                "exists": cross_sample_recommended_negative_replay_summary_path.is_file(),
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
            "cross_sample_violation_group_count": cross_sample_violation_group_count,
            "linked_cross_sample_negative_count": linked_cross_sample_count,
            "unlinked_cross_sample_negative_count": unlinked_cross_sample_count,
            "recommended_replay_group_count": recommended_summary_metrics["group_count"],
            "recommended_replay_passed_group_count": recommended_summary_metrics["passed_group_count"],
            "recommended_replay_failed_group_count": recommended_summary_metrics["failed_group_count"],
            "recommended_replay_sample_count": recommended_summary_metrics["sample_count"],
            "recommended_replay_violation_count": recommended_summary_metrics["violation_count"],
            "cross_sample_recommended_replay_group_count": cross_sample_recommended_summary_metrics["group_count"],
            "cross_sample_recommended_replay_passed_group_count": cross_sample_recommended_summary_metrics["passed_group_count"],
            "cross_sample_recommended_replay_failed_group_count": cross_sample_recommended_summary_metrics["failed_group_count"],
            "cross_sample_recommended_replay_sample_count": cross_sample_recommended_summary_metrics["sample_count"],
            "cross_sample_recommended_replay_violation_count": cross_sample_recommended_summary_metrics["violation_count"],
        },
        "recommended_negative_replays": build_recommended_negative_replays(
            artifact_summary_path=artifact_summary_path,
            replay_index=replay_index,
            cross_sample_replay_index=cross_sample_replay_index,
        ),
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
    if args.recommended_groups_top < 0:
        raise SystemExit("--recommended-groups-top must be greater than or equal to 0")
    if args.cross_sample_recommended_groups_top < 0:
        raise SystemExit("--cross-sample-recommended-groups-top must be greater than or equal to 0")
    if args.build_recommended_negative_replay_summary and args.skip_audit:
        raise SystemExit("--build-recommended-negative-replay-summary cannot be used together with --skip-audit")
    if args.build_recommended_negative_replay_summary and not args.build_negative_replay:
        raise SystemExit("--build-recommended-negative-replay-summary requires --build-negative-replay")

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
    cross_sample_replay_index_path = (
        resolve_repo_relative(args.cross_sample_replay_index_output)
        if args.cross_sample_replay_index_output.strip()
        else output_root / f"{args.collection_batch}.cross-sample-replay-index.json"
    )
    artifact_summary_path = (
        resolve_repo_relative(args.artifact_summary_output)
        if args.artifact_summary_output.strip()
        else output_root / f"{args.collection_batch}.artifacts.json"
    )
    explicit_recommended_replay_mode = args.recommended_replay_mode.strip()
    recommended_negative_replay_summary_requested = False
    recommended_negative_replay_exit_code: int | None = None
    recommended_negative_replay_summary_path = (
        resolve_repo_relative(args.recommended_summary_output)
        if args.recommended_summary_output.strip() and explicit_recommended_replay_mode != "cross_sample"
        else artifact_summary_path.parent / "pending.recommended-negative-replay.summary.json"
    )
    cross_sample_recommended_negative_replay_summary_requested = False
    cross_sample_recommended_negative_replay_exit_code: int | None = None
    cross_sample_recommended_negative_replay_summary_path = (
        resolve_repo_relative(args.cross_sample_recommended_summary_output)
        if args.cross_sample_recommended_summary_output.strip()
        else resolve_repo_relative(args.recommended_summary_output)
        if args.recommended_summary_output.strip() and explicit_recommended_replay_mode == "cross_sample"
        else artifact_summary_path.parent / "pending.cross-sample-recommended-negative-replay.summary.json"
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

        if args.cross_sample_negative_sample_dir.strip():
            cross_sample_index_command = [
                sys.executable,
                str(REPO_ROOT / "scripts/build-negative-replay-index.py"),
                "--audit-report",
                str(audit_report_path),
                "--negative-sample-dir",
                args.cross_sample_negative_sample_dir.strip(),
                "--output",
                str(cross_sample_replay_index_path),
            ]
            run_command(cross_sample_index_command)

    summary_document = build_artifact_summary_document(
        args=args,
        output_root=output_root,
        manifest_path=manifest_path,
        audit_report_path=audit_report_path,
        replay_index_path=replay_index_path,
        cross_sample_replay_index_path=cross_sample_replay_index_path,
        artifact_summary_path=artifact_summary_path,
        inference_exit_code=inference_result.returncode,
        audit_exit_code=audit_exit_code,
        recommended_negative_replay_summary_path=recommended_negative_replay_summary_path,
        recommended_negative_replay_requested=recommended_negative_replay_summary_requested,
        recommended_negative_replay_exit_code=recommended_negative_replay_exit_code,
        cross_sample_recommended_negative_replay_summary_path=cross_sample_recommended_negative_replay_summary_path,
        cross_sample_recommended_negative_replay_requested=cross_sample_recommended_negative_replay_summary_requested,
        cross_sample_recommended_negative_replay_exit_code=cross_sample_recommended_negative_replay_exit_code,
    )
    write_json_document(artifact_summary_path, summary_document)
    print(f"wrote artifact summary to {make_repo_relative(artifact_summary_path)}", file=sys.stderr)

    if args.build_recommended_negative_replay_summary:
        recommended_negative_replays = summary_document.get("recommended_negative_replays") or {}
        if not isinstance(recommended_negative_replays, dict):
            raise SystemExit("recommended_negative_replays must be a json object")

        replay_modes_to_generate: list[str]
        if explicit_recommended_replay_mode:
            replay_modes_to_generate = [explicit_recommended_replay_mode]
        else:
            replay_modes_to_generate = ["same_sample"]
            if get_recommended_group_ids(recommended_negative_replays, "cross_sample"):
                replay_modes_to_generate.append("cross_sample")

        for replay_mode in replay_modes_to_generate:
            group_ids = get_recommended_group_ids(recommended_negative_replays, replay_mode)
            if not group_ids:
                print(
                    f"artifact summary does not contain {replay_mode} recommended negative replay groups; skipping",
                    file=sys.stderr,
                )
                continue

            requested_top = (
                args.cross_sample_recommended_groups_top
                if replay_mode == "cross_sample"
                else args.recommended_groups_top
            )
            recommended_top = select_recommended_top(group_ids, requested_top)
            summary_output_override = (
                args.cross_sample_recommended_summary_output.strip()
                if replay_mode == "cross_sample"
                else args.recommended_summary_output.strip()
            )
            summary_output_path = (
                resolve_repo_relative(summary_output_override)
                if summary_output_override
                else derive_recommended_summary_output_path(artifact_summary_path, recommended_top, replay_mode)
            )

            recommended_command = [
                sys.executable,
                str(REPO_ROOT / "scripts/run-radish-docs-qa-negative-recommended.py"),
                "--batch-artifact-summary",
                make_repo_relative(artifact_summary_path),
                "--top",
                str(recommended_top),
                "--replay-mode",
                replay_mode,
                "--summary-output",
                make_repo_relative(summary_output_path),
            ]
            if args.fail_on_recommended_replay_violation:
                recommended_command.append("--fail-on-violation")

            recommended_result = run_command(recommended_command)
            if replay_mode == "cross_sample":
                cross_sample_recommended_negative_replay_summary_requested = True
                cross_sample_recommended_negative_replay_summary_path = summary_output_path
                cross_sample_recommended_negative_replay_exit_code = recommended_result.returncode
            else:
                recommended_negative_replay_summary_requested = True
                recommended_negative_replay_summary_path = summary_output_path
                recommended_negative_replay_exit_code = recommended_result.returncode

        summary_document = build_artifact_summary_document(
            args=args,
            output_root=output_root,
            manifest_path=manifest_path,
            audit_report_path=audit_report_path,
            replay_index_path=replay_index_path,
            cross_sample_replay_index_path=cross_sample_replay_index_path,
            artifact_summary_path=artifact_summary_path,
            inference_exit_code=inference_result.returncode,
            audit_exit_code=audit_exit_code,
            recommended_negative_replay_summary_path=recommended_negative_replay_summary_path,
            recommended_negative_replay_requested=recommended_negative_replay_summary_requested,
            recommended_negative_replay_exit_code=recommended_negative_replay_exit_code,
            cross_sample_recommended_negative_replay_summary_path=cross_sample_recommended_negative_replay_summary_path,
            cross_sample_recommended_negative_replay_requested=cross_sample_recommended_negative_replay_summary_requested,
            cross_sample_recommended_negative_replay_exit_code=cross_sample_recommended_negative_replay_exit_code,
        )
        write_json_document(artifact_summary_path, summary_document)
        print(
            f"updated artifact summary with recommended replay summary metadata: {make_repo_relative(artifact_summary_path)}",
            file=sys.stderr,
        )

    if inference_result.returncode != 0:
        return inference_result.returncode
    if args.skip_audit:
        return 0
    if audit_exit_code:
        return audit_exit_code
    return recommended_negative_replay_exit_code or cross_sample_recommended_negative_replay_exit_code or 0


if __name__ == "__main__":
    raise SystemExit(main())
