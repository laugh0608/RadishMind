#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (  # noqa: E402
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_relative_to_repo,
    write_json_document,
)
from services.runtime.eval_regression import parse_regression_output  # noqa: E402


BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/batch-orchestration-summary.schema.json"
RADISHFLOW_BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH = REPO_ROOT / "datasets/eval/radishflow-batch-artifact-summary.schema.json"
NEGATIVE_REPLAY_INDEX_SCHEMA_PATH = REPO_ROOT / "datasets/eval/negative-replay-index.schema.json"
RECOMMENDED_NEGATIVE_REPLAY_SUMMARY_SCHEMA_PATH = (
    REPO_ROOT / "datasets/eval/recommended-negative-replay-summary.schema.json"
)
NEGATIVE_REPLAY_PIPELINES = {
    "radish-docs-qa": {
        "eval_task": "radish-docs-qa-negative",
        "pipeline": "radish-docs-qa-recommended-negative-replay",
    },
    "radishflow-suggest-edits": {
        "eval_task": "radishflow-suggest-edits-negative",
        "pipeline": "radishflow-suggest-edits-recommended-negative-replay",
    },
    "radishflow-ghost-completion": {
        "eval_task": "radishflow-ghost-completion-negative",
        "pipeline": "radishflow-ghost-completion-recommended-negative-replay",
    },
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Run the recommended Radish docs QA negative replay groups from a batch artifact summary "
            "and write a structured replay governance summary."
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
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional structured recommended replay summary output path override.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Do not write files; fail if the generated recommended replay summary differs from the on-disk file.",
    )
    return parser.parse_args()


def expect_object(document: Any, label: str) -> dict[str, Any]:
    if not isinstance(document, dict):
        raise SystemExit(f"{label} must be a json object")
    return document


def unique_strings(values: list[str]) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for value in values:
        normalized = str(value or "").strip()
        if normalized and normalized not in seen:
            seen.add(normalized)
            result.append(normalized)
    return result


def load_batch_artifact_summary(summary_path: Path) -> dict[str, Any]:
    document = expect_object(load_json_document(summary_path), make_repo_relative(summary_path))
    eval_task = str(document.get("eval_task") or "").strip()
    schema_path = (
        RADISHFLOW_BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH
        if eval_task in {"radishflow-suggest-edits", "radishflow-ghost-completion"}
        else BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH
    )
    ensure_schema(document, schema_path, make_repo_relative(summary_path))
    return document


def resolve_negative_replay_index(summary_document: dict[str, Any], summary_path: Path, replay_mode: str) -> Path:
    artifacts = expect_object(summary_document.get("artifacts"), f"{make_repo_relative(summary_path)} artifacts")
    artifact_key = "cross_sample_negative_replay_index" if replay_mode == "cross_sample" else "negative_replay_index"
    negative_replay_index = expect_object(
        artifacts.get(artifact_key),
        f"{make_repo_relative(summary_path)} artifacts.{artifact_key}",
    )
    index_path_value = str(negative_replay_index.get("path") or "").strip()
    if not index_path_value:
        raise SystemExit(f"{make_repo_relative(summary_path)}: artifacts.{artifact_key}.path is required")
    if negative_replay_index.get("requested") is not True:
        raise SystemExit(f"{make_repo_relative(summary_path)}: artifacts.{artifact_key}.requested must be true")
    if negative_replay_index.get("exists") is not True:
        raise SystemExit(f"{make_repo_relative(summary_path)}: artifacts.{artifact_key}.exists must be true")

    index_path = resolve_relative_to_repo(index_path_value)
    if not index_path.is_file():
        raise SystemExit(
            f"{make_repo_relative(summary_path)}: referenced negative replay index file not found: {index_path_value}"
        )
    return index_path


