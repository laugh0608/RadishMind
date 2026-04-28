#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
TRAINING_SAMPLE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-training-sample.schema.json"
COPILOT_REQUEST_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-request.schema.json"
COPILOT_RESPONSE_SCHEMA_PATH = REPO_ROOT / "contracts/copilot-response.schema.json"
FIXTURE_PATH = REPO_ROOT / "scripts/checks/fixtures/copilot-training-sample-basic.json"


def load_json_document(path: Path) -> Any:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"failed to parse json document '{path.relative_to(REPO_ROOT)}': {exc}") from exc


def assert_condition(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


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
    assert_condition(
        request.get("safety", {}).get("mode") == "advisory",
        "training input_request must preserve advisory safety mode",
    )
    assert_condition(
        request.get("safety", {}).get("requires_confirmation_for_actions") is True,
        "training input_request must require confirmation for actions",
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
        raise SystemExit(f"target_response references missing citation id: {missing_citations[0]}")

    proposed_actions = response.get("proposed_actions") if isinstance(response.get("proposed_actions"), list) else []
    if response.get("requires_confirmation") is True:
        assert_condition(
            any(isinstance(action, dict) and action.get("requires_confirmation") is True for action in proposed_actions),
            "target_response requiring confirmation must include at least one confirming action",
        )
    for action in proposed_actions:
        if not isinstance(action, dict):
            continue
        if action.get("risk_level") in {"medium", "high"}:
            assert_condition(
                action.get("requires_confirmation") is True,
                "medium/high risk training actions must require confirmation",
            )


def main() -> int:
    training_sample_schema = load_json_document(TRAINING_SAMPLE_SCHEMA_PATH)
    copilot_request_schema = load_json_document(COPILOT_REQUEST_SCHEMA_PATH)
    copilot_response_schema = load_json_document(COPILOT_RESPONSE_SCHEMA_PATH)
    fixture = load_json_document(FIXTURE_PATH)

    jsonschema.Draft202012Validator.check_schema(training_sample_schema)
    jsonschema.validate(fixture, training_sample_schema)
    if not isinstance(fixture, dict):
        raise SystemExit("copilot training sample fixture must be an object")
    jsonschema.validate(fixture["input_request"], copilot_request_schema)
    jsonschema.validate(fixture["target_response"], copilot_response_schema)
    check_training_sample(fixture)

    print("copilot training sample contract smoke passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
