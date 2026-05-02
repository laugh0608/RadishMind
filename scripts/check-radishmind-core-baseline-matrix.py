#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
MATRIX_PATH = REPO_ROOT / "scripts/checks/fixtures/radishmind-core-baseline-matrix.json"
DOC_PATH = REPO_ROOT / "docs/radishmind-core-baseline-evaluation.md"

REQUIRED_MODEL_IDS = {
    "minimind-v",
    "radishmind-core-3b",
    "radishmind-core-4b",
    "radishmind-core-7b",
    "qwen2.5-vl",
    "smolvlm",
}
REQUIRED_DIMENSIONS = {
    "protocol_following",
    "structured_response_validity",
    "chinese_task_understanding",
    "citation_alignment",
    "risk_confirmation",
    "action_boundary",
    "training_sample_fit",
    "image_intent_planning",
    "local_deployment_cost",
}
REQUIRED_GATES = {
    "validates_copilot_training_sample",
    "keeps_requires_confirmation_for_medium_high_risk",
    "preserves_citation_ids",
    "does_not_generate_image_pixels",
    "can_be_checked_by_existing_eval_and_gateway_smoke",
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


def check_matrix(matrix: dict[str, Any]) -> None:
    assert_condition(matrix.get("schema_version") == 1, "baseline matrix schema_version must be 1")
    assert_condition(matrix.get("kind") == "radishmind_core_baseline_matrix", "baseline matrix kind mismatch")
    assert_condition(matrix.get("phase") == "M4-preparation", "baseline matrix phase mismatch")

    hardware_budget = matrix.get("hardware_budget") if isinstance(matrix.get("hardware_budget"), dict) else {}
    assert_condition(hardware_budget.get("target_local_memory_gb") == 32, "baseline matrix must keep 32GB budget")
    assert_condition(hardware_budget.get("maximum_local_core_size") == "7B", "7B must remain the local upper bound")
    excluded_sizes = set(require_string_list(hardware_budget, "excluded_default_sizes"))
    assert_condition({"14B", "32B"}.issubset(excluded_sizes), "14B/32B must not be default core targets")

    boundaries = matrix.get("capability_boundaries") if isinstance(matrix.get("capability_boundaries"), dict) else {}
    assert_condition(
        boundaries.get("radishmind_core_is_from_scratch_pretraining") is False,
        "RadishMind-Core must remain base-adapted, not from-scratch pretraining",
    )
    assert_condition(
        boundaries.get("radishmind_core_direct_pixel_generation") is False,
        "RadishMind-Core must not directly generate image pixels",
    )
    assert_condition(
        boundaries.get("image_pixel_generation_owner") == "radishmind-image-adapter-backend",
        "image pixel generation must stay with image adapter/backend",
    )
    assert_condition(boundaries.get("advisory_only") is True, "baseline matrix must preserve advisory-only boundary")

    dimensions = set(require_string_list(matrix, "evaluation_dimensions"))
    missing_dimensions = sorted(REQUIRED_DIMENSIONS - dimensions)
    if missing_dimensions:
        raise SystemExit(f"baseline matrix is missing evaluation dimension: {missing_dimensions[0]}")

    gates = set(require_string_list(matrix, "minimum_entry_gates"))
    missing_gates = sorted(REQUIRED_GATES - gates)
    if missing_gates:
        raise SystemExit(f"baseline matrix is missing entry gate: {missing_gates[0]}")

    models = matrix.get("models")
    assert_condition(isinstance(models, list), "baseline matrix models must be a list")
    model_by_id = {
        str(model.get("id") or "").strip(): model
        for model in models
        if isinstance(model, dict) and str(model.get("id") or "").strip()
    }
    missing_models = sorted(REQUIRED_MODEL_IDS - set(model_by_id))
    if missing_models:
        raise SystemExit(f"baseline matrix is missing model: {missing_models[0]}")

    assert_condition(
        model_by_id["minimind-v"].get("role") == "student_base_mainline",
        "minimind-v must remain the student/base mainline",
    )
    assert_condition(
        model_by_id["qwen2.5-vl"].get("role") == "teacher_strong_baseline",
        "Qwen2.5-VL must remain teacher/strong baseline",
    )
    assert_condition(
        model_by_id["smolvlm"].get("role") == "lightweight_baseline",
        "SmolVLM must remain lightweight baseline",
    )
    assert_condition(
        model_by_id["radishmind-core-3b"].get("current_status") == "preferred_first_round_size",
        "3B must remain the preferred first-round size",
    )
    assert_condition(
        model_by_id["radishmind-core-7b"].get("current_status") == "deferred_until_3b_4b_gap_is_proven",
        "7B must remain deferred until 3B/4B gaps are proven",
    )

    for model_id, model in model_by_id.items():
        must_not_do = set(require_string_list(model, "must_not_do"))
        if model_id.startswith("radishmind-core") or model_id == "minimind-v":
            assert_condition(
                "handle_direct_pixel_generation" in must_not_do
                or "generate_image_pixels" in must_not_do,
                f"{model_id} must explicitly exclude direct image pixel generation",
            )

    decision_rules = matrix.get("decision_rules") if isinstance(matrix.get("decision_rules"), dict) else {}
    never_do = set(require_string_list(decision_rules, "never_do_by_default"))
    assert_condition(
        "merge_image_pixel_generation_into_radishmind_core" in never_do,
        "decision rules must forbid merging image pixel generation into RadishMind-Core",
    )
    assert_condition(
        "treat_14b_or_32b_as_default_core_target" in never_do,
        "decision rules must forbid 14B/32B default core targets",
    )


def check_doc() -> None:
    content = DOC_PATH.read_text(encoding="utf-8")
    for pattern in (
        "scripts/checks/fixtures/radishmind-core-baseline-matrix.json",
        "scripts/check-radishmind-core-baseline-matrix.py",
        "图片像素生成不进入主模型参数目标",
        "`3B`",
        "`4B`",
        "`7B`",
        "`minimind-v`",
        "`Qwen2.5-VL`",
        "`SmolVLM`",
    ):
        assert_condition(pattern in content, f"baseline evaluation doc is missing expected content: {pattern}")


def main() -> int:
    matrix = load_json_document(MATRIX_PATH)
    if not isinstance(matrix, dict):
        raise SystemExit("baseline matrix fixture must be an object")
    check_matrix(matrix)
    check_doc()
    print("radishmind core baseline matrix check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
