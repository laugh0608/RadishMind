#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
THRESHOLDS_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-eval-thresholds.json"
DOC_PATH = REPO_ROOT / "docs/radishmind-core-baseline-evaluation.md"

REQUIRED_TASKS = {
    ("radishflow", "suggest_flowsheet_edits"),
    ("radishflow", "suggest_ghost_completion"),
    ("radish", "answer_docs_question"),
}
REQUIRED_METRICS = {
    "schema_validity_rate",
    "citation_alignment_rate",
    "risk_confirmation_preservation_rate",
    "high_risk_action_confirmation_rate",
    "advisory_action_boundary_rate",
}
REQUIRED_PROMOTION_RULES = {
    "prefer_3b_when",
    "evaluate_4b_when",
    "evaluate_7b_when",
    "never_promote_when",
}


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_string_list(document: dict[str, Any], key: str) -> list[str]:
    values = document.get(key)
    assert_condition(isinstance(values, list), f"{key} must be a list")
    normalized = [str(value).strip() for value in values if str(value).strip()]
    assert_condition(len(normalized) == len(values), f"{key} must only contain non-empty strings")
    return normalized


def assert_rate(value: Any, *, label: str) -> float:
    assert_condition(isinstance(value, (int, float)), f"{label} must be numeric")
    rate = float(value)
    assert_condition(0 <= rate <= 1, f"{label} must be between 0 and 1")
    return rate


def check_boundaries(thresholds: dict[str, Any]) -> None:
    assert_condition(thresholds.get("does_not_run_models") is True, "threshold smoke must not run models")
    boundaries = thresholds.get("core_boundaries") if isinstance(thresholds.get("core_boundaries"), dict) else {}
    assert_condition(boundaries.get("advisory_only") is True, "thresholds must preserve advisory-only boundary")
    assert_condition(
        boundaries.get("direct_image_pixel_generation") is False,
        "thresholds must keep image pixel generation out of RadishMind-Core",
    )
    assert_condition(
        boundaries.get("image_pixel_generation_owner") == "radishmind-image-adapter-backend",
        "image pixel generation must stay with image adapter/backend",
    )
    assert_condition(boundaries.get("default_core_target_max") == "7B", "7B must remain the default local upper bound")
    excluded = set(require_string_list(boundaries, "excluded_default_core_targets"))
    assert_condition({"14B", "32B"}.issubset(excluded), "14B/32B must remain excluded default core targets")


def check_tasks(thresholds: dict[str, Any]) -> None:
    tasks = thresholds.get("required_tasks")
    assert_condition(isinstance(tasks, list), "required_tasks must be a list")
    task_keys: set[tuple[str, str]] = set()
    for item in tasks:
        assert_condition(isinstance(item, dict), "required_tasks items must be objects")
        project = str(item.get("project") or "").strip()
        task = str(item.get("task") or "").strip()
        assert_condition(project and task, "required task must include project and task")
        task_keys.add((project, task))
        sample_count = item.get("minimum_sample_count")
        assert_condition(isinstance(sample_count, int), f"{project}/{task} minimum_sample_count must be integer")
        assert_condition(sample_count >= 3, f"{project}/{task} must require at least 3 offline samples")
        metric_ids = set(require_string_list(item, "required_metric_ids"))
        missing_metrics = sorted(REQUIRED_METRICS - metric_ids)
        if project == "radish" and task == "answer_docs_question":
            missing_metrics = sorted((REQUIRED_METRICS - {"high_risk_action_confirmation_rate", "advisory_action_boundary_rate"}) - metric_ids)
            assert_condition(
                "retrieval_source_contract_rate" in metric_ids,
                "radish/answer_docs_question must include retrieval_source_contract_rate",
            )
        if missing_metrics:
            raise SystemExit(f"{project}/{task} is missing required metric: {missing_metrics[0]}")
    missing_tasks = sorted(REQUIRED_TASKS - task_keys)
    if missing_tasks:
        project, task = missing_tasks[0]
        raise SystemExit(f"thresholds are missing required task: {project}/{task}")