def resolve_recommended_selection(summary_document: dict[str, Any], top_n: int, replay_mode: str) -> tuple[list[str], str]:
    if top_n <= 0:
        raise SystemExit("--top must be greater than 0")

    recommended = expect_object(summary_document.get("recommended_negative_replays"), "recommended_negative_replays")
    default_replay_mode = str(recommended.get("default_replay_mode") or "").strip()
    if default_replay_mode not in {"same_sample", "cross_sample"}:
        raise SystemExit(
            "recommended_negative_replays.default_replay_mode must be one of: cross_sample, same_sample"
        )

    selected_replay_mode = replay_mode.strip() or default_replay_mode
    if selected_replay_mode not in {"same_sample", "cross_sample"}:
        raise SystemExit(
            f"recommended_negative_replays replay mode must be one of: cross_sample, same_sample; got '{selected_replay_mode}'"
        )
    group_id_field = "cross_sample_recommended_group_ids" if selected_replay_mode == "cross_sample" else "recommended_group_ids"
    recommended_group_ids = unique_strings(
        [str(group_id) for group_id in recommended.get(group_id_field) or []]
    )
    if not recommended_group_ids:
        raise SystemExit(f"recommended_negative_replays.{group_id_field} is empty")

    return recommended_group_ids[:top_n], selected_replay_mode


def load_negative_replay_index(index_path: Path) -> dict[str, Any]:
    document = expect_object(load_json_document(index_path), make_repo_relative(index_path))
    ensure_schema(document, NEGATIVE_REPLAY_INDEX_SCHEMA_PATH, make_repo_relative(index_path))
    return document


def resolve_negative_eval_task(batch_artifact_summary: dict[str, Any], summary_path: Path) -> str:
    eval_task = str(batch_artifact_summary.get("eval_task") or "").strip()
    config = NEGATIVE_REPLAY_PIPELINES.get(eval_task)
    if config is None:
        raise SystemExit(
            f"{make_repo_relative(summary_path)}: unsupported eval_task for recommended negative replay: "
            f"{eval_task or '<empty>'}"
        )
    return config["eval_task"]


def resolve_recommended_pipeline(eval_task: str) -> str:
    for config in NEGATIVE_REPLAY_PIPELINES.values():
        if config["eval_task"] == eval_task:
            return config["pipeline"]
    raise SystemExit(f"unsupported negative replay eval_task: {eval_task}")


def build_group_map(index_document: dict[str, Any]) -> dict[str, dict[str, Any]]:
    groups: dict[str, dict[str, Any]] = {}
    for group in index_document.get("violation_groups") or []:
        group_object = expect_object(group, "negative replay index violation_groups entry")
        group_id = str(group_object.get("group_id") or "").strip()
        if not group_id:
            raise SystemExit("negative replay index violation_groups entries must include group_id")
        groups[group_id] = group_object
    return groups


def derive_summary_output_path(batch_artifact_summary_path: Path, top_n: int, replay_mode: str) -> Path:
    name = batch_artifact_summary_path.name
    if name.endswith(".artifacts.json"):
        base_name = name[: -len(".artifacts.json")]
    elif name.endswith(".json"):
        base_name = name[: -len(".json")]
    else:
        base_name = name
    return batch_artifact_summary_path.with_name(
        f"{base_name}.recommended-negative-replay-top{top_n}-{replay_mode}.summary.json"
    )


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


