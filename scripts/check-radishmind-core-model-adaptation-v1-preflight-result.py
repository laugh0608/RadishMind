#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
RESULT_PATH = REPO_ROOT / "training/experiments/radishmind-core-model-adaptation-v1-preflight-result-v0.json"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse {path.relative_to(REPO_ROOT)}: {exc}") from exc


def require_existing_file(path_value: str, *, field_name: str) -> None:
    require(path_value.strip() == path_value and path_value.strip(), f"{field_name} must be a clean non-empty path")
    require((REPO_ROOT / path_value).is_file(), f"{field_name} points to a missing file: {path_value}")


def require_tmp_path(path_value: str, *, field_name: str) -> None:
    require(path_value.startswith("tmp/"), f"{field_name} must stay under tmp/: {path_value}")


def check_scope(document: dict[str, Any]) -> None:
    scope = document.get("scope")
    require(isinstance(scope, dict), "scope must be an object")
    require(scope.get("sample_set_id") == "full-holdout-9", "sample_set_id mismatch")
    require(scope.get("sample_count") == 9, "sample_count must be 9")
    require(scope.get("task_counts") == {
        "radish/answer_docs_question": 3,
        "radishflow/suggest_flowsheet_edits": 3,
        "radishflow/suggest_ghost_completion": 3,
    }, "task_counts mismatch")
    require_existing_file(str(scope.get("candidate_manifest") or ""), field_name="candidate_manifest")
    require_existing_file(str(scope.get("candidate_eval_manifest") or ""), field_name="candidate_eval_manifest")


def check_artifact_paths(section: dict[str, Any], *, section_name: str) -> None:
    artifact_paths = section.get("artifact_paths")
    require(isinstance(artifact_paths, dict), f"{section_name}.artifact_paths must be an object")
    for key in ("candidate_output_dir", "candidate_summary", "offline_eval_run"):
        require_tmp_path(str(artifact_paths.get(key) or ""), field_name=f"{section_name}.{key}")


def check_raw_student(document: dict[str, Any]) -> None:
    raw = document.get("raw_student")
    require(isinstance(raw, dict), "raw_student must be an object")
    require(raw.get("candidate_track") == "raw_student", "raw candidate_track mismatch")
    check_artifact_paths(raw, section_name="raw_student")
    observation = raw.get("candidate_summary_observation")
    require(isinstance(observation, dict), "raw observation must be an object")
    require(observation.get("sample_count") == 9, "raw sample_count mismatch")
    require(observation.get("schema_valid_rate") == 0.6666666666666666, "raw schema_valid_rate mismatch")
    require(observation.get("schema_invalid_rate") == 0.3333333333333333, "raw schema_invalid_rate mismatch")
    require(observation.get("task_valid_rate_for_valid_responses") == 1.0, "raw task rate mismatch")
    require(observation.get("task_validation_attempted") == 6, "raw task validation attempted mismatch")
    require(observation.get("timeout_count") == 0, "raw timeout_count must be 0")
    require(observation.get("hit_max_new_tokens_count") == 0, "raw hit_max_new_tokens_count must be 0")
    require(observation.get("json_extracted_count") == 9, "raw json_extracted_count must be 9")

    task_results = raw.get("task_results")
    require(isinstance(task_results, list) and len(task_results) == 3, "raw task_results must include three tasks")
    by_task = {
        f"{entry.get('project')}/{entry.get('task')}": entry
        for entry in task_results
        if isinstance(entry, dict)
    }
    edits = by_task.get("radishflow/suggest_flowsheet_edits")
    require(isinstance(edits, dict), "raw edits result missing")
    require(edits.get("status") == "blocked", "raw edits must remain blocked")
    require(edits.get("schema_valid_rate") == 0.0, "raw edits schema_valid_rate must be 0")
    require("copy_from_hard_field_freeze" in str(edits.get("blocking_reason") or ""), "raw edits blocker must mention scaffold leakage")
    require(by_task.get("radishflow/suggest_ghost_completion", {}).get("status") == "passed_machine_gate", "raw ghost status mismatch")
    require(by_task.get("radish/answer_docs_question", {}).get("status") == "passed_machine_gate", "raw docs status mismatch")

    decision = raw.get("offline_eval_decision")
    require(isinstance(decision, dict), "raw offline_eval_decision must be an object")
    require(decision.get("promotion_status") == "blocked", "raw promotion_status must stay blocked")
    require(decision.get("requires_human_review") is True, "raw result must require human review")


