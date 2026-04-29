from __future__ import annotations

from pathlib import Path
from typing import Any


TRAIN_FIELDS: tuple[str, ...] = (
    "summary",
    "answers",
    "issues",
    "proposed_actions",
    "citations",
    "confidence",
    "risk_level",
    "requires_confirmation",
)


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def collect_referenced_citation_ids(response: dict[str, Any]) -> set[str]:
    referenced_ids: set[str] = set()
    for section_name in ("answers", "issues", "proposed_actions"):
        values = response.get(section_name)
        if not isinstance(values, list):
            continue
        for item in values:
            if not isinstance(item, dict):
                continue
            citation_ids = item.get("citation_ids")
            if isinstance(citation_ids, list):
                referenced_ids.update(str(value) for value in citation_ids if str(value).strip())
    return referenced_ids


def iter_proposed_actions(response: dict[str, Any]) -> list[dict[str, Any]]:
    values = response.get("proposed_actions")
    if not isinstance(values, list):
        return []
    return [value for value in values if isinstance(value, dict)]


def action_needs_confirmation_boundary(action: dict[str, Any]) -> bool:
    risk_level = action.get("risk_level")
    if risk_level == "high":
        return True
    if risk_level != "medium":
        return False
    if action.get("kind") != "ghost_completion":
        return True
    preview = action.get("preview") if isinstance(action.get("preview"), dict) else {}
    return preview.get("accept_key") != "manual_only"


def check_training_sample(sample: dict[str, Any]) -> None:
    request = sample.get("input_request")
    response = sample.get("target_response")
    distillation = sample.get("distillation") if isinstance(sample.get("distillation"), dict) else {}
    quality_gates = sample.get("quality_gates") if isinstance(sample.get("quality_gates"), dict) else {}

    assert_condition(sample.get("schema_version") == 1, "copilot training sample schema_version must be 1")
    assert_condition(sample.get("kind") == "copilot_training_sample", "copilot training sample kind mismatch")
    assert_condition(isinstance(request, dict), "copilot training sample input_request must be an object")
    assert_condition(isinstance(response, dict), "copilot training sample target_response must be an object")
    assert_condition(sample.get("project") == request.get("project"), "sample project must match input_request.project")
    assert_condition(sample.get("project") == response.get("project"), "sample project must match target_response.project")
    assert_condition(sample.get("task") == request.get("task"), "sample task must match input_request.task")
    assert_condition(sample.get("task") == response.get("task"), "sample task must match target_response.task")

    safety = request.get("safety", {}) if isinstance(request.get("safety"), dict) else {}
    assert_condition(
        safety.get("mode") == "advisory",
        "training input_request must preserve advisory safety mode",
    )
    assert_condition(quality_gates.get("schema_validated") is True, "training sample must be schema validated")
    assert_condition(quality_gates.get("risk_reviewed") is True, "training sample must be risk reviewed")
    assert_condition(quality_gates.get("citation_checked") is True, "training sample must be citation checked")

    train_fields = distillation.get("train_fields") if isinstance(distillation.get("train_fields"), list) else []
    for required_field in ("summary", "risk_level", "requires_confirmation"):
        assert_condition(required_field in train_fields, f"training sample train_fields must include {required_field}")

    citation_ids = {
        str(citation.get("id"))
        for citation in response.get("citations", [])
        if isinstance(citation, dict) and str(citation.get("id") or "").strip()
    }
    referenced_ids = collect_referenced_citation_ids(response)
    missing_citations = sorted(referenced_ids - citation_ids)
    if missing_citations:
        raise ValueError(f"target_response references missing citation id: {missing_citations[0]}")

    proposed_actions = iter_proposed_actions(response)
    has_confirming_action = any(action.get("requires_confirmation") is True for action in proposed_actions)
    has_confirmation_boundary_action = any(action_needs_confirmation_boundary(action) for action in proposed_actions)
    request_requires_confirmation = safety.get("requires_confirmation_for_actions") is True

    if response.get("requires_confirmation") is True:
        assert_condition(
            has_confirming_action,
            "target_response requiring confirmation must include at least one confirming action",
        )
        assert_condition(
            request_requires_confirmation,
            "training input_request must require confirmation when target_response does",
        )

    for action in proposed_actions:
        if action_needs_confirmation_boundary(action):
            assert_condition(
                action.get("requires_confirmation") is True,
                "confirmation-boundary training actions must require confirmation",
            )

    if has_confirming_action or has_confirmation_boundary_action:
        assert_condition(
            request_requires_confirmation,
            "training input_request must require confirmation for confirmation-boundary actions",
        )


