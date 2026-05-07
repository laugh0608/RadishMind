#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path
from typing import Any

import jsonschema


REPO_ROOT = Path(__file__).resolve().parent.parent
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


EFFICIENCY_RANGE_SAMPLE = (
    REPO_ROOT / "datasets/eval/radishflow/suggest-flowsheet-edits-efficiency-range-ordering-001.json"
)
CROSS_OBJECT_SAMPLE = (
    REPO_ROOT
    / "datasets/eval/radishflow/suggest-flowsheet-edits-cross-object-mixed-risk-reconnect-plus-pump-parameter-001.json"
)
RESPONSE_SCHEMA = REPO_ROOT / "contracts/copilot-response.schema.json"


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


def assert_valid_response(runner: Any, sample: dict[str, Any], response: dict[str, Any]) -> None:
    response_schema = load_json(RESPONSE_SCHEMA)
    jsonschema.validate(response, response_schema)
    runner.validate_candidate_response(response, sample=sample, response_schema=response_schema)
    violations = runner.validate_task_candidate_response(
        response,
        sample=sample,
        sample_name=f"{sample['sample_id']}.json",
    )
    require(not violations, "response builder output must pass suggest_flowsheet_edits task validation: " + "; ".join(violations))


def main() -> int:
    runner = load_candidate_runner()

    cross_object = load_json(CROSS_OBJECT_SAMPLE)
    built_cross_object, cross_paths = runner.build_suggest_edits_response(None, sample=cross_object)
    assert_valid_response(runner, cross_object, built_cross_object)
    require("$.proposed_actions" in cross_paths, "builder must report the action scaffold path")
    require(
        [action.get("target") for action in built_cross_object["proposed_actions"]]
        == [{"type": "stream", "id": "cooler-outlet-81"}, {"type": "unit", "id": "pump-12"}],
        "builder must preserve cross-object ordered action targets",
    )
    require(
        [citation.get("id") for citation in built_cross_object["citations"]]
        == [
            "diag-1",
            "diag-2",
            "diag-3",
            "diag-4",
            "flowdoc-stream-3",
            "flowdoc-unit-1",
            "flowdoc-unit-2",
            "flowdoc-stream-2",
            "snapshot-1",
        ],
        "builder must preserve cross-object ordered citations",
    )

    efficiency_range = load_json(EFFICIENCY_RANGE_SAMPLE)
    raw_intent_response = {
        "schema_version": 1,
        "status": "ok",
        "project": "radishflow",
        "task": "suggest_flowsheet_edits",
        "summary": "raw model summary about pump efficiency",
        "answers": [
            {
                "kind": "edit_rationale",
                "text": "raw model rationale text",
                "citation_ids": [],
            }
        ],
        "issues": [
            {
                "code": "WRONG_CODE",
                "message": "raw model issue message",
                "severity": "warning",
                "citation_ids": [],
            }
        ],
        "proposed_actions": [
            {
                "kind": "candidate_edit",
                "title": "raw model title",
                "target": {"type": "unit", "id": "wrong"},
                "rationale": "raw model action rationale",
                "patch": {},
                "risk_level": "low",
                "requires_confirmation": False,
                "citation_ids": [],
            }
        ],
        "citations": [],
        "confidence": 0.42,
        "risk_level": "low",
        "requires_confirmation": False,
    }
    built_efficiency, efficiency_paths = runner.build_suggest_edits_response(
        raw_intent_response,
        sample=efficiency_range,
    )
    assert_valid_response(runner, efficiency_range, built_efficiency)
    require(
        built_efficiency["summary"] == "当前更合适的是围绕 pump-3 输出局部效率复核提案，并保持 suggested_range 继续按下界到上界的顺序表达。",
        "builder must keep efficiency summary task-grounded instead of preserving misleading raw summary",
    )
    require(
        built_efficiency["answers"][0]["text"]
        == "pump-3 的 efficiency_percent 当前已超出建议运行区间，因此更合适的是保留单对象参数 patch；同时 suggested_range 应继续保持 65 到 82 这种下界在前、上界在后的可审查表达。",
        "builder must keep efficiency answer task-grounded while rebuilding answer citations",
    )
    require(
        built_efficiency["issues"][0]["message"] == "pump-3 的 efficiency_percent=109 超出了当前建议运行区间。",
        "builder must keep efficiency issue message task-grounded while rebuilding code/severity/citations",
    )
    require(
        built_efficiency["proposed_actions"][0]["title"] == "复核 pump-3 的效率建议区间",
        "builder must keep efficiency action title task-grounded",
    )
    require(
        built_efficiency["proposed_actions"][0]["rationale"]
        == "pump-3 的效率参数当前仍需要局部复核，因此更合适的是保留带 suggested_range 的单对象参数 patch 供人工审查。",
        "builder must keep efficiency action rationale task-grounded",
    )
    require(built_efficiency["confidence"] == 0.42, "builder must preserve bounded raw confidence")
    require(
        built_efficiency["proposed_actions"][0]["target"] == {"type": "unit", "id": "pump-3"},
        "builder must rebuild the stable action target instead of trusting raw target drift",
    )
    require(
        built_efficiency["proposed_actions"][0]["patch"]
        == {
            "parameter_updates": {
                "efficiency_percent": {
                    "action": "clamp_to_review_range",
                    "suggested_range": [65, 82],
                }
            },
            "preserve_topology": True,
        },
        "builder must rebuild the stable parameter update patch",
    )
    require(
        [citation.get("id") for citation in built_efficiency["citations"]] == ["diag-1", "flowdoc-unit-1", "snapshot-1"],
        "builder must keep efficiency citations aligned with diagnostic, unit, and snapshot evidence",
    )
    require("$.summary" not in efficiency_paths, "builder must reject misleading raw efficiency summary merge")
    require("$.answers[0].text" not in efficiency_paths, "builder must reject misleading raw efficiency answer merge")
    require("$.proposed_actions[0].title" not in efficiency_paths, "builder must reject misleading raw efficiency action title")

    print("radishmind core suggest edits response builder check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
