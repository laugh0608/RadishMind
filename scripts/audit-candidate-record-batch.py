#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
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
    resolve_relative_to_repo,
)
from services.runtime.eval_regression import parse_regression_output  # noqa: E402


TASK_SAMPLE_DIRS = {
    "radish-docs-qa": REPO_ROOT / "datasets/eval/radish",
    "radishflow-suggest-edits": REPO_ROOT / "datasets/eval/radishflow",
    "radishflow-ghost-completion": REPO_ROOT / "datasets/eval/radishflow",
}
DEFAULT_NEGATIVE_OUTPUT_DIR = "datasets/eval/radish-negative"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Audit a candidate record batch by replaying it through the existing eval regression rules.",
    )
    parser.add_argument("task", choices=sorted(TASK_SAMPLE_DIRS), help="Eval task name to audit against.")
    parser.add_argument("--manifest", required=True, help="Candidate record batch manifest path.")
    parser.add_argument("--sample-dir", default="", help="Optional sample directory override.")
    parser.add_argument(
        "--expected-source",
        default="",
        help="Optional candidate_response_record.expected_source override. Defaults to manifest.source when present.",
    )
    parser.add_argument(
        "--required-capture-origin",
        default="",
        help="Optional candidate_response_record.required_capture_origin override. Defaults to manifest.capture_origin when present.",
    )
    parser.add_argument(
        "--required-collection-batch",
        default="",
        help="Optional candidate_response_record.required_collection_batch override. Defaults to manifest.collection_batch.",
    )
    parser.add_argument(
        "--required-tag",
        action="append",
        default=[],
        help="Optional candidate_response_record.required_tags entry. Can be repeated.",
    )
    parser.add_argument(
        "--report-output",
        default="",
        help="Optional path to write a structured audit report json.",
    )
    parser.add_argument(
        "--replay-index-output",
        default="",
        help="Optional path to write a negative replay index json after the audit report is generated.",
    )
    parser.add_argument(
        "--build-negative-replay",
        action="store_true",
        help="After writing audit metadata, also build same-sample negative replay fixtures from the replay index.",
    )
    parser.add_argument(
        "--negative-output-dir",
        default="",
        help="Optional output directory override used with --build-negative-replay.",
    )
    parser.add_argument("--fail-on-violation", action="store_true")
    return parser.parse_args()


def normalize_required_tags(values: list[str]) -> list[str]:
    normalized: list[str] = []
    for value in values:
        tag = str(value).strip()
        if tag and tag not in normalized:
            normalized.append(tag)
    return normalized


def load_manifest(manifest_path: Path) -> dict[str, Any]:
    manifest = load_json_document(manifest_path)
    if not isinstance(manifest, dict):
        raise SystemExit(f"manifest must be a json object: {make_repo_relative(manifest_path)}")
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))
    return manifest


def build_manifest_sample_map(manifest: dict[str, Any], manifest_label: str) -> dict[str, str]:
    sample_map: dict[str, str] = {}
    for entry in manifest.get("records") or []:
        if not isinstance(entry, dict):
            continue
        sample_id = str(entry.get("sample_id") or "").strip()
        record_id = str(entry.get("record_id") or "").strip()
        if not sample_id or not record_id:
            raise SystemExit(f"{manifest_label}: each record must include non-empty sample_id and record_id")
        if sample_id in sample_map:
            raise SystemExit(f"{manifest_label}: duplicate sample_id found in manifest: {sample_id}")
        sample_map[sample_id] = record_id
    if not sample_map:
        raise SystemExit(f"{manifest_label}: manifest does not contain any records")
    return sample_map


def resolve_sample_dir(task: str, sample_dir_value: str) -> Path:
    if sample_dir_value.strip():
        return resolve_relative_to_repo(sample_dir_value)
    return TASK_SAMPLE_DIRS[task]