def build_training_sample_from_eval(
    eval_sample: dict[str, Any],
    *,
    source_eval_sample: Path,
    source_eval_sample_label: str,
    created_for: str = "teacher-student-distillation",
) -> dict[str, Any]:
    sample_id = str(eval_sample.get("sample_id") or "").strip()
    project = str(eval_sample.get("project") or "").strip()
    task = str(eval_sample.get("task") or "").strip()
    input_request = eval_sample.get("input_request")
    golden_response = eval_sample.get("golden_response")
    if not sample_id:
        raise ValueError(f"{source_eval_sample_label} is missing sample_id")
    if not project or not task:
        raise ValueError(f"{source_eval_sample_label} is missing project/task")
    if not isinstance(input_request, dict):
        raise ValueError(f"{source_eval_sample_label} is missing input_request")
    if not isinstance(golden_response, dict):
        raise ValueError(f"{source_eval_sample_label} is missing golden_response")

    return {
        "schema_version": 1,
        "kind": "copilot_training_sample",
        "sample_id": f"{sample_id}-training-golden-001",
        "training_mode": "distillation",
        "project": project,
        "task": task,
        "input_request": input_request,
        "target_response": golden_response,
        "distillation": {
            "source": "golden_response",
            "teacher": {
                "provider": "fixture",
                "model": "golden-response",
                "record_id": sample_id,
            },
            "train_fields": list(TRAIN_FIELDS),
        },
        "quality_gates": {
            "schema_validated": True,
            "risk_reviewed": True,
            "citation_checked": True,
            "human_review_required": False,
        },
        "metadata": {
            "source_eval_sample": source_eval_sample.as_posix(),
            "created_for": created_for,
            "notes": [
                "Generated from committed eval input_request and golden_response; no model inference is run.",
            ],
        },
    }


def build_training_sample_from_candidate_record(
    eval_sample: dict[str, Any],
    candidate_record: dict[str, Any],
    *,
    source_eval_sample: Path,
    source_eval_sample_label: str,
    candidate_record_label: str,
    created_for: str = "teacher-student-distillation",
) -> dict[str, Any]:
    sample_id = str(eval_sample.get("sample_id") or "").strip()
    project = str(eval_sample.get("project") or "").strip()
    task = str(eval_sample.get("task") or "").strip()
    input_request = eval_sample.get("input_request")
    target_response = candidate_record.get("response")
    record_id = str(candidate_record.get("record_id") or "").strip()
    model = str(candidate_record.get("model") or "").strip()
    source = str(candidate_record.get("source") or "").strip()
    if not sample_id:
        raise ValueError(f"{source_eval_sample_label} is missing sample_id")
    if not project or not task:
        raise ValueError(f"{source_eval_sample_label} is missing project/task")
    if not isinstance(input_request, dict):
        raise ValueError(f"{source_eval_sample_label} is missing input_request")
    if not isinstance(target_response, dict):
        raise ValueError(f"{candidate_record_label} is missing response")
    if candidate_record.get("sample_id") != sample_id:
        raise ValueError(f"{candidate_record_label} sample_id does not match {source_eval_sample_label}")
    if candidate_record.get("project") != project or candidate_record.get("task") != task:
        raise ValueError(f"{candidate_record_label} project/task does not match {source_eval_sample_label}")
    if not record_id or not model or not source:
        raise ValueError(f"{candidate_record_label} is missing record_id/model/source")

    return {
        "schema_version": 1,
        "kind": "copilot_training_sample",
        "sample_id": f"{sample_id}-training-capture-001",
        "training_mode": "distillation",
        "project": project,
        "task": task,
        "input_request": input_request,
        "target_response": target_response,
        "distillation": {
            "source": "teacher_capture",
            "teacher": {
                "provider": source,
                "model": model,
                "record_id": record_id,
            },
            "train_fields": list(TRAIN_FIELDS),
        },
        "quality_gates": {
            "schema_validated": True,
            "risk_reviewed": True,
            "citation_checked": True,
            "human_review_required": False,
        },
        "metadata": {
            "source_eval_sample": source_eval_sample.as_posix(),
            "created_for": created_for,
            "notes": [
                f"Generated from audited candidate response record: {candidate_record_label}.",
                "No model inference is run during conversion.",
            ],
        },
    }