def build_group_summary(
    *,
    batch_artifact_summary_path: Path,
    negative_eval_task: str,
    group_id: str,
    replay_mode: str,
    sample_dir_override: str,
    fail_on_violation: bool,
    group_document: dict[str, Any],
) -> tuple[dict[str, Any], int]:
    command = [
        sys.executable,
        str(REPO_ROOT / "scripts" / "run-eval-regression.py"),
        negative_eval_task,
        "--batch-artifact-summary",
        make_repo_relative(batch_artifact_summary_path),
        "--group-id",
        group_id,
        "--replay-mode",
        replay_mode,
    ]
    if sample_dir_override:
        command.extend(["--sample-dir", sample_dir_override])
    if fail_on_violation:
        command.append("--fail-on-violation")

    result = run_command(command)
    parsed = parse_regression_output(result.stdout, result.stderr)
    expected_candidate_violations = unique_strings(
        [str(value) for value in group_document.get("expected_candidate_violations") or []]
    )

    entries: list[dict[str, Any]] = []
    for raw_entry in group_document.get("entries") or []:
        entry = expect_object(raw_entry, f"negative replay index group '{group_id}' entry")
        if str(entry.get("replay_mode") or "").strip() != replay_mode:
            continue
        entries.append(entry)
    sample_files = sorted(str(sample.get("sample_file") or "").strip() for sample in parsed["samples"] if sample.get("sample_file"))
    summary_document = {
        "group_id": group_id,
        "replay_mode": replay_mode,
        "requested_entry_count": len(entries),
        "selected_sample_count": len(parsed["samples"]),
        "exit_code": result.returncode,
        "passed_sample_count": parsed["passed_count"],
        "failed_sample_count": parsed["failed_count"],
        "violation_count": parsed["violation_count"],
        "expected_candidate_violations": expected_candidate_violations,
        "sample_files": sample_files,
        "summary_line": str(parsed["summary_line"] or ""),
        "warning_line": str(parsed["warning_line"] or ""),
        "command": {
            "python": [
                "python3",
                "./scripts/run-eval-regression.py",
                negative_eval_task,
                "--batch-artifact-summary",
                make_repo_relative(batch_artifact_summary_path),
                "--group-id",
                group_id,
                "--replay-mode",
                replay_mode,
                *(["--sample-dir", sample_dir_override] if sample_dir_override else []),
                *(["--fail-on-violation"] if fail_on_violation else []),
            ],
        },
    }
    if result.stderr.strip():
        summary_document["stderr"] = result.stderr
    failed_samples = [sample for sample in parsed["samples"] if sample.get("status") == "fail"]
    if failed_samples:
        summary_document["failed_samples"] = failed_samples
    return summary_document, result.returncode


def build_summary_document(
    *,
    summary_path: Path,
    batch_artifact_summary_path: Path,
    batch_artifact_summary: dict[str, Any],
    negative_replay_index_path: Path,
    negative_eval_task: str,
    requested_top: int,
    selected_group_ids: list[str],
    replay_mode: str,
    sample_dir_override: str,
    fail_on_violation: bool,
    group_summaries: list[dict[str, Any]],
    overall_exit_code: int,
) -> dict[str, Any]:
    summary_document: dict[str, Any] = {
        "schema_version": 1,
        "pipeline": resolve_recommended_pipeline(negative_eval_task),
        "project": str(batch_artifact_summary.get("project") or "").strip(),
        "task": str(batch_artifact_summary.get("task") or "").strip(),
        "eval_task": negative_eval_task,
        "provider": str(batch_artifact_summary.get("provider") or "").strip(),
        "collection_batch": str(batch_artifact_summary.get("collection_batch") or "").strip(),
        "batch_artifact_summary_path": make_repo_relative(batch_artifact_summary_path),
        "negative_replay_index_path": make_repo_relative(negative_replay_index_path),
        "output_path": make_repo_relative(summary_path),
        "selection": {
            "requested_top": requested_top,
            "selected_group_ids": selected_group_ids,
            "replay_mode": replay_mode,
            "sample_dir_override": sample_dir_override,
            "fail_on_violation": fail_on_violation,
        },
        "execution": {
            "overall_exit_code": overall_exit_code,
        },
        "summary": {
            "selected_group_count": len(group_summaries),
            "passed_group_count": sum(1 for group in group_summaries if int(group.get("exit_code") or 0) == 0),
            "failed_group_count": sum(1 for group in group_summaries if int(group.get("exit_code") or 0) != 0),
            "selected_sample_count": sum(int(group.get("selected_sample_count") or 0) for group in group_summaries),
            "passed_sample_count": sum(int(group.get("passed_sample_count") or 0) for group in group_summaries),
            "failed_sample_count": sum(int(group.get("failed_sample_count") or 0) for group in group_summaries),
            "violation_count": sum(int(group.get("violation_count") or 0) for group in group_summaries),
        },
        "groups": group_summaries,
    }

    model = str(batch_artifact_summary.get("model") or "").strip()
    if model:
        summary_document["model"] = model

    capture_origin = str(batch_artifact_summary.get("capture_origin") or "").strip()
    if capture_origin:
        summary_document["capture_origin"] = capture_origin

    ensure_schema(summary_document, RECOMMENDED_NEGATIVE_REPLAY_SUMMARY_SCHEMA_PATH, "recommended negative replay summary")
    return summary_document


