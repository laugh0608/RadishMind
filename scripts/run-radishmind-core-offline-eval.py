#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any, Callable

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from scripts.eval.regression_diagnostics_suggest import validate_suggest_request, validate_suggest_response  # noqa: E402
from scripts.eval.regression_docs import validate_radish_docs_response, validate_radish_docs_retrieval  # noqa: E402
from scripts.eval.regression_ghost import validate_ghost_completion_request, validate_ghost_completion_response  # noqa: E402
from scripts.eval.regression_shared import TASK_CONFIG, get_array, test_document_against_schema  # noqa: E402


DEFAULT_MANIFEST_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-offline-eval-fixture-run-manifest.json"
DEFAULT_CHECK_OUTPUT_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-offline-eval-golden-run.json"
RUN_SCHEMA_PATH = REPO_ROOT / "contracts/radishmind-core-offline-eval-run.schema.json"
THRESHOLDS_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-eval-thresholds.json"
TASK_TO_CONFIG_KEY = {
    ("radishflow", "suggest_flowsheet_edits"): "radishflow-suggest-edits",
    ("radishflow", "suggest_ghost_completion"): "radishflow-ghost-completion",
    ("radish", "answer_docs_question"): "radish-docs-qa",
}
TASK_VALIDATORS: dict[
    tuple[str, str],
    tuple[Callable[[dict[str, Any], str, list[str]], None] | None, Callable[[dict[str, Any], dict[str, Any], str, str, list[str]], None]],
] = {
    ("radishflow", "suggest_flowsheet_edits"): (validate_suggest_request, validate_suggest_response),
    ("radishflow", "suggest_ghost_completion"): (validate_ghost_completion_request, validate_ghost_completion_response),
    ("radish", "answer_docs_question"): (None, validate_radish_docs_response),
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", default=str(DEFAULT_MANIFEST_PATH.relative_to(REPO_ROOT)))
    parser.add_argument("--output")
    parser.add_argument("--check-output", default=str(DEFAULT_CHECK_OUTPUT_PATH.relative_to(REPO_ROOT)))
    return parser.parse_args()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def write_json(path: Path, document: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def normalize_repo_path(path_value: str) -> Path:
    path = Path(path_value)
    return path if path.is_absolute() else REPO_ROOT / path


def load_thresholds() -> tuple[dict[str, dict[str, float]], dict[tuple[str, str], list[str]]]:
    thresholds = load_json(THRESHOLDS_PATH)
    metrics = thresholds.get("metrics")
    require(isinstance(metrics, list), "radishmind core eval thresholds metrics must be a list")
    metric_thresholds: dict[str, dict[str, float]] = {}
    for metric in metrics:
        require(isinstance(metric, dict), "threshold metric must be an object")
        metric_id = str(metric.get("id") or "").strip()
        require(metric_id, "threshold metric is missing id")
        metric_thresholds[metric_id] = {
            "3B": float(metric["minimum_for_3b"]),
            "4B": float(metric["minimum_for_4b"]),
            "7B": float(metric["minimum_for_7b"]),
        }

    required_tasks = thresholds.get("required_tasks")
    require(isinstance(required_tasks, list), "radishmind core eval thresholds required_tasks must be a list")
    task_metrics: dict[tuple[str, str], list[str]] = {}
    for task_entry in required_tasks:
        require(isinstance(task_entry, dict), "required task threshold entry must be an object")
        key = (str(task_entry.get("project") or "").strip(), str(task_entry.get("task") or "").strip())
        task_metrics[key] = [str(metric_id).strip() for metric_id in get_array(task_entry.get("required_metric_ids")) if str(metric_id).strip()]
    return metric_thresholds, task_metrics


def validate_manifest(manifest: dict[str, Any]) -> None:
    require(manifest.get("schema_version") == 1, "offline eval fixture run manifest schema_version must be 1")
    require(manifest.get("kind") == "radishmind_core_offline_eval_fixture_run_manifest", "offline eval fixture run manifest kind mismatch")
    require(manifest.get("phase") == "M4-preparation", "offline eval fixture run phase mismatch")
    require(manifest.get("response_source") == "golden_response", "first offline eval runner only supports golden_response fixture source")
    execution_policy = manifest.get("execution_policy")
    require(isinstance(execution_policy, dict), "offline eval fixture run execution_policy must be an object")
    require(execution_policy.get("run_status") == "completed", "offline eval fixture run must produce completed records")
    require(execution_policy.get("does_not_run_models") is True, "offline eval fixture run must not run models")
    require(execution_policy.get("provider_access") == "none", "offline eval fixture run must not access providers")
    require(execution_policy.get("model_artifacts_downloaded") is False, "offline eval fixture run must not download model artifacts")


def validate_sample_and_response(
    sample: dict[str, Any],
    response: dict[str, Any],
    *,
    project: str,
    task: str,
    sample_name: str,
) -> dict[str, bool]:
    config = TASK_CONFIG[TASK_TO_CONFIG_KEY[(project, task)]]
    request_violations: list[str] = []
    response_violations: list[str] = []
    retrieval_violations: list[str] = []

    test_document_against_schema(sample, config["sample_schema"], f"{sample_name} sample", request_violations)
    test_document_against_schema(sample.get("input_request"), config["request_schema"], f"{sample_name} input_request", request_violations)
    test_document_against_schema(response, config["response_schema"], f"{sample_name} candidate_response", response_violations)

    request_validator, response_validator = TASK_VALIDATORS[(project, task)]
    if request_validator is not None:
        request_validator(sample, sample_name, request_violations)
    if (project, task) == ("radish", "answer_docs_question"):
        validate_radish_docs_retrieval(sample, sample_name, retrieval_violations)
    response_validator(sample, response, "candidate_response", sample_name, response_violations)

    all_response_violations = response_violations + retrieval_violations
    return {
        "schema_validity_rate": not any("schema validation failed" in violation for violation in request_violations + response_violations),
        "citation_alignment_rate": not any("citation" in violation for violation in all_response_violations),
        "risk_confirmation_preservation_rate": not any(
            ("risk_level" in violation or "requires_confirmation" in violation) for violation in all_response_violations
        ),
        "high_risk_action_confirmation_rate": not any(
            ("requires_confirmation" in violation and ("candidate_edit" in violation or "risk" in violation))
            for violation in all_response_violations
        ),
        "advisory_action_boundary_rate": not any(
            ("direct business write" in violation or "proposed_action" in violation or "candidate_edit" in violation or "ghost_completion" in violation)
            for violation in all_response_violations
        ),
        "retrieval_source_contract_rate": len(retrieval_violations) == 0,
    }


def build_metric_results(
    *,
    task_key: tuple[str, str],
    model_size: str,
    sample_results: list[dict[str, bool]],
    metric_thresholds: dict[str, dict[str, float]],
    task_metrics: dict[tuple[str, str], list[str]],
) -> list[dict[str, Any]]:
    metric_results: list[dict[str, Any]] = []
    for metric_id in task_metrics[task_key]:
        values = [1.0 if result.get(metric_id) else 0.0 for result in sample_results]
        observed_value = sum(values) / len(values) if values else 0.0
        threshold = metric_thresholds[metric_id][model_size]
        metric_results.append(
            {
                "metric_id": metric_id,
                "threshold": threshold,
                "observed_value": observed_value,
                "passed": observed_value >= threshold,
            }
        )
    return metric_results


def build_result_records(
    manifest: dict[str, Any],
    *,
    metric_thresholds: dict[str, dict[str, float]],
    task_metrics: dict[tuple[str, str], list[str]],
) -> list[dict[str, Any]]:
    candidate_model = manifest["candidate_model"]
    model_id = str(candidate_model["model_id"])
    model_size = str(candidate_model["target_size"])
    result_records: list[dict[str, Any]] = []
    for group in manifest["sample_selection"]:
        project = str(group["project"])
        task = str(group["task"])
        task_key = (project, task)
        sample_results: list[dict[str, bool]] = []
        evidence_paths: list[str] = []
        selected_samples = group.get("selected_samples")
        require(isinstance(selected_samples, list) and selected_samples, f"{project}/{task} selected_samples must be non-empty")
        require(len(selected_samples) >= int(group["minimum_sample_count"]), f"{project}/{task} selected sample count below minimum")
        for selected_sample in selected_samples:
            sample_id = str(selected_sample.get("sample_id") or "").strip()
            sample_path_value = str(selected_sample.get("path") or "").strip()
            require(sample_id and sample_path_value, f"{project}/{task} selected sample must include sample_id and path")
            sample_path = normalize_repo_path(sample_path_value)
            require(sample_path.is_file(), f"offline eval selected sample path is missing: {sample_path_value}")
            sample = load_json(sample_path)
            require(isinstance(sample, dict), f"{sample_path_value} must be a JSON object")
            require(sample.get("sample_id") == sample_id, f"{sample_path_value} sample_id mismatch")
            require(sample.get("project") == project, f"{sample_path_value} project mismatch")
            require(sample.get("task") == task, f"{sample_path_value} task mismatch")
            response = sample.get(str(manifest["response_source"]))
            require(isinstance(response, dict), f"{sample_path_value} is missing {manifest['response_source']} object")
            sample_results.append(
                validate_sample_and_response(
                    sample,
                    response,
                    project=project,
                    task=task,
                    sample_name=sample_path.name,
                )
            )
            evidence_paths.append(sample_path_value)

        result_records.append(
            {
                "model_id": model_id,
                "project": project,
                "task": task,
                "sample_count": len(selected_samples),
                "status": "completed",
                "metric_results": build_metric_results(
                    task_key=task_key,
                    model_size=model_size,
                    sample_results=sample_results,
                    metric_thresholds=metric_thresholds,
                    task_metrics=task_metrics,
                ),
                "cost_observation": {
                    "memory_gb": None,
                    "latency_p95_ms": None,
                    "notes": [
                        "Fixture response source; no model process was started."
                    ],
                },
                "evidence_paths": evidence_paths,
            }
        )
    return result_records


def build_eval_run(manifest: dict[str, Any]) -> dict[str, Any]:
    metric_thresholds, task_metrics = load_thresholds()
    result_records = build_result_records(manifest, metric_thresholds=metric_thresholds, task_metrics=task_metrics)
    all_passed = all(
        metric_result.get("passed") is True
        for record in result_records
        for metric_result in get_array(record.get("metric_results"))
    )
    candidate_model = manifest["candidate_model"]
    return {
        "schema_version": 1,
        "kind": "radishmind_core_offline_eval_run",
        "phase": manifest["phase"],
        "run_id": manifest["run_id"],
        "created_at": manifest["created_at"],
        "execution": manifest["execution_policy"],
        "sample_selection": manifest["sample_selection"],
        "candidate_models": [candidate_model],
        "metrics": [
            {
                "metric_id": metric_id,
                "blocking": True,
                "threshold_by_size": metric_thresholds[metric_id],
            }
            for metric_id in metric_thresholds
        ],
        "cost_budget": {
            "target_local_memory_gb": 32,
            "maximum_local_core_size": "7B",
            "excluded_default_sizes": [
                "14B",
                "32B"
            ],
            "per_candidate": [
                {
                    "model_id": candidate_model["model_id"],
                    "run_allowed": True,
                    "max_memory_gb": candidate_model["local_memory_budget_gb"],
                    "max_latency_p95_ms": None,
                    "notes": [
                        "Fixture response source; no runtime cost observed."
                    ],
                }
            ],
        },
        "result_records": result_records,
        "decision": {
            "promotion_status": "no_promotion_planned" if all_passed else "blocked",
            "recommended_next_step": (
                "Offline eval runner is ready for real model candidate output."
                if all_passed
                else "Fix blocking fixture-run metrics before real model inference."
            ),
            "requires_human_review": True,
            "never_promote_when": [
                "model_weakens_requires_confirmation",
                "model_claims_direct_business_write",
                "model_requires_image_pixel_generation_inside_core",
                "model_requires_14b_or_32b_as_default_target",
                "missing_observed_metric_values"
            ],
        },
    }


def main() -> int:
    args = parse_args()
    manifest_path = normalize_repo_path(args.manifest)
    manifest = load_json(manifest_path)
    require(isinstance(manifest, dict), "offline eval fixture run manifest must be an object")
    validate_manifest(manifest)

    run_document = build_eval_run(manifest)
    run_schema = load_json(RUN_SCHEMA_PATH)
    jsonschema.Draft202012Validator.check_schema(run_schema)
    jsonschema.validate(run_document, run_schema)

    if args.output:
        write_json(normalize_repo_path(args.output), run_document)
    if args.check_output:
        expected_path = normalize_repo_path(args.check_output)
        expected = load_json(expected_path)
        if expected != run_document:
            raise SystemExit(f"offline eval run output does not match {expected_path.relative_to(REPO_ROOT)}")

    print("radishmind core offline eval fixture run passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
