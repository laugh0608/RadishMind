#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

ENTRY_PATH = REPO_ROOT / "training/experiments/radishmind-core-task-scoped-builder-broader-review-entry-v0.json"

EXPECTED_SAMPLE_SET_IDS = ["full-holdout-9", "holdout6-v2-non-overlap"]
EXPECTED_TASK_COUNTS = {
    "radishflow/suggest_flowsheet_edits": 5,
    "radishflow/suggest_ghost_completion": 5,
    "radish/answer_docs_question": 5,
}
EXPECTED_FULL_HOLDOUT_IDS = {
    "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
    "radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001",
    "radishflow-suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001",
    "radishflow-suggest-ghost-completion-flash-inlet-001",
    "radishflow-suggest-ghost-completion-heater-stream-name-001",
    "radishflow-suggest-ghost-completion-mixer-standard-outlet-001",
    "radish-answer-docs-question-attachment-mixed-001",
    "radish-answer-docs-question-docs-attachments-faq-001",
    "radish-answer-docs-question-navigation-001",
}
EXPECTED_V2_IDS = {
    "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001",
    "radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001",
    "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
    "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
    "radish-answer-docs-question-docs-faq-forum-conflict-001",
    "radish-answer-docs-question-evidence-gap-001",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check the committed broader task-scoped builder review entry.",
    )
    parser.add_argument(
        "--summary-output",
        default="",
        help="Optional path to write the normalized broader review entry summary.",
    )
    parser.add_argument(
        "--check-summary",
        default="",
        help="Optional expected summary path. Fails if regenerated summary differs.",
    )
    return parser.parse_args()


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{repo_rel(path)}': {exc}") from exc


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a non-empty clean path")
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def require_dict(document: dict[str, Any], key: str) -> dict[str, Any]:
    value = document.get(key)
    require(isinstance(value, dict), f"{key} must be an object")
    return value


def collect_sample_selection(manifest: dict[str, Any]) -> tuple[dict[str, list[str]], dict[str, int], list[str]]:
    sample_ids_by_task: dict[str, list[str]] = defaultdict(list)
    sample_counts_by_task: dict[str, int] = defaultdict(int)
    ordered_sample_ids: list[str] = []
    for group in manifest.get("sample_selection") or []:
        require(isinstance(group, dict), "sample_selection groups must be objects")
        project = str(group.get("project") or "").strip()
        task = str(group.get("task") or "").strip()
        require(project and task, "sample_selection entries must include project and task")
        task_key = f"{project}/{task}"
        selected_samples = group.get("selected_samples") or []
        require(isinstance(selected_samples, list), f"{task_key} selected_samples must be a list")
        for sample in selected_samples:
            require(isinstance(sample, dict), f"{task_key} selected sample must be an object")
            sample_id = str(sample.get("sample_id") or "").strip()
            require(sample_id, f"{task_key} selected sample_id must be non-empty")
            sample_ids_by_task[task_key].append(sample_id)
            sample_counts_by_task[task_key] += 1
            ordered_sample_ids.append(sample_id)
    return dict(sample_ids_by_task), dict(sample_counts_by_task), ordered_sample_ids


def collect_track(document: dict[str, Any], track_id: str) -> dict[str, Any]:
    tracks = document.get("tracks")
    require(isinstance(tracks, list), "tracks must be a list")
    for entry in tracks:
        if isinstance(entry, dict) and entry.get("track_id") == track_id:
            return entry
    raise SystemExit(f"missing track: {track_id}")