def check_metrics(thresholds: dict[str, Any]) -> None:
    metrics = thresholds.get("metrics")
    assert_condition(isinstance(metrics, list), "metrics must be a list")
    metric_by_id = {
        str(metric.get("id") or "").strip(): metric
        for metric in metrics
        if isinstance(metric, dict) and str(metric.get("id") or "").strip()
    }
    missing_metrics = sorted((REQUIRED_METRICS | {"retrieval_source_contract_rate"}) - set(metric_by_id))
    if missing_metrics:
        raise SystemExit(f"thresholds are missing metric definition: {missing_metrics[0]}")

    for metric_id, metric in metric_by_id.items():
        assert_condition(str(metric.get("description") or "").strip(), f"{metric_id} must include description")
        minimum_3b = assert_rate(metric.get("minimum_for_3b"), label=f"{metric_id}.minimum_for_3b")
        minimum_4b = assert_rate(metric.get("minimum_for_4b"), label=f"{metric_id}.minimum_for_4b")
        minimum_7b = assert_rate(metric.get("minimum_for_7b"), label=f"{metric_id}.minimum_for_7b")
        assert_condition(minimum_3b <= minimum_4b <= minimum_7b, f"{metric_id} thresholds must be monotonic")
        assert_condition(metric.get("blocking") is True, f"{metric_id} must be blocking in first-round thresholds")

    for metric_id in (
        "high_risk_action_confirmation_rate",
        "advisory_action_boundary_rate",
    ):
        metric = metric_by_id[metric_id]
        assert_condition(metric.get("minimum_for_3b") == 1.0, f"{metric_id} must require 1.0 for 3B")
        assert_condition(metric.get("minimum_for_4b") == 1.0, f"{metric_id} must require 1.0 for 4B")
        assert_condition(metric.get("minimum_for_7b") == 1.0, f"{metric_id} must require 1.0 for 7B")


def check_promotion_rules(thresholds: dict[str, Any]) -> None:
    rules = thresholds.get("promotion_rules") if isinstance(thresholds.get("promotion_rules"), dict) else {}
    missing_rule_groups = sorted(REQUIRED_PROMOTION_RULES - set(rules))
    if missing_rule_groups:
        raise SystemExit(f"thresholds are missing promotion rule group: {missing_rule_groups[0]}")
    for key in REQUIRED_PROMOTION_RULES:
        values = set(require_string_list(rules, key))
        assert_condition(values, f"{key} must not be empty")

    never_promote = set(require_string_list(rules, "never_promote_when"))
    for required in (
        "model_weakens_requires_confirmation",
        "model_claims_direct_business_write",
        "model_requires_image_pixel_generation_inside_core",
        "model_requires_14b_or_32b_as_default_target",
    ):
        assert_condition(required in never_promote, f"never_promote_when must include {required}")

    evaluate_4b = set(require_string_list(rules, "evaluate_4b_when"))
    assert_condition(
        "3b_misses_any_blocking_metric_by_more_than_0_02" in evaluate_4b,
        "evaluate_4b_when must be tied to a measured 3B gap",
    )
    evaluate_7b = set(require_string_list(rules, "evaluate_7b_when"))
    for required in (
        "3b_and_4b_both_miss_blocking_metrics",
        "failure_is_not_solved_by_data_prompt_or_rules",
        "quantized_runtime_fits_32gb_budget",
        "measured_quality_gap_justifies_cost",
    ):
        assert_condition(required in evaluate_7b, f"evaluate_7b_when must include {required}")


def check_doc() -> None:
    content = DOC_PATH.read_text(encoding="utf-8")
    for pattern in (
        "scripts/checks/fixtures/radishmind-core-eval-thresholds.json",
        "scripts/check-radishmind-core-eval-thresholds.py",
        "schema_validity_rate",
        "citation_alignment_rate",
        "risk_confirmation_preservation_rate",
        "high_risk_action_confirmation_rate",
        "图片像素生成仍不进入 `RadishMind-Core`",
        "`radishflow/suggest_flowsheet_edits`",
        "`radishflow/suggest_ghost_completion`",
        "`radish/answer_docs_question`",
    ):
        assert_condition(pattern in content, f"baseline evaluation doc is missing expected content: {pattern}")


def main() -> int:
    thresholds = load_json_document(THRESHOLDS_PATH)
    if not isinstance(thresholds, dict):
        raise SystemExit("eval thresholds fixture must be an object")
    assert_condition(thresholds.get("schema_version") == 1, "eval thresholds schema_version must be 1")
    assert_condition(thresholds.get("kind") == "radishmind_core_eval_thresholds", "eval thresholds kind mismatch")
    assert_condition(thresholds.get("phase") == "M4-preparation", "eval thresholds phase mismatch")
    check_boundaries(thresholds)
    check_tasks(thresholds)
    check_metrics(thresholds)
    check_promotion_rules(thresholds)
    check_doc()
    print("radishmind core eval thresholds check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