def make_override_record_ref(
    *,
    manifest_path: Path,
    record_id: str,
    expected_source: str,
    required_capture_origin: str,
    required_collection_batch: str,
    required_tags: list[str],
) -> dict[str, Any]:
    record_ref: dict[str, Any] = {
        "manifest_path": make_repo_relative(manifest_path),
        "record_id": record_id,
    }
    if expected_source:
        record_ref["expected_source"] = expected_source
    if required_capture_origin:
        record_ref["required_capture_origin"] = required_capture_origin
    if required_collection_batch:
        record_ref["required_collection_batch"] = required_collection_batch
    if required_tags:
        record_ref["required_tags"] = required_tags
    return record_ref


def derive_negative_replay_index_output(report_path: Path) -> Path:
    name = report_path.name
    if name.endswith(".audit.json"):
        return report_path.with_name(name[: -len(".audit.json")] + ".negative-replay-index.json")
    if name.endswith(".json"):
        return report_path.with_name(name[: -len(".json")] + ".negative-replay-index.json")
    return report_path.with_name(name + ".negative-replay-index.json")


def run_repo_script(script_name: str, script_args: list[str]) -> None:
    command = [sys.executable, str(REPO_ROOT / "scripts" / script_name), *script_args]
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


def build_negative_replay_index_args(report_path: Path, replay_index_output: Path, negative_sample_dir: str) -> list[str]:
    script_args = [
        "--audit-report",
        make_repo_relative(report_path),
        "--output",
        make_repo_relative(replay_index_output),
    ]
    if negative_sample_dir.strip():
        script_args.extend(["--negative-sample-dir", negative_sample_dir.strip()])
    return script_args


def resolve_same_sample_negative_output_dir(args: argparse.Namespace) -> str:
    return args.negative_output_dir.strip() or DEFAULT_NEGATIVE_OUTPUT_DIR


