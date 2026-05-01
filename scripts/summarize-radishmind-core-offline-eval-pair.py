#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--raw-run", required=True)
    parser.add_argument("--repaired-run", required=True)
    parser.add_argument("--raw-summary", required=True)
    parser.add_argument("--repaired-summary", required=True)
    parser.add_argument("--output", required=True)
    return parser.parse_args()


def repo_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def repo_rel(path: Path) -> str:
    try:
        return path.relative_to(REPO_ROOT).as_posix()
    except ValueError:
        return path.as_posix()


def load_json(path_value: str) -> dict[str, Any]:
    path = repo_path(path_value)
    try:
        document = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{repo_rel(path)}': {exc}") from exc
    if not isinstance(document, dict):
        raise SystemExit(f"{repo_rel(path)} must be a JSON object")
    return document


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def write_json(path_value: str, document: dict[str, Any]) -> None:
    path = repo_path(path_value)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def metric_map(run: dict[str, Any]) -> dict[str, dict[str, float | bool | None]]:
    metrics: dict[str, dict[str, float | bool | None]] = {}
    for record in run.get("result_records") or []:
        if not isinstance(record, dict):
            continue
        task_key = f"{record.get('project')}/{record.get('task')}"
        for metric in record.get("metric_results") or []:
            if not isinstance(metric, dict):
                continue
            metric_id = str(metric.get("metric_id") or "").strip()
            if not metric_id:
                continue
            metrics[f"{task_key}/{metric_id}"] = {
                "observed_value": metric.get("observed_value"),
                "passed": metric.get("passed"),
            }
    return metrics


def build_task_comparisons(raw_run: dict[str, Any], repaired_run: dict[str, Any]) -> list[dict[str, Any]]:
    repaired_by_task = {
        f"{record.get('project')}/{record.get('task')}": record
        for record in repaired_run.get("result_records") or []
        if isinstance(record, dict)
    }
    comparisons: list[dict[str, Any]] = []
    for raw_record in raw_run.get("result_records") or []:
        if not isinstance(raw_record, dict):
            continue
        task_key = f"{raw_record.get('project')}/{raw_record.get('task')}"
        repaired_record = repaired_by_task.get(task_key)
        require(repaired_record is not None, f"repaired run is missing task record: {task_key}")
        raw_metrics = {
            metric["metric_id"]: metric
            for metric in raw_record.get("metric_results") or []
            if isinstance(metric, dict) and metric.get("metric_id")
        }
        repaired_metrics = {
            metric["metric_id"]: metric
            for metric in repaired_record.get("metric_results") or []
            if isinstance(metric, dict) and metric.get("metric_id")
        }
        metric_deltas = []
        for metric_id, raw_metric in raw_metrics.items():
            repaired_metric = repaired_metrics.get(metric_id)
            require(repaired_metric is not None, f"repaired run is missing metric {task_key}/{metric_id}")
            raw_value = raw_metric.get("observed_value")
            repaired_value = repaired_metric.get("observed_value")
            delta = None
            if isinstance(raw_value, (int, float)) and isinstance(repaired_value, (int, float)):
                delta = round(float(repaired_value) - float(raw_value), 6)
            metric_deltas.append(
                {
                    "metric_id": metric_id,
                    "raw_observed_value": raw_value,
                    "repaired_observed_value": repaired_value,
                    "delta": delta,
                    "raw_passed": raw_metric.get("passed"),
                    "repaired_passed": repaired_metric.get("passed"),
                }
            )
        comparisons.append(
            {
                "task": task_key,
                "sample_count": raw_record.get("sample_count"),
                "raw_all_metrics_passed": all(metric.get("passed") is True for metric in raw_metrics.values()),
                "repaired_all_metrics_passed": all(metric.get("passed") is True for metric in repaired_metrics.values()),
                "metric_deltas": metric_deltas,
            }
        )
    return comparisons


def build_summary(args: argparse.Namespace) -> dict[str, Any]:
    raw_run = load_json(args.raw_run)
    repaired_run = load_json(args.repaired_run)
    raw_summary = load_json(args.raw_summary)
    repaired_summary = load_json(args.repaired_summary)

    require(raw_run.get("kind") == "radishmind_core_offline_eval_run", "raw run kind mismatch")
    require(repaired_run.get("kind") == "radishmind_core_offline_eval_run", "repaired run kind mismatch")
    require(raw_summary.get("kind") == "radishmind_core_candidate_run_summary", "raw candidate summary kind mismatch")
    require(
        repaired_summary.get("kind") == "radishmind_core_candidate_run_summary",
        "repaired candidate summary kind mismatch",
    )
    require(raw_summary.get("sample_count") == repaired_summary.get("sample_count"), "raw/repaired sample count mismatch")

    raw_provider = raw_summary.get("provider") if isinstance(raw_summary.get("provider"), dict) else {}
    repaired_provider = repaired_summary.get("provider") if isinstance(repaired_summary.get("provider"), dict) else {}
    model_id = str(repaired_provider.get("model_id") or raw_provider.get("model_id") or "unknown")
    raw_obs = raw_summary.get("observation_summary") if isinstance(raw_summary.get("observation_summary"), dict) else {}
    repaired_obs = (
        repaired_summary.get("observation_summary") if isinstance(repaired_summary.get("observation_summary"), dict) else {}
    )
    postprocess = repaired_summary.get("postprocess_policy") if isinstance(repaired_summary.get("postprocess_policy"), dict) else {}
    generation_summary = repaired_obs.get("generation_summary") if isinstance(repaired_obs.get("generation_summary"), dict) else {}

    return {
        "schema_version": 1,
        "kind": "radishmind_core_offline_eval_pair_summary",
        "phase": "M4-preparation",
        "model_id": model_id,
        "sample_count": raw_summary.get("sample_count"),
        "inputs": {
            "raw_run": args.raw_run,
            "repaired_run": args.repaired_run,
            "raw_summary": args.raw_summary,
            "repaired_summary": args.repaired_summary,
        },
        "raw": {
            "decision": raw_run.get("decision", {}).get("promotion_status"),
            "schema_valid_rate": raw_obs.get("schema_valid_rate"),
            "task_valid_rate": raw_obs.get("task_valid_rate"),
            "task_validation_attempted": raw_obs.get("task_validation_attempted"),
            "schema_failure_categories": raw_obs.get("schema_failure_categories") or {},
            "task_failure_categories": raw_obs.get("task_failure_categories") or {},
        },
        "repaired": {
            "decision": repaired_run.get("decision", {}).get("promotion_status"),
            "schema_valid_rate": repaired_obs.get("schema_valid_rate"),
            "task_valid_rate": repaired_obs.get("task_valid_rate"),
            "task_validation_attempted": repaired_obs.get("task_validation_attempted"),
            "postprocess_policy": postprocess,
        },
        "generation_summary": generation_summary,
        "task_comparisons": build_task_comparisons(raw_run, repaired_run),
        "interpretation": {
            "raw_model_is_blocked": raw_run.get("decision", {}).get("promotion_status") == "blocked",
            "repaired_output_passes_current_fixture": repaired_run.get("decision", {}).get("promotion_status")
            == "no_promotion_planned",
            "repaired_is_not_raw_capability": True,
            "recommended_next_step": (
                "Keep raw and repaired tracks separate, expand beyond the 9 fixture samples, "
                "and review repaired paths before any training or promotion decision."
            ),
        },
    }


def main() -> int:
    args = parse_args()
    write_json(args.output, build_summary(args))
    print("radishmind core offline eval pair summary written.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