def check_repaired_comparison(document: dict[str, Any]) -> None:
    repaired = document.get("repaired_student_comparison")
    require(isinstance(repaired, dict), "repaired_student_comparison must be an object")
    require(repaired.get("candidate_track") == "repair_hard_fields", "repaired candidate_track mismatch")
    check_artifact_paths(repaired, section_name="repaired_student_comparison")
    observation = repaired.get("candidate_summary_observation")
    require(isinstance(observation, dict), "repaired observation must be an object")
    require(observation.get("sample_count") == 9, "repaired sample_count mismatch")
    require(observation.get("schema_valid_rate") == 1.0, "repaired schema_valid_rate mismatch")
    require(observation.get("task_valid_rate") == 1.0, "repaired task_valid_rate mismatch")
    require(observation.get("task_validation_attempted") == 9, "repaired task_validation_attempted mismatch")
    require(observation.get("timeout_count") == 0, "repaired timeout_count must be 0")
    require(observation.get("hit_max_new_tokens_count") == 0, "repaired hit_max_new_tokens_count must be 0")
    require(observation.get("json_extracted_count") == 9, "repaired json_extracted_count must be 9")

    postprocess = repaired.get("postprocess_observation")
    require(isinstance(postprocess, dict), "postprocess_observation must be an object")
    require(postprocess.get("repaired_output_count") == 3, "repaired_output_count mismatch")
    counts = postprocess.get("repaired_path_counts")
    require(isinstance(counts, dict), "repaired_path_counts must be an object")
    require(counts.get("$.citations") == 3, "repaired citations count must be 3")

    decision = repaired.get("offline_eval_decision")
    require(isinstance(decision, dict), "repaired offline_eval_decision must be an object")
    require(decision.get("promotion_status") == "no_promotion_planned", "repaired promotion_status mismatch")
    require(decision.get("requires_human_review") is True, "repaired result must require human review")


def check_decision_and_policy(document: dict[str, Any]) -> None:
    decision = document.get("decision")
    require(isinstance(decision, dict), "decision must be an object")
    require(decision.get("raw_student_promotion_status") == "blocked", "raw_student_promotion_status must be blocked")
    require(decision.get("repaired_result_role") == "postprocess_comparison_only", "repaired result role mismatch")
    route_conclusion = str(decision.get("route_conclusion") or "")
    require("raw" in route_conclusion and "blocked" in route_conclusion, "route conclusion must keep raw blocked")
    require("must not be treated as raw model promotion" in route_conclusion, "route conclusion must reject raw promotion")
    require("training acceptance" in route_conclusion, "route conclusion must reject training acceptance")

    artifact_policy = document.get("artifact_policy")
    require(isinstance(artifact_policy, dict), "artifact_policy must be an object")
    require(artifact_policy.get("generated_outputs_default_location") == "tmp/", "generated outputs must stay under tmp/")
    disallowed = artifact_policy.get("committed_disallowed")
    require(isinstance(disallowed, list), "committed_disallowed must be a list")
    for item in (
        "candidate response files",
        "prompt files",
        "provider raw dump",
        "training JSONL",
        "model weights",
        "checkpoint",
        "adapter binary",
    ):
        require(item in disallowed, f"artifact policy must disallow {item}")


def main() -> int:
    document = load_json(RESULT_PATH)
    require(isinstance(document, dict), "preflight result must be a JSON object")
    require(document.get("schema_version") == 1, "schema_version must be 1")
    require(document.get("kind") == "radishmind_core_model_adaptation_v1_preflight_result", "kind mismatch")
    require(document.get("result_id") == "radishmind-core-model-adaptation-v1-preflight-result-v0", "result_id mismatch")
    require(document.get("status") == "completed_no_raw_promotion", "status must keep no raw promotion")
    require(document.get("phase") == "P4-preflight", "phase mismatch")

    model = document.get("model")
    require(isinstance(model, dict), "model must be an object")
    require(model.get("provider") == "local_transformers", "provider mismatch")
    require(model.get("model_id") == "Qwen2.5-1.5B-Instruct", "model_id mismatch")
    require(model.get("device") == "cpu", "device mismatch")
    require(model.get("does_not_download_model_artifacts") is True, "model artifacts must not be downloaded")

    check_scope(document)
    check_raw_student(document)
    check_repaired_comparison(document)
    check_decision_and_policy(document)

    print("radishmind core model adaptation v1 preflight result check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
