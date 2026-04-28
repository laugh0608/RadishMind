#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
SCHEMA_PATH = REPO_ROOT / "contracts/radishmind-core-offline-eval-run.schema.json"
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-offline-eval-run-basic.json"
THRESHOLDS_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-eval-thresholds.json"
DOC_PATH = REPO_ROOT / "docs/radishmind-core-baseline-evaluation.md"

REQUIRED_TASKS = {
    ("radishflow", "suggest_flowsheet_edits"),
    ("radishflow", "suggest_ghost_completion"),
    ("radish", "answer_docs_question"),
}
REQUIRED_MODEL_IDS = {
    "minimind-v",
    "radishmind-core-3b",
    "radishmind-core-4b",
    "radishmind-core-7b",
    "qwen2.5-vl",
    "smolvlm",
}
CORE_MODEL_SIZE = {
    "radishmind-core-3b": "3B",
    "radishmind-core-4b": "4B",
    "radishmind-core-7b": "7B",
}
REQUIRED_NEVER_PROMOTE = {
    "model_weakens_requires_confirmation",
    "model_claims_direct_business_write",
    "model_requires_image_pixel_generation_inside_core",
    "model_requires_14b_or_32b_as_default_target",
    "missing_observed_metric_values",
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


def build_threshold_maps(thresholds: dict[str, Any]) -> tuple[dict[str, dict[str, float]], dict[tuple[str, str], set[str]]]:
    metrics = thresholds.get("metrics")
    assert_condition(isinstance(metrics, list), "thresholds.metrics must be a list")
    metric_thresholds: dict[str, dict[str, float]] = {}
    for metric in metrics:
        assert_condition(isinstance(metric, dict), "threshold metric must be an object")
        metric_id = str(metric.get("id") or "").strip()
        if not metric_id:
            continue
        metric_thresholds[metric_id] = {
            "3B": float(metric.get("minimum_for_3b")),
            "4B": float(metric.get("minimum_for_4b")),
            "7B": float(metric.get("minimum_for_7b")),
        }

    required_tasks = thresholds.get("required_tasks")
    assert_condition(isinstance(required_tasks, list), "thresholds.required_tasks must be a list")
    task_metrics: dict[tuple[str, str], set[str]] = {}
    for task in required_tasks:
        assert_condition(isinstance(task, dict), "threshold required task must be an object")
        key = (str(task.get("project") or "").strip(), str(task.get("task") or "").strip())
        task_metrics[key] = set(require_string_list(task, "required_metric_ids"))
    return metric_thresholds, task_metrics


def check_execution(document: dict[str, Any]) -> None:
    assert_condition(document.get("schema_version") == 1, "offline eval run schema_version must be 1")
    assert_condition(document.get("kind") == "radishmind_core_offline_eval_run", "offline eval run kind mismatch")
    assert_condition(document.get("phase") == "M4-preparation", "offline eval run phase mismatch")
    execution = document.get("execution") if isinstance(document.get("execution"), dict) else {}
    assert_condition(execution.get("run_status") == "planned", "contract fixture must remain a planned eval run")
    assert_condition(execution.get("does_not_run_models") is True, "contract smoke must not run models")
    assert_condition(execution.get("provider_access") == "none", "contract fixture must not access providers")
    assert_condition(execution.get("model_artifacts_downloaded") is False, "contract fixture must not download models")


def check_sample_selection(document: dict[str, Any]) -> dict[tuple[str, str], int]:
    groups = document.get("sample_selection")
    assert_condition(isinstance(groups, list), "sample_selection must be a list")
    group_keys: set[tuple[str, str]] = set()
    sample_counts: dict[tuple[str, str], int] = {}
    seen_sample_ids: set[str] = set()

    for group in groups:
        assert_condition(isinstance(group, dict), "sample_selection items must be objects")
        project = str(group.get("project") or "").strip()
        task = str(group.get("task") or "").strip()
        key = (project, task)
        group_keys.add(key)
        minimum_sample_count = group.get("minimum_sample_count")
        assert_condition(isinstance(minimum_sample_count, int), f"{project}/{task} minimum_sample_count must be integer")
        assert_condition(minimum_sample_count >= 3, f"{project}/{task} must select at least 3 samples")
        selected_samples = group.get("selected_samples")
        assert_condition(isinstance(selected_samples, list), f"{project}/{task} selected_samples must be a list")
        assert_condition(
            len(selected_samples) >= minimum_sample_count,
            f"{project}/{task} selected sample count is below minimum",
        )
        sample_counts[key] = len(selected_samples)

        for selected_sample in selected_samples:
            assert_condition(isinstance(selected_sample, dict), "selected sample must be an object")
            sample_id = str(selected_sample.get("sample_id") or "").strip()
            relative_path = str(selected_sample.get("path") or "").strip()
            assert_condition(sample_id, "selected sample must include sample_id")
            assert_condition(relative_path, f"{sample_id} must include path")
            assert_condition(sample_id not in seen_sample_ids, f"duplicate selected sample_id: {sample_id}")
            seen_sample_ids.add(sample_id)
            sample_path = REPO_ROOT / relative_path
            assert_condition(sample_path.is_file(), f"selected sample path does not exist: {relative_path}")
            sample_document = load_json_document(sample_path)
            assert_condition(isinstance(sample_document, dict), f"selected sample must be object: {relative_path}")
            assert_condition(sample_document.get("sample_id") == sample_id, f"{relative_path} sample_id mismatch")
            assert_condition(sample_document.get("project") == project, f"{relative_path} project mismatch")
            assert_condition(sample_document.get("task") == task, f"{relative_path} task mismatch")
            coverage_tags = set(require_string_list(selected_sample, "coverage_tags"))
            assert_condition(coverage_tags, f"{sample_id} must include coverage tags")

    missing_tasks = sorted(REQUIRED_TASKS - group_keys)
    if missing_tasks:
        project, task = missing_tasks[0]
        raise SystemExit(f"offline eval run is missing sample selection for {project}/{task}")
    return sample_counts


def check_candidate_models(document: dict[str, Any]) -> None:
    models = document.get("candidate_models")
    assert_condition(isinstance(models, list), "candidate_models must be a list")
    model_by_id = {
        str(model.get("model_id") or "").strip(): model
        for model in models
        if isinstance(model, dict) and str(model.get("model_id") or "").strip()
    }
    missing_models = sorted(REQUIRED_MODEL_IDS - set(model_by_id))
    if missing_models:
        raise SystemExit(f"offline eval run is missing candidate model: {missing_models[0]}")

    for disallowed_size in ("14B", "32B"):
        assert_condition(
            all(model.get("target_size") != disallowed_size for model in model_by_id.values()),
            f"{disallowed_size} must not be a candidate model target size",
        )

    for model_id, size in CORE_MODEL_SIZE.items():
        model = model_by_id[model_id]
        assert_condition(model.get("role") == "core_candidate", f"{model_id} must be a core candidate")
        assert_condition(model.get("target_size") == size, f"{model_id} target_size mismatch")
        assert_condition(model.get("local_memory_budget_gb") == 32, f"{model_id} must stay within 32GB budget")

    assert_condition(model_by_id["qwen2.5-vl"].get("role") == "teacher_baseline", "Qwen2.5-VL must stay teacher baseline")
    assert_condition(model_by_id["smolvlm"].get("role") == "lightweight_baseline", "SmolVLM must stay lightweight baseline")


def check_metrics(document: dict[str, Any], metric_thresholds: dict[str, dict[str, float]]) -> None:
    metrics = document.get("metrics")
    assert_condition(isinstance(metrics, list), "metrics must be a list")
    fixture_metrics = {
        str(metric.get("metric_id") or "").strip(): metric
        for metric in metrics
        if isinstance(metric, dict) and str(metric.get("metric_id") or "").strip()
    }
    missing_metrics = sorted(set(metric_thresholds) - set(fixture_metrics))
    if missing_metrics:
        raise SystemExit(f"offline eval run is missing metric: {missing_metrics[0]}")

    for metric_id, expected_by_size in metric_thresholds.items():
        metric = fixture_metrics[metric_id]
        assert_condition(metric.get("blocking") is True, f"{metric_id} must remain blocking")
        threshold_by_size = metric.get("threshold_by_size") if isinstance(metric.get("threshold_by_size"), dict) else {}
        for size, expected in expected_by_size.items():
            assert_condition(
                float(threshold_by_size.get(size)) == expected,
                f"{metric_id}.{size} threshold must match radishmind-core-eval-thresholds fixture",
            )


def check_cost_budget(document: dict[str, Any]) -> None:
    cost_budget = document.get("cost_budget") if isinstance(document.get("cost_budget"), dict) else {}
    assert_condition(cost_budget.get("target_local_memory_gb") == 32, "offline eval run must keep 32GB budget")
    assert_condition(cost_budget.get("maximum_local_core_size") == "7B", "7B must remain max local core size")
    excluded = set(require_string_list(cost_budget, "excluded_default_sizes"))
    assert_condition({"14B", "32B"}.issubset(excluded), "14B/32B must stay excluded default sizes")

    per_candidate = cost_budget.get("per_candidate")
    assert_condition(isinstance(per_candidate, list), "cost_budget.per_candidate must be a list")
    budget_by_model = {
        str(item.get("model_id") or "").strip(): item
        for item in per_candidate
        if isinstance(item, dict) and str(item.get("model_id") or "").strip()
    }
    for model_id in ("radishmind-core-3b", "radishmind-core-4b", "radishmind-core-7b"):
        assert_condition(model_id in budget_by_model, f"cost budget is missing {model_id}")
        assert_condition(budget_by_model[model_id].get("max_memory_gb") == 32, f"{model_id} must keep 32GB memory budget")
    assert_condition(budget_by_model["radishmind-core-3b"].get("run_allowed") is True, "3B must remain runnable")
    assert_condition(budget_by_model["radishmind-core-4b"].get("run_allowed") is True, "4B must remain runnable")
    assert_condition(budget_by_model["radishmind-core-7b"].get("run_allowed") is False, "7B must remain deferred")


def check_result_records(
    document: dict[str, Any],
    task_metrics: dict[tuple[str, str], set[str]],
    sample_counts: dict[tuple[str, str], int],
    metric_thresholds: dict[str, dict[str, float]],
) -> None:
    records = document.get("result_records")
    assert_condition(isinstance(records, list), "result_records must be a list")
    expected_record_keys = {
        (model_id, project, task)
        for model_id in CORE_MODEL_SIZE
        for project, task in REQUIRED_TASKS
    }
    record_keys: set[tuple[str, str, str]] = set()

    for record in records:
        assert_condition(isinstance(record, dict), "result_records items must be objects")
        model_id = str(record.get("model_id") or "").strip()
        project = str(record.get("project") or "").strip()
        task = str(record.get("task") or "").strip()
        key = (model_id, project, task)
        assert_condition(key not in record_keys, f"duplicate result record: {model_id}/{project}/{task}")
        record_keys.add(key)
        assert_condition(model_id in CORE_MODEL_SIZE, f"result record model must be a core candidate: {model_id}")
        assert_condition((project, task) in REQUIRED_TASKS, f"unexpected result record task: {project}/{task}")
        assert_condition(
            record.get("sample_count") == sample_counts[(project, task)],
            f"{model_id}/{project}/{task} sample_count must match sample selection",
        )

        expected_metric_ids = task_metrics[(project, task)]
        metric_results = record.get("metric_results")
        assert_condition(isinstance(metric_results, list), f"{model_id}/{project}/{task} metric_results must be a list")
        metric_result_by_id = {
            str(metric_result.get("metric_id") or "").strip(): metric_result
            for metric_result in metric_results
            if isinstance(metric_result, dict) and str(metric_result.get("metric_id") or "").strip()
        }
        missing_metrics = sorted(expected_metric_ids - set(metric_result_by_id))
        if missing_metrics:
            raise SystemExit(f"{model_id}/{project}/{task} is missing metric result: {missing_metrics[0]}")
        unexpected_metrics = sorted(set(metric_result_by_id) - expected_metric_ids)
        if unexpected_metrics:
            raise SystemExit(f"{model_id}/{project}/{task} has unexpected metric result: {unexpected_metrics[0]}")

        size = CORE_MODEL_SIZE[model_id]
        for metric_id, metric_result in metric_result_by_id.items():
            expected_threshold = metric_thresholds[metric_id][size]
            assert_condition(
                float(metric_result.get("threshold")) == expected_threshold,
                f"{model_id}/{project}/{task}/{metric_id} threshold mismatch",
            )
            status = record.get("status")
            observed_value = metric_result.get("observed_value")
            passed = metric_result.get("passed")
            if status == "completed":
                assert_condition(observed_value is not None, f"{model_id}/{project}/{task}/{metric_id} completed result needs observed_value")
                assert_condition(isinstance(passed, bool), f"{model_id}/{project}/{task}/{metric_id} completed result needs passed")
                assert_condition(passed == (float(observed_value) >= expected_threshold), f"{metric_id} passed flag mismatch")
            else:
                assert_condition(observed_value is None, f"{model_id}/{project}/{task}/{metric_id} planned result must not fake observed_value")
                assert_condition(passed is None, f"{model_id}/{project}/{task}/{metric_id} planned result must not fake passed")

    missing_records = sorted(expected_record_keys - record_keys)
    if missing_records:
        model_id, project, task = missing_records[0]
        raise SystemExit(f"offline eval run is missing result record: {model_id}/{project}/{task}")


def check_decision(document: dict[str, Any]) -> None:
    decision = document.get("decision") if isinstance(document.get("decision"), dict) else {}
    assert_condition(decision.get("promotion_status") == "no_promotion_planned", "planned fixture must not promote a model")
    assert_condition(decision.get("requires_human_review") is True, "offline eval decision must require human review")
    never_promote = set(require_string_list(decision, "never_promote_when"))
    missing_rules = sorted(REQUIRED_NEVER_PROMOTE - never_promote)
    if missing_rules:
        raise SystemExit(f"offline eval decision is missing never-promote rule: {missing_rules[0]}")


def check_doc() -> None:
    content = DOC_PATH.read_text(encoding="utf-8")
    for pattern in (
        "contracts/radishmind-core-offline-eval-run.schema.json",
        "scripts/checks/fixtures/radishmind-core-offline-eval-run-basic.json",
        "scripts/check-radishmind-core-offline-eval-run-contract.py",
        "离线评测样本选择与结果记录",
        "不下载模型、不启动训练、不访问外部 provider",
    ):
        assert_condition(pattern in content, f"baseline evaluation doc is missing expected content: {pattern}")


def main() -> int:
    schema = load_json_document(SCHEMA_PATH)
    fixture = load_json_document(FIXTURE_PATH)
    thresholds = load_json_document(THRESHOLDS_PATH)

    jsonschema.Draft202012Validator.check_schema(schema)
    jsonschema.validate(fixture, schema)
    if not isinstance(fixture, dict):
        raise SystemExit("offline eval run fixture must be an object")
    if not isinstance(thresholds, dict):
        raise SystemExit("eval thresholds fixture must be an object")

    metric_thresholds, task_metrics = build_threshold_maps(thresholds)
    check_execution(fixture)
    sample_counts = check_sample_selection(fixture)
    check_candidate_models(fixture)
    check_metrics(fixture, metric_thresholds)
    check_cost_budget(fixture)
    check_result_records(fixture, task_metrics, sample_counts, metric_thresholds)
    check_decision(fixture)
    check_doc()

    print("radishmind core offline eval run contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
