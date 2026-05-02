#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


DOCS_CONFLICT_SAMPLE = (
    REPO_ROOT / "datasets/eval/radish/answer-docs-question-docs-faq-forum-conflict-001.json"
)
EVIDENCE_GAP_SAMPLE = REPO_ROOT / "datasets/eval/radish/answer-docs-question-evidence-gap-001.json"
EFFICIENCY_RANGE_SAMPLE = (
    REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-efficiency-range-ordering-001.json"
)


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def load_json(path: Path) -> dict[str, Any]:
    document = json.loads(path.read_text(encoding="utf-8"))
    require(isinstance(document, dict), f"{path.relative_to(REPO_ROOT)} must be a JSON object")
    return document


def load_candidate_runner() -> Any:
    module_path = REPO_ROOT / "scripts/run-radishmind-core-candidate.py"
    spec = importlib.util.spec_from_file_location("radishmind_core_candidate_runner", module_path)
    require(spec is not None and spec.loader is not None, "candidate runner module spec must be loadable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def freeze_fields_by_path(prompt_document: dict[str, Any]) -> dict[str, Any]:
    freeze = prompt_document.get("hard_field_freeze")
    require(isinstance(freeze, dict), "prompt document must include hard_field_freeze")
    fields = freeze.get("fields")
    require(isinstance(fields, list) and fields, "hard_field_freeze.fields must be a non-empty list")
    result: dict[str, Any] = {}
    for field in fields:
        require(isinstance(field, dict), "hard_field_freeze.fields entries must be objects")
        path = field.get("path")
        require(isinstance(path, str) and path.startswith("$."), "freeze field path must be a JSON path")
        require(path not in result, f"freeze field path must be unique: {path}")
        result[path] = field.get("value")
    return result


def build_prompt(runner: Any, sample: dict[str, Any]) -> dict[str, Any]:
    return runner.build_prompt_document(sample, model_id="check-hard-field-freeze")


def assert_prompt_mentions_freeze(prompt_document: dict[str, Any]) -> None:
    user_content = str(prompt_document["messages"][1]["content"])
    require("hard_field_freeze" in user_content, "prompt text must include the hard_field_freeze block")
    require(
        "必须逐项照抄 value" in user_content,
        "prompt text must describe freeze values as copy-exact constraints",
    )


def main() -> int:
    runner = load_candidate_runner()

    evidence_gap = load_json(EVIDENCE_GAP_SAMPLE)
    evidence_prompt = build_prompt(runner, evidence_gap)
    assert_prompt_mentions_freeze(evidence_prompt)
    evidence_fields = freeze_fields_by_path(evidence_prompt)
    require(evidence_fields.get("$.status") == "partial", "evidence-gap status must be frozen as partial")
    require(evidence_fields.get("$.risk_level") == "medium", "evidence-gap risk_level must be frozen as medium")
    require(evidence_fields.get("$.requires_confirmation") is False, "evidence-gap confirmation must be frozen false")
    require(evidence_fields.get("$.proposed_actions") == [], "evidence-gap no-action boundary must be frozen")
    require(
        "$.answers[0].kind" not in evidence_fields,
        "evidence-gap answer kind must not be frozen without an explicit answer-kind assertion",
    )

    docs_conflict = load_json(DOCS_CONFLICT_SAMPLE)
    docs_fields = freeze_fields_by_path(build_prompt(runner, docs_conflict))
    require(docs_fields.get("$.status") == "ok", "docs source-conflict status must be frozen as ok")
    require(docs_fields.get("$.risk_level") == "low", "docs source-conflict risk_level must be frozen as low")
    require(
        docs_fields.get("$.proposed_actions[0].kind") == "read_only_check",
        "docs source-conflict required read_only_check action kind must be frozen",
    )
    require(
        docs_fields.get("$.proposed_actions[0].requires_confirmation") is False,
        "docs source-conflict read_only_check confirmation boundary must be frozen",
    )

    efficiency_range = load_json(EFFICIENCY_RANGE_SAMPLE)
    efficiency_fields = freeze_fields_by_path(build_prompt(runner, efficiency_range))
    require(efficiency_fields.get("$.status") == "partial", "efficiency-range status must be frozen as partial")
    require(efficiency_fields.get("$.risk_level") == "medium", "efficiency-range risk_level must be frozen as medium")
    require(
        efficiency_fields.get("$.requires_confirmation") is True,
        "efficiency-range top-level confirmation must be frozen true",
    )
    require(
        efficiency_fields.get("$.proposed_actions[0].patch")
        == {
            "parameter_updates": {
                "efficiency_percent": {
                    "action": "clamp_to_review_range",
                    "suggested_range": [65, 82],
                }
            },
            "preserve_topology": True,
        },
        "efficiency-range parameter update patch must be frozen",
    )

    print("radishmind core candidate hard-field freeze check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