def check_summary_output(path: Path, document: dict[str, Any]) -> None:
    if not path.is_file():
        raise SystemExit(f"recommended negative replay summary file not found for --check: {make_repo_relative(path)}")
    existing = load_json_document(path)
    if existing != document:
        expected_text = json.dumps(document, ensure_ascii=False, indent=2) + "\n"
        actual_text = path.read_text(encoding="utf-8")
        raise SystemExit(
            "recommended negative replay summary is stale: "
            f"{make_repo_relative(path)}\nexpected:\n{expected_text}\nactual:\n{actual_text}"
        )


def main() -> int:
    args = parse_args()
    batch_artifact_summary_path = resolve_relative_to_repo(args.batch_artifact_summary)
    if not batch_artifact_summary_path.is_file():
        raise SystemExit(f"batch artifact summary file not found: {args.batch_artifact_summary}")

    batch_artifact_summary = load_batch_artifact_summary(batch_artifact_summary_path)
    negative_eval_task = resolve_negative_eval_task(batch_artifact_summary, batch_artifact_summary_path)
    selected_group_ids, replay_mode = resolve_recommended_selection(batch_artifact_summary, args.top, args.replay_mode)
    negative_replay_index_path = resolve_negative_replay_index(batch_artifact_summary, batch_artifact_summary_path, replay_mode)
    negative_replay_index = load_negative_replay_index(negative_replay_index_path)
    group_map = build_group_map(negative_replay_index)

    missing_group_ids = [group_id for group_id in selected_group_ids if group_id not in group_map]
    if missing_group_ids:
        raise SystemExit(
            "recommended_negative_replays references missing group_id value(s): " + ", ".join(sorted(missing_group_ids))
        )

    summary_path = (
        resolve_relative_to_repo(args.summary_output)
        if args.summary_output.strip()
        else derive_summary_output_path(batch_artifact_summary_path, args.top, replay_mode)
    )

    sample_dir_override = args.sample_dir.strip()
    group_summaries: list[dict[str, Any]] = []
    overall_exit_code = 0

    for group_id in selected_group_ids:
        print(f"replaying recommended group {group_id} ({replay_mode})", file=sys.stderr)
        group_summary, exit_code = build_group_summary(
            batch_artifact_summary_path=batch_artifact_summary_path,
            negative_eval_task=negative_eval_task,
            group_id=group_id,
            replay_mode=replay_mode,
            sample_dir_override=sample_dir_override,
            fail_on_violation=args.fail_on_violation,
            group_document=group_map[group_id],
        )
        group_summaries.append(group_summary)
        if exit_code != 0:
            overall_exit_code = exit_code

    summary_document = build_summary_document(
        summary_path=summary_path,
        batch_artifact_summary_path=batch_artifact_summary_path,
        batch_artifact_summary=batch_artifact_summary,
        negative_replay_index_path=negative_replay_index_path,
        negative_eval_task=negative_eval_task,
        requested_top=args.top,
        selected_group_ids=selected_group_ids,
        replay_mode=replay_mode,
        sample_dir_override=sample_dir_override,
        fail_on_violation=args.fail_on_violation,
        group_summaries=group_summaries,
        overall_exit_code=overall_exit_code,
    )

    if args.check:
        check_summary_output(summary_path, summary_document)
    else:
        write_json_document(summary_path, summary_document)
        print(f"wrote recommended replay summary to {make_repo_relative(summary_path)}", file=sys.stderr)

    return overall_exit_code


if __name__ == "__main__":
    raise SystemExit(main())
