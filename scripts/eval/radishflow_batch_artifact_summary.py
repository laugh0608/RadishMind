from __future__ import annotations

import shlex
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from services.runtime.candidate_records import (
    RECORD_BATCH_SCHEMA_PATH,
    SHORT_LAYOUT_DUMPS_DIR,
    SHORT_LAYOUT_RECORDS_DIR,
    SHORT_LAYOUT_RESPONSES_DIR,
    ensure_schema,
    load_json_document,
    make_repo_relative,
    resolve_manifest_output_root,
)

RADISHFLOW_BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH = (
    REPO_ROOT / "datasets/eval/radishflow-batch-artifact-summary.schema.json"
)

TASK_SUMMARY_CONFIG: dict[str, dict[str, str]] = {
    "suggest_flowsheet_edits": {
        "pipeline": "radishflow-suggest-edits-poc-batch",
        "eval_task": "radishflow-suggest-edits",
    },
    "suggest_ghost_completion": {
        "pipeline": "radishflow-ghost-completion-poc-batch",
        "eval_task": "radishflow-ghost-completion",
    },
}


def derive_output_root_from_record_dir(record_dir: Path) -> Path:
    resolved = record_dir.resolve()
    if resolved.name in {"records", SHORT_LAYOUT_RECORDS_DIR}:
        return resolved.parent
    return resolved


def derive_artifact_summary_path(output_root: Path, collection_batch: str) -> Path:
    return output_root / "artifacts.json"


def load_object_if_exists(path: Path) -> dict[str, Any] | None:
    if not path.is_file():
        return None
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"{make_repo_relative(path)} must be a json object")
    return document


def count_replay_entries(index_document: dict[str, Any] | None, replay_mode: str) -> tuple[int, int]:
    if index_document is None:
        return 0, 0
    linked_count = 0
    for group in index_document.get("violation_groups") or []:
        if not isinstance(group, dict):
            raise SystemExit("negative replay index violation_groups entries must be json objects")
        for entry in group.get("entries") or []:
            if not isinstance(entry, dict):
                raise SystemExit("negative replay index group entries must be json objects")
            if str(entry.get("replay_mode") or "").strip() == replay_mode:
                linked_count += 1
    return linked_count, 0