def main() -> int:
    args = parse_args()
    if (args.replay_index_output.strip() or args.build_negative_replay) and not args.report_output.strip():
        raise SystemExit("--report-output is required when using --replay-index-output or --build-negative-replay")
    if args.build_negative_replay and args.task != "radish-docs-qa":
        raise SystemExit("--build-negative-replay is currently only supported for task 'radish-docs-qa'")

    manifest_path = resolve_relative_to_repo(args.manifest)
    if not manifest_path.is_file():
        raise SystemExit(f"manifest file not found: {args.manifest}")

    manifest = load_manifest(manifest_path)
    manifest_label = make_repo_relative(manifest_path)
    sample_to_record_id = build_manifest_sample_map(manifest, manifest_label)
    sample_dir = resolve_sample_dir(args.task, args.sample_dir)

    expected_source = args.expected_source.strip() or str(manifest.get("source") or "").strip()
    required_capture_origin = args.required_capture_origin.strip() or str(manifest.get("capture_origin") or "").strip()
    required_collection_batch = (
        args.required_collection_batch.strip() or str(manifest.get("collection_batch") or "").strip()
    )
    required_tags = normalize_required_tags(args.required_tag)
    report_path = resolve_relative_to_repo(args.report_output) if args.report_output.strip() else None

    with tempfile.TemporaryDirectory(prefix=f"{args.task}-batch-audit-") as temp_dir:
        temp_path = Path(temp_dir)
        matched_sample_ids: set[str] = set()
        sample_files = sorted(sample_dir.glob("*.json"))
        if not sample_files:
            raise SystemExit(f"no sample json files found in: {make_repo_relative(sample_dir)}")

        for sample_file in sample_files:
            sample = load_json_document(sample_file)
            if not isinstance(sample, dict):
                raise SystemExit(f"sample must be a json object: {make_repo_relative(sample_file)}")

            sample_id = str(sample.get("sample_id") or "").strip()
            if sample_id not in sample_to_record_id:
                continue

            matched_sample_ids.add(sample_id)
            sample["candidate_response_record"] = make_override_record_ref(
                manifest_path=manifest_path,
                record_id=sample_to_record_id[sample_id],
                expected_source=expected_source,
                required_capture_origin=required_capture_origin,
                required_collection_batch=required_collection_batch,
                required_tags=required_tags,
            )
            (temp_path / sample_file.name).write_text(
                json.dumps(sample, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )

        unmatched_sample_ids = sorted(set(sample_to_record_id) - matched_sample_ids)
        if unmatched_sample_ids:
            raise SystemExit(
                "manifest contains sample_id values that are missing from the audit sample directory: "
                + ", ".join(unmatched_sample_ids)
            )
        if not matched_sample_ids:
            raise SystemExit("no samples matched the manifest sample_id set")

        print(
            f"auditing {len(matched_sample_ids)} sample(s) from {manifest_label} against task '{args.task}'",
            file=sys.stderr,
        )
        command = [
            sys.executable,
            str(REPO_ROOT / "scripts/run-eval-regression.py"),
            args.task,
            "--sample-dir",
            str(temp_path),
        ]
        if args.fail_on_violation:
            command.append("--fail-on-violation")
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

        if report_path is not None:
            parsed = parse_regression_output(result.stdout, result.stderr)
            report_document = {
                "schema_version": 1,
                "task": args.task,
                "manifest_path": manifest_label,
                "sample_dir": make_repo_relative(sample_dir),
                "matched_sample_count": len(matched_sample_ids),
                "expected_source": expected_source,
                "required_capture_origin": required_capture_origin,
                "required_collection_batch": required_collection_batch,
                "required_tags": required_tags,
                "exit_code": result.returncode,
                "passed_count": parsed["passed_count"],
                "failed_count": parsed["failed_count"],
                "violation_count": parsed["violation_count"],
                "warning_line": parsed["warning_line"],
                "summary_line": parsed["summary_line"],
                "stderr": result.stderr,
                "samples": parsed["samples"],
            }
            report_path.parent.mkdir(parents=True, exist_ok=True)
            report_path.write_text(json.dumps(report_document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            print(f"wrote audit report to {make_repo_relative(report_path)}", file=sys.stderr)

            if args.replay_index_output.strip() or args.build_negative_replay:
                replay_index_output = (
                    resolve_relative_to_repo(args.replay_index_output)
                    if args.replay_index_output.strip()
                    else derive_negative_replay_index_output(report_path)
                )
                initial_negative_sample_dir = ""
                same_sample_negative_output_dir = resolve_same_sample_negative_output_dir(args)
                if args.build_negative_replay:
                    bootstrap_negative_output_dir = temp_path / "bootstrap-negative-replay"
                    bootstrap_negative_output_dir.mkdir(parents=True, exist_ok=True)
                    initial_negative_sample_dir = str(bootstrap_negative_output_dir)
                elif args.negative_output_dir.strip():
                    initial_negative_output_dir = resolve_relative_to_repo(args.negative_output_dir)
                    if initial_negative_output_dir.is_dir():
                        initial_negative_sample_dir = args.negative_output_dir
                run_repo_script(
                    "build-negative-replay-index.py",
                    build_negative_replay_index_args(report_path, replay_index_output, initial_negative_sample_dir),
                )

                if args.build_negative_replay:
                    if parsed["failed_count"] == 0:
                        print(
                            "audit report does not contain failed samples; skipping same-sample negative replay generation",
                            file=sys.stderr,
                        )
                    else:
                        negative_args = [
                            "--index",
                            make_repo_relative(replay_index_output),
                            "--output-dir",
                            same_sample_negative_output_dir,
                        ]
                        run_repo_script(
                            "build-radish-docs-negative-replay.py",
                            negative_args,
                        )
                        run_repo_script(
                            "build-negative-replay-index.py",
                            build_negative_replay_index_args(
                                report_path,
                                replay_index_output,
                                same_sample_negative_output_dir,
                            ),
                        )
        return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