def build_summary(entry: dict[str, Any]) -> dict[str, Any]:
    require(entry.get("schema_version") == 1, "entry schema_version must be 1")
    require(
        entry.get("kind") == "radishmind_core_task_scoped_builder_broader_review_entry",
        "entry kind mismatch",
    )
    require(entry.get("entry_id") == "radishmind-core-task-scoped-builder-broader-review-entry-v0", "entry_id mismatch")
    require(entry.get("status") == "ready_for_broader_review", "entry status must be ready_for_broader_review")
    require(entry.get("phase") == "M4-preparation", "entry phase mismatch")
    require(entry.get("does_not_run_models") is True, "entry must not run models")
    require(entry.get("does_not_generate_jsonl") is True, "entry must not generate JSONL")
    require(entry.get("does_not_mark_review_pass") is True, "entry must not mark review pass")
    require(entry.get("candidate_track") == "--build-task-scoped-response", "candidate_track mismatch")

    for field_name in (
        "source_experiment",
        "source_review_plan",
        "source_full_holdout_review_records",
        "source_run_set_summary",
        "source_natural_language_audit",
    ):
        require_existing_file(str(entry.get(field_name) or ""), field_name=field_name)

    source_sample_sets = require_dict(entry, "source_sample_sets")
    require(set(source_sample_sets) == {"full_holdout_9", "holdout6_v2_non_overlap"}, "source sample sets mismatch")
    full_holdout_source = require_dict(source_sample_sets, "full_holdout_9")
    v2_source = require_dict(source_sample_sets, "holdout6_v2_non_overlap")
    for label, source in (("full_holdout", full_holdout_source), ("v2", v2_source)):
        require(str(source.get("sample_set_id") or "").strip(), f"{label} sample_set_id must be non-empty")
        require_existing_file(str(source.get("candidate_manifest") or ""), field_name=f"{label}.candidate_manifest")
        require_existing_file(str(source.get("candidate_eval_manifest") or ""), field_name=f"{label}.candidate_eval_manifest")

    full_holdout_manifest = load_json(REPO_ROOT / str(full_holdout_source["candidate_eval_manifest"]))
    v2_manifest = load_json(REPO_ROOT / str(v2_source["candidate_eval_manifest"]))
    _, full_counts_by_task, full_ordered_ids = collect_sample_selection(full_holdout_manifest)
    _, v2_counts_by_task, v2_ordered_ids = collect_sample_selection(v2_manifest)

    require(set(full_ordered_ids) == EXPECTED_FULL_HOLDOUT_IDS, "full-holdout sample set mismatch")
    require(set(v2_ordered_ids) == EXPECTED_V2_IDS, "v2 sample set mismatch")
    require(not set(full_ordered_ids).intersection(v2_ordered_ids), "broader review sample sets must not overlap")

    merged_sample_ids = full_ordered_ids + v2_ordered_ids
    merged_unique_ids = list(dict.fromkeys(merged_sample_ids))
    require(len(merged_sample_ids) == 15, "broader sample surface must contain 15 samples")
    require(len(merged_unique_ids) == 15, "broader sample surface must contain 15 unique samples")

    task_counts = {
        task: full_counts_by_task.get(task, 0) + v2_counts_by_task.get(task, 0)
        for task in EXPECTED_TASK_COUNTS
    }
    require(task_counts == EXPECTED_TASK_COUNTS, "task counts must total 5 per task")

    full_holdout_records = load_json(REPO_ROOT / str(entry["source_full_holdout_review_records"]))
    require(full_holdout_records.get("status") == "reviewed_pass", "full-holdout review records must be reviewed_pass")
    batch_summary = require_dict(full_holdout_records, "batch_summary")
    require(batch_summary.get("reviewed_pass_count") == 9, "full-holdout reviewed_pass_count mismatch")
    run_set_summary = load_json(REPO_ROOT / str(entry["source_run_set_summary"]))
    full_track = collect_track(run_set_summary, "task_scoped_builder_full_holdout_9")
    require(full_track.get("human_review_status") == "reviewed_pass", "full-holdout track human review must pass")
    require(full_track.get("reviewed_pass_count") == 9, "full-holdout track reviewed_pass count mismatch")
    require(
        str(full_track.get("route_signal") or "") == "broader_tooling_track_machine_and_review_passed_citation_fixture_tightened",
        "full-holdout route signal mismatch",
    )
    fixed_track = collect_track(run_set_summary, "task_scoped_builder_guardrail_fixed_v2")
    require(fixed_track.get("promotion_status") == "no_promotion_planned", "fixed v2 track must remain no_promotion_planned")
    require(
        str(fixed_track.get("route_signal") or "") == "task_scoped_builder_tooling_track_ready_for_broader_review",
        "fixed v2 route signal mismatch",
    )
    audit_document = load_json(REPO_ROOT / str(entry["source_natural_language_audit"]))
    audit_summary = require_dict(audit_document, "summary")
    require(audit_summary.get("status") == "pass", "natural-language audit must pass")
    require(audit_summary.get("violation_count") == 0, "natural-language audit must stay violation free")

    computed_evidence_summary = {
        "full_holdout_review_status": full_holdout_records.get("status"),
        "full_holdout_reviewed_pass_count": batch_summary.get("reviewed_pass_count"),
        "full_holdout_route_signal": full_track.get("route_signal"),
        "v2_route_signal": fixed_track.get("route_signal"),
        "v2_natural_language_audit_status": audit_summary.get("status"),
        "v2_natural_language_violation_count": audit_summary.get("violation_count"),
    }
    require(entry.get("evidence_summary") == computed_evidence_summary, "entry evidence_summary mismatch")

    sample_surface = require_dict(entry, "sample_surface")
    require(sample_surface.get("surface_id") == "broader-task-scoped-builder-15-sample-surface-v0", "surface_id mismatch")
    require(sample_surface.get("sample_count") == 15, "sample_count mismatch")
    require(sample_surface.get("unique_sample_count") == 15, "unique_sample_count mismatch")
    require(sample_surface.get("overlap_detected") is False, "overlap_detected must remain false")
    require(sample_surface.get("task_counts") == EXPECTED_TASK_COUNTS, "sample_surface task_counts mismatch")
    require(sample_surface.get("sample_set_ids") == EXPECTED_SAMPLE_SET_IDS, "sample_set_ids mismatch")
    require(sample_surface.get("sample_ids_by_task") == {
        "radishflow/suggest_flowsheet_edits": [
            "radishflow-suggest-flowsheet-edits-action-citation-ordering-diagnostic-artifact-snapshot-001",
            "radishflow-suggest-flowsheet-edits-compressor-parameter-update-ordering-001",
            "radishflow-suggest-flowsheet-edits-valve-local-fix-vs-global-balance-001",
            "radishflow-suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001",
            "radishflow-suggest-flowsheet-edits-efficiency-range-ordering-001",
        ],
        "radishflow/suggest_ghost_completion": [
            "radishflow-suggest-ghost-completion-flash-inlet-001",
            "radishflow-suggest-ghost-completion-heater-stream-name-001",
            "radishflow-suggest-ghost-completion-mixer-standard-outlet-001",
            "radishflow-suggest-ghost-completion-flash-vapor-outlet-001",
            "radishflow-suggest-ghost-completion-valve-ambiguous-no-tab-001",
        ],
        "radish/answer_docs_question": [
            "radish-answer-docs-question-attachment-mixed-001",
            "radish-answer-docs-question-docs-attachments-faq-001",
            "radish-answer-docs-question-navigation-001",
            "radish-answer-docs-question-docs-faq-forum-conflict-001",
            "radish-answer-docs-question-evidence-gap-001",
        ],
    }, "sample_ids_by_task mismatch")
    require(sample_surface.get("sample_ids_by_sample_set") == {
        "full-holdout-9": full_ordered_ids,
        "holdout6-v2-non-overlap": v2_ordered_ids,
    }, "sample_ids_by_sample_set mismatch")

    broader_review_gate = require_dict(entry, "broader_review_gate")
    require(
        "full_holdout_review_records_are_reviewed_pass" in (broader_review_gate.get("may_execute_when_all_true") or []),
        "broader review gate missing full-holdout pass condition",
    )
    require(
        "sample_surface_has_15_unique_samples" in (broader_review_gate.get("may_execute_when_all_true") or []),
        "broader review gate missing sample surface condition",
    )
    require(
        "candidate_outputs_leave_tmp" in (broader_review_gate.get("must_not_execute_when_any_true") or []),
        "broader review gate missing tmp boundary",
    )

    review_batch = require_dict(entry, "review_batch")
    require(review_batch.get("review_batch_id") == "task-scoped-builder-broader-15-sample-review-v0", "review_batch_id mismatch")
    require(review_batch.get("status") == "planned", "review_batch status must remain planned")
    require(review_batch.get("manual_review_ratio") == 1.0, "manual_review_ratio mismatch")
    require(review_batch.get("requires_offline_eval_run") is True, "offline eval requirement mismatch")
    require(review_batch.get("requires_natural_language_audit") is True, "natural language audit requirement mismatch")
    require(review_batch.get("requires_human_reviewer") is True, "human reviewer requirement mismatch")

    validation_entry = require_dict(entry, "validation_entry")
    require(validation_entry.get("script") == "scripts/check-radishmind-core-task-scoped-builder-broader-review-entry.py", "validation script mismatch")
    require(validation_entry.get("check_summary") == "scripts/checks/fixtures/radishmind-core-task-scoped-builder-broader-review-entry-summary.json", "check_summary mismatch")
    validation_args = validation_entry.get("command_args")
    require(isinstance(validation_args, list), "validation_entry.command_args must be a list")
    require(
        validation_args
        == [
            "python3",
            "scripts/check-radishmind-core-task-scoped-builder-broader-review-entry.py",
            "--summary-output",
            "tmp/radishmind-core-task-scoped-builder-broader-review-entry-summary.json",
            "--check-summary",
            "scripts/checks/fixtures/radishmind-core-task-scoped-builder-broader-review-entry-summary.json",
        ],
        "validation command args mismatch",
    )

    route_signal = require_dict(entry, "route_signal")
    require(route_signal.get("status") == "broader_task_scoped_builder_review_entry_fixed", "route signal status mismatch")

    return {
        "schema_version": 1,
        "kind": "radishmind_core_task_scoped_builder_broader_review_entry_summary",
        "entry_id": entry.get("entry_id"),
        "status": entry.get("status"),
        "phase": entry.get("phase"),
        "candidate_track": entry.get("candidate_track"),
        "sample_surface": {
            "surface_id": sample_surface.get("surface_id"),
            "sample_set_ids": sample_surface.get("sample_set_ids"),
            "sample_count": sample_surface.get("sample_count"),
            "unique_sample_count": sample_surface.get("unique_sample_count"),
            "task_counts": sample_surface.get("task_counts"),
            "sample_ids_by_task": sample_surface.get("sample_ids_by_task"),
            "sample_ids_by_sample_set": sample_surface.get("sample_ids_by_sample_set"),
            "overlap_detected": sample_surface.get("overlap_detected"),
        },
        "evidence_summary": computed_evidence_summary,
        "validation_entry": {
            "script": validation_entry.get("script"),
            "check_summary": validation_entry.get("check_summary"),
            "command_args": validation_args,
        },
        "review_batch": {
            "review_batch_id": review_batch.get("review_batch_id"),
            "status": review_batch.get("status"),
            "manual_review_ratio": review_batch.get("manual_review_ratio"),
            "requires_offline_eval_run": review_batch.get("requires_offline_eval_run"),
            "requires_natural_language_audit": review_batch.get("requires_natural_language_audit"),
            "requires_human_reviewer": review_batch.get("requires_human_reviewer"),
        },
        "broader_review_gate": broader_review_gate,
        "route_signal": route_signal,
    }


def main() -> int:
    args = parse_args()
    entry = load_json(ENTRY_PATH)
    summary = build_summary(entry)

    if args.summary_output:
        write_json(Path(args.summary_output), summary)
    if args.check_summary:
        expected = load_json(Path(args.check_summary))
        require(summary == expected, "broader review entry summary mismatch")

    print("radishmind core task-scoped builder broader review entry check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