def build_recommended_negative_replays(
    *,
    artifact_summary_path: Path,
    eval_task: str,
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

            python_command = " ".join(
                [
                    "python3",
                    shlex.quote("./scripts/run-eval-regression.py"),
                    eval_task,
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
    document: dict[str, Any] = {
        "default_replay_mode": "same_sample",
        "recommended_group_ids": [group["group_id"] for group in same_sample_groups],
        "groups": same_sample_groups,
    }
    if cross_sample_groups:
        document["cross_sample_recommended_group_ids"] = [group["group_id"] for group in cross_sample_groups]
        document["cross_sample_groups"] = cross_sample_groups
    return document


def collect_record_paths(output_root: Path) -> list[Path]:
    return sorted(path for path in output_root.rglob("*.record.json") if path.is_file())


def uses_short_layout_output_root(output_root: Path) -> bool:
    resolved = output_root.resolve()
    return resolved.parent.parent.name == "batches"


def infer_provider_from_record(record: dict[str, Any]) -> str:
    provider = str(record.get("provider") or "").strip()
    if provider:
        return provider
    capture_metadata = record.get("capture_metadata") or {}
    tags = capture_metadata.get("tags") or []
    if isinstance(tags, list):
        normalized_tags = {str(tag).strip() for tag in tags if str(tag).strip()}
        if "openai-compatible" in normalized_tags:
            return "openai-compatible"
        if "mock" in normalized_tags:
            return "mock"
    source = str(record.get("source") or "").strip()
    if source == "simulated_candidate_response":
        return "mock"
    if source == "captured_candidate_response":
        return "openai-compatible"
    return ""


def collect_unique_non_empty(values: list[str]) -> str:
    normalized = sorted({value.strip() for value in values if value and value.strip()})
    if len(normalized) > 1:
        raise SystemExit(f"inconsistent non-empty values found: {normalized}")
    return normalized[0] if normalized else ""


def load_radishflow_batch_artifact_summary(path: Path) -> dict[str, Any]:
    document = load_json_document(path)
    if not isinstance(document, dict):
        raise SystemExit(f"batch artifact summary must be a json object: {make_repo_relative(path)}")
    ensure_schema(
        document,
        RADISHFLOW_BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH,
        f"radishflow batch artifact summary '{make_repo_relative(path)}'",
    )
    return document


def build_radishflow_batch_artifact_summary_document(
    *,
    output_root: Path,
    manifest_path: Path,
    audit_report_path: Path,
    capture_exit_code: int = 0,
    audit_requested: bool = True,
    provider_override: str = "",
    model_override: str = "",
) -> dict[str, Any]:
    manifest = load_json_document(manifest_path)
    ensure_schema(manifest, RECORD_BATCH_SCHEMA_PATH, make_repo_relative(manifest_path))
    if not isinstance(manifest, dict):
        raise SystemExit(f"{make_repo_relative(manifest_path)} must be a json object")
    resolved_output_root = resolve_manifest_output_root(manifest_path, manifest)

    task = str(manifest.get("task") or "").strip()
    task_config = TASK_SUMMARY_CONFIG.get(task)
    if task_config is None:
        raise SystemExit(f"unsupported RadishFlow batch manifest task: {task or '<empty>'}")

    audit_report = load_object_if_exists(audit_report_path)
    record_paths = collect_record_paths(resolved_output_root)
    records = [load_json_document(path) for path in record_paths]
    for path, record in zip(record_paths, records):
        if not isinstance(record, dict):
            raise SystemExit(f"{make_repo_relative(path)} must be a json object")

    provider = provider_override.strip() or collect_unique_non_empty(
        [infer_provider_from_record(record) for record in records]
    )
    if not provider:
        raise SystemExit(
            f"unable to infer provider for RadishFlow batch: manifest={make_repo_relative(manifest_path)}"
        )

    model = model_override.strip() or collect_unique_non_empty([str(record.get("model") or "").strip() for record in records])
    capture_origin = str(manifest.get("capture_origin") or "").strip() or collect_unique_non_empty(
        [
            str(((record.get("capture_metadata") or {}).get("capture_origin")) or "").strip()
            for record in records
        ]
    )

    uses_short_layout = uses_short_layout_output_root(resolved_output_root)
    records_dir = resolved_output_root / (SHORT_LAYOUT_RECORDS_DIR if uses_short_layout else "records")
    if not records_dir.is_dir():
        records_dir = resolved_output_root

    response_dir = resolved_output_root / (SHORT_LAYOUT_RESPONSES_DIR if uses_short_layout else "responses")
    dump_dir = resolved_output_root / (SHORT_LAYOUT_DUMPS_DIR if uses_short_layout else "dumps")
    negative_replay_index_path = resolved_output_root / "negative-replay-index.json"
    cross_sample_negative_replay_index_path = resolved_output_root / "cross-sample-replay-index.json"
    recommended_negative_replay_summary_path = resolved_output_root / "recommended-negative-replay.summary.json"
    cross_sample_recommended_negative_replay_summary_path = (
        resolved_output_root / "cross-sample-recommended-negative-replay.summary.json"
    )
    response_count = len(list(response_dir.glob("*.response.json"))) if response_dir.is_dir() else 0
    dump_count = len(list(dump_dir.glob("*.dump.json"))) if dump_dir.is_dir() else 0
    root_record_count = sum(1 for path in record_paths if path.parent == records_dir)
    nested_record_count = len(record_paths) - root_record_count
    manifest_records = manifest.get("records") or []
    if not isinstance(manifest_records, list):
        raise SystemExit(f"{make_repo_relative(manifest_path)} records must be an array")
    replay_index = load_object_if_exists(negative_replay_index_path)
    cross_sample_replay_index = load_object_if_exists(cross_sample_negative_replay_index_path)
    linked_same_sample_count, _ = count_replay_entries(replay_index, "same_sample")
    linked_cross_sample_count, _ = count_replay_entries(cross_sample_replay_index, "cross_sample")
    violation_group_count = len(list((replay_index or {}).get("violation_groups") or []))
    cross_sample_violation_group_count = len(list((cross_sample_replay_index or {}).get("violation_groups") or []))
    recommended_negative_replays = build_recommended_negative_replays(
        artifact_summary_path=derive_artifact_summary_path(resolved_output_root, str(manifest.get("collection_batch") or "")),
        eval_task=task_config["eval_task"] + "-negative",
        replay_index=replay_index,
        cross_sample_replay_index=cross_sample_replay_index,
    )

    summary_document: dict[str, Any] = {
        "schema_version": 1,
        "pipeline": task_config["pipeline"],
        "project": str(manifest.get("project") or "").strip(),
        "task": task,
        "eval_task": task_config["eval_task"],
        "provider": provider,
        "collection_batch": str(manifest.get("collection_batch") or "").strip(),
        "source": str(manifest.get("source") or "").strip(),
        "output_root": make_repo_relative(resolved_output_root),
        "execution": {
            "capture_exit_code": max(0, int(capture_exit_code)),
            "audit_exit_code": None if audit_report is None else int(audit_report.get("exit_code") or 0),
        },
        "artifacts": {
            "manifest": {
                "path": make_repo_relative(manifest_path),
                "exists": manifest_path.is_file(),
                "record_count": len(manifest_records),
            },
            "audit_report": {
                "requested": audit_requested,
                "path": make_repo_relative(audit_report_path),
                "exists": audit_report_path.is_file(),
            },
            "output_root": {
                "path": make_repo_relative(resolved_output_root),
                "exists": resolved_output_root.is_dir(),
            },
            "records": {
                "path": make_repo_relative(records_dir),
                "exists": bool(record_paths),
                "file_count": len(record_paths),
                "root_file_count": root_record_count,
                "nested_file_count": nested_record_count,
            },
            "responses": {
                "path": make_repo_relative(response_dir),
                "exists": response_dir.is_dir(),
                "file_count": response_count,
            },
            "dumps": {
                "path": make_repo_relative(dump_dir),
                "exists": dump_dir.is_dir(),
                "file_count": dump_count,
            },
            "negative_replay_index": {
                "requested": negative_replay_index_path.is_file(),
                "path": make_repo_relative(negative_replay_index_path),
                "exists": negative_replay_index_path.is_file(),
            },
            "cross_sample_negative_replay_index": {
                "requested": cross_sample_negative_replay_index_path.is_file(),
                "path": make_repo_relative(cross_sample_negative_replay_index_path),
                "exists": cross_sample_negative_replay_index_path.is_file(),
            },
            "recommended_negative_replay_summary": {
                "requested": recommended_negative_replay_summary_path.is_file(),
                "path": make_repo_relative(recommended_negative_replay_summary_path),
                "exists": recommended_negative_replay_summary_path.is_file(),
            },
            "cross_sample_recommended_negative_replay_summary": {
                "requested": cross_sample_recommended_negative_replay_summary_path.is_file(),
                "path": make_repo_relative(cross_sample_recommended_negative_replay_summary_path),
                "exists": cross_sample_recommended_negative_replay_summary_path.is_file(),
            },
        },
        "summary": {
            "captured_record_count": len(manifest_records),
            "discovered_record_count": len(record_paths),
            "matched_sample_count": int((audit_report or {}).get("matched_sample_count") or 0),
            "passed_sample_count": int((audit_report or {}).get("passed_count") or 0),
            "failed_sample_count": int((audit_report or {}).get("failed_count") or 0),
            "violation_count": int((audit_report or {}).get("violation_count") or 0),
            "root_record_count": root_record_count,
            "nested_record_count": nested_record_count,
            "response_file_count": response_count,
            "dump_file_count": dump_count,
            "violation_group_count": violation_group_count,
            "cross_sample_violation_group_count": cross_sample_violation_group_count,
            "linked_same_sample_negative_count": linked_same_sample_count,
            "linked_cross_sample_negative_count": linked_cross_sample_count,
            "recommended_replay_group_count": len(list(recommended_negative_replays.get("recommended_group_ids") or [])),
            "cross_sample_recommended_replay_group_count": len(
                list(recommended_negative_replays.get("cross_sample_recommended_group_ids") or [])
            ),
        },
        "recommended_negative_replays": recommended_negative_replays,
    }
    if model:
        summary_document["model"] = model
    if capture_origin:
        summary_document["capture_origin"] = capture_origin

    ensure_schema(
        summary_document,
        RADISHFLOW_BATCH_ARTIFACT_SUMMARY_SCHEMA_PATH,
        "radishflow batch artifact summary",
    )
    return summary_document
